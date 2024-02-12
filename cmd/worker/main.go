package main

import (
	"fmt"
	"github.com/koor-tech/genesis/internal/worker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "clusters",
	Short: "Monitors the clusters",
}

var workerCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitors the state of the each cluster",
	Run: func(cmd *cobra.Command, args []string) {
		w := worker.NewWorker()
		w.ResumeCluster()
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
}

func main() {
	loadConfiguration()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func loadConfiguration() {
	viper.SetConfigName("config")
	viper.AddConfigPath("../../")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading configuration, %s", err)
	}
}
