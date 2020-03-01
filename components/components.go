package components

import "time"

// ChannelItem contains di.fm channel metadata
type ChannelItem struct {
	ID       int64  `json:"id"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	Playlist string `json:"playlist"`
}

// CurrentlyPlaying contains the currently playing metadata for a di.fm channel
type CurrentlyPlaying struct {
	ChannelID  int64  `json:"channel_id"`
	ChannelKey string `json:"channel_key"`
	Track      Track  `json:"track"`
}

type Track struct {
	DisplayArtist string    `json:"display_artist"`
	DisplayTitle  string    `json:"display_title"`
	Duration      float64   `json:"duration"`
	StartTime     time.Time `json:"start_time"`
}
