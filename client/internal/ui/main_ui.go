package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"image"
	"p2p_remote_desk/client/internal/capture"
	"time"
)

// MainUI 主界面组件
type MainUI struct {
	Window    fyne.Window
	Container *fyne.Container

	// 工具栏
	toolbar *ToolbarUI

	// 远程画面相关
	remoteScreen  *canvas.Raster
	lastImage     image.Image
	screenCapture *capture.ScreenCapture
	isCapturing   bool
	captureTimer  *time.Timer

	// 性能监控
	perfStats *PerformanceStatsView
}

// NewMainUI 创建主界面
func NewMainUI(window fyne.Window) *MainUI {
	ui := &MainUI{
		Window: window,
	}

	ui.initComponents()
	ui.setupLayout()
	ui.setupKeyboardHandling()

	ui.perfStats.Hide()

	return ui
}

func (ui *MainUI) initComponents() {
	// 创建性能监控
	ui.perfStats = NewPerformanceStatsView()

	// 创建远程画面
	ui.remoteScreen = canvas.NewRaster(ui.updateScreen)
	ui.screenCapture = capture.NewScreenCapture()

	// 创建工具栏
	ui.NewToolbar()
}

// setupLayout 设置布局
func (ui *MainUI) setupLayout() {
	ui.Container = container.NewBorder(
		ui.toolbar.Toolbar,
		nil,
		nil,
		ui.perfStats.GetContainer(),
		ui.remoteScreen,
	)
}

// setupKeyboardHandling 设置键盘事件处理
func (ui *MainUI) setupKeyboardHandling() {
	ui.Window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		switch ke.Name {
		case fyne.KeyF11:
			ui.onFullScreenClick()
		case fyne.KeyF3:
			ui.togglePerformanceStats()
		case fyne.KeyEscape:
			if ui.toolbar.isFullScreen {
				ui.onFullScreenClick()
			}
		}
	})
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
