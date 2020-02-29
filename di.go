package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/acaloiaro/dicli/components"
	"github.com/acaloiaro/dicli/config"
	"github.com/acaloiaro/dicli/context"
	"github.com/acaloiaro/dicli/difm"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

/* di.fm API
Track details: http://www.di.fm/tracks/<track id>
Listen history: POST /_papi/v1/di/listen_history
       Payload: {track_id: 2918701, playlist_id: 63675}
Currently playing (all stations): https://www.di.fm/_papi/v1/di/currently_playing
Skip track: https://www.di.fm/_papi/v1/di/skip_events
*/
var ctx *context.AppContext

func init() {

	ctx = context.CreateAppContext()
	ctx.View.NowPlaying.
		SetBorder(true).
		SetTitle(" Now Playing ")
	ctx.View.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			ctx.View.App.Stop()
		case 'j': //scroll down
			ctx.View.ChannelList.SetCurrentItem(ctx.View.ChannelList.GetCurrentItem() + 1)
		case 'k': //scroll up
			current := ctx.View.ChannelList.GetCurrentItem()
			if current > 0 {
				ctx.View.ChannelList.SetCurrentItem(current - 1)
			}
		case 'p': // pause/resume
			togglePause()
		}

		return event
	})

	channels := difm.ListChannels()
	for _, chn := range channels {
		ctx.View.ChannelList.AddItem(chn.Name, "", 0, func() {
			go func() {
				chn := channels[ctx.View.ChannelList.GetCurrentItem()]
				playChannel(&chn)
			}()
		})
	}
}

func main() {
	pflag.String("username", "", "your di.fm username")
	pflag.String("password", "", "your di.fm password")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	username := viper.GetString("username")
	password := viper.GetString("password")
	var token string
	if len(username) > 0 && len(password) > 0 {
		token = difm.Authenticate(username, password)
	}

	token = config.GetToken()
	if token == "" {
		fmt.Println("First, authenticate with by running: dicli -username USER -password PASSWORD")
		os.Exit(1)
	}

	ctx.DifmToken = token

	run()
}

func run() {

	flex := tview.NewFlex()
	flex.
		AddItem(ctx.View.ChannelList, 0, 1, false).
		AddItem(ctx.View.NowPlaying, 0, 2, false)

	err := ctx.View.App.
		SetRoot(flex, true).
		SetFocus(ctx.View.ChannelList).
		Run()

	if err != nil {
		panic(err)
	}
}

// togglePause pauses/unpauses audio when a channel is playing
func togglePause() {

	// nothing to do if nothing has been streamed
	if ctx.AudioStream == nil {
		return
	}

	ctx.AudioStream.Close()
	if !ctx.IsPlaying {
		playChannel(ctx.CurrentChannel)
	}

	ctx.IsPlaying = !ctx.IsPlaying
}

// stream streams the provided URL using the given di.fm premium token
func stream(url string) {
	client := &http.Client{}
	u := fmt.Sprintf("%s?%s", url, ctx.DifmToken)
	req, _ := http.NewRequest("GET", u, nil)
	resp, err := client.Do(req)
	if err != nil {
		// TODO: Don't exit here. Once there's a status message area in the app, populate it with the error
		log.Println("unable to stream channel", err.Error())
		os.Exit(1)
	}

	var format beep.Format
	ctx.AudioStream, format, err = mp3.Decode(resp.Body)
	if err != nil {
		// TODO: Don't exit here. Once there's a status message area in the app, populate it with the error
		log.Println("unable to stream channel:", resp.StatusCode)
		os.Exit(1)
	}

	if !ctx.SpeakerInitialized {
		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	}

	speaker.Play(ctx.AudioStream)
	ctx.IsPlaying = true
}

// playChannel begins streaming the provided channel after fetching its playlist
// If a channel is already playing, the old stream is stopped first, clearing up resources.
// This function is asynchronous and creates a single streaming resource: he audio stream held by the application
// context. To clean up resources created by this function, Close() the application's audio stream.
func playChannel(chn *components.ChannelItem) {

	// when other channels are already playing, close their stream before playing a new one
	if ctx.AudioStream != nil {
		ctx.AudioStream.Close()
	}

	go func() {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", chn.Playlist, nil)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if streamURL, ok := difm.GetStreamURL(body); ok {
			stream(streamURL)
			setCurrentChannel(chn)
		}
	}()
}

func setCurrentChannel(chn *components.ChannelItem) {
	ctx.CurrentChannel = chn
	ctx.View.App.QueueUpdateDraw(func() {
		ctx.View.NowPlaying.SetChannel(chn)
	})
}
