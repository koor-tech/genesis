package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func RunMigrationCommand(command string, db *sql.DB, args ...string) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.RunContext(context.Background(), command, db, ".", args...); err != nil {
		return fmt.Errorf("goose %v: dir %s. %w", command, ".", err)
	}

	return nil
}
