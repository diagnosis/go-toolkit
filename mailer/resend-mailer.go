package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/diagnosis/go-toolkit/v2/apperr"
)

// ResendMailer sends email through the Resend HTTP API.
type ResendMailer struct {
	apiKey  string
	from    string
	client  *http.Client
	baseURL string
}

// NewResendMailer returns a ResendMailer that authenticates with apiKey
// and sends from the given address.
func NewResendMailer(apiKey, from string) *ResendMailer {
	return &ResendMailer{
		apiKey:  apiKey,
		from:    from,
		client:  &http.Client{Timeout: 20 * time.Second},
		baseURL: "https://api.resend.com/emails",
	}
}

// Send delivers an HTML email to the given recipients via the Resend API.
func (m *ResendMailer) Send(ctx context.Context, to []string, subject, body string) error {
	ctxWithTimeOut, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	payload := map[string]any{
		"from":    m.from,
		"to":      to,
		"subject": subject,
		"html":    body,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return apperr.Internal("failed to deliver email", "failed to marshal payload", err)
	}

	req, err := http.NewRequestWithContext(ctxWithTimeOut, "POST", m.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return apperr.Internal("failed to deliver email", "http request initialization failed", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return apperr.Internal("failed to deliver email", "http client request failed", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))

		var resendJSON struct {
			Message string `json:"message"`
		}
		if err = json.Unmarshal(b, &resendJSON); err != nil {
			resendJSON.Message = string(b)
		}


		return apperr.Internal("email delivery failed", fmt.Sprintf("resend status=%d: %s", resp.StatusCode, resendJSON.Message))
	}

	return nil
}
