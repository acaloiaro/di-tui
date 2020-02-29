package context

import (
	"github.com/acaloiaro/dicli/components"
	"github.com/acaloiaro/dicli/views"
	"github.com/faiface/beep"
)

func CreateAppContext() *AppContext {
	return &AppContext{
		View: views.CreateAppView(),
	}
}

type AppContext struct {
	AudioStream        beep.StreamSeekCloser
	CurrentChannel     *components.ChannelItem
	DifmToken          string
	IsPlaying          bool
	SpeakerInitialized bool
	View               *views.AppView
}
