package sender

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
)

type MailSender interface {
	Send(ctx context.Context, to, cc, bcc []string, subject, textBody, htmlBody string, replyTo string, headers map[string]string) error
}

type SMTPSender struct {
	host string
	port string
	auth smtp.Auth
	from string
}

func NewSMTPSender(host, port, username, password, from string) *SMTPSender {
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	return &SMTPSender{host: host, port: port, auth: auth, from: from}
}

func (s *SMTPSender) Send(ctx context.Context, to, cc, bcc []string, subject, textBody, htmlBody string, replyTo string, headers map[string]string) error {
	recipients := append(append([]string{}, to...), cc...)
	recipients = append(recipients, bcc...)
	if len(recipients) == 0 {
		return fmt.Errorf("no recipients")
	}

	header := make(map[string]string)
	header["From"] = s.from
	header["To"] = strings.Join(to, ", ")
	if len(cc) > 0 {
		header["Cc"] = strings.Join(cc, ", ")
	}
	header["Subject"] = subject
	if replyTo != "" {
		header["Reply-To"] = replyTo
	}
	header["MIME-Version"] = "1.0"
	boundary := "----=_Part_0_123456789.123456789"
	header["Content-Type"] = fmt.Sprintf("multipart/alternative; boundary=%s", boundary)

	// merge custom headers without overwriting standard ones
	for k, v := range headers {
		lk := strings.ToLower(k)
		if lk == "from" || lk == "to" || lk == "subject" || lk == "mime-version" || lk == "content-type" || lk == "reply-to" || lk == "cc" {
			continue
		}
		header[k] = v
	}

	// Build multipart/alternative message
	var msgBuilder strings.Builder
	for k, v := range header {
		msgBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msgBuilder.WriteString("\r\n")

	// text part
	msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msgBuilder.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	msgBuilder.WriteString(textBody)
	msgBuilder.WriteString("\r\n")

	// html part
	msgBuilder.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msgBuilder.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
	msgBuilder.WriteString(htmlBody)
	msgBuilder.WriteString("\r\n")

	// end
	msgBuilder.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	msg := msgBuilder.String()

	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	// Try TLS connection if possible
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: s.host})
	if err == nil {
		c, err := smtp.NewClient(conn, s.host)
		if err == nil {
			if s.auth != nil {
				if ok, _ := c.Extension("AUTH"); ok {
					if err := c.Auth(s.auth); err != nil {
						slog.Error("smtp auth failed", "error", err.Error())
					}
				}
			}
			if err := c.Mail(s.from); err != nil {
				return err
			}
			for _, rcpt := range recipients {
				if err := c.Rcpt(rcpt); err != nil {
					return err
				}
			}
			w, err := c.Data()
			if err != nil {
				return err
			}
			_, err = w.Write([]byte(msg))
			if err != nil {
				return err
			}
			if err := w.Close(); err != nil {
				return err
			}
			c.Quit()
			return nil
		}
	}

	// Fallback to plain smtp.SendMail
	if err := smtp.SendMail(addr, s.auth, s.from, recipients, []byte(msg)); err != nil {
		return err
	}

	return nil
}
