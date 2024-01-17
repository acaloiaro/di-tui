package mpris

import (
	"time"

	"github.com/acaloiaro/di-tui/app"
	"github.com/acaloiaro/di-tui/context"
	"github.com/godbus/dbus/v5"
	"github.com/quarckster/go-mpris-server/pkg/server"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

var (
	s             *server.Server
	statusPlaying = map[string]dbus.Variant{
		"PlaybackStatus": dbus.MakeVariant(types.PlaybackStatusPlaying),
	}
)

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
	ctx      *context.AppContext
	metaData types.Metadata
}

func (p Player) Next() error {
	return nil
}

func (p Player) Previous() error {
	return nil
}

func (p Player) Pause() error {
	return nil
}

func (p Player) PlayPause() error {
	app.TogglePause(p.ctx)
	newStatus := types.PlaybackStatusPaused
	if p.ctx.IsPlaying {
		newStatus = types.PlaybackStatusPlaying
	}
	status := map[string]dbus.Variant{
		"PlaybackStatus": dbus.MakeVariant(newStatus),
	}
	s.Conn.Emit("/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Properties.PropertiesChanged", "org.mpris.MediaPlayer2.Player", status, []string{})
	return nil
}

func (p Player) Stop() error {
	return nil
}

func (p Player) Play() error {
	return nil
}

func (p Player) Seek(offset types.Microseconds) error {
	return nil
}

func (p Player) SetPosition(trackId string, position types.Microseconds) error {
	return nil
}

func (p Player) OpenUri(uri string) error {
	return nil
}

func (p Player) PlaybackStatus() (types.PlaybackStatus, error) {
	return types.PlaybackStatusStopped, nil
}

func (p Player) Rate() (float64, error) {
	return 1.2, nil
}

func (p Player) SetRate(float64) error {
	return nil
}

func (p Player) Metadata() (types.Metadata, error) {
	return p.metaData, nil
}
func (p Player) Volume() (float64, error) {
	return 0, nil
}
func (p Player) SetVolume(in float64) error {
	return nil
}
func (p Player) Position() (int64, error) {
	return 0, nil
}
func (p Player) MinimumRate() (float64, error) {
	return 0, nil
}
func (p Player) MaximumRate() (float64, error) {
	return 0, nil
}
func (p Player) CanGoNext() (bool, error) {
	return false, nil
}
func (p Player) CanGoPrevious() (bool, error) {
	return false, nil
}
func (p Player) CanPlay() (bool, error) {
	return true, nil
}
func (p Player) CanPause() (bool, error) {
	return true, nil
}
func (p Player) CanSeek() (bool, error) {
	return false, nil
}
func (p Player) CanControl() (bool, error) {
	return true, nil
}

// Start starts the mpris server, handling play/pause events, and announces the currently playing track on an interval
func Start(ctx *context.AppContext) {
	metaData := types.Metadata{
		TrackId:        "/TrackList/Track1",
		Length:         100,
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
	p := Player{ctx: ctx, metaData: metaData}
	s = server.NewServer("di-tui", r, p)
	// _ = events.NewEventHandler(s)
	go s.Listen()
	go func() {
		for {
			<-time.Tick(1 * time.Second)
			if !p.ctx.IsPlaying {
				continue
			}

			p.metaData.Artist = []string{p.ctx.View.NowPlaying.Track.Artist}
			p.metaData.Title = p.ctx.View.NowPlaying.Track.Title
			props := map[string]any{
				"Metadata": p.metaData.MakeMap(),
			}
			s.Conn.Emit("/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Properties.PropertiesChanged", "org.mpris.MediaPlayer2.Player", props, []string{})

			s.Conn.Emit("/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Properties.PropertiesChanged", "org.mpris.MediaPlayer2.Player", statusPlaying, []string{})
		}
	}()
}
