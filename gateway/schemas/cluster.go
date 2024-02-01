package schemas

import "github.com/google/uuid"

type ClusterCreated struct {
	ID uuid.UUID `json:"id"`
}
