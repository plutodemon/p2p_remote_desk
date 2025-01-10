package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type CustomButtonToolbarItem struct {
	widget.BaseWidget
	OnActivated func() `json:"-"`
	Button      *widget.Button
}

func NewCustomButtonToolbarItem(name string, onActivated func()) *CustomButtonToolbarItem {
	Button := widget.NewButton(name, onActivated)
	Button.Resize(fyne.NewSize(5, 10))

	item := &CustomButtonToolbarItem{Button: Button, OnActivated: onActivated}
	return item
}

func (c *CustomButtonToolbarItem) ToolbarObject() fyne.CanvasObject {
	c.Button.OnTapped = c.OnActivated
	return c.Button
}
