package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	BaseUrl string `mapstructure:"bazarr_url"`
	Address string
	Port    string
	Protocol string
	ApiToken   string `mapstructure:"bazarr_token"`
	ComputedBaseUrl string
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
		fmt.Fprintln(os.Stderr,"Configuration Error: Unable to read config. ", err)
		os.Exit(1)
	}
	viper.Unmarshal(&cfg)

	var (
		computedBaseUrl string
		err error
	)

	if cfg.BaseUrl != "" {
		// If bazarr_url is provided, parse it directly
		parsedUrl, err := url.Parse(cfg.BaseUrl)
		if err != nil {
			fmt.Fprintln(os.Stderr, "URL Error: ", err)
			os.Exit(1)
		}
		cfg.Protocol = parsedUrl.Scheme
		cfg.Address = parsedUrl.Host
		computedBaseUrl = cfg.Protocol + "://" + cfg.Address
		cfg.ApiUrl = computedBaseUrl + "/api/"
	fmt.Fprintln(os.Stderr, "[DEBUG] ComputedBaseUrl:", cfg.ComputedBaseUrl)
	fmt.Fprintln(os.Stderr, "[DEBUG] ApiUrl:", cfg.ApiUrl)
	} else if strings.Contains(cfg.Address, "/") {
		// this is a check in case the Address is a subpath
		computedBaseUrl, err = url.JoinPath(cfg.Protocol + "://" + cfg.Address)
		if err != nil {
			fmt.Fprintln(os.Stderr, "URL Error: ", err)
			os.Exit(1)
		}
		apiUrl, _ := url.JoinPath(computedBaseUrl, "api/")
		cfg.ApiUrl = apiUrl
	} else {
		computedBaseUrl, err = url.JoinPath(cfg.Protocol + "://" + cfg.Address + ":" + cfg.Port)
		if err != nil {
			fmt.Fprintln(os.Stderr, "URL Error: ", err)
			os.Exit(1)
		}
		apiUrl, _ := url.JoinPath(computedBaseUrl, "api/")
		cfg.ApiUrl = apiUrl
	}

	cfg.ComputedBaseUrl = computedBaseUrl
}
