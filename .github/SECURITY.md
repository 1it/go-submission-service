# Security Policy

## Supported Versions

We release security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| older   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

If you believe you have found a security vulnerability, please report it by one of the following means:

1. **Preferred**: Open a [private security advisory](https://github.com/1it/go-submission-service/security/advisories/new) on GitHub.
2. If you cannot use the advisory: email the maintainers (see repository description or CONTRIBUTING.md for contact) with a clear description and steps to reproduce.

Please include:

- Type of issue (e.g. authentication bypass, injection, information disclosure)
- Full path of affected code or configuration
- Step-by-step instructions to reproduce
- Impact and suggested fix if you have one

We will acknowledge your report and work on a fix. We ask that you do not disclose the issue publicly until it has been addressed.

## Security Best Practices for Deployments

- Do not run with `ADMIN_ENDPOINTS_ENABLED=true` on public endpoints without strong auth and IP whitelisting.
- Use HTTPS in production and set `ADMIN_REQUIRE_HTTPS=true` when using admin APIs.
- Keep secrets (API keys, SMTP credentials, reCAPTCHA/Turnstile keys) in environment or secret managers, not in config files committed to git.
- Use a dedicated database user with minimal required permissions when using PostgreSQL.
