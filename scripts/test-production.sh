#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Configuration variables - use env vars if set, otherwise use defaults
VPS_USER="${VPS_USER:-root}"
VPS_HOST="${VPS_HOST:-node-0.1it.dev}"
DOMAIN_NAME="${DOMAIN_NAME:-api-submissions.1it.dev}"
CONTAINER_NAME="${CONTAINER_NAME:-go-submission-service}"
TIMEOUT=5  # Timeout for curl requests in seconds
TEST_EMAIL="${TEST_EMAIL:-admin@1it.dev}"  # Email for testing

# Source environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment variables from .env"
    source .env
fi

echo "=== Testing production deployment ==="
echo "Domain: ${DOMAIN_NAME}"
echo "Server: ${VPS_USER}@${VPS_HOST}"

# Function to print test result
test_result() {
    if [ $1 -eq 0 ]; then
        echo "✅ $2"
    else
        echo "❌ $2"
        FAILED=1
    fi
}

# Initialize failure flag
FAILED=0

# Test 1: Check if the server is accessible via HTTPS
echo -n "Testing HTTPS accessibility... "
if curl -s -o /dev/null -w "%{http_code}" --max-time ${TIMEOUT} https://${DOMAIN_NAME}/health | grep -q "200\|30[0-9]"; then
    test_result 0 "Server is accessible via HTTPS"
else
    test_result 1 "Server is not accessible via HTTPS"
fi

# Test 2: Check SSL certificate validity
echo -n "Testing SSL certificate validity... "
CERT_EXPIRY=$(curl -v --silent https://${DOMAIN_NAME}/health 2>&1 | grep "expire date" || echo "")
if [ -n "$CERT_EXPIRY" ]; then
    test_result 0 "SSL certificate is valid: $CERT_EXPIRY"
else
    test_result 1 "Failed to get SSL certificate information"
fi

# Test 3: Check server connectivity for key endpoints (adjust these to match your actual API endpoints)
echo -n "Testing API endpoint response... "
if curl -s -o /dev/null -w "%{http_code}" --max-time ${TIMEOUT} https://${DOMAIN_NAME}/api/beta/check | grep -q "200\|404"; then
    test_result 0 "API endpoint is responding"
else
    test_result 1 "API endpoint is not responding"
fi

# Test 4: Check if Docker container is running on the server
echo -n "Checking if container is running on server... "
ssh -q ${VPS_USER}@${VPS_HOST} "docker ps | grep -q ${CONTAINER_NAME}"
test_result $? "Docker container '${CONTAINER_NAME}' status"

# Test 5: Check container logs for errors
echo -n "Checking container logs for errors... "
CONTAINER_ERRORS=$(ssh -q ${VPS_USER}@${VPS_HOST} "docker logs ${CONTAINER_NAME} 2>&1 | grep -i 'error\|panic\|fatal' | wc -l")
if [ "$CONTAINER_ERRORS" -gt 0 ]; then
    test_result 1 "Found $CONTAINER_ERRORS error messages in container logs"
    echo "Last 5 errors from logs:"
    ssh -q ${VPS_USER}@${VPS_HOST} "docker logs ${CONTAINER_NAME} 2>&1 | grep -i 'error\|panic\|fatal' | tail -5"
else
    test_result 0 "No errors found in container logs"
fi

# Test 6: Check server resources
echo -n "Checking server resources... "
ssh -q ${VPS_USER}@${VPS_HOST} "docker stats ${CONTAINER_NAME} --no-stream --format \"Memory usage: {{.MemPerc}}%, CPU usage: {{.CPUPerc}}\""
test_result $? "Resource usage"

# Test 7: Email submission test
echo -n "Testing email submission with ${TEST_EMAIL}... "
# Generate a unique ID for this test submission
TEST_ID=$(date +%s)
# Prepare JSON payload for the API request
JSON_PAYLOAD="{\"email\":\"${TEST_EMAIL}\",\"platform\":\"test\",\"source\":\"e2e-test-${TEST_ID}\"}"

# Send the POST request to the beta signup endpoint
RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d "${JSON_PAYLOAD}" --max-time ${TIMEOUT} https://${DOMAIN_NAME}/api/beta/signup)
HTTP_STATUS=$?

# Check if curl command succeeded
if [ $HTTP_STATUS -eq 0 ]; then
    # Check if the response contains success indicators
    if echo "$RESPONSE" | grep -q "success\|ok\|true"; then
        test_result 0 "Email submission successful: $RESPONSE"
    else 
        test_result 1 "Email submission failed with response: $RESPONSE"
    fi
else
    test_result 1 "Failed to submit email - connection error"
fi

# Test 8: Check if test email appears in logs
echo -n "Verifying email submission in logs... "
EMAIL_LOG_CHECK=$(ssh -q ${VPS_USER}@${VPS_HOST} "docker logs ${CONTAINER_NAME} --since 1m 2>&1 | grep -c '${TEST_EMAIL}'")
if [ "$EMAIL_LOG_CHECK" -gt 0 ]; then
    test_result 0 "Found test email in logs - submission processed"
else
    test_result 1 "Test email not found in recent logs"
fi

# Summary
echo "=== Test Summary ==="
if [ $FAILED -eq 0 ]; then
    echo "✅ All tests passed successfully!"
    exit 0
else
    echo "❌ Some tests failed. Please check the log above for details."
    exit 1
fi 