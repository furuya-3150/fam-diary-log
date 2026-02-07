package domain

import "github.com/furuya-3150/fam-diary-log/pkg/events"

// MailSendEvent represents an event to request sending a mail
type MailSendEvent struct {
	TemplateID      string                 `json:"template_id"`
	TemplateVersion string                 `json:"template_version,omitempty"`
	Subject         string                 `json:"subject,omitempty"`
	To              []string               `json:"to,omitempty"`
	Cc              []string               `json:"cc,omitempty"`
	Bcc             []string               `json:"bcc,omitempty"`
	Locale          string                 `json:"locale,omitempty"`
	ReplyTo         string                 `json:"reply_to,omitempty"`
	Headers         map[string]string      `json:"headers,omitempty"`
	Payload         map[string]interface{} `json:"payload,omitempty"`
}

// EventType implements events.Event
func (m *MailSendEvent) EventType() string {
	return "mail.send"
}

var _ events.Event = (*MailSendEvent)(nil)
