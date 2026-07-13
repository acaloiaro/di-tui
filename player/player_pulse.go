package player

import (
	"io"

	"github.com/acaloiaro/di-tui/context"
	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/pulse"
	"github.com/jfreymuth/pulse/proto"
)

func playPulse(ctx *context.AppContext, stream io.ReadWriter, playbackLatency int) error {
	d, err := mp3.NewDecoder(stream)
	if err != nil {
		return err
	}

	c, err := pulse.NewClient()
	if err != nil {
		return err
	}
	defer c.Close()

	p, err := c.NewPlayback(
		// proto.FormatInt16LE convinces pulse to expect the format emitted by go-mp3.
		pulse.NewReader(d, proto.FormatInt16LE),
		pulse.PlaybackSampleRate(d.SampleRate()),
		pulse.PlaybackStereo,
		pulse.PlaybackLatency(float64(playbackLatency)),
	)
	if err != nil {
		return err
	}

	ctx.Player = p
	p.Start()
	p.Drain()
	return p.Error()
}
