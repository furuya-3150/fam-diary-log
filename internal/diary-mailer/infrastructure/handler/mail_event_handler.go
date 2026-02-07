package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/furuya-3150/fam-diary-log/internal/diary-mailer/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary-mailer/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/consumer"
)

type MailEventHandler struct {
	uc usecase.MailUsecase
}

func NewMailEventHandler(uc usecase.MailUsecase) consumer.EventHandler {
	return &MailEventHandler{uc: uc}
}

func (h *MailEventHandler) Handle(ctx context.Context, eventType string, content []byte) error {
	var m domain.MailMessage
	if err := json.Unmarshal(content, &m); err != nil {
		slog.Error("failed to unmarshal mail message (poison)", "error", err.Error(), "raw", string(content))
		// malformed message: ack and drop to avoid infinite requeue
		return nil
	}

	slog.Info("received mail message", "type", "template", "template_id", m.TemplateID, "locale", m.Locale, "to", m.To)
	if err := h.uc.Send(ctx, &m); err != nil {
		// treat template-not-found as unrecoverable (ack and drop)
		if strings.Contains(err.Error(), "template not found") {
			slog.Error("template not found for mail message (dropping)", "error", err.Error(), "template_id", m.TemplateID, "locale", m.Locale, "raw", string(content))
			return nil
		}
		slog.Error("failed to send mail (transient)", "error", err.Error(), "template_id", m.TemplateID)
		return err
	}

	return nil
}
