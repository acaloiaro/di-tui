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
	viper.AddConfigPath("$HOME/.config/dicli")
	viper.AddConfigPath("$HOME/.dicli/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		saveConfig()
	}
}

func SaveToken(token string) {
	viper.Set("username", "")
	viper.Set("password", "")
	viper.Set("token", token)

	saveConfig()
}

func GetToken() (token string) {
	return viper.GetString("token")
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
