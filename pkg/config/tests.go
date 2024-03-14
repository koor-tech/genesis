package config

import (
	"fmt"

	"github.com/creasty/defaults"
	"go.uber.org/fx"
)

var TestModule = fx.Module("config_test",
	fx.Provide(
		LoadTestConfig,
	),
)

func LoadTestConfig() (*Config, error) {
	c := &Config{}
	if err := defaults.Set(c); err != nil {
		return nil, fmt.Errorf("failed to set config defaults: %w", err)
	}

	return c, nil
}
