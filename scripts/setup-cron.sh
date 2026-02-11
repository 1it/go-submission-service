#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Configuration variables
APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"  # Get absolute path to app directory
SCRIPTS_DIR="${APP_DIR}/scripts"
MONITOR_SCRIPT="${SCRIPTS_DIR}/monitor.sh"
CRON_SCHEDULE="${CRON_SCHEDULE:-*/15 * * * *}"  # Every 15 minutes by default
USER=$(whoami)

# Source environment variables from .env file if it exists
if [ -f "${APP_DIR}/.env" ]; then
    echo "Loading environment variables from .env"
    source "${APP_DIR}/.env"
fi

echo "=== Setting up monitoring cron job ==="
echo "Application directory: ${APP_DIR}"
echo "Monitor script: ${MONITOR_SCRIPT}"
echo "Cron schedule: ${CRON_SCHEDULE}"
echo "Running as user: ${USER}"

# Make sure the script is executable
chmod +x "${MONITOR_SCRIPT}"

# Create a temporary file for the crontab
TEMP_CRON=$(mktemp)

# Export existing crontab
crontab -l > "${TEMP_CRON}" 2>/dev/null || echo "# New crontab for ${USER}" > "${TEMP_CRON}"

# Check if the cron job already exists
if grep -q "${MONITOR_SCRIPT}" "${TEMP_CRON}"; then
    echo "Cron job already exists. Updating it..."
    sed -i.bak "/.*${MONITOR_SCRIPT//\//\\/}.*/d" "${TEMP_CRON}"
fi

# Add environment setting for the script
echo "# Beta server monitoring job - Added $(date)" >> "${TEMP_CRON}"
echo "SHELL=/bin/bash" >> "${TEMP_CRON}"
echo "PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin" >> "${TEMP_CRON}"

# Add environment variables
echo "VPS_USER=${VPS_USER:-root}" >> "${TEMP_CRON}"
echo "VPS_HOST=${VPS_HOST:-node-0.iaac.dev}" >> "${TEMP_CRON}"
echo "DOMAIN_NAME=${DOMAIN_NAME:-api-submissions.1it.dev}" >> "${TEMP_CRON}"
echo "CONTAINER_NAME=${CONTAINER_NAME:-beta-server}" >> "${TEMP_CRON}"
echo "LOG_DIR=${APP_DIR}/logs" >> "${TEMP_CRON}"
echo "ALERT_EMAIL=${ALERT_EMAIL:-admin@1it.dev}" >> "${TEMP_CRON}"
echo "TEST_EMAIL=${TEST_EMAIL:-admin@1it.dev}" >> "${TEMP_CRON}"

# Add the cron job
echo "${CRON_SCHEDULE} cd ${APP_DIR} && ${MONITOR_SCRIPT} > /dev/null 2>&1" >> "${TEMP_CRON}"

# Install the new crontab
crontab "${TEMP_CRON}"
rm "${TEMP_CRON}"

echo "=== Cron job installed successfully ==="
echo "To view current cron jobs, run: crontab -l"
echo "To edit cron jobs manually, run: crontab -e"

# Create the logs directory if it doesn't exist
mkdir -p "${APP_DIR}/logs"
echo "Log directory created at: ${APP_DIR}/logs"

# Show current crontab
echo "Current crontab entries:"
crontab -l 