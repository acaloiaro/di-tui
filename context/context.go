package context

import (
	"github.com/acaloiaro/dicli/components"
	"github.com/acaloiaro/dicli/views"
	"github.com/faiface/beep"
)

func CreateAppContext() *AppContext {
	ctx := &AppContext{
		View: views.CreateAppView(),
	}

	return ctx
}

type AppContext struct {
	AudioStream        beep.StreamSeekCloser
	CurrentChannel     *components.ChannelItem
	DifmToken          string
	IsPlaying          bool
	SpeakerInitialized bool
	View               *views.AppView
}
