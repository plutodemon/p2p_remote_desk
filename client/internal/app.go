package internal

import (
	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal/window"
	"p2p_remote_desk/llog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
)

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

	// 设置图标
	icon, err := fyne.LoadResourceFromPath("../png/icon.png")
	if err != nil {
		llog.Fatal("加载图标失败: ", err)
	}
	// 设置应用图标
	newApp.FyneApp.SetIcon(icon)

	// 设置系统托盘菜单
	if desk, ok := newApp.FyneApp.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show", func() {
				llog.Debug("Show")
			}))
		desk.SetSystemTrayMenu(m)
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

func (a *App) deviceNewAndRun(username string) {
	// 创建设备管理窗口
	deviceWindow := a.FyneApp.NewWindow("设备管理")

	// 创建设备管理窗口管理器
	a.DeviceWindow = window.NewDeviceWindow(deviceWindow, username, func(device *window.DeviceInfo) {
		a.confirmDialog(device)
	}, func() {
		a.DeviceWindow.Window.Close()
		a.LoginWindow.Window.Show()
	}, func() {
	})

	// 显示设备管理窗口
	deviceWindow.Show()
}

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
	a.MainWindow = window.NewMainWindow(mainWindow, func() {
		a.MainWindow.Cleanup()
		a.MainWindow.Window.Close()
		a.DeviceWindow.Window.Show()
	})

	a.MainWindow.OnConnectClick()

	// 显示主窗口
	mainWindow.Show()
}
