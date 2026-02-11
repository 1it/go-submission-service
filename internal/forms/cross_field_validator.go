package forms

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/1it/go-submission-service/internal/config"
)

// CrossFieldValidationFunc defines a function type for cross-field validation
type CrossFieldValidationFunc func(values map[string]interface{}, rule config.CrossFieldRule) error

// crossFieldValidators maps rule types to validation functions
var crossFieldValidators = map[string]CrossFieldValidationFunc{
	"equal":           validateFieldsEqual,
	"not_equal":       validateFieldsNotEqual,
	"one_of_required": validateOneOfRequired,
	"all_or_none":     validateAllOrNone,
	"dependent":       validateDependent,
}

// ValidateCrossFieldRules validates rules that span multiple fields
func ValidateCrossFieldRules(formData map[string]interface{}, rules []config.CrossFieldRule) (map[string]string, error) {
	errors := make(map[string]string)

	for _, rule := range rules {
		// Check if condition is met
		if rule.Condition != "" && !evaluateCondition(formData, rule.Condition) {
			continue
		}

		// Get the validation function
		validateFunc, exists := crossFieldValidators[rule.Type]
		if !exists {
			errors["_form"] = fmt.Sprintf("Unknown cross-field validation rule type: %s", rule.Type)
			continue
		}

		// Run the validation
		if err := validateFunc(formData, rule); err != nil {
			// Use the first field as the error key, or "_form" for form-level errors
			errorKey := "_form"
			if len(rule.Fields) > 0 {
				errorKey = rule.Fields[0]
			}

			if rule.Message != "" {
				errors[errorKey] = rule.Message
			} else {
				errors[errorKey] = err.Error()
			}
		}
	}

	if len(errors) > 0 {
		return errors, fmt.Errorf("cross-field validation failed")
	}

	return nil, nil
}

// validateFieldsEqual validates that all specified fields have the same value
func validateFieldsEqual(formData map[string]interface{}, rule config.CrossFieldRule) error {
	if len(rule.Fields) < 2 {
		return fmt.Errorf("equal rule requires at least two fields")
	}

	var firstValue interface{}
	var firstField string

	for i, field := range rule.Fields {
		value, exists := formData[field]
		if !exists {
			return fmt.Errorf("field '%s' is missing", field)
		}

		if i == 0 {
			firstValue = value
			firstField = field
		} else if !reflect.DeepEqual(value, firstValue) {
			return fmt.Errorf("field '%s' must be equal to '%s'", field, firstField)
		}
	}

	return nil
}

// validateFieldsNotEqual validates that all specified fields have different values
func validateFieldsNotEqual(formData map[string]interface{}, rule config.CrossFieldRule) error {
	if len(rule.Fields) < 2 {
		return fmt.Errorf("not_equal rule requires at least two fields")
	}

	values := make(map[interface{}]string)

	for _, field := range rule.Fields {
		value, exists := formData[field]
		if !exists {
			continue // Skip missing fields
		}

		if existingField, found := values[value]; found {
			return fmt.Errorf("field '%s' must not be equal to field '%s'", field, existingField)
		}

		values[value] = field
	}

	return nil
}

// validateOneOfRequired validates that at least one of the specified fields is present
func validateOneOfRequired(formData map[string]interface{}, rule config.CrossFieldRule) error {
	if len(rule.Fields) < 1 {
		return fmt.Errorf("one_of_required rule requires at least one field")
	}

	for _, field := range rule.Fields {
		value, exists := formData[field]
		if exists && !isEmpty(value) {
			return nil // At least one field is present and not empty
		}
	}

	return fmt.Errorf("at least one of these fields is required: %s", strings.Join(rule.Fields, ", "))
}

// validateAllOrNone validates that either all specified fields are present or none are
func validateAllOrNone(formData map[string]interface{}, rule config.CrossFieldRule) error {
	if len(rule.Fields) < 2 {
		return fmt.Errorf("all_or_none rule requires at least two fields")
	}

	presentCount := 0
	for _, field := range rule.Fields {
		value, exists := formData[field]
		if exists && !isEmpty(value) {
			presentCount++
		}
	}

	if presentCount > 0 && presentCount < len(rule.Fields) {
		return fmt.Errorf("either all of these fields must be provided or none: %s", strings.Join(rule.Fields, ", "))
	}

	return nil
}

// validateDependent validates that if the first field is present, the dependent fields are also present
func validateDependent(formData map[string]interface{}, rule config.CrossFieldRule) error {
	if len(rule.Fields) < 2 {
		return fmt.Errorf("dependent rule requires at least two fields")
	}

	primaryField := rule.Fields[0]
	primaryValue, exists := formData[primaryField]

	// If primary field doesn't exist or is empty, no validation needed
	if !exists || isEmpty(primaryValue) {
		return nil
	}

	// Check that all dependent fields exist and are not empty
	for i := 1; i < len(rule.Fields); i++ {
		dependentField := rule.Fields[i]
		dependentValue, exists := formData[dependentField]

		if !exists || isEmpty(dependentValue) {
			return fmt.Errorf("field '%s' is required when '%s' is provided", dependentField, primaryField)
		}
	}

	return nil
}

// evaluateCondition evaluates a simple condition string against form data
func evaluateCondition(formData map[string]interface{}, condition string) bool {
	// Simple condition format: "field=value" or "field!=value" or "field"
	parts := strings.SplitN(condition, "=", 2)

	if len(parts) == 1 {
		// Check if field exists and is not empty
		field := strings.TrimSpace(parts[0])
		if strings.HasPrefix(field, "!") {
			// Check if field doesn't exist or is empty
			field = strings.TrimPrefix(field, "!")
			value, exists := formData[field]
			return !exists || isEmpty(value)
		} else {
			// Check if field exists and is not empty
			value, exists := formData[field]
			return exists && !isEmpty(value)
		}
	} else {
		// Check if field equals value
		field := strings.TrimSpace(parts[0])
		expectedValue := strings.TrimSpace(parts[1])

		if strings.HasPrefix(field, "!") {
			// Check if field doesn't equal value
			field = strings.TrimPrefix(field, "!")
			value, exists := formData[field]
			if !exists {
				return true
			}

			strValue, ok := value.(string)
			if !ok {
				return true // Non-string values are considered not equal
			}

			return strValue != expectedValue
		} else {
			// Check if field equals value
			value, exists := formData[field]
			if !exists {
				return false
			}

			strValue, ok := value.(string)
			if !ok {
				return false // Non-string values are considered not equal
			}

			return strValue == expectedValue
		}
	}
}
