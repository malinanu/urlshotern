#!/bin/bash
# Secret Rotation Automation Script
# This script helps automate common secret rotation tasks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG_FILE="${PROJECT_ROOT}/logs/secret-rotation.log"

# Ensure log directory exists
mkdir -p "${PROJECT_ROOT}/logs"

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}[SUCCESS] $1${NC}" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}[WARNING] $1${NC}" | tee -a "$LOG_FILE"
}

# Generate secure random string
generate_secret() {
    local length=${1:-32}
    openssl rand -base64 $length | tr -d "=+/" | cut -c1-$length
}

# Generate secure password (alphanumeric + special chars)
generate_password() {
    openssl rand -base64 32 | head -c 32
}

# Check if running in correct environment
check_environment() {
    if [ ! -f "${PROJECT_ROOT}/.env" ]; then
        error ".env file not found in project root"
        exit 1
    fi

    # Source environment variables
    set -a
    source "${PROJECT_ROOT}/.env"
    set +a
}

# Display menu
show_menu() {
    echo ""
    echo "====================================="
    echo "    Secret Rotation Tool"
    echo "====================================="
    echo "1. Rotate Database Password"
    echo "2. Rotate JWT Secret"
    echo "3. Rotate Redis Password"
    echo "4. Rotate API Keys (Bulk)"
    echo "5. Check Secret Expiry Status"
    echo "6. Emergency Rotation (All Secrets)"
    echo "7. Exit"
    echo "====================================="
    echo -n "Select option: "
}

# Rotate database password
rotate_database_password() {
    log "Starting database password rotation..."

    # Generate new password
    NEW_PASSWORD=$(generate_password)

    echo -n "Enter PostgreSQL admin username (default: postgres): "
    read PG_ADMIN_USER
    PG_ADMIN_USER=${PG_ADMIN_USER:-postgres}

    echo -n "Enter PostgreSQL host (default: localhost): "
    read PG_HOST
    PG_HOST=${PG_HOST:-localhost}

    echo -n "Enter database name (default: urlshortener): "
    read DB_NAME
    DB_NAME=${DB_NAME:-urlshortener}

    echo -n "Enter database user to rotate (default: urlshortener): "
    read DB_USER
    DB_USER=${DB_USER:-urlshortener}

    warning "This will change the password for user: $DB_USER"
    echo -n "Continue? (yes/no): "
    read confirm
    if [ "$confirm" != "yes" ]; then
        log "Database password rotation aborted"
        return
    fi

    # Update password in PostgreSQL
    log "Updating password in PostgreSQL..."
    PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "$PG_HOST" -U "$PG_ADMIN_USER" -d "$DB_NAME" <<EOF
ALTER USER $DB_USER WITH PASSWORD '$NEW_PASSWORD';
EOF

    if [ $? -eq 0 ]; then
        success "Database password updated in PostgreSQL"

        # Update .env file
        log "Updating .env file..."
        sed -i.bak "s/^DB_PASSWORD=.*/DB_PASSWORD=$NEW_PASSWORD/" "${PROJECT_ROOT}/.env"

        success "Database password rotated successfully"
        warning "New password: $NEW_PASSWORD"
        warning "IMPORTANT: Store this password securely and update all services using it"
        warning "Old .env backed up to .env.bak"

        # Log rotation event
        log "Database password rotation completed for user: $DB_USER"
    else
        error "Failed to update database password"
        exit 1
    fi
}

# Rotate JWT secret
rotate_jwt_secret() {
    log "Starting JWT secret rotation..."

    # Generate new secret
    NEW_JWT_SECRET=$(generate_secret 64)

    # Get current secret
    CURRENT_JWT_SECRET=$(grep "^JWT_SECRET=" "${PROJECT_ROOT}/.env" | cut -d '=' -f2)

    if [ -z "$CURRENT_JWT_SECRET" ]; then
        error "Current JWT secret not found in .env"
        exit 1
    fi

    warning "This will rotate the JWT secret with a grace period"
    echo "During the grace period, both old and new secrets will be valid"
    echo -n "Continue? (yes/no): "
    read confirm
    if [ "$confirm" != "yes" ]; then
        log "JWT secret rotation aborted"
        return
    fi

    # Update .env with dual-secret configuration
    log "Setting up dual-secret configuration..."
    sed -i.bak "s/^JWT_SECRET=.*/JWT_SECRET=$NEW_JWT_SECRET/" "${PROJECT_ROOT}/.env"

    # Add old secret as backup
    if ! grep -q "^JWT_SECRET_OLD=" "${PROJECT_ROOT}/.env"; then
        echo "JWT_SECRET_OLD=$CURRENT_JWT_SECRET" >> "${PROJECT_ROOT}/.env"
    else
        sed -i "s/^JWT_SECRET_OLD=.*/JWT_SECRET_OLD=$CURRENT_JWT_SECRET/" "${PROJECT_ROOT}/.env"
    fi

    success "JWT secret rotated successfully"
    warning "New JWT_SECRET: $NEW_JWT_SECRET"
    warning "Old secret stored in JWT_SECRET_OLD for grace period"
    warning "IMPORTANT:"
    echo "  1. Restart your application to apply changes"
    echo "  2. Wait 48 hours for all tokens to expire"
    echo "  3. Remove JWT_SECRET_OLD from .env"
    warning "Old .env backed up to .env.bak"

    # Log rotation event
    log "JWT secret rotation completed with grace period"
}

