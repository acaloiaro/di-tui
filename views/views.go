package views

import (
	"fmt"
	"os"
	"strings"

	"github.com/acaloiaro/di-tui/components"
	"github.com/acaloiaro/di-tui/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var primaryTextColorString string
var secondaryTextColorString string
var primaryColor tcell.Color
var backgroundColor tcell.Color
var primaryTextColor tcell.Color
var secondaryTextColor tcell.Color

func init() {
	if config.HasTheme() {
		os.Setenv("COLORTERM", "24bit")
		primaryTextColorString = config.GetThemePrimaryTextColor()
		secondaryTextColorString = config.GetThemeSecondaryTextColor()
		secondaryTextColor = tcell.GetColor(secondaryTextColorString)
		primaryColor = tcell.GetColor(config.GetThemePrimaryColor())
		backgroundColor = tcell.GetColor(config.GetThemeBackgroundColor())
	} else {
		// do not override the terminal's theme, use all of its defaults
		os.Setenv("TCELL_TRUECOLOR", "disable")
		primaryTextColorString = "white"
		secondaryTextColorString = "blue"
		primaryColor = tcell.ColorBlue
		primaryTextColor = tcell.GetColor(primaryTextColorString)
		secondaryTextColor = tcell.GetColor(secondaryTextColorString)
	}

	primaryTextColor = tcell.GetColor(primaryTextColorString)
}

// ViewContext holds references to all the top-level UI elements in the application
type ViewContext struct {
	App          *tview.Application
	ChannelList  *tview.List
	FavoriteList *tview.List
	Keybindings  *KeybindingView
	NowPlaying   *NowPlayingView
	Status       *StatusView
}

// KeybindingView is a custom view for dispalying the keyboard bindings available to users
type KeybindingView struct {
	*tview.Box
	Bindings []UIKeybinding
}

// UIKeybinding is a helper struct for building a ControlsView
type UIKeybinding struct {
	Description string // a description of what the keybinding does
	Func        func() // the funcion to execute when the binding is pressed
	Shortcut    string // a keybinding to bind to this control
}

// NowPlayingView is a custom view for dispalying the currently playing channel
type NowPlayingView struct {
	*tview.Box
	Art     string
	Channel *components.ChannelItem
	Elapsed float64
	Track   components.Track
}

// StatusView shows temporary status messages in the application
type StatusView struct {
	*tview.Box
	Message string
}

// CreateViewContext creates the primary application view of di-tui
func CreateViewContext() *ViewContext {
	return &ViewContext{
		App:          tview.NewApplication(),
		ChannelList:  createChannelList(),
		FavoriteList: createFavoriteList(),
		Keybindings:  createKeybindings(),
		NowPlaying:   createNowPlaying(),
		Status:       createStatusView(),
	}
}

// Draw draws a NowPlayingView onto the scren
func (n *NowPlayingView) Draw(screen tcell.Screen) {
	n.Box.Draw(screen)
	n.Box.SetBackgroundColor(backgroundColor)

	x, y, width, _ := n.GetInnerRect()

	line := fmt.Sprintf("%s[%s] %s", "Channel:", secondaryTextColorString, n.Channel.Name)
	tview.Print(screen, line, x, y, width, tview.AlignLeft, primaryTextColor)

	line = fmt.Sprintf("%s[%s]  %s", "Artist:", secondaryTextColorString, n.Track.Artist)
	tview.Print(screen, line, x, y+1, width, tview.AlignLeft, primaryTextColor)

	line = fmt.Sprintf("%s[%s]   %s", "Track:", secondaryTextColorString, n.Track.Title)
	tview.Print(screen, line, x, y+2, width, tview.AlignLeft, primaryTextColor)

	line = fmt.Sprintf("%s[%s] %s", "Elapsed:", secondaryTextColorString, n.elapsedString())
	tview.Print(screen, line, x, y+3, width, tview.AlignLeft, primaryTextColor)

	if n.Art == "" {
		return
	}

	artLines := strings.Split(n.Art, "\n")
	for i, line := range artLines {
		if i == 0 {
			label := fmt.Sprintf("[%s] %s", primaryTextColorString, "Album Cover")
			tview.Print(screen, label, x, y+4, width, tview.AlignCenter, secondaryTextColor)
		}

		l := fmt.Sprintf("[%s] %s", primaryTextColorString, line)
		tview.Print(screen, l, x, y+5+i, width, tview.AlignCenter, secondaryTextColor)
	}
}

func (n *NowPlayingView) elapsedString() (str string) {

	if n.Elapsed >= 0 {
		minutes := int(n.Elapsed / 60)
		seconds := int(n.Elapsed) % 60
		str = fmt.Sprintf("%02d:%02d", minutes, seconds)
	}

	return
}

// Draw draws a NowPlayingView onto the scren
func (s *StatusView) Draw(screen tcell.Screen) {

	s.Box.Draw(screen)
	s.Box.SetBackgroundColor(backgroundColor)

	x, y, width, _ := s.GetInnerRect()
	line := fmt.Sprintf("%s[%s] %s", "Message:", secondaryTextColorString, s.Message)
	tview.Print(screen, line, x, y, width, tview.AlignLeft, primaryTextColor)
}

// Draw draws the key bindings view on to the screen
func (k *KeybindingView) Draw(screen tcell.Screen) {
	k.Box.Draw(screen)
	k.Box.SetBackgroundColor(backgroundColor)

	x, y, width, _ := k.GetInnerRect()

	previousWidth := 0
	for j, bnd := range k.Bindings {
		line := fmt.Sprintf("(%s)[%s] %s", bnd.Shortcut, primaryTextColorString, bnd.Description)
		tview.Print(screen, line, x+previousWidth, y, width, tview.AlignLeft, secondaryTextColor)
		previousWidth += len(bnd.Shortcut) + len(bnd.Description) + 4
		// virtically separate playback controls from ui controls
		// yes, this is hacky, but there's a comment, so it's ok, right?
		if j == 4 {
			y = y + 1
			previousWidth = 0
		}
	}
}

func createChannelList() *tview.List {
	list := tview.NewList()
	list.
		ShowSecondaryText(false).
		SetBorder(true).
		SetTitle(" Channels ").
		SetTitleColor(primaryTextColor).
		SetBorderColor(primaryColor)

	list.SetMainTextColor(primaryTextColor)
	list.Box.SetBackgroundColor(backgroundColor)

	return list
}

func createFavoriteList() *tview.List {
	list := tview.NewList()
	list.
		ShowSecondaryText(false).
		SetBorder(true).
		SetBorderColor(primaryColor).
		SetTitle(" Favorites ").
		SetTitleColor(primaryTextColor)

	list.Box.SetBackgroundColor(backgroundColor)
	list.SetMainTextColor(primaryTextColor)

	return list
}

func createKeybindings() *KeybindingView {
	kbv := &KeybindingView{Box: tview.NewBox()}
	kbv.
		SetBorder(true).
		SetBorderColor(primaryColor).
		SetTitleColor(primaryTextColor).
		SetTitle(" Keyboard Controls ")

	kbv.SetBorderColor(primaryColor)

	return kbv
}

func createNowPlaying() *NowPlayingView {
	np := &NowPlayingView{
		Box:     tview.NewBox(),
		Channel: &components.ChannelItem{},
		Elapsed: 0.0,
	}

	np.SetTitle(" Now Playing ").
		SetBorder(true).
		SetBorderColor(primaryColor).
		SetTitleColor(primaryTextColor)

	return np
}

func createStatusView() *StatusView {
	sv := &StatusView{
		Box:     tview.NewBox(),
		Message: "Ready to Play",
	}
	sv.SetTitle(" Status ")
	sv.SetTitleColor(primaryTextColor)
	sv.SetBorder(true)
	sv.SetBorderColor(primaryColor)

	return sv
}

func GetKeybindings() (bindings []UIKeybinding) {
	bindings = []UIKeybinding{
		UIKeybinding{Shortcut: "c", Description: "Channels", Func: func() {}},
		UIKeybinding{Shortcut: "f", Description: "Favorites", Func: func() {}},
		UIKeybinding{Shortcut: "F", Description: "Toggle Favorite", Func: func() {}},
		UIKeybinding{Shortcut: "j", Description: "Scroll Down", Func: func() {}},
		UIKeybinding{Shortcut: "k", Description: "Scroll Up", Func: func() {}},
		UIKeybinding{Shortcut: "q", Description: "Quit", Func: func() {}},
		UIKeybinding{Shortcut: "p", Description: "Pause", Func: func() {}},
		UIKeybinding{Shortcut: "Enter", Description: "Play", Func: func() {}},
	}

	return bindings
}
