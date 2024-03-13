package notification

import "github.com/koor-tech/genesis/pkg/models"

type Notifier interface {
	Send(customer models.Customer) error
}
