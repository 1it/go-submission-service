package email

// EmailSender defines an interface for sending emails, allowing for different implementations.
type EmailSender interface {
	SendEmail(recipientEmail string, subject string, htmlContent string) error
}
