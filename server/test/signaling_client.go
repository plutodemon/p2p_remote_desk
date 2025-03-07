package main

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/coder/websocket"
	"log"
	"os"
)

type SignalMessage struct {
	From    string      `json:"from"`    // 发送方ID
	To      string      `json:"to"`      // 接收方ID
	Type    string      `json:"type"`    // 消息类型（offer/answer/candidate）
	Payload interface{} `json:"payload"` // 消息内容（SDP或Candidate）
}

func main() {
	// 连接信令服务器
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, "ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("连接服务器失败:", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	// 注册客户端ID
	reader := bufio.NewReader(os.Stdin)
	log.Print("输入你的客户端ID: ")
	clientID, _ := reader.ReadString('\n')
	clientID = clientID[:len(clientID)-1] // 去除换行符

	// 发送注册消息
	regMsg, err := json.Marshal(map[string]string{
		"action": "register",
		"id":     clientID,
	})
	if err != nil {
		log.Fatal("序列化注册消息失败:", err)
	}
	if err := conn.Write(ctx, websocket.MessageText, regMsg); err != nil {
		log.Fatal("注册失败:", err)
	}

	// 接收消息的协程
	go func() {
		for {
			_, msgBytes, err := conn.Read(ctx)
			if err != nil {
				log.Fatal("读取消息失败:", err)
			}

			var msg SignalMessage
			if err := json.Unmarshal(msgBytes, &msg); err != nil {
				log.Printf("解析消息失败: %v", err)
				continue
			}
			log.Printf("收到来自 %s 的消息: Type=%s, Payload=%v", msg.From, msg.Type, msg.Payload)
		}
	}()

	// 发送消息的循环
	for {
		log.Print("输入目标客户端ID: ")
		to, _ := reader.ReadString('\n')
		to = to[:len(to)-1]

		log.Print("输入消息类型（offer/answer/candidate）: ")
		msgType, _ := reader.ReadString('\n')
		msgType = msgType[:len(msgType)-1]

		log.Print("输入消息内容（JSON）: ")
		payload, _ := reader.ReadString('\n')
		payload = payload[:len(payload)-1]

		// 构造消息
		msg := SignalMessage{
			From:    clientID,
			To:      to,
			Type:    msgType,
			Payload: payload,
		}
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			log.Fatal("序列化消息失败:", err)
		}
		if err := conn.Write(ctx, websocket.MessageText, msgBytes); err != nil {
			log.Fatal("发送消息失败:", err)
		}
	}
}
