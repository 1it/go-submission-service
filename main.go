package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1it/go-submission-service/internal/api"
	"github.com/1it/go-submission-service/internal/config"
	"github.com/1it/go-submission-service/internal/database"
	"github.com/1it/go-submission-service/internal/jobs"
	"github.com/1it/go-submission-service/internal/metrics"
	"github.com/1it/go-submission-service/internal/middleware"
)

// RequestLoggerMiddleware logs incoming HTTP requests
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())

		// Log request information
		log.Printf("[%s] [REQUEST] %s %s from %s", requestID, r.Method, r.URL.Path, r.RemoteAddr)

		// Create a custom response writer to capture status code
		crw := &customResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code
		}

		// Process the request using the wrapped handler
		next.ServeHTTP(crw, r)

		// Log response information
		duration := time.Since(startTime)
		log.Printf("[%s] [RESPONSE] %s %s - %d (%s) - took %v",
			requestID, r.Method, r.URL.Path, crw.statusCode,
			http.StatusText(crw.statusCode), duration)
	})
}

// CORSMiddleware adds CORS headers to responses based on configuration
func CORSMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Origin header from the request
			origin := r.Header.Get("Origin")

			// Check if the origin is allowed
			if origin != "" {
				for _, allowedOrigin := range cfg.Security.CORS.AllowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			// Set allowed methods from config
			if len(cfg.Security.CORS.AllowedMethods) > 0 {
				methods := ""
				for i, method := range cfg.Security.CORS.AllowedMethods {
					if i > 0 {
						methods += ", "
					}
					methods += method
				}
				w.Header().Set("Access-Control-Allow-Methods", methods)
			}

			// Set allowed headers from config
			if len(cfg.Security.CORS.AllowedHeaders) > 0 {
				headers := ""
				for i, header := range cfg.Security.CORS.AllowedHeaders {
					if i > 0 {
						headers += ", "
					}
					headers += header
				}
				w.Header().Set("Access-Control-Allow-Headers", headers)
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Process other requests
			next.ServeHTTP(w, r)
		})
	}
}

// CustomResponseWriter captures the status code of an HTTP response
type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and passes it to the underlying ResponseWriter
func (crw *customResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.ResponseWriter.WriteHeader(code)
}

func setupLogging() {
	// Configure logging with timestamps
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
	log.Println("Logging initialized")
}

func main() {
	// Set up logging first
	setupLogging()
	log.Println("Starting form submission service...")

	// Load configuration
	log.Println("Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded. Using port: %s, DB: %s, Email: %s",
		cfg.Server.Port, cfg.Database.Path, cfg.Email.Provider)

	// Initialize database
	log.Printf("Connecting to database at %s...", cfg.Database.Path)
	db, err := database.NewSQLiteDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create required tables
	log.Println("Initializing database schema...")
	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	// Set up API routes
	log.Println("Setting up API routes...")

	// Create repository instance
	repo := db.GetRepository()

	// Create and configure API router
	apiRouter := api.NewRouter(repo, cfg)
	apiRouter.RegisterRoutes()
	mux := apiRouter.GetMux()

	// Initialize security middleware
	rateLimitConfig := middleware.RateLimitMiddlewareConfig{
		Enabled:         cfg.Security.RateLimit.Enabled,
		RequestsPerMin:  cfg.Security.RateLimit.RequestsPerMin,
		BurstSize:       cfg.Security.RateLimit.BurstSize,
		CleanupInterval: 5 * time.Minute,
	}
	rateLimiter := middleware.NewRateLimitMiddleware(rateLimitConfig)

	// Build middleware chain: metrics -> security headers -> rate limiting -> API key auth -> logging -> CORS
	// The order is important: metrics should be the outermost middleware to capture all requests
	var handler http.Handler = mux

	// Apply middleware in reverse order (innermost to outermost)
	handler = CORSMiddleware(cfg)(handler)
	handler = RequestLoggerMiddleware(handler)
	// Use comprehensive admin authentication middleware
	adminAuth := middleware.NewAdminAuthMiddleware(cfg)
	handler = adminAuth.Middleware(handler)
	handler = rateLimiter.Middleware(handler)

	// Apply security headers if enabled
	if cfg.Security.SecurityHeaders.Enabled {
		handler = middleware.SecurityHeadersMiddleware(handler)
		log.Println("Security headers middleware enabled")
	}

	// Metrics should be the outermost middleware
	handler = metrics.PrometheusMiddleware(handler)

	log.Printf("Routes configured with enhanced security middleware: rate limiting (%d req/min), API key auth, security headers", cfg.Security.RateLimit.RequestsPerMin)

	// Use port from config
	port := cfg.Server.Port
	if port == "" { // Fallback if config somehow missed it
		port = "8080"
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background job processor
	log.Println("Starting background job processor...")
	jobProcessor := jobs.NewJobProcessor(repo, cfg)
	go jobProcessor.Start(ctx)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")

	// Cancel context to stop job processor
	cancel()

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown server gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server shutdown complete")
	}
}
