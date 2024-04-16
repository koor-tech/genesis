package models

import "github.com/google/uuid"

type Customer struct {
	ID      uuid.UUID `db:"id"`
	Company string    `db:"company"`
	Email   string    `db:"email"`
}
