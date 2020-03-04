package views

import (
	"fmt"

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

// Draw draws the key bindings view on to the screen
func (n *KeybindingView) Draw(screen tcell.Screen) {
	n.Box.Draw(screen)
	x, y, width, _ := n.GetInnerRect()

	previousWidth := 0
	for j, bnd := range n.Bindings {
		line := fmt.Sprintf("(%s)[white] %s", bnd.Shortcut, bnd.Description)
		tview.Print(screen, line, x+previousWidth, y, width, tview.AlignLeft, tcell.ColorBlue)
		previousWidth += len(bnd.Shortcut) + len(bnd.Description) + 4

		// virtically separate playback controls from ui controls
		// yes, this is hacky, but there's a comment, so it's ok, right?
		// TODO clean up this mess
		if j == 4 {
			y = y + 1
			previousWidth = 0
		}
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

	var minutes, seconds int
	if n.Elapsed > 0 {
		minutes = int(n.Elapsed / 60)
		seconds = int(n.Elapsed) % 60
	}
	elapsedStr := fmt.Sprintf("%02d:%02d", minutes, seconds)
	line = fmt.Sprintf("%s[white] %s", "Elapsed:", elapsedStr)
	tview.Print(screen, line, x, y+3, width, tview.AlignLeft, tcell.ColorBlue)
}

// Draw draws a NowPlayingView onto the scren
func (s *StatusView) Draw(screen tcell.Screen) {

	s.Box.Draw(screen)
	x, y, width, _ := s.GetInnerRect()

	line := fmt.Sprintf("%s[white] %s", "Message:", s.Message)
	tview.Print(screen, line, x, y, width, tview.AlignLeft, tcell.ColorBlue)
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
