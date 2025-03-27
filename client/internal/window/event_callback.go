package window

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"p2p_remote_desk/llog"
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

func (w *MainWindow) OnConnectClick() {
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

func (w *MainWindow) onFullScreenClick() func() {
	return func() {
		if w.isFullScreen {
			w.fullScreenTool.SetIcon(theme.ViewFullScreenIcon())
		} else {
			w.fullScreenTool.SetIcon(theme.ViewRestoreIcon())
		}
		w.isFullScreen = !w.isFullScreen
		w.Window.SetFullScreen(w.isFullScreen)
		w.toolbar.Refresh()
	}
}

func (w *MainWindow) onDisplayChanged(s string) {
	w.SetDisplayIndex(s)
}

func (w *MainWindow) togglePerformanceStats() func() {
	return func() {
		if w.isShowStats {
			w.showStatsTool.SetIcon(theme.VisibilityIcon())
			w.perfStats.Hide()
		} else {
			w.showStatsTool.SetIcon(theme.VisibilityOffIcon())
			w.perfStats.Show()
		}
		w.isShowStats = !w.isShowStats
		w.toolbar.Refresh()
	}
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
