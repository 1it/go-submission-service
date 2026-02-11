package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Standard HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests received",
		},
		[]string{"method", "path", "status_code"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Application-specific metrics
	SignupRequestsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "signup_requests_total",
			Help: "Total number of signup requests",
		},
	)

	SignupSuccessTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "signup_success_total",
			Help: "Total number of successful signups",
		},
	)

	SignupFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "signup_failures_total",
			Help: "Total number of failed signups",
		},
		[]string{"reason"},
	)

	RecaptchaVerificationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "recaptcha_verifications_total",
			Help: "Total number of reCAPTCHA verification attempts",
		},
		[]string{"result"},
	)
)

// IncrementSignupFailure increments the signup failures counter with the specified reason
func IncrementSignupFailure(reason string) {
	SignupFailuresTotal.WithLabelValues(reason).Inc()
}

// IncrementRecaptchaVerification increments the reCAPTCHA verifications counter with the specified result
func IncrementRecaptchaVerification(result string) {
	RecaptchaVerificationsTotal.WithLabelValues(result).Inc()
}
