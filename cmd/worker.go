package cmd

import (
	"github.com/koor-tech/genesis/internal/worker"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var workerCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitors the state of the each cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		fxOpts := getFxBaseOpts()

		fxOpts = append(fxOpts, fx.Invoke(func(*worker.Worker) {}))

		app := fx.New(fxOpts...)
		app.Run()

		return nil
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)
}
