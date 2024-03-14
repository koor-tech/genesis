package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/koor-tech/genesis/pkg/config"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

const (
	DriverNamePostgres = "postgres"
	maxRetriesConn     = 10
)

// DB struct for manage database connection,
type DB struct {
	logger *slog.Logger

	Conn *sqlx.DB
}

type Params struct {
	fx.In

	LC fx.Lifecycle

	Logger *slog.Logger
	Config *config.Config
}

func NewDB(p Params) (*DB, error) {
	d := &DB{
		logger: p.Logger,
	}

	p.LC.Append(fx.StartHook(func(ctx context.Context) error {
		if err := d.Connect(BuilDSNUri(p.Config.Database)); err != nil {
			return err
		}

		return nil
	}))

	p.LC.Append(fx.StopHook(func(ctx context.Context) error {
		if d.Conn != nil {
			return d.Conn.Close()
		}

		return nil
	}))

	return d, nil
}

func (d *DB) Connect(dbUri string) error {
	var (
		tries int
		conn  *sqlx.DB
		err   error
	)

	for {
		conn, err = sqlx.Connect(DriverNamePostgres, dbUri)
		if err != nil {
			tries++
			d.logger.Error("db connection attempt failed", "err", err, "try", tries)

			if tries > maxRetriesConn {
				return fmt.Errorf("failed to connect to db. %w", err)
			}

			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	d.Conn = conn

	return nil
}

func BuilDSNUri(cfg config.Database) string {
	ssl := "sslmode=disable"
	if cfg.SSLEnabled {
		ssl = ""
	}

	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, ssl)
}
