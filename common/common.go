package common

type SignalMessage struct {
	From    string      `json:"from"`    // 发送方ID
	To      string      `json:"to"`      // 接收方ID
	Type    string      `json:"type"`    // 消息类型（offer/answer/candidate）
	Payload interface{} `json:"payload"` // 消息内容（SDP或Candidate）
}

var SignalReg struct {
	Action string `json:"action"`
	ID     string `json:"id"`
}

type SignalRegAction = string

const (
	SignalRegActionRegister SignalRegAction = "register"
)
