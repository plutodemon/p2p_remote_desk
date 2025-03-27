package window

import (
	"fmt"
	"image"
	"time"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal/capture"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MainWindow struct {
	Window         fyne.Window
	toolbar        *widget.Toolbar
	fullScreenTool *widget.ToolbarAction
	showStatsTool  *widget.ToolbarAction

	perfStats     *PerformanceStatsView
	remoteScreen  *canvas.Raster
	screenCapture *capture.ScreenCapture

	lastImage    image.Image
	captureTimer *time.Timer

	isConnected  bool
	isFullScreen bool
	isShowStats  bool
	isCapturing  bool
}

func NewMainWindow(window fyne.Window) *MainWindow {
	w := &MainWindow{
		Window: window,
	}

	w.setupUI()

	return w
}

func (w *MainWindow) setupUI() {
	// 创建工具栏
	w.fullScreenTool = widget.NewToolbarAction(theme.ViewFullScreenIcon(), w.onFullScreenClick())
	w.showStatsTool = widget.NewToolbarAction(theme.VisibilityIcon(), w.togglePerformanceStats())
	w.toolbar = widget.NewToolbar(
		w.fullScreenTool,
		//widget.NewToolbarSeparator(),
		w.showStatsTool,
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.HelpIcon(), func() {}),
	)

	// 创建性能监控
	w.perfStats = newPerformanceStatsView()
	w.perfStats.Hide()

	// 创建远程画面
	w.remoteScreen = canvas.NewRaster(w.updateScreen)
	w.screenCapture = capture.NewScreenCapture()

	content := container.NewBorder(
		w.toolbar,
		nil,
		nil,
		w.perfStats.GetContainer(),
		w.remoteScreen,
	)
	w.Window.SetContent(content)
	w.Window.Resize(config.WindowSize)
	w.Window.CenterOnScreen()
	// w.Window.SetMaster()
}

func (w *MainWindow) SetStatus(status string) {
}

func (w *MainWindow) updateScreen(weight, high int) image.Image {
	if w.lastImage == nil {
		img := image.NewRGBA(image.Rect(0, 0, weight, high))
		return img
	}
	return w.lastImage
}

func (w *MainWindow) Cleanup() {
	w.StopCapture()
	if w.screenCapture != nil {
		// w.screenCapture.Close()
	}
	if w.perfStats != nil {
		w.perfStats.Cleanup()
	}
}

func (w *MainWindow) SetFPS(fps float64) {
	//if w.statusBar != config.FPS {
	//	return
	//}
	//w.SetStatus(fmt.Sprintf("%.1f", fps))
}

// SetQuality 设置画面质量
func (w *MainWindow) SetQuality(quality string) {
	if w.screenCapture == nil {
		return
	}

	// 设置屏幕捕获的质量
	w.screenCapture.Quality = quality
	w.SetStatus(fmt.Sprintf("已设置画面质量: %s", quality))
}

// SetDisplayIndex 设置显示器索引
func (w *MainWindow) SetDisplayIndex(displayName string) {
	// 设置屏幕捕获的显示器索引
	if w.screenCapture != nil {
		// 从字符串中提取显示器索引
		// 格式为 "显示器名称(索引)"
		index := 0
		// 这里可以添加解析逻辑，从字符串中提取索引

		w.screenCapture.DisplayIndex = index
		w.SetStatus(fmt.Sprintf("已切换到显示器: %s", displayName))
	}
}

// StopCapture 停止屏幕捕获
func (w *MainWindow) StopCapture() {
	if !w.isCapturing {
		return
	}
	w.isCapturing = false
	w.lastImage = nil
	w.remoteScreen.Refresh()
}

// StartCapture 开始屏幕捕获
func (w *MainWindow) StartCapture() {
	if w.isCapturing {
		return
	}

	w.isCapturing = true
	interval := time.Second / time.Duration(60)

	ticker := time.NewTicker(interval)
	lastCaptureTime := time.Now()

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			if !w.isCapturing {
				return
			}

			// 更新FPS
			now := time.Now()
			actualFPS := 1.0 / now.Sub(lastCaptureTime).Seconds()
			lastCaptureTime = now
			w.SetFPS(actualFPS)

			// 捕获并显示画面
			if w.screenCapture != nil {
				img, err := w.screenCapture.CaptureScreen()
				if err != nil {
					continue
				}

				w.lastImage = img
				w.remoteScreen.Refresh()
			}
		}
	}()
}
