package window

import (
	"image/color"

	"p2p_remote_desk/client/config"
	"p2p_remote_desk/lkit"
	"p2p_remote_desk/llog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type DeviceWindow struct {
	Window           fyne.Window
	onDeviceSelected func(device *DeviceInfo) // 设备选择回调函数
	username         string                   // 当前登录的用户名
	deviceList       *widget.List             // 设备列表组件
	devices          []*DeviceInfo            // 设备信息列表
}

type DeviceInfo struct {
	ID       string // 设备ID
	Name     string // 设备名称
	IP       uint32 // 设备IP
	IsOnline bool   // 是否在线
}

func NewDeviceWindow(window fyne.Window, username string, onDeviceSelected func(device *DeviceInfo)) *DeviceWindow {
	w := &DeviceWindow{
		Window:           window,
		username:         username,
		onDeviceSelected: onDeviceSelected,
		devices:          make([]*DeviceInfo, 0),
	}

	// 创建设备管理界面
	w.setupUI()

	// 加载设备列表
	w.loadDevices()

	return w
}

func (w *DeviceWindow) setupUI() {
	// 创建设备列表
	w.deviceList = widget.NewList(
		func() int {
			return len(w.devices)
		},
		func() fyne.CanvasObject {
			info := container.NewCenter(
				container.NewHBox(
					widget.NewLabel(""),
					widget.NewLabel(""),
				),
			)
			return container.NewBorder(nil, nil,
				widget.NewIcon(nil),
				widget.NewIcon(nil),
				info,
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			border := obj.(*fyne.Container)
			device := w.devices[id]

			// 设备名称
			labelList := border.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container)
			nameLabel := labelList.Objects[0].(*widget.Label)
			nameLabel.SetText(device.Name)

			// 设备ip
			ipLabel := labelList.Objects[1].(*widget.Label)
			ipLabel.SetText(lkit.AnyToStr(device.IP))

			// 设备图标
			border.Objects[1].(*widget.Icon).SetResource(theme.ComputerIcon())

			// 设备状态
			var res fyne.Resource
			if device.IsOnline {
				res = theme.RadioButtonCheckedIcon()
			} else {
				res = theme.RadioButtonIcon()
			}
			border.Objects[2].(*widget.Icon).SetResource(res)
		},
	)

	// 设置列表选择事件
	w.deviceList.OnSelected = func(id widget.ListItemID) {
		device := w.devices[id]
		if device.IsOnline {
			llog.Info("选择设备: %s", device.Name)
			if w.onDeviceSelected != nil {
				w.onDeviceSelected(device)
			}
		} else {
			ShowError(w.Window, "该设备当前不在线, 无法连接")
			w.deviceList.UnselectAll()
		}
	}

	// 创建刷新按钮
	refreshBtn := widget.NewButton("刷新设备列表", func() {
		w.loadDevices()
	})

	title := widget.NewLabel("设备管理 - " + w.username)
	title.TextStyle = fyne.TextStyle{Bold: true}

	self := container.NewCenter(
		title,
		canvas.NewRadialGradient(color.Gray16{Y: 0xffff}, nil),
	)

	refresh := container.NewVBox(
		canvas.NewRadialGradient(color.Gray16{Y: 0xffff}, nil),
		refreshBtn,
	)

	device := container.NewBorder(
		nil,
		refresh,
		nil,
		nil,
		w.deviceList,
	)

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("个人信息", theme.AccountIcon(), self),
		container.NewTabItemWithIcon("设备列表", theme.ListIcon(), device),
		container.NewTabItemWithIcon("设置", theme.SettingsIcon(), widget.NewLabel("设置页面")),
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

	//network.Clients.Range(func(key, value any) bool {
	//	client := value.(common.ClientInfo)
	//	w.devices = append(w.devices, &DeviceInfo{
	//		ID:       client.UID,
	//		Name:     "设备" + client.Name,
	//		IP:       client.IP,
	//		IsOnline: true,
	//	})
	//	return true
	//})

	w.devices = []*DeviceInfo{
		{ID: "device1", Name: "我的电脑", IP: 111, IsOnline: true},
		{ID: "device2", Name: "办公室电脑", IP: 222, IsOnline: false},
		{ID: "device3", Name: "家里的笔记本", IP: 333, IsOnline: true},
	}

	// 刷新列表显示
	w.deviceList.UnselectAll()
	w.deviceList.Refresh()
}
