package clients

import (
	"context"
	"database/sql"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/genesis"
	"github.com/koor-tech/genesis/pkg/models"
)

type ClientsRepository struct {
	db *database.DB
}

func NewClientsRepository(db *database.DB) *ClientsRepository {
	return &ClientsRepository{
		db: db,
	}
}

func (r *ClientsRepository) Save(ctx context.Context, client models.Client) (*models.Client, error) {
	sqlStmt, args, _ := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert(`clients`).
		Columns(`id`, `name`).
		Values(client.ID, client.Name).
		ToSql()

	_, err := r.db.Conn.ExecContext(ctx, sqlStmt, args...)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *ClientsRepository) QueryByID(ctx context.Context, ID uuid.UUID) (*models.Client, error) {
	var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(
			`c.id`,
			`cs.name`,
		).
		From(`clients c`)
	var c models.Client

	sqlStmt, args, _ := builder.Where(`c.id = $1`, ID).ToSql()
	if err := r.db.Conn.GetContext(ctx, &c, sqlStmt, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, genesis.ErrClusterNotFound
		}
		return nil, err
	}
	return &c, nil
}
