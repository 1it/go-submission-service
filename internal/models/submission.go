package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Submission represents a form submission with flexible data structure
type Submission struct {
	ID          string                 `json:"id"`
	FormData    map[string]interface{} `json:"form_data"`
	Status      string                 `json:"status"`              // e.g., "pending", "processed", "failed"
	FormType    string                 `json:"form_type,omitempty"` // e.g., "contact", "newsletter", "signup"
	Source      string                 `json:"source,omitempty"`    // e.g., "website", "api", "mobile_app"
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Referrer    string                 `json:"referrer,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
}

// JSONMap is a type that can be stored as JSON in the database
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface for JSONMap
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONMap
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, j)
}

// GetEmail returns the email address from the form data
func (s *Submission) GetEmail(emailField string) string {
	if s.FormData == nil {
		return ""
	}

	if email, ok := s.FormData[emailField].(string); ok {
		return email
	}
	return ""
}

// GetString returns a string value from the form data
func (s *Submission) GetString(field string) string {
	if s.FormData == nil {
		return ""
	}

	if value, ok := s.FormData[field].(string); ok {
		return value
	}
	return ""
}

// GetInt returns an integer value from the form data
func (s *Submission) GetInt(field string) int {
	if s.FormData == nil {
		return 0
	}

	switch v := s.FormData[field].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		// Try to parse as int
		var i int
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			return i
		}
	}
	return 0
}

// GetBool returns a boolean value from the form data
func (s *Submission) GetBool(field string) bool {
	if s.FormData == nil {
		return false
	}

	switch v := s.FormData[field].(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "yes" || v == "1"
	case int:
		return v != 0
	case float64:
		return v != 0
	}
	return false
}
