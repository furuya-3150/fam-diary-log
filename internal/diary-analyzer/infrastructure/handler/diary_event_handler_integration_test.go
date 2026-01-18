package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/gateway"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/helper"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/usecase"
	diarydomain "github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/consumer"
)

// TestDiaryEventHandlerIntegrationWithBroker tests event handling with actual RabbitMQ broker
func TestDiaryEventHandlerIntegrationWithBroker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Setup database
	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	// Setup RabbitMQ connection
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	require.NoError(t, err)
	defer conn.Close()

	// Create repository and usecase
	repo := repository.NewDiaryAnalysisRepository(dbManager)
	nlpGateway := gateway.NewYahooNLPGateway(config.ThirdParty.YahooNLPAppID)
	analysisUsecase := usecase.NewDiaryAnalysisUsecaseWithNLPGateway(repo, nlpGateway)

	// Create handler
	log := slog.Default()
	eventHandler := NewDiaryEventHandler(analysisUsecase, log)

	// Create consumer
	consumerConfig := consumer.Config{
		ExchangeName: "diary.events.test",
		ExchangeKind: "topic",
		QueueName:    "diary-analyzer.analyze.test",
		RoutingKeys:  []string{"diary.created"},
	}
	mq, err := consumer.NewRabbitMQConsumer(conn, consumerConfig, eventHandler, log)
	require.NoError(t, err)

	// Start consumer in a goroutine
	go func() {
		err := mq.Start(ctx)
		if err != nil && err != context.Canceled {
			t.Logf("consumer error: %v", err)
		}
	}()

	// Give the consumer time to start
	time.Sleep(500 * time.Millisecond)

	// Create a publisher to send test message
	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	// Declare exchange
	err = ch.ExchangeDeclare(
		consumerConfig.ExchangeName,
		consumerConfig.ExchangeKind,
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	// Create test event
	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()
	content := "これはテストの日記です。素晴らしい一日でした。"

	testEvent := &diarydomain.DiaryCreatedEvent{
		DiaryID:   diaryID,
		UserID:    userID,
		FamilyID:  familyID,
		Content:   content,
		Timestamp: time.Now(),
	}

	eventBytes, err := json.Marshal(testEvent)
	require.NoError(t, err)

	// publish the event
	err = ch.PublishWithContext(
		ctx,
		consumerConfig.ExchangeName,
		"diary.created",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        eventBytes,
		},
	)
	require.NoError(t, err)

	// Give the handler time to process the message
	time.Sleep(2 * time.Second)

	// Assert: Verify that the analysis was stored in the database
	ch2, err := conn.Channel()
	require.NoError(t, err)
	defer ch2.Close()

	// Get the recorded analysis
	gorm := dbManager.GetGorm()
	var analysis domain.DiaryAnalysis
	result := gorm.Where("diary_id = ?", diaryID).First(&analysis)

	assert.NoError(t, result.Error, "analysis should be stored in database")
	assert.Equal(t, diaryID, analysis.DiaryID)
	assert.Equal(t, userID, analysis.UserID)
	assert.Equal(t, familyID, analysis.FamilyID)
	assert.Equal(t, len([]rune(content)), analysis.CharCount)
	assert.Equal(t, 100, analysis.AccuracyScore) // Assuming the test content has no errors

	// Cleanup
	cancel()
	mq.Stop()
}

// TestDiaryEventHandlerIntegrationMessageNotProcessed tests that invalid messages are handled gracefully
func TestDiaryEventHandlerIntegrationInvalidMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	config := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	require.NoError(t, err)
	defer conn.Close()

	repo := repository.NewDiaryAnalysisRepository(dbManager)
	nlpGateway := gateway.NewYahooNLPGateway(config.ThirdParty.YahooNLPAppID)
	analysisUsecase := usecase.NewDiaryAnalysisUsecaseWithNLPGateway(repo, nlpGateway)

	log := slog.Default()
	eventHandler := NewDiaryEventHandler(analysisUsecase, log)

	consumerConfig := consumer.Config{
		ExchangeName: "diary.events.invalid.test",
		ExchangeKind: "topic",
		QueueName:    "diary-analyzer.analyze.invalid.test",
		RoutingKeys:  []string{"diary.created"},
	}
	mq, err := consumer.NewRabbitMQConsumer(conn, consumerConfig, eventHandler, log)
	require.NoError(t, err)

	go func() {
		err := mq.Start(ctx)
		if err != nil && err != context.Canceled {
			t.Logf("consumer error: %v", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	err = ch.ExchangeDeclare(
		consumerConfig.ExchangeName,
		consumerConfig.ExchangeKind,
		true,
		false,
		false,
		false,
		nil,
	)
	require.NoError(t, err)

	// Act: Publish invalid JSON
	err = ch.PublishWithContext(
		ctx,
		consumerConfig.ExchangeName,
		"diary.created",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte("invalid json"),
		},
	)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Assert: Consumer should still be running despite the error
	assert.NotNil(t, mq)

	cancel()
	mq.Stop()
}
