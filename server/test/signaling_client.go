package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/coder/websocket"
)

type SignalMessage struct {
	From    string      `json:"from"`    // 发送方ID
	To      string      `json:"to"`      // 接收方ID
	Type    string      `json:"type"`    // 消息类型（offer/answer/candidate）
	Payload interface{} `json:"payload"` // 消息内容（SDP或Candidate）
}

// 注册客户端并返回客户端ID
func registerClient(ctx context.Context, conn *websocket.Conn, reader *bufio.Reader) (string, error) {
	registerDone := make(chan struct {
		clientID string
		err      error
	})

	go func() {
		// 获取客户端ID
		log.Print("输入你的客户端ID: ")
		clientID, err := reader.ReadString('\n')
		if err != nil {
			registerDone <- struct {
				clientID string
				err      error
			}{clientID: "", err: err}
			return
		}
		clientID = clientID[:len(clientID)-1] // 去除换行符

		// 发送注册消息
		regMsg, err := json.Marshal(map[string]string{
			"action": "register",
			"id":     clientID,
		})
		if err != nil {
			registerDone <- struct {
				clientID string
				err      error
			}{clientID: "", err: err}
			return
		}

		if err := conn.Write(ctx, websocket.MessageText, regMsg); err != nil {
			registerDone <- struct {
				clientID string
				err      error
			}{clientID: "", err: err}
			return
		}

		registerDone <- struct {
			clientID string
			err      error
		}{clientID: clientID, err: nil}
	}()

	result := <-registerDone
	return result.clientID, result.err
}

// 启动消息接收协程
func startReceiveMessages(ctx context.Context, conn *websocket.Conn) {
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
}

// 启动消息发送协程
func startSendMessages(ctx context.Context, conn *websocket.Conn, reader *bufio.Reader, clientID string) {
	go func() {
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
	}()
}

func main() {
	// 连接信令服务器
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, "ws://localhost:8081/ws", nil)
	if err != nil {
		log.Fatal("连接服务器失败:", err)
	}
	defer conn.CloseNow()

	// 创建读取器
	reader := bufio.NewReader(os.Stdin)

	// 注册客户端
	clientID, err := registerClient(ctx, conn, reader)
	if err != nil {
		log.Fatal("注册失败:", err)
	}
	log.Printf("客户端 %s 注册成功", clientID)

	// 启动消息接收协程
	startReceiveMessages(ctx, conn)

	// 启动消息发送协程
	startSendMessages(ctx, conn, reader, clientID)

	// 主线程等待，防止程序退出
	select {}
}
