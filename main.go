package main

import (
	"github.com/koor-tech/genesis/gateway"
	"github.com/spf13/viper"
	"log"
)

func main() {
	loadConfiguration()
	r := gateway.SetupRouter()
	r.Run() // listen and serve on 0.0.0.0:8080
}

func loadConfiguration() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading configuration, %s", err)
	}
}
