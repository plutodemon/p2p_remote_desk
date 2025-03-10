package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"p2p_remote_desk/common"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/server/config"

	"github.com/coder/websocket"
	"github.com/panjf2000/ants/v2"
	"github.com/plutodemon/llog"
)

type Client struct {
	Id             string
	Conn           *websocket.Conn
	LastActiveTime time.Time
}

// 服务器配置参数
type ServerConfig struct {
	MaxConnections    int           // 最大并发连接数
	IdleTimeout       time.Duration // 连接空闲超时时间
	GoroutinePoolSize int           // goroutine池大小
	MessageBufferSize int           // 消息缓冲区大小
	CleanupInterval   time.Duration // 清理间隔时间
}

type ClientsInfo struct {
	Clients     map[string]*Client // 已注册的客户端
	clientsMu   sync.RWMutex       // 客户端注册表锁
	activeConns int32              // 当前活跃连接数
	config      ServerConfig       // 服务器配置
	messagePool sync.Pool          // 消息对象池
}

func (c *ClientsInfo) AddClient(client *Client) bool {
	// 检查是否超过最大连接数
	if c.config.MaxConnections > 0 && atomic.LoadInt32(&c.activeConns) >= int32(c.config.MaxConnections) {
		llog.Warn("达到最大连接数限制，拒绝新连接")
		return false
	}

	// 更新活跃时间
	client.LastActiveTime = time.Now()

	c.clientsMu.Lock()
	c.Clients[client.Id] = client
	c.clientsMu.Unlock()

	// 增加活跃连接计数
	atomic.AddInt32(&c.activeConns, 1)
	llog.InfoF("客户端[ %s ]已注册，当前连接数: %d", client.Id, atomic.LoadInt32(&c.activeConns))
	return true
}

func (c *ClientsInfo) RemoveClient(clientID string) {
	c.clientsMu.Lock()
	_, exists := c.Clients[clientID]
	if exists {
		delete(c.Clients, clientID)
		c.clientsMu.Unlock()

		// 减少活跃连接计数
		atomic.AddInt32(&c.activeConns, -1)
		llog.InfoF("客户端[ %s ]已注销，当前连接数: %d", clientID, atomic.LoadInt32(&c.activeConns))
	} else {
		c.clientsMu.Unlock()
	}
}

func (c *ClientsInfo) GetClient(clientID string) (*Client, bool) {
	c.clientsMu.RLock()
	client, ok := c.Clients[clientID]
	c.clientsMu.RUnlock()

	if ok {
		// 更新活跃时间
		client.LastActiveTime = time.Now()
		return client, true
	}
	return nil, false
}

// 创建消息对象的工厂函数
func createSignalMessage() interface{} {
	return new(common.SignalMessage)
}

var SignalClients = &ClientsInfo{
	Clients:     make(map[string]*Client),
	clientsMu:   sync.RWMutex{},
	activeConns: 0,
	config: ServerConfig{
		MaxConnections:    1000,             // 最大允许1000个并发连接
		IdleTimeout:       5 * time.Minute,  // 5分钟不活跃则断开
		GoroutinePoolSize: 500,              // 工作协程池大小
		MessageBufferSize: 100,              // 每个连接的消息缓冲区大小
		CleanupInterval:   30 * time.Second, // 每30秒清理一次不活跃连接
	},
	messagePool: sync.Pool{
		New: createSignalMessage,
	},
}

// 定期清理不活跃的连接
func startCleanupRoutine() {
	ticker := time.NewTicker(SignalClients.config.CleanupInterval)
	go func() {
		for range ticker.C {
			cleanupInactiveConnections()
		}
	}()
}

func cleanupInactiveConnections() {
	now := time.Now()
	var inactiveClients []string

	// 使用读锁查找不活跃的连接
	SignalClients.clientsMu.RLock()
	for id, client := range SignalClients.Clients {
		if now.Sub(client.LastActiveTime) > SignalClients.config.IdleTimeout {
			inactiveClients = append(inactiveClients, id)
		}
	}
	SignalClients.clientsMu.RUnlock()

	// 移除不活跃的连接
	for _, id := range inactiveClients {
		llog.InfoF("移除不活跃客户端[ %s ]", id)
		SignalClients.RemoveClient(id)
	}

	llog.InfoF("清理完成，当前活跃连接数: %d", atomic.LoadInt32(&SignalClients.activeConns))
}

