package mpris

import (
	"sync"
	"time"

	"adriano.fyi/di-tui/app"
	"adriano.fyi/di-tui/context"
	"github.com/godbus/dbus/v5"
	"github.com/quarckster/go-mpris-server/pkg/server"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

var s *server.Server

type Root struct{}

func (r Root) Raise() error {
	return nil
}

func (r Root) Quit() error {
	return nil
}

func (r Root) CanQuit() (bool, error) {
	return false, nil
}

func (r Root) CanRaise() (bool, error) {
	return false, nil
}

func (r Root) HasTrackList() (bool, error) {
	return true, nil
}

func (r Root) Identity() (string, error) {
	return "di-tui - di.fm player", nil
}

func (r Root) SupportedUriSchemes() ([]string, error) {
	return []string{}, nil
}

func (r Root) SupportedMimeTypes() ([]string, error) {
	return []string{}, nil
}

var _ types.OrgMprisMediaPlayer2Adapter = Root{}

type Player struct {
	ctx *context.AppContext

	mu       sync.Mutex
	metaData types.Metadata
	status   types.PlaybackStatus
}

// emitStatus records the playback status and announces it if it changed.
func (p *Player) emitStatus(status types.PlaybackStatus) {
	p.mu.Lock()
	changed := p.status != status
	p.status = status
	p.mu.Unlock()

	if !changed || s.Conn == nil {
		return
	}
	props := map[string]dbus.Variant{
		"PlaybackStatus": dbus.MakeVariant(status),
	}
	s.Conn.Emit("/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Properties.PropertiesChanged", "org.mpris.MediaPlayer2.Player", props, []string{})
}

// emitMetadata records the current track and announces it if it changed.
//
// Only the populated keys are sent: some MPRIS consumers (gnome-shell's media
// widget, see https://gitlab.gnome.org/GNOME/gnome-shell/-/issues/9394) do
// work proportional to the number of keys on every change.
func (p *Player) emitMetadata(artist, title string) {
	p.mu.Lock()
	changed := p.metaData.Title != title || len(p.metaData.Artist) != 1 || p.metaData.Artist[0] != artist
	p.metaData.Artist = []string{artist}
	p.metaData.Title = title
	trackID := p.metaData.TrackId
	p.mu.Unlock()

	if !changed || s.Conn == nil {
		return
	}
	metadata := map[string]dbus.Variant{
		"mpris:trackid": dbus.MakeVariant(dbus.ObjectPath(trackID)),
	}
	if title != "" {
		metadata["xesam:title"] = dbus.MakeVariant(title)
	}
	if artist != "" {
		metadata["xesam:artist"] = dbus.MakeVariant([]string{artist})
	}
	props := map[string]dbus.Variant{
		"Metadata": dbus.MakeVariant(metadata),
	}
	s.Conn.Emit("/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Properties.PropertiesChanged", "org.mpris.MediaPlayer2.Player", props, []string{})
}

func (p *Player) Next() error {
	return nil
}

func (p *Player) Previous() error {
	return nil
}

func (p *Player) Pause() error {
	return nil
}

func (p *Player) PlayPause() error {
	app.TogglePause(p.ctx)
	newStatus := types.PlaybackStatusPaused
	if p.ctx.IsPlaying {
		newStatus = types.PlaybackStatusPlaying
	}
	p.emitStatus(newStatus)
	return nil
}

func (p *Player) Stop() error {
	app.Stop(p.ctx)
	p.emitStatus(types.PlaybackStatusStopped)
	return nil
}

func (p *Player) Play() error {
	app.Play(p.ctx)
	p.emitStatus(types.PlaybackStatusPlaying)
	return nil
}

func (p *Player) Seek(offset types.Microseconds) error {
	return nil
}

func (p *Player) SetPosition(trackId string, position types.Microseconds) error {
	return nil
}

func (p *Player) OpenUri(uri string) error {
	return nil
}

func (p *Player) PlaybackStatus() (types.PlaybackStatus, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.status, nil
}

func (p *Player) Rate() (float64, error) {
	return 1.2, nil
}

func (p *Player) SetRate(float64) error {
	return nil
}

func (p *Player) Metadata() (types.Metadata, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.metaData, nil
}
func (p *Player) Volume() (float64, error) {
	return 0, nil
}
func (p *Player) SetVolume(in float64) error {
	return nil
}
func (p *Player) Position() (int64, error) {
	return 0, nil
}
func (p *Player) MinimumRate() (float64, error) {
	return 0, nil
}
func (p *Player) MaximumRate() (float64, error) {
	return 0, nil
}
func (p *Player) CanGoNext() (bool, error) {
	return false, nil
}
func (p *Player) CanGoPrevious() (bool, error) {
	return false, nil
}
func (p *Player) CanPlay() (bool, error) {
	return true, nil
}
func (p *Player) CanPause() (bool, error) {
	return true, nil
}
func (p *Player) CanSeek() (bool, error) {
	return false, nil
}
func (p *Player) CanControl() (bool, error) {
	return true, nil
}

// Start starts the mpris server, handling play/pause events, and announces the currently playing track on an interval
func Start(ctx *context.AppContext) {
	metaData := types.Metadata{
		TrackId:        "/TrackList/Track1",
		ArtUrl:         "",
		Album:          "",
		AlbumArtist:    []string{},
		Artist:         []string{""},
		AsText:         "",
		AudioBPM:       0,
		AutoRating:     0.0,
		Comment:        []string{},
		Composer:       []string{},
		ContentCreated: "",
		DiscNumber:     0,
		FirstUsed:      "",
		Genre:          []string{},
		LastUsed:       "",
		Lyricist:       []string{},
		Title:          "",
		TrackNumber:    0,
		Url:            "",
		UseCount:       0,
		UserRating:     0.0,
	}
	r := Root{}
	p := &Player{ctx: ctx, metaData: metaData, status: types.PlaybackStatusStopped}
	s = server.NewServer("di-tui", r, p)
	go s.Listen()

	// Playback can be started, paused and stopped from the TUI without going
	// through the MPRIS methods above, and the track changes on its own, so
	// poll the application state and announce it -- but only when something
	// actually changed. Unconditional PropertiesChanged signals every second
	// make some desktops (gnome-shell) stall on every one of them.
	go func() {
		for range time.Tick(1 * time.Second) {
			if s.Conn == nil {
				continue
			}

			if p.ctx.IsPlaying {
				p.emitStatus(types.PlaybackStatusPlaying)
			} else if status, _ := p.PlaybackStatus(); status == types.PlaybackStatusPlaying {
				p.emitStatus(types.PlaybackStatusPaused)
			}

			track := p.ctx.View.NowPlaying.Track
			p.emitMetadata(track.Artist, track.Title)
		}
	}()
}
