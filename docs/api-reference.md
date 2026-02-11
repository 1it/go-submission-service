# API Reference

Comprehensive API documentation for the Go Submission Service with detailed usage examples.

## Base URLs

- **Development**: `http://localhost:8080`
- **Production**: `https://your-domain.com`

## Authentication

### Public Endpoints

Most endpoints are public and don't require authentication:
- Form submission endpoints
- Health checks
- Metrics (Prometheus)

### Admin Endpoints

Admin endpoints require API key authentication:

```bash
# Using X-API-Key header
curl -H "X-API-Key: your-api-key" https://your-domain.com/api/v1/mgmt/submissions

# Using Authorization Bearer token
curl -H "Authorization: Bearer your-api-key" https://your-domain.com/api/v1/mgmt/submissions
```

## Form Submission API

### Submit Form Data

Submit form data for processing and email delivery.

#### Legacy Endpoint

```http
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

#### V1 Endpoint (Recommended)

```http
POST /api/v1/submit
Content-Type: application/json

{
  "form_data": {
    "email": "user@example.com",
    "name": "John Doe",
    "message": "Hello, this is a test message",
    "company": "Example Corp",
    "phone": "+1-555-0123"
  },
  "recaptcha_token": "optional-recaptcha-token"
}
```

#### Request Body

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `form_data` | object | Yes | Form field data as key-value pairs |
| `recaptcha_token` | string | No | reCAPTCHA v3 token for spam protection |

#### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "message": "Thank you for your submission!",
  "submission_id": "uuid-string"
}
```

**Validation Error (400 Bad Request)**:
```json
{
  "success": false,
  "error": "Validation failed",
  "details": {
    "email": "Invalid email format",
    "name": "Name is required"
  }
}
```

**Rate Limited (429 Too Many Requests)**:
```json
{
  "success": false,
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

#### Examples

##### Contact Form

```bash
curl -X POST https://your-domain.com/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "email": "customer@example.com",
      "name": "Jane Smith",
      "subject": "Product Inquiry",
      "message": "I would like to know more about your services.",
      "phone": "+1-555-0123"
    }
  }'
```

##### Newsletter Signup

```bash
curl -X POST https://your-domain.com/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "email": "subscriber@example.com",
      "name": "John Doe",
      "interests": ["technology", "business"],
      "frequency": "weekly"
    }
  }'
```

##### Business Contact Form

```bash
curl -X POST https://your-domain.com/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "email": "business@company.com",
      "name": "Alice Johnson",
      "company": "Tech Solutions Inc",
      "title": "CTO",
      "message": "Interested in enterprise solutions",
      "budget": "$10,000-$50,000",
      "timeline": "Q2 2024"
    },
    "recaptcha_token": "03AGdBq25..."
  }'
```

##### JavaScript Frontend Integration

```javascript
// Vanilla JavaScript
async function submitForm(formData) {
  try {
    const response = await fetch('/api/v1/submit', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        form_data: formData,
        recaptcha_token: await getRecaptchaToken()
      })
    });
    
    const result = await response.json();
    
    if (result.success) {
      showSuccessMessage(result.message);
    } else {
      showErrorMessage(result.error, result.details);
    }
  } catch (error) {
    showErrorMessage('Network error occurred');
  }
}

// React example
import { useState } from 'react';

function ContactForm() {
  const [formData, setFormData] = useState({});
  const [loading, setLoading] = useState(false);
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      const response = await fetch('/api/v1/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ form_data: formData })
      });
      
      const result = await response.json();
      // Handle result...
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {/* Form fields */}
    </form>
  );
}
```

## Health Check API

### Basic Health Check

Check if the service is running.

```http
GET /health
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

### V1 Health Check

Enhanced health check with more details.

```http
GET /api/v1/health
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "checks": {
    "database": "healthy",
    "email_service": "healthy"
  },
  "uptime": "72h30m15s"
}
```

## Admin API

### List Submissions

Retrieve paginated list of form submissions.

```http
GET /api/v1/mgmt/submissions
X-API-Key: your-api-key
```

