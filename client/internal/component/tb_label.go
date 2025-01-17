package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"p2p_remote_desk/client/config"
)

type CustomLabelToolbarItem struct {
	widget.BaseWidget
	Label *widget.Label
}

func NewCustomLabelToolbarItem(text string) *CustomLabelToolbarItem {
	Label := widget.NewLabel(text)
	Label.Resize(config.ToolbarItemSize)

	item := &CustomLabelToolbarItem{Label: Label}
	return item
}

func (c *CustomLabelToolbarItem) ToolbarObject() fyne.CanvasObject {
	return c.Label
}
