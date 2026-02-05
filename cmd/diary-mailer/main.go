package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	mailBroker "github.com/furuya-3150/fam-diary-log/internal/mail/infrastructure/broker"
	mailHandler "github.com/furuya-3150/fam-diary-log/internal/mail/infrastructure/handler"
	mailSender "github.com/furuya-3150/fam-diary-log/internal/mail/infrastructure/sender"
	mailTemplate "github.com/furuya-3150/fam-diary-log/internal/mail/infrastructure/template"
	mailUsecase "github.com/furuya-3150/fam-diary-log/internal/mail/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/consumer"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/rabbit"
	"github.com/furuya-3150/fam-diary-log/pkg/logger"
	"github.com/joho/godotenv"
)

func init() {
	var log *slog.Logger
	if os.Getenv("GO_ENV") == "dev" {
		log = logger.New(slog.LevelDebug)
	} else {
		log = logger.New(slog.LevelInfo)
	}
	slog.SetDefault(log)

	if os.Getenv("GO_ENV") == "dev" {
		_ = godotenv.Load("./cmd/mail/.env")
	}
}

func main() {
	ctx := context.Background()
	log := slog.Default()

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		log.Error("RABBITMQ_URL is not set")
		os.Exit(1)
	}

	conn, err := rabbit.NewConnection(rabbit.Config{URL: rabbitmqURL})
	if err != nil {
		log.Error("failed to connect to RabbitMQ", "error", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// prepare components
	tplStore := mailTemplate.NewInMemoryStore()
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")
	sender := mailSender.NewSMTPSender(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)

	uc := mailUsecase.NewMailUsecase(sender, tplStore)

	handler := mailHandler.NewMailEventHandler(uc)

	consumerConfig := mailBroker.MailConsumerConfig()

	c, err := consumer.NewRabbitMQConsumer(conn, consumerConfig, handler, log)
	if err != nil {
		log.Error("failed to create consumer", "error", err.Error())
		os.Exit(1)
	}

	if err := c.Start(ctx); err != nil {
		log.Error("failed to start consumer", "error", err.Error())
		os.Exit(1)
	}

	log.Info("mail service started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("shutting down mail service")
	if err := c.Stop(); err != nil {
		log.Error("failed to stop consumer", "error", err.Error())
		os.Exit(1)
	}
	log.Info("mail service stopped")
}
