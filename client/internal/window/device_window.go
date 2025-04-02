package window

import (
	"image/color"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal/network"
	"p2p_remote_desk/common"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type DeviceWindow struct {
	Window           fyne.Window
	onDeviceSelected func(device *DeviceInfo) // 设备选择回调函数
	onLogout         func()                   // 注销回调函数
	username         string                   // 当前登录的用户名
	deviceCard       *widget.Card             // 设备卡片列表
	devices          []*DeviceInfo            // 设备信息列表
}

type DeviceInfo struct {
	ID       string // 设备ID
	Name     string // 设备名称
	IP       uint32 // 设备IP
	IsOnline bool   // 是否在线
}

func (newApp *App) newDeviceWindow(username string) {
	newApp.deviceWindow = &DeviceWindow{
		username: username,
		devices:  make([]*DeviceInfo, 0),
	}

	newApp.deviceWindow.Window = newApp.fyneApp.NewWindow("设备管理")

	newApp.deviceWindow.onDeviceSelected = func(device *DeviceInfo) {
		newApp.confirmDialog(device)
	}

	newApp.deviceWindow.onLogout = func() {
		newApp.deviceWindow.Window.Hide()
		newApp.loginWindow.Window.Show()
	}

	// 创建设备管理界面
	newApp.deviceWindow.setupUI()

	// 加载设备列表
	newApp.deviceWindow.loadDevices()

	newApp.deviceWindow.Window.SetOnClosed(func() {
	})

	newApp.deviceWindow.Window.Show()
}

func (newApp *App) confirmDialog(device *DeviceInfo) {
	callback := func(conn bool) {
		if !conn {
			return
		}
		newApp.deviceWindow.Window.Hide()
		newApp.newMainWindow(device)
	}

	confirm := dialog.NewConfirm("连接确认", "是否连接到设备: "+device.Name, callback, newApp.deviceWindow.Window)
	confirm.SetConfirmText("连接")
	confirm.SetDismissText("取消")
	confirm.Resize(config.DialogSize)
	confirm.Show()
}

func (w *DeviceWindow) setupUI() {
	// 个人信息
	title := widget.NewLabel("设备管理 - " + w.username)
	title.TextStyle = fyne.TextStyle{Bold: true}

	self := container.NewCenter(
		title,
		canvas.NewRadialGradient(color.Gray16{Y: 0xffff}, nil),
	)

	// 设备列表
	refreshBtn := widget.NewButton("刷新设备列表", func() {
		w.loadDevices()
	})

	refresh := container.NewVBox(
		canvas.NewRadialGradient(color.Gray16{Y: 0xffff}, nil),
		refreshBtn,
	)

	w.deviceCard = widget.NewCard("", "", nil)

	device := container.NewBorder(nil, refresh, nil, nil, w.deviceCard)

	// 设置页面
	logoutBtn := widget.NewButton("退出登录", func() {
		w.onLogout()
	})

	logout := container.NewVBox(
		canvas.NewRadialGradient(color.Gray16{Y: 0xffff}, nil),
		logoutBtn,
	)

	logoutCon := container.NewBorder(nil, logout, nil, nil, widget.NewLabel("设置页面"))

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("个人信息", theme.AccountIcon(), self),
		container.NewTabItemWithIcon("设备列表", theme.ListIcon(), device),
		container.NewTabItemWithIcon("设置", theme.SettingsIcon(), logoutCon),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	// 设置窗口内容
	w.Window.SetContent(tabs)
	w.Window.Resize(config.WindowSize)
	w.Window.CenterOnScreen()
	w.Window.SetMaster()
}

func (w *DeviceWindow) loadDevices() {
	// 先清空设备列表
	w.devices = make([]*DeviceInfo, 0)

	network.Clients.Range(func(key, value any) bool {
		client := value.(common.ClientInfo)
		w.devices = append(w.devices, &DeviceInfo{
			ID:       client.UID,
			Name:     "设备" + client.Name,
			IP:       client.IP,
			IsOnline: true,
		})
		return true
	})

	if config.IsDevelopment() && len(w.devices) == 0 {
		w.devices = []*DeviceInfo{
			{ID: "device1", Name: "我的电脑", IP: 111, IsOnline: true},
			{ID: "device2", Name: "办公室电脑", IP: 222, IsOnline: false},
			{ID: "device3", Name: "家里的笔记本", IP: 333, IsOnline: true},
		}
	}

	w.refreshDeviceList()
}

func (w *DeviceWindow) refreshDeviceList() {
	// 创建设备列表
	cardList := make([]fyne.CanvasObject, 0)
	for _, device := range w.devices {
		button := widget.NewButton("连接", w.buttonTapped(device))
		if !device.IsOnline {
			button.Disable()
		}
		con := container.NewHBox(layout.NewSpacer(), button)
		card := widget.NewCard("", device.Name+": "+lkit.AnyToStr(device.IP), con)
		cardList = append(cardList, card)
	}

	w.deviceCard.SetContent(container.NewScroll(container.NewVBox(cardList...)))
	w.deviceCard.Refresh()
}

func (w *DeviceWindow) buttonTapped(device *DeviceInfo) func() {
	if !device.IsOnline {
		return func() {
			ShowError(w.Window, "该设备当前不在线, 无法连接")
		}
	}
	return func() {
		llog.Info("选择设备: %s", device.Name)
		if w.onDeviceSelected != nil {
			w.onDeviceSelected(device)
		}
	}
}
