# Retry Mechanism & Error Handling

The Go Submission Service includes a robust retry mechanism and error handling system to ensure reliable form submission processing even when external services (like email providers) experience temporary failures.

## Overview

The retry system is designed to handle transient failures gracefully, particularly for email delivery, which is a critical but potentially unreliable part of the form submission process. The system includes:

1. **Immediate Retry**: Quick retries during the initial request
2. **Background Processing**: Asynchronous processing of failed submissions
3. **Exponential Backoff**: Smart delay calculation between retry attempts
4. **Error Classification**: Distinguishing between retryable and non-retryable errors
5. **Configurable Policies**: Customizable retry parameters

## Configuration

Retry behavior can be configured via environment variables or in the `config.json` file:

```bash
# Retry configuration
export RETRY_MAX_ATTEMPTS=3       # Maximum number of retry attempts
export RETRY_INITIAL_DELAY_MS=1000 # Initial delay in milliseconds
export RETRY_MAX_DELAY_MS=30000   # Maximum delay in milliseconds
export RETRY_BACKOFF_FACTOR=2.0   # Exponential backoff multiplier
```

Or in `config.json`:

```json
{
  "retry": {
    "max_attempts": 3,
    "initial_delay_ms": 1000,
    "max_delay_ms": 30000,
    "backoff_factor": 2.0
  }
}
```

## How It Works

### 1. Submission Flow

When a form is submitted:

1. The form data is validated and stored in the database with status `pending`
2. The service attempts to send confirmation emails immediately
3. If email sending succeeds, the submission is marked as `processed`
4. If email sending fails with a retryable error, the submission remains `pending`
5. A background job processor periodically checks for pending submissions and retries them

### 2. Retryable vs. Non-Retryable Errors

The system classifies errors into two categories:

**Retryable Errors** (will be retried):
- HTTP 5xx server errors
- Network timeouts
- Connection refused/reset
- DNS resolution failures
- "Service unavailable" responses
- Rate limiting (temporary)

**Non-Retryable Errors** (will not be retried):
- HTTP 4xx client errors
- Authentication failures
- Invalid email addresses
- Permanent quota exceeded errors

### 3. Background Job Processing

A background job processor runs every 5 minutes to:

1. Find submissions older than 5 minutes with status `pending`
2. Attempt to process them with the configured retry policy
3. Mark successful submissions as `processed`
4. Mark submissions older than 24 hours as `failed` if they still can't be processed

### 4. Exponential Backoff

The retry mechanism uses exponential backoff to increase the delay between retry attempts:

```
delay = initial_delay * (backoff_factor ^ (attempt - 1))
```

For example, with default settings:
- 1st retry: 1 second delay
- 2nd retry: 2 seconds delay
- 3rd retry: 4 seconds delay

The delay is capped at the configured maximum delay.

## Monitoring

### Logs

Retry attempts are logged with detailed information:

```
2025/07/18 15:34:53 Executing email for submission abc123 (attempt 1/3)
2025/07/18 15:34:53 email for submission abc123 failed (attempt 1/3): connection refused. Retrying in 1s
2025/07/18 15:34:54 Executing email for submission abc123 (attempt 2/3)
2025/07/18 15:34:54 email for submission abc123 succeeded on attempt 2
```

### Health Check

The `/health` endpoint includes information about the retry mechanism:

```json
{
  "status": "ok",
  "timestamp": "2025-07-18T15:34:53Z",
  "service": "go-submission-service",
  "version": "1.0.0",
  "features": {
    "email_retry": true,
    "background_jobs": true,
    "form_validation": true,
    "database_migration": true
  }
}
```

## Testing

You can test the retry mechanism using the provided script:

```bash
./test-retry-mechanism.sh
```

This script:
1. Tests normal submission flow
2. Simulates email service failure by stopping MailHog
3. Verifies the service still accepts submissions
4. Restarts MailHog to allow background processing to succeed

## Best Practices

1. **Configure Appropriate Timeouts**: Set reasonable timeouts for email services
2. **Monitor Failed Submissions**: Regularly check for submissions stuck in `pending` or `failed` status
3. **Adjust Retry Parameters**: Tune retry parameters based on your email provider's characteristics
4. **Use Reliable Email Services**: Choose email providers with good uptime and clear error responses

## Implementation Details

The retry mechanism is implemented in the following components:

- `internal/retry/retry.go`: Core retry logic and error classification
- `internal/jobs/processor.go`: Background job processing
- `internal/handlers/submit.go`: Integration with submission handler
- `internal/config/config.go`: Configuration options

## Limitations

1. The background processor runs in-process and is not distributed
2. Failed submissions are not automatically retried after being marked as `failed`
3. There is no admin UI for managing failed submissions (use CLI or direct database access)

## Future Improvements

- Add webhook notifications for failed submissions
- Implement a distributed job queue for better scalability
- Create an admin UI for managing and retrying failed submissions
- Add more granular metrics for retry attempts and failures