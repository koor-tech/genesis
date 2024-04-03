package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/koor-tech/genesis/pkg/fxslog"
	"github.com/koor-tech/genesis/tmpls"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/koor-tech/genesis/gateway"
	"github.com/koor-tech/genesis/gateway/handler"
	"github.com/koor-tech/genesis/internal/cluster"
	sshSvc "github.com/koor-tech/genesis/internal/ssh"
	"github.com/koor-tech/genesis/internal/worker"
	"github.com/koor-tech/genesis/pkg/config"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/notification"
	"github.com/koor-tech/genesis/pkg/providers/hetzner"
	"github.com/koor-tech/genesis/pkg/rabbitmq"
	clusters "github.com/koor-tech/genesis/pkg/repositories/postgres/cluster"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/customers"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/providers"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/ssh"
	"github.com/koor-tech/genesis/pkg/repositories/postgres/state"
)

func getFxBaseOpts() []fx.Option {
	return []fx.Option{
		fx.StartTimeout(90 * time.Second),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return fxslog.New(logger)
		}),

		config.Module,
		LoggerModule,
		notification.Module,
		gateway.HTTPServerModule,

		fx.Provide(
			state.NewClusterStateRepository,
			customers.NewCustomersRepository,
			providers.NewProviderRepository,
			clusters.NewClusterRepository,
			ssh.NewSshRepository,

			sshSvc.NewService,

			tmpls.New,

			hetzner.New,
			worker.NewWorker,
		),

		fx.Provide(
			database.NewDB,
			rabbitmq.NewClient,
			handler.NewCluster,
			cluster.NewService,

			// Add any dependencies here so they can just be injected
		),
	}
}

var RootCmd = &cobra.Command{
	Use: "genesis",
	RunE: func(cmd *cobra.Command, args []string) error {
		fxOpts := getFxBaseOpts()

		fxOpts = append(fxOpts, fx.Invoke(func(*gin.Engine) {}))

		app := fx.New(fxOpts...)
		app.Run()

		return nil
	},
}

var LoggerModule = fx.Module("logger",
	fx.Provide(
		NewLogger,
	),
)

func NewLogger(cfg *config.Config) (*slog.Logger, error) {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil)), nil
}
