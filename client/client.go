package main

import (
	"fmt"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal/window"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
	lkit.InitCrashLog()
	defer lkit.CrashLog()

	// 确保只有一个实例运行
	if err := config.EnsureSingleInstance(); err != nil {
		panic(fmt.Sprintf("程序已在运行: %v", err))
	}
	defer config.CleanupLock()

	// 初始化配置
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("初始化配置失败: %v", err))
	}

	// 初始化日志系统
	if err := llog.Init(config.GetConfig().LogConfig); err != nil {
		panic(fmt.Sprintf("初始化日志系统失败: %v", err))
	}
	defer llog.Cleanup()

	// 设置全局panic处理
	defer llog.HandlePanic()

	if err := glfw.Init(); err != nil {
		llog.Fatal("初始化glfw失败: ", err)
	}
	defer glfw.Terminate()

	window.NewAppAndRun()
}
