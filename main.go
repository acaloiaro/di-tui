package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/acaloiaro/di-tui/app"
	"github.com/acaloiaro/di-tui/components"
	"github.com/acaloiaro/di-tui/config"
	"github.com/acaloiaro/di-tui/context"
	"github.com/acaloiaro/di-tui/difm"
	"github.com/acaloiaro/di-tui/mpris"
	"github.com/acaloiaro/di-tui/views"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var ctx *context.AppContext
var favoritesDebouncer *time.Timer

const VERSION = "1.14.1"

func main() {
	var err error
	usernameFlag := pflag.String("username", "", "your di.fm username")
	passwordFlag := pflag.String("password", "", "your di.fm password")
	versionFlag := pflag.Bool("version", false, "print the current di-tui version")
	networkFlag := pflag.String("network", viper.GetString("network.shortname"), "the audioaddict network to connect to")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if *versionFlag {
		fmt.Printf("di-tui %s\n", VERSION)
		return
	}
	network, err := difm.GetNetwork(*networkFlag)
	if err != nil {
		var networks []string
		for network := range difm.Networks {
			networks = append(networks, network)
		}
		fmt.Printf("Invalid network: %s \nPlease choose from the following: %s\n", *networkFlag, strings.Join(networks, ", "))
		return
	}

	if *usernameFlag != "" && *passwordFlag != "" {
		err = difm.Authenticate(ctx, *usernameFlag, *passwordFlag)
		if err != nil {
			fmt.Printf("unable to authenticate %s", err)
			return
		}
	}

	token := config.GetToken()
	if token == "" {
		fmt.Println("Authenticate by running: di-tui --username USER --password PASSWORD\n\n" +
			"Or, visit https://github.com/acaloiaro/di-tui#authenticate for other options.")
		os.Exit(1)
	}

	ctx = context.CreateAppContext(views.CreateViewContext(network))
	ctx.DifmToken = token

	run()
}

func run() {
	configureEventHandling()
	updateScreenLayout()
	FetchFavoritesAndChannels()

	ctx.View.Keybindings.Bindings = views.GetKeybindings()

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

// configureEventHandling handles key press events, and regular UI updates such as the currently playing track
func configureEventHandling() {
	ctx.View.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		favoritesEmpty := len(ctx.FavoriteList) == 0
		if favoritesEmpty {
			ctx.View.App.SetFocus(ctx.View.ChannelList)
		}

		focus := ctx.View.App.GetFocus().(*tview.List)
		switch event.Key() {
		case tcell.KeyEnter:
			current := focus.GetCurrentItem()
			if focus != ctx.View.ChannelList && current < len(ctx.FavoriteList) {
				highlightedFavorite := ctx.FavoriteList[current]
				ctx.HighlightedChannel = difm.FavoriteItemChannel(ctx, highlightedFavorite)
			} else if current < len(ctx.ChannelList) {
				ctx.HighlightedChannel = &ctx.ChannelList[current]
			}
			app.Play(ctx)
		case tcell.KeyRune:
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
			case 'j': // scroll down
				current := focus.GetCurrentItem() + 1
				focus.SetCurrentItem(current)
			case 'k': // scroll up
				current := focus.GetCurrentItem()
				if current > 0 {
					focus.SetCurrentItem(current - 1)
				}
			case 'J': // Shift+J: move favorite down
				if focus == ctx.View.FavoriteList {
					moveFavoriteDown(ctx, focus.GetCurrentItem())
				}
			case 'K': // Shift+K: move favorite up
				if focus == ctx.View.FavoriteList {
					moveFavoriteUp(ctx, focus.GetCurrentItem())
				}
			case 'F': // Shift+F: toggle favorite
				current := focus.GetCurrentItem()
				var channel *components.ChannelItem
				if focus == ctx.View.FavoriteList && current < len(ctx.FavoriteList) {
					channel = difm.FavoriteItemChannel(ctx, ctx.FavoriteList[current])
				} else if focus == ctx.View.ChannelList && current < len(ctx.ChannelList) {
					channel = &ctx.ChannelList[current]
				}
				if channel != nil {
					toggleFavorite(ctx, channel)
				}
			case 'p', 32: // tcell has no constant for the space bar rune (32)
				app.TogglePause(ctx)
			}
		}

		return event
	})

	// keep 'now playing' up to date second-by-second
	go func() {
		c := time.Tick(1 * time.Second)
		for range c {
			// don't show elapsed time updates when playback is paused/not playing
			if !ctx.IsPlaying {
				continue
			}
			elapsed := time.Since(ctx.View.NowPlaying.Track.StartTime)

			// If the current time is past the end of the track, then a new track is playing and the now playing track needs
			// to be refreshed.
			if ctx.View.NowPlaying.Track.Duration > 0 && ctx.View.NowPlaying.Track.Duration < elapsed.Seconds() {
				app.UpdateNowPlaying(ctx, ctx.CurrentChannel)
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
			ctx.ShowStatus = true
			ctx.View.Status.Message = status.Message
			updateScreenLayout() // add the status pane to the screen

			<-time.Tick(status.Duration)
			ctx.ShowStatus = false
			updateScreenLayout() // remove the status pane from the screen
		}
	}()

	// Start the mpris server for d-bus support
	mpris.Start(ctx)
}

