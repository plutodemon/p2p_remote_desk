package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/go-gl/glfw/v3.3/glfw"
	"p2p_remote_desk/config"
	"p2p_remote_desk/internal/component"
	"p2p_remote_desk/kit"
	"runtime"
)

// ToolbarUI 工具栏组件
type ToolbarUI struct {
	Toolbar *widget.Toolbar

	// 模式选择
	ModeSelect   *component.CustomSliderToolbarItem
	isController bool // 是否为控制端
	ModeState    *component.CustomLabelToolbarItem

	// 连接按钮
	ConnectBtn  *component.CustomButtonToolbarItem
	isConnected bool // 是否已连接
	StatusLabel *component.CustomLabelToolbarItem

	// 全屏按钮
	FullScreenBtn *component.CustomButtonToolbarItem
	isFullScreen  bool // 是否全屏

	// 监控按钮
	PerfStatsBtn *component.CustomButtonToolbarItem
	isShowStats  bool // 是否显示监控

	// 显示选择
	DisplaySelect *component.CustomRadioToolbarItem
	QualitySelect *component.CustomSelectToolbarItem
	FpsSelect     *component.CustomSelectToolbarItem

	// 帧率显示
	FpsLabel *component.CustomLabelToolbarItem

	createDisplay bool // 是否创建显示选择
	isShowToolbar bool // 工具栏是否可见

}

// NewToolbar 创建工具栏
func (ui *MainUI) NewToolbar() {
	ui.toolbar = &ToolbarUI{
		Toolbar: widget.NewToolbar(),
	}

	ui.initToolBarUI()

	ui.toolbar.Toolbar.Resize(fyne.NewSize(20, 5))

	ui.toolbar.Toolbar.Append(ui.toolbar.ModeSelect)
	ui.toolbar.Toolbar.Append(ui.toolbar.ModeState)
	ui.toolbar.Toolbar.Append(widget.NewToolbarSeparator())

	ui.toolbar.Toolbar.Append(ui.toolbar.ConnectBtn)
	ui.toolbar.Toolbar.Append(ui.toolbar.StatusLabel)
	ui.toolbar.Toolbar.Append(widget.NewToolbarSeparator())

	ui.toolbar.Toolbar.Append(ui.toolbar.FullScreenBtn)
	ui.toolbar.Toolbar.Append(ui.toolbar.PerfStatsBtn)
	ui.toolbar.Toolbar.Append(widget.NewToolbarSeparator())

	ui.toolbar.Toolbar.Append(ui.toolbar.QualitySelect)
	ui.toolbar.Toolbar.Append(ui.toolbar.FpsSelect)
	ui.toolbar.Toolbar.Append(ui.toolbar.DisplaySelect)
	ui.toolbar.Toolbar.Append(widget.NewToolbarSpacer())

	ui.toolbar.Toolbar.Append(ui.toolbar.FpsLabel)
}

func (ui *MainUI) initToolBarUI() {
	ui.toolbar.isShowToolbar = true

	ui.toolbar.StatusLabel = component.NewCustomLabelToolbarItem("就绪")
	ui.toolbar.ConnectBtn = component.NewCustomButtonToolbarItem(config.ConnectBtnNameOpen, ui.onConnectClick)

	ui.toolbar.FullScreenBtn = component.NewCustomButtonToolbarItem("全屏", ui.onFullScreenClick)

	disPlayList := make([]string, 0)
	switch runtime.GOOS {
	case "windows", "darwin", "linux":
		monitors := glfw.GetMonitors()
		if len(monitors) == 1 {
			break
		}
		for index, monitor := range monitors {
			disPlayList = append(disPlayList, monitor.GetName()+"("+kit.AnyToStr(index)+")")
		}
		ui.toolbar.createDisplay = true
	default:

	}
	if ui.toolbar.createDisplay {
		ui.toolbar.DisplaySelect = component.NewCustomRadioToolbarItem(disPlayList, ui.onDisplayChanged)
		ui.toolbar.DisplaySelect.Radio.Selected = disPlayList[0]
	}

	// 创建性能监控按钮
	ui.toolbar.PerfStatsBtn = component.NewCustomButtonToolbarItem("隐藏监控", ui.togglePerformanceStats)

	// 创建质量选择
	screenConfig := config.GetConfig().ScreenConfig
	qualityList := make([]string, 0, len(screenConfig.QualityList))
	for _, setting := range screenConfig.QualityList {
		qualityList = append(qualityList, setting.Name)
	}
	ui.toolbar.QualitySelect = component.NewCustomSelectToolbarItem(qualityList, ui.onQualityChanged)
	ui.toolbar.QualitySelect.Select.SetSelected(screenConfig.DefaultQuality)

	// 创建帧率选择
	ui.toolbar.FpsSelect = component.NewCustomSelectToolbarItem(kit.SliceToStrList(screenConfig.FrameRates), ui.onFPSChanged)
	ui.toolbar.FpsSelect.Select.SetSelected(kit.AnyToStr(screenConfig.DefaultFrameRate))

	ui.toolbar.FpsLabel = component.NewCustomLabelToolbarItem("FPS: 0")

	// 创建模式选择
	ui.toolbar.ModeSelect = component.NewCustomSliderToolbarItem(ui.onModeChanged)
	ui.toolbar.ModeState = component.NewCustomLabelToolbarItem("被控端")
}

func (t *ToolbarUI) SetStatus(status string) {
	t.StatusLabel.Label.SetText(status)
}

func (t *ToolbarUI) SetFPS(fps float64) {
	t.FpsLabel.Label.SetText(fmt.Sprintf("FPS: %.1f", fps))
}
