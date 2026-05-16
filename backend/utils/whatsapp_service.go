package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type WhatsAppService struct {
	apiURL        string
	apiKey        string
	SenderNumber  string
	enabled       bool
	httpClient    *http.Client
}

type WhatsAppMessage struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

func NewWhatsAppService() (*WhatsAppService, error) {
	apiURL       := strings.TrimSpace(os.Getenv("WHATSAPP_API_URL"))
	apiKey       := strings.TrimSpace(os.Getenv("WHATSAPP_API_KEY"))
	senderNumber := strings.TrimSpace(os.Getenv("WHATSAPP_SENDER_NUMBER"))
	enabled      := strings.ToLower(strings.TrimSpace(os.Getenv("WHATSAPP_ENABLED"))) == "true"

	if !enabled {
		return &WhatsAppService{enabled: false}, nil
	}

	if apiURL == "" || apiKey == "" {
		return nil, fmt.Errorf("WHATSAPP_API_URL ou WHATSAPP_API_KEY não configurados")
	}

	fmt.Println("📡 WhatsApp API URL:", apiURL)
	fmt.Println("📱 WhatsApp Sender:", senderNumber)

	return &WhatsAppService{
		apiURL:       apiURL,
		apiKey:       apiKey,
		SenderNumber: senderNumber,
		enabled:      true,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// 🔥 Padroniza número para 55XXXXXXXXXX
func normalizePhoneNumber(phone string) (string, error) {
	clean := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	if len(clean) < 10 {
		return "", fmt.Errorf("número inválido: %s", phone)
	}

	// Se não começar com 55, adiciona
	if !strings.HasPrefix(clean, "55") {
		clean = "55" + clean
	}

	if len(clean) < 12 || len(clean) > 13 {
		return "", fmt.Errorf("número fora do padrão BR: %s", clean)
	}

	return clean, nil
}

func (w *WhatsAppService) SendMessage(to, message string) error {
	if !w.enabled {
		return nil
	}

	if to == "" || message == "" {
		return fmt.Errorf("telefone ou mensagem vazios")
	}

	normalized, err := normalizePhoneNumber(to)
	if err != nil {
		fmt.Printf("❌ normalizePhoneNumber(%q): %v\n", to, err)
		return err
	}

	payload := WhatsAppMessage{
		Number: normalized,
		Text:   message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar JSON: %w", err)
	}

	fmt.Println("📤 Enviando WhatsApp para:", normalized)
	fmt.Println("📝 Mensagem:", message)

	req, err := http.NewRequest("POST", w.apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", w.apiKey)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro HTTP: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Println("📥 Status:", resp.StatusCode)
	fmt.Println("📥 Response:", string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("erro API (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// 🔥 AGORA RETORNA ERRO
func (w *WhatsAppService) SendMessageWithRetry(to, message string) error {
	if !w.enabled {
		return nil
	}

	const maxRetries = 3
	const baseDelay = 2 * time.Second

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := w.SendMessage(to, message)
		if err == nil {
			fmt.Println("✅ WhatsApp enviado com sucesso")
			return nil
		}

		lastErr = err
		fmt.Printf("⚠️ Tentativa %d falhou: %v\n", attempt+1, err)

		if attempt < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<attempt)
			time.Sleep(delay)
		}
	}

	fmt.Printf("❌ Falha final ao enviar para %s: %v\n", maskPhoneNumber(to), lastErr)
	return lastErr
}

func maskPhoneNumber(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}