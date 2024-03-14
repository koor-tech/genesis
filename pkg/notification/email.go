package notification

import (
	"html/template"
	"log/slog"

	"github.com/koor-tech/genesis/pkg/config"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/resend/resend-go/v2"
)

const emailText = `Your managed Koor Ceph Cluster is ready to use!`

type Email struct {
	Notifier

	logger *slog.Logger
	cfg    config.EmailNotifications

	tpl *template.Template

	client *resend.Client
}

func NewEmail(logger *slog.Logger, cfg config.EmailNotifications) (Notifier, error) {
	client := resend.NewClient(cfg.Token)

	tpl, err := template.New("email").Parse(emailText)
	if err != nil {
		return nil, err
	}

	return &Email{
		logger: logger,
		cfg:    cfg,

		tpl: tpl,

		client: client,
	}, nil
}

func (n *Email) Send(customer models.Customer) error {
	params := &resend.SendEmailRequest{
		To:      []string{customer.Email},
		From:    n.cfg.From,
		Subject: n.cfg.Subject,
		ReplyTo: n.cfg.ReplyTo,
		Text:    emailText,
	}

	sent, err := n.client.Emails.Send(params)
	if err != nil {
		return err
	}

	n.logger.Debug("customer notification email sent", "email_id", sent.Id)

	return nil
}
