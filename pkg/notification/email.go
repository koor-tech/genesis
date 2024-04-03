package notification

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/koor-tech/genesis/pkg/config"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/resend/resend-go/v2"
)

const (
	emailSubject = `Your managed Koor Ceph Cluster is ready to use!`
)

type Email struct {
	Notifier

	logger *slog.Logger
	cfg    config.EmailNotifications

	tmpls *template.Template

	client *resend.Client
}

func NewEmail(logger *slog.Logger, tmpls *template.Template, cfg config.EmailNotifications) (Notifier, error) {
	client := resend.NewClient(cfg.Token)

	return &Email{
		logger: logger,
		cfg:    cfg,

		tmpls: tmpls,

		client: client,
	}, nil
}

func (n *Email) Send(customer models.Customer) error {
	var out bytes.Buffer
	if err := n.tmpls.ExecuteTemplate(&out, "notification_email.gotmpl", &TmplData{
		AccessLink: fmt.Sprintf("https://koor.tech/cluster?customer_id=%s", customer.ID), // TODO
	}); err != nil {
		return err
	}

	params := &resend.SendEmailRequest{
		To:      []string{customer.Email},
		From:    n.cfg.From,
		ReplyTo: n.cfg.ReplyTo,
		Subject: emailSubject,
		Text:    "emailText",
	}

	sent, err := n.client.Emails.Send(params)
	if err != nil {
		return err
	}

	n.logger.Debug("customer notification email sent", "email_id", sent.Id)

	return nil
}
