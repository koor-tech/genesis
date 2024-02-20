package main

import (
	"fmt"
	"github.com/koor-tech/genesis/internal/worker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	viper.AutomaticEnv()
	rootCmd.AddCommand(workerCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
