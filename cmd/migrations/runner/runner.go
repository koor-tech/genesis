package runner

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"log"
)

const (
	migrationsDir = "/app/migrations"
)

func RunMigration(command string, db *sql.DB, args ...string) {
	if err := goose.RunContext(context.Background(), command, db, migrationsDir, args...); err != nil {
		log.Fatalf("goose %v: %v dir: %s", command, err, migrationsDir)
	}
}
