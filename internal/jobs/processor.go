package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/1it/go-submission-service/internal/config"
	"github.com/1it/go-submission-service/internal/database"
	"github.com/1it/go-submission-service/internal/email"
	"github.com/1it/go-submission-service/internal/models"
	"github.com/1it/go-submission-service/internal/retry"
)

// JobProcessor handles background job processing
type JobProcessor struct {
	repo        *database.Repository
	config      *config.Config
	retryConfig retry.RetryConfig
	stopChan    chan struct{}
	running     bool
}

// NewJobProcessor creates a new job processor
func NewJobProcessor(repo *database.Repository, cfg *config.Config) *JobProcessor {
	return &JobProcessor{
		repo:        repo,
		config:      cfg,
		retryConfig: retry.DefaultRetryConfig(),
		stopChan:    make(chan struct{}),
		running:     false,
	}
}

// Start begins processing background jobs
func (jp *JobProcessor) Start(ctx context.Context) {
	if jp.running {
		log.Println("Job processor is already running")
		return
	}

	jp.running = true
	log.Println("Starting job processor...")

	// Process pending submissions every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Process immediately on start
	jp.processPendingSubmissions()

	for {
		select {
		case <-ctx.Done():
			log.Println("Job processor stopping due to context cancellation")
			jp.running = false
			return
		case <-jp.stopChan:
			log.Println("Job processor stopping due to stop signal")
			jp.running = false
			return
		case <-ticker.C:
			jp.processPendingSubmissions()
		}
	}
}

// Stop stops the job processor
func (jp *JobProcessor) Stop() {
	if !jp.running {
		return
	}

	log.Println("Stopping job processor...")
	close(jp.stopChan)
}

// processPendingSubmissions finds and processes submissions that are still pending
func (jp *JobProcessor) processPendingSubmissions() {
	log.Println("Processing pending submissions...")

	// Get all pending submissions older than 5 minutes
	cutoffTime := time.Now().Add(-5 * time.Minute)
	submissions, err := jp.repo.Submissions.GetByDateRange(time.Time{}, cutoffTime)
	if err != nil {
		log.Printf("Error fetching pending submissions: %v", err)
		return
	}

	pendingCount := 0
	processedCount := 0
	failedCount := 0

	for _, submission := range submissions {
		if submission.Status == "pending" {
			pendingCount++

			// Try to process the submission
			if jp.processSubmission(submission) {
				processedCount++
			} else {
				failedCount++
			}
		}
	}

	if pendingCount > 0 {
		log.Printf("Processed %d pending submissions: %d succeeded, %d failed",
			pendingCount, processedCount, failedCount)
	}
}

// processSubmission attempts to process a single pending submission
func (jp *JobProcessor) processSubmission(submission *models.Submission) bool {
	log.Printf("Processing pending submission: %s", submission.ID)

	// Extract email from form data
	email := jp.getEmailFromSubmission(submission)
	if email == "" {
		log.Printf("No email found in submission %s, marking as failed", submission.ID)
		jp.repo.Submissions.UpdateStatus(submission.ID, "failed")
		return false
	}

	// Attempt to send email with retry
	success := false
	err := retry.ExecuteWithRetry(jp.retryConfig, func() error {
		return jp.sendEmailForSubmission(email, submission)
	}, fmt.Sprintf("email for submission %s", submission.ID))

	if err != nil {
		log.Printf("Failed to send email for submission %s after retries: %v", submission.ID, err)

		// Mark as failed if we've exceeded retry attempts
		if jp.shouldMarkAsFailed(submission) {
			jp.repo.Submissions.UpdateStatus(submission.ID, "failed")
			log.Printf("Marked submission %s as failed after multiple retry attempts", submission.ID)
		}
	} else {
		// Mark as processed
		jp.repo.Submissions.MarkProcessed(submission.ID)
		log.Printf("Successfully processed pending submission: %s", submission.ID)
		success = true
	}

	return success
}

// sendEmailForSubmission sends email for a specific submission
func (jp *JobProcessor) sendEmailForSubmission(emailAddr string, submission *models.Submission) error {
	// Create email service
	emailService, err := email.NewService(jp.config)
	if err != nil {
		return retry.RetryableError{
			Err:       fmt.Errorf("failed to create email service: %w", err),
			Retryable: true,
		}
	}

	// Add timestamp to form data for templates
	formData := make(map[string]interface{})
	for k, v := range submission.FormData {
		formData[k] = v
	}
	formData["timestamp"] = time.Now().Format("January 2, 2006 at 3:04 PM MST")

	// Determine templates based on form type
	confirmationTemplate := "confirmation.html"
	adminTemplate := "form_submission.html"

	switch submission.FormType {
	case "business_contact":
		confirmationTemplate = "business_contact_confirmation.html"
		adminTemplate = "business_contact.html"
	}

	// Send confirmation email to the submitter
	err = emailService.SendTemplatedEmail(emailAddr, confirmationTemplate, formData)
	if err != nil {
		return retry.RetryableError{
			Err:       fmt.Errorf("failed to send confirmation email: %w", err),
			Retryable: jp.isEmailErrorRetryable(err),
		}
	}

	// If admin email is configured, send notification to admin
	if adminEmail := jp.config.Email.AdminEmail; adminEmail != "" {
		err = emailService.SendTemplatedEmail(adminEmail, adminTemplate, formData)
		if err != nil {
			// Log but don't fail if admin notification fails
			log.Printf("Failed to send admin notification email for submission %s: %v", submission.ID, err)
		}
	}

	return nil
}

// getEmailFromSubmission extracts email from submission based on form type
func (jp *JobProcessor) getEmailFromSubmission(submission *models.Submission) string {
	var emailField string
	switch submission.FormType {
	case "business_contact":
		emailField = "work_email"
	default:
		emailField = jp.config.Form.EmailField
	}

	if emailValue, ok := submission.FormData[emailField].(string); ok {
		return emailValue
	}

	// Fallback to default email field
	if emailField != jp.config.Form.EmailField {
		if emailValue, ok := submission.FormData[jp.config.Form.EmailField].(string); ok {
			return emailValue
		}
	}

	return ""
}

// isEmailErrorRetryable determines if an email error should be retried
func (jp *JobProcessor) isEmailErrorRetryable(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Retryable email errors
	retryableErrors := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"dial tcp",
		"no such host",
		"network is unreachable",
		"temporary failure",
		"service unavailable",
		"rate limit",
		"quota exceeded",
		"server error",
	}

	for _, retryableErr := range retryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// shouldMarkAsFailed determines if a submission should be marked as failed
func (jp *JobProcessor) shouldMarkAsFailed(submission *models.Submission) bool {
	// Mark as failed if submission is older than 24 hours
	return time.Since(submission.CreatedAt) > 24*time.Hour
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstring(s, substr)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
