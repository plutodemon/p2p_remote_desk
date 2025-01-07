package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"image"
	"p2p_remote_desk/config"
	"p2p_remote_desk/logger"
	"remote_desk/internal/core/capture"
	"strconv"
	"time"
)

// MainUI 主界面组件
type MainUI struct {
	Window    fyne.Window
	Container *fyne.Container

	// 工具栏
	toolbar *Toolbar

	// 远程画面相关
	remoteScreen  *canvas.Raster
	lastImage     image.Image
	screenCapture *capture.ScreenCapture
	isCapturing   bool
	captureTimer  *time.Timer

	// 性能监控
	perfStats *PerformanceStatsView
	showStats bool

	isMaximized bool
}

// NewMainUI 创建主界面
func NewMainUI(window fyne.Window) *MainUI {
	ui := &MainUI{
		Window: window,
	}

	ui.initComponents()
	ui.setupLayout()
	ui.setupKeyboardHandling()

	return ui
}

// initComponents 初始化组件
func (ui *MainUI) initComponents() {
	// 创建工具栏
	NewToolbar(ui)
	ui.setupToolbarCallbacks()

	// 创建远程画面
	ui.remoteScreen = canvas.NewRaster(ui.updateScreen)
	ui.screenCapture = capture.NewScreenCapture()

	// 创建性能监控
	ui.perfStats = NewPerformanceStatsView()
}

// setupLayout 设置布局
func (ui *MainUI) setupLayout() {
	ui.Container = container.NewBorder(
		ui.toolbar.GetContainer(),   // 顶部工具栏
		nil,                         // 底部
		nil,                         // 左侧
		ui.perfStats.GetContainer(), // 右侧性能监控
		ui.remoteScreen,             // 中间的远程画面
	)
}

// setupKeyboardHandling 设置键盘事件处理
func (ui *MainUI) setupKeyboardHandling() {
	ui.Window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		switch ke.Name {
		case fyne.KeyF11:
			ui.toggleFullScreen()
		case fyne.KeyF3:
			ui.togglePerformanceStats()
		case fyne.KeyEscape:
			if ui.isMaximized {
				ui.toggleFullScreen()
			}
		}
	})
}

// setupToolbarCallbacks 设置工具栏回调
func (ui *MainUI) setupToolbarCallbacks() {
	ui.toolbar.modeSelect.OnChanged = ui.onModeChanged
	ui.toolbar.connectBtn.OnTapped = ui.onConnectClick
	ui.toolbar.fullScreenBtn.OnTapped = ui.onFullScreenClick
	ui.toolbar.qualitySelect.OnChanged = ui.onQualityChanged
	ui.toolbar.fpsSelect.OnChanged = ui.onFPSChanged
	ui.toolbar.perfStatsBtn.OnTapped = ui.togglePerformanceStats
}

// 屏幕显示相关方法
func (ui *MainUI) updateScreen(w, h int) image.Image {
	if ui.lastImage == nil {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		return img
	}
	return ui.lastImage
}

func (ui *MainUI) StartCapture() {
	if ui.isCapturing {
		return
	}

	fps, _ := strconv.Atoi(ui.toolbar.fpsSelect.Selected)
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
					logger.Error("屏幕捕获失败: %v", err)
					continue
				}

				ui.lastImage = img
				ui.remoteScreen.Refresh()
			}
		}
	}()
}

func (ui *MainUI) StopCapture() {
	if !ui.isCapturing {
		return
	}
	ui.isCapturing = false
	ui.lastImage = nil
	ui.remoteScreen.Refresh()
}

// 事件处理方法
func (ui *MainUI) onModeChanged(s string) {
	if s == "控制端" {
		ui.toolbar.isController = true
		ui.toolbar.connectBtn.Enable()
		ui.toolbar.SetStatus("等待连接到被控端...")
	} else {
		ui.toolbar.isController = false
		ui.toolbar.connectBtn.Disable()
		ui.toolbar.SetStatus("等待控制端连接...")
	}
	NewToolbar(ui)
}

