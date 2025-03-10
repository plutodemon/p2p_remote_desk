package signaling

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"syscall"

	"p2p_remote_desk/common"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/server/config"

	"github.com/coder/websocket"
	"github.com/plutodemon/llog"
)

type Client struct {
	Id   string
	Conn *websocket.Conn
}

type ClientsInfo struct {
	Clients   map[string]*Client // 已注册的客户端
	clientsMu sync.Mutex         // 客户端注册表锁
}

func (c *ClientsInfo) AddClient(client *Client) {
	c.clientsMu.Lock()
	c.Clients[client.Id] = client
	c.clientsMu.Unlock()
	llog.InfoF("客户端[ %s ]已注册", client.Id)
}

func (c *ClientsInfo) RemoveClient(clientID string) {
	c.clientsMu.Lock()
	delete(c.Clients, clientID)
	c.clientsMu.Unlock()
	llog.InfoF("客户端[ %s ]已注销", clientID)
}

func (c *ClientsInfo) GetClient(clientID string) (*Client, bool) {
	c.clientsMu.Lock()
	defer c.clientsMu.Unlock()
	if client, ok := c.Clients[clientID]; ok {
		return client, true
	}
	return nil, false
}

var SignalClients = &ClientsInfo{
	Clients:   make(map[string]*Client),
	clientsMu: sync.Mutex{},
}

func Start() {
	http.HandleFunc("/signaling", handleSignaling)

	cfg := config.GetConfig()
	add := lkit.GetAddr(cfg.Server.Host, cfg.Server.SignalPort)

	err := http.ListenAndServe(add, nil)
	if err != nil {
		llog.Error("start handleSignaling error:", err)
		lkit.SigChan <- syscall.SIGTERM
		return
	}

	llog.Info("信令服务器启动, 地址:", add)
}

func handleSignaling(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		llog.Warn("WebSocket升级失败:", err)
		return
	}
	defer conn.CloseNow()

	ctx := context.Background()
	var clientID string
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			llog.Warn("读取注册消息失败:", err)
			return
		}

		reg := common.SignalReg
		if err := json.Unmarshal(msg, &reg); err != nil {
			llog.Warn("解析注册消息失败:", err)
			continue
		}

		if reg.Action == common.SignalRegActionRegister && reg.ID != "" {
			clientID = reg.ID
			SignalClients.AddClient(&Client{Id: clientID, Conn: conn})
			break
		}
	}

	for {
		_, msgBytes, err := conn.Read(ctx)
		if err != nil {
			llog.WarnF("客户端[ %s ]断开连接:", clientID, err)
			return
		}

		var msg common.SignalMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			llog.Warn("解析消息失败:", err)
			continue
		}

		targetClient, exists := SignalClients.GetClient(msg.To)
		if exists {
			msgBytes, err = json.Marshal(msg)
			if err != nil {
				llog.Warn("序列化消息失败:", err)
				continue
			}
			if err := targetClient.Conn.Write(ctx, websocket.MessageText, msgBytes); err != nil {
				llog.WarnF("转发消息到[ %s ]失败:", msg.To, err)
			}
		} else {
			llog.Warn("目标客户端[ %s ]不存在", msg.To)
		}
	}
}
