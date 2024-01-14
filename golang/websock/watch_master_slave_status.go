package websock

import (
	"encoding/json"
	"fmt"
	"golang/database"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func getInstanceStatus(conn *websocket.Conn) {
	var err error
	var msg []byte
	var instanceList []string //format is ip_port
	var lock sync.RWMutex
	var closeConn bool
	newGoFunc := true
	GofuncLock := make(chan int, 0)
	instanceStatus := make(map[string]*InstanceStatus)
	dbConn := make(map[string]*database.MySQLConn)
	dbCheck := make(map[string]bool)
	defer func() {
		conn.Close()
		close(GofuncLock)
		for _, v := range dbConn {
			if v != nil {
				v.Close()
			}
		}
	}()
	for {
		// get message from client
		_, msg, err = conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read message:", err)
			closeConn = true
			break
		}
		data := new(QueryInstanceInstanceRequest)
		resp := new(QueryInstanceSendMsg)
		if err = json.Unmarshal(msg, data); err != nil {
			log.Println("Failed to read message:", err)
			err = fmt.Errorf("unmarshal message error： %s", err.Error())
			resp.Status = "0001"
			resp.Msg = err.Error()
			senMsg(conn, resp)
			continue
		}
		if len(data.InstanceMark) == 0 {
			conn.WriteMessage(websocket.PongMessage, []byte{})
		} else {
			instanceList = []string{}
			//you can confirm data current in this module
			if err = checkDataCurrent(data.InstanceMark); err != nil {
				resp.Status = "0001"
				resp.Msg = err.Error()
				senMsg(conn, resp)
				return
			}
			for _, v := range data.InstanceMark {
				instanceList = append(instanceList, v)
			}
			//when newGoFunc is false, it means that the coroutine being queried is still running
			//it is necessary to lock and wait for the coroutine to exit before opening a new coroutine to query
			if newGoFunc == false {
				newGoFunc = true
				GofuncLock <- 1
			}
		}
		if len(instanceList) != 0 && newGoFunc {
			//change status means current query coroutine is running
			newGoFunc = false
			resp := new(QueryInstanceSendMsg)
			resp.Data = new(InstanceStatus)
			//check the truly runnning status
			go func() {
				//recover
				defer recoverWsAndSentMsg(conn, resp)
				for {
					for _, v := range instanceList {
						if instanceStatus[v] == nil {
							instanceStatus[v] = new(InstanceStatus)
						}
						lock.RLock()
						check := dbCheck[v]
						lock.RUnlock()
						if check == false {
							lock.Lock()
							go func(iMark string) {
								defer recoverWsAndSentMsg(conn, resp)
								dbCheck[iMark] = true
								lock.Unlock()
								var err error
								defer func() {
									lock.Lock()
									dbCheck[iMark] = false
									lock.Unlock()
								}()

								resp.Data.InstanceMark = iMark
								instanceData := new(database.QuerySlaveStatusResult)
								lock.RLock()
								connect := dbConn[iMark]
								lock.RUnlock()
								if connect == nil {
									info := strings.Split(iMark, "_")
									host := ""
									port := 0
									//when length is not 2,it means format is not ip_port
									if len(info) != 2 {
										resp.Status = "0001"
										resp.Msg = fmt.Sprintf("format of %s is error", iMark)
										lock.Lock()
										senMsg(conn, resp)
										lock.Unlock()
										return
									} else {
										host = info[0]
										port, _ = strconv.Atoi(info[1])
									}

									dbSvc := &database.MySQLConn{
										Host:     host,
										Port:     port,
										Username: "your username",
										Password: "your password",
										DBName:   "mysql",
									}
									err = dbSvc.NewConn()
									if err != nil {
										resp.Status = "0001"
										resp.Msg = err.Error()
										resp.Data.LastError = err.Error()
										if checkIfInstanceStatusChange(instanceStatus[iMark], resp.Data) {
											lock.Lock()
											senMsg(conn, resp)
											lock.Unlock()
											return
										}
										return
									}
									resp.Data.RunningStatus = true
									lock.Lock()
									dbConn[iMark] = new(database.MySQLConn)
									dbConn[iMark] = dbSvc
									lock.Unlock()
								} else {
									a := &struct {
										A string
									}{}
									if err = dbConn[iMark].DB.Raw("select 1 as a").Scan(a).Error; err != nil {
										resp.Status = "0001"
										resp.Msg = fmt.Sprintf("连接已断开 %s", err.Error())
										resp.Data.RunningStatus = false
										lock.Lock()
										dbConn[iMark] = nil
										lock.Unlock()
									} else {
										resp.Data.RunningStatus = true
									}
								}
								if resp.Data.RunningStatus {
									//get master-slave status and latency
									if instanceData, err = dbConn[iMark].QuerySlaveStatus(); err != nil {
										resp.Status = "0001"
										resp.Msg = err.Error()
										resp.Data.SlaveStatus = false
										//compare to instanceData.LastError
										instanceData.LastError = err.Error()
									}
									if (instanceData.LastError != "" && instanceData.LastError != "0") || instanceData.SlaveIORunning != "Yes" || instanceData.SlaveSQLRunning != "Yes" {
										resp.Data.SlaveStatus = false
										resp.Data.LastError = instanceData.LastError
									} else {
										resp.Data.SlaveStatus = true
										resp.Data.SecondsBehindMaster = instanceData.SecondsBehindMaster
										resp.Data.LastError = ""
									}
								}

								//send message when somgthings is changed
								if checkIfInstanceStatusChange(instanceStatus[iMark], resp.Data) {
									lock.Lock()
									senMsg(conn, resp)
									lock.Unlock()
									return
								}
							}(v)
						}

					}
					time.Sleep(time.Second)
					//If newGoFunc is true, it means that the upper-level loop has received a new message and the current coroutine needs to be returned and re-run
					if newGoFunc {
						defer func() {
							<-GofuncLock
						}()
						return
					}
					//close
					if closeConn {
						return
					}
				}
			}()
		}
	}

	return
}
func senMsg(conn *websocket.Conn, data interface{}) (err error) {
	msg := []byte{}
	if msg, err = json.Marshal(data); err != nil {
		err = fmt.Errorf("send message error %s", err.Error())
		return
	}
	return conn.WriteMessage(websocket.TextMessage, msg)
}

func checkIfInstanceStatusChange(old *InstanceStatus, new *InstanceStatus) (res bool) {
	res = !(old.RunningStatus == new.RunningStatus && old.LastError == new.LastError && old.SecondsBehindMaster == new.SecondsBehindMaster && old.SlaveStatus == new.SlaveStatus)
	old.RunningStatus = new.RunningStatus
	old.SlaveStatus = new.SlaveStatus
	old.LastError = new.LastError
	old.SecondsBehindMaster = new.SecondsBehindMaster
	return
}

func checkDataCurrent(insatanceList []string) (err error) {
	return
}

func recoverWsAndSentMsg(conn *websocket.Conn, data interface{}) {
	if err := recover(); err != nil {
		stack := make([]byte, 1024*8)
		stack = stack[:runtime.Stack(stack, false)]
		resp := new(QueryInstanceSendMsg)
		resp.Status = "0001"
		resp.Msg = string(stack)
		senMsg(conn, resp)
	}
}
