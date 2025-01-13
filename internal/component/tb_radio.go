package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"p2p_remote_desk/config"
)

type CustomRadioToolbarItem struct {
	widget.BaseWidget
	OnActivated func(string) `json:"-"`
	Radio       *widget.RadioGroup
}

func NewCustomRadioToolbarItem(options []string, changed func(string)) *CustomRadioToolbarItem {
	Radio := widget.NewRadioGroup(options, changed)
	Radio.Resize(config.ToolbarItemSize)

	item := &CustomRadioToolbarItem{Radio: Radio, OnActivated: changed}
	return item
}

func (c *CustomRadioToolbarItem) ToolbarObject() fyne.CanvasObject {
	c.Radio.OnChanged = c.OnActivated
	return c.Radio
}
