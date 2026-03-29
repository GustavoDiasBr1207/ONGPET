package utils

// EmailTestData contém dados para testar envio de email
type EmailTestData struct {
	To       string
	Template string // "adoption_request", "pet_registered", "adoption_confirmed", "contact"
	Data     map[string]string
}

// SendTestEmail envia um email de teste
func SendTestEmail(testData EmailTestData) error {
	mailer := GetMailer()
	if mailer == nil {
		return nil // Falha silenciosa se mailer não estiver disponível
	}

	switch testData.Template {
	case "pet_registered":
		petName := testData.Data["pet_name"]
		ownerName := testData.Data["owner_name"]
		subject, body := NewPetRegisteredEmail(petName, ownerName)
		return mailer.Send([]string{testData.To}, subject, body)

	case "adoption_request":
		petName := testData.Data["pet_name"]
		requesterName := testData.Data["requester_name"]
		ongName := testData.Data["ong_name"]
		subject, body := NewAdoptionRequestEmail(petName, requesterName, ongName)
		return mailer.Send([]string{testData.To}, subject, body)

	case "adoption_confirmed":
		petName := testData.Data["pet_name"]
		requesterName := testData.Data["requester_name"]
		subject, body := NewAdoptionConfirmedEmail(petName, requesterName)
		return mailer.Send([]string{testData.To}, subject, body)

	case "contact":
		name := testData.Data["name"]
		email := testData.Data["email"]
		message := testData.Data["message"]
		subject, body := NewContactEmail(name, email, message)
		return mailer.Send([]string{testData.To}, subject, body)

	default:
		// Envio genérico
		subject := testData.Data["subject"]
		body := testData.Data["body"]
		return mailer.Send([]string{testData.To}, subject, body)
	}
}
