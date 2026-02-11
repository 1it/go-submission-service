package metrics

import (
	"net/http"
	"strconv"
	"time"
)

// PrometheusMiddleware is middleware that instruments HTTP requests with Prometheus metrics
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture the status code
		rww := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code if not set
		}

		// Process the request using the wrapped handler
		next.ServeHTTP(rww, r)

		// Record metrics after the request is processed
		duration := time.Since(start).Seconds()
		path := r.URL.Path
		method := r.Method
		statusCode := strconv.Itoa(rww.statusCode)

		// Increment request counter
		HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()

		// Observe request duration
		HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	})
}

// responseWriterWrapper is a wrapper around http.ResponseWriter to capture the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and passes it to the underlying ResponseWriter
func (rww *responseWriterWrapper) WriteHeader(code int) {
	rww.statusCode = code
	rww.ResponseWriter.WriteHeader(code)
}

// WrapHandler wraps an http.HandlerFunc with Prometheus instrumentation
func WrapHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture the status code
		rww := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status code if not set
		}

		// Process the request
		handlerFunc(rww, r)

		// Record metrics after the request is processed
		duration := time.Since(start).Seconds()
		path := r.URL.Path
		method := r.Method
		statusCode := strconv.Itoa(rww.statusCode)

		// Increment request counter
		HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()

		// Observe request duration
		HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
