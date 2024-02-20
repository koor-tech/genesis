package customers

import (
	"context"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/models"
)

type CustomerInterface interface {
	Save(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	QueryByID(ctx context.Context, ID uuid.UUID) (*models.Customer, error)
}
