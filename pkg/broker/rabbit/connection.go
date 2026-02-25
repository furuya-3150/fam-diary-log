package rabbit

import (
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Config holds RabbitMQ connection configuration
type Config struct {
	URL string
}

// NewConnection creates a new RabbitMQ connection
func NewConnection(config Config) (*amqp.Connection, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("RabbitMQ URL is required")
	}

	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	closeCh := make(chan *amqp.Error)
	conn.NotifyClose(closeCh)

	go func() {
		err := <-closeCh
		if err != nil {
			slog.Warn("rabbitmq connection closed", "err", err)
		} else {
			slog.Warn("rabbitmq connection closed")
		}
	}()

	return conn, nil
}

// DeclareExchange declares an exchange
func DeclareExchange(ch *amqp.Channel, name, kind string) error {
	return ch.ExchangeDeclare(
		name,  // name
		kind,  // kind
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}

// DeclareQueue declares a queue
func DeclareQueue(ch *amqp.Channel, name string, args amqp.Table) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
}

// BindQueue binds a queue to an exchange
func BindQueue(ch *amqp.Channel, queueName, exchangeName, routingKey string) error {
	return ch.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange name
		false,        // no-wait
		nil,          // arguments
	)
}
