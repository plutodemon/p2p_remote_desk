package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type SignalMessageType = string

const (
	SignalMessageTypeRegister      SignalMessageType = "register"
	SignalMessageTypeGetClientList SignalMessageType = "getClientList"
)

type SignalMessage struct {
	From    string            `json:"from"`    // 发送方
	Type    SignalMessageType `json:"type"`    // 消息类型 SignalMessageType
	Message json.RawMessage   `json:"message"` // 消息内容，使用json.RawMessage替代interface{}
}

// CreateSignalMessage 创建信令消息
func CreateSignalMessage(from string, msgType SignalMessageType, message interface{}) (*SignalMessage, error) {
	var rawMsg json.RawMessage
	var err error

	if message != nil {
		rawMsg, err = json.Marshal(message)
		if err != nil {
			return nil, err
		}
	} else {
		// 空消息使用空JSON对象
		rawMsg = json.RawMessage([]byte(`{}`))
	}

	return &SignalMessage{
		From:    from,
		Type:    msgType,
		Message: rawMsg,
	}, nil
}

// GetMessage 通用的消息获取方法，可以获取任意类型的消息
// messageType: 期望的消息类型
// result: 用于存储解析结果的指针，必须是指针类型
func (s *SignalMessage) GetMessage(messageType SignalMessageType, result interface{}) error {
	// 检查消息类型是否匹配
	if s.Type != messageType {
		return fmt.Errorf("message type mismatch: expected %s, got %s", messageType, s.Type)
	}

	// 检查result是否为指针类型
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() != reflect.Ptr || resultValue.IsNil() {
		return errors.New("result must be a non-nil pointer")
	}

	// 解析消息内容
	if err := json.Unmarshal(s.Message, result); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}

type ClientInfo struct {
	Id string `json:"id"`
	IP uint32 `json:"ip"`
}
