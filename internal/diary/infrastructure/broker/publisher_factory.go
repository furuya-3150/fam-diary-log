package broker

import (
	"errors"
	"log/slog"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/furuya-3150/fam-diary-log/pkg/broker/publisher"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/rabbit"
)

var (
	// Singleton RabbitMQ connection
	rabbitConn *amqp.Connection
	connMutex  sync.Once
	connErr    error
)

// initConnection initializes the RabbitMQ connection once
func initConnection() (*amqp.Connection, error) {
	connMutex.Do(func() {
		rabbitmqURL := os.Getenv("RABBITMQ_URL")
		if rabbitmqURL == "" {
			connErr = errors.New("RABBITMQ_URL environment variable is not set")
			return
		}

		rabbitConfig := rabbit.Config{URL: rabbitmqURL}
		rabbitConn, connErr = rabbit.NewConnection(rabbitConfig)
	})

	return rabbitConn, connErr
}

// GetRabbitMQConnection returns the singleton RabbitMQ connection
func GetRabbitMQConnection() (*amqp.Connection, error) {
	if rabbitConn != nil {
		return rabbitConn, nil
	}
	return initConnection()
}

// NewDiaryPublisher initializes and returns a RabbitMQ publisher for diary events
func NewDiaryPublisher(log *slog.Logger) (publisher.Publisher) {
	// Get or initialize the shared connection
	conn, err := GetRabbitMQConnection()
	if err != nil {
		log.Error("failed to get RabbitMQ connection", "error", err.Error())
		os.Exit(1)
	}

	// Create publisher with shared connection
	publisherConfig := DiaryPublisherConfig()
	pub, err := publisher.NewRabbitMQPublisher(conn, publisherConfig, log)
	if err != nil {
		log.Error("failed to create publisher", "error", err.Error())
		os.Exit(1)
	}

	return pub
}

// CloseRabbitMQConnection closes the singleton connection
func CloseRabbitMQConnection() error {
	if rabbitConn != nil {
		return rabbitConn.Close()
	}
	return nil
}
