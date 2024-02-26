package ssh

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/models"
)

type SshRepository struct {
	db *database.DB
}

func NewSshRepository(db *database.DB) *SshRepository {
	return &SshRepository{
		db: db,
	}
}

func (r *SshRepository) Save(ctx context.Context, sshModel models.Ssh) (*models.Ssh, error) {
	sqlStmt, args, _ := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Insert(`ssh`).
		Columns(`id`, `cluster_id`, `private_file_path`, `public_file_path`, `private_key`, `public_key`).
		Values(sshModel.ID, sshModel.ClusterID, sshModel.PrivateFilePath, sshModel.PublicFilePath, sshModel.PrivateKey, sshModel.PublicKey).
		ToSql()

	_, err := r.db.Conn.ExecContext(ctx, sqlStmt, args...)
	if err != nil {
		return nil, err
	}
	return &sshModel, nil
}
