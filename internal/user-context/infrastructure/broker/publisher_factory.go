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
	// singleton connection for diary-mailer
	mailRabbitConn *amqp.Connection
	mailConnOnce   sync.Once
	mailConnErr    error
)

func initMailConnection() (*amqp.Connection, error) {
	mailConnOnce.Do(func() {
		rabbitmqURL := os.Getenv("RABBITMQ_URL")
		if rabbitmqURL == "" {
			mailConnErr = errors.New("RABBITMQ_URL environment variable is not set")
			return
		}

		rabbitConfig := rabbit.Config{URL: rabbitmqURL}
		mailRabbitConn, mailConnErr = rabbit.NewConnection(rabbitConfig)
	})

	return mailRabbitConn, mailConnErr
}

// GetMailRabbitConnection returns the singleton connection for mail publisher
func GetMailRabbitConnection() (*amqp.Connection, error) {
	if mailRabbitConn != nil {
		return mailRabbitConn, nil
	}
	return initMailConnection()
}

// NewDiaryMailerPublisher initializes and returns a RabbitMQ publisher for mail commands
func NewDiaryMailerPublisher(log *slog.Logger) publisher.Publisher {
	conn, err := GetMailRabbitConnection()
	if err != nil {
		log.Error("failed to get RabbitMQ connection for mail publisher", "error", err.Error())
		os.Exit(1)
	}

	cfg := DiaryPublisherConfig()
	pub, err := publisher.NewRabbitMQPublisher(conn, cfg, log)
	if err != nil {
		log.Error("failed to create mail publisher", "error", err.Error())
		os.Exit(1)
	}

	return pub
}
