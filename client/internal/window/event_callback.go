package window

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"strconv"
	"time"

	"github.com/plutodemon/llog"
	"p2p_remote_desk/client/config"
)

func ShowError(window fyne.Window, message string) {
	dialog := widget.NewLabel(message)
	popup := widget.NewModalPopUp(dialog, window.Canvas())
	popup.Show()

	// 2秒后自动关闭错误提示
	go func() {
		time.Sleep(2 * time.Second)
		popup.Hide()
	}()
}

// 菜单事件处理函数
func (w *MainWindow) onConnectClick() {
	if w.isConnected {
		// 断开连接
		w.SetStatus("已断开连接")
		w.isConnected = false
		w.StopCapture()
		return
	}

	w.handleControllerConnect()
}

func (w *MainWindow) handleControllerConnect() {
	//cfg := config.GetConfig()
	//serverAddr := cfg.ServerConfig.Address + ":" + cfg.ServerConfig.Port

	w.SetStatus("正在连接...")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				llog.Error("连接过程发生panic: %v", r)
				w.SetStatus("连接异常")
			}
		}()

		// todo 连接逻辑
		//if err := w.mainUI.screenCapture.Connect(serverAddr); err != nil {
		//	llog.Error("连接失败: %v", err)
		//	w.mainUI.SetStatus("连接失败")
		//	return
		//}

		w.SetStatus("已连接")
		w.isConnected = true
		w.StartCapture()
	}()
}

func (w *MainWindow) onFullScreenClick() {
	if w.isFullScreen {
		// 退出全屏
		w.Window.SetFullScreen(false)

		// 更新菜单项文本
		for _, item := range w.mainMenu.Items[1].Items {
			if item.Label == config.ExitFullScreen {
				item.Label = config.FullScreen
				break
			}
		}

		// 性能监控
		w.perfStats.Hide()
		w.isShowStats = false
	} else {
		// 进入全屏
		w.Window.SetFullScreen(true)

		// 更新菜单项文本
		for _, item := range w.mainMenu.Items[1].Items {
			if item.Label == config.FullScreen {
				item.Label = config.ExitFullScreen
				break
			}
		}
	}

	// 性能监控
	w.perfStats.Hide()
	w.isShowStats = false

	w.isFullScreen = !w.isFullScreen
}

func (w *MainWindow) onDisplayChanged(s string) {
	w.SetDisplayIndex(s)
}

func (w *MainWindow) togglePerformanceStats() {
	if w.isShowStats {
		w.perfStats.Hide()

		// 更新菜单项文本
		for _, item := range w.mainMenu.Items[1].Items {
			if item.Label == config.HiddenPerfStats {
				item.Label = config.ShowPerfStats
				break
			}
		}
	} else {
		w.perfStats.Show()

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
	w.SetQuality(s)
}

func (w *MainWindow) onFPSChanged(s string) {
	_, err := strconv.Atoi(s)
	if err != nil {
		llog.Error("解析帧率失败: %v", err)
		return
	}

	w.SetStatus(fmt.Sprintf("已设置帧率: %s", s))

	if w.isConnected {
		w.StopCapture()
		w.StartCapture()
	}
}

func (w *MainWindow) onSettingChanged(setting string) {
	w.statusBar = setting
}
