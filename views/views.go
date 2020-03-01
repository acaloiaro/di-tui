package views

import (
	"fmt"

	"github.com/acaloiaro/dicli/components"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type AppView struct {
	App          *tview.Application
	ChannelList  *tview.List
	FavoriteList *tview.List
	NowPlaying   *NowPlayingView
	Keybindings  *KeybindingView
}

// KeybindingView is a custom view for dispalying the keyboard bindings available to users
type KeybindingView struct {
	*tview.Box
	Bindings []UIKeybinding
}

// UIKeybinding is a helper struct for building a ControlsView
type UIKeybinding struct {
	Shortcut    string // a keybinding to bind to this control
	Description string // a description of what the keybinding does
	Func        func() // the funcion to execute when the binding is pressed
}

// CreateAppView creates the primary application view of di-tui
func CreateAppView() *AppView {
	return &AppView{
		App:          tview.NewApplication(),
		ChannelList:  createChannelList(),
		FavoriteList: createFavoriteList(),
		NowPlaying:   newNowPlaying(&components.ChannelItem{}),
		Keybindings:  createKeybindings(),
	}
}

// NowPlayingView is a custom view for dispalying the currently playing channel
type NowPlayingView struct {
	*tview.Box
	Channel *components.ChannelItem
	Artist  string
	Track   string
}

func newNowPlaying(chn *components.ChannelItem) *NowPlayingView {
	np := &NowPlayingView{
		Box:     tview.NewBox(),
		Channel: chn,
	}

	np.SetTitle(" Now Playing ")
	np.SetBorder(true)

	return np
}

// Draw draws this primitive onto the screen.
func (n *NowPlayingView) Draw(screen tcell.Screen) {
	n.Box.Draw(screen)
	x, y, width, _ := n.GetInnerRect()

	line := fmt.Sprintf("%s[white] %s", "Channel:", n.Channel.Name)
	tview.Print(screen, line, x, y, width, tview.AlignLeft, tcell.ColorBlue)

	line = fmt.Sprintf("%s[white]  %s", "Artist:", n.Artist)
	tview.Print(screen, line, x, y+1, width, tview.AlignLeft, tcell.ColorBlue)

	line = fmt.Sprintf("%s[white]   %s", "Track:", n.Track)
	tview.Print(screen, line, x, y+2, width, tview.AlignLeft, tcell.ColorBlue)

}

// Draw draws this primitive onto the screen.
func (n *KeybindingView) Draw(screen tcell.Screen) {
	n.Box.Draw(screen)
	x, y, width, _ := n.GetInnerRect()

	previousWidth := 0
	for _, bnd := range n.Bindings {
		line := fmt.Sprintf("(%s)[white] %s", bnd.Shortcut, bnd.Description)
		tview.Print(screen, line, x+previousWidth, y, width, tview.AlignLeft, tcell.ColorBlue)
		previousWidth += len(bnd.Shortcut) + len(bnd.Description) + 5
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
