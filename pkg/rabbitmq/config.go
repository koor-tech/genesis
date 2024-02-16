package rabbitmq

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
}

func NewConfig() *Config {
	return &Config{
		Host:     viper.GetString("RABBITMQ_HOST"),
		Port:     viper.GetInt("RABBITMQ_PORT"),
		User:     viper.GetString("RABBITMQ_USER"),
		Password: viper.GetString("RABBITMQ_PASSWORD"),
	}
}

func (c *Config) Url() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", c.User, c.Password, c.Host, c.Port)
}
