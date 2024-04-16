package gateway

import (
	"context"
	"github.com/koor-tech/genesis/gateway/middleware"
	"log/slog"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/koor-tech/genesis/gateway/handler"
	"github.com/koor-tech/genesis/pkg/config"
	sloggin "github.com/samber/slog-gin"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	LC fx.Lifecycle

	Logger *slog.Logger
	Config *config.Config

	Shutdowner     fx.Shutdowner
	ClusterHandler *handler.Cluster // Use annotations for easily adding new handlers on the go instead of "just one"
	AuthMiddleware *middleware.AuthMiddleware
}

func New(p Params) (*gin.Engine, error) {
	// Gin HTTP Server
	gin.SetMode(p.Config.Mode)
	r := gin.New()

	// Add the sloggin middleware to all routes.
	// The middleware will log all requests attributes.
	r.Use(sloggin.New(p.Logger))
	r.Use(gin.Recovery())

	RegisterRoutes(r, p.ClusterHandler, p.AuthMiddleware)

	// Create HTTP Server for graceful shutdown handling
	srv := &http.Server{
		Addr:    p.Config.Listen,
		Handler: r,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			p.Logger.Info("http server listening", "address", srv.Addr)
			go func() {
				if err := srv.Serve(ln); err != nil {
					p.Logger.Error("unable to start the server", "err", err)
					if shutdownErr := p.Shutdowner.Shutdown(); shutdownErr != nil {
						p.Logger.Error("failed to shutdown the application", "err", shutdownErr)
					}
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return r, nil
}

func RegisterRoutes(r *gin.Engine, clusterHandler *handler.Cluster, authMiddleware *middleware.AuthMiddleware) *gin.Engine {
	r.GET("/status", handler.GetStatus)
	v1 := r.Group("/api/v1").Use(authMiddleware.Validate)
	{
		v1.POST("/clusters", clusterHandler.CreateCluster)
		v1.GET("/clusters/:id", clusterHandler.GetCluster)
		v1.DELETE("/clusters/:id", clusterHandler.DeleteCluster)
		v1.PUT("/clusters/:id", clusterHandler.ResumeCluster)
	}

	return r
}
