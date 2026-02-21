package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/furuya-3150/fam-diary-log/pkg/events"
)

// Publisher defines the interface for publishing events
type Publisher interface {
	Publish(ctx context.Context, event events.Event) error
	Close() error
}

// Config holds configuration for RabbitMQPublisher
type Config struct {
	ExchangeName string
	ExchangeKind string
}

// RabbitMQPublisher implements Publisher for RabbitMQ
type RabbitMQPublisher struct {
	conn   *amqp.Connection
	config Config
	l      *slog.Logger
}

// NewRabbitMQPublisher creates a new RabbitMQPublisher
func NewRabbitMQPublisher(conn *amqp.Connection, config Config, l *slog.Logger) (*RabbitMQPublisher, error) {
	if config.ExchangeName == "" {
		return nil, fmt.Errorf("exchange name is required")
	}
	if config.ExchangeKind == "" {
		config.ExchangeKind = "topic" // default to topic
	}

	// チャネルを作成してexchangeの存在を確認（初期化時のみ）
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	// Ensure exchange exists
	if err := ch.ExchangeDeclare(
		config.ExchangeName, // name
		config.ExchangeKind, // kind
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	); err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	publisher := &RabbitMQPublisher{
		conn:   conn,
		config: config,
		l:      l,
	}

	return publisher, nil
}

// Publish publishes an event to RabbitMQ
// リクエスト毎にchannelを作成してクローズすることで、channel/connection is not openエラーを回避
func (p *RabbitMQPublisher) Publish(ctx context.Context, event events.Event) error {
	// リクエスト毎にchannelを作成
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	}

	if err := ch.PublishWithContext(
		ctx,
		p.config.ExchangeName, // exchange
		event.EventType(),     // routing key
		false,                 // mandatory
		false,                 // immediate
		message,
	); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.l.Info("event published", "exchange", p.config.ExchangeName, "event_type", event.EventType())

	return nil
}

// Close closes the publisher
// connectionのみクローズ（channelは各Publishで都度クローズされる）
func (p *RabbitMQPublisher) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
