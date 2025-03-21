package main

import (
	"fmt"
	"os"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal"
	"p2p_remote_desk/llog"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
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

	// 设置全局panic处理
	defer llog.HandlePanic()

	if err := glfw.Init(); err != nil {
		llog.Fatal("初始化glfw失败: ", err)
	}
	defer glfw.Terminate()

	internal.NewAppAndRun()
}
