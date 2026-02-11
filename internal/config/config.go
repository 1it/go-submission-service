package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds the main application configuration
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server"`

	// Database configuration
	Database DatabaseConfig `json:"database"`

	// Email service configuration
	Email EmailConfig `json:"email"`

	// Form processing configuration
	Form FormConfig `json:"form"`

	// Security configuration
	Security SecurityConfig `json:"security"`

	// Retry configuration
	Retry RetryConfig `json:"retry"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port string `json:"port" env:"PORT"`
	Host string `json:"host" env:"HOST"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `json:"path" env:"DB_PATH"`
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	Provider   string           `json:"provider" env:"EMAIL_PROVIDER"` // "smtp" or "mailersend"
	SMTP       SMTPConfig       `json:"smtp,omitempty"`
	MailerSend MailerSendConfig `json:"mailersend,omitempty"`
	Templates  TemplateConfig   `json:"templates"`
	AdminEmail string           `json:"admin_email" env:"ADMIN_EMAIL"` // Optional admin email for notifications
}

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string `json:"host" env:"SMTP_HOST"`
	Port     string `json:"port" env:"SMTP_PORT"`
	Username string `json:"username" env:"SMTP_USERNAME"`
	Password string `json:"password" env:"SMTP_PASSWORD"`
	From     string `json:"from" env:"SMTP_FROM"`
}

// MailerSendConfig holds MailerSend API configuration
type MailerSendConfig struct {
	APIKey    string `json:"api_key" env:"MAILERSEND_API_KEY"`
	FromEmail string `json:"from_email" env:"MAILERSEND_FROM_EMAIL"`
	FromName  string `json:"from_name" env:"MAILERSEND_FROM_NAME"`
}

// TemplateConfig holds email template configuration
type TemplateConfig struct {
	Directory       string            `json:"directory" env:"TEMPLATES_DIR"`
	DefaultTemplate string            `json:"default_template" env:"DEFAULT_TEMPLATE"`
	Subject         string            `json:"subject" env:"EMAIL_SUBJECT"`
	Variables       map[string]string `json:"variables,omitempty"`
}

// FormConfig holds form processing configuration
type FormConfig struct {
	RequiredFields      []string                  `json:"required_fields"`
	EmailField          string                    `json:"email_field" env:"EMAIL_FIELD"`
	ValidationRules     map[string]ValidationRule `json:"validation_rules,omitempty"`
	SuccessMessage      string                    `json:"success_message" env:"SUCCESS_MESSAGE"`
	MaxRequestSize      int64                     `json:"max_request_size" env:"MAX_REQUEST_SIZE"`
	SanitizationOptions *SanitizationOptions      `json:"sanitization_options,omitempty"`
	CrossFieldRules     []CrossFieldRule          `json:"cross_field_rules,omitempty"`
}

