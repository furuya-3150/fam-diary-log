package rabbit

import (
	"fmt"

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
func DeclareQueue(ch *amqp.Channel, name string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
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
