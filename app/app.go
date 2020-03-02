package app

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/acaloiaro/di-tui/components"
	"github.com/acaloiaro/di-tui/context"
	"github.com/acaloiaro/di-tui/difm"
	"github.com/faiface/beep/speaker"
)

// PlayChannel begins streaming the provided channel after fetching its playlist
// If a channel is already playing, the old stream is stopped first, clearing up resources.
// This function is *asynchronous* and creates a single streaming resource: the audio stream held by the application
// context. To clean up resources created by this function, Close() the application's audio stream.
func PlayChannel(chn *components.ChannelItem, ctx *context.AppContext) {

	// when other channels are already playing, close their stream before playing a new one
	if ctx.AudioStream != nil {
		ctx.AudioStream.Close()
	}

	go func() {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", chn.Playlist, nil)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if streamURL, ok := difm.GetStreamURL(body); ok {
			format := difm.Stream(streamURL, ctx)

			if !ctx.SpeakerInitialized {
				speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
				ctx.SpeakerInitialized = true
			}

			speaker.Play(ctx.AudioStream)
			ctx.IsPlaying = true

			UpdateNowPlaying(chn, ctx)
		}
	}()
}

// TogglePause pauses/unpauses audio when a channel is playing
func TogglePause(ctx *context.AppContext) {

	// nothing to do if nothing has been streamed
	if ctx.AudioStream == nil {
		return
	}

	ctx.AudioStream.Close()
	if !ctx.IsPlaying {
		PlayChannel(ctx.CurrentChannel, ctx)
	}

	ctx.IsPlaying = !ctx.IsPlaying
}

func UpdateNowPlaying(chn *components.ChannelItem, ctx *context.AppContext) {
	ctx.CurrentChannel = chn
	cp := difm.GetCurrentlyPlaying(ctx)

	ctx.View.App.QueueUpdateDraw(func() {
		ctx.View.NowPlaying.Channel = chn
		ctx.View.NowPlaying.Track = cp.Track
	})
}