// ValidationRule defines validation rules for form fields
type ValidationRule struct {
	Type    string `json:"type"` // email, regex, length, etc.
	Pattern string `json:"pattern,omitempty"`
	Min     int    `json:"min,omitempty"`
	Max     int    `json:"max,omitempty"`
	Message string `json:"message,omitempty"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	CORS            CORSConfig            `json:"cors"`
	ReCAPTCHA       ReCAPTCHAConfig       `json:"recaptcha"`
	Turnstile       TurnstileConfig       `json:"turnstile"`
	RateLimit       RateLimitConfig       `json:"rate_limit"`
	Admin           AdminSecurityConfig   `json:"admin"`
	SecurityHeaders SecurityHeadersConfig `json:"security_headers"`
}

// TurnstileConfig holds Turnstile configuration
type TurnstileConfig struct {
	Enabled   bool   `json:"enabled" env:"TURNSTILE_ENABLED"`
	SecretKey string `json:"secret_key" env:"TURNSTILE_SECRET_KEY"`
	SiteKey   string `json:"site_key" env:"TURNSTILE_SITE_KEY"`
}

// AdminSecurityConfig holds admin-specific security configuration
type AdminSecurityConfig struct {
	Enabled      bool     `json:"enabled" env:"ADMIN_ENDPOINTS_ENABLED"`
	APIKey       string   `json:"api_key" env:"ADMIN_API_KEY"`
	PathPrefix   string   `json:"path_prefix" env:"ADMIN_PATH_PREFIX"`
	IPWhitelist  []string `json:"ip_whitelist" env:"ADMIN_IP_WHITELIST"`
	RequireHTTPS bool     `json:"require_https" env:"ADMIN_REQUIRE_HTTPS"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool `json:"enabled" env:"RATE_LIMIT_ENABLED"`
	RequestsPerMin int  `json:"requests_per_minute" env:"RATE_LIMIT_REQUESTS_PER_MINUTE"`
	BurstSize      int  `json:"burst_size" env:"RATE_LIMIT_BURST_SIZE"`
}

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	Enabled bool `json:"enabled" env:"SECURITY_HEADERS_ENABLED"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins" env:"CORS_ALLOWED_ORIGINS"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// ReCAPTCHAConfig holds reCAPTCHA configuration
type ReCAPTCHAConfig struct {
	Enabled    bool    `json:"enabled" env:"RECAPTCHA_ENABLED"`
	SecretKey  string  `json:"secret_key" env:"RECAPTCHA_SECRET_KEY"`
	SiteKey    string  `json:"site_key" env:"RECAPTCHA_SITE_KEY"`
	ActionName string  `json:"action_name" env:"RECAPTCHA_ACTION_NAME"`
	MinScore   float32 `json:"min_score" env:"RECAPTCHA_MIN_SCORE"`

	// Enterprise configuration
	EnterpriseEnabled bool   `json:"enterprise_enabled" env:"RECAPTCHA_ENTERPRISE_ENABLED"`
	ProjectID         string `json:"project_id" env:"RECAPTCHA_PROJECT_ID"`
}

// RetryConfig holds retry-related configuration
type RetryConfig struct {
	MaxAttempts   int     `json:"max_attempts" env:"RETRY_MAX_ATTEMPTS"`
	InitialDelay  int     `json:"initial_delay_ms" env:"RETRY_INITIAL_DELAY_MS"` // in milliseconds
	MaxDelay      int     `json:"max_delay_ms" env:"RETRY_MAX_DELAY_MS"`         // in milliseconds
	BackoffFactor float64 `json:"backoff_factor" env:"RETRY_BACKOFF_FACTOR"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	// Try to load environment variables from .env file
	// This is done first so that environment variables take precedence
	_ = LoadEnvFile(".env")

	// Also try environment-specific .env files
	env := os.Getenv("ENVIRONMENT")
	if env != "" {
		_ = LoadEnvFile(".env." + env)
	}

	// Default configuration with generic, non-project-specific values
	config := Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Path: "./data/submissions.db",
		},
		Email: EmailConfig{
			Provider: "smtp",
			SMTP: SMTPConfig{
				Host: "localhost",
				Port: "1025", // Default to MailHog for development
				From: "noreply@example.com",
			},
			MailerSend: MailerSendConfig{
				FromEmail: "noreply@example.com",
				FromName:  "Form Submission Service",
			},
			Templates: TemplateConfig{
				Directory:       "./templates",
				DefaultTemplate: "default.html",
				Subject:         "Form Submission Received",
			},
		},
		Form: FormConfig{
			RequiredFields: []string{"email"},
			EmailField:     "email",
			SuccessMessage: "Thank you for your submission!",
			MaxRequestSize: 1024 * 1024, // 1MB
			SanitizationOptions: &SanitizationOptions{
				DisableHTMLEscaping:   false,
				DisableHTMLTagRemoval: false,
			},
			ValidationRules: map[string]ValidationRule{
				"email": {
					Type:    "email",
					Message: "Please enter a valid email address",
				},
				"name": {
					Type:    "length",
					Min:     2,
					Max:     100,
					Message: "Name must be between 2 and 100 characters",
				},
				"message": {
					Type:    "length",
					Max:     5000,
					Message: "Message must be less than 5000 characters",
				},
			},
		},
		Security: SecurityConfig{
			CORS: CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "OPTIONS"},
				AllowedHeaders: []string{"Content-Type", "Accept"},
			},
			ReCAPTCHA: ReCAPTCHAConfig{
				Enabled:    false,
				ActionName: "form_submit",
				MinScore:   0.5,
			},
			RateLimit: RateLimitConfig{
				Enabled:        true,
				RequestsPerMin: 60,
				BurstSize:      10,
			},
			Admin: AdminSecurityConfig{
				Enabled:      false,      // Disabled by default for security
				APIKey:       "",         // Must be set via environment variable
				PathPrefix:   "mgmt",     // Less obvious than "admin"
				IPWhitelist:  []string{}, // Empty means no IP restrictions
				RequireHTTPS: true,       // Require HTTPS in production
			},
			SecurityHeaders: SecurityHeadersConfig{
				Enabled: true,
			},
		},
		Retry: RetryConfig{
			MaxAttempts:   3,
			InitialDelay:  1000,  // 1 second
			MaxDelay:      30000, // 30 seconds
			BackoffFactor: 2.0,
		},
	}

	// Load configuration from file
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = "config.json"
	}

	if file, err := os.Open(configPath); err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&config); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	// Override with environment variables
	loadEnvVars(&config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// loadEnvVars loads environment variables into the configuration
func loadEnvVars(config *Config) {
	loadServerEnv(config)
	loadDatabaseEnv(config)
	loadEmailEnv(config)
	loadFormEnv(config)
	loadSecurityEnv(config)
	loadRetryEnv(config)
}

func loadServerEnv(config *Config) {
	if port := os.Getenv("PORT"); port != "" {
		config.Server.Port = port
	}
	if host := os.Getenv("HOST"); host != "" {
		config.Server.Host = host
	}
}

func loadDatabaseEnv(config *Config) {
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}
}

