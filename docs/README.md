# Go Submission Service Documentation

Welcome to the comprehensive documentation for the Go Submission Service - a production-ready, cloud-native form submission service built with Go.

## 📚 Documentation Index

### 🚀 Getting Started
- **[Getting Started Guide](getting-started.md)** - Complete setup and first steps
- **[Quick Start Examples](examples/README.md)** - Fast examples for common use cases
- **[Configuration Guide](configuration.md)** - Detailed configuration options

### 📖 API Reference
- **[API Reference](api-reference.md)** - Complete API documentation with examples
- **[OpenAPI Specification](openapi.yaml)** - Machine-readable API specification
- **[Admin API Reference](admin-api.yaml)** - Admin endpoints documentation

### 🔧 Usage & Examples
- **[Examples & Tutorials](examples/README.md)** - Real-world usage examples
- **[Email Templates](email-templates.md)** - Template system documentation
- **[CLI Usage](cli-usage.md)** - Command-line interface guide

### 🛠️ Deployment & Operations
- **[Docker Deployment](deployment/docker.md)** - Container deployment guide
- **[Kubernetes Deployment](deployment/kubernetes.md)** - K8s deployment examples
- **[Cloud Providers](deployment/cloud-providers.md)** - AWS, GCP, Azure guides
- **[Production Deployment](deployment/production.md)** - Production best practices

### 🔒 Security & Reliability
- **[Retry Mechanism](retry-mechanism.md)** - Email delivery reliability
- **[Troubleshooting Guide](troubleshooting.md)** - Common issues and solutions

## 🎯 Quick Navigation

### For New Users
1. Start with the **[Getting Started Guide](getting-started.md)**
2. Try the **[Quick Examples](examples/README.md)**
3. Review **[Configuration Options](configuration.md)**

### For Developers
1. Check the **[API Reference](api-reference.md)**
2. Explore **[Usage Examples](examples/README.md)**
3. Review **[OpenAPI Spec](openapi.yaml)**

### For DevOps
1. See **[Docker Deployment](deployment/docker.md)**
2. Check **[Kubernetes Setup](deployment/kubernetes.md)**
3. Review **[Production Guide](deployment/production.md)**

### For Troubleshooting
1. Start with **[Troubleshooting Guide](troubleshooting.md)**
2. Check **[Configuration Guide](configuration.md)**
3. Review **[Retry Mechanism](retry-mechanism.md)**

## 📋 Documentation Structure

```
docs/
├── README.md                    # This file - documentation index
├── getting-started.md           # Complete setup guide
├── api-reference.md             # API documentation
├── openapi.yaml                 # OpenAPI 3.0 specification
├── admin-api.yaml               # Admin API documentation
├── configuration.md             # Configuration options
├── email-templates.md           # Email template system
├── cli-usage.md                 # CLI tool documentation
├── retry-mechanism.md           # Email reliability
├── troubleshooting.md           # Common issues and solutions
├── examples/                    # Usage examples and tutorials
│   └── README.md               # Examples index
└── deployment/                  # Deployment guides
    ├── README.md               # Deployment overview
    ├── docker.md               # Docker deployment
    ├── kubernetes.md           # Kubernetes deployment
    ├── cloud-providers.md      # Cloud provider guides
    └── production.md           # Production best practices
```

## 🚀 Quick Start

### 1. Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/1it/go-submission-service.git
cd go-submission-service

# Start with Docker Compose
make run

# Test the service
curl http://localhost:8080/api/v1/health
```

### 2. Submit Your First Form

```bash
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "email": "test@example.com",
      "name": "Test User",
      "message": "Hello from the documentation!"
    }
  }'
```

### 3. Check Admin Dashboard

```bash
# Enable admin endpoints first
export ADMIN_ENDPOINTS_ENABLED=true
export ADMIN_API_KEY=your-api-key

# List submissions
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/admin/submissions
```

## 📖 Key Features

### ✅ Core Functionality
- **Form Submission Handling** - Accept and validate form data
- **Email Notifications** - Send emails using SMTP or MailerSend
- **Admin Dashboard** - Manage submissions with API key authentication
- **Database Storage** - SQLite with flexible JSON data storage

### 🔒 Security & Reliability
- **Rate Limiting** - Prevent abuse with configurable limits
- **CORS Support** - Secure cross-origin requests
- **reCAPTCHA Integration** - Spam protection
- **Security Headers** - Production-ready security
- **Retry Mechanism** - Reliable email delivery

### ⚙️ Configuration & Deployment
- **Environment Variables** - Flexible configuration
- **Docker Support** - Container-ready deployment
- **Kubernetes Ready** - Cloud-native deployment
- **Prometheus Metrics** - Monitoring and observability

### 🎨 Customization
- **Email Templates** - Customizable HTML templates
- **Validation Rules** - Configurable field validation
- **API Versioning** - Backward compatibility
- **Extensible Architecture** - Plugin-ready design

## 🔗 External Resources

- **[GitHub Repository](https://github.com/1it/go-submission-service)** - Source code and issues
- **[Docker Hub](https://hub.docker.com/r/1it/go-submission-service)** - Container images
- **[Go Report Card](https://goreportcard.com/report/github.com/1it/go-submission-service)** - Code quality
- **[License](https://opensource.org/licenses/MIT)** - MIT License

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md) for details.

### Documentation Contributions

To improve this documentation:

1. **Report Issues** - Create a GitHub issue for documentation problems
2. **Suggest Improvements** - Open a discussion for new documentation ideas
3. **Submit PRs** - Fork the repo and submit pull requests for improvements

### Documentation Standards

- Use clear, concise language
- Include practical examples
- Provide copy-paste ready code snippets
- Keep examples up-to-date with the latest version
- Use consistent formatting and structure

## 📝 Documentation Updates

This documentation is maintained alongside the codebase. When new features are added:

1. **Update relevant guides** - Modify existing documentation
2. **Add new examples** - Create practical usage examples
3. **Update API docs** - Keep OpenAPI spec current
4. **Review troubleshooting** - Add solutions for new issues

## 🆘 Getting Help

If you need help with the Go Submission Service:

1. **Check the Troubleshooting Guide** - [troubleshooting.md](troubleshooting.md)
2. **Review Configuration** - [configuration.md](configuration.md)
3. **Search Issues** - [GitHub Issues](https://github.com/1it/go-submission-service/issues)
4. **Create New Issue** - [GitHub Issues](https://github.com/1it/go-submission-service/issues/new)

## 📊 Documentation Status

| Document | Status | Last Updated |
|----------|--------|--------------|
| Getting Started | ✅ Complete | 2025-08-05 |
| API Reference | ✅ Complete | 2025-08-05 |
| OpenAPI Spec | ✅ Complete | 2025-08-05 |
| Configuration | ✅ Complete | 2025-08-05 |
| Examples | ✅ Complete | 2025-08-05 |
| Troubleshooting | ✅ Complete | 2025-08-05 |
| Deployment Guides | ✅ Complete | 2025-08-05 |
| Email Templates | ✅ Complete | 2025-08-05 |
| CLI Usage | ✅ Complete | 2025-08-05 |
| Retry Mechanism | ✅ Complete | 2025-08-05 |

---

**Happy coding! 🚀**

If you find this documentation helpful, please consider giving us a ⭐ on GitHub! 