func FetchFavoritesAndChannels() {
	ctx.View.ChannelList.Clear()
	ctx.View.FavoriteList.Clear()

	channels := difm.ListChannels(ctx)
	if len(channels) == 0 {
		ctx.SetStatusMessage("Unable to get the channel list.")
		return
	}
	ctx.ChannelList = channels

	for _, chn := range channels {
		ctx.View.ChannelList.AddItem(chn.Name, "", 0, func() {})
	}

	remoteFavs := difm.ListFavorites(ctx)
	// Populate ChannelID on remote favorites by matching channel names
	channelByName := make(map[string]*components.ChannelItem, len(channels))
	for i := range channels {
		channelByName[channels[i].Name] = &channels[i]
	}
	for i, rf := range remoteFavs {
		if ch, ok := channelByName[rf.Name]; ok {
			remoteFavs[i].ChannelID = ch.ID
		}
	}

	favorites := mergeFavorites(remoteFavs, config.GetLocalFavorites(ctx.Network.ShortName))
	for i, fav := range favorites {
		ctx.View.FavoriteList.InsertItem(i, fav.Name, "", 0, func() {})
	}
	ctx.FavoriteList = favorites

	if len(favorites) == 0 {
		if len(channels) > 0 {
			ctx.HighlightedChannel = &channels[0]
		}
		return
	}

	// default the highlighted channel to the first favorite; even before users select a channel manually. This way,
	// when di-tui starts and the user presses the "Play" media key, di-tui will start playing the first favorite
	// instead of requiring them to choose the channel to be played
	ctx.HighlightedChannel = difm.FavoriteItemChannel(ctx, ctx.FavoriteList[0])
}

// mergeFavorites combines local config favorites with remote favorites.
// Local favorites define ordering; remote favorites not present locally are appended.
// Local entries with hidden=true suppress both themselves and their remote counterpart.
// If there are no local favorites, remote favorites are returned as-is.
func mergeFavorites(remoteFavs, localFavs []components.FavoriteItem) []components.FavoriteItem {
	if len(localFavs) == 0 {
		return remoteFavs
	}

	localIDs := make(map[int64]bool, len(localFavs))
	result := make([]components.FavoriteItem, 0, len(localFavs)+len(remoteFavs))
	for _, lf := range localFavs {
		localIDs[lf.ChannelID] = true
		if !lf.Hidden {
			result = append(result, lf)
		}
	}
	for _, rf := range remoteFavs {
		if !localIDs[rf.ChannelID] {
			result = append(result, rf)
		}
	}
	return result
}

func moveFavoriteUp(ctx *context.AppContext, current int) {
	if current <= 0 || current >= len(ctx.FavoriteList) {
		return
	}
	newCurrentIDx := current - 1
	ctx.FavoriteList[current], ctx.FavoriteList[newCurrentIDx] = ctx.FavoriteList[newCurrentIDx], ctx.FavoriteList[current]
	buildFavoriteList(ctx)
	ctx.View.FavoriteList.SetCurrentItem(newCurrentIDx)
	saveFavoritesDebounced(ctx)
}

func moveFavoriteDown(ctx *context.AppContext, current int) {
	if current < 0 || current >= len(ctx.FavoriteList)-1 {
		return
	}
	newCurrentIDx := current + 1
	ctx.FavoriteList[current], ctx.FavoriteList[newCurrentIDx] = ctx.FavoriteList[newCurrentIDx], ctx.FavoriteList[current]
	buildFavoriteList(ctx)
	ctx.View.FavoriteList.SetCurrentItem(newCurrentIDx)
	saveFavoritesDebounced(ctx)
}

// saveFavoritesDebounced defers saveFavorites by 200ms, resetting the timer on each call.
func saveFavoritesDebounced(ctx *context.AppContext) {
	if favoritesDebouncer != nil {
		favoritesDebouncer.Stop()
	}
	favoritesDebouncer = time.AfterFunc(200*time.Millisecond, func() { saveFavorites(ctx) })
}

// saveFavorites persists the current visible order while preserving hidden entries.
func saveFavorites(ctx *context.AppContext) {
	var hidden []components.FavoriteItem
	for _, lf := range config.GetLocalFavorites(ctx.Network.ShortName) {
		if lf.Hidden {
			hidden = append(hidden, lf)
		}
	}
	config.SaveLocalFavorites(ctx.Network.ShortName, append(ctx.FavoriteList, hidden...))
}

func buildFavoriteList(ctx *context.AppContext) {
	ctx.View.FavoriteList.Clear()
	for i, fav := range ctx.FavoriteList {
		ctx.View.FavoriteList.InsertItem(i, fav.Name, "", 0, func() {})
	}
}

// toggleFavorite adds or removes channel from the favorites list.
// Removals are persisted as hidden=true so remote favorites for the same channel stay suppressed.
func toggleFavorite(ctx *context.AppContext, channel *components.ChannelItem) {
	for i, fav := range ctx.FavoriteList {
		if fav.ChannelID == channel.ID {
			ctx.FavoriteList = append(ctx.FavoriteList[:i], ctx.FavoriteList[i+1:]...)
			buildFavoriteList(ctx)
			setLocalFavoriteHidden(ctx.Network.ShortName, channel.ID, channel.Name, true)
			return
		}
	}
	ctx.FavoriteList = append(ctx.FavoriteList, components.FavoriteItem{
		ChannelID: channel.ID,
		Name:      channel.Name,
	})
	buildFavoriteList(ctx)
	setLocalFavoriteHidden(ctx.Network.ShortName, channel.ID, channel.Name, false)
}

// setLocalFavoriteHidden updates or inserts a local favorite entry with the given hidden state.
func setLocalFavoriteHidden(network string, channelID int64, name string, hidden bool) {
	localFavs := config.GetLocalFavorites(network)
	for i, lf := range localFavs {
		if lf.ChannelID == channelID {
			localFavs[i].Hidden = hidden
			config.SaveLocalFavorites(network, localFavs)
			return
		}
	}
	config.SaveLocalFavorites(network, append(localFavs, components.FavoriteItem{
		Name:      name,
		ChannelID: channelID,
		Hidden:    hidden,
	}))
}
