package handler

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/furuya-3150/fam-diary-log/internal/mail/domain"
	"github.com/furuya-3150/fam-diary-log/internal/mail/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/consumer"
)

type MailEventHandler struct {
	uc  usecase.MailUsecase
}

func NewMailEventHandler(uc usecase.MailUsecase) consumer.EventHandler {
	return &MailEventHandler{uc: uc}
}

func (h *MailEventHandler) Handle(ctx context.Context, eventType string, content []byte) error {
	var m domain.MailMessage
	if err := json.Unmarshal(content, &m); err != nil {
		slog.Error("failed to unmarshal mail message", "error", err.Error())
		return err
	}

	slog.Info("received mail message", "type", "template", m.TemplateID)
	if err := h.uc.Send(ctx, &m); err != nil {
		slog.Error("failed to send mail", "error", err.Error())
		return err
	}

	return nil
}
