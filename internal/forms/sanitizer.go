package forms

import (
	"html"
	"regexp"
	"strings"
)

// SanitizationOptions defines options for data sanitization
type SanitizationOptions struct {
	TrimWhitespace      bool // Trim whitespace from string values
	RemoveHTMLTags      bool // Remove HTML tags from string values
	EscapeHTML          bool // Escape HTML entities
	NormalizeLineBreaks bool // Convert all line breaks to \n
	LowerCase           bool // Convert strings to lowercase
	UpperCase           bool // Convert strings to uppercase
}

// DefaultSanitizationOptions returns default sanitization options
func DefaultSanitizationOptions() SanitizationOptions {
	return SanitizationOptions{
		TrimWhitespace:      true,
		RemoveHTMLTags:      true,
		EscapeHTML:          true,
		NormalizeLineBreaks: true,
		LowerCase:           false,
		UpperCase:           false,
	}
}

// SanitizeString sanitizes a string value based on options
func SanitizeString(value string, options SanitizationOptions) string {
	// Trim whitespace
	if options.TrimWhitespace {
		value = strings.TrimSpace(value)
	}

	// Remove HTML tags
	if options.RemoveHTMLTags {
		re := regexp.MustCompile("<[^>]*>")
		value = re.ReplaceAllString(value, "")
	}

	// Escape HTML entities
	if options.EscapeHTML {
		value = html.EscapeString(value)
	}

	// Normalize line breaks
	if options.NormalizeLineBreaks {
		// Replace all types of line breaks with \n
		value = regexp.MustCompile(`\r\n|\r`).ReplaceAllString(value, "\n")
	}

	// Convert to lowercase
	if options.LowerCase {
		value = strings.ToLower(value)
	}

	// Convert to uppercase
	if options.UpperCase {
		value = strings.ToUpper(value)
	}

	return value
}

// SanitizeMap sanitizes all string values in a map based on options
func SanitizeMap(data map[string]interface{}, options SanitizationOptions) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case string:
			result[key] = SanitizeString(v, options)
		case map[string]interface{}:
			result[key] = SanitizeMap(v, options)
		case []interface{}:
			result[key] = SanitizeArray(v, options)
		default:
			result[key] = v
		}
	}

	return result
}

// SanitizeArray sanitizes all string values in an array based on options
func SanitizeArray(data []interface{}, options SanitizationOptions) []interface{} {
	result := make([]interface{}, len(data))

	for i, value := range data {
		switch v := value.(type) {
		case string:
			result[i] = SanitizeString(v, options)
		case map[string]interface{}:
			result[i] = SanitizeMap(v, options)
		case []interface{}:
			result[i] = SanitizeArray(v, options)
		default:
			result[i] = v
		}
	}

	return result
}
