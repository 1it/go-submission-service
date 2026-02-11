package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/1it/go-submission-service/internal/config"
	"github.com/1it/go-submission-service/internal/metrics"
	"github.com/1it/go-submission-service/internal/recaptcha"
)

// VerifyTurnstileToken verifies a Cloudflare Turnstile token
func VerifyTurnstileToken(w http.ResponseWriter, r *http.Request, cfg *config.Config, token string, requestID string) bool {
	if !cfg.Security.Turnstile.Enabled || cfg.Security.Turnstile.SecretKey == "" {
		log.Printf("[%s] Turnstile not enabled or secret key not set", requestID)
		return true
	}

	log.Printf("[%s] Using Cloudflare Turnstile verification", requestID)
	remoteIP := r.RemoteAddr
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		remoteIP = strings.Split(ip, ",")[0]
	} else if ip := r.Header.Get("X-Real-IP"); ip != "" {
		remoteIP = ip
	} else if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		remoteIP = host
	}
	valid, errMsg, err := recaptcha.VerifyTurnstileToken(
		cfg.Security.Turnstile.SecretKey,
		token,
		remoteIP,
	)
	if err != nil {
		log.Printf("[%s] Turnstile verification error: %v", requestID, err)
		WriteJSONError(w, "Failed to verify Turnstile", http.StatusInternalServerError, nil)
		metrics.IncrementSignupFailure("turnstile_error")
		return false
	}
	if !valid {
		log.Printf("[%s] Turnstile verification failed: %s", requestID, errMsg)
		WriteJSONError(w, "Turnstile verification failed", http.StatusBadRequest, nil)
		metrics.IncrementSignupFailure("turnstile_failed")
		return false
	}
	log.Printf("[%s] Turnstile verification passed", requestID)
	metrics.IncrementRecaptchaVerification("success")
	return true
}

// VerifyRecaptchaToken verifies a reCAPTCHA token using the configured method
func VerifyRecaptchaToken(w http.ResponseWriter, r *http.Request, cfg *config.Config, token string, requestID string) bool {
	if cfg.Security.ReCAPTCHA.EnterpriseEnabled && cfg.Security.ReCAPTCHA.ProjectID != "" {
		// Use reCAPTCHA Enterprise verification
		log.Printf("[%s] Using reCAPTCHA Enterprise verification", requestID)

		valid, score, err := recaptcha.VerifyTokenEnterprise(
			cfg.Security.ReCAPTCHA.ProjectID,
			cfg.Security.ReCAPTCHA.SiteKey,
			token,
			cfg.Security.ReCAPTCHA.ActionName,
		)

		if err != nil {
			log.Printf("[%s] reCAPTCHA Enterprise verification error: %v", requestID, err)
			WriteJSONError(w, "Failed to verify reCAPTCHA", http.StatusInternalServerError, nil)
			metrics.IncrementRecaptchaVerification("error")
			metrics.IncrementSignupFailure("recaptcha_error")
			return false
		}

		if !valid {
			log.Printf("[%s] reCAPTCHA Enterprise verification failed with score: %v", requestID, score)
			WriteJSONError(w, "reCAPTCHA verification failed", http.StatusBadRequest, nil)
			metrics.IncrementRecaptchaVerification("failure")
			metrics.IncrementSignupFailure("recaptcha_failed")
			return false
		}

		log.Printf("[%s] reCAPTCHA Enterprise verification passed with score: %v", requestID, score)
		metrics.IncrementRecaptchaVerification("success")
	} else if cfg.Security.ReCAPTCHA.Enabled && cfg.Security.ReCAPTCHA.SecretKey != "" {
		// Use standard reCAPTCHA verification
		log.Printf("[%s] Using standard reCAPTCHA verification", requestID)
		valid, err := verifyRecaptcha(token, requestID, cfg.Security.ReCAPTCHA.SecretKey)
		if err != nil {
			log.Printf("[%s] Failed to verify reCAPTCHA: %v", requestID, err)
			WriteJSONError(w, "Failed to verify reCAPTCHA", http.StatusInternalServerError, nil)
			metrics.IncrementSignupFailure("recaptcha_error")
			return false
		}

		if !valid {
			log.Printf("[%s] reCAPTCHA verification failed", requestID)
			WriteJSONError(w, "reCAPTCHA verification failed", http.StatusBadRequest, nil)
			metrics.IncrementSignupFailure("recaptcha_failed")
			return false
		}
	}

	return true
}

// verifyRecaptcha verifies the reCAPTCHA token with Google's API
func verifyRecaptcha(token, requestID, secretKey string) (bool, error) {
	if token == "" {
		return false, nil
	}

	// Skip verification if no secret key is provided (for development/testing)
	if secretKey == "" {
		log.Println("RecaptchaSecretKey not set, skipping verification")
		return true, nil
	}

	// Prepare form data
	formData := url.Values{
		"secret":   {secretKey},
		"response": {token},
	}

	// Send request to Google's reCAPTCHA API
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", formData)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var result struct {
		Success     bool      `json:"success"`
		Score       float64   `json:"score"`
		Action      string    `json:"action"`
		ChallengeTS time.Time `json:"challenge_ts"`
		Hostname    string    `json:"hostname"`
		ErrorCodes  []string  `json:"error-codes"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}

	// For v3 reCAPTCHA, we can also check the score
	// Score ranges from 0.0 (bot) to 1.0 (human)
	if result.Success && result.Score >= 0.5 {
		log.Printf("[%s] reCAPTCHA verification passed with score: %v", requestID, result.Score)
		metrics.IncrementRecaptchaVerification("success")
		return true, nil
	}
	log.Printf("[%s] reCAPTCHA verification failed with score: %v", requestID, result.Score)
	metrics.IncrementRecaptchaVerification("failure")
	return false, nil
}
