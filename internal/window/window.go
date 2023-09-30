package window

import (
	"context"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotkit/app"
	"github.com/diamondburned/gotkit/gtkutil/cssutil"
	"github.com/thekrafter/gtkcord-spacebar/internal/window/login"
)

// SetPreferDarkTheme sets whether or not GTK should use a dark theme.
func SetPreferDarkTheme(prefer bool) {
	settings := gtk.SettingsGetDefault()
	settings.SetObjectProperty("gtk-application-prefer-dark-theme", prefer)

	scheme := adw.ColorSchemePreferLight
	if prefer {
		scheme = adw.ColorSchemePreferDark
	}

	adwStyles := adw.StyleManagerGetDefault()
	adwStyles.SetColorScheme(scheme)
}

var _ = cssutil.WriteCSS(`
	.titlebar {
		background-color: @headerbar_bg_color;
	}

	window.devel .titlebar {
		background-image: cross-fade(
			5% -gtk-recolor(url("resource:/org/gnome/Adwaita/styles/assets/devel-symbolic.svg")),
			image(transparent));
		background-repeat: repeat-x;
	}
`)

// Window is the main gtkcord window.
type Window struct {
	*app.Window
	ctx context.Context

	Stack   *gtk.Stack
	Login   *login.Page
	Loading *login.LoadingPage
	Chat    *ChatPage
}

// NewWindow creates a new Window.
func NewWindow(ctx context.Context) *Window {
	win := app.FromContext(ctx).NewWindow()
	win.SetTitle("")
	win.SetTitlebar(newEmptyHeader())
	win.SetDefaultSize(800, 600)

	ctx = app.WithWindow(ctx, win)

	w := Window{
		Window: win,
		ctx:    ctx,
	}

	w.Login = login.NewPage(ctx, (*loginWindow)(&w))
	w.Login.LoadKeyring()

	w.Loading = login.NewLoadingPage(ctx)

	w.Stack = gtk.NewStack()
	w.Stack.SetTransitionType(gtk.StackTransitionTypeCrossfade)
	w.Stack.AddChild(w.Login)
	w.Stack.AddChild(w.Loading)
	w.Stack.SetVisibleChild(w.Login)

	win.SetChild(w.Stack)
	return &w
}

func (w *Window) Context() context.Context {
	return w.ctx
}

func (w *Window) SwitchToChatPage() {
	if w.Chat == nil {
		w.Chat = NewChatPage(w.ctx)
		w.Stack.AddChild(w.Chat)
	}
	w.Stack.SetVisibleChild(w.Chat)
	w.Chat.SwitchToMessages()
	w.SetTitle("")
}

func (w *Window) SwitchToLoginPage() {
	w.Stack.SetVisibleChild(w.Login)
	w.SetTitle("Login")
}

var emptyHeaderCSS = cssutil.Applier("empty-header", `
	.empty-header {
		min-height: 0;
		min-width: 0;
		padding: 0;
		margin: 0;
		border: 0;
	}
`)

func newEmptyHeader() *gtk.Box {
	b := gtk.NewBox(gtk.OrientationVertical, 0)
	emptyHeaderCSS(b)
	return b
}
