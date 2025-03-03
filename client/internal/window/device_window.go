package window

import (
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"image/color"
	"p2p_remote_desk/client/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/plutodemon/llog"
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
			return container.NewHBox(
				widget.NewIcon(nil),
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			hBox := obj.(*fyne.Container)
			device := w.devices[id]

			// 设备图标
			hBox.Objects[0].(*widget.Icon).SetResource(theme.ComputerIcon())
			// 设备名称
			nameLabel := hBox.Objects[1].(*widget.Label)
			nameLabel.SetText(device.Name)

			// 设备状态
			statusLabel := hBox.Objects[2].(*widget.Label)
			if device.IsOnline {
				statusLabel.SetText("[在线]")
			} else {
				statusLabel.SetText("[离线]")
			}
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
	// TODO: 与信令服务器通信，获取设备列表
	// 这里先使用模拟数据
	w.devices = []*DeviceInfo{
		{ID: "device1", Name: "我的电脑", IsOnline: true},
		{ID: "device2", Name: "办公室电脑", IsOnline: false},
		{ID: "device3", Name: "家里的笔记本", IsOnline: true},
	}

	// 刷新列表显示
	w.deviceList.Refresh()
}
