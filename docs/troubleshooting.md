# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the Go Submission Service.

## Table of Contents

1. [Quick Diagnostics](#quick-diagnostics)
2. [Common Issues](#common-issues)
3. [Error Codes](#error-codes)
4. [Performance Issues](#performance-issues)
5. [Security Issues](#security-issues)
6. [Deployment Issues](#deployment-issues)
7. [FAQ](#faq)

## Quick Diagnostics

### 1. Service Health Check

```bash
# Check if the service is running
curl http://localhost:8080/api/v1/health

# Expected response:
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": 3600.5,
  "dependencies": {
    "database": "healthy",
    "email": "healthy"
  }
}
```

### 2. Configuration Validation

```bash
# Generate configuration documentation
./go-submission-service config-docs

# Check current configuration
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/admin/health
```

### 3. Database Status

```bash
# Check database file
ls -la ./data/submissions.db

# Check database tables
sqlite3 ./data/submissions.db ".tables"

# Check recent submissions
sqlite3 ./data/submissions.db "SELECT COUNT(*) FROM submissions;"
```

### 4. Log Analysis

```bash
# View recent logs
docker logs go-submission-service --tail 100

# Follow logs in real-time
docker logs go-submission-service -f

# Search for errors
docker logs go-submission-service 2>&1 | grep -i error
```

## Common Issues

### 1. Service Won't Start

**Symptoms**: Service fails to start or crashes immediately

**Diagnosis**:
```bash
# Check if port is already in use
netstat -tulpn | grep :8080

# Check file permissions
ls -la ./data/
ls -la ./templates/

# Check configuration syntax
./go-submission-service config-docs
```

**Solutions**:

#### Port Already in Use
```bash
# Kill process using port 8080
sudo lsof -ti:8080 | xargs kill -9

# Or change port in configuration
export PORT=8081
```

#### Permission Issues
```bash
# Fix data directory permissions
sudo chown -R $USER:$USER ./data/
chmod 755 ./data/

# Fix template directory permissions
sudo chown -R $USER:$USER ./templates/
chmod 755 ./templates/
```

#### Configuration Errors
```bash
# Validate configuration
./go-submission-service config-docs

# Check environment variables
env | grep -E "(EMAIL|SMTP|MAILERSEND|ADMIN)"
```

### 2. Form Submissions Failing

**Symptoms**: Form submissions return errors or don't work

**Diagnosis**:
```bash
# Test basic submission
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{"form_data": {"email": "test@example.com", "name": "Test"}}'

# Check validation rules
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{"form_data": {}}'
```

**Solutions**:

#### Validation Errors
```json
// Ensure required fields are present
{
  "form_data": {
    "email": "valid@email.com",
    "name": "John Doe"
  }
}

// Check validation rules in config
{
  "form": {
    "required_fields": ["email", "name"],
    "validation_rules": {
      "email": {"type": "email"},
      "name": {"type": "length", "min": 2, "max": 100}
    }
  }
}
```

#### Rate Limiting
```bash
# Check rate limit configuration
curl http://localhost:8080/api/v1/health

# Increase rate limits if needed
export RATE_LIMIT_REQUESTS_PER_MINUTE=120
export RATE_LIMIT_BURST_SIZE=20
```

### 3. Email Not Sending

**Symptoms**: Form submissions succeed but no emails are received

**Diagnosis**:
```bash
# Check email configuration
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/admin/health

# Test email service directly
./go-submission-service test-email
```

**Solutions**:

#### SMTP Issues
```bash
# Test SMTP connection
telnet smtp.gmail.com 587

# Check SMTP credentials
export SMTP_USERNAME=your-email@gmail.com
export SMTP_PASSWORD=your-app-password
export SMTP_FROM=your-email@gmail.com

# For Gmail, use App Password instead of regular password
# https://support.google.com/accounts/answer/185833
```

#### MailerSend Issues
```bash
# Verify API key
export MAILERSEND_API_KEY=your-api-key

# Test API key
curl -H "Authorization: Bearer your-api-key" \
  https://api.mailersend.com/v1/domains
```

#### Template Issues
```bash
# Check template syntax
cat ./templates/default.html

# Verify template variables
# Templates should use {{.variable_name}} syntax
```

### 4. Admin Endpoints Not Working

**Symptoms**: Admin API calls return 401 or 404

**Diagnosis**:
```bash
# Check if admin endpoints are enabled
curl http://localhost:8080/api/v1/admin/health

# Test with API key
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/admin/health
```

**Solutions**:

#### Enable Admin Endpoints
```bash
export ADMIN_ENDPOINTS_ENABLED=true
export ADMIN_API_KEY=your-secure-api-key
```

#### Fix API Key
```bash
# Generate secure API key
openssl rand -hex 32

# Set API key
export ADMIN_API_KEY=generated-api-key

# Test API key
curl -H "X-API-Key: generated-api-key" \
  http://localhost:8080/api/v1/admin/health
```

#### Check Path Prefix
```bash
# Default path is /api/v1/admin/
# If you changed it, use the correct path
export ADMIN_PATH_PREFIX=custom-admin

# Test with custom path
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/custom-admin/health
```

### 5. CORS Errors

**Symptoms**: Frontend requests fail with CORS errors

**Diagnosis**:
```bash
# Test CORS preflight
curl -H "Origin: https://yourdomain.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -X OPTIONS \
  http://localhost:8080/api/v1/submit
```

**Solutions**:

#### Configure CORS
```bash
# Add your domain to allowed origins
export CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com

# For development
export CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

#### Check Origin Header
```javascript
// Ensure your frontend sends the correct Origin header
fetch('http://localhost:8080/api/v1/submit', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Origin': 'https://yourdomain.com'
  },
  body: JSON.stringify(data)
});
```

### 6. Database Issues

**Symptoms**: Database errors or data corruption

**Diagnosis**:
```bash
# Check database file
ls -la ./data/submissions.db

# Check database integrity
sqlite3 ./data/submissions.db "PRAGMA integrity_check;"

# Check table structure
sqlite3 ./data/submissions.db ".schema submissions"
```

**Solutions**:

#### Database Corruption
```bash
# Backup current database
cp ./data/submissions.db ./data/submissions.db.backup

# Recreate database
rm ./data/submissions.db
./go-submission-service migrate
```

#### Permission Issues
```bash
# Fix database permissions
sudo chown $USER:$USER ./data/submissions.db
chmod 644 ./data/submissions.db

# Fix directory permissions
sudo chown -R $USER:$USER ./data/
chmod 755 ./data/
```

#### Disk Space
```bash
# Check disk space
df -h

# Clean up old logs
docker system prune -f
```

## Error Codes

### HTTP Status Codes

| Code | Meaning | Solution |
|------|---------|----------|
| 200 | Success | - |
| 400 | Bad Request | Check request format and validation rules |
| 401 | Unauthorized | Check API key for admin endpoints |
| 403 | Forbidden | Check IP whitelist and permissions |
| 404 | Not Found | Check endpoint URL and admin endpoint enablement |
| 429 | Too Many Requests | Increase rate limits or wait |
| 500 | Internal Server Error | Check server logs and configuration |

### Common Error Messages

#### "Invalid email format"
- Ensure email field contains a valid email address
- Check email validation rules in configuration

#### "Field 'name' is required"
- Include all required fields in form_data
- Check required_fields configuration

#### "Rate limit exceeded"
- Wait before making more requests
- Increase rate limit configuration
- Check if multiple clients share same IP

#### "Database connection failed"
- Check database file permissions
- Ensure sufficient disk space
- Verify database path configuration

#### "Email service unavailable"
- Check email provider configuration
- Verify SMTP credentials or API key
- Test email service connectivity

## Performance Issues

### 1. Slow Response Times

**Diagnosis**:
```bash
# Check response times
curl -w "@curl-format.txt" -o /dev/null -s \
  http://localhost:8080/api/v1/health

# Monitor metrics
curl http://localhost:8080/api/v1/metrics | grep http_request_duration
```

**Solutions**:

#### Database Optimization
```sql
-- Add indexes for better performance
CREATE INDEX idx_submissions_status ON submissions(status);
CREATE INDEX idx_submissions_created_at ON submissions(created_at);
CREATE INDEX idx_submissions_form_type ON submissions(form_type);
```

#### Rate Limiting
```bash
# Adjust rate limits for your traffic
export RATE_LIMIT_REQUESTS_PER_MINUTE=200
export RATE_LIMIT_BURST_SIZE=50
```

#### Connection Pooling
```bash
# Optimize database connections
export DB_MAX_OPEN_CONNS=25
export DB_MAX_IDLE_CONNS=5
export DB_CONN_MAX_LIFETIME=300s
```

### 2. High Memory Usage

**Diagnosis**:
```bash
# Check memory usage
docker stats go-submission-service

# Monitor Go runtime metrics
curl http://localhost:8080/api/v1/metrics | grep go_
```

**Solutions**:

#### Garbage Collection
```bash
# Set Go GC parameters
export GOGC=100
export GOMEMLIMIT=512MiB
```

#### Request Size Limits
```bash
# Limit request size
export MAX_REQUEST_SIZE=1048576  # 1MB
```

### 3. Database Performance

**Diagnosis**:
```sql
-- Check slow queries
SELECT * FROM submissions ORDER BY created_at DESC LIMIT 100;

-- Check table size
SELECT COUNT(*) FROM submissions;
```

**Solutions**:

#### Database Maintenance
```sql
-- Optimize database
VACUUM;
ANALYZE;

-- Clean old data
DELETE FROM submissions WHERE created_at < datetime('now', '-90 days');
```

## Security Issues

### 1. API Key Exposure

**Symptoms**: Unauthorized access to admin endpoints

**Solutions**:
```bash
# Rotate API key immediately
export ADMIN_API_KEY=$(openssl rand -hex 32)

# Check logs for unauthorized access
docker logs go-submission-service | grep "unauthorized"
```

### 2. CORS Misconfiguration

**Symptoms**: Security warnings or unauthorized cross-origin requests

**Solutions**:
```bash
# Restrict CORS to specific domains
export CORS_ALLOWED_ORIGINS=https://yourdomain.com

# Remove wildcard origins
# Don't use: export CORS_ALLOWED_ORIGINS=*
```

### 3. Rate Limiting Bypass

**Symptoms**: Excessive requests bypassing rate limits

**Solutions**:
```bash
# Enable rate limiting
export RATE_LIMIT_ENABLED=true

# Use stricter limits
export RATE_LIMIT_REQUESTS_PER_MINUTE=30
export RATE_LIMIT_BURST_SIZE=5
```

## Deployment Issues

### 1. Docker Issues

**Symptoms**: Container fails to start or crashes

**Solutions**:

#### Container Won't Start
```bash
# Check container logs
docker logs go-submission-service

# Check resource limits
docker stats go-submission-service

# Restart container
docker restart go-submission-service
```

#### Volume Mount Issues
```bash
# Check volume mounts
docker inspect go-submission-service | grep -A 10 "Mounts"

# Fix volume permissions
docker run --rm -v $(pwd)/data:/data alpine chown -R 1000:1000 /data
```

### 2. Kubernetes Issues

**Symptoms**: Pods not starting or services not accessible

**Solutions**:

#### Pod Status
```bash
# Check pod status
kubectl get pods -l app=go-submission-service

# Check pod logs
kubectl logs -l app=go-submission-service

# Check pod events
kubectl describe pod -l app=go-submission-service
```

#### Service Issues
```bash
# Check service
kubectl get svc go-submission-service

# Test service connectivity
kubectl port-forward svc/go-submission-service 8080:80
curl http://localhost:8080/api/v1/health
```

### 3. Environment Variables

**Symptoms**: Configuration not applied correctly

**Solutions**:
```bash
# Check environment variables in container
docker exec go-submission-service env | grep -E "(EMAIL|SMTP|ADMIN)"

# Verify config file
docker exec go-submission-service cat /app/config.json
```

## FAQ

### General Questions

**Q: How do I change the port the service runs on?**
A: Set the `PORT` environment variable or update the `server.port` in your config file.

**Q: Can I use a different database?**
A: Currently, the service supports SQLite. PostgreSQL support is planned for future releases.

**Q: How do I backup my data?**
A: Copy the SQLite database file: `cp ./data/submissions.db ./backup/submissions.db`

**Q: Can I run multiple instances?**
A: Yes, but you'll need to use a shared database or implement database clustering.

### Email Questions

**Q: Which email providers are supported?**
A: SMTP (Gmail, Outlook, etc.) and MailerSend are currently supported.

**Q: How do I set up Gmail SMTP?**
A: Use an App Password instead of your regular password. Enable 2FA and generate an App Password in your Google Account settings.

**Q: Can I use custom email templates?**
A: Yes, create HTML templates in the `templates` directory and reference them in your configuration.

**Q: How do I test email sending?**
A: Use MailHog for local testing or check the admin health endpoint for email service status.

### Security Questions

**Q: How do I generate a secure API key?**
A: Use `openssl rand -hex 32` to generate a secure random API key.

**Q: Can I restrict admin access by IP?**
A: Yes, configure `ADMIN_IP_WHITELIST` with allowed IP addresses.

**Q: How do I enable HTTPS?**
A: Use a reverse proxy like Nginx or Traefik to handle SSL termination.

**Q: Is reCAPTCHA required?**
A: No, reCAPTCHA is optional but recommended for production use.

### Performance Questions

**Q: How many requests can the service handle?**
A: Performance depends on your hardware and configuration. The service can handle hundreds of requests per second on modest hardware.

**Q: How do I monitor performance?**
A: Use the Prometheus metrics endpoint and health checks to monitor performance.

**Q: Can I scale horizontally?**
A: Yes, run multiple instances behind a load balancer, but ensure they share the same database.

### Troubleshooting Questions

**Q: How do I enable debug logging?**
A: Set the log level in your configuration or use the `-debug` flag when running the service.

**Q: Where are the logs stored?**
A: Logs are written to stdout/stderr. In Docker, use `docker logs` to view them.

**Q: How do I reset the database?**
A: Stop the service, delete the database file, and restart. The service will recreate the database automatically.

**Q: Can I export my data?**
A: Use the admin API endpoints to export data or directly query the SQLite database.

### Getting Help

If you're still experiencing issues:

1. **Check the logs** for detailed error messages
2. **Review the configuration** using `./go-submission-service config-docs`
3. **Test with minimal configuration** to isolate the issue
4. **Search existing issues** on GitHub
5. **Create a new issue** with detailed information including:
   - Error messages and logs
   - Configuration (with sensitive data removed)
   - Steps to reproduce
   - Environment details (OS, Docker version, etc.)

### Useful Commands

```bash
# Quick health check
curl http://localhost:8080/api/v1/health

# Test form submission
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{"form_data": {"email": "test@example.com", "name": "Test"}}'

# Check admin endpoints
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/admin/health

# View metrics
curl http://localhost:8080/api/v1/metrics

# Generate config docs
./go-submission-service config-docs

# Check database
sqlite3 ./data/submissions.db "SELECT COUNT(*) FROM submissions;"
``` 