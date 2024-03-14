package ssh

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/models"
	sshRepo "github.com/koor-tech/genesis/pkg/repositories/postgres/ssh"
	sshPkg "github.com/koor-tech/genesis/pkg/ssh"
	"github.com/koor-tech/genesis/pkg/utils"
)

type Service struct {
	logger *slog.Logger

	sshManager    *sshPkg.Manager
	sshRepository sshRepo.SshInterface
}

func NewService(logger *slog.Logger, db *database.DB) *Service {
	return &Service{
		logger: logger,

		sshManager:    sshPkg.NewManager(logger),
		sshRepository: sshRepo.NewSshRepository(db),
	}
}

func (s *Service) BuildAndRunSSH(ctx context.Context, clusterID uuid.UUID, dirPath string) (*models.Ssh, error) {
	const (
		filePermission = 0600
		fileKeyName    = "id_ed25519"
	)
	//Generate SSH key
	sshModel, err := s.sshManager.GenerateKey()
	if err != nil {
		s.logger.Error("unable to generate ssh keys", "err", err)
		return nil, err
	}

	privateKeyFile := fmt.Sprintf("%s/%s", dirPath, fileKeyName)
	publicKeyFile := fmt.Sprintf("%s/%s.pub", dirPath, fileKeyName)

	sshModel.PrivateFilePath = privateKeyFile
	sshModel.PublicFilePath = publicKeyFile
	sshModel.ClusterID = clusterID

	//Save ssh into clients folder
	err = utils.SaveInFile(privateKeyFile, sshModel.PrivateKey, filePermission)
	if err != nil {
		s.logger.Error("unable to generate ssh keys", "err", err)
		return nil, err
	}
	err = utils.SaveInFile(publicKeyFile, sshModel.PublicKey, filePermission)
	if err != nil {
		s.logger.Error("unable to generate ssh keys", "err", err)
		return nil, err
	}

	s.logger.Info("============ Running SSH agent  ==============")
	err = s.sshManager.RunAgent(sshModel.PrivateFilePath)
	if err != nil {
		s.logger.Error("unable to execute ssh agent", "err", err)
		return nil, err
	}

	_, err = s.sshRepository.Save(ctx, *sshModel)
	if err != nil {
		s.logger.Error("unable to save ssh keys in database", "err", err)
		return nil, err
	}
	return sshModel, nil
}
