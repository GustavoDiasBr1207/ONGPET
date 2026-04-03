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
	apiURL    string
	apiKey    string
	enabled   bool
	httpClient *http.Client
}

// WhatsAppMessage representa uma mensagem para envio
type WhatsAppMessage struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

// NewWhatsAppService cria uma nova instância do serviço WhatsApp
func NewWhatsAppService() (*WhatsAppService, error) {
	apiURL := strings.TrimSpace(os.Getenv("WHATSAPP_API_URL"))
	apiKey := strings.TrimSpace(os.Getenv("WHATSAPP_API_KEY"))
	enabled := strings.ToLower(strings.TrimSpace(os.Getenv("WHATSAPP_ENABLED"))) == "true"

	// Se WhatsApp não está habilitado, retorna sem erro (graceful degradation)
	if !enabled {
		return &WhatsAppService{
			enabled: false,
		}, nil
	}

	if apiURL == "" || apiKey == "" {
		return nil, fmt.Errorf("WHATSAPP_API_URL ou WHATSAPP_API_KEY não configurados")
	}

	return &WhatsAppService{
		apiURL:    apiURL,
		apiKey:    apiKey,
		enabled:   true,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// SendMessage envia uma mensagem de WhatsApp usando Evolution API
// O número 'to' deve estar no formato: 55XXXXXXXXXX (código país + DDD + número, sem caracteres especiais)
// A instância do WhatsApp deve estar conectada na Evolution API
func (w *WhatsAppService) SendMessage(to, message string) error {
	if !w.enabled {
		return nil // Silenciosamente ignora se WhatsApp desabilitado
	}

	if to == "" || message == "" {
		return fmt.Errorf("número de telefone e mensagem são obrigatórios")
	}

	// Limpar número: remover caracteres especiais, deixar apenas dígitos
	cleanedTo := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		if r == '+' {
			return r
		}
		return -1
	}, to)

	cleanedTo = strings.TrimPrefix(cleanedTo, "+")

	// Validar formato básico: deve ter pelo menos 11 dígitos (55 + 11 dígitos do Brasil)
	if len(cleanedTo) < 11 {
		return fmt.Errorf("número de telefone inválido: %s (esperado formato 55XXXXXXXXXX)", to)
	}

	// Payload para Evolution API (formato JSON)
	payload := map[string]interface{}{
		"number": cleanedTo,
		"text":   message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %w", err)
	}

	// Construir requisição para Evolution API
	req, err := http.NewRequest("POST", w.apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.apiKey))

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem WhatsApp: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao ler resposta: %w", err)
	}

	// Evolution API retorna sucesso em 200, 201 ou formato variável
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendMessageWithRetry envia mensagem com tentativas de repetição
// Máximo 2 tentativas com backoff (não bloqueia fluxo principal)
func (w *WhatsAppService) SendMessageWithRetry(to, message string) {
	if !w.enabled {
		return
	}

	const maxRetries = 2
	const baseDelay = 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := w.SendMessage(to, message)
		if err == nil {
			// Sucesso
			return
		}

		if attempt < maxRetries-1 {
			// Backoff exponencial: 1s, 2s
			delay := baseDelay * time.Duration(1<<uint(attempt))
			time.Sleep(delay)
		} else {
			// Última tentativa falhou
			fmt.Printf("⚠️ Erro ao enviar WhatsApp para %s (após %d tentativas): %v\n",
				maskPhoneNumber(to), maxRetries, err)
		}
	}
}

// maskPhoneNumber mascara o número para logging (mostra apenas últimos 4 dígitos)
func maskPhoneNumber(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}
