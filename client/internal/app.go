package internal

import (
	"fyne.io/fyne/v2/dialog"
	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal/window"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// App 应用程序结构体
type App struct {
	FyneApp      fyne.App
	LoginWindow  *window.LoginWindow
	DeviceWindow *window.DeviceWindow
	MainWindow   *window.MainWindow
}

func NewAppAndRun() {
	newApp := &App{
		FyneApp: app.NewWithID("remote_desk"),
	}

	newApp.loginNewAndRun()

	newApp.FyneApp.Run()
}

func (a *App) loginNewAndRun() {
	loginWindow := a.FyneApp.NewWindow("登录")

	a.LoginWindow = window.NewLoginWindow(loginWindow, func(userName string) {
		loginWindow.Hide()
		a.deviceNewAndRun(userName)
	})

	loginWindow.Show()
}

// 设备管理窗口
func (a *App) deviceNewAndRun(username string) {
	// 创建设备管理窗口
	deviceWindow := a.FyneApp.NewWindow("设备管理")

	// 创建设备管理窗口管理器
	a.DeviceWindow = window.NewDeviceWindow(deviceWindow, username, func(device *window.DeviceInfo) {
		a.confirmDialog(device)
	})

	// 显示设备管理窗口
	deviceWindow.Show()
}

// 确认连接dialog
func (a *App) confirmDialog(device *window.DeviceInfo) {
	confirm := dialog.NewConfirm("连接确认", "是否连接到设备: "+device.Name, func(flag bool) {
		if flag {
			a.DeviceWindow.Window.Hide()
			a.mainNewAndRun(device)
		}
	}, a.DeviceWindow.Window)
	confirm.SetConfirmText("连接")
	confirm.SetDismissText("取消")
	confirm.Resize(config.DialogSize)
	confirm.Show()
}

func (a *App) mainNewAndRun(device *window.DeviceInfo) {
	// 创建主窗口
	mainWindow := a.FyneApp.NewWindow("远程桌面: " + device.Name)

	// 创建主窗口管理器
	a.MainWindow = window.NewMainWindow(mainWindow)

	// 显示主窗口
	mainWindow.Show()
}
