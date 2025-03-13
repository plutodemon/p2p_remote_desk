package network

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/common"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"

	"github.com/coder/websocket"
)

var clientID string
var ctx context.Context
var wsConn *websocket.Conn

// ConnectSignalingServer 连接信令服务器
func ConnectSignalingServer() error {
	ctx = context.Background()
	cfg := config.GetConfig().ServerConfig
	addr := lkit.GetAddr(cfg.Address, cfg.SignalPort)
	url := "ws://" + addr + "/signaling"

	name, err := os.Hostname()
	if err != nil {
		return err
	}
	clientID = name

	wsConn, _, err = websocket.Dial(ctx, url, nil)
	if err != nil {
		llog.Warn("连接信令服务器失败", "url:", url, "error:", err)
		return err
	}
	defer func() {
		_ = wsConn.CloseNow()
	}()

	// 注册客户端
	if err := sendMessage(common.SignalMessageTypeRegister, nil); err != nil {
		llog.Warn("注册失败:", err)
		return err
	}
	llog.InfoF("客户端 %s 注册成功", clientID)

	for {
		_, msgBytes, err := wsConn.Read(ctx)
		if err != nil {
			llog.Warn("读取消息失败:", err)
			return err
		}

		var msg common.SignalMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			llog.Warn("解析消息失败:", err)
			continue
		}

		// 处理消息
		log.Printf("收到来自 %s 的消息: Type=%s, Payload=%v", msg.From, msg.Type, msg.Message)
	}
}

func sendMessage(messageType common.SignalMessageType, message interface{}) error {
	if ctx == nil || wsConn == nil {
		return errors.New("ctx or wsConn is nil")
	}

	regMsg, _ := json.Marshal(common.SignalMessage{
		From:    clientID,
		Type:    messageType,
		Message: message,
	})

	return wsConn.Write(ctx, websocket.MessageText, regMsg)
}
