package context

import (
	"time"

	"github.com/acaloiaro/di-tui/components"
	"github.com/acaloiaro/di-tui/views"
	"github.com/faiface/beep"
)

// CreateAppContext creates the application context
func CreateAppContext(view *views.ViewContext) *AppContext {
	ctx := &AppContext{
		View:          view,
		StatusChannel: make(chan components.StatusMessage, 1),
	}

	return ctx
}

// AppContext is a shared context to be shared across the application
// AudioStream - The raw audio stream used by the beep library to stream audio
// CurrentChannel - The ChannelItem representing the currently playing di.fm channel
// DfmToken - The token used to authenticate to di.fm
// IsPlaying - Is there audio playing?
// SpeakerInitialized - Has the speaker been initialized with a bitrate?
// View - The view context
// Status - Gets and sets current application status messages
type AppContext struct {
	AudioStream        beep.StreamSeekCloser
	CurrentChannel     *components.ChannelItem
	DifmToken          string
	IsPlaying          bool
	ShowStatus         bool // The status pane will be visible when true
	SpeakerInitialized bool
	StatusChannel      chan components.StatusMessage
	View               *views.ViewContext
}

// SetStatusMessage sets the application's status message for five seconds.
func (c *AppContext) SetStatusMessage(msg string) {
	c.SetStatusMessageTimed(msg, 5*time.Second)
}

// SetStatusMessageTimed sets the application's status message for a fixed period of time.
func (c *AppContext) SetStatusMessageTimed(msg string, d time.Duration) {
	c.ShowStatus = true
	c.StatusChannel <- components.StatusMessage{Message: msg, Duration: d}
}