func main() {
	// 初始化goroutine池
	pool, err := ants.NewPool(SignalClients.config.GoroutinePoolSize)
	if err != nil {
		llog.Error("创建goroutine池失败:", err)
		lkit.SigChan <- syscall.SIGTERM
		return
	}
	defer pool.Release()

	// 启动定期清理不活跃连接的协程
	startCleanupRoutine()

	http.HandleFunc("/signaling", func(w http.ResponseWriter, r *http.Request) {
		// 使用协程池处理新连接
		pool.Submit(func() {
			handleSignaling(w, r)
		})
	})

	cfg := config.GetConfig()
	add := lkit.GetAddr(cfg.Server.Host, cfg.Server.SignalPort)

	llog.Info("信令服务器启动, 地址:", add, "，最大连接数:", SignalClients.config.MaxConnections)

	err = http.ListenAndServe(add, nil)
	if err != nil {
		llog.Error("start handleSignaling error:", err)
		lkit.SigChan <- syscall.SIGTERM
		return
	}
}

func handleSignaling(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		llog.Warn("WebSocket升级失败:", err)
		return
	}
	defer conn.CloseNow()

	// 创建一个带30秒超时的上下文用于客户端注册
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var clientID string
	registerChan := make(chan bool)

	// 启动一个goroutine处理注册
	go func() {
		for {
			_, msg, err := conn.Read(ctx)
			if err != nil {
				llog.Warn("读取注册消息失败:", err)
				registerChan <- false
				return
			}

			reg := common.SignalReg
			if err := json.Unmarshal(msg, &reg); err != nil {
				llog.Warn("解析注册消息失败:", err)
				continue
			}

			if reg.Action == common.SignalRegActionRegister && reg.ID != "" {
				clientID = reg.ID
				// 使用改进的AddClient方法，检查连接数限制
				if !SignalClients.AddClient(&Client{Id: clientID, Conn: conn, LastActiveTime: time.Now()}) {
					llog.WarnF("客户端[ %s ]注册失败: 达到最大连接数限制", clientID)
					registerChan <- false
					return
				}
				registerChan <- true
				return
			}
		}
	}()

	// 等待注册完成或超时
	select {
	case success := <-registerChan:
		if !success {
			return
		}
		llog.InfoF("客户端[ %s ]注册成功", clientID)
	case <-ctx.Done():
		llog.Warn("客户端注册超时")
		return
	}

	// 注册成功后，确保客户端断开连接时移除注册信息
	defer SignalClients.RemoveClient(clientID)

	// 重置为无超时的上下文用于后续通信
	ctx = context.Background()

	// 创建消息缓冲通道，限制消息处理速率
	msgChan := make(chan []byte, SignalClients.config.MessageBufferSize)

	// 启动消息处理协程
	go func() {
		for msgBytes := range msgChan {
			// 从对象池获取消息对象
			msgObj := SignalClients.messagePool.Get().(*common.SignalMessage)

			if err := json.Unmarshal(msgBytes, msgObj); err != nil {
				llog.Warn("解析消息失败:", err)
				// 将对象放回池中
				SignalClients.messagePool.Put(msgObj)
				continue
			}

			targetClient, exists := SignalClients.GetClient(msgObj.To)
			if exists {
				// 直接转发原始消息，避免重新序列化
				if err := targetClient.Conn.Write(ctx, websocket.MessageText, msgBytes); err != nil {
					llog.WarnF("转发消息到[ %s ]失败: %v", msgObj.To, err)
				}
			} else {
				llog.WarnF("目标客户端[ %s ]不存在", msgObj.To)
			}

			// 将对象放回池中
			SignalClients.messagePool.Put(msgObj)
		}
	}()

	// 主循环读取消息并发送到处理通道
	for {
		_, msgBytes, err := conn.Read(ctx)
		if err != nil {
			llog.WarnF("客户端[ %s ]断开连接: %v", clientID, err)
			close(msgChan) // 关闭消息通道
			return
		}

		// 更新客户端活跃时间
		if client, exists := SignalClients.GetClient(clientID); exists {
			client.LastActiveTime = time.Now()
		}

		// 将消息发送到处理通道
		select {
		case msgChan <- msgBytes:
			// 消息成功加入队列
		default:
			// 队列已满，丢弃消息
			llog.WarnF("客户端[ %s ]消息队列已满，丢弃消息", clientID)
		}
	}
}
