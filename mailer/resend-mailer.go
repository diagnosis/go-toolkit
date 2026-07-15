package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// ResendMailer sends email through the Resend HTTP API.
type ResendMailer struct {
	apiKey string
	from   string
}

// NewResendMailer returns a ResendMailer that authenticates with apiKey
// and sends from the given address.
func NewResendMailer(apiKey, from string) *ResendMailer {
	return &ResendMailer{apiKey: apiKey, from: from}
}

// Send delivers an HTML email to the given recipients via the Resend API.
func (m *ResendMailer) Send(to []string, subject, body string) error {
	payload := map[string]any{
		"from":    m.from,
		"to":      to,
		"subject": subject,
		"html":    body,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}

	return nil
}
