package ssh

import (
	"context"
	"github.com/koor-tech/genesis/pkg/models"
)

type SshInterface interface {
	Save(ctx context.Context, sshModel models.Ssh) (*models.Ssh, error)
}
