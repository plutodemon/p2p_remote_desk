package network

import (
	"p2p_remote_desk/common"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"
)

var Connected = make(chan bool)

func StartNetWorkClient() {
	lkit.SafeGo(func() {
		// 连接信令服务器
		if err := ConnectSignalingServer(); err != nil {
			Connected <- false
			llog.Warn("连接信令服务器失败:", err)
			return
		}
	})

	select {
	case isConn := <-Connected:
		if !isConn {
			return
		}
		SendMessage <- SendMessStr{MsgType: common.SignalMessageTypeRegister}
		SendMessage <- SendMessStr{MsgType: common.SignalMessageTypeGetClientList}
		llog.InfoF("客户端 %s 注册成功", ClientName)
	}
}
