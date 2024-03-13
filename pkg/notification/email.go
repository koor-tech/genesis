package notification

import (
	"log/slog"

	"github.com/koor-tech/genesis/pkg/models"
	"github.com/resend/resend-go/v2"
)

type Email struct {
	Notifier

	logger *slog.Logger

	client *resend.Client
}

func NewEmail(logger *slog.Logger) *Email {
	client := resend.NewClient("")

	return &Email{
		logger: logger,

		client: client,
	}
}

func (n *Email) Send(customer models.Customer) error {
	params := &resend.SendEmailRequest{
		To:      []string{customer.Email},
		From:    "me@exemple.io", // TODO make configurable
		Text:    "hello world",
		Subject: "Hello from Golang",
		ReplyTo: "replyto@example.com",
	}

	sent, err := n.client.Emails.Send(params)
	if err != nil {
		return err
	}

	n.logger.Debug("customer notification email sent", "email_id", sent.Id)

	return nil
}
