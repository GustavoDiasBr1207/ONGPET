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

// NewAdoptionApprovedEmail envia email para o solicitante quando a adoção é aprovada
func NewAdoptionApprovedEmail(solicitanteName, petName, ongName, telefoneOng string) (subject, body string) {
	subject = fmt.Sprintf("🎉 Parabéns! Sua adoção de %s foi aprovada!", petName)
	body = fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; color: #333; }
				.container { max-width: 700px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; border-radius: 5px; }
				.section { margin: 20px 0; }
				.footer { text-align: center; color: #666; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; }
				.action-box { background-color: #c8e6c9; border-left: 4px solid #4CAF50; padding: 15px; margin: 15px 0; border-radius: 3px; }
				.highlight { font-size: 18px; font-weight: bold; color: #2e7d32; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>🎉 Parabéns!</h1>
				</div>

				<div class="section">
					<p>Olá <strong>%s</strong>,</p>
					<p>Temos o prazer em informar que sua solicitação de adoção foi <span class="highlight">APROVADA</span>! 🐾</p>
				</div>

				<div class="section">
					<h3>✅ Sua Adoção foi Aprovada</h3>
					<p>Congratulações! Sua solicitação para adotar <strong>%s</strong> foi avaliada e aprovada pela equipe da <strong>%s</strong>.</p>
				</div>

				<div class="action-box">
					<p><strong>Próximos Passos:</strong></p>
					<p>A ONG entrará em contato com você em breve para agendar a entrega do seu novo companheiro! 🐾</p>
					<p><strong>Telefone para contato:</strong> %s</p>
				</div>

				<div class="section">
					<p>Obrigado por escolher adotar e dar um lar amoroso para um animal!</p>
					<p>Qualquer dúvida, entre em contato com a ONG.</p>
				</div>

				<div class="footer">
					<p><strong>Sistema OngPet</strong> - Plataforma de Adoção de Animais</p>
					<p>Aqui você muda vidas!</p>
				</div>
			</div>
		</body>
		</html>
	`, solicitanteName, petName, ongName, telefoneOng)
	return
}

// NewAdoptionRejectedEmail envia email para o solicitante quando a adoção é rejeitada
func NewAdoptionRejectedEmail(solicitanteName, petName, ongName, motivo string) (subject, body string) {
	subject = fmt.Sprintf("Atualização sobre sua solicitação de adoção de %s", petName)
	body = fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; color: #333; }
				.container { max-width: 700px; margin: 0 auto; padding: 20px; }
				.header { background-color: #ff9800; color: white; padding: 20px; text-align: center; border-radius: 5px; }
				.section { margin: 20px 0; }
				.footer { text-align: center; color: #666; font-size: 12px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; }
				.info-box { background-color: #fff3cd; border-left: 4px solid #ff9800; padding: 15px; margin: 15px 0; border-radius: 3px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>⚠️ Atualização Importante</h1>
				</div>

				<div class="section">
					<p>Olá <strong>%s</strong>,</p>
					<p>Agradecemos sinceramente seu interesse em adotar <strong>%s</strong> através da <strong>%s</strong>.</p>
				</div>

				<div class="info-box">
					<p><strong>Infelizmente,</strong> sua solicitação de adoção foi avaliada e <strong>não foi aprovada</strong> neste momento.</p>
					<p><strong>Motivo:</strong> %s</p>
				</div>

				<div class="section">
					<p>Não desista! Existem muitos outros animais incríveis esperando por um lar amoroso. 🐾</p>
					<p>Você pode:</p>
					<ul>
						<li>Explorar outros pets disponíveis para adoção</li>
						<li>Melhorar algumas características mencionadas e solicitar novamente</li>
						<li>Entrar em contato com a ONG para mais informações</li>
					</ul>
				</div>

				<div class="section">
					<p>Continuamos à disposição para ajudar você a encontrar seu companheiro perfeito!</p>
				</div>

				<div class="footer">
					<p><strong>Sistema OngPet</strong> - Plataforma de Adoção de Animais</p>
					<p>Aqui você muda vidas!</p>
				</div>
			</div>
		</body>
		</html>
	`, solicitanteName, petName, ongName, motivo)
	return
}
