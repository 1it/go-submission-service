# Deployment Guide

This directory contains deployment guides for different platforms and environments.

## Available Deployment Guides

- [Docker Deployment](docker.md) - Deploy using Docker and Docker Compose
- [Kubernetes Deployment](kubernetes.md) - Deploy to Kubernetes clusters
- [Cloud Providers](cloud-providers.md) - Deploy to AWS, GCP, Azure
- [Production Setup](production.md) - Production-ready configuration and best practices
- [SSL/TLS Configuration](ssl-tls.md) - Setting up HTTPS and certificates

## Quick Start

For most users, we recommend starting with the [Docker deployment guide](docker.md) for local development and testing, then moving to [Kubernetes](kubernetes.md) or [cloud providers](cloud-providers.md) for production deployments.

## Prerequisites

Before deploying, ensure you have:

- Go 1.23+ (for building from source)
- Docker and Docker Compose (for containerized deployments)
- Access to an email service (SMTP or MailerSend)
- Database storage (SQLite for development, PostgreSQL/MySQL for production)

## Environment Configuration

All deployment methods require proper environment configuration. See the [configuration guide](../configuration.md) for detailed information about:

- Environment variables
- Configuration files
- Security settings
- Email provider setup

## Support

If you encounter issues during deployment:

1. Check the troubleshooting section in each deployment guide
2. Review the [configuration documentation](../configuration.md)
3. Check the [GitHub issues](https://github.com/1it/go-submission-service/issues)
4. Join our [discussions](https://github.com/1it/go-submission-service/discussions)