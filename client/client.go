package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/plutodemon/slog"
	"os"
	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal"
)

func main() {
	// 设置全局panic处理
	defer slog.HandlePanic()

	// 初始化配置
	if err := config.Init(); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	if err := slog.Init(config.GetConfig().LogConfig); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}
	defer slog.Cleanup()

	if err := glfw.Init(); err != nil {
		slog.Error("初始化glfw失败: %v", err)
		os.Exit(1)
	}
	defer glfw.Terminate()

	// 初始化并运行应用
	newApp := internal.NewApp()
	newApp.Run()
}
