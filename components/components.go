package components

import "time"

// ChannelItem contains di.fm channel metadata
type ChannelItem struct {
	ID       int64  `json:"id"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	Playlist string `json:"playlist"`
}

// FavoriteItem contains a di.fm favorite channel
type FavoriteItem struct {
	Name        string
	PlaylistURL string
}

// CurrentlyPlaying contains the currently playing metadata for a di.fm channel
type CurrentlyPlaying struct {
	ChannelID  int64  `json:"channel_id"`
	ChannelKey string `json:"channel_key"`
	Track      Track  `json:"track"`
}

// Track is metadata about a currently playing di.fm track
type Track struct {
	Artist    string    `json:"display_artist"`
	Title     string    `json:"display_title"`
	Duration  float64   `json:"duration"`
	StartTime time.Time `json:"start_time"`
}

// StatusMessage is a message to display in the application for a fixed period of time
type StatusMessage struct {
	Message  string
	Duration time.Duration
}
