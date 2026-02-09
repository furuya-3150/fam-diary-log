package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/broker"
	analyzerConfig "github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/gateway"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/handler"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/consumer"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/rabbit"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/furuya-3150/fam-diary-log/pkg/logger"
	"github.com/joho/godotenv"
)

func init() {
	// ログ設定
	var log *slog.Logger
	if os.Getenv("GO_ENV") == "dev" {
		log = logger.New(slog.LevelDebug)
	} else {
		log = logger.New(slog.LevelInfo)
	}

	slog.SetDefault(log)

	// env読み込み
	if os.Getenv("GO_ENV") == "dev" {
		_ = godotenv.Load("./cmd/diary-analyzer/.env")
	}
}

func main() {
	ctx := context.Background()
	log := slog.Default()

	// Get RabbitMQ URL from environment
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		log.Error("RABBITMQ_URL is not set")
		os.Exit(1)
	}

	// Connect to RabbitMQ
	rabbitConfig := rabbit.Config{URL: rabbitmqURL}
	conn, err := rabbit.NewConnection(rabbitConfig)
	if err != nil {
		log.Error("failed to connect to RabbitMQ", "error", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// Initialize database and repository
	config := analyzerConfig.Load()
	dbManager := db.NewDBManager(config.DB.DatabaseURL)
	diaryAnalysisRepository := repository.NewDiaryAnalysisRepository(dbManager)

	gateway := gateway.NewYahooNLPGateway(config.ThirdParty.YahooNLPAppID)

	analyzerUsecase := usecase.NewDiaryAnalysisUsecase(diaryAnalysisRepository, gateway)

	eventHandler := handler.NewDiaryEventHandler(analyzerUsecase, log)

	consumerConfig := broker.DiaryConsumerConfig()

	// Create consumer
	c, err := consumer.NewRabbitMQConsumer(conn, consumerConfig, eventHandler, log)
	if err != nil {
		log.Error("failed to create consumer", "error", err.Error())
		os.Exit(1)
	}

	// Start consuming events
	if err := c.Start(ctx); err != nil {
		log.Error("failed to start consumer", "error", err.Error())
		os.Exit(1)
	}

	log.Info("diary-analyzer started", "version", "1.0.0")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// block until a signal is received
	<-sigChan

	log.Info("shutting down diary-analyzer")

	// Stop consumer
	if err := c.Stop(); err != nil {
		log.Error("failed to stop consumer", "error", err.Error())
		os.Exit(1)
	}

	log.Info("diary-analyzer stopped")
}