#### Query Parameters

| Parameter | Type | Description | Default |
|-----------|------|-------------|----------|
| `page` | integer | Page number (1-based) | 1 |
| `limit` | integer | Items per page (max 1000) | 50 |
| `status` | string | Filter by status | - |
| `form_type` | string | Filter by form type | - |
| `start_date` | string | Filter from date (YYYY-MM-DD) | - |
| `end_date` | string | Filter to date (YYYY-MM-DD) | - |

#### Examples

```bash
# Get first page of submissions
curl -H "X-API-Key: your-api-key" \
  "https://your-domain.com/api/v1/mgmt/submissions"

# Get submissions with pagination
curl -H "X-API-Key: your-api-key" \
  "https://your-domain.com/api/v1/mgmt/submissions?page=2&limit=25"

# Filter by date range
curl -H "X-API-Key: your-api-key" \
  "https://your-domain.com/api/v1/mgmt/submissions?start_date=2024-01-01&end_date=2024-01-31"

# Filter by status
curl -H "X-API-Key: your-api-key" \
  "https://your-domain.com/api/v1/mgmt/submissions?status=processed"
```

#### Response

```json
{
  "submissions": [
    {
      "id": "uuid-string",
      "form_data": {
        "email": "user@example.com",
        "name": "John Doe",
        "message": "Hello world"
      },
      "status": "processed",
      "created_at": "2024-01-15T10:30:00Z",
      "processed_at": "2024-01-15T10:30:05Z",
      "source_ip": "192.168.1.100",
      "user_agent": "Mozilla/5.0..."
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 150,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  }
}
```

### Get Submission by ID

Retrieve a specific submission.

```http
GET /api/v1/mgmt/submissions/{id}
X-API-Key: your-api-key
```

#### Example

```bash
curl -H "X-API-Key: your-api-key" \
  "https://your-domain.com/api/v1/mgmt/submissions/123e4567-e89b-12d3-a456-426614174000"
```

#### Response

```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "form_data": {
    "email": "user@example.com",
    "name": "John Doe",
    "message": "Hello world",
    "company": "Example Corp"
  },
  "status": "processed",
  "created_at": "2024-01-15T10:30:00Z",
  "processed_at": "2024-01-15T10:30:05Z",
  "source_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
  "email_sent": true,
  "email_sent_at": "2024-01-15T10:30:05Z",
  "retry_count": 0
}
```

### Get Statistics

Retrieve submission statistics.

```http
GET /api/v1/mgmt/stats
X-API-Key: your-api-key
```

#### Example

```bash
curl -H "X-API-Key: your-api-key" \
  "https://your-domain.com/api/v1/mgmt/stats"
```

#### Response

```json
{
  "total_submissions": 1250,
  "submissions_today": 45,
  "submissions_this_week": 320,
  "submissions_this_month": 1100,
  "status_breakdown": {
    "processed": 1200,
    "pending": 30,
    "failed": 15,
    "archived": 5
  },
  "top_sources": [
    {
      "source": "contact-form",
      "count": 800
    },
    {
      "source": "newsletter",
      "count": 300
    }
  ],
  "email_delivery_rate": 98.5,
  "average_processing_time_ms": 150
}
```

### Enhanced Health Check (Admin)

Detailed health check with dependency status.

```http
GET /api/v1/mgmt/health
X-API-Key: your-api-key
```

