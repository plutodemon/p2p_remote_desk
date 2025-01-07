package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	windows "p2p_remote_desk/internal/ui"
	"p2p_remote_desk/logger"
)

// App 应用程序结构体
type App struct {
	fyneApp     fyne.App
	mainWindow  *windows.MainWindow
	loginWindow *windows.LoginWindow
}

// NewApp 创建新的应用程序实例
func NewApp() *App {
	return &App{}
}

// Run 运行应用程序
func (a *App) Run() {
	a.fyneApp = app.NewWithID("remote_desk")

	// 创建登录窗口
	loginWindow := a.fyneApp.NewWindow("登录页")
	a.loginWindow = windows.NewLoginWindow(loginWindow, func() {
		loginWindow.Hide()
		a.showMainWindow()
	})

	// 显示登录窗口
	loginWindow.Show()
	a.fyneApp.Run()
}

// showMainWindow 显示主窗口
func (a *App) showMainWindow() {
	// 创建主窗口
	window := a.fyneApp.NewWindow("远程桌面")
	window.Resize(fyne.NewSize(1024, 768))

	// 创建主窗口管理器
	a.mainWindow = windows.NewMainWindow(window)

	// 显示主窗口
	window.Show()
	window.CenterOnScreen()
	window.SetMaster()
}

// Cleanup 清理资源
func (a *App) Cleanup() {
	if a.mainWindow != nil {
		a.mainWindow.Cleanup()
	}
	logger.Cleanup()
}
