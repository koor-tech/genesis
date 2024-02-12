package state

import (
	"context"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/models"
)

type ClusterStateInterface interface {
	Save(ctx context.Context, state models.ClusterState) (*models.ClusterState, error)
	QueryByID(ctx context.Context, clusterID uuid.UUID) (*models.ClusterState, error)
	Update(ctx context.Context, params models.ClusterState) error
}
