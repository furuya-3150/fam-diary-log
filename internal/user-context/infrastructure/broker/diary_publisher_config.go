package broker

import "github.com/furuya-3150/fam-diary-log/pkg/broker/publisher"

// DiaryPublisherConfig returns the publisher configuration for mail commands
func DiaryPublisherConfig() publisher.Config {
	return publisher.Config{
		ExchangeName: "mail.commands",
		ExchangeKind: "topic",
	}
}
