package network

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/common"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"

	"github.com/coder/websocket"
	"github.com/denisbrodbeck/machineid"
)

var (
	ClientUID  string          // 设备唯一标识
	ClientName string          // 设备名称(可修改备注)
	UserToken  string          // 设备token
	ctx        context.Context // 上下文
	wsConn     *websocket.Conn // websocket连接
	Clients    = sync.Map{}    // 客户端列表
)

func initClient() error {
	var err error
	ClientName, err = os.Hostname()
	UserToken = ClientName
	if err != nil {
		return err
	}
	ClientUID, err = machineid.ID()
	return err
}

// ConnectSignalingServer 连接信令服务器
func ConnectSignalingServer() error {
	var err error

	if err = initClient(); err != nil {
		llog.Warn("初始化客户端失败", "error:", err)
		return err
	}

	ctx = context.Background()
	cfg := config.GetConfig().ServerConfig
	addr := lkit.GetAddr(cfg.Address, cfg.SignalPort)
	url := "ws://" + addr + "/signaling"

	wsConn, _, err = websocket.Dial(ctx, url, nil)
	if err != nil {
		llog.Warn("连接信令服务器失败", "url:", url, "error:", err)
		return err
	}
	defer func() {
		_ = wsConn.CloseNow()
	}()

	afterErr := make(chan error)
	go func() {
		readMessage(afterErr)
	}()
	go func() {
		afterConnectSignalingServer(afterErr)
	}()

	select {
	case e := <-afterErr:
		return e
	}
}

func readMessage(afterErr chan error) {
	for {
		_, msgBytes, err := wsConn.Read(ctx)
		if err != nil {
			llog.Warn("读取消息失败:", err)
			afterErr <- err
			return
		}

		var msg common.SignalMessage
		if err = json.Unmarshal(msgBytes, &msg); err != nil {
			llog.Warn("解析消息失败:", err)
			continue
		}
		UserToken = msg.Sender.Token

		switch msg.Type {
		case common.SignalMessageTypeGetClientList:
			clients := make([]common.ClientInfo, 0)
			err = msg.GetMessage(common.SignalMessageTypeGetClientList, &clients)
			if err != nil {
				llog.Warn("解析客户端列表失败:", err)
				continue
			}
			for _, c := range clients {
				Clients.Store(c.Id, c)
			}
		}
	}
}

func afterConnectSignalingServer(afterErr chan error) {
	if err := sendMessage(common.SignalMessageTypeRegister, nil); err != nil {
		llog.Warn("注册失败:", err)
		afterErr <- err
		return
	}
	llog.InfoF("客户端 %s 注册成功", ClientName)

	// 获取列表
	if err := sendMessage(common.SignalMessageTypeGetClientList, nil); err != nil {
		llog.Warn("获取列表失败:", err)
		afterErr <- err
		return
	}
}

func sendMessage(messageType common.SignalMessageType, message interface{}) error {
	if ctx == nil || wsConn == nil {
		return errors.New("ctx or wsConn is nil")
	}

	sender := &common.MessageSender{
		From:  common.SignalMessageSenderTypeClient,
		UID:   ClientUID,
		Token: UserToken,
	}
	msg, err := common.CreateSignalMessage(sender, messageType, message)
	if err != nil {
		return err
	}

	regMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return wsConn.Write(ctx, websocket.MessageText, regMsg)
}
