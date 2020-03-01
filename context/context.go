package context

import (
	"github.com/acaloiaro/dicli/components"
	"github.com/acaloiaro/dicli/views"
	"github.com/faiface/beep"
)

// CreateAppContext creates the application context
func CreateAppContext(view *views.AppView) *AppContext {
	ctx := &AppContext{View: view}

	return ctx
}

// AppContext is a shared context to be shared across the application
type AppContext struct {
	AudioStream        beep.StreamSeekCloser
	CurrentChannel     *components.ChannelItem
	DifmToken          string
	IsPlaying          bool
	SpeakerInitialized bool
	View               *views.AppView
}
