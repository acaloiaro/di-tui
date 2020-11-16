package player

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/acaloiaro/di-tui/context"
	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/pulse"
	"github.com/jfreymuth/pulse/proto"
)

func Play(ctx *context.AppContext, stream io.Reader) {
	d, err := mp3.NewDecoder(stream)
	if err != nil {
		log.Println("Unable to decode mp3 file")
	}

	c, err := pulse.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	ctx.AudioStream, err = c.NewPlayback(
		// proto.FormatInt16LE convinces `pulse` to expect 2 bytes per sample; the format that go-mp3 lays out bytes
		pulse.NewReader(d, proto.FormatInt16LE),
		pulse.PlaybackSampleRate(d.SampleRate()),
		pulse.PlaybackStereo,
		pulse.PlaybackBufferSize(16482),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		ctx.AudioStream.Start()
		ctx.AudioStream.Drain()
		if ctx.AudioStream.Underflow() {
			log.Println("Underflow!")
			os.Exit(1)
		}
	}()
}
