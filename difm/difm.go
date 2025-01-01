package difm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/acaloiaro/di-tui/components"
	"github.com/acaloiaro/di-tui/config"
	"github.com/acaloiaro/di-tui/context"
	"github.com/acaloiaro/di-tui/player"
	ini "gopkg.in/ini.v1"
)

var Networks = map[string]*components.Network{
	"classicalradio": {
		Name:      "Classical Radio",
		ListenURL: "http://listen.classicalradio.com",
		ShortName: "classicalradio",
	},
	"di": {
		Name:      "DI.fm",
		ListenURL: "http://listen.di.fm",
		ShortName: "di",
	},
	"radiotunes": {
		Name:      "RadioTunes",
		ListenURL: "http://listen.radiotunes.com",
		ShortName: "radiotunes",
	},
	"rockradio": {
		Name:      "ROCKRADIO.COM",
		ListenURL: "http://listen.rockradio.com",
		ShortName: "rockradio",
	},
	"jazzradio": {
		Name:      "JazzRadio",
		ListenURL: "http://listen.jazzradio.com",
		ShortName: "jazzradio",
	},
	"zenradio": {
		Name:      "Zen Radio",
		ListenURL: "http://listen.zenradio.com",
		ShortName: "zenradio",
	},
}

func GetNetwork(name string) (network *components.Network, err error) {
	var ok bool
	if network, ok = Networks[name]; !ok {
		return nil, fmt.Errorf("network does not exist: %s", network)
	}

	return
}

type authResponse struct {
	ListenKey string `json:"listen_key"`
	APIKey    string `json:"api_key"`
}

// Authenticate authenticates to the audio addict API
func Authenticate(ctx *context.AppContext, username, password string) (err error) {
	authURL := fmt.Sprintf("https://api.audioaddict.com/v1/%s/members/authenticate", ctx.Network.ShortName)
	log.Println(authURL)
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
	if err != nil {
		log.Fatal("unable to authenticate", err.Error())
	}

	json.Unmarshal(body, &res)
	err = json.Unmarshal([]byte(body), &res)
	if err != nil {
		log.Fatalf("unable to reason API response: %v", err)
	}

	config.SaveListenToken(res.ListenKey)
	config.SaveAPIKey(res.APIKey)
	config.SaveNetwork(ctx.Network)

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
	url := fmt.Sprintf("https://api.audioaddict.com/v1/%s/currently_playing", ctx.Network.ShortName)
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		ctx.SetStatusMessage("Unable to fetch currently playing track info")
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
	url := fmt.Sprintf("%s/premium_high", ctx.Network.ListenURL)
	req, _ := http.NewRequest("GET", url, nil)
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
	url := fmt.Sprintf("%s%s?%s", ctx.Network.ListenURL, "/premium_high/favorites.pls", ctx.DifmToken)
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
	for i := 0; i < numEntries; i++ {
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
	client := http.DefaultClient
	u := fmt.Sprintf("%s?%s", url, config.GetToken())

	// Keep increasing playback latency by one second for every time that the player exits with EOF
	// while ctx.IsPlaying
	for playbackLatency := 1; ctx.IsPlaying; playbackLatency++ {
		ctx.SetStatusMessage("Buffering stream...")
		if ctx.Player != nil {
			ctx.Player.Close()
		}

		req, _ := http.NewRequest("GET", u, nil)
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			ctx.SetStatusMessage("There was a problem streaming audio.")
			return
		}

		ctx.AudioStream = resp.Body
		audioBytes := &bytes.Buffer{}
		audioStream := bufio.NewReadWriter(bufio.NewReader(audioBytes), bufio.NewWriter(audioBytes))

		go func() {
			for {
				_, err = io.CopyN(audioStream, resp.Body, 512)
				if err != nil {
					return
				}
			}
		}()

		time.Sleep(time.Duration(playbackLatency) * time.Second)
		err = player.Play(ctx, audioStream, playbackLatency)
		if err == nil {
			return
		} else {
			resp.Body.Close()
			continue
		}
	}
}

// FavoriteItemChannel identifies the ChannelItem that corresponds with a FavoriteItem
func FavoriteItemChannel(ctx *context.AppContext, favorite components.FavoriteItem) (channel *components.ChannelItem) {
	// favorites are prefixed with "<NETWORK-NAME> - <CHANNEL NAME>", e.g. "DI.fm - Liquid Trap"
	// shave it off before comparing
	prefix := fmt.Sprintf("%s - ", ctx.Network.Name)
	for _, chn := range ctx.ChannelList {
		if chn.Name == favorite.Name[len(prefix):len(favorite.Name)] {
			channel = &chn
			return
		}
	}

	return
}
