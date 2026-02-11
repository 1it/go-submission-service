package forms

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/1it/go-submission-service/internal/config"
)

// Validator handles form data validation
type Validator struct {
	config *config.Config
}

// NewValidator creates a new form validator
func NewValidator(config *config.Config) *Validator {
	return &Validator{
		config: config,
	}
}

// ValidateSubmission validates form data against configuration rules
func (v *Validator) ValidateSubmission(formData map[string]interface{}) (map[string]string, error) {
	errors := make(map[string]string)
	validateRequiredFields(formData, v.config.Form.RequiredFields, errors)
	validateEmailField(formData, v.config.Form.EmailField, errors)
	validateConfigRules(formData, v.config.Form.ValidationRules, errors)
	validateCrossFieldRules(formData, v.config.Form.CrossFieldRules, errors)
	if len(errors) > 0 {
		return errors, fmt.Errorf("validation failed")
	}
	return nil, nil
}

func validateRequiredFields(formData map[string]interface{}, required []string, errors map[string]string) {
	for _, field := range required {
		if value, exists := formData[field]; !exists || isEmpty(value) {
			errors[field] = fmt.Sprintf("Field '%s' is required", field)
		}
	}
}

func validateEmailField(formData map[string]interface{}, emailField string, errors map[string]string) {
	if emailField == "" {
		return
	}
	if email, ok := formData[emailField].(string); ok {
		if email != "" && !isValidEmail(email) {
			errors[emailField] = "Invalid email address"
		}
	} else if _, exists := formData[emailField]; exists {
		errors[emailField] = "Email must be a string"
	}
}

func validateConfigRules(formData map[string]interface{}, rules map[string]config.ValidationRule, errors map[string]string) {
	for field, rule := range rules {
		if value, exists := formData[field]; exists {
			if err := validateField(field, value, rule); err != nil {
				if rule.Message != "" {
					errors[field] = rule.Message
				} else {
					errors[field] = err.Error()
				}
			}
		}
	}
}

func validateCrossFieldRules(formData map[string]interface{}, rules []config.CrossFieldRule, errors map[string]string) {
	if len(rules) == 0 {
		return
	}
	crossFieldErrors, err := ValidateCrossFieldRules(formData, rules)
	if err != nil {
		for field, message := range crossFieldErrors {
			errors[field] = message
		}
	}
}

// SanitizeSubmission sanitizes form data
func (v *Validator) SanitizeSubmission(formData map[string]interface{}) map[string]interface{} {
	options := DefaultSanitizationOptions()

	// Check if custom sanitization options are defined in config
	if v.config != nil && v.config.Form.SanitizationOptions != nil {
		if v.config.Form.SanitizationOptions.DisableHTMLEscaping {
			options.EscapeHTML = false
		}
		if v.config.Form.SanitizationOptions.DisableHTMLTagRemoval {
			options.RemoveHTMLTags = false
		}
	}

	return SanitizeMap(formData, options)
}

// validateField is now imported from validation_rules.go

// isValidEmail checks if a string is a valid email address
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// isEmpty checks if a value is empty
func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return false // Numbers are never empty
	case bool:
		return false // Booleans are never empty
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// ValidateBusinessContactForm validates a business contact form specifically
func (v *Validator) ValidateBusinessContactForm(formData map[string]interface{}) (map[string]string, error) {
	errors := make(map[string]string)
	validateBusinessContactRequired(formData, errors)
	validateBusinessContactWorkEmail(formData, errors)
	validateBusinessContactCompanySize(formData, errors)
	validateBusinessContactFullName(formData, errors)
	validateBusinessContactCompany(formData, errors)
	validateBusinessContactPhone(formData, errors)
	validateBusinessContactJobTitle(formData, errors)
	validateBusinessContactMessage(formData, errors)
	if len(errors) > 0 {
		return errors, fmt.Errorf("business contact form validation failed")
	}
	return nil, nil
}

func validateBusinessContactRequired(formData map[string]interface{}, errors map[string]string) {
	for _, field := range []string{"full_name", "work_email", "company", "company_size"} {
		if value, exists := formData[field]; !exists || isEmpty(value) {
			errors[field] = fmt.Sprintf("Field '%s' is required", field)
		}
	}
}

func validateBusinessContactWorkEmail(formData map[string]interface{}, errors map[string]string) {
	if workEmail, ok := formData["work_email"].(string); ok && workEmail != "" && !isValidEmail(workEmail) {
		errors["work_email"] = "Invalid email address"
	}
}

func validateBusinessContactCompanySize(formData map[string]interface{}, errors map[string]string) {
	if companySize, ok := formData["company_size"].(string); ok && companySize != "" {
		validSizes := []string{"50-200", "200-500", "500-1000", "1000-5000"}
		if !contains(validSizes, companySize) {
			errors["company_size"] = "Invalid company size. Must be one of: 50-200, 200-500, 500-1000, 1000-5000"
		}
	}
}

func validateBusinessContactFullName(formData map[string]interface{}, errors map[string]string) {
	if fullName, ok := formData["full_name"].(string); ok && fullName != "" {
		if len(strings.TrimSpace(fullName)) < 2 {
			errors["full_name"] = "Full name must be at least 2 characters long"
		} else if len(fullName) > 100 {
			errors["full_name"] = "Full name must be less than 100 characters"
		}
	}
}

func validateBusinessContactCompany(formData map[string]interface{}, errors map[string]string) {
	if company, ok := formData["company"].(string); ok && company != "" {
		if len(strings.TrimSpace(company)) < 2 {
			errors["company"] = "Company name must be at least 2 characters long"
		} else if len(company) > 200 {
			errors["company"] = "Company name must be less than 200 characters"
		}
	}
}

func validateBusinessContactPhone(formData map[string]interface{}, errors map[string]string) {
	if phone, ok := formData["phone"].(string); ok && phone != "" {
		phoneRegex := regexp.MustCompile(`^[\+]?[1-9][\d]{0,15}$|^[\+]?[1-9][\d\s\-\(\)]{7,20}$`)
		if !phoneRegex.MatchString(phone) {
			errors["phone"] = "Invalid phone number format"
		}
	}
}

func validateBusinessContactJobTitle(formData map[string]interface{}, errors map[string]string) {
	if jobTitle, ok := formData["job_title"].(string); ok && jobTitle != "" && len(jobTitle) > 100 {
		errors["job_title"] = "Job title must be less than 100 characters"
	}
}

func validateBusinessContactMessage(formData map[string]interface{}, errors map[string]string) {
	if message, ok := formData["message"].(string); ok && message != "" && len(message) > 1000 {
		errors["message"] = "Message must be less than 1000 characters"
	}
}

// ValidateFormByType validates form data based on form type
func (v *Validator) ValidateFormByType(formData map[string]interface{}) (map[string]string, error) {
	formType, _ := formData["form_type"].(string)

	switch formType {
	case "business_contact":
		return v.ValidateBusinessContactForm(formData)
	default:
		// Use generic validation for other form types
		return v.ValidateSubmission(formData)
	}
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
