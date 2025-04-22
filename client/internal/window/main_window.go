package window

import (
	"fmt"
	"image"
	"time"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal/capture"
	"p2p_remote_desk/lkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/glfw-js"
)

type MainWindow struct {
	Window         fyne.Window
	onClose        func()
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

func (newApp *App) newMainWindow(device *DeviceInfo) {
	newApp.mainWindow = &MainWindow{}

	newApp.mainWindow.Window = newApp.fyneApp.NewWindow("远程桌面: " + device.Name)

	newApp.mainWindow.onClose = func() {
		newApp.mainWindow.Cleanup()
		newApp.mainWindow.Window.Close()
		newApp.deviceWindow.Window.Show()
	}

	newApp.mainWindow.setupUI()

	newApp.mainWindow.Window.SetCloseIntercept(newApp.mainWindow.onClose)

	newApp.mainWindow.OnConnectClick()

	newApp.mainWindow.Window.Show()
}

func (w *MainWindow) setupUI() {
	// 创建工具栏
	w.fullScreenTool = widget.NewToolbarAction(theme.ViewFullScreenIcon(), w.onFullScreenClick())
	w.showStatsTool = widget.NewToolbarAction(theme.VisibilityIcon(), w.togglePerformanceStats())
	w.toolbar = widget.NewToolbar(
		w.fullScreenTool,
		w.showStatsTool,
		widget.NewToolbarSpacer(),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.LogoutIcon(), w.onClose),
	)

	// 创建性能监控
	w.perfStats = newPerformanceStatsView()
	w.perfStats.Hide()

	// 创建远程画面
	w.remoteScreen = canvas.NewRaster(w.updateScreen)
	w.screenCapture = capture.NewScreenCapture()

	videoMode := glfw.GetPrimaryMonitor().GetVideoMode()
	width, height := videoMode.Width, videoMode.Height

	content := container.NewBorder(
		w.toolbar,
		nil,
		nil,
		w.perfStats.GetContainer(),
		container.New(&AspectRatioLayout{Ratio: float64(width) / float64(height)}, w.remoteScreen),
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
	interval := time.Second / time.Duration(glfw.GetPrimaryMonitor().GetVideoMode().RefreshRate)

	ticker := time.NewTicker(interval)
	lastCaptureTime := time.Now()

	lkit.SafeGo(func() {
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
	})
}
