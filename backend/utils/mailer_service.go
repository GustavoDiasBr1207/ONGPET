package utils

import (
	"log"
	"sync"
)

var (
	mailerInstance *Mailer
	mailerOnce     sync.Once
)

// InitMailer inicializa o mailer global
func InitMailer() error {
	var err error
	mailerOnce.Do(func() {
		m, initErr := New()
		if initErr != nil {
			err = initErr
			log.Println("⚠️ Falha ao inicializar mailer:", initErr)
			return
		}
		mailerInstance = m
		log.Println("✅ Mailer inicializado com sucesso")
	})
	return err
}

// GetMailer retorna a instância global do mailer
func GetMailer() *Mailer {
	if mailerInstance == nil {
		log.Println("⚠️ Mailer não inicializado, tentando inicializar...")
		_ = InitMailer()
	}
	return mailerInstance
}

// SendEmailPetRegistered envia email quando um pet é registrado
func SendEmailPetRegistered(to []string, petName, ownerName string) error {
	mailer := GetMailer()
	if mailer == nil {
		log.Println("⚠️ Mailer indisponível, email não enviado")
		return nil // Não falha a operação principal
	}

	subject, body := NewPetRegisteredEmail(petName, ownerName)
	return mailer.Send(to, subject, body)
}

// SendEmailAdoptionRequest envia email sobre nova solicitação de adoção
func SendEmailAdoptionRequest(to []string, petName, requesterName, ongName string) error {
	mailer := GetMailer()
	if mailer == nil {
		log.Println("⚠️ Mailer indisponível, email não enviado")
		return nil
	}

	subject, body := NewAdoptionRequestEmail(petName, requesterName, ongName)
	return mailer.Send(to, subject, body)
}

// SendEmailAdoptionConfirmed envia email confirmando adoção aprovada
func SendEmailAdoptionConfirmed(to []string, petName, requesterName string) error {
	mailer := GetMailer()
	if mailer == nil {
		log.Println("⚠️ Mailer indisponível, email não enviado")
		return nil
	}

	subject, body := NewAdoptionConfirmedEmail(petName, requesterName)
	return mailer.Send(to, subject, body)
}

// SendEmailContact envia email de contato
func SendEmailContact(to []string, name, email, message string) error {
	mailer := GetMailer()
	if mailer == nil {
		log.Println("⚠️ Mailer indisponível, email não enviado")
		return nil
	}

	subject, body := NewContactEmail(name, email, message)
	return mailer.Send(to, subject, body)
}

// SendEmailAdoptionApproved envia email para o solicitante quando a adoção é aprovada
func SendEmailAdoptionApproved(to []string, solicitanteName, petName, ongName, telefoneOng string) error {
	mailer := GetMailer()
	if mailer == nil {
		log.Println("⚠️ Mailer indisponível, email não enviado")
		return nil
	}

	subject, body := NewAdoptionApprovedEmail(solicitanteName, petName, ongName, telefoneOng)
	return mailer.Send(to, subject, body)
}

// SendEmailAdoptionRejected envia email para o solicitante quando a adoção é rejeitada
func SendEmailAdoptionRejected(to []string, solicitanteName, petName, ongName, motivo string) error {
	mailer := GetMailer()
	if mailer == nil {
		log.Println("⚠️ Mailer indisponível, email não enviado")
		return nil
	}

	subject, body := NewAdoptionRejectedEmail(solicitanteName, petName, ongName, motivo)
	return mailer.Send(to, subject, body)
}
