package ui

import (
	"fmt"
	"image"
	"time"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal/capture"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MainUI 主界面组件
type MainUI struct {
	Window    fyne.Window
	Container *fyne.Container

	// 远程画面相关
	remoteScreen  *canvas.Raster
	lastImage     image.Image
	screenCapture *capture.ScreenCapture
	isCapturing   bool
	captureTimer  *time.Timer

	// 状态栏
	statusBar *widget.Label

	// 性能监控
	perfStats *PerformanceStatsView

	// 状态标志
	isConnected  bool
	isFullScreen bool
	isShowStats  bool
	isController bool
}

// NewMainUI 创建主界面
func NewMainUI(window fyne.Window) *MainUI {
	ui := &MainUI{
		Window: window,
	}

	ui.initComponents()
	ui.setupLayout()

	ui.perfStats.Hide()

	return ui
}

func (ui *MainUI) initComponents() {
	// 创建性能监控
	ui.perfStats = NewPerformanceStatsView()

	// 创建远程画面
	ui.remoteScreen = canvas.NewRaster(ui.updateScreen)
	ui.screenCapture = capture.NewScreenCapture()

	// 创建状态栏
	ui.statusBar = widget.NewLabel("就绪")
}

// setupLayout 设置布局
func (ui *MainUI) setupLayout() {
	ui.Container = container.NewBorder(
		nil,
		container.NewHBox(ui.statusBar),
		nil,
		ui.perfStats.GetContainer(),
		ui.remoteScreen,
	)
}

// 屏幕显示相关方法
func (ui *MainUI) updateScreen(w, h int) image.Image {
	if ui.lastImage == nil {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		return img
	}
	return ui.lastImage
}

// Cleanup 清理资源
func (ui *MainUI) Cleanup() {
	ui.StopCapture()
	if ui.screenCapture != nil {
		// ui.screenCapture.Close()
	}
	if ui.perfStats != nil {
		ui.perfStats.Cleanup()
	}
}

// SetStatus 设置状态栏文本
func (ui *MainUI) SetStatus(status string) {
	ui.statusBar.SetText(status)
}

// SetFPS 设置FPS显示
func (ui *MainUI) SetFPS(fps float64) {
	// 在状态栏显示FPS
	ui.statusBar.SetText(fmt.Sprintf("就绪 | FPS: %.1f", fps))
}

// PerfStats 获取性能监控视图
func (ui *MainUI) PerfStats() *PerformanceStatsView {
	return ui.perfStats
}

// SetQuality 设置画面质量
func (ui *MainUI) SetQuality(quality string) {
	if ui.screenCapture == nil {
		return
	}

	// 设置屏幕捕获的质量
	ui.screenCapture.Quality = quality
	ui.SetStatus(fmt.Sprintf("已设置画面质量: %s", quality))
}

// SetDisplayIndex 设置显示器索引
func (ui *MainUI) SetDisplayIndex(displayName string) {
	// 设置屏幕捕获的显示器索引
	if ui.screenCapture != nil {
		// 从字符串中提取显示器索引
		// 格式为 "显示器名称(索引)"
		index := 0
		// 这里可以添加解析逻辑，从字符串中提取索引

		ui.screenCapture.DisplayIndex = index
		ui.SetStatus(fmt.Sprintf("已切换到显示器: %s", displayName))
	}
}

// StopCapture 停止屏幕捕获
func (ui *MainUI) StopCapture() {
	if !ui.isCapturing {
		return
	}
	ui.isCapturing = false
	ui.lastImage = nil
	ui.remoteScreen.Refresh()
}

// StartCapture 开始屏幕捕获
func (ui *MainUI) StartCapture() {
	if ui.isCapturing {
		return
	}

	// 从配置中获取默认帧率
	fps := config.GetConfig().ScreenConfig.DefaultFrameRate
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
			ui.SetFPS(actualFPS)

			// 捕获并显示画面
			if ui.screenCapture != nil {
				img, err := ui.screenCapture.CaptureScreen()
				if err != nil {
					continue
				}

				ui.lastImage = img
				ui.remoteScreen.Refresh()
			}
		}
	}()
}
