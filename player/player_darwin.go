//go:build darwin

// This is the macOS audio dispatcher. It defaults to native Core Audio while
// allowing the user to select the shared PulseAudio backend at runtime.
package player

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/acaloiaro/di-tui/context"
	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

var (
	audioContext     *oto.Context
	audioContextErr  error
	audioContextOnce sync.Once
	audioSampleRate  int
)

const defaultAudioProvider = CoreAudioProvider

type darwinPlayer struct {
	*oto.Player
}

func (p *darwinPlayer) Close() {
	_ = p.Player.Close()
}

func ConfigureAudioProvider(provider string) error {
	switch provider {
	case CoreAudioProvider, PulseProvider:
		audioProvider = provider
		return nil
	default:
		return fmt.Errorf("unsupported audio provider %q (choose %s or %s)", provider, CoreAudioProvider, PulseProvider)
	}
}

func Play(ctx *context.AppContext, stream io.ReadWriter, playbackLatency int) error {
	if audioProvider == PulseProvider {
		return playPulse(ctx, stream, playbackLatency)
	}
	return playCoreAudio(ctx, stream)
}

func playCoreAudio(ctx *context.AppContext, stream io.ReadWriter) error {
	d, err := mp3.NewDecoder(stream)
	if err != nil {
		return err
	}

	audioContextOnce.Do(func() {
		audioSampleRate = d.SampleRate()
		var ready chan struct{}
		audioContext, ready, audioContextErr = oto.NewContext(&oto.NewContextOptions{
			SampleRate:   audioSampleRate,
			ChannelCount: 2,
			Format:       oto.FormatSignedInt16LE,
		})
		if audioContextErr == nil {
			<-ready
		}
	})
	if audioContextErr != nil {
		return audioContextErr
	}
	if d.SampleRate() != audioSampleRate {
		return fmt.Errorf("audio sample rate changed from %d Hz to %d Hz", audioSampleRate, d.SampleRate())
	}

	p := &darwinPlayer{Player: audioContext.NewPlayer(d)}
	ctx.Player = p
	p.Play()
	for p.IsPlaying() {
		time.Sleep(10 * time.Millisecond)
	}
	return p.Err()
}
