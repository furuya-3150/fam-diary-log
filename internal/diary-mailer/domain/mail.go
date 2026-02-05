package domain

type MailMessage struct {
	TemplateID      string                 `json:"template_id"`
	Subject         string                 `json:"subject,omitempty"`
	To              []string               `json:"to"`
	Cc              []string               `json:"cc,omitempty"`
	Bcc             []string               `json:"bcc,omitempty"`
	Locale          string                 `json:"locale,omitempty"`
	ReplyTo         string                 `json:"reply_to,omitempty"`
	Headers         map[string]string      `json:"headers,omitempty"`
	Payload         map[string]interface{} `json:"payload,omitempty"`
}
