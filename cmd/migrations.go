package cmd

import (
	"log/slog"
	"os"

	"github.com/koor-tech/genesis/cmd/migrations"
	"github.com/koor-tech/genesis/pkg/config"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/spf13/cobra"
	"go.uber.org/fx/fxtest"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "run migration commands using goose",
	RunE: func(cmd *cobra.Command, args []string) error {
		var arguments []string
		if len(args) > 1 {
			arguments = append(arguments, args[1:]...)
		}
		var command string
		if len(args) == 0 {
			command = "status"

		} else {
			command = args[0]
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Setup database connection for goose
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
		d, err := database.NewDB(database.Params{
			LC:     fxtest.NewLifecycle(nil),
			Logger: logger,
			Config: cfg.Config,
		})
		if err != nil {
			return err
		}

		if err := d.Connect(database.BuilDSNUri(cfg.Config.Database)); err != nil {
			return err
		}
		defer d.Conn.Close()

		return migrations.RunMigrationCommand(command, d.Conn.DB, arguments...)
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)
}
