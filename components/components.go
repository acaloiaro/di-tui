package components

// ChannelItem contains di.fm channel metadata
type ChannelItem struct {
	ID       int64  `json:"id"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	Playlist string `json:"playlist"`
}
