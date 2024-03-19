package clusters

import (
	"context"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/models"
)

type ClustersInterface interface {
	Save(ctx context.Context, cluster models.Cluster) (*models.Cluster, error)
	QueryByID(ctx context.Context, clusterID uuid.UUID) (*models.Cluster, error)
	Update(ctx context.Context, params models.Cluster) error
}
