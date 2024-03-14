package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/koor-tech/genesis/pkg/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/fx"
)

type Client struct {
	logger *slog.Logger

	conn *amqp.Connection
	ch   *amqp.Channel
}

type Params struct {
	fx.In

	LC fx.Lifecycle

	Logger *slog.Logger
	Config *config.Config
}

func NewClient(p Params) (*Client, error) {
	c := &Client{
		logger: p.Logger,
	}

	rabbitMqURI := builRabbitMqURI(p.Config.RabbitMQ)
	p.LC.Append(fx.StartHook(func(ctx context.Context) error {
		conn, err := amqp.Dial(rabbitMqURI)
		if err != nil {
			return fmt.Errorf("unable to connect with rabbitmq (url: %s). %w", rabbitMqURI, err)
		}
		c.conn = conn

		ch, err := conn.Channel()
		if err != nil {
			return fmt.Errorf("unable to connect with rabbitmq channel. %w", err)
		}
		c.ch = ch

		return nil
	}))

	p.LC.Append(fx.StopHook(func(ctx context.Context) error {
		if c.conn != nil {
			// "Graceful" close of rabbitmq connection
			return c.conn.CloseDeadline(time.Now().Add(5 * time.Second))
		}

		return nil
	}))

	return c, nil
}

func builRabbitMqURI(cfg config.RabbitMQ) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/",
		cfg.User, cfg.Password, cfg.Host, cfg.Port)
}

var (
	queueName = "clusters"
)

func (c *Client) QueueDeclare() (amqp.Queue, error) {
	queue, err := c.ch.QueueDeclare(queueName, true, true, false, false, nil)
	if err != nil {
		return queue, fmt.Errorf("exchange.declare: %w", err)
	}
	return queue, nil
}

func (c *Client) Publish(ctx context.Context, body []byte) error {
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		ContentType:  "application/json",
		Body:         body,
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// This is not a mandatory delivery, so it will be dropped if there are no
	// queues bound to the logs exchange.
	err := c.ch.PublishWithContext(ctx, "", "clusters", false, false, msg)
	if err != nil {
		// Since publish is asynchronous this can happen if the network connection
		// is reset or if the server has run out of resources.
		return fmt.Errorf("basic.publish. %w", err)
	}
	c.logger.Debug("sent message", "body", body)

	return nil
}

func (c *Client) Consume() (<-chan amqp.Delivery, error) {
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
		return nil, fmt.Errorf("failed to register a consumer. %w", err)
	}

	return msgs, nil
}
