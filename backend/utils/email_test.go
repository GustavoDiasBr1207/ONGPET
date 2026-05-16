package utils

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// EmailTestData contém dados para testar envio de email
type EmailTestData struct {
	To       string
	Template string // "adoption_request", "pet_registered", "adoption_confirmed", "contact"
	Data     map[string]string
}

// SendTestEmail envia um email de teste usando um template
func SendTestEmail(testData EmailTestData) error {
	mailer := GetMailer()
	if mailer == nil {
		return nil
	}

	switch testData.Template {
	case "pet_registered":
		subject, body := NewPetRegisteredEmail(testData.Data["pet_name"], testData.Data["owner_name"])
		return mailer.Send([]string{testData.To}, subject, body)
	case "adoption_request":
		subject, body := NewAdoptionRequestEmail(testData.Data["pet_name"], testData.Data["requester_name"], testData.Data["ong_name"])
		return mailer.Send([]string{testData.To}, subject, body)
	case "adoption_confirmed":
		subject, body := NewAdoptionConfirmedEmail(testData.Data["pet_name"], testData.Data["requester_name"])
		return mailer.Send([]string{testData.To}, subject, body)
	case "contact":
		subject, body := NewContactEmail(testData.Data["name"], testData.Data["email"], testData.Data["message"])
		return mailer.Send([]string{testData.To}, subject, body)
	default:
		return mailer.Send([]string{testData.To}, testData.Data["subject"], testData.Data["body"])
	}
}

// TestMailerInit verifica se o mailer inicializa corretamente com as variáveis de ambiente.
func TestMailerInit(t *testing.T) {
	_ = godotenv.Load("../.env")

	if err := InitMailer(); err != nil {
		t.Fatalf("falha ao inicializar mailer: %v", err)
	}

	if GetMailer() == nil {
		t.Fatal("GetMailer() retornou nil após inicialização bem-sucedida")
	}

	t.Log("✅ Mailer inicializado com sucesso")
}

// TestSendRealEmail envia um email de verdade para verificar o fluxo completo.
// Usa TEST_EMAIL_TO se definido, caso contrário usa SMTP_FROM do .env.
func TestSendRealEmail(t *testing.T) {
	_ = godotenv.Load("../.env")

	to := os.Getenv("TEST_EMAIL_TO")
	if to == "" {
		to = os.Getenv("SMTP_FROM")
	}
	if to == "" {
		t.Skip("SMTP_FROM e TEST_EMAIL_TO não definidos — configure o .env para executar")
	}

	if err := InitMailer(); err != nil {
		t.Fatalf("falha ao inicializar mailer: %v", err)
	}

	m := GetMailer()
	if m == nil {
		t.Fatal("mailer é nil após inicialização")
	}

	subject, body := NewPetRegisteredEmail("Rex (teste)", "Equipe OngPet")
	if err := m.Send([]string{to}, subject, body); err != nil {
		t.Fatalf("❌ falha ao enviar email: %v", err)
	}

	t.Logf("✅ Email enviado com sucesso para: %s", to)
}
