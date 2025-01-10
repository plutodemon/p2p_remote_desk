package ui

import (
	"p2p_remote_desk/config"
	"p2p_remote_desk/logger"
	"strconv"
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
				logger.Error("连接过程发生panic: %v", r)
				ui.toolbar.SetStatus("连接异常")
				ui.toolbar.ConnectBtn.Button.Enable()
			}
		}()

		// todo 连接逻辑
		//if err := ui.screenCapture.Connect(serverAddr); err != nil {
		//	logger.Error("连接失败: %v", err)
		//	ui.toolbar.SetStatus("连接失败")
		//	ui.toolbar.ConnectBtn.Button.Enable()
		//	return
		//}

		ui.toolbar.SetStatus("已连接")
		ui.toolbar.ConnectBtn.Button.SetText(config.ConnectBtnNameClose)
		ui.toolbar.ConnectBtn.Button.Enable()
		ui.toolbar.isConnected = true
		// ui.StartCapture()
	}()
}

func (ui *MainUI) onFullScreenClick() {
	cfg := config.GetConfig()

	if ui.toolbar.isFullScreen {
		// 退出全屏
		ui.Window.SetFullScreen(false)
		ui.toolbar.FullScreenBtn.Button.SetText("全屏")

		// 退出全屏时总是显示工具栏
		ui.toolbar.Toolbar.Show()

		// 性能监控
		ui.perfStats.Hide()
		ui.toolbar.PerfStatsBtn.Button.SetText("性能监控")
		ui.toolbar.isShowStats = false
	} else {
		// 进入全屏
		ui.Window.SetFullScreen(true)
		ui.toolbar.FullScreenBtn.Button.SetText("退出全屏")

		// 根据配置决定是否隐藏工具栏
		if cfg.UIConfig.HideToolbarInFullscreen {
			ui.toolbar.Toolbar.Hide()
		}
	}

	// 性能监控
	ui.perfStats.Hide()
	ui.toolbar.PerfStatsBtn.Button.SetText("性能监控")
	ui.toolbar.isShowStats = false

	ui.toolbar.isFullScreen = !ui.toolbar.isFullScreen
}

func (ui *MainUI) onDisplayChanged(s string) {
	// todo
}

func (ui *MainUI) togglePerformanceStats() {
	if ui.toolbar.isShowStats {
		ui.perfStats.Hide()
		ui.toolbar.PerfStatsBtn.Button.SetText("性能监控")
	} else {
		ui.perfStats.Show()
		ui.toolbar.PerfStatsBtn.Button.SetText("隐藏监控")
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
		logger.Error("解析帧率失败: %v", err)
		return
	}

	if ui.isCapturing {
		ui.StopCapture()
		ui.StartCapture()
	}
}

func (ui *MainUI) onModeChanged(_ float64) {
	if ui.toolbar.isController {
		ui.toolbar.ModeState.Label.SetText("被控端")
	} else {
		ui.toolbar.ModeState.Label.SetText("控制端")
	}

	ui.toolbar.isController = !ui.toolbar.isController
}
