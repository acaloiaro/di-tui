package main

import (
	"fmt"
	"os"

	"github.com/acaloiaro/dicli/app"
	"github.com/acaloiaro/dicli/config"
	"github.com/acaloiaro/dicli/context"
	"github.com/acaloiaro/dicli/difm"

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
		fmt.Println("First, authenticate with by running: dicli --username USER --password PASSWORD")
		os.Exit(1)
	}

	ctx = context.CreateAppContext()
	ctx.DifmToken = token

	run()
}

func run() {
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
			app.TogglePause(ctx)
		}

		return event
	})

	channels := difm.ListChannels()
	for _, chn := range channels {
		ctx.View.ChannelList.AddItem(chn.Name, "", 0, func() {
			go func() {
				chn := channels[ctx.View.ChannelList.GetCurrentItem()]
				app.PlayChannel(&chn, ctx)
			}()
		})
	}

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
