package player

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/acaloiaro/di-tui/context"
	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/pulse"
	"github.com/jfreymuth/pulse/proto"
)

var IsPlaying = false

var pulseClient *pulse.Client

func init() {
	var err error
	pulseClient, err = pulse.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func Play(ctx *context.AppContext, stream io.Reader) (err error) {
	var d *mp3.Decoder
	err = errors.New("wait for buffer to fill")
	// Wait for the buffer to contain audio data
	for err != nil {
		d, err = mp3.NewDecoder(stream)
		if err != nil {
			// TODO print a message explaining that the fetched content can't be played
			return
		}
		time.Sleep(time.Second * 1)
	}

	ctx.AudioStream, err = pulseClient.NewPlayback(
		// proto.FormatInt16LE convinces `pulse` to expect 2 bytes per sample; the format that go-mp3 lays out bytes
		pulse.NewReader(d, proto.FormatInt16LE),
		pulse.PlaybackSampleRate(d.SampleRate()),
		pulse.PlaybackStereo,
		pulse.PlaybackBufferSize(16482),
		pulse.PlaybackLatency(3.0),
	)
	if err != nil {
		return
	}

	IsPlaying = true
	ctx.AudioStream.Start()
	//ctx.AudioStream.Drain()
	IsPlaying = false

	return
}
