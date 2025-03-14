package signaling

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"p2p_remote_desk/common"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"
	"p2p_remote_desk/server/config"

	"github.com/coder/websocket"
	"github.com/panjf2000/ants/v2"
)

var Pool *ants.Pool

var SignalClients = &ClientsInfo{
	Clients:   make(map[string]*Client),
	clientsMu: sync.RWMutex{},
	messagePool: sync.Pool{
		New: func() interface{} {
			return new(common.SignalMessage)
		},
	},
}

func Start() {
	cfg := config.GetConfig().Server

	poolOptions := ants.Options{
		ExpiryDuration:   10 * time.Minute, // 空闲worker的过期时间
		PreAlloc:         true,             // 预分配goroutine队列内存
		MaxBlockingTasks: 1000,             // 最大阻塞任务数
		Nonblocking:      false,            // 设置为true时，当池满时Submit会返回ErrPoolOverload错误
		PanicHandler: func(p interface{}) {
			llog.Error("协程池处理任务时发生panic:", p)
		},
	}

	var err error
	Pool, err = ants.NewPool(cfg.GoroutinePoolSize, ants.WithOptions(poolOptions))
	if err != nil {
		llog.Error("创建goroutine池失败:", err)
		lkit.SigChan <- syscall.SIGTERM
		return
	}
	defer Pool.Release()

	// 启动定期清理不活跃连接的协程
	startCleanupRoutine()

	http.HandleFunc("/signaling", handleSignaling)

	addr := lkit.GetAddr(cfg.Host, cfg.SignalPort)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		llog.Error("start handleSignaling error:", err)
		lkit.SigChan <- syscall.SIGTERM
		return
	}

	llog.Info("信令服务器启动, 地址:", addr)
}

func handleSignaling(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig().Server
	// 检查当前连接数是否超过限制
	if cfg.MaxConnections > 0 && atomic.LoadInt32(&SignalClients.activeConns) >= int32(cfg.MaxConnections) {
		llog.Warn("达到最大连接数限制，拒绝新连接")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// 使用协程池处理WebSocket连接
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		llog.Warn("WebSocket升级失败:", err)
		return
	}

	// 提交任务到协程池，使用重试机制
	var submitted bool
	for retries := 0; retries < 3; retries++ {
		err = Pool.Submit(func() {
			handleWebSocketConn(conn)
		})

		if err == nil {
			submitted = true
			break
		}

		if errors.Is(err, ants.ErrPoolOverload) {
			llog.Warn("协程池已满，等待重试...")
			time.Sleep(100 * time.Millisecond) // 短暂等待后重试
		} else {
			break // 其他错误直接退出
		}
	}

	if submitted {
		return
	}

	llog.Warn("提交任务到协程池失败:", err)
	_ = conn.CloseNow()
}

// 处理WebSocket连接的函数，由协程池调用
func handleWebSocketConn(conn *websocket.Conn) {
	defer func() {
		err := conn.CloseNow()
		if err != nil {
			llog.Warn("关闭连接错误:", err)
		}
	}()

	// 创建一个带30秒超时的上下文用于客户端注册
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var clientID string
	registerChan := make(chan bool)

	// 启动一个goroutine处理注册
	dealRegisterMessage(ctx, conn, registerChan, &clientID)

	// 等待注册完成或超时
	select {
	case success := <-registerChan:
		if !success {
			return
		}
	case <-ctx.Done():
		llog.Warn("客户端注册超时")
		return
	}

	// 注册成功后，确保客户端断开连接时移除注册信息
	defer SignalClients.RemoveClient(clientID)

	// 重置为无超时的上下文用于后续通信
	ctx = context.Background()

	// 创建消息缓冲通道，限制消息处理速率
	cfg := config.GetConfig().Server
	msgChan := make(chan []byte, cfg.MessageBufferSize)

	// 启动消息处理协程
	dealMessage(ctx, msgChan)

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
			client.UpdateInfo(nil)
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

func dealRegisterMessage(ctx context.Context, conn *websocket.Conn, registerChan chan bool, clientID *string) {
	go func() {
		for {
			_, msg, err := conn.Read(ctx)
			if err != nil {
				llog.Warn("读取注册消息失败:", err)
				registerChan <- false
				return
			}

			var message common.SignalMessage
			if err := json.Unmarshal(msg, &message); err != nil {
				llog.Warn("解析注册消息失败:", err)
				continue
			}

			if message.Type == common.SignalMessageTypeRegister && message.From != "" {
				*clientID = message.From
				if !SignalClients.AddClient(NewClient(message.From, conn)) {
					llog.WarnF("客户端[ %s ]注册失败: 达到最大连接数限制", message.From)
					registerChan <- false
					return
				}
				registerChan <- true
				return
			}
		}
	}()
}

func dealMessage(ctx context.Context, msgChan chan []byte) {
	go func() {
		for msgBytes := range msgChan {
			// 从对象池获取消息对象
			message := SignalClients.messagePool.Get().(*common.SignalMessage)

			if err := json.Unmarshal(msgBytes, message); err != nil {
				llog.Warn("解析消息失败:", err)
				// 将对象放回池中
				SignalClients.messagePool.Put(message)
				continue
			}

			// todo 这里根据消息类型处理消息
			switch message.Type {
			case common.SignalMessageTypeGetClientList:
				ret := make([]common.ClientInfo, 0)
				SignalClients.clientsMu.RLock()
				for _, client := range SignalClients.Clients {
					ret = append(ret, common.ClientInfo{
						Id: client.Id,
						IP: 123123123,
					})
				}
				SignalClients.clientsMu.RUnlock()
				msg, _ := json.Marshal(common.SignalMessage{
					From:    "server",
					Type:    common.SignalMessageTypeGetClientList,
					Message: ret,
				})
				targetClient, _ := SignalClients.GetClient(message.From)
				if err := targetClient.Write(ctx, websocket.MessageText, msg); err != nil {
					llog.WarnF("转发消息到[ %s ]失败: %v", message.From, err)
				}

			default:

			}

			// 将对象放回池中
			SignalClients.messagePool.Put(message)
		}
	}()
}
