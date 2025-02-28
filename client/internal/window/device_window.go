package window

import (
	"time"

	"p2p_remote_desk/client/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/plutodemon/llog"
)

// DeviceWindow 设备管理窗口
type DeviceWindow struct {
	window           fyne.Window
	onDeviceSelected func(deviceID string) // 设备选择回调函数
	username         string                // 当前登录的用户名
	deviceList       *widget.List          // 设备列表组件
	devices          []DeviceInfo          // 设备信息列表
}

// DeviceInfo 设备信息结构体
type DeviceInfo struct {
	ID       string // 设备ID
	Name     string // 设备名称
	IsOnline bool   // 是否在线
}

// NewDeviceWindow 创建设备管理窗口
func NewDeviceWindow(window fyne.Window, username string, onDeviceSelected func(deviceID string)) *DeviceWindow {
	w := &DeviceWindow{
		window:           window,
		username:         username,
		onDeviceSelected: onDeviceSelected,
		devices:          make([]DeviceInfo, 0),
	}

	// 创建设备管理界面
	w.setupUI()

	// 加载设备列表
	w.loadDevices()

	return w
}

// setupUI 设置设备管理界面
func (w *DeviceWindow) setupUI() {
	// 创建设备列表
	w.deviceList = widget.NewList(
		func() int {
			return len(w.devices)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			container := obj.(*fyne.Container)
			device := w.devices[id]

			// 设备名称
			nameLabel := container.Objects[0].(*widget.Label)
			nameLabel.SetText(device.Name)

			// 设备状态
			statusLabel := container.Objects[1].(*widget.Label)
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
				w.onDeviceSelected(device.ID)
			}
		} else {
			w.showError("该设备当前不在线，无法连接")
			w.deviceList.UnselectAll()
		}
	}

	// 创建刷新按钮
	refreshBtn := widget.NewButton("刷新设备列表", func() {
		w.loadDevices()
	})

	// 创建标题
	title := widget.NewLabel("设备管理 - " + w.username)
	title.TextStyle = fyne.TextStyle{Bold: true}

	// 创建布局
	content := container.NewBorder(
		container.NewVBox(
			title,
			widget.NewSeparator(),
		),
		container.NewVBox(
			refreshBtn,
		),
		nil,
		nil,
		w.deviceList,
	)

	// 设置窗口内容
	w.window.SetContent(content)
	w.window.Resize(config.WindowSize)
	w.window.CenterOnScreen()
	w.window.SetMaster()
}

// loadDevices 加载设备列表
func (w *DeviceWindow) loadDevices() {
	// TODO: 与信令服务器通信，获取设备列表
	// 这里先使用模拟数据
	w.devices = []DeviceInfo{
		{ID: "device1", Name: "我的电脑", IsOnline: true},
		{ID: "device2", Name: "办公室电脑", IsOnline: false},
		{ID: "device3", Name: "家里的笔记本", IsOnline: true},
	}

	// 刷新列表显示
	w.deviceList.Refresh()
}

// showError 显示错误信息
func (w *DeviceWindow) showError(message string) {
	dialog := widget.NewLabel(message)
	popup := widget.NewModalPopUp(dialog, w.window.Canvas())
	popup.Show()

	// 2秒后自动关闭错误提示
	go func() {
		time.Sleep(2 * time.Second)
		popup.Hide()
	}()
}
