package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"p2p_remote_desk/client/config"
)

type CustomSliderToolbarItem struct {
	widget.BaseWidget
	OnActivated func(float64) `json:"-"`
	slider      *widget.Slider
}

func NewCustomSliderToolbarItem(onActivated func(float64)) *CustomSliderToolbarItem {
	slider := widget.NewSlider(config.SliderClose, config.SliderOpen)
	slider.Resize(config.ToolbarItemSize)
	slider.Step = config.SliderStep
	slider.Value = config.SliderClose

	item := &CustomSliderToolbarItem{slider: slider, OnActivated: onActivated}
	return item
}

func (c *CustomSliderToolbarItem) ToolbarObject() fyne.CanvasObject {
	c.slider.OnChanged = c.OnActivated
	return c.slider
}
