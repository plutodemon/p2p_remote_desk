package network

import (
	"p2p_remote_desk/common"
	"p2p_remote_desk/llog"
)

func StartNetWorkClient() {
	go func() {
		defer llog.HandlePanic()

		// 连接信令服务器
		if err := ConnectSignalingServer(); err != nil {
			llog.Warn("连接信令服务器失败:", err)
			return
		}
	}()
	SendMessage <- SendMessStr{MsgType: common.SignalMessageTypeRegister}
	SendMessage <- SendMessStr{MsgType: common.SignalMessageTypeGetClientList}
	llog.InfoF("客户端 %s 注册成功", ClientName)
}
