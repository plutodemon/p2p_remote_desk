package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type CustomSelectToolbarItem struct {
	widget.BaseWidget
	OnActivated func(string) `json:"-"`
	Select      *widget.Select
}

func NewCustomSelectToolbarItem(options []string, changed func(string)) *CustomSelectToolbarItem {
	Select := widget.NewSelect(options, changed)
	Select.Resize(fyne.NewSize(5, 10))
	Select.Selected = options[0]
	item := &CustomSelectToolbarItem{Select: Select, OnActivated: changed}
	return item
}

func (c *CustomSelectToolbarItem) ToolbarObject() fyne.CanvasObject {
	c.Select.OnChanged = c.OnActivated
	return c.Select
}
