package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/1it/go-submission-service/internal/config"
	"github.com/1it/go-submission-service/internal/database"
	"github.com/1it/go-submission-service/internal/email"
	"github.com/1it/go-submission-service/internal/forms"
	"github.com/1it/go-submission-service/internal/metrics"
	"github.com/1it/go-submission-service/internal/models"
	"github.com/1it/go-submission-service/internal/retry"
)

// SubmitRequest represents a form submission request
type SubmitRequest struct {
	FormData       map[string]interface{} `json:"form_data"`
	RecaptchaToken string                 `json:"recaptcha_token,omitempty"`
	TurnstileToken string                 `json:"turnstile_token,omitempty"`
}

// SubmitHandler handles form submissions
func SubmitHandler(repo *database.Repository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())

		// Increment total submission requests counter
		metrics.SignupRequestsTotal.Inc()

		// Log request details
		log.Printf("[%s] Request received: %s %s from %s",
			requestID, r.Method, r.URL.Path, r.RemoteAddr)

		// 1. Validate Method
		if r.Method != http.MethodPost {
			WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed, nil)
			log.Printf("[%s] Method not allowed: %s for path %s", requestID, r.Method, r.URL.Path)
			metrics.IncrementSignupFailure("method_not_allowed")
			return
		}

		// 2. Validate and parse request
		options := forms.DefaultRequestValidationOptions()

		// Set max request size from config if available
		if cfg.Form.MaxRequestSize > 0 {
			options.MaxBodySize = cfg.Form.MaxRequestSize
		}

		var req SubmitRequest
		if err := forms.ValidateJSONRequest(r, options, &req); err != nil {
			WriteJSONError(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest, nil)
			log.Printf("[%s] Invalid request: %v", requestID, err)
			metrics.IncrementSignupFailure("invalid_request")
			return
		}

		// 3. Validate Form Data
		validator := forms.NewValidator(cfg)

		// Sanitize input data
		req.FormData = validator.SanitizeSubmission(req.FormData)

		// Validate form data using form-type-specific validation
		validationErrors, err := validator.ValidateFormByType(req.FormData)
		if err != nil {
			WriteJSONError(w, "Validation failed", http.StatusBadRequest, validationErrors)
			log.Printf("[%s] Validation failed: %v", requestID, validationErrors)
			metrics.IncrementSignupFailure("validation_failed")
			return
		}

		// Get email from form data for logging (handle different form types)
		email := getEmailFromFormData(req.FormData, cfg.Form.EmailField)

		log.Printf("[%s] Processing form submission with email: %s", requestID, email)

		// 4. Verify reCAPTCHA token if enabled
		if cfg.Security.ReCAPTCHA.Enabled || cfg.Security.ReCAPTCHA.EnterpriseEnabled {
			if !VerifyRecaptchaToken(w, r, cfg, req.RecaptchaToken, requestID) {
				return
			}
		}
		// 4b. Verify Turnstile token if enabled
		if cfg.Security.Turnstile.Enabled {
			if !VerifyTurnstileToken(w, r, cfg, req.TurnstileToken, requestID) {
				return
			}
		}

		// 5. Check if submission with this email already exists
		if email != "" {
			// Determine the email field based on form type
			formType := getFormType(req.FormData)
			var emailField string
			switch formType {
			case "business_contact":
				emailField = "work_email"
			default:
				emailField = cfg.Form.EmailField
			}

			_, err := repo.Submissions.GetByEmail(email, emailField)
			if err == nil {
				// Email already exists
				WriteJSONError(w, "Email already registered", http.StatusConflict, nil)
				log.Printf("[%s] Email already registered: %s", requestID, email)
				metrics.IncrementSignupFailure("email_already_exists")
				return
			} else if err.Error() != fmt.Sprintf("submission not found for email: %s", email) {
				// Other database error during check
				log.Printf("[%s] Error checking submission %s: %v", requestID, email, err)
				WriteJSONError(w, "Failed to check registration status", http.StatusInternalServerError, nil)
				metrics.IncrementSignupFailure("db_error")
				return
			}
		}

		// 6. Add submission to database
		submission := &models.Submission{
			FormData:  req.FormData,
			Status:    "pending",
			FormType:  getFormType(req.FormData),
			Source:    "api",
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
			Referrer:  r.Referer(),
		}

		if err := repo.Submissions.Create(submission); err != nil {
			log.Printf("Error adding submission: %v", err)
			WriteJSONError(w, "Failed to process submission", http.StatusInternalServerError, nil)
			metrics.IncrementSignupFailure("db_error")
			return
		}

		submissionID := submission.ID

		// 7. Send email if email field is present
		emailSent := false
		if email != "" {
			// Try to send email with retry logic
			retryConfig := retry.DefaultRetryConfig()
			retryConfig.MaxAttempts = 2 // Limit to 2 attempts for immediate processing
			retryConfig.InitialDelay = 500 * time.Millisecond

			err := retry.ExecuteWithRetry(retryConfig, func() error {
				return sendEmailWithRetryableError(cfg, email, req.FormData)
			}, fmt.Sprintf("email for submission %s", submissionID))

			if err != nil {
				log.Printf("Error sending email to %s: %v", email, err)
				log.Printf("Submission %s will remain pending for background processing", submissionID)
				// Don't fail the request if email sending fails - it will be retried by background job
			} else {
				emailSent = true
			}
		}

		// 8. Update submission status only if email was sent successfully
		if emailSent {
			if err := repo.Submissions.MarkProcessed(submissionID); err != nil {
				log.Printf("Failed to update status for submission %s: %v", submissionID, err)
				// Continue successfully even if status update fails
			}
		} else {
			log.Printf("Submission %s remains pending - will be processed by background job", submissionID)
		}

		// 9. Return Success Response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ErrorResponse{
			Success: true,
			Message: cfg.Form.SuccessMessage,
		})

		// Increment successful submission counter
		metrics.SignupSuccessTotal.Inc()

		duration := time.Since(startTime)
		log.Printf("[%s] Successfully processed submission. Duration: %v", requestID, duration)
	}
}

