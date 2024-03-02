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
	rabbitMQClient *rabbitmq.Client
	db             *database.DB
}

func NewWorker() *Worker {
	queue := rabbitmq.NewClient()
	_, err := queue.QueueDeclare()
	if err != nil {
		log.Fatalf("unable to declare the queue")
	}
	return &Worker{
		rabbitMQClient: queue,
		db:             database.NewDB(),
	}
}

func (w *Worker) ResumeCluster() {
	msgs := w.rabbitMQClient.Consume()
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

	k := cluster.NewService(w.db, w.rabbitMQClient)
	err = k.ResumeCluster(context.Background(), state.ClusterID)
	if err != nil {
		log.Printf("Error running ResumeCluster: %s", err)
		return
	}
}
