package context

import (
	"io"
	"time"

	"github.com/acaloiaro/di-tui/components"
	"github.com/acaloiaro/di-tui/views"
	"github.com/jfreymuth/pulse"
)

// CreateAppContext creates the application context
func CreateAppContext(view *views.ViewContext) *AppContext {
	ctx := &AppContext{
		View:          view,
		StatusChannel: make(chan components.StatusMessage, 10),
	}

	return ctx
}

// AppContext is a shared context to be shared across the application
// AudioStream - The raw audio stream from which audio is streamed by the player
// CurrentChannel - The ChannelItem representing the currently playing on the network
// DfmToken - The token used to authenticate to the network
// IsPlaying - Is there audio playing?
// Netowrk - the network to connect to
// View - The view context
// Status - Gets and sets current application status messages
type AppContext struct {
	AudioStream        io.ReadCloser
	ChannelList        []components.ChannelItem
	CurrentChannel     *components.ChannelItem
	DifmToken          string
	FavoriteList       []components.FavoriteItem
	Network            *components.Network
	HighlightedChannel *components.ChannelItem
	IsPlaying          bool
	Player             *pulse.PlaybackStream
	ShowStatus         bool // The status pane will be visible when true
	StatusChannel      chan components.StatusMessage
	View               *views.ViewContext
}

// SetStatusMessage sets the application's status message for five seconds.
func (c *AppContext) SetStatusMessage(msg string) {
	c.SetStatusMessageTimed(msg, 5*time.Second)
}

// SetStatusMessageTimed sets the application's status message for a fixed period of time.
func (c *AppContext) SetStatusMessageTimed(msg string, d time.Duration) {
	c.StatusChannel <- components.StatusMessage{Message: msg, Duration: d}
}
