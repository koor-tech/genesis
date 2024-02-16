package main

import (
	"github.com/koor-tech/genesis/gateway"
	"github.com/spf13/viper"
)

func init() {
	viper.AutomaticEnv()
}

func main() {
	r := gateway.SetupRouter()
	r.Run() // listen and serve on 0.0.0.0:8080
}
