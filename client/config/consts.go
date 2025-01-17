package config

import "fyne.io/fyne/v2"

const (
	SliderClose = 0 // 滑块关闭
	SliderOpen  = 1 // 滑块打开
	SliderStep  = 1 // 滑块步长
)

const (
	ConnectBtnNameOpen  = "连接"
	ConnectBtnNameClose = "断开"

	HiddenPerfStats = "隐藏监控"
	ShowPerfStats   = "性能监控"

	ControlledEnd = "被控端"
	ControlEnd    = "控制端"

	FullScreen     = "全屏"
	ExitFullScreen = "退出全屏"
)

var ToolbarSize = fyne.NewSize(ToolbarItemWide, ToolbarItemHigh)
var ToolbarItemSize = fyne.NewSize(ToolbarItemWide, ToolbarItemHigh)

const (
	ToolbarItemWide = 10
	ToolbarItemHigh = 5
)

var WindowSize = fyne.NewSize(960, 550)
