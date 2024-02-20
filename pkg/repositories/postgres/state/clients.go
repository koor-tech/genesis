package state

import (
	"context"
	"database/sql"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/genesis"
	"github.com/koor-tech/genesis/pkg/models"
	"time"
)

type ClusterStateRepository struct {
	db *database.DB
}

func NewClusterStateRepository(db *database.DB) *ClusterStateRepository {
	return &ClusterStateRepository{
		db: db,
	}
}

func (r *ClusterStateRepository) Save(ctx context.Context, clusterState models.ClusterState) (*models.ClusterState, error) {
	sqlStmt, args, _ := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert(`cluster_state`).
		Columns(`id`, `cluster_id`, `phase`, `time`).
		Values(clusterState.ID, clusterState.ClusterID, clusterState.Phase, time.Now()).
		ToSql()

	_, err := r.db.Conn.ExecContext(ctx, sqlStmt, args...)
	if err != nil {
		return nil, err
	}
	return &clusterState, nil
}

func (r *ClusterStateRepository) QueryByID(ctx context.Context, clusterID uuid.UUID) (*models.ClusterState, error) {
	var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select(
			`cs.id`,
			`cs.cluster_id`,
			`cs.phase`,
		).
		From(`cluster_state cs`)
	var cs models.ClusterState

	sqlStmt, args, _ := builder.Where(`cs.id = $1`, clusterID).ToSql()
	if err := r.db.Conn.GetContext(ctx, &cs, sqlStmt, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, genesis.ErrClusterNotFound
		}
		return nil, err
	}
	return &cs, nil
}

func (r *ClusterStateRepository) Update(ctx context.Context, params models.ClusterState) error {
	cs, err := r.QueryByID(ctx, params.ID)
	if err != nil {
		return err
	}
	cs.Phase = params.Phase
	updateStmt, args, _ := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update(`cluster_state`).
		Set(`phase`, cs.Phase).
		Where("id = $2", cs.ID).
		ToSql()

	_, err = r.db.Conn.ExecContext(ctx, updateStmt, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return genesis.ErrClusterStateNotFound
		}
		return err
	}

	return nil
}