func loadEmailEnv(config *Config) {
	if provider := os.Getenv("EMAIL_PROVIDER"); provider != "" {
		config.Email.Provider = provider
	}
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		config.Email.SMTP.Host = smtpHost
	}
	if smtpPort := os.Getenv("SMTP_PORT"); smtpPort != "" {
		config.Email.SMTP.Port = smtpPort
	}
	if smtpUsername := os.Getenv("SMTP_USERNAME"); smtpUsername != "" {
		config.Email.SMTP.Username = smtpUsername
	}
	if smtpPassword := os.Getenv("SMTP_PASSWORD"); smtpPassword != "" {
		config.Email.SMTP.Password = smtpPassword
	}
	if smtpFrom := os.Getenv("SMTP_FROM"); smtpFrom != "" {
		config.Email.SMTP.From = smtpFrom
	}
	if apiKey := os.Getenv("MAILERSEND_API_KEY"); apiKey != "" {
		config.Email.MailerSend.APIKey = apiKey
	}
	if fromEmail := os.Getenv("MAILERSEND_FROM_EMAIL"); fromEmail != "" {
		config.Email.MailerSend.FromEmail = fromEmail
	}
	if fromName := os.Getenv("MAILERSEND_FROM_NAME"); fromName != "" {
		config.Email.MailerSend.FromName = fromName
	}
	if templatesDir := os.Getenv("TEMPLATES_DIR"); templatesDir != "" {
		config.Email.Templates.Directory = templatesDir
	}
	if defaultTemplate := os.Getenv("DEFAULT_TEMPLATE"); defaultTemplate != "" {
		config.Email.Templates.DefaultTemplate = defaultTemplate
	}
	if subject := os.Getenv("EMAIL_SUBJECT"); subject != "" {
		config.Email.Templates.Subject = subject
	}
	if adminEmail := os.Getenv("ADMIN_EMAIL"); adminEmail != "" {
		config.Email.AdminEmail = adminEmail
	}
}

