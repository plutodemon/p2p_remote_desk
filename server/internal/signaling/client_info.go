package signaling

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"
	"p2p_remote_desk/server/config"

	"github.com/coder/websocket"
)

type Client struct {
	Id             string
	LastActiveTime int64
	Conn           *websocket.Conn
	Mu             sync.Mutex
}

func NewClient(clientID string, conn *websocket.Conn) *Client {
	return &Client{
		Id:             clientID,
		LastActiveTime: lkit.GetNowUnix(),
		Conn:           conn,
		Mu:             sync.Mutex{},
	}
}

func (c *Client) GetLastActiveTime() int64 {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return c.LastActiveTime
}

func (c *Client) Write(ctx context.Context, messageType websocket.MessageType, data []byte) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return c.Conn.Write(ctx, messageType, data)
}

func (c *Client) UpdateInfo(conn *websocket.Conn) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if conn != nil {
		c.Conn = conn
	}
	c.LastActiveTime = lkit.GetNowUnix()
}

type ClientsInfo struct {
	Clients     map[string]*Client // 已注册的客户端
	clientsMu   sync.RWMutex       // 客户端注册表锁
	activeConns int32              // 当前活跃连接数
	messagePool sync.Pool          // 消息对象池
}

func (c *ClientsInfo) AddClient(client *Client) bool {
	// 检查是否超过最大连接数
	cfg := config.GetConfig().Server
	if cfg.MaxConnections > 0 && atomic.LoadInt32(&c.activeConns) >= int32(cfg.MaxConnections) {
		llog.Warn("达到最大连接数限制，拒绝新连接")
		return false
	}

	c.clientsMu.Lock()
	c.Clients[client.Id] = client
	c.clientsMu.Unlock()

	// 增加活跃连接计数
	atomic.AddInt32(&c.activeConns, 1)
	llog.InfoF("客户端 %s 已注册，当前连接数: %d", client.Id, atomic.LoadInt32(&c.activeConns))
	return true
}

func (c *ClientsInfo) RemoveClient(clientID string) {
	c.clientsMu.Lock()
	defer c.clientsMu.Unlock()
	if _, exists := c.Clients[clientID]; exists {
		delete(c.Clients, clientID)
		// 减少活跃连接计数
		atomic.AddInt32(&c.activeConns, -1)
		llog.InfoF("客户端 %s 已注销，当前连接数: %d", clientID, atomic.LoadInt32(&c.activeConns))
	}
}

func (c *ClientsInfo) GetClient(clientID string) (*Client, bool) {
	c.clientsMu.RLock()
	client, ok := c.Clients[clientID]
	c.clientsMu.RUnlock()

	if ok {
		client.UpdateInfo(nil)
		return client, true
	}
	return nil, false
}

func (c *ClientsInfo) ClientRange(f func(client *Client) bool) {
	c.clientsMu.RLock()
	defer c.clientsMu.RUnlock()
	for _, client := range c.Clients {
		if !f(client) {
			break
		}
	}
}

// 定期清理不活跃的连接
func startCleanupRoutine() {
	cfg := config.GetConfig().Server
	ticker := time.NewTicker(time.Duration(cfg.CleanupInterval) * time.Second)
	go func() {
		for range ticker.C {
			cleanupInactiveConnections()
		}
	}()
}

func cleanupInactiveConnections() {
	now := lkit.GetNowUnix()
	cfg := config.GetConfig().Server
	var inactiveClients []string

	// 使用读锁查找不活跃的连接
	SignalClients.clientsMu.RLock()
	for id, client := range SignalClients.Clients {
		if now-client.GetLastActiveTime() > cfg.IdleTimeout {
			inactiveClients = append(inactiveClients, id)
		}
	}
	SignalClients.clientsMu.RUnlock()

	// 移除不活跃的连接
	for _, id := range inactiveClients {
		SignalClients.RemoveClient(id)
	}

	llog.InfoF("清理完成，当前活跃连接数: %d", atomic.LoadInt32(&SignalClients.activeConns))
}
