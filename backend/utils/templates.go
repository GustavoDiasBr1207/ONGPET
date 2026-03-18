package mailer

import "fmt"

func NewPetRegisteredEmail(petName, ownerName string) (subject, body string) {
    subject = fmt.Sprintf("🐾 Pet registered: %s", petName)
    body = fmt.Sprintf(`
        <h2>Hi %s!</h2>
        <p>Your pet <strong>%s</strong> was successfully registered.</p>
    `, ownerName, petName)
    return
}

// Adicione outros templates conforme precisar