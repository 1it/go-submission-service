# Usage Examples and Tutorials

This directory contains examples and tutorials for using the Go Submission Service in various scenarios.

## Quick Examples

### 1. Basic Contact Form

**Configuration**:
```json
{
  "form": {
    "required_fields": ["name", "email", "message"],
    "email_field": "email",
    "success_message": "Thank you for your message!"
  }
}
```

**Submission**:
```bash
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "name": "John Doe",
      "email": "john@example.com",
      "message": "Hello, this is a test message"
    }
  }'
```

### 2. React Integration

```jsx
import React, { useState } from 'react';

const ContactForm = () => {
  const [formData, setFormData] = useState({ name: '', email: '', message: '' });
  const [status, setStatus] = useState('idle');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setStatus('submitting');

    try {
      const response = await fetch('http://localhost:8080/api/v1/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ form_data: formData })
      });

      const result = await response.json();
      setStatus(result.success ? 'success' : 'error');
    } catch (err) {
      setStatus('error');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder="Name"
        value={formData.name}
        onChange={(e) => setFormData({...formData, name: e.target.value})}
        required
      />
      <input
        type="email"
        placeholder="Email"
        value={formData.email}
        onChange={(e) => setFormData({...formData, email: e.target.value})}
        required
      />
      <textarea
        placeholder="Message"
        value={formData.message}
        onChange={(e) => setFormData({...formData, message: e.target.value})}
        required
      />
      <button type="submit" disabled={status === 'submitting'}>
        {status === 'submitting' ? 'Sending...' : 'Send'}
      </button>
      {status === 'success' && <p>Thank you!</p>}
      {status === 'error' && <p>Error sending message</p>}
    </form>
  );
};
```

### 3. Business Contact Form

**Configuration**:
```json
{
  "form": {
    "required_fields": ["full_name", "work_email", "company", "message"],
    "email_field": "work_email",
    "validation_rules": {
      "work_email": {"type": "email"},
      "phone": {
        "type": "regex",
        "pattern": "^\\+?[1-9]\\d{1,14}$"
      }
    }
  }
}
```

**Submission**:
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
      "message": "Interested in your enterprise solutions"
    }
  }'
```

### 4. Newsletter Signup

**Configuration**:
```json
{
  "form": {
    "required_fields": ["email"],
    "email_field": "email",
    "success_message": "Thank you for subscribing!"
  }
}
```

**Submission**:
```bash
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "form_type": "newsletter",
      "email": "subscriber@example.com"
    }
  }'
```

### 5. Admin API Usage

**List Submissions**:
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/admin/submissions
```

**Get Statistics**:
```bash
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/admin/stats
```

**Update Submission Status**:
```bash
curl -X PATCH \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"status": "processed"}' \
  http://localhost:8080/api/v1/admin/submissions/550e8400-e29b-41d4-a716-446655440000/status
```

### 6. Email Templates

**Default Template** (`templates/default.html`):
```html
<!DOCTYPE html>
<html>
<head>
    <title>New Form Submission</title>
</head>
<body>
    <h2>New Form Submission</h2>
    <p><strong>Submitted:</strong> {{.timestamp}}</p>
    <p><strong>IP Address:</strong> {{.ip_address}}</p>
    
    <h3>Form Data:</h3>
    <ul>
        {{range $key, $value := .form_data}}
        <li><strong>{{$key}}:</strong> {{$value}}</li>
        {{end}}
    </ul>
</body>
</html>
```

**Business Contact Template** (`templates/business_contact.html`):
```html
<!DOCTYPE html>
<html>
<head>
    <title>New Business Contact</title>
</head>
<body>
    <h2>New Business Contact Submission</h2>
    <div>
        <p><strong>Name:</strong> {{.form_data.full_name}}</p>
        <p><strong>Email:</strong> {{.form_data.work_email}}</p>
        <p><strong>Company:</strong> {{.form_data.company}}</p>
        <p><strong>Job Title:</strong> {{.form_data.job_title}}</p>
    </div>
    <div>
        <h3>Message</h3>
        <p>{{.form_data.message}}</p>
    </div>
</body>
</html>
```

### 7. Testing Examples

**Health Check**:
```bash
curl http://localhost:8080/api/v1/health
```

**Load Testing**:
```bash
# Test with Apache Bench
ab -n 100 -c 10 -p test-data.json -T application/json \
  http://localhost:8080/api/v1/submit
```

**Validation Testing**:
```bash
# Test missing fields
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{"form_data": {}}'

# Test invalid email
curl -X POST http://localhost:8080/api/v1/submit \
  -H "Content-Type: application/json" \
  -d '{
    "form_data": {
      "name": "Test",
      "email": "invalid-email",
      "message": "Test"
    }
  }'
```

### 8. Configuration Examples

**Production Configuration**:
```json
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
      "api_key": "your-api-key",
      "from_email": "noreply@yourdomain.com",
      "from_name": "Your Company"
    }
  },
  "security": {
    "admin": {
      "enabled": true,
      "api_key": "your-secure-api-key"
    },
    "rate_limit": {
      "enabled": true,
      "requests_per_minute": 100
    },
    "cors": {
      "allowed_origins": ["https://yourdomain.com"]
    }
  }
}
```

**Development Configuration**:
```json
{
  "server": {
    "port": "8080"
  },
  "database": {
    "path": "./data/submissions.db"
  },
  "email": {
    "provider": "smtp",
    "smtp": {
      "host": "localhost",
      "port": "1025"
    }
  },
  "security": {
    "admin": {
      "enabled": true,
      "api_key": "dev-api-key"
    },
    "cors": {
      "allowed_origins": ["http://localhost:3000"]
    }
  }
}
```

## Next Steps

1. **Customize Email Templates**: Create branded templates for your use case
2. **Add Validation Rules**: Configure field-specific validation
3. **Set Up Monitoring**: Use Prometheus metrics and health checks
4. **Implement Frontend**: Build a web interface for form submission
5. **Security Hardening**: Configure rate limiting, CORS, and API keys

For more detailed examples, see the [Getting Started Guide](../getting-started.md) and [API Reference](../api-reference.md). 