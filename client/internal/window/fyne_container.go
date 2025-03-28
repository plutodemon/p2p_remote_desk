package window

import (
	"fyne.io/fyne/v2"
)

type AspectRatioLayout struct {
	Ratio float64 // 例如 4/3
}

func (a *AspectRatioLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	// 根据容器大小计算目标尺寸，保证比例不变
	width := containerSize.Width
	height := containerSize.Height
	if float64(width)/float64(height) > a.Ratio {
		width = float32(float64(height) * a.Ratio)
	} else {
		height = float32(float64(width) / a.Ratio)
	}
	// 将所有对象置于容器中央并设置相同尺寸
	offsetX := (containerSize.Width - width) / 2
	offsetY := (containerSize.Height - height) / 2
	for _, o := range objects {
		o.Resize(fyne.NewSize(width, height))
		o.Move(fyne.NewPos(offsetX, offsetY))
	}
}

func (a *AspectRatioLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	// 返回一个最小尺寸 16:9
	return fyne.NewSize(960, 540)
}
