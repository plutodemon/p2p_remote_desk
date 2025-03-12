package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"
	"p2p_remote_desk/server/config"
	"p2p_remote_desk/server/internal/auth"
	"p2p_remote_desk/server/internal/signaling"
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

	// 注册要接收的信号
	// kill pid 是发送SIGTERM的信号 ; kill -9 pid 是发送SIGKILL的信号(无法捕获)
	signal.Notify(lkit.SigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	// 每个服务都用一个goroutine来运行

	// 验证服务
	go func() {
		auth.Start()
	}()

	// 信令服务
	go func() {
		signaling.Start()
	}()

	// ice服务
	go func() {
		// todo
	}()

	// 主goroutine等待信号 一直阻塞
	select {
	case sig := <-lkit.SigChan:
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			llog.Info("收到退出信号: ", sig)
			// todo 执行清理操作
			return
		default:
			llog.Info("收到未知信号: ", sig)
		}
	}
}
