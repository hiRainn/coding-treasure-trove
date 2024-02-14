package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type MySQLConn struct {
	DB       *gorm.DB
	Username string
	Password string
	Host     string
	Port     int
	DBName   string
}

func (cfg *MySQLConn) NewConn() (err error) {
	dsn := fmt.Sprintf(`%s:%s@tcp(%s:%d)/?charset=utf8&parseTime=True&loc=Local&readTimeout=30s&multiStatements=true`, cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // Slow SQL threshold
			LogLevel:      logger.Info, // Log level
			Colorful:      false,       // Disable color
		},
	)
	cfg.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:         newLogger,
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})

	db, _ := cfg.DB.DB()
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Second * 60)
	return
}

//close mysql connect
func (cfg *MySQLConn) Close() {
	sqlDB, err := cfg.DB.DB()
	if err != nil {
		return
	}
	sqlDB.Close()
}

func (cfg *MySQLConn) QuerySlaveStatus() (data *QuerySlaveStatusResult, err error) {
	data = new(QuerySlaveStatusResult)
	if err = cfg.DB.Raw("show slave status").Scan(data).Error; err != nil {
		err = fmt.Errorf("get slave status info error: %s", err.Error())
	}
	if data.LastError != "" || data.LastIOError != "" || data.LastSQLError != "" {
		var sb strings.Builder
		if data.LastError != "" {
			sb.WriteString(fmt.Sprintf("Last_Error:%s", data.LastError))
		}
		if data.LastIOError != "" {
			if sb.Len() > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("Last_IO_Error:%s", data.LastIOError))
		}
		if data.LastSQLError != "" {
			if sb.Len() > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("Last_SQL_Error:%s", data.LastSQLError))
		}
		data.LastError = sb.String()
	}

	return
}
