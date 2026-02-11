# Production Setup Guide

This guide covers production-ready configuration and best practices for deploying the Go Submission Service.

## Production Checklist

### Security
- [ ] Enable HTTPS/TLS encryption
- [ ] Configure proper CORS origins
- [ ] Enable reCAPTCHA protection
- [ ] Set up API key authentication for admin endpoints
- [ ] Configure IP whitelisting for admin access
- [ ] Use environment variables for secrets
- [ ] Enable security headers
- [ ] Set up rate limiting

### Performance
- [ ] Configure database connection pooling
- [ ] Set up horizontal scaling
- [ ] Configure caching (if applicable)
- [ ] Optimize resource limits
- [ ] Set up CDN for static assets

### Reliability
- [ ] Configure health checks
- [ ] Set up monitoring and alerting
- [ ] Configure log aggregation
- [ ] Set up backup and recovery
- [ ] Configure graceful shutdown
- [ ] Set up retry mechanisms

### Observability
- [ ] Configure structured logging
- [ ] Set up metrics collection
- [ ] Configure distributed tracing
- [ ] Set up error tracking
- [ ] Configure performance monitoring

## Environment Configuration

### Production Environment Variables

```bash
# Environment
ENVIRONMENT=production

# Server Configuration
PORT=8080
HOST=0.0.0.0

# Database
DB_PATH=/data/submissions.db
# Or for PostgreSQL:
# DATABASE_URL=postgres://user:password@host:5432/dbname

# Email Service (MailerSend recommended for production)
EMAIL_SERVICE=mailersend
MAILERSEND_API_KEY=your_mailersend_api_key
MAILERSEND_FROM_EMAIL=noreply@yourdomain.com
MAILERSEND_FROM_NAME="Your Service Name"

# Security
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
RECAPTCHA_ENABLED=true
RECAPTCHA_SECRET_KEY=your_recaptcha_secret_key
RECAPTCHA_SITE_KEY=your_recaptcha_site_key
RECAPTCHA_MIN_SCORE=0.5

# Admin Security
ADMIN_ENDPOINTS_ENABLED=true
ADMIN_API_KEY=your_secure_admin_api_key
ADMIN_PATH_PREFIX=mgmt
ADMIN_IP_WHITELIST=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
ADMIN_REQUIRE_HTTPS=true

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=100
RATE_LIMIT_BURST=20

# Logging
LOG_LEVEL=info
LOG_FORMAT=json

# Retry Configuration
RETRY_MAX_ATTEMPTS=3
RETRY_INITIAL_DELAY_MS=1000
RETRY_MAX_DELAY_MS=30000
RETRY_BACKOFF_FACTOR=2.0

# Templates
TEMPLATES_DIR=/app/templates
DEFAULT_TEMPLATE=default.html
EMAIL_SUBJECT="Form Submission Received"

# Form Configuration
REQUIRED_FIELDS=email,name
EMAIL_FIELD=email
SUCCESS_MESSAGE="Thank you for your submission!"
```

### Configuration File (config.json)

```json
{
  "environment": "production",
  "server": {
    "port": "8080",
    "host": "0.0.0.0",
    "read_timeout": "30s",
    "write_timeout": "30s",
    "idle_timeout": "120s"
  },
  "database": {
    "path": "/data/submissions.db",
    "max_connections": 25,
    "connection_timeout": "30s"
  },
  "email": {
    "provider": "mailersend",
    "mailersend": {
      "api_key": "${MAILERSEND_API_KEY}",
      "from_email": "noreply@yourdomain.com",
      "from_name": "Your Service Name"
    },
    "templates": {
      "directory": "/app/templates",
      "default_template": "default.html",
      "subject": "Form Submission Received"
    }
  },
  "security": {
    "cors": {
      "allowed_origins": ["https://yourdomain.com", "https://www.yourdomain.com"],
      "allowed_methods": ["GET", "POST", "OPTIONS"],
      "allowed_headers": ["Content-Type", "Accept"],
      "max_age": 3600
    },
    "recaptcha": {
      "enabled": true,
      "secret_key": "${RECAPTCHA_SECRET_KEY}",
      "site_key": "${RECAPTCHA_SITE_KEY}",
      "min_score": 0.5,
      "action_name": "form_submit"
    },
    "admin": {
      "enabled": true,
      "api_key": "${ADMIN_API_KEY}",
      "path_prefix": "mgmt",
      "ip_whitelist": ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"],
      "require_https": true
    },
    "rate_limiting": {
      "enabled": true,
      "requests_per_minute": 100,
      "burst": 20
    }
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout"
  },
  "retry": {
    "max_attempts": 3,
    "initial_delay_ms": 1000,
    "max_delay_ms": 30000,
    "backoff_factor": 2.0
  },
  "form": {
    "required_fields": ["email", "name"],
    "email_field": "email",
    "success_message": "Thank you for your submission!",
    "max_file_size": "10MB"
  }
}
```

