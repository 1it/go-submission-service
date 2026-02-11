#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Configuration variables - use env vars if set, otherwise use defaults
# REGISTRY_DOMAIN="${REGISTRY_DOMAIN:-hub.docker.com}"
# REGISTRY_USER="${REGISTRY_USER:-1it}"
# REGISTRY_PASSWORD="${REGISTRY_PASSWORD:-1it}"
IMAGE_NAME="${IMAGE_NAME:-1it/go-submission-service}"
IMAGE_TAG="${IMAGE_TAG:-latest}"

# Source environment variables from .env file if it exists
echo "Loading environment variables from .env"
source .env

# Full image reference
FULL_IMAGE="${REGISTRY_DOMAIN}/${IMAGE_NAME}:${IMAGE_TAG}"

echo "=== Building and pushing Docker image ==="
echo "Image: ${FULL_IMAGE}"
 
# Login to registry
echo "Logging in to registry ${REGISTRY_DOMAIN}..."
echo "${REGISTRY_PASSWORD}" | docker login ${REGISTRY_DOMAIN} -u ${REGISTRY_USER} --password-stdin

# Build the image
echo "Building Docker image..."
docker build -t ${FULL_IMAGE} .

# Push the image
echo "Pushing image to registry..."
docker push ${FULL_IMAGE}

echo "=== Build and push completed successfully ==="
echo "Image is available at: ${FULL_IMAGE}" 