package mailer

import (
	"fmt"
	"net/smtp"
)

type Mailer interface {
	Send(to []string, subject, body string) error
}

type SMTPMailer struct {
	config *Config
}

func NewSMTPMailer(cfg *Config) *SMTPMailer {
	return &SMTPMailer{config: cfg}
}

func (m *SMTPMailer) Send(to []string, subject, body string) error {

	auth := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		m.config.From,
		to[0],
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
