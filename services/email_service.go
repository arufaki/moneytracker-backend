package services

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/wneessen/go-mail"
)

type EmailService interface {
	SendVerificationEmail(to, name, token string) error
}

type emailService struct{}

func NewEmailService() EmailService {
	return &emailService{}
}

func (s *emailService) SendVerificationEmail(to, name, token string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	verifyLink := fmt.Sprintf("%s/auth/verify?token=%s", frontendURL, token)

	if smtpHost == "" {
		log.Printf("[EMAIL MOCK] Verification email for %s (%s): %s", name, to, verifyLink)
		return nil
	}

	smtpPort := 587
	if p, err := strconv.Atoi(smtpPortStr); err == nil && p > 0 {
		smtpPort = p
	}

	m := mail.NewMsg()
	if smtpFrom != "" {
		if err := m.From(smtpFrom); err != nil {
			return err
		}
	} else {
		if err := m.From(smtpUser); err != nil {
			return err
		}
	}
	if err := m.To(to); err != nil {
		return err
	}
	m.Subject("Verifikasi Email - Money Tracker")
	m.SetBodyString(mail.TypeTextPlain, fmt.Sprintf("Halo %s,\n\nSilakan klik link berikut untuk verifikasi akun Anda:\n%s\n\nTerima kasih!", name, verifyLink))

	opts := []mail.Option{
		mail.WithPort(smtpPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(smtpUser),
		mail.WithPassword(smtpPass),
	}

	client, err := mail.NewClient(smtpHost, opts...)
	if err != nil {
		return err
	}

	return client.DialAndSend(m)
}
