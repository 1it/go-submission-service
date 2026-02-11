# Email Templates Guide

The Go Submission Service includes a flexible email templating system that allows you to create and customize email templates for different purposes.

## Available Templates

The service comes with two default templates:

1. **form_submission.html** - Used for admin notifications about new form submissions
2. **confirmation.html** - Used for confirmation emails sent to the submitter

## Template Variables

Templates can access any variables provided in the form data. For example, if your form includes fields like `name`, `email`, and `message`, these will be available in the template.

Additionally, the following standard variables are always available:

- `timestamp` - The time the form was submitted (in RFC3339 format)
- `subject` - The email subject (can be overridden in the template)

## Creating Custom Templates

To create a custom template:

1. Create an HTML file in the `templates` directory
2. Use Go's template syntax to access form data variables
3. Use the template in your configuration or specify it in the form data

### Example Template

```html
<!DOCTYPE html>
<html>
<head>
    <title>Custom Template</title>
</head>
<body>
    <h1>Hello, {{.name}}!</h1>
    <p>Thank you for your submission.</p>
    <p>Your email: {{.email}}</p>
    <p>Your message: {{.message}}</p>
    <p>Submitted at: {{.timestamp}}</p>
</body>
</html>
```

## Using Templates

### Default Template

The default template is specified in the configuration:

```json
{
  "email": {
    "templates": {
      "directory": "./templates",
      "default_template": "form_submission.html",
      "subject": "Form Submission Received"
    }
  }
}
```

### Specifying a Template in Form Data

You can specify which template to use in the form data:

```json
{
  "form_data": {
    "name": "John Doe",
    "email": "john@example.com",
    "message": "Hello, world!",
    "template": "custom_template.html"
  }
}
```

### Admin Notifications

If you configure an admin email, the service will send a notification to that address whenever a form is submitted:

```json
{
  "email": {
    "admin_email": "admin@example.com"
  }
}
```

Or via environment variable:

```
ADMIN_EMAIL=admin@example.com
```

## Template Variables from Configuration

You can define default template variables in your configuration:

```json
{
  "email": {
    "templates": {
      "variables": {
        "company_name": "ACME Inc.",
        "website": "https://example.com",
        "support_email": "support@example.com"
      }
    }
  }
}
```

These variables will be available in all templates, but can be overridden by form data with the same keys.

## Customizing Email Subjects

The default subject is specified in the configuration, but you can override it for specific emails by including a `subject` field in your form data:

```json
{
  "form_data": {
    "name": "John Doe",
    "email": "john@example.com",
    "subject": "Custom Subject Line"
  }
}
```