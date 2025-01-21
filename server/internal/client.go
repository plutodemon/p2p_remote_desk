package internal

import (
	"encoding/json"
	"fmt"
	"github.com/plutodemon/llog"
	"net"
	"p2p_remote_desk/server/config"
	"sync"
	"time"
)

// Client 客户端连接结构体
type Client struct {
	ID        string
	conn      net.Conn
	server    *Server
	sendChan  chan []byte
	closeChan chan struct{}
	closeOnce sync.Once
}

// NewClient 创建新的客户端连接
func NewClient(conn net.Conn, server *Server) *Client {
	return &Client{
		ID:        conn.RemoteAddr().String(),
		conn:      conn,
		server:    server,
		sendChan:  make(chan []byte, config.GetConfig().Performance.BufferSize),
		closeChan: make(chan struct{}),
	}
}

// Handle 处理客户端连接
func (c *Client) Handle() {
	defer func() {
		c.server.removeClient(c.ID)
		c.Close()
	}()

	// 启动发送协程
	go c.sendLoop()

	// 读取循环
	buffer := make([]byte, config.GetConfig().Performance.BufferSize)
	for {
		select {
		case <-c.closeChan:
			return
		default:
			// 设置读取超时
			_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * 30))
			n, err := c.conn.Read(buffer)
			if err != nil {
				llog.Error("读取客户端数据失败: %v", err)
				return
			}

			// 处理接收到的数据
			if err := c.handleMessage(buffer[:n]); err != nil {
				llog.Error("处理客户端消息失败: %v", err)
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(data []byte) error {
	var message struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	switch message.Type {
	case "ping":
		return c.handlePing()
	// 添加其他消息类型的处理
	default:
		return fmt.Errorf("未知的消息类型: %s", message.Type)
	}
}

// handlePing 处理ping消息
func (c *Client) handlePing() error {
	response := struct {
		Type    string `json:"type"`
		Payload struct {
			Time int64 `json:"time"`
		} `json:"payload"`
	}{
		Type: "pong",
		Payload: struct {
			Time int64 `json:"time"`
		}{
			Time: time.Now().UnixNano() / 1e6,
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("生成响应失败: %w", err)
	}

	c.SendMessage(data)
	return nil
}

// SendMessage 发送消息
func (c *Client) SendMessage(message []byte) {
	select {
	case c.sendChan <- message:
	case <-c.closeChan:
	default:
		llog.Warn("客户端发送缓冲区已满: %s", c.ID)
	}
}

// sendLoop 发送循环
func (c *Client) sendLoop() {
	for {
		select {
		case message := <-c.sendChan:
			if err := c.writeMessage(message); err != nil {
				llog.Error("发送消息失败: %v", err)
				c.Close()
				return
			}
		case <-c.closeChan:
			return
		}
	}
}

// writeMessage 写入消息
func (c *Client) writeMessage(message []byte) error {
	_ = c.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
	_, err := c.conn.Write(message)
	return err
}

// Close 关闭客户端连接
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.closeChan)
		_ = c.conn.Close()
		llog.Info("客户端连接已关闭: %s", c.ID)
	})
}
