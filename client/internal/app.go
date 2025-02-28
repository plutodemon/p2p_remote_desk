package internal

import (
	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal/window"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/plutodemon/llog"
)

// App 应用程序结构体
type App struct {
	FyneApp      fyne.App
	MainWindow   *window.MainWindow
	LoginWindow  *window.LoginWindow
	DeviceWindow *window.DeviceWindow
}

// NewApp 创建新的应用程序实例
func NewApp() *App {
	return &App{}
}

// Run 运行应用程序
func (a *App) Run() {
	a.FyneApp = app.NewWithID("remote_desk")

	// 创建登录窗口
	loginWindow := a.FyneApp.NewWindow("登录页")
	a.LoginWindow = window.NewLoginWindow(loginWindow, func() {
		loginWindow.Hide()
		a.showDeviceWindow("root")
	})

	// 显示登录窗口
	loginWindow.Show()
	a.FyneApp.Run()
}

// showDeviceWindow 显示设备管理窗口
func (a *App) showDeviceWindow(username string) {
	// 创建设备管理窗口
	deviceWindow := a.FyneApp.NewWindow("设备管理")
	deviceWindow.Resize(config.WindowSize)

	// 创建设备管理窗口管理器
	a.DeviceWindow = window.NewDeviceWindow(deviceWindow, username, func(deviceID string) {
		deviceWindow.Hide()
		a.showMainWindow()
	})

	// 显示设备管理窗口
	deviceWindow.Show()
	deviceWindow.CenterOnScreen()
	deviceWindow.SetMaster()
}

// showMainWindow 显示主窗口
func (a *App) showMainWindow() {
	// 创建主窗口
	mainWindow := a.FyneApp.NewWindow("远程桌面")
	mainWindow.Resize(config.WindowSize)

	// 创建主窗口管理器
	a.MainWindow = window.NewMainWindow(mainWindow)

	// 显示主窗口
	mainWindow.Show()
	mainWindow.CenterOnScreen()
	mainWindow.SetMaster()
}

// Cleanup 清理资源
func (a *App) Cleanup() {
	if a.MainWindow != nil {
		a.MainWindow.Cleanup()
	}
	llog.Cleanup()
}