func loadFormEnv(config *Config) {
	if emailField := os.Getenv("EMAIL_FIELD"); emailField != "" {
		config.Form.EmailField = emailField
	}
	if successMessage := os.Getenv("SUCCESS_MESSAGE"); successMessage != "" {
		config.Form.SuccessMessage = successMessage
	}
	if maxRequestSize := os.Getenv("MAX_REQUEST_SIZE"); maxRequestSize != "" {
		if size, err := strconv.ParseInt(maxRequestSize, 10, 64); err == nil && size > 0 {
			config.Form.MaxRequestSize = size
		}
	}
	if os.Getenv("DISABLE_HTML_ESCAPING") == "true" {
		if config.Form.SanitizationOptions == nil {
			config.Form.SanitizationOptions = &SanitizationOptions{}
		}
		config.Form.SanitizationOptions.DisableHTMLEscaping = true
	}
	if os.Getenv("DISABLE_HTML_TAG_REMOVAL") == "true" {
		if config.Form.SanitizationOptions == nil {
			config.Form.SanitizationOptions = &SanitizationOptions{}
		}
		config.Form.SanitizationOptions.DisableHTMLTagRemoval = true
	}
}

func loadSecurityEnv(config *Config) {
	loadCORSEnv(config)
	loadTurnstileEnv(config)
	loadRecaptchaEnv(config)
	loadRateLimitEnv(config)
	loadAdminEnv(config)
	loadSecurityHeadersEnv(config)
}

func loadCORSEnv(config *Config) {
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		config.Security.CORS.AllowedOrigins = parseCommaSeparated(origins)
	}
	if methods := os.Getenv("CORS_ALLOWED_METHODS"); methods != "" {
		config.Security.CORS.AllowedMethods = parseCommaSeparated(methods)
	}
	if headers := os.Getenv("CORS_ALLOWED_HEADERS"); headers != "" {
		config.Security.CORS.AllowedHeaders = parseCommaSeparated(headers)
	}
}

func loadTurnstileEnv(config *Config) {
	if os.Getenv("TURNSTILE_ENABLED") == "true" {
		config.Security.Turnstile.Enabled = true
	}
	if secretKey := os.Getenv("TURNSTILE_SECRET_KEY"); secretKey != "" {
		config.Security.Turnstile.SecretKey = secretKey
	}
	if siteKey := os.Getenv("TURNSTILE_SITE_KEY"); siteKey != "" {
		config.Security.Turnstile.SiteKey = siteKey
	}
}

func loadRecaptchaEnv(config *Config) {
	if os.Getenv("RECAPTCHA_ENABLED") == "true" {
		config.Security.ReCAPTCHA.Enabled = true
	}
	if secretKey := os.Getenv("RECAPTCHA_SECRET_KEY"); secretKey != "" {
		config.Security.ReCAPTCHA.SecretKey = secretKey
	}
	if siteKey := os.Getenv("RECAPTCHA_SITE_KEY"); siteKey != "" {
		config.Security.ReCAPTCHA.SiteKey = siteKey
	}
	if actionName := os.Getenv("RECAPTCHA_ACTION_NAME"); actionName != "" {
		config.Security.ReCAPTCHA.ActionName = actionName
	}
	if os.Getenv("RECAPTCHA_ENTERPRISE_ENABLED") == "true" {
		config.Security.ReCAPTCHA.EnterpriseEnabled = true
	}
	if projectID := os.Getenv("RECAPTCHA_PROJECT_ID"); projectID != "" {
		config.Security.ReCAPTCHA.ProjectID = projectID
	}
}

func loadRateLimitEnv(config *Config) {
	if os.Getenv("RATE_LIMIT_ENABLED") == "false" {
		config.Security.RateLimit.Enabled = false
	}
	if requestsPerMin := os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"); requestsPerMin != "" {
		if rpm, err := strconv.Atoi(requestsPerMin); err == nil && rpm > 0 {
			config.Security.RateLimit.RequestsPerMin = rpm
		}
	}
	if burstSize := os.Getenv("RATE_LIMIT_BURST_SIZE"); burstSize != "" {
		if burst, err := strconv.Atoi(burstSize); err == nil && burst > 0 {
			config.Security.RateLimit.BurstSize = burst
		}
	}
}

