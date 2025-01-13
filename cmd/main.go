package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.3/glfw"
	"os"
	"p2p_remote_desk/config"
	"p2p_remote_desk/internal"
	"p2p_remote_desk/logger"
)

func main() {
	// 设置全局panic处理
	defer logger.HandlePanic()

	// 初始化配置
	if err := config.Init(); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	if err := logger.Init(); err != nil {
		fmt.Printf("初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Cleanup()

	if err := glfw.Init(); err != nil {
		logger.Error("初始化glfw失败: %v", err)
		os.Exit(1)
	}
	defer glfw.Terminate()

	// 初始化并运行应用
	newApp := internal.NewApp()
	newApp.Run()
}
