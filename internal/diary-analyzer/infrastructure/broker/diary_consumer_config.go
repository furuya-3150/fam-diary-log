package broker

import "github.com/furuya-3150/fam-diary-log/pkg/broker/consumer"

// DiaryConsumerConfig returns the consumer configuration for diary events
func DiaryConsumerConfig() consumer.Config {
	return consumer.Config{
		ExchangeName: "diary.events",
		ExchangeKind: "topic",
		QueueName:    "diary-analyzer.analyze",
		RoutingKeys:  []string{"diary.created"},
	}
}
