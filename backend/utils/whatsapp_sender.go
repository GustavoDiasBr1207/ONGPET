package utils

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	whatsappInstance *WhatsAppService
	whatsappMu       sync.Mutex
)

// InitWhatsAppService inicializa o serviço WhatsApp global. Pode ser chamado
// novamente se a inicialização anterior falhou.
func InitWhatsAppService() error {
	whatsappMu.Lock()
	defer whatsappMu.Unlock()

	if whatsappInstance != nil {
		return nil
	}

	svc, err := NewWhatsAppService()
	if err != nil {
		log.Println("⚠️ Falha ao inicializar WhatsApp service:", err)
		return err
	}

	whatsappInstance = svc
	if svc.enabled {
		log.Println("✅ WhatsApp service inicializado com sucesso")
	} else {
		log.Println("ℹ️ WhatsApp service desabilitado (WHATSAPP_ENABLED != true)")
	}
	return nil
}

// GetWhatsAppService retorna a instância global do serviço WhatsApp
func GetWhatsAppService() *WhatsAppService {
	whatsappMu.Lock()
	inst := whatsappInstance
	whatsappMu.Unlock()

	if inst != nil {
		return inst
	}

	_ = InitWhatsAppService()

	whatsappMu.Lock()
	defer whatsappMu.Unlock()
	return whatsappInstance
}

// ─────────────────────────────────────────────────────────────
// Funções de Envio de Mensagens WhatsApp
// Cada função segue o padrão: extrai dados, cria mensagem, envia assincronamente
// ─────────────────────────────────────────────────────────────

// SendWhatsAppAdoptionRequest envia WhatsApp para ONG sobre nova solicitação
// telefoneOng: número no formato 55XXXXXXXXXX (sem caracteres especiais)
func SendWhatsAppAdoptionRequest(telefoneOng, petName, solicitanteName, ongName string) {
	svc := GetWhatsAppService()
	if svc == nil || !svc.enabled {
		return
	}

	// Enviar em goroutine assíncrona (não bloqueia)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ Panic ao enviar WhatsApp (nova adoção para ONG): %v\n", r)
			}
		}()

		message := NewAdoptionRequestWhatsAppMessage(petName, solicitanteName, ongName)
		start := time.Now()

		err := svc.SendMessageWithRetry(telefoneOng, message)

		duration := time.Since(start).Milliseconds()
		if err != nil {
			log.Printf("❌ Falha ao enviar WhatsApp (nova adoção → ONG) para %s em %dms: %v\n",
				maskPhoneNumber(telefoneOng), duration, err)
		} else {
			log.Printf("✉️ WhatsApp enviado (nova adoção → ONG) para %s em %dms\n",
				maskPhoneNumber(telefoneOng), duration)
		}
	}()
}

// SendWhatsAppAdoptionConfirmation envia confirmação para solicitante que pedido foi recebido
// telefoneSolicitante: número no formato 55XXXXXXXXXX
func SendWhatsAppAdoptionConfirmation(telefoneSolicitante, solicitanteName, petName, ongName string) {
	svc := GetWhatsAppService()
	if svc == nil || !svc.enabled {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ Panic ao enviar WhatsApp (confirmação para solicitante): %v\n", r)
			}
		}()

		message := NewAdoptionConfirmationWhatsAppMessage(solicitanteName, petName, ongName)
		start := time.Now()

		err := svc.SendMessageWithRetry(telefoneSolicitante, message)

		duration := time.Since(start).Milliseconds()
		if err != nil {
			log.Printf("❌ Falha ao enviar WhatsApp (confirmação → solicitante) para %s em %dms: %v\n",
				maskPhoneNumber(telefoneSolicitante), duration, err)
		} else {
			log.Printf("✉️ WhatsApp enviado (confirmação → solicitante) para %s em %dms\n",
				maskPhoneNumber(telefoneSolicitante), duration)
		}
	}()
}

