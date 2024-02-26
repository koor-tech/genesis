package customers

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

type CustomersRepository struct {
	db *database.DB
}

func NewCustomersRepository(db *database.DB) *CustomersRepository {
	return &CustomersRepository{
		db: db,
	}
}

func (r *CustomersRepository) Save(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	sqlStmt, args, _ := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert(`customers`).
		Columns(`id`, `name`, `email`).
		Values(customer.ID, customer.Name, customer.Email).
		ToSql()

	_, err := r.db.Conn.ExecContext(ctx, sqlStmt, args...)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func (r *CustomersRepository) QueryByID(ctx context.Context, ID uuid.UUID) (*models.Customer, error) {
	var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(
			`c.id`,
			`cs.name`,
		).
		From(`customers c`)
	var c models.Customer

	sqlStmt, args, _ := builder.Where(`c.id = $1`, ID).ToSql()
	if err := r.db.Conn.GetContext(ctx, &c, sqlStmt, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, genesis.ErrClusterNotFound
		}
		return nil, err
	}
	return &c, nil
}
