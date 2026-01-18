package broker

import "github.com/furuya-3150/fam-diary-log/pkg/broker/publisher"

// DiaryPublisherConfig returns the publisher configuration for diary events
func DiaryPublisherConfig() publisher.Config {
	return publisher.Config{
		ExchangeName: "diary.events",
		ExchangeKind: "topic",
	}
}
