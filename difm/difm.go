package difm

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/acaloiaro/di-tui/context"
	"github.com/acaloiaro/di-tui/player"

	"github.com/acaloiaro/di-tui/components"
	"github.com/acaloiaro/di-tui/config"
	"github.com/bradfitz/iter"
	ini "gopkg.in/ini.v1"
)

// Authenticate authenticates to the di.fm API with username and password, returning the listen token
func Authenticate(username, password string) (token string) {
	authURL := "https://api.audioaddict.com/v1/di/members/authenticate"
	client := &http.Client{}
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	encodedData := data.Encode()

	req, _ := http.NewRequest("POST", authURL, strings.NewReader(encodedData))
	req.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("authentication failed", err)
	}
	defer resp.Body.Close()

	var res authResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("Unable to authenticate to di.fm. Status code:", resp.StatusCode)
		os.Exit(1)
	}

	json.Unmarshal(body, &res)
	token = res.ListenKey
	config.SaveToken(token)

	return
}

// GetStreamURL extracts a playlist's stream URL from raw INI bytes (pls file)
func GetStreamURL(data []byte, ctx *context.AppContext) (streamURL string, ok bool) {
	cfg, err := ini.Load(data)
	if err != nil {
		ctx.SetStatusMessage("Unable to fetch channel playlist file.")
		ok = false
		return
	}

	streamURL = cfg.Section("playlist").Key("File1").String()
	ok = streamURL != ""

	return
}

// GetCurrentlyPlaying fetches the list of all currently playing tracks site-side
func GetCurrentlyPlaying(ctx *context.AppContext) (currentlyPlaying components.CurrentlyPlaying) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.audioaddict.com/v1/di/currently_playing", nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		msg := fmt.Sprintf("Unable to fetch currently playing track info: %v", resp.StatusCode)
		ctx.SetStatusMessage(msg)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		ctx.SetStatusMessage("Unable to fetch currently playing track info.")

		return
	}

	var currentlyPlayingStations []components.CurrentlyPlaying
	json.Unmarshal(body, &currentlyPlayingStations)

	for _, cp := range currentlyPlayingStations {
		if cp.ChannelID == ctx.CurrentChannel.ID {
			return cp
		}
	}

	return
}

// ListChannels lists all premium MP3 channels
func ListChannels(ctx *context.AppContext) (channels []components.ChannelItem) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://listen.di.fm/premium_high", nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		ctx.SetStatusMessage("Unable to fetch the list of channels")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.SetStatusMessage("Unable to fetch the list of channels")
		return
	}

	err = json.Unmarshal(body, &channels)
	if err != nil {
		ctx.SetStatusMessage("Unable to fetch the list of channels")
		return
	}

	return
}

// ListFavorites lists a user's favorite channels
func ListFavorites(ctx *context.AppContext) (favorites []components.FavoriteItem) {

	client := &http.Client{}
	url := fmt.Sprintf("%s?%s", "http://listen.di.fm/premium_high/favorites.pls", ctx.DifmToken)
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		ctx.SetStatusMessage("There was a problem fetching your favorites")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	cfg, err := ini.Load(body)
	if err != nil {
		ctx.SetStatusMessage("There was a problem fetching your favorites")
		return
	}

	sec := "playlist"
	numEntries := cfg.Section(sec).Key("NumberOfEntries").MustInt(0)
	for i := range iter.N(numEntries) {
		// di.fm's PLS keys begin at 1
		k := i + 1
		favorites = append(favorites, components.FavoriteItem{
			Name:        cfg.Section(sec).Key(fmt.Sprintf("Title%d", k)).String(),
			PlaylistURL: cfg.Section(sec).Key(fmt.Sprintf("File%d", k)).String(),
		})
	}

	return
}

// Stream streams the provided URL using the given di.fm premium token
func Stream(url string, ctx *context.AppContext) {
	client := &http.Client{}
	u := fmt.Sprintf("%s?%s", url, config.GetToken())
	req, _ := http.NewRequest("GET", u, nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		ctx.SetStatusMessage("There was a problem streaming audio.")
		return
	}

	go func() { player.Play(ctx, resp.Body) }()
	if err != nil {
		ctx.SetStatusMessage(fmt.Sprintf("There was a problem streaming audio: %s", err.Error()))
		return
	}
}

type authResponse struct {
	ListenKey string `json:"listen_key"`
}
