#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# This script allows easy access to the CLI tool inside the Docker container
# Usage: ./cli.sh [command] [args]
# Example: ./cli.sh list --status=invited

CONTAINER_NAME="${CONTAINER_NAME:-beta-server}"

# Check if the container is running
if ! docker ps | grep -q "${CONTAINER_NAME}"; then
    echo "Error: Container '${CONTAINER_NAME}' is not running"
    exit 1
fi

# Execute the CLI tool inside the container
echo "Running CLI command in container ${CONTAINER_NAME}: ./cli $@"
docker exec -it ${CONTAINER_NAME} ./cli "$@" 