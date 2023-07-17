package config

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/viper"
)

// TODO This application is migrating to having all of its configuration read into this `Config` struct, rather than
// using individual `ReadType` functions from viper. Migrating all `Get*` methods to use `Config` instead.
type Config struct {
	Theme map[string]string
}

var C Config

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/di-tui")
	viper.AddConfigPath("$HOME/.di-tui/")
	viper.AddConfigPath(".")

	viper.SetDefault("album_art", true)

	err := viper.ReadInConfig()
	if err != nil {
		saveConfig()
	}

	if err := viper.Unmarshal(&C); err != nil {
		fmt.Fprintf(os.Stderr, "unable to load di-tui configuration file: %v", err)
		os.Exit(1)
	}
}

// AlbumArt returns true if album art should be fetched when a new song begins playing
func AlbumArt() bool {
	return viper.GetBool("album_art")
}

// GetToken returns the di.fm API token if one is available
func GetToken() (token string) {
	return viper.GetString("token")
}

// GetUserID returns the di.fm user ID
func GetUserID() (userID int) {
	return viper.GetInt("user_id")
}

// GetAudioToken gets the audio token from disk
func GetAudioToken() string {
	return viper.GetString("audioToken")
}

// SaveAudioToken saves the audio token to disk
func SaveAudioToken(audioToken string) {
	viper.Set("audioToken", audioToken)

	saveConfig()
}

// GetSessionKey gets the session key from disk
func GetSessionKey() string {
	return viper.GetString("sessionKey")
}

// SaveSessionKey saves the session key to disk
func SaveSessionKey(sessionKey string) {
	viper.Set("sessionKey", sessionKey)

	saveConfig()
}

// SaveUserID saves the user's ID
func SaveUserID(userID int64) {
	viper.Set("user_id", userID)

	saveConfig()
}

// SaveListenToken saves the user's listen token
func SaveListenToken(token string) {
	viper.Set("token", token)

	saveConfig()
}

// HasTheme returns true if the di-tui config has a Theme set
func HasTheme() bool {
	if C.Theme["primary_color"] != "" {
		return true
	} else {
		return false
	}
}

// GetThemePrimaryColor gets the theme's primary color
func GetThemePrimaryColor() string {
	return C.Theme["primary_color"]
}

// GetThemePrimaryColor gets the theme's background color
func GetThemeBackgroundColor() string {
	return C.Theme["background_color"]
}

// GetThemePrimaryTextColor gets the theme's primary text color
func GetThemePrimaryTextColor() string {
	return C.Theme["primary_text_color"]
}

// GetThemePrimaryColor gets the theme's primary color
func GetThemeSecondaryTextColor() string {
	return C.Theme["secondary_text_color"]
}

func saveConfig() {
	viper.SetConfigFile(configFilePath())
	viper.SetConfigType("yaml")

	viper.Set("username", "")
	viper.Set("password", "")

	viper.WriteConfig()

}

func configFilePath() string {
	var home string
	if runtime.GOOS == "windows" {
		home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}

	home = os.Getenv("HOME")
	dir := fmt.Sprintf("%s/.config/di-tui/", home)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}

	return fmt.Sprintf("%s/config.yml", dir)
}
