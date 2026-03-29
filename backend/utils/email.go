package utils

import (
    "fmt"
    "os"
    "strconv"

    gomail "gopkg.in/gomail.v2"
)

type Mailer struct {
    dialer *gomail.Dialer
    from   string
}

func New() (*Mailer, error) {
    host := os.Getenv("SMTP_HOST")
    portStr := os.Getenv("SMTP_PORT")
    user := os.Getenv("SMTP_USER")
    pass := os.Getenv("SMTP_PASS")
    from := os.Getenv("SMTP_FROM")

    port, err := strconv.Atoi(portStr)
    if err != nil {
        return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
    }

    d := gomail.NewDialer(host, port, user, pass)
    return &Mailer{dialer: d, from: from}, nil
}

func (m *Mailer) Send(to []string, subject, body string) error {
    msg := gomail.NewMessage()
    msg.SetHeader("From", m.from)
    msg.SetHeader("To", to...)
    msg.SetHeader("Subject", subject)
    msg.SetBody("text/html", body)

    return m.dialer.DialAndSend(msg)
}