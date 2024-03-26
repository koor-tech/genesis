package schemas

import "github.com/google/uuid"

type Customer struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}

type Provider struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Cluster struct {
	ID         uuid.UUID `json:"id"`
	Phase      int       `json:"status"`
	Customer   Customer  `json:"customer"`
	Provider   Provider  `json:"provider"`
	KubeConfig *string   `json:"kube-config,omitempty"`
}
