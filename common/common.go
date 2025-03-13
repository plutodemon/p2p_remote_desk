package common

type SignalMessageType = string

const (
	SignalMessageTypeRegister      SignalMessageType = "register"
	SignalMessageTypeGetClientList SignalMessageType = "getClientList"
)

type SignalMessage struct {
	From    string            `json:"from"`    // 发送方
	Type    SignalMessageType `json:"type"`    // 消息类型 SignalMessageType
	Message interface{}       `json:"message"` // 消息内容
}
