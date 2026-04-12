package utils

import "fmt"

// NewAcompanhamentoReminderEmail gera o email de lembrete enviado 1 dia antes da próxima data de acompanhamento.
// Destinatário: ONG responsável.
func NewAcompanhamentoReminderEmail(
	petNome string,
	adotanteNome string,
	adotanteTelefone string,
	adotanteEmail string,
	proximaData string, // formato: "02/01/2006"
	frequencia string,
) (subject, body string) {
	subject = fmt.Sprintf("🐾 Lembrete de Acompanhamento: %s - amanhã (%s)", petNome, proximaData)

	emailSection := ""
	if adotanteEmail != "" {
		emailSection = fmt.Sprintf(`<p><strong>Email do Adotante:</strong> %s</p>`, adotanteEmail)
	}

	body = fmt.Sprintf(`
<!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
  <div style="background-color: #f9a825; padding: 20px; border-radius: 8px 8px 0 0; text-align: center;">
    <h1 style="color: white; margin: 0;">🐾 Lembrete de Acompanhamento</h1>
  </div>
  <div style="background-color: #fff; padding: 30px; border: 1px solid #e0e0e0; border-top: none; border-radius: 0 0 8px 8px;">
    <p>Olá! Este é um lembrete automático do sistema <strong>OngPet</strong>.</p>
    <p>Amanhã, <strong>%s</strong>, está agendado um acompanhamento pós-adoção:</p>

    <div style="background-color: #f5f5f5; border-left: 4px solid #f9a825; padding: 16px; margin: 20px 0; border-radius: 4px;">
      <h3 style="margin-top: 0; color: #f9a825;">🐶 Pet</h3>
      <p><strong>Nome do Pet:</strong> %s</p>

      <h3 style="color: #f9a825;">👤 Adotante</h3>
      <p><strong>Nome:</strong> %s</p>
      <p><strong>Telefone:</strong> %s</p>
      %s

      <h3 style="color: #f9a825;">📅 Acompanhamento</h3>
      <p><strong>Data Agendada:</strong> %s</p>
      <p><strong>Frequência:</strong> %s</p>
    </div>

    <p>Entre em contato com o adotante para verificar como está o bem-estar do pet.</p>
    <p>Acesse o painel da instituição para registrar o log do acompanhamento após o contato.</p>

    <br>
    <p style="color: #999; font-size: 12px;">
      Atenciosamente,<br>
      Sistema OngPet 🐾<br>
      <em>Este é um email automático, por favor não responda.</em>
    </p>
  </div>
</body>
</html>
`, proximaData, petNome, adotanteNome, adotanteTelefone, emailSection, proximaData, frequencia)

	return
}