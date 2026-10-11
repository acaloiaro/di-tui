package player

import "testing"

func TestConfigureAudioProvider(t *testing.T) {
	if err := ConfigureAudioProvider(DefaultAudioProvider()); err != nil {
		t.Fatalf("default provider was rejected: %v", err)
	}
	if err := ConfigureAudioProvider("unknown"); err == nil {
		t.Fatal("unknown provider was accepted")
	}
}
