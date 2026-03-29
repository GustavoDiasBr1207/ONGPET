package utils

import "fmt"

func NewPetRegisteredEmail(petName, ownerName string) (subject, body string) {
    subject = fmt.Sprintf("🐾 Pet registered: %s", petName)
    body = fmt.Sprintf(`
        <h2>Hi %s!</h2>
        <p>Your pet <strong>%s</strong> was successfully registered.</p>
    `, ownerName, petName)
    return
}

func NewAdoptionRequestEmail(petName, requesterName, ongName string) (subject, body string) {
    subject = fmt.Sprintf("🐾 Nova solicitação de adoção: %s", petName)
    body = fmt.Sprintf(`
        <h2>Nova Solicitação de Adoção</h2>
        <p>Olá <strong>%s</strong>,</p>
        <p>Você recebeu uma nova solicitação de adoção para o pet <strong>%s</strong>.</p>
        <p><strong>Solicitante:</strong> %s</p>
        <p>Acesse o painel da ONG para visualizar os detalhes da solicitação.</p>
        <br>
        <p>Atenciosamente,<br>Sistema OngPet</p>
    `, ongName, petName, requesterName)
    return
}

func NewAdoptionConfirmedEmail(petName, requesterName string) (subject, body string) {
    subject = fmt.Sprintf("🎉 Adoção aprovada: %s", petName)
    body = fmt.Sprintf(`
        <h2>Parabéns! Sua adoção foi aprovada!</h2>
        <p>Olá <strong>%s</strong>,</p>
        <p>Sua solicitação de adoção para o pet <strong>%s</strong> foi aprovada!</p>
        <p>A ONG entrará em contato com você em breve para os próximos passos.</p>
        <br>
        <p>Obrigado por escolher adotar!<br>Sistema OngPet</p>
    `, requesterName, petName)
    return
}

func NewContactEmail(name, email, message string) (subject, body string) {
    subject = fmt.Sprintf("📧 Novo contato: %s", name)
    body = fmt.Sprintf(`
        <h2>Novo Contato</h2>
        <p><strong>Nome:</strong> %s</p>
        <p><strong>Email:</strong> %s</p>
        <p><strong>Mensagem:</strong></p>
        <p>%s</p>
    `, name, email, message)
    return
}