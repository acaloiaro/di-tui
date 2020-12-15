package config

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/viper"
)

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
}

// AlbumArt returns true if album art should be fetched when a new song begins playing
func AlbumArt() bool {
	return viper.GetBool("album_art")
}

// GetToken returns the di.fm API token if one is available
func GetToken() (token string) {
	return viper.GetString("token")
}

// SaveToken persists the di.fm API token to disk
func SaveToken(token string) {
	viper.Set("username", "")
	viper.Set("password", "")
	viper.Set("token", token)

	saveConfig()
}

func saveConfig() {
	viper.SetConfigFile(configFilePath())
	viper.SetConfigType("yaml")
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
