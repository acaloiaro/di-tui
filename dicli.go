package main

import (
	"fmt"
	"os"

	"github.com/acaloiaro/dicli/app"
	"github.com/acaloiaro/dicli/config"
	"github.com/acaloiaro/dicli/context"
	"github.com/acaloiaro/dicli/difm"
	"github.com/acaloiaro/dicli/views"
	"github.com/rivo/tview"

	"github.com/gdamore/tcell"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var ctx *context.AppContext

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

	ctx = context.CreateAppContext(views.CreateAppView())
	ctx.DifmToken = token

	run()
}

func run() {
	configureEventHandling()
	configureUIComponents()
	layout := buildUILayout()

	err := ctx.View.App.
		SetRoot(layout, true).
		SetFocus(ctx.View.FavoriteList).
		Run()

	if err != nil {
		panic(err)
	}
}

func buildUILayout() *tview.Flex {
	main := tview.NewFlex()
	main.SetDirection(tview.FlexRow)

	favsAndChannels := tview.NewFlex().
		AddItem(ctx.View.FavoriteList, 0, 1, false).
		AddItem(ctx.View.ChannelList, 0, 2, false).
		SetDirection(tview.FlexRow)

	flex := tview.NewFlex()
	flex.
		AddItem(favsAndChannels, 0, 1, false).
		AddItem(ctx.View.NowPlaying, 0, 2, false)

	main.
		AddItem(flex, 0, 15, false).
		AddItem(ctx.View.Keybindings, 0, 1, false)

	return main
}

func configureEventHandling() {

	ctx.View.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		focus := ctx.View.App.GetFocus().(*tview.List)

		switch event.Rune() {
		case 'q':
			ctx.View.App.Stop()
		case 'j': //scroll down
			focus.SetCurrentItem(focus.GetCurrentItem() + 1)
		case 'k': //scroll up
			current := focus.GetCurrentItem()
			if current > 0 {
				focus.SetCurrentItem(current - 1)
			}
		case 'p': // pause/resume
			app.TogglePause(ctx)
		}

		return event
	})
}

func configureUIComponents() {

	// configure the channel list
	channels := difm.ListChannels()
	for _, chn := range channels {
		ctx.View.ChannelList.AddItem(chn.Name, "", 0, func() {
			chn := channels[ctx.View.ChannelList.GetCurrentItem()]
			app.PlayChannel(&chn, ctx)
		})
	}

	favorites := difm.ListFavorites(ctx)
	for _, fav := range favorites {
		ctx.View.FavoriteList.AddItem(fav.Name, "", 0, func() {
			f := favorites[ctx.View.FavoriteList.GetCurrentItem()]
			for _, chn := range channels {
				// favorites are prefixed with "DI.fm - <CHANNEL NAME>", shave it off before comparing
				if chn.Name == f.Name[8:len(f.Name)] {
					app.PlayChannel(&chn, ctx)
					return
				}
			}
		})
	}

	// configure the keybinding view
	bindings := []views.UIKeybinding{
		views.UIKeybinding{Shortcut: "q", Description: "Quit", Func: func() {}},
		views.UIKeybinding{Shortcut: "p", Description: "Pause", Func: func() {}},
		views.UIKeybinding{Shortcut: "j", Description: "Scroll Up", Func: func() {}},
		views.UIKeybinding{Shortcut: "k", Description: "Scroll Down", Func: func() {}},
		views.UIKeybinding{Shortcut: "Enter", Description: "Play Selected", Func: func() {}},
	}

	ctx.View.Keybindings.Bindings = bindings
}
