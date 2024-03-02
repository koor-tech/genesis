package main

import (
	"fmt"
	"github.com/koor-tech/genesis/gateway"
	"github.com/spf13/viper"
	"log"
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
	if err := r.Run(port); err != nil {
		log.Fatalf("Could not setup server on port %s: %v", port, err)
	}
}
