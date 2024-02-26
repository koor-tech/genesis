package main

import (
	"fmt"
	"github.com/koor-tech/genesis/gateway"
	"github.com/spf13/viper"
)

const (
	GenesisApiDefaultPort = "8080"
)

func init() {
	viper.AutomaticEnv()
}

func main() {
	portNumber := viper.GetString("GENESIS_PORT")
	if len(portNumber) == 0 {
		portNumber = GenesisApiDefaultPort
	}

	port := fmt.Sprintf(":%s", portNumber)
	r := gateway.SetupRouter()
	r.Run(port) // listen and serve on 0.0.0.0:8080
}
