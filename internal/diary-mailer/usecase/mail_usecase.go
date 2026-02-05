package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/furuya-3150/fam-diary-log/internal/diary-mailer/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary-mailer/infrastructure/sender"
	"github.com/furuya-3150/fam-diary-log/internal/diary-mailer/infrastructure/template"
)

type MailUsecase interface {
	Send(ctx context.Context, m *domain.MailMessage) error
}

type mailUsecase struct {
	sender sender.MailSender
	tpl    template.TemplateStore
}

func NewMailUsecase(s sender.MailSender, t template.TemplateStore) MailUsecase {
	return &mailUsecase{sender: s, tpl: t}
}

func (u *mailUsecase) Send(ctx context.Context, m *domain.MailMessage) error {
	tpl, err := u.tpl.Get(m.TemplateID, m.Locale)
	if err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	textBody, htmlBody, err := tpl.Render(m.Payload)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	subject := tpl.Subject
	if m.Subject != "" {
		subject = m.Subject
	}

	if err := u.sender.Send(ctx, m.To, m.Cc, m.Bcc, subject, textBody, htmlBody, m.ReplyTo, m.Headers); err != nil {
		return err
	}

	slog.Info("mail sent", "to", m.To, "template", m.TemplateID)
	return nil
}
