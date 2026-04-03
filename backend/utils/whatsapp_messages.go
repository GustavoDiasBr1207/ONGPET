package utils

import "fmt"

// WhatsApp Message Templates
// Mensagens otimizadas para WhatsApp (texto simples, sem HTML)

// NewAdoptionRequestWhatsAppMessage cria mensagem para ONG sobre novo pedido
// Formato: número + DDD, ex: 5527992345678
func NewAdoptionRequestWhatsAppMessage(petName, solicitanteName, ongName string) string {
	return fmt.Sprintf(`🐾 *Nova Solicitação de Adoção!*

Olá *%s*,

Você recebeu uma nova solicitação de adoção para:
🐶 *%s*

👤 *Solicitante:* %s

Acesse o painel da ONG ou clique aqui para revisar os detalhes completos da solicitação.

---
Sistema OngPet 🐾
`, ongName, petName, solicitanteName)
}

// NewAdoptionConfirmationWhatsAppMessage envia confirmação para solicitante que pedido foi recebido
func NewAdoptionConfirmationWhatsAppMessage(solicitanteName, petName, ongName string) string {
	return fmt.Sprintf(`🎉 *Solicitação Recebida!*

Olá *%s*,

Sua solicitação de adoção para *%s* foi recebida com sucesso! ✅

A *%s* está analisando seu pedido. Você receberá uma mensagem assim que houver uma resposta.

Obrigado por nos escolher! 🐾

---
Sistema OngPet
`, solicitanteName, petName, ongName)
}

// NewAdoptionApprovedWhatsAppMessage envia confirmação de aprovação para solicitante
// Inclui telefone da ONG para contato
func NewAdoptionApprovedWhatsAppMessage(solicitanteName, petName, ongName, ongPhone string) string {
	return fmt.Sprintf(`🎉 *PARABÉNS! Sua Adoção foi Aprovada!* 🎉

Olá *%s*,

Excelente notícia! Sua solicitação de adoção para *%s* foi *APROVADA*! ✅

*Próximos Passos:*
📞 Entre em contato com a ONG para combinar os detalhes finais:
📱 *%s* (WhatsApp)

Obrigado por escolher adotar! 🐾❤️

---
Sistema OngPet
`, solicitanteName, petName, ongPhone)
}

// NewAdoptionRejectedWhatsAppMessage envia rejeição para solicitante
func NewAdoptionRejectedWhatsAppMessage(solicitanteName, petName, ongName string) string {
	return fmt.Sprintf(`😔 *Solicitação não aprovada*

Olá *%s*,

Infelizmente, sua solicitação de adoção para *%s* não foi aprovada neste momento. 😢

A *%s* agradece seu interesse e incentiva você a explorar outros animais disponíveis para adoção.

Para mais informações ou questões, entre em contato diretamente com a instituição.

---
Sistema OngPet
`, solicitanteName, petName, ongName)
}

// NewPetRegisteredWhatsAppMessage envia notificação de novo pet registrado
func NewPetRegisteredWhatsAppMessage(petName, ownerName string) string {
	return fmt.Sprintf(`🐾 *Novo Pet Registrado!*

Olá *%s*,

O pet *%s* foi registrado com sucesso no sistema OngPet! ✅

Agora outras pessoas podem fazer solicitações de adoção deste animal.

---
Sistema OngPet
`, ownerName, petName)
}
