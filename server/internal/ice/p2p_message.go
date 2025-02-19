package ice

import (
	"encoding/json"
	"fmt"
	"net"
)

// MessageType 消息类型
type MessageType string

const (
	// MsgRegister 注册消息
	MsgRegister MessageType = "register"
	// MsgConnect 请求连接对方
	MsgConnect MessageType = "connect"
	// MsgPunch 打洞消息
	MsgPunch MessageType = "punch"
	// MsgNATDetect NAT类型检测
	MsgNATDetect MessageType = "nat_detect"
)

// Message P2P消息结构
type Message struct {
	Type    MessageType    `json:"type"`
	Payload map[string]any `json:"payload"`
}

// RegisterPayload 注册消息负载
type RegisterPayload struct {
	PeerID    string `json:"peer_id"`
	LocalIP   string `json:"local_ip"`
	LocalPort int    `json:"local_port"`
}

// ConnectPayload 连接请求负载
type ConnectPayload struct {
	TargetID string `json:"target_id"`
}

// PunchPayload 打洞消息负载
type PunchPayload struct {
	PeerID     string `json:"peer_id"`
	PublicIP   string `json:"public_ip"`
	PublicPort int    `json:"public_port"`
	LocalIP    string `json:"local_ip"`
	LocalPort  int    `json:"local_port"`
}

// NATInfo NAT信息
type NATInfo struct {
	Type      string `json:"type"`      // NAT类型
	Symmetric bool   `json:"symmetric"` // 是否是对称NAT
}

// WriteMessage 发送消息
func WriteMessage(conn net.Conn, msgType MessageType, payload interface{}) error {
	msg := Message{
		Type:    msgType,
		Payload: make(map[string]any),
	}

	// 将payload转换为map
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化payload失败: %w", err)
	}

	if err := json.Unmarshal(data, &msg.Payload); err != nil {
		return fmt.Errorf("转换payload失败: %w", err)
	}

	// 发送消息
	if err := json.NewEncoder(conn).Encode(msg); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}

// ReadMessage 读取消息
func ReadMessage(conn net.Conn) (*Message, error) {
	var msg Message
	if err := json.NewDecoder(conn).Decode(&msg); err != nil {
		return nil, fmt.Errorf("读取消息失败: %w", err)
	}
	return &msg, nil
}

// ParsePayload 解析消息负载
func ParsePayload(msg *Message, payload interface{}) error {
	data, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("序列化payload失败: %w", err)
	}

	if err := json.Unmarshal(data, payload); err != nil {
		return fmt.Errorf("解析payload失败: %w", err)
	}

	return nil
}
