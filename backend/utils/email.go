package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const sendgridURL = "https://api.sendgrid.com/v3/mail/send"

type Mailer struct {
	apiKey string
	from   string
	client *http.Client
}

func New() (*Mailer, error) {
	apiKey := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if apiKey == "" {
		return nil, fmt.Errorf("SMTP_PASS (SendGrid API key) não configurado")
	}

	return &Mailer{
		apiKey: apiKey,
		from:   from,
		client: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

type sgAddress struct {
	Email string `json:"email"`
}

type sgPersonalization struct {
	To []sgAddress `json:"to"`
}

type sgContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type sgPayload struct {
	Personalizations []sgPersonalization `json:"personalizations"`
	From             sgAddress           `json:"from"`
	Subject          string              `json:"subject"`
	Content          []sgContent         `json:"content"`
}

func (m *Mailer) Send(to []string, subject, body string) error {
	toAddrs := make([]sgAddress, len(to))
	for i, addr := range to {
		toAddrs[i] = sgAddress{Email: addr}
	}

	payload := sgPayload{
		Personalizations: []sgPersonalization{{To: toAddrs}},
		From:             sgAddress{Email: m.from},
		Subject:          subject,
		Content:          []sgContent{{Type: "text/html", Value: body}},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	req, err := http.NewRequest("POST", sendgridURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("erro HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SendGrid erro (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
