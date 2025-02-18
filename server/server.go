package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"p2p_remote_desk/server/config"
	"p2p_remote_desk/server/internal"

	"github.com/plutodemon/llog"
)

// SigChan 创建一个通道来接收信号
var SigChan = make(chan os.Signal)

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
	signal.Notify(SigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	// 每个服务都用一个goroutine来运行\

	// 验证服务
	go func() {
		internal.Authentic()
	}()

	// 信令服务
	go func() {
		internal.SignalServer()
	}()

	// ice服务
	go func() {
		// todo
	}()

	// 主goroutine等待信号 一直阻塞
	select {
	case sig := <-SigChan:
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
