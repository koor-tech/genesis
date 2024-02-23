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
)

func init() {
	viper.SetDefault("GENESIS_DATA", "/koor/clients")
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
