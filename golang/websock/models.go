package websock

type QueryInstanceInstanceRequest struct {
	InstanceMark []string `json:"instance_mark"` //format is ip_port
}

type QueryInstanceSendMsg struct {
	Status string          `json:"status"`
	Msg    string          `json:"msg"`
	Data   *InstanceStatus `json:"data"`
}
type InstanceStatus struct {
	InstanceMark        string `json:"instance_mark"`
	RunningStatus       bool   `json:"running_status"`
	SlaveStatus         bool   `json:"slave_status"`
	SecondsBehindMaster int    `json:"seconds_behind_master"`
	LastError           string `json:"last_error"`
}
