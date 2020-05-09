package views

import (
	"fmt"
	"strings"

	"github.com/acaloiaro/di-tui/components"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

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
	x, y, width, _ := n.GetInnerRect()

	line := fmt.Sprintf("%s[white] %s", "Channel:", n.Channel.Name)
	tview.Print(screen, line, x, y, width, tview.AlignLeft, tcell.ColorBlue)

	line = fmt.Sprintf("%s[white]  %s", "Artist:", n.Track.Artist)
	tview.Print(screen, line, x, y+1, width, tview.AlignLeft, tcell.ColorBlue)

	line = fmt.Sprintf("%s[white]   %s", "Track:", n.Track.Title)
	tview.Print(screen, line, x, y+2, width, tview.AlignLeft, tcell.ColorBlue)

	line = fmt.Sprintf("%s[white] %s", "Elapsed:", n.elapsedString())
	tview.Print(screen, line, x, y+3, width, tview.AlignLeft, tcell.ColorBlue)

	if n.Art == "" {
		return
	}

	artLines := strings.Split(n.Art, "\n")
	for i, line := range artLines {
		if i == 0 {
			label := fmt.Sprintf("[white] %s", "Album Cover")
			tview.Print(screen, label, x, y+4, width, tview.AlignCenter, tcell.ColorBlue)
		}

		l := fmt.Sprintf("[white] %s", line)
		tview.Print(screen, l, x, y+5+i, width, tview.AlignCenter, tcell.ColorBlue)
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
	x, y, width, _ := s.GetInnerRect()
	line := fmt.Sprintf("%s[white] %s", "Message:", s.Message)
	tview.Print(screen, line, x, y, width, tview.AlignLeft, tcell.ColorBlue)
}

// Draw draws the key bindings view on to the screen
func (k *KeybindingView) Draw(screen tcell.Screen) {
	k.Box.Draw(screen)
	x, y, width, _ := k.GetInnerRect()

	previousWidth := 0
	for j, bnd := range k.Bindings {
		line := fmt.Sprintf("(%s)[white] %s", bnd.Shortcut, bnd.Description)
		tview.Print(screen, line, x+previousWidth, y, width, tview.AlignLeft, tcell.ColorBlue)
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
		SetTitle(" Channels ")

	return list
}

func createFavoriteList() *tview.List {
	list := tview.NewList()
	list.
		ShowSecondaryText(false).
		SetBorder(true).
		SetTitle(" Favorites ")

	return list
}

func createKeybindings() *KeybindingView {
	kbv := &KeybindingView{Box: tview.NewBox()}
	kbv.
		SetBorder(true).
		SetTitle(" Key Bindings ")

	return kbv
}

func createNowPlaying() *NowPlayingView {
	np := &NowPlayingView{
		Box:     tview.NewBox(),
		Channel: &components.ChannelItem{},
		Elapsed: 0.0,
	}

	np.SetTitle(" Now Playing ")
	np.SetBorder(true)

	return np
}

func createStatusView() *StatusView {
	sv := &StatusView{
		Box:     tview.NewBox(),
		Message: "Ready to Play",
	}
	sv.SetTitle(" Status ")
	sv.SetBorder(true)

	return sv
}
