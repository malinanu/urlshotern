#!/bin/bash

# URL Shortener Deployment Script for Oracle VM
# Usage: ./deploy.sh [shared|separate] [domain]

set -e

# Configuration
DEPLOYMENT_TYPE=${1:-shared}
DOMAIN=${2:-yourdomain.com}
PROJECT_DIR="/home/urlshortener/URLshorter"
DOCKER_DIR="$PROJECT_DIR/docker"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
    exit 1
}

# Check if running as correct user
check_user() {
    if [ "$USER" != "urlshortener" ]; then
        error "This script should be run as the 'urlshortener' user"
    fi
}

# Check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        error "Docker is not running. Please start Docker and try again."
    fi
}

# Check if environment file exists
check_env_file() {
    if [ ! -f "$PROJECT_DIR/.env.prod" ]; then
        warn "Production environment file not found at $PROJECT_DIR/.env.prod"
        echo "Please copy .env.prod.example to .env.prod and configure it:"
        echo "cp $PROJECT_DIR/.env.prod.example $PROJECT_DIR/.env.prod"
        echo "nano $PROJECT_DIR/.env.prod"
        exit 1
    fi
}

# Create shared network if it doesn't exist
create_shared_network() {
    if [ "$DEPLOYMENT_TYPE" == "shared" ]; then
        if ! docker network ls | grep -q "shared-backend"; then
            log "Creating shared-backend network..."
            docker network create shared-backend
        else
            log "shared-backend network already exists"
        fi
    fi
}

# Update nginx configuration with domain
update_nginx_config() {
    log "Updating nginx configuration with domain: $DOMAIN"
    sed -i "s/yourdomain\.com/$DOMAIN/g" "$DOCKER_DIR/nginx.prod.conf"
}

# Deploy the application
deploy_application() {
    log "Starting deployment of URL shortener ($DEPLOYMENT_TYPE mode)..."

    cd "$DOCKER_DIR"

    # Load environment variables
    export $(cat "$PROJECT_DIR/.env.prod" | grep -v '^#' | xargs)

    # Choose the right compose file
    if [ "$DEPLOYMENT_TYPE" == "shared" ]; then
        COMPOSE_FILE="docker-compose.shared.yml"
        log "Using shared PostgreSQL deployment"
    else
        COMPOSE_FILE="docker-compose.separate.yml"
        log "Using separate PostgreSQL deployment"
    fi

    # Stop existing services
    log "Stopping existing services..."
    docker-compose -f "$COMPOSE_FILE" down --remove-orphans

    # Build and start services
    log "Building and starting services..."
    docker-compose -f "$COMPOSE_FILE" up -d --build

    # Wait for services to be healthy
    log "Waiting for services to start..."
    sleep 30

    # Check if services are running
    docker-compose -f "$COMPOSE_FILE" ps
}

# Setup SSL certificate
setup_ssl() {
    log "Setting up SSL certificate for $DOMAIN..."

    # Stop nginx temporarily
    docker-compose -f "$COMPOSE_FILE" stop nginx

    # Get SSL certificate
    if sudo certbot certonly --standalone -d "$DOMAIN" -d "www.$DOMAIN" --non-interactive --agree-tos --email "admin@$DOMAIN"; then
        log "SSL certificate obtained successfully"
    else
        warn "Failed to obtain SSL certificate. You may need to configure it manually."
    fi

    # Start nginx again
    docker-compose -f "$COMPOSE_FILE" start nginx
}

# Test deployment
test_deployment() {
    log "Testing deployment..."

    # Test health endpoint
    if curl -f "http://localhost:8080/health" >/dev/null 2>&1; then
        log "Health check passed"
    else
        error "Health check failed"
    fi

    # Test with domain if SSL is configured
    if [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ]; then
        if curl -f "https://$DOMAIN/health" >/dev/null 2>&1; then
            log "HTTPS health check passed"
        else
            warn "HTTPS health check failed"
        fi
    fi
}

# Setup database for shared deployment
setup_shared_database() {
    if [ "$DEPLOYMENT_TYPE" == "shared" ]; then
        log "Please ensure you have created the database and user in your Odoo PostgreSQL:"
        echo ""
        echo "Connect to your Odoo PostgreSQL container and run:"
        echo "CREATE USER urlshortener WITH PASSWORD 'your_password';"
        echo "CREATE DATABASE urlshortener_db OWNER urlshortener;"
        echo "GRANT ALL PRIVILEGES ON DATABASE urlshortener_db TO urlshortener;"
        echo ""
        echo "Then run the schema initialization:"
        echo "docker exec -i <odoo_postgres_container> psql -U urlshortener -d urlshortener_db < $PROJECT_DIR/configs/database.sql"
        echo ""
        read -p "Press Enter after setting up the database..."
    fi
}

# Main deployment process
main() {
    echo -e "${BLUE}"
    echo "========================================"
    echo "    URL Shortener Deployment Script    "
    echo "========================================"
    echo -e "${NC}"

    log "Deployment Type: $DEPLOYMENT_TYPE"
    log "Domain: $DOMAIN"
    log "Project Directory: $PROJECT_DIR"

    check_user
    check_docker
    check_env_file
    create_shared_network
    setup_shared_database
    update_nginx_config
    deploy_application

    # Ask about SSL setup
    echo ""
    read -p "Do you want to setup SSL certificate? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        setup_ssl
    fi

    test_deployment

    echo ""
    log "Deployment completed successfully!"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Configure your domain DNS to point to this server's IP"
    echo "2. Test your application at http://$DOMAIN or https://$DOMAIN"
    echo "3. Monitor logs: docker-compose -f $COMPOSE_FILE logs -f"
    echo "4. Set up backups using the backup script in the deployment guide"
    echo ""
    echo -e "${GREEN}Your URL shortener is now running! 🚀${NC}"
}

# Show usage if invalid arguments
if [ "$DEPLOYMENT_TYPE" != "shared" ] && [ "$DEPLOYMENT_TYPE" != "separate" ]; then
    echo "Usage: $0 [shared|separate] [domain]"
    echo ""
    echo "Examples:"
    echo "  $0 shared myurl.com     # Deploy with shared PostgreSQL"
    echo "  $0 separate myurl.com   # Deploy with separate PostgreSQL"
    echo ""
    exit 1
fi

# Run main function
main