func loadAdminEnv(config *Config) {
	if os.Getenv("ADMIN_ENDPOINTS_ENABLED") == "true" {
		config.Security.Admin.Enabled = true
	}
	if adminAPIKey := os.Getenv("ADMIN_API_KEY"); adminAPIKey != "" {
		config.Security.Admin.APIKey = adminAPIKey
	}
	if adminPathPrefix := os.Getenv("ADMIN_PATH_PREFIX"); adminPathPrefix != "" {
		config.Security.Admin.PathPrefix = adminPathPrefix
	}
	if adminIPWhitelist := os.Getenv("ADMIN_IP_WHITELIST"); adminIPWhitelist != "" {
		config.Security.Admin.IPWhitelist = parseCommaSeparated(adminIPWhitelist)
	}
	if os.Getenv("ADMIN_REQUIRE_HTTPS") == "false" {
		config.Security.Admin.RequireHTTPS = false
	}
}

func loadSecurityHeadersEnv(config *Config) {
	if os.Getenv("SECURITY_HEADERS_ENABLED") == "false" {
		config.Security.SecurityHeaders.Enabled = false
	}
}

func loadRetryEnv(config *Config) {
	if maxAttempts := os.Getenv("RETRY_MAX_ATTEMPTS"); maxAttempts != "" {
		if attempts, err := strconv.Atoi(maxAttempts); err == nil && attempts > 0 {
			config.Retry.MaxAttempts = attempts
		}
	}
	if initialDelay := os.Getenv("RETRY_INITIAL_DELAY_MS"); initialDelay != "" {
		if delay, err := strconv.Atoi(initialDelay); err == nil && delay > 0 {
			config.Retry.InitialDelay = delay
		}
	}
	if maxDelay := os.Getenv("RETRY_MAX_DELAY_MS"); maxDelay != "" {
		if delay, err := strconv.Atoi(maxDelay); err == nil && delay > 0 {
			config.Retry.MaxDelay = delay
		}
	}
	if backoffFactor := os.Getenv("RETRY_BACKOFF_FACTOR"); backoffFactor != "" {
		if factor, err := strconv.ParseFloat(backoffFactor, 64); err == nil && factor > 0 {
			config.Retry.BackoffFactor = factor
		}
	}
}

