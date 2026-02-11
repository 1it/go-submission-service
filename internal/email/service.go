package email

import (
	"fmt"
	"log"

	"github.com/1it/go-submission-service/internal/config"
	"github.com/1it/go-submission-service/internal/templates"
)

// Service provides email sending functionality with template support
type Service struct {
	config          *config.Config
	templateManager *templates.Manager
	sender          EmailSender
}

// NewService creates a new email service
func NewService(cfg *config.Config) (*Service, error) {
	// Create template manager
	templateManager := templates.NewManager(cfg)
	if err := templateManager.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize template manager: %w", err)
	}

	// Create email sender based on configuration
	var sender EmailSender
	if cfg.Email.Provider == "mailersend" {
		if cfg.Email.MailerSend.APIKey == "" {
			return nil, fmt.Errorf("MailerSend API key is required when using MailerSend provider")
		}
		mailerClient := NewMailerSendClient(cfg.Email.MailerSend.APIKey)
		mailerClient.SetSender(cfg.Email.MailerSend.FromEmail, cfg.Email.MailerSend.FromName)
		sender = mailerClient
	} else {
		// Default to SMTP
		sender = NewSMTPClient(SMTPConfig{
			Host:     cfg.Email.SMTP.Host,
			Port:     cfg.Email.SMTP.Port,
			Username: cfg.Email.SMTP.Username,
			Password: cfg.Email.SMTP.Password,
			From:     cfg.Email.SMTP.From,
		})
	}

	return &Service{
		config:          cfg,
		templateManager: templateManager,
		sender:          sender,
	}, nil
}

// SendTemplatedEmail sends an email using a template
func (s *Service) SendTemplatedEmail(recipientEmail, templateName string, data map[string]interface{}) error {
	// Use default template if none specified
	if templateName == "" {
		templateName = s.config.Email.Templates.DefaultTemplate
	}

	// Render the template
	log.Printf("Rendering email template '%s' for recipient: %s", templateName, recipientEmail)
	htmlContent, err := s.templateManager.RenderEmail(templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Get subject from config or data
	subject := s.config.Email.Templates.Subject
	if subjectFromData, ok := data["subject"].(string); ok && subjectFromData != "" {
		subject = subjectFromData
	}

	// Send the email
	log.Printf("Sending email to %s with subject '%s'", recipientEmail, subject)
	return s.sender.SendEmail(recipientEmail, subject, htmlContent)
}

// GetAvailableTemplates returns a list of available template names
func (s *Service) GetAvailableTemplates() []string {
	return s.templateManager.ListTemplates()
}
