package utils

import (
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

// TestValidatePhoneNumber testa validação de números de telefone
func TestValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		phone   string
		wantErr bool
		name    string
	}{
		{"5527992345678", false, "Número válido Brasil (13 dígitos)"},
		{"55 27 99234-5678", false, "Número válido com formatação"},
		{"27992345678", false, "Número sem código de país (11 dígitos)"},
		{"123", true, "Número muito curto"},
		{"", true, "Telefone vazio"},
		{"+55 27 99234-5678", false, "Número com + e formatação"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePhoneNumber(tt.phone)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePhoneNumber(%q) error = %v, wantErr %v", tt.phone, err, tt.wantErr)
			}
		})
	}
}

// TestNormalizePhoneNumber testa normalização de números para o padrão 55XXXXXXXXXX
func TestNormalizePhoneNumber(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
		name    string
	}{
		{"5527992345678", "5527992345678", false, "Número já normalizado"},
		{"27992345678", "5527992345678", false, "Adiciona 55 automaticamente"},
		{"(27) 99234-5678", "5527992345678", false, "Remove formatação e adiciona 55"},
		{"+55 27 99234-5678", "5527992345678", false, "Remove + e espaços"},
		{"123", "", true, "Muito curto — deve falhar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizePhoneNumber(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizePhoneNumber(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("normalizePhoneNumber(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestMaskPhoneNumber testa mascaramento de número para logging
func TestMaskPhoneNumber(t *testing.T) {
	tests := []struct {
		phone    string
		expected string
		name     string
	}{
		{"5527992345678", "****5678", "Máscara telefone completo"},
		{"1234", "****", "Telefone muito curto"},
		{"", "****", "Telefone vazio"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskPhoneNumber(tt.phone)
			if result != tt.expected {
				t.Errorf("maskPhoneNumber(%q) = %q, want %q", tt.phone, result, tt.expected)
			}
		})
	}
}

// TestNewAdoptionMessages testa criação de mensagens WhatsApp
func TestNewAdoptionMessages(t *testing.T) {
	t.Run("NewAdoptionRequestWhatsAppMessage", func(t *testing.T) {
		msg := NewAdoptionRequestWhatsAppMessage("Rex", "João", "ONG Amigos")
		if !strings.Contains(msg, "Rex") {
			t.Error("mensagem não contém nome do pet")
		}
		if !strings.Contains(msg, "João") {
			t.Error("mensagem não contém nome do solicitante")
		}
		if !strings.Contains(msg, "ONG Amigos") {
			t.Error("mensagem não contém nome da ONG")
		}
	})

	t.Run("NewAdoptionApprovedWhatsAppMessage", func(t *testing.T) {
		msg := NewAdoptionApprovedWhatsAppMessage("Maria", "Mimi", "ONG Cat", "5527992345678")
		if !strings.Contains(msg, "APROVADA") {
			t.Error("mensagem de aprovação não contém 'APROVADA'")
		}
		if !strings.Contains(msg, "5527992345678") {
			t.Error("mensagem não contém telefone da ONG")
		}
	})

	t.Run("NewAdoptionRejectedWhatsAppMessage", func(t *testing.T) {
		msg := NewAdoptionRejectedWhatsAppMessage("Pedro", "Buddy", "ONG Dogs")
		if !strings.Contains(msg, "não aprovada") {
			t.Error("mensagem de rejeição não contém 'não aprovada'")
		}
	})

	t.Run("NewAdoptionConfirmationWhatsAppMessage", func(t *testing.T) {
		msg := NewAdoptionConfirmationWhatsAppMessage("Ana", "Thor", "ONG Pets")
		if !strings.Contains(msg, "Ana") {
			t.Error("mensagem de confirmação não contém nome do solicitante")
		}
		if !strings.Contains(msg, "Thor") {
			t.Error("mensagem de confirmação não contém nome do pet")
		}
	})
}

// TestWhatsAppServiceInit verifica se o serviço inicializa corretamente
func TestWhatsAppServiceInit(t *testing.T) {
	_ = godotenv.Load("../.env")

	if err := InitWhatsAppService(); err != nil {
		t.Fatalf("falha ao inicializar WhatsApp service: %v", err)
	}

	svc := GetWhatsAppService()
	if svc == nil {
		t.Fatal("GetWhatsAppService() retornou nil após inicialização")
	}

	if svc.enabled {
		t.Logf("✅ WhatsApp service inicializado e habilitado (sender: %s)", svc.SenderNumber)
		if svc.apiURL == "" {
			t.Error("WHATSAPP_API_URL está vazio com serviço habilitado")
		}
		if svc.apiKey == "" {
			t.Error("WHATSAPP_API_KEY está vazio com serviço habilitado")
		}
		if svc.SenderNumber == "" {
			t.Error("WHATSAPP_SENDER_NUMBER está vazio com serviço habilitado")
		}
	} else {
		t.Log("ℹ️ WhatsApp service inicializado mas desabilitado (WHATSAPP_ENABLED != true)")
	}
}

// TestSendRealWhatsApp envia uma mensagem real para verificar o fluxo completo.
// Usa TEST_WHATSAPP_NUMBER se definido, caso contrário usa WHATSAPP_SENDER_NUMBER do .env.
func TestSendRealWhatsApp(t *testing.T) {
	_ = godotenv.Load("../.env")

	to := os.Getenv("TEST_WHATSAPP_NUMBER")
	if to == "" {
		to = os.Getenv("WHATSAPP_SENDER_NUMBER")
	}
	if to == "" {
		t.Skip("WHATSAPP_SENDER_NUMBER e TEST_WHATSAPP_NUMBER não definidos — configure o .env para executar")
	}

	if err := InitWhatsAppService(); err != nil {
		t.Fatalf("falha ao inicializar WhatsApp service: %v", err)
	}

	svc := GetWhatsAppService()
	if svc == nil {
		t.Fatal("WhatsApp service é nil após inicialização")
	}

	if !svc.enabled {
		t.Skip("WhatsApp desabilitado — verifique WHATSAPP_ENABLED=true no .env")
	}

	msg := NewAdoptionConfirmationWhatsAppMessage("Equipe OngPet", "Rex (teste)", "ONG Teste")
	if err := svc.SendMessageWithRetry(to, msg); err != nil {
		t.Fatalf("❌ falha ao enviar WhatsApp: %v", err)
	}

	t.Logf("✅ WhatsApp enviado com sucesso para: %s", maskPhoneNumber(to))
}

// BenchmarkPhoneValidation benchmark de validação de telefone
func BenchmarkPhoneValidation(b *testing.B) {
	phone := "5527992345678"
	for i := 0; i < b.N; i++ {
		_ = ValidatePhoneNumber(phone)
	}
}
