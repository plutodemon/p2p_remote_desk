package window

import (
	"p2p_remote_desk/llog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
)

type App struct {
	fyneApp      fyne.App
	loginWindow  *LoginWindow
	deviceWindow *DeviceWindow
	mainWindow   *MainWindow
}

func NewAppAndRun() {
	newApp := &App{}
	newApp.fyneApp = app.NewWithID("remote_desk")

	// 设置图标
	icon, err := fyne.LoadResourceFromPath("../png/icon.png")
	if err != nil {
		llog.Fatal("加载图标失败: ", err)
	}
	// 设置应用图标
	newApp.fyneApp.SetIcon(icon)

	// 设置系统托盘菜单
	if desk, ok := newApp.fyneApp.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show", func() {
				llog.Debug("Show")
			}))
		desk.SetSystemTrayMenu(m)
	}

	// 登录界面 程序入口
	newApp.newLoginWindow()

	newApp.fyneApp.Run()
}
