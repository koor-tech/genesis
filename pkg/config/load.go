package config

import (
	"fmt"
	"os"

	"github.com/creasty/defaults"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

var Module = fx.Module("config",
	fx.Provide(
		Load,
	),
)

type Result struct {
	fx.Out

	Config *Config
}

func Load() (Result, error) {
	// Viper Config reading setup
	viper.SetEnvPrefix("GENESIS")
	viper.SetConfigType("yaml")

	if configFile := os.Getenv("GENESIS_CONFIG_FILE"); configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/config")
	}

	res := Result{}
	// Find and read the config file
	if err := viper.ReadInConfig(); err != nil {
		return res, fmt.Errorf("fatal error config file: %w", err)
	}

	c := &Config{}
	if err := defaults.Set(c); err != nil {
		return res, fmt.Errorf("failed to set config defaults: %w", err)
	}
	res.Config = c

	if err := viper.Unmarshal(c); err != nil {
		return res, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return res, nil
}
