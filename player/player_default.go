//go:build !darwin

// This is the default audio dispatcher. It makes PulseAudio the only available
// provider and supplies the platform-specific functions shared code calls.
// macOS has a specialized implementation that overrides this one.
package player

import (
	"fmt"
	"io"

	"github.com/acaloiaro/di-tui/context"
)

const defaultAudioProvider = PulseProvider

func ConfigureAudioProvider(provider string) error {
	if provider != PulseProvider {
		return fmt.Errorf("unsupported audio provider %q (choose %s)", provider, PulseProvider)
	}
	audioProvider = provider
	return nil
}

func Play(ctx *context.AppContext, stream io.ReadWriter, playbackLatency int) error {
	return playPulse(ctx, stream, playbackLatency)
}
