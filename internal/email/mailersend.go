package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type MailerSendClient struct {
	APIKey      string
	BaseURL     string
	FromEmail   string
	FromName    string
	SubjectLine string
}

type EmailRequest struct {
	From    EmailAddress   `json:"from"`
	To      []EmailAddress `json:"to"`
	Subject string         `json:"subject"`
	HTML    string         `json:"html"`
}

type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

func NewMailerSendClient(apiKey string) *MailerSendClient {
	return &MailerSendClient{
		APIKey:      apiKey,
		BaseURL:     "https://api.mailersend.com/v1",
		FromEmail:   "noreply@example.com",
		FromName:    "Form Submission Service",
		SubjectLine: "Form Submission Received",
	}
}

// SetSender sets the from email address and name
func (m *MailerSendClient) SetSender(email, name string) {
	if email != "" {
		m.FromEmail = email
	}
	if name != "" {
		m.FromName = name
	}
}

// SetSubject sets the email subject line
func (m *MailerSendClient) SetSubject(subject string) {
	if subject != "" {
		m.SubjectLine = subject
	}
}

// SendEmail sends an email using the MailerSend API.
func (m *MailerSendClient) SendEmail(recipientEmail string, subject string, htmlContent string) error {
	endpoint := m.BaseURL + "/email"

	log.Printf("Preparing to send email to: %s", recipientEmail)

	emailReq := EmailRequest{
		From: EmailAddress{
			Email: m.FromEmail,
			Name:  m.FromName,
		},
		To: []EmailAddress{
			{
				Email: recipientEmail,
			},
		},
		Subject: subject,
		HTML:    htmlContent,
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		log.Printf("failed to marshal email request: %v", err)
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	log.Printf("MailerSend request prepared with subject: %s, from: %s <%s>",
		m.SubjectLine, m.FromName, m.FromEmail)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("failed to create http request: %v", err)
		return fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.APIKey)

	log.Printf("Sending request to MailerSend API endpoint: %s", endpoint)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("failed to send http request: %v", err)
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for logging and error details
	var respBody bytes.Buffer
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		log.Printf("failed to read response body: %v", err)
	}

	log.Printf("MailerSend API response status: %d, body: %s", resp.StatusCode, respBody.String())

	if resp.StatusCode != http.StatusAccepted {
		log.Printf("MailerSend API error: status code: %d, response: %s",
			resp.StatusCode, respBody.String())
		return fmt.Errorf("failed to send email, status code: %d, response: %s",
			resp.StatusCode, respBody.String())
	}

	log.Printf("Successfully sent email to %s via MailerSend API", recipientEmail)
	return nil
}
