package difm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/acaloiaro/di-tui/context"

	"github.com/acaloiaro/di-tui/components"
	"github.com/acaloiaro/di-tui/config"
	"github.com/bradfitz/iter"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	ini "gopkg.in/ini.v1"
)

/* di.fm API
Track details: http://www.di.fm/tracks/<track id>
Listen history: POST /_papi/v1/di/listen_history
       Payload: {track_id: 2918701, playlist_id: 63675}
Currently playing (all stations): https://www.di.fm/_papi/v1/di/currently_playing
Skip track: https://www.di.fm/_papi/v1/di/skip_events
Favorites: http://listen.di.fm/public3/favorites.pls?<listen key>
*/

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
	body, err := ioutil.ReadAll(resp.Body)
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
func GetStreamURL(data []byte) (streamURL string, ok bool) {
	cfg, err := ini.Load(data)
	if err != nil {
		fmt.Printf("playlist file parsing failed : %v\n", err.Error())
		return
	}

	streamURL = cfg.Section("playlist").Key("File1").String()
	ok = streamURL != ""

	return
}

// GetCurrentlyPlaying fetches the list of all currently playing tracks site-side
func GetCurrentlyPlaying(ctx *context.AppContext) (currentlyPlaying components.CurrentlyPlaying) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://www.di.fm/_papi/v1/di/currently_playing", nil)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// TODO: Don't exit here. Once there's a status message area in the app, populate it with the error
		log.Println("unable to list channels", err.Error())
		os.Exit(1)
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
func ListChannels() (channels []components.ChannelItem) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://listen.di.fm/premium_high", nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// TODO: Don't exit here. Once there's a status message area in the app, populate it with the error
		log.Println("unable to list channels", err.Error())
		os.Exit(1)
	}

	err = json.Unmarshal(body, &channels)
	if err != nil {
		// TODO: Don't exit here. Once there's a status message area in the app, populate it with the error
		log.Panicf("unable to fetch channel list: %e", err)
	}

	return
}

// ListFavorites lists a user's favorite channels
func ListFavorites(ctx *context.AppContext) (favorites []components.FavoriteItem) {

	client := &http.Client{}
	url := fmt.Sprintf("%s?%s", "http://listen.di.fm/public3/favorites.pls", ctx.DifmToken)
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	cfg, err := ini.Load(body)
	if err != nil {
		// TODO: show an message in a status UI element of the app
		fmt.Printf("favorite list parsing failed: %v\n", err.Error())
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
func Stream(url string, ctx *context.AppContext) (format beep.Format) {
	client := &http.Client{}
	u := fmt.Sprintf("%s?%s", url, config.GetToken())
	req, _ := http.NewRequest("GET", u, nil)
	resp, err := client.Do(req)
	if err != nil {
		// TODO: Don't exit here. Once there's a status message area in the app, populate it with the error
		log.Println("unable to stream channel", err.Error())
		os.Exit(1)
	}

	ctx.AudioStream, format, err = mp3.Decode(resp.Body)
	if err != nil {
		// TODO: Don't exit here. Once there's a status message area in the app, populate it with the error
		log.Println("unable to stream channel:", resp.StatusCode)
		os.Exit(1)
	}

	return
}

type authResponse struct {
	ListenKey string `json:"listen_key"`
}
