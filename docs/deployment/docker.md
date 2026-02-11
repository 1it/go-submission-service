# Docker Deployment Guide

This guide covers deploying the Go Submission Service using Docker and Docker Compose.

## Quick Start with Docker Compose

The easiest way to get started is using the provided Docker Compose configuration:

```bash
# Clone the repository
git clone https://github.com/1it/go-submission-service.git
cd go-submission-service

# Start the service with MailHog for local email testing
make run

# Or manually with docker-compose
docker-compose up -d
```

This will start:
- **Form Submission Service** on `http://localhost:8080`
- **MailHog** (email testing) on `http://localhost:8025`

## Docker Image

### Using Pre-built Image

```bash
# Pull the latest image
docker pull ghcr.io/1it/go-submission-service:latest

# Run the container
docker run -d \
  --name form-submission-service \
  -p 8080:8080 \
  -e EMAIL_SERVICE=smtp \
  -e SMTP_HOST=your-smtp-host \
  -e SMTP_PORT=587 \
  -e SMTP_FROM=noreply@example.com \
  ghcr.io/1it/go-submission-service:latest
```

### Building Your Own Image

```bash
# Build the image
docker build -t go-submission-service .

# Run the container
docker run -d \
  --name form-submission-service \
  -p 8080:8080 \
  -e EMAIL_SERVICE=smtp \
  -e SMTP_HOST=your-smtp-host \
  go-submission-service
```

## Docker Compose Configuration

### Basic Configuration

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  form-submission-service:
    image: ghcr.io/1it/go-submission-service:latest
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DB_PATH=/data/submissions.db
      - EMAIL_SERVICE=smtp
      - SMTP_HOST=smtp.example.com
      - SMTP_PORT=587
      - SMTP_USERNAME=your-username
      - SMTP_PASSWORD=your-password
      - SMTP_FROM=noreply@example.com
    volumes:
      - ./data:/data
      - ./templates:/app/templates
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Production Configuration

For production deployments:

```yaml
version: '3.8'

services:
  form-submission-service:
    image: ghcr.io/1it/go-submission-service:latest
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
      - PORT=8080
      - DB_PATH=/data/submissions.db
      - EMAIL_SERVICE=mailersend
      - MAILERSEND_API_KEY=${MAILERSEND_API_KEY}
      - MAILERSEND_FROM_EMAIL=noreply@yourdomain.com
      - RECAPTCHA_ENABLED=true
      - RECAPTCHA_SECRET_KEY=${RECAPTCHA_SECRET_KEY}
      - CORS_ALLOWED_ORIGINS=https://yourdomain.com
      - ADMIN_ENDPOINTS_ENABLED=true
      - ADMIN_API_KEY=${ADMIN_API_KEY}
      - ADMIN_REQUIRE_HTTPS=true
    volumes:
      - form_data:/data
      - ./templates:/app/templates
      - ./config.json:/app/config.json:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

volumes:
  form_data:
```

### With Reverse Proxy (Nginx)

```yaml
version: '3.8'

services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - form-submission-service
    restart: unless-stopped

  form-submission-service:
    image: ghcr.io/1it/go-submission-service:latest
    expose:
      - "8080"
    environment:
      - ENVIRONMENT=production
      - PORT=8080
      # ... other environment variables
    volumes:
      - form_data:/data
      - ./templates:/app/templates
    restart: unless-stopped

volumes:
  form_data:
```

## Environment Variables

Key environment variables for Docker deployment:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Server port | `8080` | No |
| `DB_PATH` | Database file path | `./data/submissions.db` | No |
| `EMAIL_SERVICE` | Email provider (`smtp` or `mailersend`) | `smtp` | No |
| `SMTP_HOST` | SMTP server hostname | `localhost` | Yes (for SMTP) |
| `SMTP_PORT` | SMTP server port | `1025` | No |
| `SMTP_USERNAME` | SMTP username | - | No |
| `SMTP_PASSWORD` | SMTP password | - | No |
| `SMTP_FROM` | From email address | `noreply@example.com` | No |
| `MAILERSEND_API_KEY` | MailerSend API key | - | Yes (for MailerSend) |
| `RECAPTCHA_ENABLED` | Enable reCAPTCHA | `false` | No |
| `RECAPTCHA_SECRET_KEY` | reCAPTCHA secret key | - | Yes (if enabled) |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins | `*` | No |

## Volumes and Persistence

### Database Persistence

To persist form submissions:

```yaml
volumes:
  - ./data:/data  # Host directory
  # or
  - form_data:/data  # Named volume
```

### Template Customization

To use custom email templates:

```yaml
volumes:
  - ./custom-templates:/app/templates
```

### Configuration Files

To use a configuration file:

```yaml
volumes:
  - ./config.json:/app/config.json:ro
```

## Health Checks

The service provides health check endpoints:

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health check (admin endpoint)
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/mgmt/health
```

## Monitoring

### Prometheus Metrics

The service exposes Prometheus metrics at `/metrics`:

```yaml
services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
```

### Logging

Configure structured logging:

```yaml
services:
  form-submission-service:
    # ... other configuration
    environment:
      - LOG_LEVEL=info
      - LOG_FORMAT=json
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## Troubleshooting

### Common Issues

1. **Service won't start**
   ```bash
   # Check logs
   docker-compose logs form-submission-service
   
   # Check configuration
   docker-compose config
   ```

2. **Database permission issues**
   ```bash
   # Ensure data directory is writable
   mkdir -p ./data
   chmod 755 ./data
   ```

3. **Email not sending**
   ```bash
   # Test SMTP connection
   docker exec -it form-submission-service telnet smtp.example.com 587
   
   # Check email configuration
   docker exec -it form-submission-service env | grep SMTP
   ```

4. **Health check failing**
   ```bash
   # Test health endpoint
   docker exec -it form-submission-service curl -f http://localhost:8080/health
   ```

### Debug Mode

Run with debug logging:

```yaml
environment:
  - LOG_LEVEL=debug
  - ENVIRONMENT=development
```

## Security Considerations

1. **Use secrets management**:
   ```yaml
   secrets:
     smtp_password:
       external: true
   
   services:
     form-submission-service:
       secrets:
         - smtp_password
   ```

2. **Limit container privileges**:
   ```yaml
   services:
     form-submission-service:
       user: "1000:1000"
       read_only: true
       tmpfs:
         - /tmp
   ```

3. **Use specific image tags**:
   ```yaml
   image: ghcr.io/1it/go-submission-service:v1.0.0  # Instead of :latest
   ```

## Next Steps

- [Production Setup Guide](production.md)
- [SSL/TLS Configuration](ssl-tls.md)
- [Kubernetes Deployment](kubernetes.md)
- [Monitoring and Observability](../monitoring.md)