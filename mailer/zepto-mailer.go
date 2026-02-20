package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ZeptoMailer struct {
	apiKey string
	from   ZeptoAddress
}

type ZeptoAddress struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type zeptoPayload struct {
	From     ZeptoAddress `json:"from"`
	To       []zeptoTo    `json:"to"`
	Subject  string       `json:"subject"`
	HTMLBody string       `json:"htmlbody"`
}

type zeptoTo struct {
	EmailAddress ZeptoAddress `json:"email_address"`
}

func NewZeptoMailer(apiKey, fromEmail, fromName string) *ZeptoMailer {
	return &ZeptoMailer{
		apiKey: apiKey,
		from: ZeptoAddress{
			Address: fromEmail,
			Name:    fromName,
		},
	}
}

func (m *ZeptoMailer) Send(to []string, subject, body string) error {
	recipients := make([]zeptoTo, len(to))
	for i, addr := range to {
		recipients[i] = zeptoTo{
			EmailAddress: ZeptoAddress{Address: addr},
		}
	}

	payload := zeptoPayload{
		From:     m.from,
		To:       recipients,
		Subject:  subject,
		HTMLBody: body,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.zeptomail.com/v1.1/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Zoho-enczapikey "+m.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("zepto API error: status %d", resp.StatusCode)
	}

	return nil
}
