package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/1it/go-submission-service/internal/config"
)

// AdminAuthMiddleware provides authentication and authorization for admin endpoints
type AdminAuthMiddleware struct {
	cfg *config.Config
}

// NewAdminAuthMiddleware creates a new admin authentication middleware
func NewAdminAuthMiddleware(cfg *config.Config) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{
		cfg: cfg,
	}
}

// Middleware returns an HTTP middleware that wraps handlers with admin authentication
func (m *AdminAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get configured path prefix
		pathPrefix := m.cfg.Security.Admin.PathPrefix
		if pathPrefix == "" {
			pathPrefix = "admin" // fallback to default
		}
		adminPath := "/api/v1/" + pathPrefix + "/"

		// Only apply admin authentication to admin endpoints
		if !strings.HasPrefix(r.URL.Path, adminPath) {
			next.ServeHTTP(w, r)
			return
		}

		m.authenticateAdmin(w, r, func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	})
}

// Authenticate wraps an HTTP handler with admin authentication
func (m *AdminAuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.authenticateAdmin(w, r, next)
	}
}

// authenticateAdmin performs the actual admin authentication logic
func (m *AdminAuthMiddleware) authenticateAdmin(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Check if admin endpoints are enabled
	if !m.cfg.Security.Admin.Enabled {
		log.Printf("Admin endpoint access denied: endpoints disabled")
		http.Error(w, "Admin endpoints are disabled", http.StatusNotFound)
		return
	}

	// Check HTTPS requirement in production
	if m.cfg.Security.Admin.RequireHTTPS && m.cfg.IsProductionMode() {
		if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
			log.Printf("Admin endpoint access denied: HTTPS required")
			http.Error(w, "HTTPS required for admin endpoints", http.StatusForbidden)
			return
		}
	}

	// Check IP whitelist if configured
	if len(m.cfg.Security.Admin.IPWhitelist) > 0 {
		clientIP := m.getClientIP(r)
		if !m.isIPWhitelisted(clientIP) {
			log.Printf("Admin endpoint access denied: IP %s not whitelisted", clientIP)
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
	}

	// Check API key
	if m.cfg.Security.Admin.APIKey == "" {
		log.Printf("Admin endpoint access denied: no API key configured")
		http.Error(w, "Admin endpoints not properly configured", http.StatusServiceUnavailable)
		return
	}

	apiKey := m.extractAPIKey(r)
	if apiKey == "" {
		log.Printf("Admin endpoint access denied: no API key provided")
		w.Header().Set("WWW-Authenticate", `Bearer realm="admin"`)
		http.Error(w, "API key required", http.StatusUnauthorized)
		return
	}

	if apiKey != m.cfg.Security.Admin.APIKey {
		log.Printf("Admin endpoint access denied: invalid API key")
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	// Authentication successful
	log.Printf("Admin endpoint access granted for %s", r.RemoteAddr)
	next(w, r)
}

// extractAPIKey extracts the API key from the request
// Supports both Authorization header (Bearer token) and X-API-Key header
func (m *AdminAuthMiddleware) extractAPIKey(r *http.Request) string {
	// Try Authorization header first (Bearer token)
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Try X-API-Key header
	return r.Header.Get("X-API-Key")
}

// getClientIP extracts the real client IP from the request
func (m *AdminAuthMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// isIPWhitelisted checks if the given IP is in the whitelist
func (m *AdminAuthMiddleware) isIPWhitelisted(clientIP string) bool {
	for _, allowedIP := range m.cfg.Security.Admin.IPWhitelist {
		// Support CIDR notation
		if strings.Contains(allowedIP, "/") {
			_, network, err := net.ParseCIDR(allowedIP)
			if err != nil {
				log.Printf("Invalid CIDR in IP whitelist: %s", allowedIP)
				continue
			}
			ip := net.ParseIP(clientIP)
			if ip != nil && network.Contains(ip) {
				return true
			}
		} else {
			// Exact IP match
			if clientIP == allowedIP {
				return true
			}
		}
	}
	return false
}
