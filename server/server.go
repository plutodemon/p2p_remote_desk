package main

import (
	"fmt"
	"github.com/plutodemon/llog"
	"os"
	"p2p_remote_desk/server/config"
	"p2p_remote_desk/server/internal"
)

func main() {
	// 设置全局panic处理
	defer llog.HandlePanic()

	// 初始化配置
	if err := config.Init(); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	if err := llog.Init(config.GetConfig().LogConfig); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}
	defer llog.Cleanup()

	// 创建并启动服务器
	server := internal.NewServer()
	if err := server.Start(); err != nil {
		llog.Error("服务器启动失败: %v", err)
		os.Exit(1)
	}
}
