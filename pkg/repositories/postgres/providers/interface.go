package providers

import (
	"context"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/models"
)

type ProvidersInterface interface {
	Save(ctx context.Context, provider models.Provider) (*models.Provider, error)
	QueryByID(ctx context.Context, ID uuid.UUID) (*models.Provider, error)
}
