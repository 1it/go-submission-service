package recaptcha

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// VerifyTurnstileToken validates a Cloudflare Turnstile token via /siteverify.
// It returns (isValid, human-readableErr, lowLevelErr).
func VerifyTurnstileToken(secretKey, token, remoteIP string) (bool, string, error) {
	if token == "" || secretKey == "" {
		return false, "missing token or secret", nil
	}

	form := url.Values{
		"secret":   {secretKey},
		"response": {token},
	}
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", form)
	if err != nil {
		return false, "failed to contact Turnstile API", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "failed to read Turnstile response", err
	}

	var result struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
		Hostname   string   `json:"hostname"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, "failed to parse Turnstile response", err
	}

	if !result.Success {
		msg := "Turnstile verification failed"
		if len(result.ErrorCodes) > 0 {
			msg = fmt.Sprintf("Turnstile error: %v", result.ErrorCodes)
		}
		return false, msg, nil
	}
	return true, "", nil
}
