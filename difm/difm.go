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

	"github.com/acaloiaro/dicli/context"

	"github.com/acaloiaro/dicli/components"
	"github.com/acaloiaro/dicli/config"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	ini "gopkg.in/ini.v1"
)

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

	log.Println("Token", res)
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
		log.Panicf("unable to fetch channel list: %e", err)
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
