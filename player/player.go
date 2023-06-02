package player

import (
	"io"

	"github.com/acaloiaro/di-tui/context"
	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/pulse"
	"github.com/jfreymuth/pulse/proto"
)

func Play(ctx *context.AppContext, stream io.ReadWriter, playbackLatency int) (err error) {
	var d *mp3.Decoder
	d, err = mp3.NewDecoder(stream)
	if err != nil {
		return
	}

	var c *pulse.Client
	c, err = pulse.NewClient()
	if err != nil {
		return
	}
	defer c.Close()

	ctx.Player, err = c.NewPlayback(
		// proto.FormatInt16LE convinces `pulse` to expect 2 bytes per sample; the format that go-mp3 lays out bytes
		pulse.NewReader(d, proto.FormatInt16LE),
		pulse.PlaybackSampleRate(d.SampleRate()),
		pulse.PlaybackStereo,
		pulse.PlaybackLatency(float64(playbackLatency)),
	)
	if err != nil {
		return
	}

	ctx.Player.Start()
	ctx.Player.Drain()

	err = ctx.Player.Error()

	return
}

func Stop(ctx *context.AppContext) {
	ctx.IsPlaying = false
	if ctx.Player != nil {
		ctx.Player.Close()
	}
}
