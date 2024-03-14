package genesis

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	DataDir string
}

const (
	templateDirPrefix = "templates"
	clientsDirPrefix  = "clients"
	chartsDirPrefix   = "charts"
)

func init() {
	viper.SetDefault("GENESIS_DATA", "/koor")
}

func NewConfig() *Config {
	return &Config{
		DataDir: viper.GetString("GENESIS_DATA"),
	}
}

func (c *Config) TemplatesDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, templateDirPrefix)
}

func (c *Config) ClientsDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, clientsDirPrefix)
}

func (c *Config) ChartsDir() string {
	return fmt.Sprintf("%s/%s", c.DataDir, chartsDirPrefix)
}
