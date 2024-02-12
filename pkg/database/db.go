package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"time"
)

const (
	DriverNamePostgres = "postgres"
	maxRetriesConn     = 20
)

// DB struct for manage database connection,
type DB struct {
	Conn *sqlx.DB
}

func NewDB() *DB {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &DB{
		Conn: connect(logger, viper.GetString("database.uri")),
	}
}

func connect(logger *slog.Logger, dbUri string) *sqlx.DB {
	var (
		tries int
		conn  *sqlx.DB
		err   error
	)

	for {
		conn, err = sqlx.Connect(DriverNamePostgres, dbUri)
		if err != nil {
			logger.Info("connection attempt", "err", err.Error(), "try", tries)
			tries++
			if tries > maxRetriesConn {
				logger.Warn("failed to connect to db", "err", err)
				os.Exit(1)
			}
			time.Sleep(time.Duration(5) * time.Second)
		} else {
			break
		}
	}
	return conn
}
