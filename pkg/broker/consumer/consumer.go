package consumer

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/furuya-3150/fam-diary-log/pkg/broker/rabbit"
)

const (
	DefaultExchangeKind = "topic"
)

// EventHandler defines the interface for handling events
// Implementation should handle specific event types
type EventHandler interface {
	Handle(ctx context.Context, eventType string, content []byte) error
}

// Consumer defines the interface for consuming events
type Consumer interface {
	Start(ctx context.Context) error
	Stop() error
}

// Config holds configuration for RabbitMQConsumer
type Config struct {
	ExchangeName string
	ExchangeKind string
	QueueName    string
	RoutingKeys  []string // multiple routing keys to bind
}

// RabbitMQConsumer implements Consumer for RabbitMQ
type RabbitMQConsumer struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	config  Config
	handler EventHandler
	l       *slog.Logger
}

// NewRabbitMQConsumer creates a new RabbitMQConsumer
func NewRabbitMQConsumer(conn *amqp.Connection, config Config, handler EventHandler, l *slog.Logger) (*RabbitMQConsumer, error) {
	if config.ExchangeName == "" {
		return nil, fmt.Errorf("exchange name is required")
	}
	if config.QueueName == "" {
		return nil, fmt.Errorf("queue name is required")
	}
	if len(config.RoutingKeys) == 0 {
		return nil, fmt.Errorf("at least one routing key is required")
	}
	if config.ExchangeKind == "" {
		config.ExchangeKind = DefaultExchangeKind
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	consumer := &RabbitMQConsumer{
		conn:    conn,
		ch:      ch,
		config:  config,
		handler: handler,
		l:       l,
	}

	return consumer, nil
}

// Start starts consuming events
func (c *RabbitMQConsumer) Start(ctx context.Context) error {
	// Declare exchange
	if err := rabbit.DeclareExchange(c.ch, c.config.ExchangeName, c.config.ExchangeKind); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	q, err := rabbit.DeclareQueue(c.ch, c.config.QueueName)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange with multiple routing keys
	for _, routingKey := range c.config.RoutingKeys {
		if err := rabbit.BindQueue(c.ch, q.Name, c.config.ExchangeName, routingKey); err != nil {
			return fmt.Errorf("failed to bind queue with routing key %s: %w", routingKey, err)
		}
	}

	// Set QoS (prefetch count)
	if err := c.ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Consume messages
	msgs, err := c.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (manual ack)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to consume: %w", err)
	}

	c.l.Info("consumer started", "queue", q.Name, "routing_keys", c.config.RoutingKeys)

	// Handle messages
	go func() {
		for msg := range msgs {
			c.handleMessage(ctx, msg)
		}
	}()

	return nil
}

// handleMessage handles a single message
func (c *RabbitMQConsumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	// Get routing key from message
	routingKey := msg.RoutingKey


	// Handle event
	if err := c.handler.Handle(ctx, routingKey, msg.Body); err != nil {
		c.l.Error("failed to handle event", "routing_key", routingKey, "error", err.Error())
		msg.Nack(false, true) // Nack and requeue
		return
	}

	// Acknowledge message
	if err := msg.Ack(false); err != nil {
		c.l.Error("failed to acknowledge message", "error", err.Error())
	}

	c.l.Info("event processed", "routing_key", routingKey)
}

// Stop stops the consumer
func (c *RabbitMQConsumer) Stop() error {
	if err := c.ch.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}
