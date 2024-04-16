package handler

import (
	"github.com/koor-tech/genesis/gateway/schemas"
	"github.com/koor-tech/genesis/pkg/models"
)

func mapCluster(koorCluster *models.Cluster) *schemas.Cluster {
	schema := schemas.Cluster{
		ID:    koorCluster.ID,
		Phase: int(koorCluster.ClusterState.Phase),
		Customer: schemas.Customer{
			ID:      koorCluster.Customer.ID,
			Company: koorCluster.Customer.Company,
			Email:   koorCluster.Customer.Email,
		},
		Provider: schemas.Provider{
			ID:   koorCluster.Provider.ID,
			Name: koorCluster.Provider.Name,
		},
	}

	if koorCluster.KubeConfig != nil {
		schema.KubeConfig = koorCluster.KubeConfig
	}
	return &schema
}
