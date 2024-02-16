package rabbitmq

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewClient() *Client {
	config := NewConfig()
	conn, err := amqp.Dial(config.Url())
	//defer conn.Close()
	if err != nil {
		log.Fatal("unable to connect with rabbit", "err", err, "url", config.Url())
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("unable to connect with rabbit", "err", err)
	}

	//defer ch.Close()
	return &Client{
		conn: conn,
		ch:   ch,
	}
}

var (
	exchangeName = "cluster-tasks"
	queueName    = "clusters"
)

func (c *Client) QueueDeclare() (amqp.Queue, error) {
	queue, err := c.ch.QueueDeclare(queueName, true, true, false, false, nil)
	if err != nil {
		log.Fatalf("exchange.declare: %v", err)
	}
	return queue, nil
}

func (c *Client) Publish(ctx context.Context, body []byte) {

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         body,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This is not a mandatory delivery, so it will be dropped if there are no
	// queues bound to the logs exchange.
	err := c.ch.PublishWithContext(ctx, "", "clusters", false, false, msg)
	if err != nil {
		// Since publish is asynchronous this can happen if the network connection
		// is reset or if the server has run out of resources.
		log.Fatalf("basic.publish: %v", err)
	}
	log.Printf("Sent %s", body)
}
func (c *Client) Consume() <-chan amqp.Delivery {
	msgs, err := c.ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}
	return msgs
}
