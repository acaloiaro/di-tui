package player

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/acaloiaro/di-tui/context"
	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/pulse"
	"github.com/jfreymuth/pulse/proto"
)

var IsPlaying bool
var pulseClient *pulse.Client

func init() {
	var err error
	pulseClient, err = pulse.NewClient()
	if err != nil {
		log.Fatalf("error initializing pulse: %v", err)
	}
}

func Play(ctx *context.AppContext, stream io.Reader) (err error) {
	if ctx.AudioStream != nil {
		ctx.AudioStream.Close()
	}

	IsPlaying = true
	var d *mp3.Decoder
	err = errors.New("waiting for buffer to fill")
	// Wait for the buffer to contain audio data
	for err != nil {
		d, err = mp3.NewDecoder(stream)
		if err != nil {
			// TODO print a message explaining that the fetched content can't be played
			continue
		}
	}

	ctx.AudioStream, err = pulseClient.NewPlayback(
		// proto.FormatInt16LE convinces `pulse` to expect 2 bytes per sample; the format that go-mp3 lays out bytes
		pulse.NewReader(d, proto.FormatInt16LE),
		pulse.PlaybackSampleRate(d.SampleRate()),
		pulse.PlaybackStereo,
		pulse.PlaybackBufferSize(16482),
	)
	if err != nil {
		log.Println(err)
		os.Exit(1)
		return
	}

	ctx.AudioStream.Start()
	ctx.AudioStream.Drain()

	IsPlaying = false

	return
}
