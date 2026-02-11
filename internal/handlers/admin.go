package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1it/go-submission-service/internal/config"
	"github.com/1it/go-submission-service/internal/database"
	"github.com/1it/go-submission-service/internal/models"
)

// AdminSubmissionsResponse represents the response for listing submissions
type AdminSubmissionsResponse struct {
	Submissions []*models.Submission `json:"submissions"`
	Total       int                  `json:"total"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
	Filters     map[string]string    `json:"filters"`
}

// AdminStatsResponse represents the response for admin statistics
type AdminStatsResponse struct {
	Overall     map[string]int            `json:"overall"`
	ByFormType  map[string]map[string]int `json:"by_form_type"`
	FormTypes   []string                  `json:"form_types"`
	GeneratedAt time.Time                 `json:"generated_at"`
}

// AdminHealthResponse represents enhanced health check response
type AdminHealthResponse struct {
	Status       string                 `json:"status"`
	Timestamp    time.Time              `json:"timestamp"`
	Service      string                 `json:"service"`
	Version      string                 `json:"version"`
	Uptime       string                 `json:"uptime"`
	Dependencies map[string]interface{} `json:"dependencies"`
	Metrics      map[string]interface{} `json:"metrics"`
}

var startTime = time.Now()

// AdminListSubmissionsHandler handles GET /admin/submissions
func AdminListSubmissionsHandler(repo *database.Repository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		page, limit, filters := adminParseListParams(r.URL.Query())

		submissions, err := repo.Submissions.List(filters)
		if err != nil {
			log.Printf("Error listing submissions: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		total, err := repo.Submissions.Count(filters)
		if err != nil {
			log.Printf("Error counting submissions: %v", err)
			total = len(submissions)
		}

		response := AdminSubmissionsResponse{
			Submissions: submissions,
			Total:       total,
			Page:        page,
			Limit:       limit,
			Filters: map[string]string{
				"status":     filters.Status,
				"form_type":  filters.FormType,
				"source":     filters.Source,
				"start_date": r.URL.Query().Get("start_date"),
				"end_date":   r.URL.Query().Get("end_date"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	}
}

func adminParseListParams(query interface{ Get(string) string }) (page, limit int, filters database.SubmissionFilters) {
	page = 1
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit = 50
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	filters = database.SubmissionFilters{
		Status:   query.Get("status"),
		FormType: query.Get("form_type"),
		Source:   query.Get("source"),
		Limit:    limit,
		Offset:   (page - 1) * limit,
		OrderBy:  "created_at",
		OrderDir: "DESC",
	}

	if startDate := query.Get("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			filters.StartDate = &parsed
		}
	}
	if endDate := query.Get("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			filters.EndDate = parseEndOfDay(parsed)
		}
	}

	return page, limit, filters
}

func parseEndOfDay(t time.Time) *time.Time {
	end := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	return &end
}

// AdminGetSubmissionHandler handles GET /api/v1/admin/submissions/{id}
func AdminGetSubmissionHandler(repo *database.Repository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from URL path
		id := extractSubmissionID(r.URL.Path, cfg.Security.Admin.PathPrefix)
		if id == "" {
			http.Error(w, "Submission ID required", http.StatusBadRequest)
			return
		}

		submission, err := repo.Submissions.GetByID(id)
		if err != nil {
			log.Printf("Error getting submission %s: %v", id, err)
			http.Error(w, "Submission not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(submission); err != nil {
			log.Printf("Failed to encode submission: %v", err)
		}
	}
}

// AdminUpdateSubmissionStatusHandler handles PATCH /api/v1/admin/submissions/{id}/status
func AdminUpdateSubmissionStatusHandler(repo *database.Repository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from URL path (remove /status suffix)
		id := extractSubmissionIDFromStatusPath(r.URL.Path, cfg.Security.Admin.PathPrefix)
		if id == "" {
			http.Error(w, "Submission ID required", http.StatusBadRequest)
			return
		}

		// Parse request body
		var req struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate status
		validStatuses := map[string]bool{
			"pending":   true,
			"processed": true,
			"failed":    true,
			"archived":  true,
		}
		if !validStatuses[req.Status] {
			http.Error(w, "Invalid status. Must be one of: pending, processed, failed, archived", http.StatusBadRequest)
			return
		}

		// Update status
		if err := repo.Submissions.UpdateStatus(id, req.Status); err != nil {
			log.Printf("Error updating submission %s status: %v", id, err)
			http.Error(w, "Failed to update submission status", http.StatusInternalServerError)
			return
		}

		// Return updated submission
		submission, err := repo.Submissions.GetByID(id)
		if err != nil {
			log.Printf("Error getting updated submission %s: %v", id, err)
			http.Error(w, "Submission updated but failed to retrieve", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(submission); err != nil {
			log.Printf("Failed to encode submission: %v", err)
		}
	}
}

// AdminDeleteSubmissionHandler handles DELETE /api/v1/admin/submissions/{id}
func AdminDeleteSubmissionHandler(repo *database.Repository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from URL path
		id := extractSubmissionID(r.URL.Path, cfg.Security.Admin.PathPrefix)
		if id == "" {
			http.Error(w, "Submission ID required", http.StatusBadRequest)
			return
		}

		// Delete submission
		if err := repo.Submissions.Delete(id); err != nil {
			log.Printf("Error deleting submission %s: %v", id, err)
			http.Error(w, "Failed to delete submission", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// AdminStatsHandler handles GET /admin/stats
func AdminStatsHandler(repo *database.Repository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get overall stats
		overallStats, err := repo.Submissions.GetStats()
		if err != nil {
			log.Printf("Error getting overall stats: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Get stats by form type
		statsByFormType, err := repo.Submissions.GetStatsByFormType()
		if err != nil {
			log.Printf("Error getting stats by form type: %v", err)
			statsByFormType = make(map[string]map[string]int)
		}

		// Get form types
		formTypes, err := repo.Submissions.GetFormTypes()
		if err != nil {
			log.Printf("Error getting form types: %v", err)
			formTypes = []string{}
		}

		response := AdminStatsResponse{
			Overall:     overallStats,
			ByFormType:  statsByFormType,
			FormTypes:   formTypes,
			GeneratedAt: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode stats response: %v", err)
		}
	}
}

// AdminHealthHandler handles GET /admin/health - enhanced health check
func AdminHealthHandler(repo *database.Repository, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check database health
		dbStatus := "ok"
		dbError := ""
		if err := repo.Submissions.HealthCheck(); err != nil {
			dbStatus = "error"
			dbError = err.Error()
		}

		// Get basic stats for metrics
		stats, err := repo.Submissions.GetStats()
		if err != nil {
			stats = map[string]int{"error": 1}
		}

		// Calculate uptime
		uptime := time.Since(startTime)

		// Build dependencies status
		dependencies := map[string]interface{}{
			"database": map[string]interface{}{
				"status": dbStatus,
				"type":   "sqlite",
				"path":   cfg.Database.Path,
			},
		}
		if dbError != "" {
			dependencies["database"].(map[string]interface{})["error"] = dbError
		}

		// Add email provider status
		dependencies["email"] = map[string]interface{}{
			"provider": cfg.Email.Provider,
			"status":   "configured",
		}

		// Overall status
		status := "ok"
		if dbStatus != "ok" {
			status = "degraded"
		}

		response := AdminHealthResponse{
			Status:       status,
			Timestamp:    time.Now(),
			Service:      "go-submission-service",
			Version:      "1.0.0",
			Uptime:       fmt.Sprintf("%.0fs", uptime.Seconds()),
			Dependencies: dependencies,
			Metrics: map[string]interface{}{
				"submissions":    stats,
				"uptime_seconds": uptime.Seconds(),
			},
		}

		// Set appropriate status code
		statusCode := http.StatusOK
		if status != "ok" {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode health response: %v", err)
		}
	}
}

// extractSubmissionID extracts the submission ID from URL paths like /api/v1/{pathPrefix}/submissions/{id}
func extractSubmissionID(path string, pathPrefix string) string {
	if pathPrefix == "" {
		pathPrefix = "admin"
	}
	prefix := fmt.Sprintf("/api/v1/%s/submissions/", pathPrefix)
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	// Remove any trailing slashes
	id = strings.TrimSuffix(id, "/")
	return id
}

// extractSubmissionIDFromStatusPath extracts the submission ID from status update paths like /api/v1/{pathPrefix}/submissions/{id}/status
func extractSubmissionIDFromStatusPath(path string, pathPrefix string) string {
	if pathPrefix == "" {
		pathPrefix = "admin"
	}
	prefix := fmt.Sprintf("/api/v1/%s/submissions/", pathPrefix)
	const suffix = "/status"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}
	// Remove prefix and suffix to get the ID
	id := strings.TrimPrefix(path, prefix)
	id = strings.TrimSuffix(id, suffix)
	return id
}
