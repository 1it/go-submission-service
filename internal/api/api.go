package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/1it/go-submission-service/internal/config"
	"github.com/1it/go-submission-service/internal/database"
	"github.com/1it/go-submission-service/internal/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Router handles all API route registration
type Router struct {
	repo *database.Repository
	cfg  *config.Config
	mux  *http.ServeMux
}

// NewRouter creates a new API router
func NewRouter(repo *database.Repository, cfg *config.Config) *Router {
	return &Router{
		repo: repo,
		cfg:  cfg,
		mux:  http.NewServeMux(),
	}
}

// RegisterRoutes registers all API routes with consistent v1 prefix
func (r *Router) RegisterRoutes() {
	// Core API endpoints
	r.mux.HandleFunc("/api/v1/health", handlers.HealthHandler)
	log.Println("Registered route: /api/v1/health")

	r.mux.HandleFunc("/api/v1/submit", handlers.SubmitHandler(r.repo, r.cfg))
	log.Println("Registered route: /api/v1/submit")

	// Legacy signup endpoint for backward compatibility
	r.mux.HandleFunc("/api/v1/signup", handlers.LegacySignupHandler(r.repo, r.cfg))
	log.Println("Registered route: /api/v1/signup (legacy compatibility)")

	// Metrics endpoint
	r.mux.Handle("/api/v1/metrics", promhttp.Handler())
	log.Println("Registered route: /api/v1/metrics")

	// Admin endpoints (only if enabled)
	if r.cfg.Security.Admin.Enabled {
		r.registerAdminRoutes()
		log.Println("Admin endpoints enabled and registered")
	} else {
		log.Println("Admin endpoints disabled - not registering admin routes")
	}
}

// registerAdminRoutes registers all admin API endpoints with consistent routing
func (r *Router) registerAdminRoutes() {
	// Use configured path prefix for admin endpoints
	pathPrefix := r.cfg.Security.Admin.PathPrefix
	if pathPrefix == "" {
		pathPrefix = "admin" // fallback to default
	}

	// Admin submissions endpoints
	r.mux.HandleFunc("/api/v1/"+pathPrefix+"/submissions", r.handleSubmissions)
	r.mux.HandleFunc("/api/v1/"+pathPrefix+"/submissions/", r.handleSubmissionsWithID)

	// Admin stats and health endpoints
	r.mux.HandleFunc("/api/v1/"+pathPrefix+"/stats", handlers.AdminStatsHandler(r.repo, r.cfg))
	r.mux.HandleFunc("/api/v1/"+pathPrefix+"/health", handlers.AdminHealthHandler(r.repo, r.cfg))

	log.Printf("Registered admin routes: /api/v1/%s/*", pathPrefix)
}

// handleSubmissions handles /api/v1/admin/submissions (list only)
func (r *Router) handleSubmissions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	handlers.AdminListSubmissionsHandler(r.repo, r.cfg)(w, req)
}

// handleSubmissionsWithID handles /api/v1/{pathPrefix}/submissions/{id} and /api/v1/{pathPrefix}/submissions/{id}/status
func (r *Router) handleSubmissionsWithID(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// Get configured path prefix
	pathPrefix := r.cfg.Security.Admin.PathPrefix
	if pathPrefix == "" {
		pathPrefix = "admin" // fallback to default
	}
	basePath := "/api/v1/" + pathPrefix + "/submissions/"

	// Remove the base path to get the ID part
	idPath := strings.TrimPrefix(path, basePath)

	// Handle empty ID (redirect to list)
	if idPath == "" {
		r.handleSubmissions(w, req)
		return
	}

	// Check if this is a status update request
	if strings.HasSuffix(idPath, "/status") {
		if req.Method != http.MethodPatch {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handlers.AdminUpdateSubmissionStatusHandler(r.repo, r.cfg)(w, req)
		return
	}

	// Handle individual submission requests
	switch req.Method {
	case http.MethodGet:
		handlers.AdminGetSubmissionHandler(r.repo, r.cfg)(w, req)
	case http.MethodDelete:
		handlers.AdminDeleteSubmissionHandler(r.repo, r.cfg)(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetMux returns the configured HTTP mux
func (r *Router) GetMux() *http.ServeMux {
	return r.mux
}
