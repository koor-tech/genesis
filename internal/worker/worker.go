package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/koor-tech/genesis/internal/cluster"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/koor-tech/genesis/pkg/notification"
	"github.com/koor-tech/genesis/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/fx"
)

type Worker struct {
	logger   *slog.Logger
	db       *database.DB
	notifier notification.Notifier

	clusterSvc *cluster.Service
}

type Params struct {
	fx.In

	LC fx.Lifecycle

	Logger     *slog.Logger
	RabbitMQ   *rabbitmq.Client
	DB         *database.DB
	Notifier   notification.Notifier
	ClusterSvc *cluster.Service
}

func NewWorker(p Params) (*Worker, error) {
	w := &Worker{
		logger:     p.Logger,
		db:         p.DB,
		notifier:   p.Notifier,
		clusterSvc: p.ClusterSvc,
	}

	p.LC.Append(fx.StartHook(func(ctx context.Context) error {
		_, err := p.RabbitMQ.QueueDeclare()
		if err != nil {
			return fmt.Errorf("unable to declare the queue. %w", err)
		}

		msgs, err := p.RabbitMQ.Consume()
		if err != nil {
			return err
		}

		go w.ResumeCluster(msgs)

		return nil
	}))

	return w, nil
}

func (w *Worker) ResumeCluster(ch <-chan amqp.Delivery) error {
	forever := make(chan bool)

	go func() {
		for d := range ch {
			w.logger.Info("received a message", "body", d.Body)
			w.processMessage(d.Body)
		}
	}()

	w.logger.Info("[*] Waiting for messages. To exit press CTRL+C")
	<-forever

	return nil
}

func (w *Worker) processMessage(body []byte) {
	var state models.ClusterState
	err := json.Unmarshal(body, &state)
	if err != nil {
		w.logger.Error("error parsing JSON", "err", err)
		return
	}

	err = w.clusterSvc.ResumeCluster(context.Background(), state.ClusterID)
	if err != nil {
		w.logger.Error("error running ResumeCluster", "err", err)
		return
	}
}
