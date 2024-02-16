package clusters

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

type ClusterRepository struct {
	db *database.DB
}

func NewClusterRepository(db *database.DB) *ClusterRepository {
	return &ClusterRepository{
		db: db,
	}
}

func (r *ClusterRepository) Save(ctx context.Context, cluster models.Cluster) (*models.Cluster, error) {
	sqlStmt, args, _ := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert(`clusters`).
		Columns(`id`, `customer_id`, `provider_id`).
		Values(cluster.ID, cluster.Customer.ID, cluster.Provider.ID).
		ToSql()

	_, err := r.db.Conn.ExecContext(ctx, sqlStmt, args...)
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterRepository) QueryByID(ctx context.Context, clusterID uuid.UUID) (*models.Cluster, error) {
	var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(
			`c.id`,
			`c.customer_id`,
			`customers.id as "customers.id"`,
			`customers.name as "customers.name"`,
			`c.provider_id`,
			`p.id as "providers.id"`,
			`p.name as "providers.name"`,
			`cs.phase as "cs.phase"`,
			`cs.id as "cs.id"`,
			`cs.cluster_id as "cs.cluster_id"`,
		).
		From(`clusters c`).
		InnerJoin("customers on customers.id = c.customer_id").
		InnerJoin("providers p on p.id = c.provider_id").
		InnerJoin("cluster_state cs on cs.cluster_id = c.id")

	var c models.Cluster

	sqlStmt, args, _ := builder.Where(`c.id = $1`, clusterID).ToSql()
	if err := r.db.Conn.GetContext(ctx, &c, sqlStmt, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, genesis.ErrClusterNotFound
		}
		return nil, err
	}
	return &c, nil
}