# Rotate Redis password
rotate_redis_password() {
    log "Starting Redis password rotation..."

    # Generate new password
    NEW_PASSWORD=$(generate_password)

    echo -n "Enter Redis host (default: localhost): "
    read REDIS_HOST
    REDIS_HOST=${REDIS_HOST:-localhost}

    echo -n "Enter Redis port (default: 6379): "
    read REDIS_PORT
    REDIS_PORT=${REDIS_PORT:-6379}

    warning "This will change the Redis password"
    warning "Ensure you have access to Redis configuration file"
    echo -n "Continue? (yes/no): "
    read confirm
    if [ "$confirm" != "yes" ]; then
        log "Redis password rotation aborted"
        return
    fi

    # Update Redis configuration
    log "Updating Redis password..."
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" CONFIG SET requirepass "$NEW_PASSWORD"

    if [ $? -eq 0 ]; then
        success "Redis password updated"

        # Update .env file
        log "Updating .env file..."
        sed -i.bak "s/^REDIS_PASSWORD=.*/REDIS_PASSWORD=$NEW_PASSWORD/" "${PROJECT_ROOT}/.env"

        success "Redis password rotated successfully"
        warning "New password: $NEW_PASSWORD"
        warning "IMPORTANT: Update redis.conf to persist across restarts"
        warning "Add or update: requirepass $NEW_PASSWORD"
        warning "Old .env backed up to .env.bak"

        # Log rotation event
        log "Redis password rotation completed"
    else
        error "Failed to update Redis password"
        exit 1
    fi
}

# Rotate API keys (bulk)
rotate_api_keys() {
    log "Starting bulk API key rotation..."

    echo "This will rotate API keys for users that are approaching expiration"
    echo -n "Days before expiration to rotate (default: 30): "
    read DAYS_BEFORE
    DAYS_BEFORE=${DAYS_BEFORE:-30}

    echo -n "Grace period in days (default: 30): "
    read GRACE_PERIOD
    GRACE_PERIOD=${GRACE_PERIOD:-30}

    warning "This will rotate all API keys expiring within $DAYS_BEFORE days"
    echo -n "Continue? (yes/no): "
    read confirm
    if [ "$confirm" != "yes" ]; then
        log "API key rotation aborted"
        return
    fi

    # Connect to database and find expiring keys
    log "Finding API keys expiring within $DAYS_BEFORE days..."

    PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST:-localhost}" -U "${DB_USER:-urlshortener}" -d "${DB_NAME:-urlshortener}" <<EOF
-- Find and display expiring keys
SELECT user_id, email, key_id, expires_at
FROM api_keys
JOIN users ON api_keys.user_id = users.id
WHERE expires_at < CURRENT_TIMESTAMP + INTERVAL '$DAYS_BEFORE days'
  AND revoked = false;

-- Note: Actual rotation logic should be implemented in the application
-- This query is for reporting purposes
EOF

    success "API key expiry check complete"
    warning "Implement actual rotation logic in your application"
    warning "Suggested: Create an API endpoint for bulk rotation"

    # Log event
    log "API key expiry check completed"
}

# Check secret expiry status
check_expiry_status() {
    log "Checking secret expiry status..."

    echo ""
    echo "====================================="
    echo "    Secret Expiry Status"
    echo "====================================="

    # Check SSL certificate expiry
    if [ ! -z "$SSL_CERT_PATH" ] && [ -f "$SSL_CERT_PATH" ]; then
        CERT_EXPIRY=$(openssl x509 -in "$SSL_CERT_PATH" -noout -enddate | cut -d= -f2)
        CERT_EXPIRY_EPOCH=$(date -d "$CERT_EXPIRY" +%s)
        CURRENT_EPOCH=$(date +%s)
        DAYS_UNTIL_EXPIRY=$(( ($CERT_EXPIRY_EPOCH - $CURRENT_EPOCH) / 86400 ))

        echo "SSL Certificate:"
        if [ $DAYS_UNTIL_EXPIRY -lt 30 ]; then
            error "  Expires in $DAYS_UNTIL_EXPIRY days (ACTION REQUIRED)"
        elif [ $DAYS_UNTIL_EXPIRY -lt 60 ]; then
            warning "  Expires in $DAYS_UNTIL_EXPIRY days"
        else
            success "  Expires in $DAYS_UNTIL_EXPIRY days"
        fi
    else
        echo "SSL Certificate: Not configured"
    fi

    # Check database password age (if logged)
    echo ""
    echo "Database Password:"
    echo "  Check your rotation log for last rotation date"
    echo "  Recommended rotation: Every 90 days"

    # Check JWT secret age (if logged)
    echo ""
    echo "JWT Secret:"
    echo "  Check your rotation log for last rotation date"
    echo "  Recommended rotation: Every 180 days"

    echo ""
    echo "====================================="
}

