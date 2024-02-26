package ssh

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/files"
	"github.com/koor-tech/genesis/pkg/models"
	sshRepo "github.com/koor-tech/genesis/pkg/repositories/postgres/ssh"
	sshPkg "github.com/koor-tech/genesis/pkg/ssh"
	"log/slog"
	"os"
)

type Service struct {
	sshManager    *sshPkg.Manager
	sshRepository sshRepo.SshInterface
	logger        *slog.Logger
}

func NewService(db *database.DB) *Service {
	return &Service{
		sshManager:    sshPkg.NewManager(),
		sshRepository: sshRepo.NewSshRepository(db),
		logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
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
	err = files.SaveInFile(privateKeyFile, sshModel.PrivateKey, filePermission)
	if err != nil {
		s.logger.Error("unable to generate ssh keys", "err", err)
		return nil, err
	}
	err = files.SaveInFile(publicKeyFile, sshModel.PublicKey, filePermission)
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
