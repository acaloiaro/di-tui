package difm

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/acaloiaro/di-tui/context"
	"github.com/acaloiaro/di-tui/player"

	"github.com/acaloiaro/di-tui/components"
	"github.com/acaloiaro/di-tui/config"
	"github.com/bradfitz/iter"
	ini "gopkg.in/ini.v1"
)

type ApplicationMetadata struct {
	User struct {
		AudioToken string `json:"audio_token"`
		SessionKey string `json:"session_key"`
	} `json:"user"`
	CsrfToken string
}

type authResponse struct {
	ListenKey string `json:"listen_key"`
}

// Authenticate authenticates to the di.fm API with a username and password
//
// Note: There is a more API-friendly way of authenticating to the audioaddict API. However, because only the "web
// player" allows on-demand content, and di-tui uses the on-demand API to stream audio, we must mimic the web-player
// login workflow. I.e. the following
// 1. GET www.di.fm/login to get the CSRF token
// 2. POST www.di.fm/login (with CSRF token and other appropriate headers)
// 3. GET www.di.fm/ to retrieve two key pieces of information
//   - an "audio_token"
//   - a "session_key"
//
// Both of which are required to make content requests to the on-demand API
// This login workflow is cookie-based
func Authenticate(ctx *context.AppContext, username, password string) (token string) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}
	client := &http.Client{Jar: jar}

	// 1. GET www.di.fm/login to get the CSRF token
	meta, _ := getApplicationMetadata(ctx, client)
	authURL := "https://www.di.fm/login"
	data := url.Values{}
	data.Set("member_session[username]", username)
	data.Set("member_session[password]", password)

	encodedData := data.Encode()

	// 2. POST www.di.fm/login (with CSRF token and other appropriate headers)
	req, _ := http.NewRequest("POST", authURL, strings.NewReader(encodedData))
	req.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Origin", "https://www.di.fm")
	req.Header.Add("Referrer", "https://www.di.fm")
	req.Header.Add("Accept", "*/*;q=0.5, text/javascript, application/javascript, application/ecmascript, application/x-ecmascript")
	//req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:106.0) Gecko/20100101 Firefox/106.0")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("te", "trailers")
	req.Header.Add("X-CSRF-Token", meta.CsrfToken)

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

	// 3. GET www.di.fm/ to retrieve two key pieces of information
	meta, err = getApplicationMetadata(ctx, client)
	if err != nil {
		fmt.Println("Unable to fetch audio token and session key")
		os.Exit(1)
	}

	config.SaveAudioToken(meta.User.AudioToken)
	config.SaveSessionKey(meta.User.SessionKey)

	return
}

// GetApplicationMetadata fetches the application's metadata (the di.fm player application) from www.di.fm
func getApplicationMetadata(ctx *context.AppContext, client *http.Client) (appMeta ApplicationMetadata, err error) {
	var req *http.Request
	req, err = http.NewRequest("GET", "https://www.di.fm", nil)
	if err != nil {
		return
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:106.0) Gecko/20100101 Firefox/106.0")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Site", "same-origin")

	var resp *http.Response
	resp, err = client.Do(req)
	defer resp.Body.Close()
	if err != nil || resp.StatusCode != 200 {
		return
	}

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		return
	}

	bodyStr := string(body)
	re := regexp.MustCompile(`.*di\.app\.start\((.*)\);.*`)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) > 0 {
		appMeta = ApplicationMetadata{}
		err = json.Unmarshal([]byte(matches[1]), &appMeta)
		if err != nil {
			return
		}
	}

	re = regexp.MustCompile(`.*meta name="csrf-token" content="(.*?)"/>`)
	matches = re.FindStringSubmatch(bodyStr)
	if len(matches) > 0 {
		appMeta.CsrfToken = matches[1]
	}

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

// ListChannels lists all premium MP3 channels
func ListChannels(ctx *context.AppContext) (channels []components.ChannelItem) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://listen.di.fm/premium_high", nil)
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		ctx.SetStatusMessage(fmt.Sprintf("Unable to fetch the list of channels: %v", err))
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

type channelContent struct {
	Tracks []struct {
		Content struct {
			Assets []struct {
				Url string `json:"url"`
			} `json:"assets"`
		} `json:"content"`
	} `json:"tracks"`
}

// FetchContent fetches a channel's currently playing content on-demand. This is the on-demand sibling of the Stream()
// function, which is for streaming-only content.
func FetchContent(ctx *context.AppContext, channelID string) (err error) {
	client := http.Client{}

	// Fetch the list of tracks currently playing on this channel
	u := fmt.Sprintf("https://api.audioaddict.com/v1/di/routines/channel/%s?tune_in=false&audio_token=%s", channelID, config.GetAudioToken())
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:106.0) Gecko/20100101 Firefox/106.0")
	req.Header.Set("X-Session-Key", config.GetSessionKey())
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		os.Exit(1)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("Unable to play content:", resp.StatusCode)
	}
	var chnContent channelContent
	json.Unmarshal(body, &chnContent)

	// Fetch the first-playing track on this channel
	u = fmt.Sprintf("https:%s", chnContent.Tracks[0].Content.Assets[0].Url)
	ctx.SetStatusMessage(u)
	req, err = http.NewRequest("GET", u, nil)
	req.Header.Add("Range", "bytes=0-1000000")
	if err != nil {
		return
	}

	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != 206 {
		return
	}

	go func() { player.Play(ctx, resp.Body) }()

	return
}
