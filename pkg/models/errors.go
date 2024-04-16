package models

import "errors"

var (
	ErrClusterNotFound      = errors.New("cluster not found")
	ErrCustomerNotFound     = errors.New("customer not found")
	ErrClusterStateNotFound = errors.New("cluster state not found")
)
