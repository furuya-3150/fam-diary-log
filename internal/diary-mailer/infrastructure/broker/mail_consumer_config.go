package broker

import "github.com/furuya-3150/fam-diary-log/pkg/broker/consumer"

// MailConsumerConfig returns the consumer configuration for mail commands
func MailConsumerConfig() consumer.Config {
	return consumer.Config{
		ExchangeName: "mail.commands",
		ExchangeKind: "topic",
		QueueName:    "mail-service.send",
		RoutingKeys:  []string{"mail.send"},
	}
}
