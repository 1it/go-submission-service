package forms

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/1it/go-submission-service/internal/config"
)

// ValidationRuleFunc defines a function type for field validation
type ValidationRuleFunc func(value interface{}, rule config.ValidationRule) error

// validationRules maps rule types to validation functions
var validationRules = map[string]ValidationRuleFunc{
	"email":    validateEmail,
	"regex":    validateRegex,
	"length":   validateLength,
	"url":      validateURL,
	"numeric":  validateNumeric,
	"boolean":  validateBoolean,
	"enum":     validateEnum,
	"date":     validateDate,
	"required": validateRequired,
	"phone":    validatePhone,
}

// validateField validates a field using the appropriate validation function
func validateField(field string, value interface{}, rule config.ValidationRule) error {
	// Skip validation for nil or empty values unless it's a required field
	if rule.Type != "required" && isEmpty(value) {
		return nil
	}

	if field == "email" {
		if err := validateEmail(value, rule); err != nil {
			return err
		}
	}
	if field == "url" {
		if err := validateURL(value, rule); err != nil {
			return err
		}
	}

	// Get the validation function for this rule type
	validateFunc, exists := validationRules[rule.Type]
	if !exists {
		return fmt.Errorf("unknown validation rule type: %s", rule.Type)
	}

	// Run the validation
	if err := validateFunc(value, rule); err != nil {
		if rule.Message != "" {
			return fmt.Errorf(rule.Message)
		}
		return err
	}

	return nil
}

// validateEmail validates an email address
func validateEmail(value interface{}, rule config.ValidationRule) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("email must be a string")
	}

	if _, err := mail.ParseAddress(strValue); err != nil {
		return fmt.Errorf("invalid email address")
	}

	return nil
}

// validateRegex validates a string against a regular expression
func validateRegex(value interface{}, rule config.ValidationRule) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	if rule.Pattern == "" {
		return fmt.Errorf("regex pattern is required")
	}

	matched, err := regexp.MatchString(rule.Pattern, strValue)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %v", err)
	}

	if !matched {
		return fmt.Errorf("value does not match pattern")
	}

	return nil
}

// validateLength validates the length of a string
func validateLength(value interface{}, rule config.ValidationRule) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	length := utf8.RuneCountInString(strValue)

	if rule.Min > 0 && length < rule.Min {
		return fmt.Errorf("value must be at least %d characters", rule.Min)
	}

	if rule.Max > 0 && length > rule.Max {
		return fmt.Errorf("value must be at most %d characters", rule.Max)
	}

	return nil
}

// validateURL validates a URL
func validateURL(value interface{}, rule config.ValidationRule) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("URL must be a string")
	}

	parsedURL, err := url.Parse(strValue)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	// Check if URL has scheme and host
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("URL must include scheme and host (e.g., https://example.com)")
	}

	// Optionally check for specific schemes
	if rule.Pattern != "" {
		schemes := strings.Split(rule.Pattern, ",")
		validScheme := false
		for _, scheme := range schemes {
			if parsedURL.Scheme == strings.TrimSpace(scheme) {
				validScheme = true
				break
			}
		}
		if !validScheme {
			return fmt.Errorf("URL must use one of these schemes: %s", rule.Pattern)
		}
	}

	return nil
}

// validateNumeric validates numeric values
func validateNumeric(value interface{}, rule config.ValidationRule) error {
	// Handle different numeric types
	var numValue float64

	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		numValue = float64(v.(int64))
	case uint, uint8, uint16, uint32, uint64:
		numValue = float64(v.(uint64))
	case float32, float64:
		numValue = v.(float64)
	case string:
		// Try to parse string as number
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("value must be a number")
		}
		numValue = parsed
	default:
		return fmt.Errorf("value must be a number")
	}

	// Check min/max constraints
	if rule.Min != 0 && numValue < float64(rule.Min) {
		return fmt.Errorf("value must be at least %d", rule.Min)
	}

	if rule.Max != 0 && numValue > float64(rule.Max) {
		return fmt.Errorf("value must be at most %d", rule.Max)
	}

	return nil
}

// validateBoolean validates boolean values
func validateBoolean(value interface{}, _ config.ValidationRule) error {
	switch value := value.(type) {
	case bool:
		return nil
	case string:
		strValue := strings.ToLower(value)
		if strValue == "true" || strValue == "false" || strValue == "yes" || strValue == "no" || strValue == "1" || strValue == "0" {
			return nil
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// 0 and 1 are valid boolean values
		return nil
	}

	return fmt.Errorf("value must be a boolean (true/false, yes/no, 1/0)")
}

// validateEnum validates that a value is one of a set of allowed values
func validateEnum(value interface{}, rule config.ValidationRule) error {
	if rule.Pattern == "" {
		return fmt.Errorf("enum values must be specified in pattern field")
	}

	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("value must be a string")
	}

	allowedValues := strings.Split(rule.Pattern, ",")
	for _, allowed := range allowedValues {
		if strings.TrimSpace(allowed) == strValue {
			return nil
		}
	}

	return fmt.Errorf("value must be one of: %s", rule.Pattern)
}

// validateDate validates date strings
func validateDate(value interface{}, rule config.ValidationRule) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("date must be a string")
	}

	// Default format is RFC3339
	format := "2006-01-02T15:04:05Z07:00"
	if rule.Pattern != "" {
		format = rule.Pattern
	}

	_, err := time.Parse(format, strValue)
	if err != nil {
		return fmt.Errorf("invalid date format, expected: %s", format)
	}

	return nil
}

// validateRequired validates that a value is not empty
func validateRequired(value interface{}, _ config.ValidationRule) error {
	if isEmpty(value) {
		return fmt.Errorf("value is required")
	}
	return nil
}

// validatePhone validates phone numbers
func validatePhone(value interface{}, rule config.ValidationRule) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("phone number must be a string")
	}

	// Default phone regex pattern
	pattern := `^[\+]?[1-9][\d]{0,15}$|^[\+]?[1-9][\d\s\-\(\)]{7,20}$`
	if rule.Pattern != "" {
		pattern = rule.Pattern
	}

	matched, err := regexp.MatchString(pattern, strValue)
	if err != nil {
		return fmt.Errorf("invalid phone regex pattern: %v", err)
	}

	if !matched {
		return fmt.Errorf("invalid phone number format")
	}

	return nil
}
