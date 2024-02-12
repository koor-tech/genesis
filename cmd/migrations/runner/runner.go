package runner

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
	"log"

	_ "github.com/lib/pq"
)

func RunMigration(command string, db *sql.DB, dir string, args ...string) {
	if err := goose.RunContext(context.Background(), command, db, dir, args...); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
}