## Database Configuration

### SQLite (Small to Medium Scale)

For smaller deployments, SQLite is sufficient:

```bash
# Ensure proper permissions
mkdir -p /data
chown app:app /data
chmod 755 /data

# Database configuration
DB_PATH=/data/submissions.db
```

### PostgreSQL (Large Scale)

For high-volume production deployments:

```bash
# Database connection
DATABASE_URL=postgres://username:password@host:5432/submissions?sslmode=require

# Connection pool settings
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_TIMEOUT=30s
```

#### PostgreSQL Setup

```sql
-- Create database and user
CREATE DATABASE submissions;
CREATE USER formservice WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE submissions TO formservice;

-- Create tables (auto-migrated by service)
-- The service will handle schema creation and migrations
```

## Email Service Configuration

### MailerSend (Recommended)

MailerSend provides better deliverability and analytics:

```bash
EMAIL_SERVICE=mailersend
MAILERSEND_API_KEY=your_api_key
MAILERSEND_FROM_EMAIL=noreply@yourdomain.com
MAILERSEND_FROM_NAME="Your Service Name"
```

### SMTP (Alternative)

For custom SMTP providers:

```bash
EMAIL_SERVICE=smtp
SMTP_HOST=smtp.yourdomain.com
SMTP_PORT=587
SMTP_USERNAME=your_username
SMTP_PASSWORD=your_password
SMTP_FROM=noreply@yourdomain.com
SMTP_TLS=true
```

## Security Configuration

### SSL/TLS Setup

#### Using Reverse Proxy (Recommended)

```nginx
# /etc/nginx/sites-available/form-submission
server {
    listen 443 ssl http2;
    server_name forms.yourdomain.com;
    
    ssl_certificate /etc/letsencrypt/live/forms.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/forms.yourdomain.com/privkey.pem;
    
    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";
    
    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req zone=api burst=20 nodelay;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # Admin endpoints with additional restrictions
    location /api/v1/mgmt/ {
        # Restrict to specific IPs
        allow 10.0.0.0/8;
        allow 172.16.0.0/12;
        allow 192.168.0.0/16;
        deny all;
        
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name forms.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```

### reCAPTCHA Configuration

