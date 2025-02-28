package window

import (
	"strconv"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal/ui"

	"fyne.io/fyne/v2"
)

// MainWindow 主窗口管理器
type MainWindow struct {
	window fyne.Window
	mainUI *ui.MainUI

	// 菜单相关
	mainMenu     *fyne.MainMenu
	isConnected  bool
	isFullScreen bool
	isShowStats  bool
	isController bool
}

// NewMainWindow 创建主窗口管理器
func NewMainWindow(window fyne.Window) *MainWindow {
	w := &MainWindow{
		window: window,
	}

	// 创建主UI
	w.mainUI = ui.NewMainUI(window)

	// 设置菜单栏
	w.setupMainMenu()

	// 设置窗口关闭回调
	window.SetCloseIntercept(func() {
		w.Cleanup()
		window.Close()
	})

	return w
}

// setupMainMenu 设置主菜单
func (w *MainWindow) setupMainMenu() {
	// 连接菜单
	connectMenu := fyne.NewMenu("连接",
		fyne.NewMenuItem(config.ConnectBtnNameOpen, w.onConnectClick),
	)

	// 视图菜单
	viewMenu := fyne.NewMenu("视图",
		fyne.NewMenuItem(config.FullScreen, w.onFullScreenClick),
		fyne.NewMenuItem(config.ShowPerfStats, w.togglePerformanceStats),
	)

	// 设置菜单
	screenConfig := config.GetConfig().ScreenConfig

	// 画面质量子菜单
	qualityItems := make([]*fyne.MenuItem, 0)
	for _, setting := range screenConfig.QualityList {
		qualityName := setting.Name
		qualityItems = append(qualityItems,
			fyne.NewMenuItem(qualityName, func() { w.onQualityChanged(qualityName) }),
		)
	}

	// 帧率子菜单
	fpsItems := make([]*fyne.MenuItem, 0)
	for _, fps := range screenConfig.FrameRates {
		fpsStr := strconv.Itoa(fps)
		fpsItems = append(fpsItems,
			fyne.NewMenuItem(fpsStr, func() { w.onFPSChanged(fpsStr) }),
		)
	}

	// 设置子菜单
	settingsItems := []*fyne.MenuItem{
		fyne.NewMenuItem("模式切换", w.onModeChanged),
	}

	// 创建主菜单
	w.mainMenu = fyne.NewMainMenu(
		connectMenu,
		viewMenu,
		fyne.NewMenu("画面质量", qualityItems...),
		fyne.NewMenu("帧率", fpsItems...),
		fyne.NewMenu("设置", settingsItems...),
	)

	// 设置窗口内容
	w.window.SetMainMenu(w.mainMenu)
}

// Cleanup 清理资源
func (w *MainWindow) Cleanup() {
	if w.mainUI != nil {
		w.mainUI.Cleanup()
	}
}
