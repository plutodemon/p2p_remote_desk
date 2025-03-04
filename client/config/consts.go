package config

import "fyne.io/fyne/v2"

const (
	Development = "development" // 开发环境
	Production  = "production"  // 生产环境
)

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

	RestoreDefault = "恢复默认"
	RestoreNormal  = "恢复普通"
)

var ToolbarSize = fyne.NewSize(ToolbarItemWide, ToolbarItemHigh)
var ToolbarItemSize = fyne.NewSize(ToolbarItemWide, ToolbarItemHigh)
var LoginEntrySize = fyne.NewSize(320, 40)
var LoginButtonSize = fyne.NewSize(200, 40)
var DialogSize = fyne.NewSize(420, 160)

const (
	ToolbarItemWide = 10
	ToolbarItemHigh = 5
)

var WindowSize = fyne.NewSize(960, 550)
