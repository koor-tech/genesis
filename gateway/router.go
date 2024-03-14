package gateway

import (
	"context"
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

	ClusterHandler *handler.Cluster // Use annotations for easily adding new handlers on the go instead of "just one"
}

func New(p Params) (*gin.Engine, error) {
	// Gin HTTP Server
	gin.SetMode(p.Config.Mode)
	r := gin.New()

	// Add the sloggin middleware to all routes.
	// The middleware will log all requests attributes.
	r.Use(sloggin.New(p.Logger))
	r.Use(gin.Recovery())

	RegisterRoutes(r, p.ClusterHandler)

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
			go srv.Serve(ln)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return r, nil
}

func RegisterRoutes(r *gin.Engine, clusterHandler *handler.Cluster) *gin.Engine {
	r.GET("/status", handler.GetStatus)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/cluster", clusterHandler.CreateCluster)
		v1.GET("/clusters/:id", clusterHandler.GetCluster)
		v1.DELETE("/cluster/:id", clusterHandler.DeleteCluster)
		v1.PUT("/cluster/:id", clusterHandler.ResumeCluster)
	}

	return r
}
