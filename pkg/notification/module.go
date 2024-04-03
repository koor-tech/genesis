package notification

import (
	"html/template"
	"log/slog"

	"github.com/koor-tech/genesis/pkg/config"
	"go.uber.org/fx"
)

var Module = fx.Module("notification",
	fx.Provide(New),
)

type Params struct {
	fx.In

	Logger *slog.Logger
	Tmpls  *template.Template
	Config *config.Config
}

func New(p Params) (Notifier, error) {
	var not Notifier
	var err error

	switch p.Config.Notifications.Type {
	case config.NotificationTypeEmail:
		not, err = NewEmail(p.Logger, p.Tmpls, p.Config.Notifications.Email)

	default:
		not = NewNoop()
	}

	return not, err
}
