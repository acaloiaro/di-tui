package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	ini "gopkg.in/ini.v1"
)

/* di.fm API
Track details: http://www.di.fm/tracks/<track id>
Listen history: POST /_papi/v1/di/listen_history
       Payload: {track_id: 2918701, playlist_id: 63675}
Currently playing (all stations): https://www.di.fm/_papi/v1/di/currently_playing
Skip track: https://www.di.fm/_papi/v1/di/skip_events
*/
var ctrls = make(chan int)
var done = make(chan bool)
var chanList *tview.List
var nowPlaying *nowPlayingView
var app *tview.Application
var paused bool
var token string
var speakerInitialized bool
var audioStream beep.StreamSeekCloser
var currentChannel = &channelItem{Name: "N/A"}

const (
	// CTRLPAUSE pauses playback
	CTRLPAUSE = iota
	// CTRLRESUME resumes playback
	CTRLRESUME
	// CTRLPLAY begins playback
	CTRLPLAY
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/dicli")
	viper.AddConfigPath("$HOME/.dicli/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		writeConfig()
	}
}

func init() {
	app = tview.NewApplication()
	chanList = createChannelList()
	nowPlaying = newNowPlaying(currentChannel)
	nowPlaying.
		SetBorder(true).
		SetTitle(" Now Playing ")

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			app.Stop()
		case 'j': //scroll down
			chanList.SetCurrentItem(chanList.GetCurrentItem() + 1)
		case 'k': //scroll up
			current := chanList.GetCurrentItem()
			if current > 0 {
				chanList.SetCurrentItem(chanList.GetCurrentItem() - 1)
			}
		case 'p': // pause/resume
			ctrls <- CTRLPAUSE
		}

		return event
	})
}

func main() {
	auth()
	controlListener()
	drawUI()
}

func auth() {
	pflag.String("username", "", "your di.fm username")
	pflag.String("password", "", "your di.fm password")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	username := viper.GetString("username")
	password := viper.GetString("password")
	if len(username) > 0 && len(password) > 0 {
		authenticate(username, password)
	}

	token = viper.GetString("token")
	if token == "" {
		fmt.Println("First, authenticate with by running: dicli -username USER -password PASSWORD")
		os.Exit(1)
	}
}

func authenticate(username, password string) {
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
	if err != nil {
		fmt.Println("unable to authenticate", err.Error())
		os.Exit(1)
	}

	json.Unmarshal(body, &res)
	viper.Set("username", "")
	viper.Set("password", "")
	viper.Set("token", res.ListenKey)

	writeConfig()
}

func writeConfig() {
	viper.SetConfigFile(configFilePath())
	viper.SetConfigType("yaml")
	viper.WriteConfig()
}

func drawUI() {

	flex := tview.NewFlex()
	flex.
		AddItem(chanList, 0, 1, false).
		AddItem(nowPlaying, 0, 2, false)

	if err := app.
		SetRoot(flex, true).
		SetFocus(chanList).
		Run(); err != nil {
		panic(err)
	}
}

func createChannelList() *tview.List {
	channels := list()
	list := tview.NewList()
	list.
		ShowSecondaryText(false).
		SetBorder(true).
		SetTitle(" Channels ")

	for _, chn := range channels {
		list.AddItem(chn.Name, "", 0, func() {
			go func() {
				chn := channels[list.GetCurrentItem()]
				app.QueueUpdateDraw(func() {
					nowPlaying.setChannel(&chn)
				})

				currentChannel = &chn
				ctrls <- CTRLPLAY
			}()
		})
	}

	return list
}

func controlListener() {
	go func() {
		for {
			switch <-ctrls {
			case CTRLPAUSE:
				// nothing to do if nothing has been streamed
				if audioStream == nil {
					continue
				}

				audioStream.Close()
				if paused {
					playChannel(currentChannel)
					paused = false
				} else {
					paused = true
				}
			case CTRLPLAY, CTRLRESUME:
				if audioStream != nil {
					audioStream.Close()
				}

				playChannel(currentChannel)
			}
		}
	}()
}

// stream streams the provided URL using the given di.fm premium token
func stream(url string) {
	client := &http.Client{}
	u := fmt.Sprintf("%s?%s", url, token)
	req, _ := http.NewRequest("GET", u, nil)
	resp, err := client.Do(req)
	if err != nil {
		// TODO: Don't exit here. Once there's a status message area in the app, populate it with the error
		log.Println("unable to stream channel", err.Error())
		os.Exit(1)
	}

	var format beep.Format
	audioStream, format, err = mp3.Decode(resp.Body)
	if err != nil {
		// TODO: Don't exit here. Once there's a status message area in the app, populate it with the error
		log.Println("unable to stream channel:", resp.StatusCode)
		os.Exit(1)
	}

	if !speakerInitialized {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	}

	speaker.Play(audioStream)
}

func playChannel(chn *channelItem) {
	go func() {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", chn.Playlist, nil)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if streamURL, ok := getStreamURL(body); ok {
			currentChannel = chn
			stream(streamURL)
		}
	}()
}

// listChannels lists all premium MP3 channels
func list() (channels []channelItem) {
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
		log.Panicf("unable to fetch channel list: %e", err)
	}

	return
}

// getStreamURL extracts a playlist's stream URL from raw INI bytes (pls file)
func getStreamURL(data []byte) (streamURL string, ok bool) {
	cfg, err := ini.Load(data)
	if err != nil {
		fmt.Printf("playlist file parsing failed : %v\n", err.Error())
		return
	}

	streamURL = cfg.Section("playlist").Key("File1").String()
	ok = streamURL != ""

	return
}

// nowPlayingView is a custom view for dispalying the currently playing channel
type nowPlayingView struct {
	*tview.Box
	channel *channelItem
}

func newNowPlaying(chn *channelItem) *nowPlayingView {
	return &nowPlayingView{
		Box:     tview.NewBox(),
		channel: chn,
	}
}

// Draw draws this primitive onto the screen.
func (n *nowPlayingView) Draw(screen tcell.Screen) {
	n.Box.Draw(screen)
	x, y, width, _ := n.GetInnerRect()

	line := fmt.Sprintf(`%s[white]  %s`, "Channel:", n.channel.Name)
	tview.Print(screen, line, x, y, width, tview.AlignLeft, tcell.ColorYellow)
}

func (n *nowPlayingView) setChannel(chn *channelItem) {
	n.channel = chn
}

// channelItem contains di.fm channel metadata
type channelItem struct {
	ID       int64  `json:"id"`
	Key      string `json:"key"`
	Name     string `json:"name"`
	Playlist string `json:"playlist"`
}

func configFilePath() string {
	var home string
	if runtime.GOOS == "windows" {
		home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	} else {
		home = os.Getenv("HOME")
	}

	dir := fmt.Sprintf("%s/.config/dicli/", home)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}

	return fmt.Sprintf("%s/config.yml", dir)
}

type authResponse struct {
	ListenKey string `json:"listen_key"`
}
