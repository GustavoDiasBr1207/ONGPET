package utils

import "fmt"

// AdoptionRequestFullData contém dados completos da solicitação
type AdoptionRequestFullData struct {
	PetNome             string
	PetEspecie          string
	PetRaca             string
	PetIdade            int
	PetDescricao        string
	PetPeso             float64
	PetPorte            string
	PetRegiao           string
	OngNome             string
	SolicitanteName     string
	RespostasFormulario map[string]string // Campo => Valor
}

// NewAdoptionRequestFullEmail cria email com todas as informações de adoção
func NewAdoptionRequestFullEmail(data AdoptionRequestFullData) (subject, body string) {
	subject = fmt.Sprintf("🐾 Nova Solicitação de Adoção: %s - %s", data.PetNome, data.SolicitanteName)

	// Monta a seção de respostas do formulário
	respostasHTML := ""
	if len(data.RespostasFormulario) > 0 {
		respostasHTML = `
		<hr>
		<h3>📋 Informações do Solicitante (Formulário)</h3>
		<table border="1" cellpadding="10" style="border-collapse: collapse; width: 100%;">
			<thead style="background-color: #f0f0f0;">
				<tr>
					<th>Campo</th>
					<th>Resposta</th>
				</tr>
			</thead>
			<tbody>`

		for campo, resposta := range data.RespostasFormulario {
			respostasHTML += fmt.Sprintf(`
				<tr>
					<td><strong>%s</strong></td>
					<td>%s</td>
				</tr>`, campo, resposta)
		}

		respostasHTML += `
			</tbody>
		</table>`
	}

	body = fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; color: #333; }
				.container { max-width: 700px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; border-radius: 5px; }
				.section { margin: 20px 0; }
				table { border-collapse: collapse; width: 100%%; margin: 10px 0; }
				th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
				th { background-color: #f0f0f0; font-weight: bold; }
				tr:nth-child(even) { background-color: #f9f9f9; }
				hr { border: 0; border-top: 2px solid #ddd; margin: 20px 0; }
				.footer { text-align: center; color: #666; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; }
				.action-box { background-color: #e8f5e9; border-left: 4px solid #4CAF50; padding: 15px; margin: 15px 0; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>🐾 Nova Solicitação de Adoção</h1>
				</div>

				<div class="section">
					<p>Olá <strong>%s</strong>,</p>
					<p>Você recebeu uma nova solicitação de adoção para o seguinte pet:</p>
				</div>

				<div class="section">
					<h3>🐾 Informações do Pet</h3>
					<table border="1" cellpadding="10">
						<tr><td width="35%%"><strong>Nome</strong></td><td>%s</td></tr>
						<tr><td><strong>Espécie</strong></td><td>%s</td></tr>
						<tr><td><strong>Raça</strong></td><td>%s</td></tr>
						<tr><td><strong>Idade</strong></td><td>%d ano(s)</td></tr>
						<tr><td><strong>Peso</strong></td><td>%.1f kg</td></tr>
						<tr><td><strong>Porte</strong></td><td>%s</td></tr>
						<tr><td><strong>Região</strong></td><td>%s</td></tr>
						<tr><td><strong>Descrição</strong></td><td>%s</td></tr>
					</table>
				</div>

				<div class="section">
					<h3>👤 Informações do Solicitante</h3>
					<p><strong>Nome:</strong> %s</p>
				</div>

				%s

				<div class="action-box">
					<strong>✅ Próximo Passo:</strong> Faça login no painel da ONG para revisar a solicitação completa, incluindo documentos anexados e formulário respondido pelo solicitante.
				</div>

				<div class="footer">
					<p><strong>Sistema OngPet</strong> - Plataforma de Adoção de Animais</p>
					<p>Para mais informações, acesse o painel da ONG</p>
				</div>
			</div>
		</body>
		</html>
	`,
		data.OngNome,
		data.PetNome,
		data.PetEspecie,
		data.PetRaca,
		data.PetIdade,
		data.PetPeso,
		data.PetPorte,
		data.PetRegiao,
		data.PetDescricao,
		data.SolicitanteName,
		respostasHTML,
	)
	return
}
