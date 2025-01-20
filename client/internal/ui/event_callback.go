package ui

import (
	"github.com/plutodemon/slog"
	config2 "p2p_remote_desk/client/config"
	"strconv"
	"time"
)

func (ui *MainUI) onConnectClick() {
	if ui.toolbar.isConnected {
		ui.handleControllerConnect()
		return
	}

	// todo 断开连接
}

func (ui *MainUI) handleControllerConnect() {
	//cfg := config.GetConfig()
	//serverAddr := cfg.ServerConfig.Address + ":" + cfg.ServerConfig.Port

	ui.toolbar.SetStatus("正在连接...")
	ui.toolbar.ConnectBtn.Button.Disable()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("连接过程发生panic: %v", r)
				ui.toolbar.SetStatus("连接异常")
				ui.toolbar.ConnectBtn.Button.Enable()
			}
		}()

		// todo 连接逻辑
		//if err := ui.screenCapture.Connect(serverAddr); err != nil {
		//	slog.Error("连接失败: %v", err)
		//	ui.toolbar.SetStatus("连接失败")
		//	ui.toolbar.ConnectBtn.Button.Enable()
		//	return
		//}

		ui.toolbar.SetStatus("已连接")
		ui.toolbar.ConnectBtn.Button.SetText(config2.ConnectBtnNameClose)
		ui.toolbar.ConnectBtn.Button.Enable()
		ui.toolbar.isConnected = true
		// ui.StartCapture()
	}()
}

func (ui *MainUI) onFullScreenClick() {
	cfg := config2.GetConfig()

	if ui.toolbar.isFullScreen {
		// 退出全屏
		ui.Window.SetFullScreen(false)
		ui.toolbar.FullScreenBtn.Button.SetText(config2.FullScreen)

		// 退出全屏时总是显示工具栏
		ui.toolbar.Toolbar.Show()

		// 性能监控
		ui.perfStats.Hide()
		ui.toolbar.PerfStatsBtn.Button.SetText(config2.ShowPerfStats)
		ui.toolbar.isShowStats = false
	} else {
		// 进入全屏
		ui.Window.SetFullScreen(true)
		ui.toolbar.FullScreenBtn.Button.SetText(config2.ExitFullScreen)

		// 根据配置决定是否隐藏工具栏
		if cfg.UIConfig.HideToolbarInFullscreen {
			ui.toolbar.Toolbar.Hide()
		}
	}

	// 性能监控
	ui.perfStats.Hide()
	ui.toolbar.PerfStatsBtn.Button.SetText(config2.ShowPerfStats)
	ui.toolbar.isShowStats = false

	ui.toolbar.isFullScreen = !ui.toolbar.isFullScreen
}

func (ui *MainUI) onDisplayChanged(s string) {
	// todo
}

func (ui *MainUI) togglePerformanceStats() {
	if ui.toolbar.isShowStats {
		ui.perfStats.Hide()
		ui.toolbar.PerfStatsBtn.Button.SetText(config2.ShowPerfStats)
	} else {
		ui.perfStats.Show()
		ui.toolbar.PerfStatsBtn.Button.SetText(config2.HiddenPerfStats)
	}
	ui.toolbar.isShowStats = !ui.toolbar.isShowStats
}

func (ui *MainUI) onQualityChanged(s string) {
	if ui.screenCapture == nil {
		return
	}

	//quality := 100
	//switch s {
	//case "低":
	//	quality = 50
	//case "中":
	//	quality = 75
	//case "高":
	//	quality = 100
	//}
	//
	//ui.screenCapture.SetQuality(quality)
}

func (ui *MainUI) onFPSChanged(s string) {
	_, err := strconv.Atoi(s)
	if err != nil {
		slog.Error("解析帧率失败: %v", err)
		return
	}

	if ui.isCapturing {
		ui.StopCapture()
		ui.StartCapture()
	}
}

func (ui *MainUI) StopCapture() {
	if !ui.isCapturing {
		return
	}
	ui.isCapturing = false
	ui.lastImage = nil
	ui.remoteScreen.Refresh()
}

func (ui *MainUI) StartCapture() {
	if ui.isCapturing {
		return
	}

	fps, _ := strconv.Atoi(ui.toolbar.FpsSelect.Select.Selected)
	if fps <= 0 {
		fps = 30
	}

	ui.isCapturing = true
	interval := time.Second / time.Duration(fps)

	ticker := time.NewTicker(interval)
	lastCaptureTime := time.Now()

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			if !ui.isCapturing {
				return
			}

			// 更新FPS
			now := time.Now()
			actualFPS := 1.0 / now.Sub(lastCaptureTime).Seconds()
			lastCaptureTime = now
			ui.toolbar.SetFPS(actualFPS)

			// 捕获并显示画面
			if ui.screenCapture != nil {
				img, err := ui.screenCapture.CaptureScreen()
				if err != nil {
					slog.Error("屏幕捕获失败: %v", err)
					continue
				}

				ui.lastImage = img
				ui.remoteScreen.Refresh()
			}
		}
	}()
}

func (ui *MainUI) onModeChanged(_ float64) {
	if ui.toolbar.isController {
		ui.toolbar.ModeState.Label.SetText(config2.ControlledEnd)
	} else {
		ui.toolbar.ModeState.Label.SetText(config2.ControlEnd)
	}

	ui.toolbar.isController = !ui.toolbar.isController
}