// sendEmailWithRetryableError sends an email with proper error classification for retries
func sendEmailWithRetryableError(cfg *config.Config, recipientEmail string, formData map[string]interface{}) error {
	err := sendEmail(cfg, recipientEmail, formData)
	if err != nil {
		// Classify error as retryable or not
		return retry.RetryableError{
			Err:       err,
			Retryable: isEmailErrorRetryable(err),
		}
	}
	return nil
}

// sendEmail sends an email for a form submission
func sendEmail(cfg *config.Config, recipientEmail string, formData map[string]interface{}) error {
	// Create email service
	emailService, err := email.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to create email service: %w", err)
	}

	// Add timestamp to form data for templates
	formData["timestamp"] = time.Now().Format("January 2, 2006 at 3:04 PM MST")

	// Determine templates based on form type
	formType := getFormType(formData)
	confirmationTemplate := "confirmation.html"
	adminTemplate := "form_submission.html"

	switch formType {
	case "business_contact":
		confirmationTemplate = "business_contact_confirmation.html"
		adminTemplate = "business_contact.html"
	}

	// Check if a specific template is requested in the form data (override)
	if template, ok := formData["template"].(string); ok && template != "" {
		confirmationTemplate = template
	}

	// Send confirmation email to the submitter
	if err := emailService.SendTemplatedEmail(recipientEmail, confirmationTemplate, formData); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	// If admin email is configured, send notification to admin
	if adminEmail := cfg.Email.AdminEmail; adminEmail != "" {
		if err := emailService.SendTemplatedEmail(adminEmail, adminTemplate, formData); err != nil {
			// Log but don't fail if admin notification fails
			log.Printf("Failed to send admin notification email: %v", err)
		}
	}

	return nil
}

// getFormType extracts the form type from form data
func getFormType(formData map[string]interface{}) string {
	if formType, ok := formData["form_type"].(string); ok && formType != "" {
		return formType
	}
	return "generic"
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// getEmailFromFormData extracts email from form data based on form type
func getEmailFromFormData(formData map[string]interface{}, defaultEmailField string) string {
	formType := getFormType(formData)

	// Use form-type-specific email field
	var emailField string
	switch formType {
	case "business_contact":
		emailField = "work_email"
	default:
		emailField = defaultEmailField
	}

	if emailValue, ok := formData[emailField].(string); ok {
		return emailValue
	}

	// Fallback to default email field if form-specific field not found
	if emailField != defaultEmailField {
		if emailValue, ok := formData[defaultEmailField].(string); ok {
			return emailValue
		}
	}

	return ""
}

// isEmailErrorRetryable determines if an email error should be retried
func isEmailErrorRetryable(err error) bool {
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
		"mailhog", // For testing - mailhog connection issues
	}

	for _, retryableErr := range retryableErrors {
		if containsError(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// containsError checks if an error string contains a specific substring
func containsError(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstringError(s, substr)
}

func containsSubstringError(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// HealthHandler handles basic health checks
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Health check requested via %s", r.Method)

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "go-submission-service",
		"version":   "1.0.0",
		"features": map[string]bool{
			"email_retry":        true,
			"background_jobs":    true,
			"form_validation":    true,
			"database_migration": true,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
