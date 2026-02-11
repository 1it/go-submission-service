package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// WriteJSONError writes a JSON error response with optional validation errors
func WriteJSONError(w http.ResponseWriter, message string, statusCode int, validationErrors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Message: message,
		Errors:  validationErrors,
	}); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}
