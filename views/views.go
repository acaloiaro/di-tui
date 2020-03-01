package views

import (
	"fmt"

	"github.com/acaloiaro/dicli/components"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type AppView struct {
	App         *tview.Application
	ChannelList *tview.List
	NowPlaying  *NowPlayingView
}

func CreateAppView() *AppView {
	return &AppView{
		App:         tview.NewApplication(),
		ChannelList: createChannelList(),
		NowPlaying:  newNowPlaying(&components.ChannelItem{Name: "N/A"}),
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

	line := fmt.Sprintf("%s[white]  %s", "Channel:", n.Channel.Name)
	tview.Print(screen, line, x, y, width, tview.AlignLeft, tcell.ColorYellow)

	line = fmt.Sprintf("%s[white]   %s", "Artist:", n.Artist)
	tview.Print(screen, line, x, y+1, width, tview.AlignLeft, tcell.ColorYellow)

	line = fmt.Sprintf("%s[white]    %s", "Track:", n.Track)
	tview.Print(screen, line, x, y+2, width, tview.AlignLeft, tcell.ColorYellow)

}

func createChannelList() *tview.List {
	list := tview.NewList()
	list.
		ShowSecondaryText(false).
		SetBorder(true).
		SetTitle(" Channels ")

	return list
}
