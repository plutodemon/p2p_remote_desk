package window

import (
	"fyne.io/fyne/v2"
	"p2p_remote_desk/client/internal/ui"
)

// MainWindow 主窗口管理器
type MainWindow struct {
	window fyne.Window
	mainUI *ui.MainUI
}

// NewMainWindow 创建主窗口管理器
func NewMainWindow(window fyne.Window) *MainWindow {
	w := &MainWindow{
		window: window,
	}

	// 创建主UI
	w.mainUI = ui.NewMainUI(window)
	window.SetContent(w.mainUI.Container)

	// 设置窗口关闭回调
	window.SetCloseIntercept(func() {
		w.Cleanup()
		window.Close()
	})

	return w
}

// Cleanup 清理资源
func (w *MainWindow) Cleanup() {
	if w.mainUI != nil {
		w.mainUI.Cleanup()
	}
}
