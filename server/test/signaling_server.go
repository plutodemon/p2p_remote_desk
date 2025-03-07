package main

import (
	"context"
	"encoding/json"
	"github.com/coder/websocket"
	"log"
	"net/http"
	"sync"
)

// 客户端注册表
type Client struct {
	conn *websocket.Conn
	id   string
}

var (
	clients   = make(map[string]*Client) // 已注册的客户端
	clientsMu sync.Mutex                 // 客户端注册表锁
)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("信令服务器启动，监听 :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type SignalMessageSer struct {
	From    string      `json:"from"`    // 发送方ID
	To      string      `json:"to"`      // 接收方ID
	Type    string      `json:"type"`    // 消息类型（offer/answer/candidate）
	Payload interface{} `json:"payload"` // 消息内容（SDP或Candidate）
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx := context.Background()

	// 等待客户端注册
	var clientID string
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			log.Printf("读取注册消息失败: %v", err)
			return
		}

		var reg struct {
			Action string `json:"action"`
			ID     string `json:"id"`
		}
		if err := json.Unmarshal(msg, &reg); err != nil {
			log.Printf("解析注册消息失败: %v", err)
			continue
		}

		if reg.Action == "register" && reg.ID != "" {
			clientID = reg.ID
			clientsMu.Lock()
			clients[clientID] = &Client{conn: conn, id: clientID}
			clientsMu.Unlock()
			log.Printf("客户端 %s 已注册", clientID)
			break
		}
	}

	// 监听客户端消息并转发
	for {
		_, msgBytes, err := conn.Read(ctx)
		if err != nil {
			log.Printf("客户端 %s 断开连接: %v", clientID, err)
			clientsMu.Lock()
			delete(clients, clientID)
			clientsMu.Unlock()
			return
		}

		var msg SignalMessageSer
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Printf("解析消息失败: %v", err)
			continue
		}

		// 转发消息到目标客户端
		clientsMu.Lock()
		targetClient, exists := clients[msg.To]
		clientsMu.Unlock()
		if exists {
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Printf("序列化消息失败: %v", err)
				continue
			}
			if err := targetClient.conn.Write(ctx, websocket.MessageText, msgBytes); err != nil {
				log.Printf("转发消息到 %s 失败: %v", msg.To, err)
			}
		} else {
			log.Printf("目标客户端 %s 不存在", msg.To)
		}
	}
}
