package models

import "github.com/google/uuid"

type Provider struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}