1. **Get reCAPTCHA keys** from [Google reCAPTCHA](https://www.google.com/recaptcha/)
2. **Configure the service**:
   ```bash
   RECAPTCHA_ENABLED=true
   RECAPTCHA_SECRET_KEY=your_secret_key
   RECAPTCHA_SITE_KEY=your_site_key
   RECAPTCHA_MIN_SCORE=0.5
   ```

3. **Frontend integration**:
   ```html
   <script src="https://www.google.com/recaptcha/api.js?render=YOUR_SITE_KEY"></script>
   <script>
   grecaptcha.ready(function() {
       grecaptcha.execute('YOUR_SITE_KEY', {action: 'form_submit'})
       .then(function(token) {
           // Include token in form submission
       });
   });
   </script>
   ```

### API Key Security

```bash
# Generate secure API key
openssl rand -hex 32

# Set in environment
ADMIN_API_KEY=your_generated_key

# Use in requests
curl -H "X-API-Key: your_generated_key" https://forms.yourdomain.com/api/v1/mgmt/submissions
```

## Monitoring and Observability

### Prometheus Metrics

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'form-submission-service'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

### Grafana Dashboard

Key metrics to monitor:

- **Request rate**: `rate(http_requests_total[5m])`
- **Error rate**: `rate(http_requests_total{status=~"4..|5.."}[5m])`
- **Response time**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- **Form submissions**: `rate(form_submissions_total[5m])`
- **Email delivery**: `rate(email_delivery_total[5m])`

### Log Aggregation

#### Using ELK Stack

```yaml
# docker-compose.yml
version: '3.8'
services:
  form-submission-service:
    # ... service configuration
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
        labels: "service=form-submission"
  
  filebeat:
    image: docker.elastic.co/beats/filebeat:8.5.0
    volumes:
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
```

#### Using Fluentd

```yaml
logging:
  driver: fluentd
  options:
    fluentd-address: localhost:24224
    tag: form-submission-service
```

### Health Checks

```bash
# Basic health check
curl -f http://localhost:8080/health || exit 1

# Detailed health check
curl -f -H "X-API-Key: $ADMIN_API_KEY" http://localhost:8080/api/v1/mgmt/health || exit 1
```

### Alerting Rules

```yaml
# alerting-rules.yml
groups:
  - name: form-submission-service
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          
      - alert: ServiceDown
        expr: up{job="form-submission-service"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Form submission service is down"
          
      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time detected"
```

## Backup and Recovery

### Database Backup

#### SQLite Backup

```bash
#!/bin/bash
# backup-sqlite.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
DB_PATH="/data/submissions.db"

# Create backup
sqlite3 $DB_PATH ".backup $BACKUP_DIR/submissions_$DATE.db"

# Compress backup
gzip "$BACKUP_DIR/submissions_$DATE.db"

# Upload to cloud storage (example with AWS S3)
aws s3 cp "$BACKUP_DIR/submissions_$DATE.db.gz" s3://your-backup-bucket/

# Clean up old local backups (keep last 7 days)
find $BACKUP_DIR -name "submissions_*.db.gz" -mtime +7 -delete
```

#### PostgreSQL Backup

```bash
#!/bin/bash
# backup-postgres.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
DB_URL="$DATABASE_URL"

# Create backup
pg_dump $DB_URL | gzip > "$BACKUP_DIR/submissions_$DATE.sql.gz"

# Upload to cloud storage
aws s3 cp "$BACKUP_DIR/submissions_$DATE.sql.gz" s3://your-backup-bucket/

# Clean up old backups
find $BACKUP_DIR -name "submissions_*.sql.gz" -mtime +7 -delete
```

### Automated Backup with Cron

```bash
# Add to crontab
0 2 * * * /path/to/backup-script.sh
```

## Performance Optimization

### Resource Limits

```yaml
# Docker Compose
services:
  form-submission-service:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### Database Optimization

#### SQLite Optimization

```sql
-- Enable WAL mode for better concurrency
PRAGMA journal_mode=WAL;

-- Optimize for performance
PRAGMA synchronous=NORMAL;
PRAGMA cache_size=10000;
PRAGMA temp_store=memory;
```

#### PostgreSQL Optimization

```sql
-- Create indexes for common queries
CREATE INDEX idx_submissions_created_at ON submissions(created_at);
CREATE INDEX idx_submissions_status ON submissions(status);
CREATE INDEX idx_submissions_email ON submissions((form_data->>'email'));
```

### Caching

```bash
# Redis for caching (if needed)
REDIS_URL=redis://localhost:6379/0
CACHE_TTL=300
```

## Scaling

### Horizontal Scaling

```yaml
# docker-compose.yml
version: '3.8'
services:
  form-submission-service:
    image: ghcr.io/1it/go-submission-service:latest
    deploy:
      replicas: 3
    # ... configuration
  
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - form-submission-service
```

### Load Balancer Configuration

```nginx
# nginx.conf
upstream form_submission_backend {
    least_conn;
    server form-submission-service_1:8080;
    server form-submission-service_2:8080;
    server form-submission-service_3:8080;
}

server {
    listen 80;
    
    location / {
        proxy_pass http://form_submission_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## Troubleshooting

### Common Production Issues

1. **High memory usage**:
   ```bash
   # Check memory usage
   docker stats form-submission-service
   
   # Adjust memory limits
   deploy:
     resources:
       limits:
         memory: 1G
   ```

2. **Database connection issues**:
   ```bash
   # Check database connectivity
   docker exec -it form-submission-service sqlite3 /data/submissions.db ".tables"
   
   # Check file permissions
   ls -la /data/
   ```

3. **Email delivery failures**:
   ```bash
   # Check email service configuration
   docker exec -it form-submission-service env | grep EMAIL
   
   # Test SMTP connectivity
   telnet smtp.yourdomain.com 587
   ```

4. **High response times**:
   ```bash
   # Check resource usage
   docker stats
   
   # Check database performance
   sqlite3 /data/submissions.db "EXPLAIN QUERY PLAN SELECT * FROM submissions ORDER BY created_at DESC LIMIT 10;"
   ```

### Debug Mode

```bash
# Enable debug logging
LOG_LEVEL=debug
ENVIRONMENT=development

# Check logs
docker logs -f form-submission-service
```

## Security Hardening

### Container Security

```dockerfile
# Use non-root user
USER 1000:1000

# Read-only filesystem
--read-only
--tmpfs /tmp

# Drop capabilities
--cap-drop=ALL
--cap-add=NET_BIND_SERVICE
```

### Network Security

```bash
# Firewall rules (iptables)
iptables -A INPUT -p tcp --dport 8080 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j DROP
```

### Regular Security Updates

```bash
# Update base image regularly
docker pull ghcr.io/1it/go-submission-service:latest

# Scan for vulnerabilities
docker scan ghcr.io/1it/go-submission-service:latest
```

## Maintenance

### Regular Tasks

1. **Update dependencies** monthly
2. **Review logs** weekly
3. **Check metrics** daily
4. **Test backups** monthly
5. **Security updates** as needed

### Maintenance Scripts

```bash
#!/bin/bash
# maintenance.sh

# Clean up old logs
find /var/log -name "*.log" -mtime +30 -delete

# Vacuum SQLite database
sqlite3 /data/submissions.db "VACUUM;"

# Check disk space
df -h

# Check service health
curl -f http://localhost:8080/health
```

## Next Steps

- [SSL/TLS Configuration](ssl-tls.md)
- [Cloud Provider Deployment](cloud-providers.md)
- [Monitoring Setup](../monitoring.md)
- [Backup Strategies](../backup.md)