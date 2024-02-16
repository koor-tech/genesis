package models

import "github.com/google/uuid"

type Cluster struct {
	ID           uuid.UUID    `db:"id"`
	CustomerID   uuid.UUID    `db:"customer_id"`
	ProviderID   uuid.UUID    `db:"provider_id"`
	Customer     Customer     `db:"customers"`
	Provider     Provider     `db:"providers"`
	ClusterState ClusterState `db:"cs"`
}
