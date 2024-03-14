package notification

import "github.com/koor-tech/genesis/pkg/models"

type Noop struct {
	Notifier
}

func NewNoop() Notifier {
	return &Noop{}
}

func (n *Noop) Send(customer models.Customer) error {
	return nil
}
