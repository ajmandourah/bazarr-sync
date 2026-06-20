package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	BaseUrl  string `mapstructure:"bazarr_url"`
	ApiToken string `mapstructure:"bazarr_token"`
	ApiUrl   string
}

var cfg Config

var CfgFile string

func GetConfigFile() string {
	if CfgFile != "" {
		return CfgFile
	}
	return "config.yaml"
}

func GetConfig() Config {
	return cfg
}

func SetConfig(newCfg Config) {
	cfg = newCfg
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
		fmt.Fprintln(os.Stderr, "Configuration Error: Unable to read config. ", err)
		os.Exit(1)
	}
	viper.Unmarshal(&cfg)

	if cfg.BaseUrl == "" {
		fmt.Fprintln(os.Stderr, "Configuration Error: bazarr_url is required")
		os.Exit(1)
	}

	parsedUrl, err := url.Parse(cfg.BaseUrl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "URL Error: ", err)
		os.Exit(1)
	}
	cfg.ApiUrl = parsedUrl.Scheme + "://" + parsedUrl.Host + "/api/"
}
