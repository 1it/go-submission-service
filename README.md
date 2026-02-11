# Go Submission Service

[![Go Report Card](https://goreportcard.com/badge/github.com/1it/go-submission-service)](https://goreportcard.com/report/github.com/1it/go-submission-service)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker Pulls](https://img.shields.io/docker/pulls/ghcr.io/1it/go-submission-service)](https://github.com/1it/go-submission-service/pkgs/container/go-submission-service)

A production-ready, cloud-native form submission service built with Go that handles form data collection, email notifications, and provides comprehensive admin management capabilities.

## ✨ Features

### 🚀 Core Functionality
- **Form Submission Handling**: Accept form data via REST API with validation
- **Email Notifications**: Send confirmation emails using multiple providers (MailerSend, SMTP)
- **Admin Dashboard API**: Complete submission management with authentication
- **Multiple Database Support**: SQLite (default) and PostgreSQL with auto-migration

### 🔒 Security & Reliability
- **Security Features**: reCAPTCHA integration, CORS, rate limiting, IP whitelisting
- **Production Ready**: Docker support, health checks, metrics, graceful shutdown
- **Retry Mechanism**: Robust email delivery with exponential backoff
- **Data Validation**: Comprehensive input validation and sanitization

### ⚙️ Configuration & Deployment
- **Flexible Configuration**: Environment variables, config files, CLI flags
- **Cloud-Native**: Ready for AWS, GCP, Azure deployment
- **Container Support**: Docker and Kubernetes ready
- **Monitoring**: Prometheus metrics, structured logging

### 🎨 Customization
- **Email Templates**: Customizable HTML email templates with variables
- **API Versioning**: Support for multiple API versions
- **Extensible**: Plugin-ready architecture for custom integrations

## Quick Start

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/1it/go-submission-service.git
cd go-submission-service

# Start the service with MailHog for local email testing
make docker-run

# The service will be available at http://localhost:8080
# MailHog web interface at http://localhost:8025
```

### Using Go directly

```bash
# Install dependencies
go mod download

# Run the service
go run main.go
```

## Configuration

The service can be configured via environment variables or a `config.json` file:

### Basic Configuration

```bash
# Server
export PORT=8080

# Database
export DB_PATH=./data/submissions.db

# Email (SMTP example)
export EMAIL_SERVICE=smtp
export SMTP_HOST=smtp.example.com
export SMTP_PORT=587
export SMTP_USERNAME=your-username
export SMTP_PASSWORD=your-password
export SMTP_FROM=noreply@example.com

# Retry Configuration (optional)
export RETRY_MAX_ATTEMPTS=3
export RETRY_INITIAL_DELAY_MS=1000
export RETRY_MAX_DELAY_MS=30000
export RETRY_BACKOFF_FACTOR=2.0
```

### Example config.json

```json
{
  "port": "8080",
  "db_path": "./data/submissions.db",
  "email_service": "smtp",
  "smtp_host": "smtp.example.com",
  "smtp_port": "587",
  "smtp_username": "your-username",
  "smtp_password": "your-password",
  "smtp_from": "noreply@example.com",
  "templates_dir": "./templates"
}
```

## API Usage

### Submit Form Data

```bash
POST /api/submit
Content-Type: application/json

{
  "form_data": {
    "email": "user@example.com",
    "name": "John Doe",
    "message": "Hello, this is a test message"
  },
  "recaptcha_token": "optional-recaptcha-token"
}
```

### V1 API Endpoints

The service provides versioned API endpoints under `/api/v1/` for enhanced functionality:

#### Form Submission (V1)
```bash
POST /api/v1/submit
Content-Type: application/json

{
  "form_data": {
    "email": "user@example.com",
    "name": "John Doe",
    "message": "Hello, this is a test message"
  },
  "recaptcha_token": "optional-recaptcha-token"
}
```

#### Health Check (V1)
```bash
GET /api/v1/health
```

#### OpenAPI Specification
```bash
GET /api/v1/openapi.json
```

#### Admin Endpoints (V1)

Admin endpoints use a configurable path prefix (default: 'mgmt') instead of 'admin' for security. These endpoints are disabled by default and include multiple security layers.

**Security Features**:
- **Disabled by default**: Must be explicitly enabled via configuration
- **Configurable path prefix**: Default 'mgmt' instead of obvious 'admin'
- **Multiple authentication methods**: API key via `X-API-Key` header or Bearer token via `Authorization: Bearer <token>` header
- **IP whitelisting**: Optional restriction to specific IP addresses or CIDR ranges
- **HTTPS enforcement**: Can require HTTPS in production environments
- **Comprehensive logging**: All access attempts are logged for security monitoring

**Configuration Environment Variables**:
- `ADMIN_ENDPOINTS_ENABLED=true` - Enable admin endpoints (default: false)
- `ADMIN_API_KEY=your-secret-key` - Set the API key for authentication
- `ADMIN_PATH_PREFIX=mgmt` - Customize the URL prefix (default: mgmt)
- `ADMIN_IP_WHITELIST=192.168.1.0/24,10.0.0.1` - Comma-separated list of allowed IPs/CIDRs
- `ADMIN_REQUIRE_HTTPS=true` - Enforce HTTPS in production (default: false)

```bash
# List all submissions
GET /api/v1/mgmt/submissions

# Get specific submission
GET /api/v1/mgmt/submissions/{id}

# Get admin statistics
GET /api/v1/mgmt/stats

# Enhanced health check
GET /api/v1/mgmt/health
```

### Legacy Endpoints

#### Health Check

```bash
GET /health
```

#### Metrics (Prometheus)

```bash
GET /metrics
```

## Email Templates

Templates are stored in the `templates/` directory and use Go's `html/template` syntax:

```html
<!DOCTYPE html>
<html>
<head>
    <title>Form Submission</title>
</head>
<body>
    <h1>New Form Submission</h1>
    <p>Name: {{.name}}</p>
    <p>Email: {{.email}}</p>
    <p>Message: {{.message}}</p>
</body>
</html>
```

## Development

### Prerequisites

- Go 1.23+
- Docker (optional)
- Make (optional)

### Available Commands

```bash
make help          # Show all available commands
make run           # Run locally (go run)
make docker-run    # Run with Docker Compose (service + MailHog)
make test          # Run unit tests
make test-e2e      # Run end-to-end tests
make docker-build  # Build Docker image
make clean         # Clean up resources
```

### Project Structure

```
├── cmd/                    # Command-line applications (e.g. CLI)
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── database/          # Database layer
│   ├── email/             # Email service implementations
│   ├── forms/             # Form validation
│   ├── handlers/          # HTTP handlers
│   ├── jobs/              # Background job processing
│   ├── metrics/           # Prometheus metrics
│   ├── models/            # Data models
│   └── retry/             # Retry mechanism
├── templates/             # Email templates
├── docs/                  # Documentation
└── examples/               # Example payloads and configs
```

## 🚀 Deployment

The service is designed for easy deployment across various environments and cloud platforms.

### Quick Deployment Options

| Platform | Guide | Best For |
|----------|-------|----------|
| 🐳 **Docker** | [Docker Guide](docs/deployment/docker.md) | Local development, simple deployments |
| ☸️ **Kubernetes** | [Kubernetes Guide](docs/deployment/kubernetes.md) | Scalable, container orchestration |
| ☁️ **Cloud Providers** | [Cloud Guide](docs/deployment/cloud-providers.md) | AWS, GCP, Azure deployments |
| 🏭 **Production** | [Production Guide](docs/deployment/production.md) | Production-ready configurations |

### Deployment Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │────│  Form Service   │────│    Database     │
│   (Nginx/ALB)   │    │   (Multiple)    │    │ (SQLite/Postgres)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                       ┌─────────────────┐
                       │  Email Service  │
                       │ (MailerSend/SMTP)│
                       └─────────────────┘
```

### Quick Start Commands

```bash
# Docker Compose (Recommended for testing)
docker-compose up -d

# Kubernetes — see docs/deployment/kubernetes.md for manifests

# Cloud Run (GCP)
gcloud run deploy --image ghcr.io/1it/go-submission-service:latest

# AWS ECS
aws ecs create-service --cli-input-json file://aws-service.json
```

### Production Checklist

- [ ] 🔒 **Security**: HTTPS, API keys, IP whitelisting
- [ ] 📊 **Monitoring**: Health checks, metrics, logging
- [ ] 🗄️ **Database**: PostgreSQL for production workloads
- [ ] 📧 **Email**: MailerSend or reliable SMTP provider
- [ ] 🔄 **Backup**: Automated database backups
- [ ] 🚀 **Scaling**: Load balancing and auto-scaling

> 📖 **Detailed Guides**: Visit our [deployment documentation](docs/deployment/) for comprehensive setup instructions.

## 📚 Documentation

Comprehensive documentation is available to help you get the most out of the Go Submission Service:

### 📖 Core Documentation

| Document | Description |
|----------|-------------|
| [Configuration Guide](docs/configuration.md) | Complete configuration reference |
| [API Reference](docs/api-reference.md) | Detailed API documentation with examples |
| [CLI Usage](docs/cli-usage.md) | Command-line interface guide |
| [Email Templates](docs/email-templates.md) | Template customization guide |
| [Retry Mechanism](docs/retry-mechanism.md) | Email delivery retry system |

### 🚀 Deployment Guides

| Platform | Guide | Description |
|----------|-------|-------------|
| [Docker](docs/deployment/docker.md) | Container deployment with Docker Compose |
| [Kubernetes](docs/deployment/kubernetes.md) | Scalable Kubernetes deployment |
| [Cloud Providers](docs/deployment/cloud-providers.md) | AWS, GCP, Azure deployment guides |
| [Production Setup](docs/deployment/production.md) | Production-ready configuration |

### 🔧 Development

- **Architecture**: Service follows clean architecture principles
- **Testing**: Comprehensive test suite with examples
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines
- **Code Style**: Go standard formatting with additional linting

## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. **🐛 Bug Reports**: [Create an issue](https://github.com/1it/go-submission-service/issues/new?template=bug_report.md)
2. **✨ Feature Requests**: [Suggest new features](https://github.com/1it/go-submission-service/issues/new?template=feature_request.md)
3. **📝 Documentation**: Improve our docs
4. **💻 Code**: Submit pull requests

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and development process.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/1it/go-submission-service.git
cd go-submission-service

# Install dependencies
go mod download

# Run tests
make test

# Run locally (Go)
make run

# Or with Docker Compose (includes MailHog)
make docker-run
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

Need help? We've got you covered:

### 📖 Self-Help Resources
1. **Documentation**: Check our [comprehensive docs](docs/)
2. **API Reference**: [Complete API guide](docs/api-reference.md)
3. **Deployment Guides**: [Platform-specific instructions](docs/deployment/)
4. **Configuration**: [Detailed configuration reference](docs/configuration.md)

### 🐛 Issues & Questions
1. **Search**: [Existing issues](https://github.com/1it/go-submission-service/issues)
2. **Report**: [Create new issue](https://github.com/1it/go-submission-service/issues/new)
3. **Discuss**: [GitHub Discussions](https://github.com/1it/go-submission-service/discussions)

---

<div align="center">

**⭐ Star this repo if you find it useful!**

[Report Bug](https://github.com/1it/go-submission-service/issues) • [Request Feature](https://github.com/1it/go-submission-service/issues) • [Documentation](docs/) • [Discussions](https://github.com/1it/go-submission-service/discussions)

</div>