# Emergency rotation (all secrets)
emergency_rotation() {
    log "Starting EMERGENCY rotation of all secrets..."

    echo ""
    error "====================================="
    error "     EMERGENCY ROTATION"
    error "====================================="
    warning "This will immediately rotate ALL secrets without grace period"
    warning "This will cause service disruption until all services are updated"
    warning "Only use this if credentials have been compromised"
    echo ""
    echo -n "Type 'EMERGENCY' to confirm: "
    read confirm

    if [ "$confirm" != "EMERGENCY" ]; then
        log "Emergency rotation aborted"
        return
    fi

    log "Emergency rotation confirmed - proceeding..."

    # Rotate database password
    log "Rotating database password..."
    NEW_DB_PASSWORD=$(generate_password)
    PGPASSWORD="${POSTGRES_PASSWORD}" psql -h "${DB_HOST:-localhost}" -U postgres -d "${DB_NAME:-urlshortener}" <<EOF
ALTER USER ${DB_USER:-urlshortener} WITH PASSWORD '$NEW_DB_PASSWORD';
EOF
    sed -i.emergency "s/^DB_PASSWORD=.*/DB_PASSWORD=$NEW_DB_PASSWORD/" "${PROJECT_ROOT}/.env"
    success "Database password rotated"

    # Rotate JWT secret (no grace period)
    log "Rotating JWT secret..."
    NEW_JWT_SECRET=$(generate_secret 64)
    sed -i.emergency "s/^JWT_SECRET=.*/JWT_SECRET=$NEW_JWT_SECRET/" "${PROJECT_ROOT}/.env"
    # Remove old secret (no grace period in emergency)
    sed -i.emergency "/^JWT_SECRET_OLD=/d" "${PROJECT_ROOT}/.env"
    success "JWT secret rotated (no grace period)"

    # Rotate Redis password
    log "Rotating Redis password..."
    NEW_REDIS_PASSWORD=$(generate_password)
    redis-cli -h "${REDIS_HOST:-localhost}" -p "${REDIS_PORT:-6379}" CONFIG SET requirepass "$NEW_REDIS_PASSWORD"
    sed -i.emergency "s/^REDIS_PASSWORD=.*/REDIS_PASSWORD=$NEW_REDIS_PASSWORD/" "${PROJECT_ROOT}/.env"
    success "Redis password rotated"

    # Clear all sessions
    log "Clearing all sessions..."
    redis-cli -h "${REDIS_HOST:-localhost}" -p "${REDIS_PORT:-6379}" -a "$NEW_REDIS_PASSWORD" FLUSHDB
    success "All sessions cleared"

    success "====================================="
    success "Emergency rotation complete"
    success "====================================="
    warning "New credentials:"
    echo "  DB_PASSWORD: $NEW_DB_PASSWORD"
    echo "  JWT_SECRET: $NEW_JWT_SECRET"
    echo "  REDIS_PASSWORD: $NEW_REDIS_PASSWORD"
    warning ""
    warning "CRITICAL NEXT STEPS:"
    echo "  1. Restart ALL application instances IMMEDIATELY"
    echo "  2. Update Kubernetes secrets / environment variables"
    echo "  3. Update Redis configuration file for persistence"
    echo "  4. Force all users to re-authenticate"
    echo "  5. Investigate how credentials were compromised"
    echo "  6. Review access logs for unauthorized access"
    warning ""
    warning "Old .env backed up to .env.emergency"

    # Log emergency rotation
    log "EMERGENCY ROTATION COMPLETED - All secrets rotated"
}

# Main script
main() {
    clear
    echo "Secret Rotation Automation Tool"
    echo "==============================="
    log "Script started by $(whoami)"

    # Check environment
    check_environment

    while true; do
        show_menu
        read option

        case $option in
            1)
                rotate_database_password
                ;;
            2)
                rotate_jwt_secret
                ;;
            3)
                rotate_redis_password
                ;;
            4)
                rotate_api_keys
                ;;
            5)
                check_expiry_status
                ;;
            6)
                emergency_rotation
                ;;
            7)
                log "Script exited normally"
                echo "Goodbye!"
                exit 0
                ;;
            *)
                error "Invalid option"
                ;;
        esac

        echo ""
        echo -n "Press Enter to continue..."
        read
        clear
    done
}

# Run main function
main