// SendWhatsAppAdoptionApproved envia confirmação de aprovação para solicitante
// telefoneSolicitante: número no formato 55XXXXXXXXXX
// ongPhone: número da ONG no formato 55XXXXXXXXXX
func SendWhatsAppAdoptionApproved(telefoneSolicitante, solicitanteName, petName, ongName, ongPhone string) {
	svc := GetWhatsAppService()
	if svc == nil || !svc.enabled {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ Panic ao enviar WhatsApp (aprovação para solicitante): %v\n", r)
			}
		}()

		message := NewAdoptionApprovedWhatsAppMessage(solicitanteName, petName, ongName, ongPhone)
		start := time.Now()

		err := svc.SendMessageWithRetry(telefoneSolicitante, message)

		duration := time.Since(start).Milliseconds()
		if err != nil {
			log.Printf("❌ Falha ao enviar WhatsApp (aprovação → solicitante) para %s em %dms: %v\n",
				maskPhoneNumber(telefoneSolicitante), duration, err)
		} else {
			log.Printf("✉️ WhatsApp enviado (aprovação → solicitante) para %s em %dms\n",
				maskPhoneNumber(telefoneSolicitante), duration)
		}
	}()
}

// SendWhatsAppAdoptionRejected envia rejeição para solicitante
// telefoneSolicitante: número no formato 55XXXXXXXXXX
func SendWhatsAppAdoptionRejected(telefoneSolicitante, solicitanteName, petName, ongName string) {
	svc := GetWhatsAppService()
	if svc == nil || !svc.enabled {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ Panic ao enviar WhatsApp (rejeição para solicitante): %v\n", r)
			}
		}()

		message := NewAdoptionRejectedWhatsAppMessage(solicitanteName, petName, ongName)
		start := time.Now()

		err := svc.SendMessageWithRetry(telefoneSolicitante, message)

		duration := time.Since(start).Milliseconds()
		if err != nil {
			log.Printf("❌ Falha ao enviar WhatsApp (rejeição → solicitante) para %s em %dms: %v\n",
				maskPhoneNumber(telefoneSolicitante), duration, err)
		} else {
			log.Printf("✉️ WhatsApp enviado (rejeição → solicitante) para %s em %dms\n",
				maskPhoneNumber(telefoneSolicitante), duration)
		}
	}()
}

// SendWhatsAppPetRegistered envia notificação de novo pet registrado
func SendWhatsAppPetRegistered(telefoneOwner, petName, ownerName string) {
	svc := GetWhatsAppService()
	if svc == nil || !svc.enabled {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ Panic ao enviar WhatsApp (novo pet): %v\n", r)
			}
		}()

		message := NewPetRegisteredWhatsAppMessage(petName, ownerName)
		start := time.Now()

		err := svc.SendMessageWithRetry(telefoneOwner, message)

		duration := time.Since(start).Milliseconds()
		if err != nil {
			log.Printf("❌ Falha ao enviar WhatsApp (novo pet) para %s em %dms: %v\n",
				maskPhoneNumber(telefoneOwner), duration, err)
		} else {
			log.Printf("✉️ WhatsApp enviado (novo pet) para %s em %dms\n",
				maskPhoneNumber(telefoneOwner), duration)
		}
	}()
}

// ValidatePhoneNumber faz validação básica de número de telefone
// Esperado: 55XXXXXXXXXX (dígitos apenas)
func ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return fmt.Errorf("número de telefone vazio")
	}

	// Remover caracteres especiais e contar dígitos
	digitCount := 0
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digitCount++
		}
	}

	// Brasil: mínimo 11 dígitos (55 + DDD + número), máximo 15 (padrão internacional)
	if digitCount < 11 || digitCount > 15 {
		return fmt.Errorf("número de telefone inválido: esperado entre 11-15 dígitos, obtive %d", digitCount)
	}

	return nil
}
