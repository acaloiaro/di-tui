package main

import (
	"fmt"
	"os"
	"time"

	"github.com/acaloiaro/di-tui/app"
	"github.com/acaloiaro/di-tui/config"
	"github.com/acaloiaro/di-tui/context"
	"github.com/acaloiaro/di-tui/difm"
	"github.com/acaloiaro/di-tui/views"
	"github.com/rivo/tview"

	"github.com/gdamore/tcell"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var ctx *context.AppContext

func init() {
	// when true color is on, tcell does not respect your terminal colors
	os.Setenv("TCELL_TRUECOLOR", "disable")
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
		difm.Authenticate(username, password)
	}

	token = config.GetToken()
	if token == "" {
		fmt.Println("Authenticate by running: di-tui --username USER --password PASSWORD\n\n" +
			"Or, visit https://github.com/acaloiaro/di-tui#authenticate for other options.")
		os.Exit(1)
	}

	ctx = context.CreateAppContext(views.CreateViewContext())
	ctx.DifmToken = token

	run()
}

func run() {
	configureEventHandling()
	updateScreenLayout()
	configureUIComponents()

	err := ctx.View.App.Run()

	if err != nil {
		panic(err)
	}
}

func updateScreenLayout() {
	focusView := ctx.View.App.GetFocus()

	main := tview.NewFlex()
	main.SetDirection(tview.FlexRow)

	favsAndChannels := tview.NewFlex().
		AddItem(ctx.View.FavoriteList, 0, 1, false).
		AddItem(ctx.View.ChannelList, 0, 2, false).
		SetDirection(tview.FlexRow)

	primaryView := tview.NewFlex()
	primaryView.
		AddItem(favsAndChannels, 30, 0, false).
		AddItem(ctx.View.NowPlaying, 0, 4, false)

	if ctx.ShowStatus {
		main.AddItem(ctx.View.Status, 4, 0, false)
	}

	main.
		AddItem(primaryView, 0, 3, false).
		AddItem(ctx.View.Keybindings, 4, 0, false)

	if focusView == nil {
		focusView = ctx.View.FavoriteList
	}

	ctx.View.App.
		SetRoot(main, true).
		SetFocus(focusView)
}

func configureEventHandling() {
	ctx.View.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		focus := ctx.View.App.GetFocus().(*tview.List)

		switch event.Rune() {
		case 'c':
			if focus != ctx.View.ChannelList {
				ctx.View.App.SetFocus(ctx.View.ChannelList)
			}
		case 'f':
			if focus != ctx.View.FavoriteList {
				ctx.View.App.SetFocus(ctx.View.FavoriteList)
			}
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

	// keep 'now playing' up to date second-by-second
	go func() {
		c := time.Tick(1 * time.Second)
		for range c {
			elapsed := time.Since(ctx.View.NowPlaying.Track.StartTime)

			// If the current time is past the end of the track, then a new track is playing and the now playing track needs
			// to be refreshed.
			if ctx.View.NowPlaying.Track.Duration > 0 && ctx.View.NowPlaying.Track.Duration < elapsed.Seconds() {
				app.UpdateNowPlaying(ctx.CurrentChannel, ctx)
			}

			if ctx.CurrentChannel != nil && elapsed.Seconds() > 0 {
				ctx.View.NowPlaying.Elapsed = elapsed.Seconds()
			}

			ctx.View.App.QueueUpdateDraw(func() {})
		}
	}()

	// display the status pane when new status messages arrive
	go func() {
		for {
			status := <-ctx.StatusChannel
			ctx.View.Status.Message = status.Message
			updateScreenLayout() // add the status pane to the screen

			<-time.Tick(status.Duration)
			ctx.ShowStatus = false
			updateScreenLayout() // remove the status pane from the screen
		}
	}()

}

func configureUIComponents() {

	// configure the channel list
	channels := difm.ListChannels(ctx)
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
				// TODO: this feels a bit hacky -- consider doing something else.
				if chn.Name == f.Name[8:len(f.Name)] {
					app.PlayChannel(&chn, ctx)
					return
				}
			}
		})
	}

	// configure the keybinding view
	bindings := []views.UIKeybinding{
		views.UIKeybinding{Shortcut: "c", Description: "Channels", Func: func() {}},
		views.UIKeybinding{Shortcut: "f", Description: "Favorites", Func: func() {}},
		views.UIKeybinding{Shortcut: "j", Description: "Scroll Down", Func: func() {}},
		views.UIKeybinding{Shortcut: "k", Description: "Scroll Up", Func: func() {}},
		views.UIKeybinding{Shortcut: "q", Description: "Quit", Func: func() {}},
		views.UIKeybinding{Shortcut: "p", Description: "Pause", Func: func() {}},
		views.UIKeybinding{Shortcut: "Enter", Description: "Play", Func: func() {}},
	}

	ctx.View.Keybindings.Bindings = bindings
}
