package worker

import (
	"context"
	"encoding/json"
	"github.com/koor-tech/genesis/internal/cluster"
	"github.com/koor-tech/genesis/pkg/database"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/koor-tech/genesis/pkg/rabbitmq"
	"log"
)

type Worker struct {
	queueClient *rabbitmq.Client
	db          *database.DB
}

func NewWorker() *Worker {
	queue := rabbitmq.NewClient()
	queue.QueueDeclare()
	return &Worker{
		queueClient: queue,
		db:          database.NewDB(),
	}
}

func (w *Worker) ResumeCluster() {
	msgs := w.queueClient.Consume()
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			w.processMessage(d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func (w *Worker) processMessage(body []byte) {
	var state models.ClusterState
	err := json.Unmarshal(body, &state)
	if err != nil {
		log.Printf("Error parsing JSON: %s", err)
		return
	}

	k := cluster.NewKoorCluster(w.db)
	err = k.ResumeCluster(context.Background(), state)
	if err != nil {
		log.Printf("Error running ResumeCluster: %s", err)
		return
	}
}
