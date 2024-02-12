package main

import (
	"fmt"
	"github.com/koor-tech/genesis/cmd/migrations/runner"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "migrations",
	Short: "Database migrations",
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "run migrations using goose",
	Run: func(cmd *cobra.Command, args []string) {
		var arguments []string
		if len(args) > 1 {
			arguments = append(arguments, args[1:]...)
		}
		command := args[0]
		db, err := goose.OpenDBWithDriver(database.DriverNamePostgres, viper.GetString("database.uri"))
		if err != nil {
			return
		}
		runner.RunMigration(command, db, "migrations", arguments...)
	},
}

func init() {
	viper.SetConfigName("../../config")
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading configuration, %s", err)
	}
	rootCmd.AddCommand(migrateCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
