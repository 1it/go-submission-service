package forms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/1it/go-submission-service/internal/config"
)

// RequestValidationOptions defines options for request validation
type RequestValidationOptions struct {
	MaxBodySize         int64    // Maximum request body size in bytes
	RequiredHeaders     []string // Headers that must be present
	AllowedMethods      []string // HTTP methods that are allowed
	AllowedContentTypes []string // Content types that are allowed
}

// DefaultRequestValidationOptions returns default request validation options
func DefaultRequestValidationOptions() RequestValidationOptions {
	return RequestValidationOptions{
		MaxBodySize:         1024 * 1024, // 1MB
		RequiredHeaders:     []string{"Content-Type"},
		AllowedMethods:      []string{http.MethodPost},
		AllowedContentTypes: []string{"application/json"},
	}
}

// ValidateRequest validates an HTTP request
func ValidateRequest(r *http.Request, options RequestValidationOptions) ([]byte, error) {
	// 1. Validate HTTP method
	methodAllowed := false
	for _, method := range options.AllowedMethods {
		if r.Method == method {
			methodAllowed = true
			break
		}
	}
	if !methodAllowed {
		return nil, fmt.Errorf("method not allowed: %s", r.Method)
	}

	// 2. Validate required headers
	for _, header := range options.RequiredHeaders {
		if r.Header.Get(header) == "" {
			return nil, fmt.Errorf("missing required header: %s", header)
		}
	}

	// 3. Validate Content-Type if specified
	if len(options.AllowedContentTypes) > 0 {
		contentType := r.Header.Get("Content-Type")
		contentTypeValid := false

		// Extract the base content type without parameters
		baseContentType := contentType
		if idx := strings.Index(contentType, ";"); idx != -1 {
			baseContentType = contentType[:idx]
		}
		baseContentType = strings.TrimSpace(baseContentType)

		for _, allowedType := range options.AllowedContentTypes {
			if baseContentType == allowedType {
				contentTypeValid = true
				break
			}
		}

		if !contentTypeValid {
			return nil, fmt.Errorf("unsupported content type: %s", contentType)
		}
	}

	// 4. Validate and read request body with size limit
	body, err := readBodyWithSizeLimit(r, options.MaxBodySize)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// readBodyWithSizeLimit reads the request body with a size limit
func readBodyWithSizeLimit(r *http.Request, maxSize int64) ([]byte, error) {
	// Set a size limit on the request body
	r.Body = http.MaxBytesReader(nil, r.Body, maxSize)

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			return nil, fmt.Errorf("request body exceeds maximum size of %d bytes", maxSize)
		}
		return nil, fmt.Errorf("error reading request body: %v", err)
	}

	// Reset the body so it can be read again if needed
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}

// ValidateJSONRequest validates a JSON request and returns the parsed body
func ValidateJSONRequest(r *http.Request, options RequestValidationOptions, v interface{}) error {
	// Ensure Content-Type is application/json
	options.AllowedContentTypes = []string{"application/json"}

	// Validate the request
	body, err := ValidateRequest(r, options)
	if err != nil {
		return err
	}

	// Parse JSON body
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	return nil
}

// ValidateFormRequest validates a form request and returns the parsed form data
func ValidateFormRequest(r *http.Request, cfg *config.Config, options RequestValidationOptions) (map[string]interface{}, error) {
	// Ensure Content-Type is application/json
	options.AllowedContentTypes = []string{"application/json"}

	// Define the request structure
	var req struct {
		FormData map[string]interface{} `json:"form_data"`
	}

	// Validate and parse the request
	if err := ValidateJSONRequest(r, options, &req); err != nil {
		return nil, err
	}

	// Check if form data is present
	if req.FormData == nil {
		return nil, fmt.Errorf("missing form_data field")
	}

	// Sanitize the form data
	validator := NewValidator(cfg)
	sanitizedData := validator.SanitizeSubmission(req.FormData)

	return sanitizedData, nil
}
