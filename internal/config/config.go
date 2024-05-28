package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Address string
	Port string
	Protocol string
	ApiToken string
	BazarrUrl string
	ApiUrl string
}

var cfg Config

var CfgFile string


func GetConfig() Config {
	return cfg
}
func InitConfig() {
	if CfgFile != "" {
		viper.SetConfigFile(CfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Fprintln(os.Stderr,"Configuration Error: ", err)
		fmt.Fprintln(os.Stderr,"Please supply a config.yaml file by using the flag --config or by placing the file in the same directory as bazarr-sync")
		os.Exit(1)
	}
	viper.Unmarshal(&cfg)

	//Bazarr url
	baseUrl, err := url.JoinPath(cfg.Protocol + "://" + cfg.Address + ":" + cfg.Port)
	if err != nil{
		fmt.Fprintln(os.Stderr, "URL Error: ", err)
	}
	apiUrl, _:= url.JoinPath(baseUrl,"api/")

	cfg.BazarrUrl = baseUrl
	cfg.ApiUrl = apiUrl
	
}
