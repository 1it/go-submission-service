# Getting Started with Go Submission Service

This guide will walk you through setting up and using the Go Submission Service for handling form submissions, email notifications, and admin management.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Installation Options](#installation-options)
3. [Basic Configuration](#basic-configuration)
4. [Your First Form Submission](#your-first-form-submission)
5. [Email Templates](#email-templates)
6. [Admin Dashboard](#admin-dashboard)
7. [Security Features](#security-features)
8. [Monitoring and Metrics](#monitoring-and-metrics)
9. [Production Deployment](#production-deployment)
10. [Troubleshooting](#troubleshooting)

## Quick Start

### Using Docker Compose (Recommended)

The fastest way to get started is using Docker Compose:

```bash
# Clone the repository
git clone https://github.com/1it/go-submission-service.git
cd go-submission-service

# Start the service with MailHog for email testing
make run

# The service will be available at:
# - API: http://localhost:8080
# - MailHog (email testing): http://localhost:8025
```

### Test Your Setup

```bash
# Test the health endpoint
curl http://localhost:8080/api/v1/health

# Submit a test form
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "email": "test@example.com",
      "name": "Test User",
      "message": "Hello from the getting started guide!"
    }
  }'
```

## Installation Options

### Option 1: Docker (Recommended for Production)

```bash
# Pull the latest image
docker pull ghcr.io/1it/go-submission-service:latest

# Run with environment variables
docker run -d \
  --name go-submission-service \
  -p 8080:8080 \
  -e PORT=8080 \
  -e EMAIL_PROVIDER=smtp \
  -e SMTP_HOST=smtp.gmail.com \
  -e SMTP_PORT=587 \
  -e SMTP_USERNAME=your-email@gmail.com \
  -e SMTP_PASSWORD=your-app-password \
  -e SMTP_FROM=your-email@gmail.com \
  ghcr.io/1it/go-submission-service:latest
```

### Option 2: Go Binary

```bash
# Download the latest release
wget https://github.com/1it/go-submission-service/releases/latest/download/go-submission-service-linux-amd64

# Make it executable
chmod +x go-submission-service-linux-amd64

# Run the service
./go-submission-service-linux-amd64
```

### Option 3: From Source

```bash
# Clone and build
git clone https://github.com/1it/go-submission-service.git
cd go-submission-service
go build -o go-submission-service .

# Run the service
./go-submission-service
```

## Basic Configuration

The service can be configured via environment variables or a `config.json` file.

### Environment Variables (Recommended)

```bash
# Server Configuration
export PORT=8080
export HOST=0.0.0.0

# Database Configuration
export DB_PATH=./data/submissions.db

# Email Configuration (SMTP)
export EMAIL_PROVIDER=smtp
export SMTP_HOST=smtp.gmail.com
export SMTP_PORT=587
export SMTP_USERNAME=your-email@gmail.com
export SMTP_PASSWORD=your-app-password
export SMTP_FROM=your-email@gmail.com

# Email Configuration (MailerSend)
export EMAIL_PROVIDER=mailersend
export MAILERSEND_API_KEY=your-api-key
export MAILERSEND_FROM_EMAIL=noreply@yourdomain.com
export MAILERSEND_FROM_NAME=Your Company

# Form Configuration
export EMAIL_FIELD=email
export SUCCESS_MESSAGE="Thank you for your submission!"
export MAX_REQUEST_SIZE=1048576

# Security Configuration
export ADMIN_ENDPOINTS_ENABLED=true
export ADMIN_API_KEY=your-secure-api-key
export RATE_LIMIT_ENABLED=true
export RATE_LIMIT_REQUESTS_PER_MINUTE=60
export SECURITY_HEADERS_ENABLED=true

# reCAPTCHA Configuration (Optional)
export RECAPTCHA_ENABLED=true
export RECAPTCHA_SECRET_KEY=your-secret-key
export RECAPTCHA_SITE_KEY=your-site-key

# CORS Configuration
export CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### Configuration File

Create a `config.json` file:

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
      "host": "smtp.gmail.com",
      "port": "587",
      "username": "your-email@gmail.com",
      "password": "your-app-password",
      "from": "your-email@gmail.com"
    },
    "templates": {
      "directory": "./templates",
      "default_template": "default.html",
      "subject": "New Form Submission"
    }
  },
  "form": {
    "required_fields": ["email", "name"],
    "email_field": "email",
    "success_message": "Thank you for your submission!",
    "max_request_size": 1048576,
    "validation_rules": {
      "email": {
        "type": "email",
        "message": "Please provide a valid email address"
      },
      "name": {
        "type": "length",
        "min": 2,
        "max": 100,
        "message": "Name must be between 2 and 100 characters"
      }
    }
  },
  "security": {
    "admin": {
      "enabled": true,
      "api_key": "your-secure-api-key",
      "path_prefix": "admin"
    },
    "rate_limit": {
      "enabled": true,
      "requests_per_minute": 60,
      "burst_size": 10
    },
    "security_headers": {
      "enabled": true
    },
    "cors": {
      "allowed_origins": ["https://yourdomain.com"]
    }
  }
}
```

## Your First Form Submission

### 1. Basic Form Submission

```bash
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "email": "user@example.com",
      "name": "John Doe",
      "message": "Hello, this is a test message"
    }
  }'
```

### 2. Business Contact Form

```bash
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "form_type": "business_contact",
      "full_name": "John Smith",
      "work_email": "john.smith@company.com",
      "company": "Tech Solutions Inc",
      "company_size": "200-500",
      "job_title": "CTO",
      "phone": "+1-555-0123",
      "message": "We are interested in your enterprise solutions"
    }
  }'
```

### 3. With reCAPTCHA Protection

```bash
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "email": "user@example.com",
      "name": "John Doe",
      "message": "Hello, this is a test message"
    },
    "recaptcha_token": "03AFcWeA..."
  }'
```

### 4. JavaScript Integration

```javascript
// Basic form submission
async function submitForm(formData) {
  try {
    const response = await fetch('http://localhost:8080/api/v1/submit', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        form_data: formData
      })
    });
    
    const result = await response.json();
    
    if (result.success) {
      console.log('Form submitted successfully:', result.submission_id);
    } else {
      console.error('Form submission failed:', result.error);
    }
  } catch (error) {
    console.error('Network error:', error);
  }
}

// Usage
submitForm({
  email: 'user@example.com',
  name: 'John Doe',
  message: 'Hello from JavaScript!'
});
```

## Email Templates

### 1. Create Email Templates

Create a `templates` directory and add HTML templates:

```bash
mkdir -p templates
```

### 2. Default Template (`templates/default.html`)

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>New Form Submission</title>
</head>
<body>
    <h2>New Form Submission Received</h2>
    
    <p><strong>Submitted:</strong> {{.timestamp}}</p>
    <p><strong>IP Address:</strong> {{.ip_address}}</p>
    
    <h3>Form Data:</h3>
    <ul>
        {{range $key, $value := .form_data}}
        <li><strong>{{$key}}:</strong> {{$value}}</li>
        {{end}}
    </ul>
    
    <hr>
    <p><em>This email was sent automatically by Go Submission Service</em></p>
</body>
</html>
```

### 3. Business Contact Template (`templates/business_contact.html`)

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>New Business Contact Submission</title>
</head>
<body>
    <h2>New Business Contact Submission</h2>
    
    <div style="background-color: #f5f5f5; padding: 20px; border-radius: 5px;">
        <h3>Contact Information</h3>
        <p><strong>Name:</strong> {{.form_data.full_name}}</p>
        <p><strong>Email:</strong> {{.form_data.work_email}}</p>
        <p><strong>Company:</strong> {{.form_data.company}}</p>
        <p><strong>Job Title:</strong> {{.form_data.job_title}}</p>
        <p><strong>Company Size:</strong> {{.form_data.company_size}}</p>
        <p><strong>Phone:</strong> {{.form_data.phone}}</p>
    </div>
    
    <div style="margin-top: 20px;">
        <h3>Message</h3>
        <p>{{.form_data.message}}</p>
    </div>
    
    <hr>
    <p><em>Submitted at {{.timestamp}} from {{.ip_address}}</em></p>
</body>
</html>
```

### 4. Configure Template Usage

Update your configuration to use specific templates for different form types:

```json
{
  "form": {
    "validation_rules": {
      "business_contact": {
        "email_template": "business_contact.html"
      }
    }
  }
}
```

## Admin Dashboard

### 1. Enable Admin Endpoints

```bash
export ADMIN_ENDPOINTS_ENABLED=true
export ADMIN_API_KEY=your-secure-api-key
```

### 2. List Submissions

```bash
curl -H "X-API-Key: your-secure-api-key" \
  http://localhost:8080/api/v1/admin/submissions
```

### 3. Get Submission Statistics

```bash
curl -H "X-API-Key: your-secure-api-key" \
  http://localhost:8080/api/v1/admin/stats
```

### 4. Get Specific Submission

```bash
curl -H "X-API-Key: your-secure-api-key" \
  http://localhost:8080/api/v1/admin/submissions/550e8400-e29b-41d4-a716-446655440000
```

### 5. Update Submission Status

```bash
curl -X PATCH \
  -H "X-API-Key: your-secure-api-key" \
  -H "Content-Type: application/json" \
  -d '{"status": "processed"}' \
  http://localhost:8080/api/v1/admin/submissions/550e8400-e29b-41d4-a716-446655440000/status
```

### 6. Delete Submission

```bash
curl -X DELETE \
  -H "X-API-Key: your-secure-api-key" \
  http://localhost:8080/api/v1/admin/submissions/550e8400-e29b-41d4-a716-446655440000
```

## Security Features

### 1. Rate Limiting

Rate limiting is enabled by default. Configure limits in your config:

```json
{
  "security": {
    "rate_limit": {
      "enabled": true,
      "requests_per_minute": 60,
      "burst_size": 10
    }
  }
}
```

### 2. reCAPTCHA Integration

Set up reCAPTCHA v3:

```bash
export RECAPTCHA_ENABLED=true
export RECAPTCHA_SECRET_KEY=your-secret-key
export RECAPTCHA_SITE_KEY=your-site-key
```

Frontend integration:

```html
<script src="https://www.google.com/recaptcha/api.js?render=your-site-key"></script>
<script>
grecaptcha.ready(function() {
    grecaptcha.execute('your-site-key', {action: 'submit'}).then(function(token) {
        // Include token in your form submission
        submitForm(formData, token);
    });
});
</script>
```

### 3. CORS Configuration

Configure allowed origins:

```bash
export CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### 4. Security Headers

Security headers are enabled by default and include:
- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- X-XSS-Protection: 1; mode=block
- Content-Security-Policy
- Strict-Transport-Security

## Monitoring and Metrics

### 1. Health Check

```bash
curl http://localhost:8080/api/v1/health
```

### 2. Prometheus Metrics

```bash
curl http://localhost:8080/api/v1/metrics
```

### 3. Admin Health Check

```bash
curl -H "X-API-Key: your-secure-api-key" \
  http://localhost:8080/api/v1/admin/health
```

### 4. Key Metrics to Monitor

- `http_requests_total` - Total HTTP requests
- `signup_requests_total` - Total form submissions
- `signup_success_total` - Successful submissions
- `signup_failures_total` - Failed submissions by reason
- `recaptcha_verifications_total` - reCAPTCHA verification attempts

## Production Deployment

### 1. Docker Production Setup

```bash
# Create production config
cat > config-production.json << EOF
{
  "server": {
    "port": "8080"
  },
  "database": {
    "path": "/data/submissions.db"
  },
  "email": {
    "provider": "mailersend",
    "mailersend": {
      "api_key": "your-production-api-key",
      "from_email": "noreply@yourdomain.com",
      "from_name": "Your Company"
    }
  },
  "security": {
    "admin": {
      "enabled": true,
      "api_key": "your-secure-production-api-key"
    },
    "rate_limit": {
      "enabled": true,
      "requests_per_minute": 100
    }
  }
}
EOF

# Run with production config
docker run -d \
  --name go-submission-service \
  -p 8080:8080 \
  -v $(pwd)/config-production.json:/app/config.json \
  -v $(pwd)/data:/data \
  -v $(pwd)/templates:/app/templates \
  ghcr.io/1it/go-submission-service:latest
```

### 2. Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-submission-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: go-submission-service
  template:
    metadata:
      labels:
        app: go-submission-service
    spec:
      containers:
      - name: go-submission-service
        image: ghcr.io/1it/go-submission-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: EMAIL_PROVIDER
          value: "mailersend"
        - name: MAILERSEND_API_KEY
          valueFrom:
            secretKeyRef:
              name: mailersend-secret
              key: api-key
        volumeMounts:
        - name: config
          mountPath: /app/config.json
          subPath: config.json
      volumes:
      - name: config
        configMap:
          name: go-submission-service-config
---
apiVersion: v1
kind: Service
metadata:
  name: go-submission-service
spec:
  selector:
    app: go-submission-service
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 3. Environment Variables for Production

```bash
# Required for production
export EMAIL_PROVIDER=mailersend
export MAILERSEND_API_KEY=your-production-api-key
export ADMIN_API_KEY=your-secure-production-api-key
export CORS_ALLOWED_ORIGINS=https://yourdomain.com

# Recommended for production
export RATE_LIMIT_ENABLED=true
export SECURITY_HEADERS_ENABLED=true
export RECAPTCHA_ENABLED=true
export DB_PATH=/data/submissions.db
```

## Troubleshooting

### Common Issues

#### 1. Email Not Sending

**Symptoms**: Form submissions succeed but no emails are received

**Solutions**:
- Check email provider configuration
- Verify SMTP credentials or MailerSend API key
- Check email template syntax
- Review server logs for email errors

```bash
# Test email configuration
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/admin/health
```

#### 2. Database Errors

**Symptoms**: Service fails to start or form submissions fail

**Solutions**:
- Ensure database directory is writable
- Check disk space
- Verify database path configuration

```bash
# Check database status
ls -la ./data/
sqlite3 ./data/submissions.db ".tables"
```

#### 3. Rate Limiting Issues

**Symptoms**: Requests return 429 status codes

**Solutions**:
- Increase rate limit configuration
- Check if multiple clients are using same IP
- Review rate limit logs

```bash
# Check current rate limit config
curl http://localhost:8080/api/v1/health
```

#### 4. CORS Errors

**Symptoms**: Frontend requests fail with CORS errors

**Solutions**:
- Add your domain to CORS_ALLOWED_ORIGINS
- Check for typos in domain names
- Ensure HTTPS is used in production

```bash
# Test CORS configuration
curl -H "Origin: https://yourdomain.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -X OPTIONS \
  http://localhost:8080/api/v1/submit
```

#### 5. Admin Endpoints Not Working

**Symptoms**: Admin API calls return 401 or 404

**Solutions**:
- Ensure ADMIN_ENDPOINTS_ENABLED=true
- Verify API key is correct
- Check API key header format

```bash
# Test admin authentication
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/admin/health
```

### Getting Help

1. **Check Logs**: Review application logs for detailed error messages
2. **Health Checks**: Use health endpoints to verify service status
3. **Configuration**: Validate your configuration with the CLI tool
4. **GitHub Issues**: Report bugs and request features on GitHub
5. **Documentation**: Review the full API reference and configuration docs

```bash
# Generate configuration documentation
./go-submission-service config-docs

# Check service health
curl http://localhost:8080/api/v1/health

# View recent logs
docker logs go-submission-service --tail 100
```

## Next Steps

Now that you have the basic service running, consider:

1. **Customizing Email Templates**: Create branded templates for your use case
2. **Setting Up Monitoring**: Configure Prometheus and Grafana for metrics
3. **Implementing Frontend**: Build a web interface for form submission
4. **Adding Custom Validation**: Configure field-specific validation rules
5. **Setting Up CI/CD**: Automate testing and deployment
6. **Security Hardening**: Review and enhance security configurations

For more advanced usage, see the [API Reference](api-reference.md) and [Configuration Guide](configuration.md). 