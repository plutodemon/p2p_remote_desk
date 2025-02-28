package window

import (
	"fmt"
	"strconv"

	"github.com/plutodemon/llog"
	"p2p_remote_desk/client/config"
)

// 菜单事件处理函数
func (w *MainWindow) onConnectClick() {
	if w.isConnected {
		// 断开连接
		w.mainUI.SetStatus("已断开连接")
		w.isConnected = false
		w.mainUI.StopCapture()
		return
	}

	w.handleControllerConnect()
}

func (w *MainWindow) handleControllerConnect() {
	//cfg := config.GetConfig()
	//serverAddr := cfg.ServerConfig.Address + ":" + cfg.ServerConfig.Port

	w.mainUI.SetStatus("正在连接...")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				llog.Error("连接过程发生panic: %v", r)
				w.mainUI.SetStatus("连接异常")
			}
		}()

		// todo 连接逻辑
		//if err := w.mainUI.screenCapture.Connect(serverAddr); err != nil {
		//	llog.Error("连接失败: %v", err)
		//	w.mainUI.SetStatus("连接失败")
		//	return
		//}

		w.mainUI.SetStatus("已连接")
		w.isConnected = true
		w.mainUI.StartCapture()
	}()
}

func (w *MainWindow) onFullScreenClick() {
	if w.isFullScreen {
		// 退出全屏
		w.window.SetFullScreen(false)

		// 更新菜单项文本
		for _, item := range w.mainMenu.Items[1].Items {
			if item.Label == config.ExitFullScreen {
				item.Label = config.FullScreen
				break
			}
		}

		// 性能监控
		w.mainUI.PerfStats().Hide()
		w.isShowStats = false
	} else {
		// 进入全屏
		w.window.SetFullScreen(true)

		// 更新菜单项文本
		for _, item := range w.mainMenu.Items[1].Items {
			if item.Label == config.FullScreen {
				item.Label = config.ExitFullScreen
				break
			}
		}
	}

	// 性能监控
	w.mainUI.PerfStats().Hide()
	w.isShowStats = false

	w.isFullScreen = !w.isFullScreen
}

func (w *MainWindow) onDisplayChanged(s string) {
	// 设置屏幕捕获的显示器索引
	w.mainUI.SetDisplayIndex(s)
}

func (w *MainWindow) togglePerformanceStats() {
	if w.isShowStats {
		w.mainUI.PerfStats().Hide()

		// 更新菜单项文本
		for _, item := range w.mainMenu.Items[1].Items {
			if item.Label == config.HiddenPerfStats {
				item.Label = config.ShowPerfStats
				break
			}
		}
	} else {
		w.mainUI.PerfStats().Show()

		// 更新菜单项文本
		for _, item := range w.mainMenu.Items[1].Items {
			if item.Label == config.ShowPerfStats {
				item.Label = config.HiddenPerfStats
				break
			}
		}
	}
	w.isShowStats = !w.isShowStats
}

func (w *MainWindow) onQualityChanged(s string) {
	w.mainUI.SetQuality(s)
}

func (w *MainWindow) onFPSChanged(s string) {
	_, err := strconv.Atoi(s)
	if err != nil {
		llog.Error("解析帧率失败: %v", err)
		return
	}

	w.mainUI.SetStatus(fmt.Sprintf("已设置帧率: %s", s))

	if w.isConnected {
		w.mainUI.StopCapture()
		w.mainUI.StartCapture()
	}
}

func (w *MainWindow) onModeChanged() {
	if w.isController {
		w.mainUI.SetStatus(config.ControlledEnd)
	} else {
		w.mainUI.SetStatus(config.ControlEnd)
	}

	w.isController = !w.isController
}
