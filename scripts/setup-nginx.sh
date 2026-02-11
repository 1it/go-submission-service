#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Source environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment variables from .env"
    source .env
fi

# Configuration variables - use env vars if set, otherwise use defaults
DOMAIN_NAME="${DOMAIN_NAME:-api-submissions.1it.dev}"
HOST_PORT="${HOST_PORT:-80}"
EMAIL="${EMAIL:-admin@1it.dev}"

echo "=== Setting up Nginx with SSL for ${DOMAIN_NAME} ==="

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root or with sudo"
    exit 1
fi

# Install Nginx if not already installed
if ! command -v nginx &> /dev/null; then
    echo "Installing Nginx..."
    apt update
    apt install -y nginx
fi

# Install Certbot if not already installed
if ! command -v certbot &> /dev/null; then
    echo "Installing Certbot..."
    apt update
    apt install -y certbot python3-certbot-nginx
fi

# Create Nginx configuration file
echo "Creating Nginx configuration..."
cat > /etc/nginx/sites-available/beta-server.conf << EOF
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN_NAME};

    location / {
        proxy_pass http://localhost:${HOST_PORT};
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# Enable the site
echo "Enabling Nginx site..."
ln -sf /etc/nginx/sites-available/beta-server.conf /etc/nginx/sites-enabled/

# Remove default site if it exists
if [ -f /etc/nginx/sites-enabled/default ]; then
    echo "Removing default Nginx site..."
    rm /etc/nginx/sites-enabled/default
fi

# Test Nginx configuration
echo "Testing Nginx configuration..."
nginx -t

# Reload Nginx to apply changes
echo "Reloading Nginx..."
systemctl reload nginx

# Request SSL certificate using Certbot
echo "Requesting SSL certificate from Let's Encrypt..."
certbot --nginx -d ${DOMAIN_NAME} --agree-tos --email ${EMAIL} --redirect

echo "=== Nginx setup with Let's Encrypt SSL completed successfully ==="
echo "Your beta server is now accessible at https://${DOMAIN_NAME}" 