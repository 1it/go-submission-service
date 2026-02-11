package retry

import (
	"fmt"
	"log"
	"math"
	"time"
)

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts   int           `json:"max_attempts"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// DefaultRetryConfig returns sensible defaults for retry behavior
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	Err        error
	Retryable  bool
	StatusCode int
}

func (r RetryableError) Error() string {
	return r.Err.Error()
}

// IsRetryable determines if an error should be retried
func IsRetryable(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	// Retry on 5xx server errors
	if statusCode >= 500 && statusCode < 600 {
		return true
	}

	// Retry on specific network/timeout errors
	errStr := err.Error()
	retryableErrors := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"no such host",
		"network is unreachable",
		"temporary failure",
		"dial tcp",
		"i/o timeout",
	}

	for _, retryableErr := range retryableErrors {
		if contains(errStr, retryableErr) {
			return true
		}
	}

	return false
}

// ExecuteWithRetry executes a function with retry logic
func ExecuteWithRetry(config RetryConfig, operation func() error, operationName string) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		log.Printf("Executing %s (attempt %d/%d)", operationName, attempt, config.MaxAttempts)

		err := operation()
		if err == nil {
			if attempt > 1 {
				log.Printf("%s succeeded on attempt %d", operationName, attempt)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if retryableErr, ok := err.(RetryableError); ok && !retryableErr.Retryable {
			log.Printf("%s failed with non-retryable error: %v", operationName, err)
			return err
		}

		// Don't sleep after the last attempt
		if attempt < config.MaxAttempts {
			delay := calculateDelay(config, attempt)
			log.Printf("%s failed (attempt %d/%d): %v. Retrying in %v",
				operationName, attempt, config.MaxAttempts, err, delay)
			time.Sleep(delay)
		} else {
			log.Printf("%s failed after %d attempts: %v", operationName, config.MaxAttempts, err)
		}
	}

	return fmt.Errorf("operation %s failed after %d attempts: %w", operationName, config.MaxAttempts, lastErr)
}

// ExecuteWithRetryAsync executes a function with retry logic asynchronously
func ExecuteWithRetryAsync(config RetryConfig, operation func() error, operationName string, callback func(error)) {
	go func() {
		err := ExecuteWithRetry(config, operation, operationName)
		if callback != nil {
			callback(err)
		}
	}()
}

// calculateDelay calculates the delay for the next retry attempt using exponential backoff
func calculateDelay(config RetryConfig, attempt int) time.Duration {
	delay := float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt-1))

	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	return time.Duration(delay)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
