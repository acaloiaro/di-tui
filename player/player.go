package player

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/acaloiaro/di-tui/context"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
)

var IsPlaying = false

var otoCtx *oto.Context
var readyChan <-chan struct{}

func init() {
	var err error
	samplingRate := 44100

	// Number of channels (aka locations) to play sounds from. Either 1 or 2.
	// 1 is mono sound, and 2 is stereo (most speakers are stereo).
	numOfChannels := 2

	// Bytes used by a channel to represent one sample. Either 1 or 2 (usually 2).
	audioBitDepth := 2

	// Remember that you should **not** create more than one context
	otoCtx, readyChan, err = oto.NewContext(samplingRate, numOfChannels, audioBitDepth)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	// It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
	<-readyChan

}

func Play(ctx *context.AppContext, stream io.Reader) (err error) {
	IsPlaying = true
	var d *mp3.Decoder
	err = errors.New("wait for buffer to fill")
	// Wait for the buffer to contain audio data
	for err != nil {
		d, err = mp3.NewDecoder(stream)
		if err != nil {
			// TODO print a message explaining that the fetched content can't be played
			continue
		}
	}

	// Create a new 'player' that will handle our sound. Paused by default.
	player := otoCtx.NewPlayer(d)

	// Play starts playing the sound and returns without waiting for it (Play() is async).
	player.Play()

	if player.Err() != nil {
		log.Println(player.Err())
		os.Exit(33)
	}
	// We can wait for the sound to finish playing using something like this
	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}

	fmt.Println("No longer playing")
	IsPlaying = false
	return
}
