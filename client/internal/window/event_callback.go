package window

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"p2p_remote_desk/client/config"
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

func (w *MainWindow) onFullScreenClick() {
	defer func() {
		w.Window.SetFullScreen(w.isFullScreen)
	}()

	for _, item := range w.mainMenu.Items[0].Items {
		if item.Label == config.FullScreen {
			item.Label = config.ExitFullScreen
			break
		}
		if item.Label == config.ExitFullScreen {
			item.Label = config.FullScreen
			break
		}
	}
	w.mainMenu.Items[0].Refresh()

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
	} else {
		w.perfStats.Show()
	}
	for _, item := range w.mainMenu.Items[0].Items {
		if item.Label == config.ShowPerfStats {
			item.Label = config.HiddenPerfStats
			break
		}
		if item.Label == config.HiddenPerfStats {
			item.Label = config.ShowPerfStats
			break
		}
	}
	w.mainMenu.Items[0].Refresh()

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

func (w *MainWindow) onSettingChanged() {
	for _, item := range w.mainMenu.Items[3].Items {
		if item.Label == config.RestoreDefault {
			item.Label = config.RestoreNormal
			w.mainMenu.Items[3].Label = "状态监控"
			break
		}
		if item.Label == config.RestoreNormal {
			item.Label = config.RestoreDefault
			break
		}
	}
	w.mainMenu.Items[3].Refresh()
}
