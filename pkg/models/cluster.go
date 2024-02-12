package models

import "github.com/google/uuid"

type Cluster struct {
	ID           uuid.UUID    `db:"id"`
	ClientID     uuid.UUID    `db:"client_id"`
	ProviderID   uuid.UUID    `db:"provider_id"`
	Client       Client       `db:"clients"`
	Provider     Provider     `db:"providers"`
	ClusterState ClusterState `db:"cs"`
}
