package utils

import "fmt"

// NewAcompanhamentoReminderWhatsAppMessage gera a mensagem de WhatsApp de lembrete
// enviada 1 dia antes da próxima data de acompanhamento.
func NewAcompanhamentoReminderWhatsAppMessage(
	petNome string,
	adotanteNome string,
	adotanteTelefone string,
	adotanteEmail string,
	proximaData string, // formato: "02/01/2006"
	frequencia string,
) string {
	emailLine := ""
	if adotanteEmail != "" {
		emailLine = fmt.Sprintf("\n📧 *Email:* %s", adotanteEmail)
	}

	return fmt.Sprintf(
		"🐾 *Lembrete de Acompanhamento - OngPet*\n\n"+
			"Olá! Amanhã, *%s*, está agendado um acompanhamento pós-adoção.\n\n"+
			"🐶 *Pet:* %s\n\n"+
			"👤 *Adotante*\n"+
			"• Nome: *%s*\n"+
			"• Telefone: *%s*%s\n\n"+
			"📅 *Frequência:* %s\n\n"+
			"Entre em contato com o adotante para verificar o bem-estar do pet e não esqueça de registrar o log no painel da instituição após o contato. 🏠",
		proximaData,
		petNome,
		adotanteNome,
		adotanteTelefone,
		emailLine,
		frequencia,
	)
}

// SendWhatsAppAcompanhamentoReminder envia WhatsApp de lembrete para a ONG
// 1 dia antes da próxima data de acompanhamento.
// telefoneOng: número no formato 55XXXXXXXXXX (somente dígitos)
func SendWhatsAppAcompanhamentoReminder(
	telefoneOng string,
	petNome string,
	adotanteNome string,
	adotanteTelefone string,
	adotanteEmail string,
	proximaData string,
	frequencia string,
) {
	svc := GetWhatsAppService()
	if svc == nil || !svc.enabled {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// log silencioso para não derrubar o scheduler
				_ = r
			}
		}()

		message := NewAcompanhamentoReminderWhatsAppMessage(
			petNome,
			adotanteNome,
			adotanteTelefone,
			adotanteEmail,
			proximaData,
			frequencia,
		)

		svc.SendMessageWithRetry(telefoneOng, message)
	}()
}