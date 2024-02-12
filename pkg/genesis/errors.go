package genesis

import "errors"

var (
	ErrClusterNotFound      = errors.New("cluster not found")
	ErrClusterStateNotFound = errors.New("cluster state not found")
)
