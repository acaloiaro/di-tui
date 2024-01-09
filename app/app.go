package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"io"
	"math"
	"net/http"
	"reflect"

	"github.com/acaloiaro/di-tui/components"
	"github.com/acaloiaro/di-tui/config"
	"github.com/acaloiaro/di-tui/context"
	"github.com/acaloiaro/di-tui/difm"
	"github.com/acaloiaro/di-tui/player"
	"github.com/nfnt/resize"
)

// Art fetches album art for the given track, converts it to ASCII, and return the ASCII stringified album art
func Art(ctx *context.AppContext, cp components.Track) (art string, err error) {
	if !config.AlbumArt() {
		return
	}

	// download track details
	var resp *http.Response
	url := fmt.Sprintf("https://api.audioaddict.com/v1/di/tracks/%d", cp.ID)
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	var trackDetails components.TrackDetails
	json.Unmarshal(body, &trackDetails)

	// fetch the album art and convert it to ascii
	url = fmt.Sprintf("https:%s", trackDetails.AlbumArtURL)
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	img, format, err := image.Decode(resp.Body)
	if format != "jpeg" && format != "png" || err != nil {
		err = errors.New("only jpeg and png images supported")
		return
	}

	_, _, windowWidth, windowHeight := ctx.View.NowPlaying.GetRect()
	scaleSize := math.Min(float64(windowWidth), float64(windowHeight))
	art = string(convertToAscii(scaleImage(img, int(scaleSize))))

	return
}

func scaleImage(img image.Image, w int) (image.Image, int, int) {
	sz := img.Bounds()
	h := (sz.Max.Y * w * 10) / (sz.Max.X * 16)
	img = resize.Resize(uint(w), uint(h), img, resize.Lanczos3)
	return img, w, h
}

func convertToAscii(img image.Image, w, h int) []byte {
	ASCIISTR := "MND8OZ$7I?+=~:,.."
	table := []byte(ASCIISTR)
	buf := new(bytes.Buffer)

	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			g := color.GrayModel.Convert(img.At(j, i))
			y := reflect.ValueOf(g).FieldByName("Y").Uint()
			pos := int(y * 16 / 255)
			_ = buf.WriteByte(table[pos])
		}
		_ = buf.WriteByte('\n')
	}

	return buf.Bytes()
}

// PlayChannel begins streaming the provided channel after fetching its playlist
// If a channel is already playing, the old stream is stopped first, clearing up resources.
// This function is *asynchronous* and creates a single streaming resource: the audio stream held by the application
// context. To clean up resources created by this function, Close() the application's audio stream.
func PlayChannel(ctx *context.AppContext, chn *components.ChannelItem) {
	player.Stop(ctx)

	if chn == nil {
		ctx.SetStatusMessage("Unable to play channel. Try again.")
		return
	}

	ctx.IsPlaying = true

	go func() {
		client := &http.Client{}
		req, _ := http.NewRequest("GET", chn.Playlist, nil)
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			ctx.SetStatusMessage(fmt.Sprintf("Unable to stream channel: %s. Try again.", chn.Name))
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			ctx.SetStatusMessage(fmt.Sprintf("Unable to stream channel: %s. Try again.", chn.Name))
			return
		}
		if streamURL, ok := difm.GetStreamURL(body, ctx); ok {
			UpdateNowPlaying(ctx, chn)
			difm.Stream(streamURL, ctx)
		}
	}()
}

// TogglePause pauses/unpauses audio when a channel is playing
func TogglePause(ctx *context.AppContext) {
	// nothing to do if nothing has been streamed
	if ctx.Player == nil {
		return
	}

	if ctx.IsPlaying {
		ctx.IsPlaying = false
		ctx.AudioStream.Close()
		player.Stop(ctx)
	} else {
		PlayChannel(ctx, ctx.CurrentChannel)
	}
}

// UpdateNowPlaying updates the application's now playing view with the currently playing channel and album art
// Artist and track information are fetched separately from album art, allowing for a more responsive UI
func UpdateNowPlaying(ctx *context.AppContext, chn *components.ChannelItem) {
	go func() {
		ctx.CurrentChannel = chn
		cp := difm.GetCurrentlyPlaying(ctx)

		ctx.View.App.QueueUpdateDraw(func() {
			ctx.View.NowPlaying.Channel = chn
			ctx.View.NowPlaying.Track = cp.Track
		})

		albumArt, err := Art(ctx, cp.Track)
		if err != nil {
			return
		}
		ctx.View.App.QueueUpdateDraw(func() {
			ctx.View.NowPlaying.Art = albumArt
		})
	}()
}
