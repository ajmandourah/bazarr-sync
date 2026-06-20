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
	c, err := LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Configuration Error: ", err)
		os.Exit(1)
	}
	cfg = c
}

func LoadConfig() (Config, error) {
	var c Config

	if CfgFile != "" {
		viper.SetConfigFile(CfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return c, fmt.Errorf("unable to read config: %w", err)
	}

	if err := viper.Unmarshal(&c); err != nil {
		return c, fmt.Errorf("unable to parse config: %w", err)
	}

	if c.BaseUrl == "" {
		return c, fmt.Errorf("bazarr_url is required")
	}

	parsedUrl, err := url.Parse(c.BaseUrl)
	if err != nil {
		return c, fmt.Errorf("invalid URL: %w", err)
	}
	c.ApiUrl = parsedUrl.Scheme + "://" + parsedUrl.Host + "/api/"

	return c, nil
}
