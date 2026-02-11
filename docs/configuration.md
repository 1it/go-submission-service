# Configuration Guide

The Go Submission Service can be configured in multiple ways, with a flexible configuration system that supports different environments and deployment scenarios.

## Configuration Methods

Configuration is loaded in the following order, with later methods overriding earlier ones:

1. **Default values** - Secure, sensible defaults
2. **Configuration file** - JSON file with comprehensive settings
3. **Environment variables** - Override specific settings
4. **.env files** - Environment-specific configuration

## Environment-Based Configuration

The service supports different environments through:

1. **ENVIRONMENT variable** - Set to "development", "production", etc.
2. **Environment-specific .env files** - Like `.env.development` or `.env.production`
3. **Environment detection methods** - `IsProductionMode()` and `IsDevelopmentMode()`

## Configuration File

By default, the service looks for a `config.json` file in the current directory. You can specify a different path using the `CONFIG_FILE` environment variable.

Example `config.json`:

```json
{
  "server": {
    "port": "8080",
    "host": "0.0.0.0"
  },
  "database": {
    "path": "./data/submissions.db"
  },
  "email": {
    "provider": "smtp",
    "smtp": {
      "host": "smtp.example.com",
      "port": "587",
      "username": "your_username",
      "password": "your_password",
      "from": "noreply@example.com"
    },
    "templates": {
      "directory": "./templates",
      "default_template": "default.html",
      "subject": "Form Submission Received"
    }
  },
  "form": {
    "required_fields": ["email", "name"],
    "email_field": "email",
    "success_message": "Thank you for your submission!"
  },
  "security": {
    "cors": {
      "allowed_origins": ["https://example.com"],
      "allowed_methods": ["GET", "POST", "OPTIONS"],
      "allowed_headers": ["Content-Type", "Accept"]
    },
    "recaptcha": {
      "enabled": true,
      "secret_key": "your_secret_key",
      "site_key": "your_site_key",
      "action_name": "form_submit",
      "min_score": 0.5
    }
  }
}
```

## Environment Variables

All configuration options can be set using environment variables. For example:

```bash
# Server configuration
export PORT=8080
export HOST=0.0.0.0

# Database configuration
export DB_PATH=./data/submissions.db

# Email configuration
export EMAIL_PROVIDER=smtp
export SMTP_HOST=smtp.example.com
export SMTP_PORT=587
export SMTP_USERNAME=your_username
export SMTP_PASSWORD=your_password
export SMTP_FROM=noreply@example.com
```

## .env Files

The service supports loading environment variables from `.env` files:

- `.env` - Base environment variables
- `.env.development` - Development-specific variables
- `.env.production` - Production-specific variables
- `.env.test` - Testing-specific variables

Example `.env` file:

```
PORT=8080
DB_PATH=./data/submissions.db
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_FROM=noreply@example.com
```

## Configuration Reference

### Server Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `PORT` | `8080` | Server port |
| `HOST` | `0.0.0.0` | Server host |

### Database Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `DB_PATH` | `./data/submissions.db` | Path to SQLite database file |

### Email Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `EMAIL_PROVIDER` | `smtp` | Email service provider (`smtp` or `mailersend`) |
| `SMTP_HOST` | `localhost` | SMTP server hostname |
| `SMTP_PORT` | `1025` | SMTP server port |
| `SMTP_USERNAME` | - | SMTP username (optional) |
| `SMTP_PASSWORD` | - | SMTP password (optional) |
| `SMTP_FROM` | `noreply@example.com` | SMTP from address |
| `MAILERSEND_API_KEY` | - | MailerSend API key |
| `MAILERSEND_FROM_EMAIL` | `noreply@example.com` | MailerSend from email |
| `MAILERSEND_FROM_NAME` | `Form Submission Service` | MailerSend from name |

### Template Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `TEMPLATES_DIR` | `./templates` | Templates directory |
| `DEFAULT_TEMPLATE` | `default.html` | Default email template |
| `EMAIL_SUBJECT` | `Form Submission Received` | Email subject line |

### Form Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `EMAIL_FIELD` | `email` | Name of the email field in form data |
| `SUCCESS_MESSAGE` | `Thank you for your submission!` | Success response message |
| `REQUIRED_FIELDS` | `email` | Comma-separated list of required fields |

### Security Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated list of allowed CORS origins |
| `CORS_ALLOWED_METHODS` | `GET,POST,OPTIONS` | Comma-separated list of allowed HTTP methods |
| `CORS_ALLOWED_HEADERS` | `Content-Type,Accept` | Comma-separated list of allowed headers |
| `RECAPTCHA_ENABLED` | `false` | Enable reCAPTCHA validation |
| `RECAPTCHA_SECRET_KEY` | - | reCAPTCHA secret key |
| `RECAPTCHA_SITE_KEY` | - | reCAPTCHA site key |
| `RECAPTCHA_ACTION_NAME` | `form_submit` | reCAPTCHA action name |
| `RECAPTCHA_MIN_SCORE` | `0.5` | Minimum reCAPTCHA score (0.0-1.0) |
| `RECAPTCHA_ENTERPRISE_ENABLED` | `false` | Enable reCAPTCHA Enterprise |
| `RECAPTCHA_PROJECT_ID` | - | Google Cloud project ID for reCAPTCHA Enterprise |

## Secure Configuration

For secure configuration:

1. **Never commit sensitive information** (passwords, API keys) to version control
2. Use **environment variables** or **.env files** for sensitive information
3. Keep **.env files** out of version control (add to .gitignore)
4. Use **different configurations** for different environments
5. Validate configuration at startup to catch issues early