#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Configuration variables - use env vars if set, otherwise use defaults
VPS_USER="${VPS_USER:-root}"
VPS_HOST="${VPS_HOST:-node-0.iaac.dev}"
REGISTRY_DOMAIN="${REGISTRY_DOMAIN:-hub.docker.com}"
REGISTRY_USER="${REGISTRY_USER:-1it}"
REGISTRY_PASSWORD="${REGISTRY_PASSWORD:-1it}"
IMAGE_NAME="${IMAGE_NAME:-1it/go-submission-service}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
CONTAINER_NAME="${CONTAINER_NAME:-go-submission-service}"
HOST_PORT="${HOST_PORT:-8080}"
CONTAINER_PORT="${CONTAINER_PORT:-8080}"

# Email configuration
EMAIL_SERVICE="${EMAIL_SERVICE:-mailersend}"
MAILERSEND_API_KEY="${MAILERSEND_API_KEY:-your_api_key_here}"
MAILERSEND_FROM_EMAIL="${MAILERSEND_FROM_EMAIL:-support@1it.dev}"
MAILERSEND_FROM_NAME="${MAILERSEND_FROM_NAME:-Go Submission Service}"
MAILERSEND_SUBJECT="${MAILERSEND_SUBJECT:-Welcome to the Go Submission Service}"

# reCAPTCHA configuration
RECAPTCHA_SECRET_KEY="${RECAPTCHA_SECRET_KEY:-1x0000000000000000000000000000000AA}"
RECAPTCHA_PROJECT_ID="${RECAPTCHA_PROJECT_ID:-10000000000000000000000000000000}"
RECAPTCHA_SITE_KEY="${RECAPTCHA_SITE_KEY:-10000000000000000000000000000000}"
RECAPTCHA_ACTION_NAME="${RECAPTCHA_ACTION_NAME:-form_submit}"
RECAPTCHA_ENTERPRISE_ENABLED="${RECAPTCHA_ENTERPRISE_ENABLED:-false}"

# Source environment variables from .env file if it exists
if [ -f .env.deploy ]; then
    echo "Loading environment variables from .env"
    source .env.deploy
fi

# Full image reference
FULL_IMAGE="${REGISTRY_DOMAIN}/${IMAGE_NAME}:${IMAGE_TAG}"

echo "=== Deploying ${FULL_IMAGE} to ${VPS_HOST} ==="

# SSH into the VPS and execute commands
echo "Connecting to ${VPS_USER}@${VPS_HOST}..."
ssh ${VPS_USER}@${VPS_HOST} << EOF
    echo "Logging in to Docker registry..."
    echo "${REGISTRY_PASSWORD}" | docker login ${REGISTRY_DOMAIN} -u ${REGISTRY_USER} --password-stdin
    
    echo "Pulling latest image: ${FULL_IMAGE}"
    docker pull ${FULL_IMAGE}
    
    echo "Creating Docker volume for persistent data if it doesn't exist..."
    docker volume create beta-server-data || true
    
    echo "Stopping and removing existing container if it exists..."
    docker rm -f ${CONTAINER_NAME} || true
    
    echo "Starting new container..."
    docker run -d \\
      --name ${CONTAINER_NAME} \\
      --dns 8.8.8.8 --dns 8.8.4.4 \\
      -p 127.0.0.1:${HOST_PORT}:${CONTAINER_PORT} \\
      -v beta-server-data:/app/data \\
      --env PORT=${CONTAINER_PORT} \\
      --env DB_PATH=/app/data/subscribers.db \\
      --env EMAIL_SERVICE="${EMAIL_SERVICE}" \\
      --env MAILERSEND_API_KEY="${MAILERSEND_API_KEY}" \\
      --env MAILERSEND_FROM_EMAIL="${MAILERSEND_FROM_EMAIL}" \\
      --env MAILERSEND_FROM_NAME="${MAILERSEND_FROM_NAME}" \\
      --env MAILERSEND_SUBJECT="${MAILERSEND_SUBJECT}" \\
      --env PLAY_STORE_BETA_LINK="${PLAY_STORE_BETA_LINK}" \\
      --env TESTFLIGHT_LINK="${TESTFLIGHT_LINK}" \\
      --env RECAPTCHA_SECRET_KEY="${RECAPTCHA_SECRET_KEY}" \\
      --env RECAPTCHA_PROJECT_ID="${RECAPTCHA_PROJECT_ID}" \\
      --env RECAPTCHA_SITE_KEY="${RECAPTCHA_SITE_KEY}" \\
      --env RECAPTCHA_ACTION_NAME="${RECAPTCHA_ACTION_NAME}" \\
      --env RECAPTCHA_ENTERPRISE_ENABLED="${RECAPTCHA_ENTERPRISE_ENABLED}" \\
      --env GIN_MODE=release \\
      --restart unless-stopped \\
      ${FULL_IMAGE}

    echo "Checking if the container is running..."
    docker ps | grep ${CONTAINER_NAME}
EOF

echo "=== Deployment completed successfully ===" 