// Validate validates the configuration and returns an error if invalid
func (c *Config) Validate() error {
	var errs []string
	errs = append(errs, validateServerConfig(c)...)
	errs = append(errs, validateDatabaseConfig(c)...)
	errs = append(errs, validateEmailConfig(c)...)
	errs = append(errs, validateFormConfig(c)...)
	errs = append(errs, validateSecurityConfig(c)...)
	if len(errs) > 0 {
		return fmt.Errorf("configuration validation errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func validateServerConfig(c *Config) []string {
	var errs []string
	if c.Server.Port == "" {
		errs = append(errs, "server port is required")
	}
	return errs
}

func validateDatabaseConfig(c *Config) []string {
	var errs []string
	if c.Database.Path == "" {
		errs = append(errs, "database path is required")
	}
	return errs
}

func validateEmailConfig(c *Config) []string {
	var errs []string
	if c.Email.Provider != "smtp" && c.Email.Provider != "mailersend" {
		errs = append(errs, "email provider must be 'smtp' or 'mailersend'")
		return errs
	}
	if c.Email.Provider == "smtp" {
		if c.Email.SMTP.Host == "" {
			errs = append(errs, "SMTP host is required when using SMTP provider")
		}
		if c.Email.SMTP.Port == "" {
			errs = append(errs, "SMTP port is required when using SMTP provider")
		}
		if c.Email.SMTP.From == "" {
			errs = append(errs, "SMTP from address is required when using SMTP provider")
		}
	}
	if c.Email.Provider == "mailersend" {
		if c.Email.MailerSend.APIKey == "" {
			errs = append(errs, "MailerSend API key is required when using MailerSend provider")
		}
		if c.Email.MailerSend.FromEmail == "" {
			errs = append(errs, "MailerSend from email is required when using MailerSend provider")
		}
	}
	if c.Email.Templates.Directory == "" {
		errs = append(errs, "templates directory is required")
	}
	if c.Email.Templates.DefaultTemplate == "" {
		errs = append(errs, "default template is required")
	}
	if c.Email.Templates.Subject == "" {
		errs = append(errs, "email subject is required")
	}
	return errs
}

func validateFormConfig(c *Config) []string {
	var errs []string
	if c.Form.EmailField == "" {
		errs = append(errs, "email field name is required")
	}
	if c.Form.SuccessMessage == "" {
		errs = append(errs, "success message is required")
	}
	return errs
}

func validateSecurityConfig(c *Config) []string {
	var errs []string
	if c.Security.ReCAPTCHA.Enabled {
		if c.Security.ReCAPTCHA.SecretKey == "" {
			errs = append(errs, "reCAPTCHA secret key is required when reCAPTCHA is enabled")
		}
		if c.Security.ReCAPTCHA.SiteKey == "" {
			errs = append(errs, "reCAPTCHA site key is required when reCAPTCHA is enabled")
		}
	}
	if c.Security.ReCAPTCHA.EnterpriseEnabled {
		if c.Security.ReCAPTCHA.ProjectID == "" {
			errs = append(errs, "reCAPTCHA project ID is required when reCAPTCHA Enterprise is enabled")
		}
		if c.Security.ReCAPTCHA.SiteKey == "" {
			errs = append(errs, "reCAPTCHA site key is required when reCAPTCHA Enterprise is enabled")
		}
	}
	if c.Security.Turnstile.Enabled {
		if c.Security.Turnstile.SecretKey == "" {
			errs = append(errs, "Turnstile secret key is required when Turnstile is enabled")
		}
		if c.Security.Turnstile.SiteKey == "" {
			errs = append(errs, "Turnstile site key is required when Turnstile is enabled")
		}
	}
	if (c.Security.ReCAPTCHA.Enabled || c.Security.ReCAPTCHA.EnterpriseEnabled) && c.Security.Turnstile.Enabled {
		errs = append(errs, "Only one CAPTCHA system (reCAPTCHA or Turnstile) can be enabled at a time")
	}
	return errs
}

// GetDatabaseConfig returns the database configuration
func (c *Config) GetDatabaseConfig() DatabaseConfig {
	return c.Database
}

// GetEmailConfig returns the email configuration
func (c *Config) GetEmailConfig() EmailConfig {
	return c.Email
}

// GetSecurityConfig returns the security configuration
func (c *Config) GetSecurityConfig() SecurityConfig {
	return c.Security
}

// GetFormConfig returns the form configuration
func (c *Config) GetFormConfig() FormConfig {
	return c.Form
}

// IsProductionMode returns true if the service is running in production mode
func (c *Config) IsProductionMode() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "production" || env == "prod"
}

// IsDevelopmentMode returns true if the service is running in development mode
func (c *Config) IsDevelopmentMode() bool {
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "development" || env == "dev" || env == ""
}

// parseCommaSeparated parses a comma-separated string into a slice of strings
func parseCommaSeparated(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// GenerateConfigDocs generates documentation for all configuration options
func GenerateConfigDocs() string {
	var docs strings.Builder

	docs.WriteString("# Configuration Documentation\n\n")
	docs.WriteString("This service can be configured using environment variables or a config.json file.\n\n")

	docs.WriteString("## Server Configuration\n\n")
	docs.WriteString("| Environment Variable | Default | Description |\n")
	docs.WriteString("|---------------------|---------|-------------|\n")
	docs.WriteString("| `PORT` | `8080` | Server port |\n")
	docs.WriteString("| `HOST` | `0.0.0.0` | Server host |\n\n")

	docs.WriteString("## Database Configuration\n\n")
	docs.WriteString("| Environment Variable | Default | Description |\n")
	docs.WriteString("|---------------------|---------|-------------|\n")
	docs.WriteString("| `DB_PATH` | `./data/submissions.db` | Path to SQLite database file |\n\n")

	docs.WriteString("## Email Configuration\n\n")
	docs.WriteString("| Environment Variable | Default | Description |\n")
	docs.WriteString("|---------------------|---------|-------------|\n")
	docs.WriteString("| `EMAIL_PROVIDER` | `smtp` | Email service provider (`smtp` or `mailersend`) |\n")
	docs.WriteString("| `SMTP_HOST` | `localhost` | SMTP server hostname |\n")
	docs.WriteString("| `SMTP_PORT` | `1025` | SMTP server port |\n")
	docs.WriteString("| `SMTP_USERNAME` | - | SMTP username (optional) |\n")
	docs.WriteString("| `SMTP_PASSWORD` | - | SMTP password (optional) |\n")
	docs.WriteString("| `SMTP_FROM` | `noreply@example.com` | SMTP from address |\n")
	docs.WriteString("| `MAILERSEND_API_KEY` | - | MailerSend API key |\n")
	docs.WriteString("| `MAILERSEND_FROM_EMAIL` | `noreply@example.com` | MailerSend from email |\n")
	docs.WriteString("| `MAILERSEND_FROM_NAME` | `Form Submission Service` | MailerSend from name |\n\n")

	docs.WriteString("## Template Configuration\n\n")
	docs.WriteString("| Environment Variable | Default | Description |\n")
	docs.WriteString("|---------------------|---------|-------------|\n")
	docs.WriteString("| `TEMPLATES_DIR` | `./templates` | Templates directory |\n")
	docs.WriteString("| `DEFAULT_TEMPLATE` | `default.html` | Default email template |\n")
	docs.WriteString("| `EMAIL_SUBJECT` | `Form Submission Received` | Email subject line |\n\n")

	docs.WriteString("## Form Configuration\n\n")
	docs.WriteString("| Environment Variable | Default | Description |\n")
	docs.WriteString("|---------------------|---------|-------------|\n")
	docs.WriteString("| `EMAIL_FIELD` | `email` | Name of the email field in form data |\n")
	docs.WriteString("| `SUCCESS_MESSAGE` | `Thank you for your submission!` | Success response message |\n\n")

	docs.WriteString("## Security Configuration\n\n")
	docs.WriteString("| Environment Variable | Default | Description |\n")
	docs.WriteString("|---------------------|---------|-------------|\n")
	docs.WriteString("| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated list of allowed CORS origins |\n")
	docs.WriteString("| `CORS_ALLOWED_METHODS` | `GET,POST,OPTIONS` | Comma-separated list of allowed HTTP methods |\n")
	docs.WriteString("| `CORS_ALLOWED_HEADERS` | `Content-Type,Accept` | Comma-separated list of allowed headers |\n")
	docs.WriteString("| `RECAPTCHA_ENABLED` | `false` | Enable reCAPTCHA validation |\n")
	docs.WriteString("| `RECAPTCHA_SECRET_KEY` | - | reCAPTCHA secret key |\n")
	docs.WriteString("| `RECAPTCHA_SITE_KEY` | - | reCAPTCHA site key |\n")
	docs.WriteString("| `RECAPTCHA_ACTION_NAME` | `form_submit` | reCAPTCHA action name |\n")
	docs.WriteString("| `RECAPTCHA_MIN_SCORE` | `0.5` | Minimum reCAPTCHA score (0.0-1.0) |\n")
	docs.WriteString("| `RECAPTCHA_ENTERPRISE_ENABLED` | `false` | Enable reCAPTCHA Enterprise |\n")
	docs.WriteString("| `RECAPTCHA_PROJECT_ID` | - | Google Cloud project ID for reCAPTCHA Enterprise |\n")
	docs.WriteString("| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |\n")
	docs.WriteString("| `RATE_LIMIT_REQUESTS_PER_MINUTE` | `60` | Maximum requests per minute per IP |\n")
	docs.WriteString("| `RATE_LIMIT_BURST_SIZE` | `10` | Burst size for rate limiting |\n")
	docs.WriteString("| `ADMIN_API_KEY` | - | API key for admin endpoints (required for admin access) |\n")
	docs.WriteString("| `SECURITY_HEADERS_ENABLED` | `true` | Enable security headers |\n\n")
	docs.WriteString("| `TURNSTILE_ENABLED` | `false` | Enable Turnstile validation |\n")
	docs.WriteString("| `TURNSTILE_SECRET_KEY` | - | Cloudflare Turnstile secret key (server-side) |\n")
	docs.WriteString("| `TURNSTILE_SITE_KEY` | - | Cloudflare Turnstile site key (client-side) |\n")

	docs.WriteString("## Example Configuration\n\n")
	docs.WriteString("### Environment Variables\n")
	docs.WriteString("```bash\n")
	docs.WriteString("export PORT=8080\n")
	docs.WriteString("export DB_PATH=./data/submissions.db\n")
	docs.WriteString("export EMAIL_PROVIDER=smtp\n")
	docs.WriteString("export SMTP_HOST=smtp.example.com\n")
	docs.WriteString("export SMTP_PORT=587\n")
	docs.WriteString("export SMTP_FROM=noreply@example.com\n")
	docs.WriteString("export EMAIL_SUBJECT=\"Contact Form Submission\"\n")
	docs.WriteString("export SUCCESS_MESSAGE=\"Thank you for contacting us!\"\n")
	docs.WriteString("```\n\n")

	docs.WriteString("### config.json\n")
	docs.WriteString("```json\n")
	docs.WriteString("{\n")
	docs.WriteString("  \"server\": {\n")
	docs.WriteString("    \"port\": \"8080\",\n")
	docs.WriteString("    \"host\": \"0.0.0.0\"\n")
	docs.WriteString("  },\n")
	docs.WriteString("  \"database\": {\n")
	docs.WriteString("    \"path\": \"./data/submissions.db\"\n")
	docs.WriteString("  },\n")
	docs.WriteString("  \"email\": {\n")
	docs.WriteString("    \"provider\": \"smtp\",\n")
	docs.WriteString("    \"smtp\": {\n")
	docs.WriteString("      \"host\": \"smtp.example.com\",\n")
	docs.WriteString("      \"port\": \"587\",\n")
	docs.WriteString("      \"from\": \"noreply@example.com\"\n")
	docs.WriteString("    },\n")
	docs.WriteString("    \"templates\": {\n")
	docs.WriteString("      \"directory\": \"./templates\",\n")
	docs.WriteString("      \"default_template\": \"default.html\",\n")
	docs.WriteString("      \"subject\": \"Form Submission Received\"\n")
	docs.WriteString("    }\n")
	docs.WriteString("  },\n")
	docs.WriteString("  \"form\": {\n")
	docs.WriteString("    \"required_fields\": [\"email\"],\n")
	docs.WriteString("    \"email_field\": \"email\",\n")
	docs.WriteString("    \"success_message\": \"Thank you for your submission!\"\n")
	docs.WriteString("  },\n")
	docs.WriteString("  \"security\": {\n")
	docs.WriteString("    \"cors\": {\n")
	docs.WriteString("      \"allowed_origins\": [\"*\"],\n")
	docs.WriteString("      \"allowed_methods\": [\"GET\", \"POST\", \"OPTIONS\"],\n")
	docs.WriteString("      \"allowed_headers\": [\"Content-Type\", \"Accept\"]\n")
	docs.WriteString("    },\n")
	docs.WriteString("    \"recaptcha\": {\n")
	docs.WriteString("      \"enabled\": false,\n")
	docs.WriteString("      \"action_name\": \"form_submit\",\n")
	docs.WriteString("      \"min_score\": 0.5\n")
	docs.WriteString("    }\n")
	docs.WriteString("  }\n")
	docs.WriteString("}\n")
	docs.WriteString("```\n")

	return docs.String()
}

// SanitizationOptions defines options for input sanitization
type SanitizationOptions struct {
	DisableHTMLEscaping   bool `json:"disable_html_escaping" env:"DISABLE_HTML_ESCAPING"`
	DisableHTMLTagRemoval bool `json:"disable_html_tag_removal" env:"DISABLE_HTML_TAG_REMOVAL"`
}

// CrossFieldRule defines a validation rule that compares multiple fields
type CrossFieldRule struct {
	Fields    []string `json:"fields"`    // Fields to compare
	Type      string   `json:"type"`      // Type of comparison (equal, not_equal, etc.)
	Condition string   `json:"condition"` // Condition for applying the rule
	Message   string   `json:"message"`   // Error message
}
