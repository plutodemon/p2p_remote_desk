package window

import (
	"time"

	"p2p_remote_desk/lkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func ShowError(window fyne.Window, message string) {
	dialog := widget.NewLabel(message)
	popup := widget.NewModalPopUp(dialog, window.Canvas())
	popup.Show()

	// 2秒后自动关闭错误提示
	lkit.SafeGo(func() {
		time.Sleep(2 * time.Second)
		popup.Hide()
	})
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

	lkit.SafeGoWithCallback(func() {
		// todo 连接逻辑
		//if err := w.mainUI.screenCapture.Connect(serverAddr); err != nil {
		//	llog.Error("连接失败: %v", err)
		//	w.mainUI.SetStatus("连接失败")
		//	return
		//}

		w.SetStatus("已连接")
		w.isConnected = true
		w.StartCapture()
	}, func(err interface{}) {
		w.SetStatus("连接异常")
	})
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
