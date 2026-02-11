#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Source environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment variables from .env"
    source .env
fi

# Configuration variables - use env vars if set, otherwise use defaults
REGISTRY_DOMAIN="${REGISTRY_DOMAIN:-registry.1it.dev}"
REGISTRY_PORT="${REGISTRY_PORT:-5000}"
REGISTRY_USER="${REGISTRY_USER:-1it}"
REGISTRY_PASSWORD="${REGISTRY_PASSWORD:-1it}"
EMAIL="${EMAIL:-admin@1it.dev}"
REGISTRY_PATH="/var/lib/registry"

echo "=== Setting up private Docker Registry on domain: ${REGISTRY_DOMAIN} ==="

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root or with sudo"
    exit 1
fi

# Create registry directories
echo "Creating registry directories..."
mkdir -p ${REGISTRY_PATH}/{data,auth,certs}

# Generate htpasswd file for basic authentication
echo "Setting up registry authentication..."
docker run --rm --entrypoint htpasswd httpd:alpine -Bbn ${REGISTRY_USER} ${REGISTRY_PASSWORD} > ${REGISTRY_PATH}/auth/htpasswd

# Create Nginx configuration for the registry
echo "Creating Nginx configuration for registry..."
cat > /etc/nginx/sites-available/docker-registry.conf << EOF
server {
    listen 80;
    listen [::]:80;
    server_name ${REGISTRY_DOMAIN};
    
    location / {
        return 301 https://\$host\$request_uri;
    }
}
EOF

# Enable the site
ln -sf /etc/nginx/sites-available/docker-registry.conf /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx

# Request SSL certificate using Certbot
echo "Requesting SSL certificate from Let's Encrypt..."
certbot --nginx -d ${REGISTRY_DOMAIN} --agree-tos --email ${EMAIL} --redirect

# Update Nginx configuration with proxy settings
echo "Updating Nginx configuration with proxy settings..."
cat > /etc/nginx/sites-available/docker-registry.conf << EOF
server {
    listen 80;
    listen [::]:80;
    server_name ${REGISTRY_DOMAIN};
    return 301 https://\$host\$request_uri;
}

server {
    listen 443 ssl;
    listen [::]:443 ssl;
    server_name ${REGISTRY_DOMAIN};

    # SSL Configuration (managed by Certbot)
    ssl_certificate /etc/letsencrypt/live/${REGISTRY_DOMAIN}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/${REGISTRY_DOMAIN}/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    # Increase max upload size
    client_max_body_size 500M;

    # Registry proxy
    location / {
        proxy_pass http://localhost:${REGISTRY_PORT};
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        chunked_transfer_encoding off;
        auth_basic "Registry";
        auth_basic_user_file ${REGISTRY_PATH}/auth/htpasswd;
    }
}
EOF

systemctl reload nginx

# Start the registry container
echo "Starting registry container..."
docker rm -f registry || true
docker run -d \
  --name registry \
  --restart=always \
  -p ${REGISTRY_PORT}:5000 \
  -v ${REGISTRY_PATH}/data:/var/lib/registry \
  -v ${REGISTRY_PATH}/auth:/auth \
  -e "REGISTRY_AUTH=htpasswd" \
  -e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm" \
  -e "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd" \
  registry:2

echo "=== Registry setup completed successfully ==="
echo "Registry URL: https://${REGISTRY_DOMAIN}"
echo "Username: ${REGISTRY_USER}"
echo "Password: ${REGISTRY_PASSWORD}" 