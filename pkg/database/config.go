package database

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Host       string
	Port       int
	User       string
	Password   string
	Name       string
	SSLEnabled bool
}

func NewConfig() *Config {
	return &Config{
		Host:       viper.GetString("DATABASE_HOST"),
		Port:       viper.GetInt("DATABASE_PORT"),
		User:       viper.GetString("DATABASE_USER"),
		Password:   viper.GetString("DATABASE_PASSWORD"),
		Name:       viper.GetString("DATABASE_NAME"),
		SSLEnabled: viper.GetBool("DATABASE_SSL_ENABLED"),
	}
}

func (c *Config) Uri() string {
	ssl := "sslmode=disable"
	if c.SSLEnabled {
		ssl = ""
	}
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?%s", c.User, c.Password, c.Host, c.Port, c.Name, ssl)
}
