package window

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/plutodemon/slog"
	config2 "p2p_remote_desk/client/config"
	"time"
)

// LoginWindow 登录窗口管理器
type LoginWindow struct {
	window     fyne.Window
	onLoggedIn func() // 登录成功的回调函数
}

// NewLoginWindow 创建登录窗口管理器
func NewLoginWindow(window fyne.Window, onLoggedIn func()) *LoginWindow {
	w := &LoginWindow{
		window:     window,
		onLoggedIn: onLoggedIn,
	}

	// 创建登录界面
	w.setupUI()

	return w
}

// setupUI 设置登录界面
func (w *LoginWindow) setupUI() {
	newSize := fyne.NewSize(200, 40)

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("用户名")
	usernameEntry.Resize(newSize)

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("密码")
	passwordEntry.Resize(newSize)

	// 创建登录按钮
	loginBtn := widget.NewButton("登录", func() {
		w.handleLogin(usernameEntry.Text, passwordEntry.Text)
	})
	loginBtn.Resize(newSize)
	loginBtn.Importance = widget.HighImportance

	// 创建表单布局
	form := container.NewVBox(
		widget.NewLabel("远程桌面登录"),
		usernameEntry,
		passwordEntry,
		loginBtn,
	)

	// 如果是开发环境，添加离线登录按钮
	if config2.IsDevelopment() {
		offlineBtn := widget.NewButton("离线登录", func() {
			slog.Info("使用离线模式登录")
			if w.onLoggedIn != nil {
				w.onLoggedIn()
			}
		})
		offlineBtn.Resize(newSize)
		offlineBtn.Importance = widget.LowImportance
		form.Add(offlineBtn)
	}

	// 设置窗口内容
	content := container.NewCenter(form)
	w.window.SetContent(content)
	w.window.Resize(config2.WindowSize)
	w.window.CenterOnScreen()
	w.window.SetMaster()
}

// handleLogin 处理登录逻辑
func (w *LoginWindow) handleLogin(username, password string) {
	if username == "" || password == "" {
		w.showError("用户名和密码不能为空")
		return
	}

	// 验证用户名和密码
	if username == "root" && password == "123" {
		slog.Info("用户登录成功: %s", username)
		if w.onLoggedIn != nil {
			w.onLoggedIn()
		}
	} else {
		w.showError("用户名或密码错误")
	}
}

// showError 显示错误信息
func (w *LoginWindow) showError(message string) {
	dialog := widget.NewLabel(message)
	popup := widget.NewModalPopUp(dialog, w.window.Canvas())
	popup.Show()

	// 2秒后自动关闭错误提示
	go func() {
		time.Sleep(2 * time.Second)
		popup.Hide()
	}()
}
