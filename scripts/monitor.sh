#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Configuration variables - use env vars if set, otherwise use defaults
VPS_USER="${VPS_USER:-root}"
VPS_HOST="${VPS_HOST:-node-0.1it.dev}"
DOMAIN_NAME="${DOMAIN_NAME:-api-submissions.1it.dev}"
CONTAINER_NAME="${CONTAINER_NAME:-go-submission-service}"
TIMEOUT=5  # Timeout for curl requests in seconds
LOG_DIR="${LOG_DIR:-./logs}"
ALERT_EMAIL="${ALERT_EMAIL:-admin@1it.dev}"
TEST_EMAIL="${TEST_EMAIL:-admin@1it.dev}"  # Email for testing
MAX_CPU_PERCENT=80
MAX_MEM_PERCENT=80

# Source environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment variables from .env"
    source .env
fi

# Create logs directory if it doesn't exist
mkdir -p "${LOG_DIR}"

# Log file with timestamp
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
LOG_FILE="${LOG_DIR}/monitor_${TIMESTAMP}.log"

# Function to log message
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "${LOG_FILE}"
}

# Function to send alert
send_alert() {
    log "ALERT: $1"
    
    # Send email alert if mail command is available
    if command -v mail > /dev/null; then
        echo "$1" | mail -s "ALERT: Beta Server Status" "${ALERT_EMAIL}"
    fi
    
    # Add your preferred alerting mechanism here (Slack, SMS, etc.)
}

log "=== Beta Server Monitoring Check: ${TIMESTAMP} ==="
log "Domain: ${DOMAIN_NAME}"
log "Server: ${VPS_USER}@${VPS_HOST}"

# Check if server is accessible
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time ${TIMEOUT} https://${DOMAIN_NAME} || echo "000")
if [[ "$HTTP_STATUS" == "000" || "$HTTP_STATUS" == "5"* ]]; then
    send_alert "Server is not accessible or returning error: HTTP ${HTTP_STATUS}"
else
    log "Server is accessible: HTTP ${HTTP_STATUS}"
fi

# Check SSL certificate expiration
SSL_DAYS=$(curl -v --silent https://${DOMAIN_NAME} 2>&1 | grep "expire date" | awk '{print $4 " " $5 " " $6 " " $7}')
if [ -n "$SSL_DAYS" ]; then
    # Calculate days left
    EXPIRE_DATE=$(date -d "$SSL_DAYS" +%s)
    CURRENT_DATE=$(date +%s)
    DAYS_LEFT=$(( ($EXPIRE_DATE - $CURRENT_DATE) / 86400 ))
    
    log "SSL certificate expires in $DAYS_LEFT days"
    
    if [ $DAYS_LEFT -lt 14 ]; then
        send_alert "SSL certificate will expire in $DAYS_LEFT days"
    fi
else
    send_alert "Failed to get SSL certificate information"
fi

# Check if Docker container is running
ssh -q ${VPS_USER}@${VPS_HOST} "docker ps | grep -q ${CONTAINER_NAME}"
if [ $? -ne 0 ]; then
    send_alert "Docker container '${CONTAINER_NAME}' is not running"
else
    log "Docker container '${CONTAINER_NAME}' is running"
fi

# Check container resource usage
STATS=$(ssh -q ${VPS_USER}@${VPS_HOST} "docker stats ${CONTAINER_NAME} --no-stream --format '{{.CPUPerc}}|{{.MemPerc}}'")
if [ $? -eq 0 ]; then
    CPU_PERCENT=$(echo "$STATS" | cut -d'|' -f1 | sed 's/%//')
    MEM_PERCENT=$(echo "$STATS" | cut -d'|' -f2 | sed 's/%//')
    
    log "Resource usage: CPU ${CPU_PERCENT}%, Memory ${MEM_PERCENT}%"
    
    # Check if resource usage is too high
    if (( $(echo "$CPU_PERCENT > $MAX_CPU_PERCENT" | bc -l) )); then
        send_alert "CPU usage is too high: ${CPU_PERCENT}%"
    fi
    
    if (( $(echo "$MEM_PERCENT > $MAX_MEM_PERCENT" | bc -l) )); then
        send_alert "Memory usage is too high: ${MEM_PERCENT}%"
    fi
else
    send_alert "Failed to get container resource stats"
fi

# Check disk space
DISK_USAGE=$(ssh -q ${VPS_USER}@${VPS_HOST} "df -h / | tail -1 | awk '{print \$5}' | sed 's/%//'")
if [ $? -eq 0 ]; then
    log "Disk usage: ${DISK_USAGE}%"
    
    if [ $DISK_USAGE -gt 85 ]; then
        send_alert "Disk space is running low: ${DISK_USAGE}% used"
    fi
else
    send_alert "Failed to get disk usage information"
fi

# Test email submission functionality
log "Testing email submission functionality..."
# Generate a unique ID for this test
TEST_ID=$(date +%s)
# Prepare JSON payload for the API request
JSON_PAYLOAD="{\"email\":\"${TEST_EMAIL}\",\"platform\":\"monitor\",\"source\":\"health-check-${TEST_ID}\"}"

# Send the POST request to the beta signup endpoint
RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d "${JSON_PAYLOAD}" --max-time ${TIMEOUT} https://${DOMAIN_NAME}/api/beta/signup)
HTTP_STATUS=$?

# Check if curl command succeeded
if [ $HTTP_STATUS -eq 0 ]; then
    # Check if the response contains success indicators
    if echo "$RESPONSE" | grep -q "success\|ok\|true"; then
        log "Email submission test passed"
    else 
        send_alert "Email submission test failed with response: $RESPONSE"
    fi
else
    send_alert "Email submission test failed - connection error"
fi

# Check for errors in container logs (last 30 minutes)
TIMESTAMP_30MIN_AGO=$(date -d "30 minutes ago" "+%Y-%m-%dT%H:%M:%S")
CONTAINER_ERRORS=$(ssh -q ${VPS_USER}@${VPS_HOST} "docker logs --since ${TIMESTAMP_30MIN_AGO} ${CONTAINER_NAME} 2>&1 | grep -i 'error\|panic\|fatal' | wc -l")
if [ $? -eq 0 ]; then
    log "Container error count (last 30 min): ${CONTAINER_ERRORS}"
    
    if [ $CONTAINER_ERRORS -gt 0 ]; then
        ERRORS=$(ssh -q ${VPS_USER}@${VPS_HOST} "docker logs --since ${TIMESTAMP_30MIN_AGO} ${CONTAINER_NAME} 2>&1 | grep -i 'error\|panic\|fatal' | head -5")
        send_alert "Found ${CONTAINER_ERRORS} errors in container logs. First 5 errors:\n${ERRORS}"
    fi
else
    send_alert "Failed to check container logs"
fi

# Verify the test email submission in logs
EMAIL_LOG_CHECK=$(ssh -q ${VPS_USER}@${VPS_HOST} "docker logs ${CONTAINER_NAME} --since 1m 2>&1 | grep -c '${TEST_EMAIL}'")
if [ "$EMAIL_LOG_CHECK" -gt 0 ]; then
    log "Test email submission successfully processed"
else
    send_alert "Test email submission not found in logs - potential processing issue"
fi

log "=== Monitoring check completed ===" 