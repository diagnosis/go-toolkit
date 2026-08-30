// Package mailer sends HTML email through SMTP, Resend, or ZeptoMail
// behind a single Mailer interface.
package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// Mailer sends an HTML email with the given subject and body to one or
// more recipients.
type Mailer interface {
	Send(ctx context.Context,to []string, subject, body string) error
}

// SMTPMailer sends email through a plain SMTP server using the settings
// in a Config.
type SMTPMailer struct {
	config *Config
}

// NewSMTPMailer returns an SMTPMailer that sends through the server
// described by cfg.
func NewSMTPMailer(cfg *Config) *SMTPMailer {
	return &SMTPMailer{config: cfg}
}

// Send delivers an HTML email to the given recipients via SMTP.
// ctx is accepted for interface compliance; net/smtp cannot honor cancellation.
func (m *SMTPMailer) Send(ctx context.Context,to []string, subject, body string) error {

	auth := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)
	tos := strings.Join(to, ",")

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		m.config.From,
		tos,
		subject,
		body,
	)

	addr := fmt.Sprintf("%s:%d", m.config.Host, m.config.Port)
	err := smtp.SendMail(addr, auth, m.config.From, to, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
