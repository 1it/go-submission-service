# Form Examples

This directory contains example form submissions for different form types supported by the Go Submission Service.

## Business Contact Form

The business contact form is designed for B2B lead generation and business inquiries.

### Form Type
`business_contact`

### Required Fields
- `full_name`: Full name of the contact person
- `work_email`: Business email address (validated)
- `company`: Company name
- `company_size`: Company size category (must be one of: `50-200`, `200-500`, `500-1000`, `1000-5000`)

### Optional Fields
- `job_title`: Job title/position
- `phone`: Phone number (validated format)
- `message`: Additional message or inquiry details (max 1000 characters)

### Email Templates
- **Admin Notification**: `business_contact.html` - Rich HTML template for admin notifications
- **User Confirmation**: `business_contact_confirmation.html` - Confirmation email sent to the submitter

### Example Usage

#### cURL Command
```bash
curl -X POST http://localhost:8080/api/submit \
  -H "Content-Type: application/json" \
  -d @examples/business-contact-form.json
```

#### JavaScript Fetch
```javascript
fetch('http://localhost:8080/api/submit', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    form_data: {
      form_type: 'business_contact',
      full_name: 'John Smith',
      work_email: 'john.smith@company.com',
      company: 'Tech Solutions Inc',
      company_size: '200-500',
      job_title: 'CTO',
      phone: '+1-555-0123',
      message: 'Interested in your enterprise solutions.'
    }
  })
})
.then(response => response.json())
.then(data => console.log(data));
```

### Validation Rules

1. **Full Name**: 2-100 characters, required
2. **Work Email**: Valid email format, required
3. **Company**: 2-200 characters, required
4. **Company Size**: Must be exactly one of the predefined options
5. **Phone**: Optional, but if provided must match international phone format
6. **Job Title**: Optional, max 100 characters
7. **Message**: Optional, max 1000 characters

### Response Format

#### Success Response
```json
{
  "success": true,
  "message": "Form submitted successfully",
  "submission_id": "uuid-here"
}
```

#### Validation Error Response
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "work_email": "Invalid email address",
    "company_size": "Invalid company size. Must be one of: 50-200, 200-500, 500-1000, 1000-5000"
  }
}
```

### Testing

Run the business contact form tests:
```bash
./test-business-contact.sh
```

Or use the enhanced e2e tests:
```bash
make test-e2e
```

## Company Size Options

The business contact form supports the following company size categories:

- **50-200**: Small to medium businesses
- **200-500**: Medium businesses  
- **500-1000**: Large businesses
- **1000-5000**: Enterprise businesses

These categories help with lead qualification and routing to appropriate sales teams.

## Template Customization

The email templates can be customized by editing:
- `templates/business_contact.html` - Admin notification template
- `templates/business_contact_confirmation.html` - User confirmation template

Both templates support all form fields as template variables using Go's `text/template` syntax (e.g., `{{.full_name}}`, `{{.company_size}}`).#
# Retry Mechanism

The Go Submission Service includes a robust retry mechanism for handling email delivery failures. This ensures that form submissions are processed reliably even when email services experience temporary outages.

### How It Works

1. When a form is submitted, the service attempts to send confirmation emails immediately
2. If email sending succeeds, the submission is marked as `processed`
3. If email sending fails with a retryable error, the submission remains `pending`
4. A background job processor periodically checks for pending submissions and retries them

### Testing the Retry Mechanism

You can test the retry mechanism using the provided script:

```bash
./test-retry-mechanism.sh
```

This script:
1. Tests normal submission flow
2. Simulates email service failure by stopping MailHog
3. Verifies the service still accepts submissions
4. Restarts MailHog to allow background processing to succeed

### Example Retry Scenario

```javascript
// Submit form while email service is down
fetch('http://localhost:8080/api/submit', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    form_data: {
      form_type: 'business_contact',
      full_name: 'John Smith',
      work_email: 'john.smith@company.com',
      company: 'Tech Solutions Inc',
      company_size: '200-500'
    }
  })
})
.then(response => response.json())
.then(data => {
  // The submission is accepted even though email delivery failed
  console.log(data); // { "success": true, "message": "Form submitted successfully" }
});

// The email will be sent automatically when the email service comes back online
```

### Configuring Retry Behavior

You can configure retry behavior via environment variables:

```bash
# Retry configuration
export RETRY_MAX_ATTEMPTS=3       # Maximum number of retry attempts
export RETRY_INITIAL_DELAY_MS=1000 # Initial delay in milliseconds
export RETRY_MAX_DELAY_MS=30000   # Maximum delay in milliseconds
export RETRY_BACKOFF_FACTOR=2.0   # Exponential backoff multiplier
```

For more details, see the [retry mechanism documentation](../docs/retry-mechanism.md).