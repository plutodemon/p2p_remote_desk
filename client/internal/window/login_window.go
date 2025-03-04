package window

import (
	"p2p_remote_desk/client/config"
	"p2p_remote_desk/client/internal/network/auth"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/plutodemon/llog"
)

type LoginWindow struct {
	Window     fyne.Window
	onLoggedIn func(userName string)
}

func NewLoginWindow(window fyne.Window, onLoggedIn func(userName string)) *LoginWindow {
	w := &LoginWindow{
		Window:     window,
		onLoggedIn: onLoggedIn,
	}

	w.setupUI()

	return w
}

func (w *LoginWindow) setupUI() {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("用户名")
	usernameEntry.Resize(config.LoginEntrySize)

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("密码")
	passwordEntry.Resize(config.LoginEntrySize)

	loginBtn := widget.NewButton("登录", func() {
		w.handleLogin(usernameEntry.Text, passwordEntry.Text)
	})
	loginBtn.Importance = widget.HighImportance
	loginBtn.Resize(config.LoginButtonSize)
	loginBtn.Refresh()

	vBox := container.NewVBox(
		widget.NewLabel("远程桌面登录"),
		usernameEntry,
		passwordEntry,
		loginBtn,
	)

	// 如果是开发环境，添加离线登录按钮
	if config.IsDevelopment() {
		offlineBtn := widget.NewButton("离线登录", func() {
			llog.Debug("使用离线模式登录")
			if w.onLoggedIn != nil {
				w.onLoggedIn("离线模式")
			}
		})
		offlineBtn.Resize(config.LoginButtonSize)
		offlineBtn.Importance = widget.LowImportance
		vBox.Add(offlineBtn)
	}

	// 设置窗口内容
	content := container.NewCenter(vBox)
	w.Window.SetContent(content)
	w.Window.Resize(config.WindowSize)
	w.Window.CenterOnScreen()
	w.Window.SetMaster()
	w.Window.SetFixedSize(true)
}

func (w *LoginWindow) handleLogin(username, password string) {
	if username == "" || password == "" {
		ShowError(w.Window, "用户名和密码不能为空")
		return
	}

	// 验证用户名和密码
	code := auth.LoginAuth(username, password)
	if code == 0 {
		llog.Info("用户登录成功: %s", username)
		if w.onLoggedIn != nil {
			w.onLoggedIn(username)
		}
	} else {
		ShowError(w.Window, "用户名或密码错误")
	}
}
