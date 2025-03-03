package window

import (
	"fmt"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"image"
	"p2p_remote_desk/client/internal/capture"
	"p2p_remote_desk/lkit"
	"time"

	"fyne.io/fyne/v2"
	"p2p_remote_desk/client/config"
)

type MainWindow struct {
	Window    fyne.Window
	mainMenu  *fyne.MainMenu
	statusBar string

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

	// 设置菜单栏
	w.setupMainMenu()

	w.setupUI()

	// 设置窗口关闭回调
	window.SetCloseIntercept(func() {
		w.Cleanup()
		window.Close()
	})

	return w
}

func (w *MainWindow) setupMainMenu() {
	// todo 换为弹窗确认
	//connectMenu := fyne.NewMenu("连接",
	//	fyne.NewMenuItem(config.ConnectBtnNameOpen, w.onConnectClick),
	//)

	// 视图菜单
	viewMenu := fyne.NewMenu("视图",
		fyne.NewMenuItem(config.FullScreen, w.onFullScreenClick),
		fyne.NewMenuItem(config.ShowPerfStats, w.togglePerformanceStats),
	)

	screenConfig := config.GetConfig().ScreenConfig
	qualityItems := make([]*fyne.MenuItem, 0)
	for _, setting := range screenConfig.QualityList {
		qualityName := setting.Name
		qualityItems = append(qualityItems, fyne.NewMenuItem(qualityName, func() { w.onQualityChanged(qualityName) }))
	}

	fpsItems := make([]*fyne.MenuItem, 0)
	for _, fps := range screenConfig.FrameRates {
		fpsStr := lkit.AnyToStr(fps)
		fpsItems = append(fpsItems, fyne.NewMenuItem(fpsStr, func() { w.onFPSChanged(fpsStr) }))
	}

	setting := fyne.NewMenu("状态监控",
		fyne.NewMenuItem(config.Normal, func() { w.onSettingChanged(config.Normal) }),
		fyne.NewMenuItem(config.NetworkDelay, func() { w.onSettingChanged(config.NetworkDelay) }),
		fyne.NewMenuItem(config.FPS, func() { w.onSettingChanged(config.FPS) }),
	)

	// 创建主菜单
	w.mainMenu = fyne.NewMainMenu(
		viewMenu,
		fyne.NewMenu("画面质量", qualityItems...),
		fyne.NewMenu("帧率", fpsItems...),
		setting,
	)

	// 设置窗口内容
	w.Window.SetMainMenu(w.mainMenu)
}

func (w *MainWindow) setupUI() {
	// 创建性能监控
	w.perfStats = newPerformanceStatsView()
	w.perfStats.Hide()

	// 创建远程画面
	w.remoteScreen = canvas.NewRaster(w.updateScreen)
	w.screenCapture = capture.NewScreenCapture()

	content := container.NewBorder(
		nil,
		nil,
		nil,
		w.perfStats.GetContainer(),
		w.remoteScreen,
	)
	w.Window.SetContent(content)
	w.Window.Resize(config.WindowSize)
	w.Window.CenterOnScreen()
	w.Window.SetMaster()
}

func (w *MainWindow) SetStatus(status string) {
	switch w.statusBar {
	case config.NetworkDelay:
		w.mainMenu.Items[3].Label = config.NetworkDelay + ":" + status
	case config.FPS:
		w.mainMenu.Items[3].Label = config.FPS + ":" + status
	default:
		w.mainMenu.Items[3].Label = status
	}
	w.mainMenu.Items[3].Refresh()
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
	if w.statusBar != config.FPS {
		return
	}
	w.SetStatus(fmt.Sprintf("%.1f", fps))
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

	// 从配置中获取默认帧率
	fps := config.GetConfig().ScreenConfig.DefaultFrameRate
	if fps <= 0 {
		fps = 30
	}

	w.isCapturing = true
	interval := time.Second / time.Duration(fps)

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
