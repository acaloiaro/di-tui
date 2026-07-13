package player

import (
	"github.com/acaloiaro/di-tui/context"
)

const (
	CoreAudioProvider = "coreaudio"
	PulseProvider     = "pulse"
)

var audioProvider = defaultAudioProvider

func DefaultAudioProvider() string {
	return defaultAudioProvider
}

func Stop(ctx *context.AppContext) {
	ctx.IsPlaying = false
	if ctx.StreamCancel != nil {
		ctx.StreamCancel()
	}
	if ctx.Player != nil {
		ctx.AudioStream.Close()
		ctx.Player.Close()
	}
}
