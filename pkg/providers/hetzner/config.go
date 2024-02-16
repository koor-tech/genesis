package hetzner

import (
	"github.com/spf13/viper"
)

type Config struct {
	Token string
}

func NewConfig() *Config {
	return &Config{
		Token: viper.GetString("HETZNER_TOKEN"),
	}
}
