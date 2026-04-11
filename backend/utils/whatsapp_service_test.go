package utils

import (
	"strings"
	"testing"
)

// TestValidatePhoneNumber testa validação de números de telefone
func TestValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		phone   string
		wantErr bool
		name    string
	}{
		{
			phone:   "5527992345678",
			wantErr: false,
			name:    "Número válido Brasil (11 dígitos)",
		},
		{
			phone:   "55 27 99234-5678",
			wantErr: false,
			name:    "Número válido com formatação",
		},
		{
			phone:   "27992345678",
			wantErr: false,
			name:    "Número sem código de país (11 dígitos)",
		},
		{
			phone:   "123",
			wantErr: true,
			name:    "Número muito curto",
		},
		{
			phone:   "",
			wantErr: true,
			name:    "Telefone vazio",
		},
		{
			phone:   "+55 27 99234-5678",
			wantErr: false,
			name:    "Número com + e formatação",
		},
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

// TestMaskPhoneNumber testa mascaramento de número para logging
func TestMaskPhoneNumber(t *testing.T) {
	tests := []struct {
		phone    string
		expected string
		name     string
	}{
		{
			phone:    "5527992345678",
			expected: "****5678",
			name:     "Máscara telefone completo",
		},
		{
			phone:    "1234",
			expected: "****",
			name:     "Telefone muito curto",
		},
		{
			phone:    "",
			expected: "****",
			name:     "Telefone vazio",
		},
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

// TestNewAdoptionMessages testa criação de mensagens
func TestNewAdoptionMessages(t *testing.T) {
	t.Run("NewAdoptionRequestWhatsAppMessage", func(t *testing.T) {
		msg := NewAdoptionRequestWhatsAppMessage("Rex", "João", "ONG Amigos")
		if !strings.Contains(msg, "Rex") {
			t.Errorf("Mensagem não contém nome do pet")
		}
		if !strings.Contains(msg, "João") {
			t.Errorf("Mensagem não contém nome do solicitante")
		}
		if !strings.Contains(msg, "ONG Amigos") {
			t.Errorf("Mensagem não contém nome da ONG")
		}
	})

	t.Run("NewAdoptionApprovedWhatsAppMessage", func(t *testing.T) {
		msg := NewAdoptionApprovedWhatsAppMessage("Maria", "Mimi", "ONG Cat", "5527992345678")
		if !strings.Contains(msg, "APROVADA") {
			t.Errorf("Mensagem de aprovação não contém 'APROVADA'")
		}
		if !strings.Contains(msg, "5527992345678") {
			t.Errorf("Mensagem não contém telefone da ONG")
		}
	})

	t.Run("NewAdoptionRejectedWhatsAppMessage", func(t *testing.T) {
		msg := NewAdoptionRejectedWhatsAppMessage("Pedro", "Buddy", "ONG Dogs")
		if !strings.Contains(msg, "não aprovada") {
			t.Errorf("Mensagem de rejeição não contém 'não aprovada'")
		}
	})
}

// BenchmarkSendMessage benchmark do envio de mensagem
// Execute com: go test -bench=. ./utils
func BenchmarkPhoneValidation(b *testing.B) {
	phone := "5527992345678"
	for i := 0; i < b.N; i++ {
		_ = ValidatePhoneNumber(phone)
	}
}

// Exemplo de teste de integração (comentado, para uso manual)
/*
func TestWhatsAppServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	svc, err := NewWhatsAppService()
	if err != nil {
		t.Fatalf("Falha ao criar WhatsApp service: %v", err)
	}

	if !svc.enabled {
		t.Skip("WhatsApp desabilitado")
	}

	// Teste com número de sandbox ou teste
	// IMPORTANTE: NÃO usar números reais em testes
	testPhone := os.Getenv("TEST_WHATSAPP_NUMBER")
	if testPhone == "" {
		t.Skip("TEST_WHATSAPP_NUMBER não configurado")
	}

	err = svc.SendMessage(testPhone, "Teste de mensagem de integração")
	if err != nil {
		t.Errorf("Falha ao enviar mensagem: %v", err)
	}
}
*/
