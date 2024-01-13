package database

type QuerySlaveStatusResult struct {
	SecondsBehindMaster int    `gorm:"column:Seconds_Behind_Master"`
	SlaveIORunning      string `gorm:"column:Slave_IO_Running"`
	SlaveSQLRunning     string `gorm:"column:Slave_SQL_Running"`
	LastError           string `gorm:"column:Last_Error"`
	LastIOError         string `gorm:"column:Last_IO_Error"`
	LastSQLError        string `gorm:"column:Last_SQL_Error"`
}
