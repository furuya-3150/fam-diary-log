package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/publisher"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/rabbit"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestInviteMembers_PublishesMailAndMailcatcherReceives(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires RabbitMQ and MailCatcher")
	}
	godotenv.Load("../../../cmd/user-context/.env")
	rUrl := os.Getenv("RABBITMQ_URL")
	mcUrl := os.Getenv("MAILCATCHER_URL")
	if rUrl == "" || mcUrl == "" {
		t.Skip("integration test skipped: set RABBITMQ_URL and MAILCATCHER_URL to run")
	}

	// connect to rabbit
	conn, err := rabbit.NewConnection(rabbit.Config{URL: rUrl})
	require.NoError(t, err)
	defer conn.Close()

	pub, err := publisher.NewRabbitMQPublisher(conn, publisher.Config{ExchangeName: "mail.commands", ExchangeKind: "topic"}, slog.Default())
	require.NoError(t, err)
	defer pub.Close()

	// create event matching templates in the repo
	event := &domain.MailSendEvent{
		TemplateID: "family_invite_v1",
		Locale:     "ja",
		To:         []string{"integration-test@example.com"},
		Payload: map[string]interface{}{
			"inviter_name": "IntegrationTester",
			"family_name":  "IntegrationFamily",
			"app_url":      "https://example.local",
		},
	}

	respBeforePublish, err := http.Get(mcUrl + "/messages")
	if err != nil {
		t.Fatalf("failed to connect to MailCatcher at %s: %v", mcUrl, err)
	}
	var msgs []map[string]interface{}
	_ = json.NewDecoder(respBeforePublish.Body).Decode(&msgs)

	respBeforePublish.Body.Close()
	initialMsgCount := len(msgs)


	require.NoError(t, pub.Publish(context.Background(), event))

	// Poll MailCatcher for the message
	found := false
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(mcUrl + "/messages")
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = json.NewDecoder(resp.Body).Decode(&msgs)
			resp.Body.Close()
			if len(msgs) > initialMsgCount {
				found = true
			}
		}
		if found {
			break
		}
		time.Sleep(1 * time.Second)
	}

	require.True(t, found, "MailCatcher did not receive the published message")
}
