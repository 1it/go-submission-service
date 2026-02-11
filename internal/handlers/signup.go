package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/1it/go-submission-service/internal/config"
	"github.com/1it/go-submission-service/internal/database"
)

type SignupRequest struct {
	Email          string `json:"email"`
	Platform       string `json:"platform,omitempty"`
	Source         string `json:"source,omitempty"`
	RecaptchaToken string `json:"recaptchaToken,omitempty"`
	TurnstileToken string `json:"turnstile_token,omitempty"`
}

type SignupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// LegacySignupHandler handles legacy beta signup requests
// This is kept for backward compatibility and converts signup requests to form submissions
func LegacySignupHandler(repo *database.Repository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Legacy signup request received, converting to form submission")

		// Parse legacy signup request
		var req SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteJSONError(w, "Invalid request body", http.StatusBadRequest, nil)
			return
		}

		// Convert to form submission format
		formData := map[string]interface{}{
			"email":     req.Email,
			"form_type": "signup",
		}

		// Add optional fields if provided
		if req.Platform != "" {
			formData["platform"] = req.Platform
		}
		if req.Source != "" {
			formData["source"] = req.Source
		}

		// Create new request body for the submit handler
		submitReq := SubmitRequest{
			FormData:       formData,
			RecaptchaToken: req.RecaptchaToken,
			TurnstileToken: req.TurnstileToken,
		}

		// Convert back to JSON and create new request
		reqBody, err := json.Marshal(submitReq)
		if err != nil {
			WriteJSONError(w, "Failed to process request", http.StatusInternalServerError, nil)
			return
		}

		// Create new request with converted body
		newReq := r.Clone(r.Context())
		newReq.Body = io.NopCloser(strings.NewReader(string(reqBody)))
		newReq.ContentLength = int64(len(reqBody))

		// Delegate to the main submit handler
		SubmitHandler(repo, cfg)(w, newReq)
	}
}
