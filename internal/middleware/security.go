package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/1it/go-submission-service/internal/config"
)

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// RateLimiter represents a rate limiter for a specific client
type RateLimiter struct {
	requests []time.Time
	mutex    sync.Mutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make([]time.Time, 0),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request is allowed under the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove old requests outside the window
	validRequests := make([]time.Time, 0)
	for _, req := range rl.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	rl.requests = validRequests

	// Check if we're under the limit
	if len(rl.requests) >= rl.limit {
		return false
	}

	// Add current request
	rl.requests = append(rl.requests, now)
	return true
}

// RateLimitMiddlewareConfig holds rate limiting middleware configuration
type RateLimitMiddlewareConfig struct {
	Enabled         bool
	RequestsPerMin  int
	BurstSize       int
	CleanupInterval time.Duration
}

// RateLimitMiddleware provides rate limiting functionality
type RateLimitMiddleware struct {
	limiters map[string]*RateLimiter
	mutex    sync.RWMutex
	config   RateLimitMiddlewareConfig
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(config RateLimitMiddlewareConfig) *RateLimitMiddleware {
	rlm := &RateLimitMiddleware{
		limiters: make(map[string]*RateLimiter),
		config:   config,
	}

	// Start cleanup goroutine
	go rlm.cleanup()

	return rlm
}

// getClientIP extracts the client IP from the request
func (rlm *RateLimitMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	if ip := strings.Split(r.RemoteAddr, ":"); len(ip) > 0 {
		return ip[0]
	}

	return r.RemoteAddr
}

// getLimiter gets or creates a rate limiter for a client
func (rlm *RateLimitMiddleware) getLimiter(clientIP string) *RateLimiter {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()

	if limiter, exists := rlm.limiters[clientIP]; exists {
		return limiter
	}

	limiter := NewRateLimiter(rlm.config.RequestsPerMin, time.Minute)
	rlm.limiters[clientIP] = limiter
	return limiter
}

// cleanup removes old limiters periodically
func (rlm *RateLimitMiddleware) cleanup() {
	ticker := time.NewTicker(rlm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rlm.mutex.Lock()
		// Remove limiters that haven't been used recently
		cutoff := time.Now().Add(-rlm.config.CleanupInterval)
		for ip, limiter := range rlm.limiters {
			limiter.mutex.Lock()
			if len(limiter.requests) == 0 || (len(limiter.requests) > 0 && limiter.requests[len(limiter.requests)-1].Before(cutoff)) {
				delete(rlm.limiters, ip)
			}
			limiter.mutex.Unlock()
		}
		rlm.mutex.Unlock()
	}
}

// Middleware returns the rate limiting middleware handler
func (rlm *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rlm.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		clientIP := rlm.getClientIP(r)
		limiter := rlm.getLimiter(clientIP)

		if !limiter.Allow() {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rlm.config.RequestsPerMin))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// Add rate limit headers to successful requests
		limiter.mutex.Lock()
		remaining := rlm.config.RequestsPerMin - len(limiter.requests)
		limiter.mutex.Unlock()

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rlm.config.RequestsPerMin))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))

		next.ServeHTTP(w, r)
	})
}

// APIKeyAuthMiddleware provides API key authentication for admin endpoints
func APIKeyAuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is an admin endpoint request
			if strings.HasPrefix(r.URL.Path, "/api/v1/admin/") {
				// Check for API key in header
				apiKey := r.Header.Get("X-API-Key")
				if apiKey == "" {
					// Also check Authorization header with Bearer token
					auth := r.Header.Get("Authorization")
					if strings.HasPrefix(auth, "Bearer ") {
						apiKey = strings.TrimPrefix(auth, "Bearer ")
					}
				}

				if apiKey == "" {
					w.Header().Set("WWW-Authenticate", `Bearer realm="admin"`)
					http.Error(w, "API key required for admin endpoints", http.StatusUnauthorized)
					return
				}

				// Validate API key
				if cfg.Security.Admin.APIKey == "" || apiKey != cfg.Security.Admin.APIKey {
					http.Error(w, "Invalid API key", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
