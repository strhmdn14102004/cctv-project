package services

import (
	"cctv-api/internal/config"
	"fmt"
	"net/smtp"
	"strings"
)

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

func (es *EmailService) SendResetEmail(to, token string) error {
	// Email content
	subject := "Device Reset Request"
	body := fmt.Sprintf(`
	<html>
	<body>
		<h2>Device Reset Request</h2>
		<p>You have requested to reset your device authorization. To complete this process, please use the following token:</p>
		<p><strong>%s</strong></p>
		<p>This token will expire in 24 hours.</p>
		<p>If you didn't request this, please ignore this email.</p>
	</body>
	</html>
	`, token)

	// SMTP auth
	auth := smtp.PlainAuth("", es.cfg.SMTPUsername, es.cfg.SMTPPassword, es.cfg.SMTPHost)

	// Email headers
	headers := make(map[string]string)
	headers["From"] = es.cfg.EmailFrom
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""

	// Build message
	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n" + body)

	// Send email
	err := smtp.SendMail(
		es.cfg.SMTPHost+":"+es.cfg.SMTPPort,
		auth,
		es.cfg.EmailFrom,
		[]string{to},
		[]byte(message.String()),
	)

	return err
}
