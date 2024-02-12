package clients

import (
	"context"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/models"
)

type ClientsInterface interface {
	Save(ctx context.Context, client models.Client) (*models.Client, error)
	QueryByID(ctx context.Context, ID uuid.UUID) (*models.Client, error)
}
