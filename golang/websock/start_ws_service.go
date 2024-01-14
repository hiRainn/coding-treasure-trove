package websock

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StartWebSock() {
	http.HandleFunc("/websock", handleWebSocket)
	http.ListenAndServe(":39076", nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	var conn *websocket.Conn
	var err error
	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	params := r.URL.Query()
	taskType := params.Get("task_type")

	switch taskType {
	case "instance_status":
		go getInstanceStatus(conn)
	}
}