func (ui *MainUI) onConnectClick() {
	cfg := config.GetConfig()

	if ui.toolbar.modeSelect.Selected == "控制端" {
		ui.handleControllerConnect(cfg)
	} else {
		ui.handleControlledConnect(cfg)
	}
}

func (ui *MainUI) handleControllerConnect(cfg *config.Config) {
	serverAddr := cfg.ServerConfig.Address + ":" + cfg.ServerConfig.Port
	ui.toolbar.SetStatus("正在连接...")
	ui.toolbar.connectBtn.Disable()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("连接过程发生panic: %v", r)
				ui.toolbar.SetStatus("连接异常")
				ui.toolbar.connectBtn.Enable()
			}
		}()

		if err := ui.screenCapture.Connect(serverAddr); err != nil {
			logger.Error("连接失败: %v", err)
			ui.toolbar.SetStatus("连接失败")
			ui.toolbar.connectBtn.Enable()
			return
		}

		ui.toolbar.SetStatus("已连接")
		ui.StartCapture()
	}()
}

func (ui *MainUI) handleControlledConnect(cfg *config.Config) {
	ui.toolbar.SetStatus("等待连接...")
	ui.toolbar.connectBtn.Disable()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("监听过程发生panic: %v", r)
				ui.toolbar.SetStatus("监听异常")
				ui.toolbar.connectBtn.Enable()
			}
		}()

		if err := ui.screenCapture.Listen(cfg.ServerConfig.Port); err != nil {
			logger.Error("启动监听失败: %v", err)
			ui.toolbar.SetStatus("启动失败")
			ui.toolbar.connectBtn.Enable()
			return
		}

		ui.toolbar.SetStatus("已连接")
		ui.StartCapture()
	}()
}

func (ui *MainUI) onFullScreenClick() {
	ui.toggleFullScreen()
}

func (ui *MainUI) onQualityChanged(s string) {
	if ui.screenCapture == nil {
		return
	}

	quality := 100
	switch s {
	case "低":
		quality = 50
	case "中":
		quality = 75
	case "高":
		quality = 100
	}

	ui.screenCapture.SetQuality(quality)
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

// 窗口控制方法
func (ui *MainUI) toggleFullScreen() {
	cfg := config.GetConfig()

	if ui.isMaximized {
		// 退出全屏
		ui.Window.SetFullScreen(false)
		ui.toolbar.fullScreenBtn.SetText("全屏")
		// 退出全屏时总是显示工具栏
		ui.toolbar.container.Show()
		// 性能监控
		ui.perfStats.Hide()
		ui.toolbar.perfStatsBtn.SetText("性能监控")
		ui.showStats = false
	} else {
		// 进入全屏
		ui.Window.SetFullScreen(true)
		ui.toolbar.fullScreenBtn.SetText("退出全屏")
		// 根据配置决定是否隐藏工具栏
		if cfg.UIConfig.HideToolbarInFullscreen {
			ui.toolbar.container.Hide()
			ui.perfStats.Hide()
			ui.toolbar.perfStatsBtn.SetText("性能监控")
			ui.showStats = false
		}
	}
	ui.isMaximized = !ui.isMaximized
}

func (ui *MainUI) togglePerformanceStats() {
	if ui.perfStats != nil {
		if ui.showStats {
			ui.perfStats.Hide()
			ui.toolbar.perfStatsBtn.SetText("性能监控")
		} else {
			ui.perfStats.Show()
			ui.toolbar.perfStatsBtn.SetText("隐藏监控")
		}
		ui.showStats = !ui.showStats
	}
}

// Cleanup 清理资源
func (ui *MainUI) Cleanup() {
	ui.StopCapture()
	if ui.screenCapture != nil {
		ui.screenCapture.Close()
	}
	if ui.perfStats != nil {
		ui.perfStats.Cleanup()
	}
}
