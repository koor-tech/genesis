package models

import "github.com/google/uuid"

type Ssh struct {
	ID              uuid.UUID `db:"id"`
	ClusterID       uuid.UUID `db:"cluster_id"`
	PrivateFilePath string    `db:"private_file_path"`
	PublicFilePath  string    `db:"public_file_path"`
	PrivateKey      string    `db:"private_key"`
	PublicKey       string    `db:"public_key"`
}
