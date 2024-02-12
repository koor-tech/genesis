package models

import "github.com/google/uuid"

type Client struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}
