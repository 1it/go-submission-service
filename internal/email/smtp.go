package email

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// SMTPConfig holds configuration details for the SMTP server.
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string // Sender email address
}

// SMTPClient is a client for sending emails via SMTP.
type SMTPClient struct {
	Config SMTPConfig
}

// NewSMTPClient creates a new SMTP client.
func NewSMTPClient(config SMTPConfig) *SMTPClient {
	log.Printf("Initializing SMTP client with host: %s, port: %s, from: %s",
		config.Host, config.Port, config.From)

	// Don't log the password
	hasAuth := config.Username != "" && config.Password != ""
	log.Printf("SMTP authentication enabled: %v", hasAuth)

	return &SMTPClient{
		Config: config,
	}
}

// SendEmail sends an email using SMTP.
func (s *SMTPClient) SendEmail(recipientEmail string, subject string, htmlContent string) error {
	log.Printf("Preparing to send email to: %s", recipientEmail)

	// Set up authentication if credentials provided
	var auth smtp.Auth
	if s.Config.Username != "" && s.Config.Password != "" {
		log.Printf("Setting up SMTP authentication for %s@%s", s.Config.Username, s.Config.Host)
		auth = smtp.PlainAuth("", s.Config.Username, s.Config.Password, s.Config.Host)
	} else {
		log.Printf("No SMTP authentication will be used (anonymous)")
	}

	// Set up email headers
	log.Println("Building email headers")
	headers := make(map[string]string)
	headers["From"] = s.Config.From
	headers["To"] = recipientEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Construct message
	log.Println("Constructing email content")
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n") // End of headers
	msg.WriteString(htmlContent)

	contentLength := len(msg.String())
	log.Printf("Email prepared, content length: %d bytes", contentLength)

	// Send email
	addr := s.Config.Host + ":" + s.Config.Port
	to := []string{recipientEmail}

	log.Printf("Sending email via SMTP to %s using server %s", recipientEmail, addr)
	err := smtp.SendMail(addr, auth, s.Config.From, to, []byte(msg.String()))
	if err != nil {
		log.Printf("ERROR: Failed to send email to %s: %v", recipientEmail, err)
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	log.Printf("Email sent successfully to %s", recipientEmail)
	return nil
}