#### Response

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "uptime": "72h30m15s",
  "checks": {
    "database": {
      "status": "healthy",
      "response_time_ms": 5,
      "last_check": "2024-01-15T10:29:55Z"
    },
    "email_service": {
      "status": "healthy",
      "provider": "smtp",
      "last_check": "2024-01-15T10:29:50Z"
    },
    "recaptcha": {
      "status": "healthy",
      "enabled": true,
      "last_check": "2024-01-15T10:29:45Z"
    }
  },
  "metrics": {
    "total_requests": 15420,
    "successful_requests": 15200,
    "failed_requests": 220,
    "average_response_time_ms": 45
  }
}
```

## OpenAPI Specification

Get the OpenAPI 3.0 specification for the API.

```http
GET /api/v1/openapi.json
```

## Metrics API

Prometheus metrics endpoint for monitoring.

```http
GET /metrics
```

#### Example Metrics

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="POST",endpoint="/api/v1/submit",status="200"} 1250
http_requests_total{method="POST",endpoint="/api/v1/submit",status="400"} 45

# HELP form_submissions_total Total number of form submissions
# TYPE form_submissions_total counter
form_submissions_total{status="processed"} 1200
form_submissions_total{status="failed"} 15

# HELP email_delivery_duration_seconds Time taken to deliver emails
# TYPE email_delivery_duration_seconds histogram
email_delivery_duration_seconds_bucket{le="0.1"} 800
email_delivery_duration_seconds_bucket{le="0.5"} 1150
email_delivery_duration_seconds_bucket{le="1.0"} 1200
```

## Error Handling

### Error Response Format

All API errors follow a consistent format:

```json
{
  "success": false,
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Specific field error"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req-uuid"
}
```

### HTTP Status Codes

| Code | Description | When |
|------|-------------|------|
| 200 | OK | Successful request |
| 400 | Bad Request | Invalid request data |
| 401 | Unauthorized | Missing or invalid API key |
| 403 | Forbidden | Access denied |
| 404 | Not Found | Resource not found |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |
| 503 | Service Unavailable | Service temporarily unavailable |

### Common Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `RECAPTCHA_FAILED` | reCAPTCHA verification failed |
| `EMAIL_DELIVERY_FAILED` | Email could not be sent |
| `DATABASE_ERROR` | Database operation failed |
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Access denied |
| `NOT_FOUND` | Resource not found |

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Public endpoints**: 100 requests per minute per IP
- **Admin endpoints**: 1000 requests per minute per API key

### Rate Limit Headers

All responses include rate limit information:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248600
```

## CORS Support

The API supports Cross-Origin Resource Sharing (CORS) for web applications:

```
Access-Control-Allow-Origin: https://yourdomain.com
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Accept, X-API-Key
```

## SDKs and Libraries

### JavaScript/TypeScript

```javascript
// npm install @1it/form-submission-client
import { FormSubmissionClient } from '@1it/form-submission-client';

const client = new FormSubmissionClient({
  baseUrl: 'https://your-domain.com',
  apiKey: 'your-api-key' // For admin operations
});

// Submit form
const result = await client.submit({
  email: 'user@example.com',
  name: 'John Doe',
  message: 'Hello world'
});

// Admin operations
const submissions = await client.admin.getSubmissions({
  page: 1,
  limit: 50
});
```

### Go

```go
// go get github.com/1it/go-submission-service/pkg/client
import "github.com/1it/go-submission-service/pkg/client"

client := client.New(client.Config{
    BaseURL: "https://your-domain.com",
    APIKey:  "your-api-key",
})

// Submit form
result, err := client.Submit(context.Background(), map[string]interface{}{
    "email":   "user@example.com",
    "name":    "John Doe",
    "message": "Hello world",
})
```

### Python

```python
# pip install form-submission-client
from form_submission_client import FormSubmissionClient

client = FormSubmissionClient(
    base_url="https://your-domain.com",
    api_key="your-api-key"
)

# Submit form
result = client.submit({
    "email": "user@example.com",
    "name": "John Doe",
    "message": "Hello world"
})

# Admin operations
submissions = client.admin.get_submissions(page=1, limit=50)
```

## Webhooks (Future)

Webhook support for real-time notifications (planned feature):

```json
{
  "event": "submission.created",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "submission_id": "uuid-string",
    "form_data": {
      "email": "user@example.com",
      "name": "John Doe"
    }
  }
}
```

## Support

For API support:

- 📖 [Documentation](https://github.com/1it/go-submission-service/docs)
- 🐛 [Report Issues](https://github.com/1it/go-submission-service/issues)
- 💬 [Discussions](https://github.com/1it/go-submission-service/discussions)
- 📧 Email: support@yourdomain.com