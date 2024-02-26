package models

import "github.com/google/uuid"

type Customer struct {
	ID    uuid.UUID `db:"id"`
	Name  string    `db:"name"`
	Email string    `db:"email"`
}
