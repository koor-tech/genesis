package providers

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

type ProviderRepository struct {
	db *database.DB
}

func NewProviderRepository(db *database.DB) *ProviderRepository {
	return &ProviderRepository{
		db: db,
	}
}

func (r *ProviderRepository) Save(ctx context.Context, provider models.Provider) (*models.Provider, error) {
	sqlStmt, args, _ := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert(`providers`).
		Columns(`id`, `name`).
		Values(provider.ID, provider.Name).
		ToSql()

	_, err := r.db.Conn.ExecContext(ctx, sqlStmt, args...)
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ProviderRepository) QueryByID(ctx context.Context, ID uuid.UUID) (*models.Provider, error) {
	var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(
			`p.id`,
			`p.name`,
		).
		From(`providers p`)
	var p models.Provider

	sqlStmt, args, _ := builder.Where(`p.id = $1`, ID).ToSql()
	if err := r.db.Conn.GetContext(ctx, &p, sqlStmt, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, genesis.ErrClusterNotFound
		}
		return nil, err
	}
	return &p, nil
}
