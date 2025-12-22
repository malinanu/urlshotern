#!/bin/bash

# ============================================================================
# Database Migration Script
# ============================================================================
# Description: Manages database migrations for the URL Shortener project
#
# Usage:
#   ./scripts/migrate.sh status                    # Show migration status
#   ./scripts/migrate.sh up 001                    # Apply migration 001
#   ./scripts/migrate.sh down 005                  # Rollback migration 005
#   ./scripts/migrate.sh rollback-to 003           # Rollback to version 003
#   ./scripts/migrate.sh dry-run down 005          # Preview rollback without executing
#
# Environment Variables:
#   DB_HOST     - Database host (default: localhost)
#   DB_PORT     - Database port (default: 5432)
#   DB_NAME     - Database name (default: urlshortener_db)
#   DB_USER     - Database user (default: urlshortener)
#   DB_PASSWORD - Database password
#   ENVIRONMENT - Environment name (production, staging, development)
# ============================================================================

set -e  # Exit on error
set -o pipefail  # Exit on pipe failure

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-urlshortener_db}
DB_USER=${DB_USER:-urlshortener}
ENVIRONMENT=${ENVIRONMENT:-development}

MIGRATIONS_DIR="configs/migrations"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Functions

function print_usage() {
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  status                    Show which migrations have been applied"
    echo "  up <version>              Apply a specific migration (e.g., up 001)"
    echo "  down <version>            Rollback a specific migration (e.g., down 005)"
    echo "  rollback-to <version>     Rollback to a specific version (e.g., rollback-to 003)"
    echo "  dry-run <direction> <version>  Preview migration without executing"
    echo ""
    echo "Environment Variables:"
    echo "  DB_HOST      Database host (default: localhost)"
    echo "  DB_PORT      Database port (default: 5432)"
    echo "  DB_NAME      Database name (default: urlshortener_db)"
    echo "  DB_USER      Database user (default: urlshortener)"
    echo "  DB_PASSWORD  Database password"
    echo "  ENVIRONMENT  Environment (production/staging/development)"
    echo ""
    echo "Examples:"
    echo "  $0 status"
    echo "  $0 up 001"
    echo "  $0 down 005"
    echo "  $0 rollback-to 003"
    echo "  $0 dry-run down 005"
}

function log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

function log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

function log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

function log_error() {
    echo -e "${RED}✗${NC} $1"
}

function check_prerequisites() {
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        log_error "psql command not found. Please install PostgreSQL client."
        exit 1
    fi

    # Check if database is accessible
    if ! PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" &> /dev/null; then
        log_error "Cannot connect to database. Please check connection settings."
        log_info "Connection: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
        exit 1
    fi

    # Check if migrations directory exists
    if [ ! -d "$PROJECT_ROOT/$MIGRATIONS_DIR" ]; then
        log_error "Migrations directory not found: $PROJECT_ROOT/$MIGRATIONS_DIR"
        exit 1
    fi
}

function ensure_schema_migrations_table() {
    local table_exists=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations');" | tr -d '[:space:]')

    if [ "$table_exists" != "t" ]; then
        log_warning "schema_migrations table does not exist. Creating it..."
        PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$PROJECT_ROOT/$MIGRATIONS_DIR/000_schema_migrations.sql"
        log_success "schema_migrations table created"
    fi
}

function get_migration_status() {
    ensure_schema_migrations_table

    echo "Migration Status:"
    echo "=================="

    # List all migration files and their status
    for file in "$PROJECT_ROOT/$MIGRATIONS_DIR"/*_up.sql; do
        if [ -f "$file" ]; then
            filename=$(basename "$file")
            version=$(echo "$filename" | grep -oP '^\d+')

            # Check if migration is applied
            applied=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
                "SELECT applied_at FROM schema_migrations WHERE version = '$version';" | tr -d '[:space:]')

            if [ -n "$applied" ]; then
                echo -e "${GREEN}✓${NC} $version - Applied at $applied"
            else
                echo -e "${YELLOW}○${NC} $version - Not applied"
            fi
        fi
    done
}

function calculate_checksum() {
    local file=$1
    if command -v sha256sum &> /dev/null; then
        sha256sum "$file" | awk '{print $1}'
    elif command -v shasum &> /dev/null; then
        shasum -a 256 "$file" | awk '{print $1}'
    else
        echo "CHECKSUM_NOT_AVAILABLE"
    fi
}

function apply_migration() {
    local version=$1
    local migration_file="$PROJECT_ROOT/$MIGRATIONS_DIR/${version}_*_up.sql"
    local actual_file=$(ls $migration_file 2>/dev/null | head -1)

    if [ ! -f "$actual_file" ]; then
        log_error "Migration file not found: ${version}_*_up.sql"
        exit 1
    fi

    # Check if already applied
    local is_applied=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT version FROM schema_migrations WHERE version = '$version';" | tr -d '[:space:]')

    if [ -n "$is_applied" ]; then
        log_warning "Migration $version is already applied"
        return 0
    fi

    log_info "Applying migration $version..."

    # Calculate checksum
    local checksum=$(calculate_checksum "$actual_file")
    local description=$(basename "$actual_file" | sed 's/_up.sql$//' | sed 's/^[0-9]*_//')

    # Execute migration and track timing
    local start_time=$(date +%s%3N)

    if PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$actual_file"; then
        local end_time=$(date +%s%3N)
        local execution_time=$((end_time - start_time))

        # Record in schema_migrations table
        PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
            "INSERT INTO schema_migrations (version, execution_time_ms, description, checksum)
             VALUES ('$version', $execution_time, '$description', '$checksum');"

        log_success "Migration $version applied successfully (${execution_time}ms)"
    else
        log_error "Migration $version failed"
        exit 1
    fi
}

function rollback_migration() {
    local version=$1
    local migration_file="$PROJECT_ROOT/$MIGRATIONS_DIR/${version}_*_down.sql"
    local actual_file=$(ls $migration_file 2>/dev/null | head -1)

    if [ ! -f "$actual_file" ]; then
        log_error "Rollback file not found: ${version}_*_down.sql"
        exit 1
    fi

    # Check if migration is applied
    local is_applied=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT version FROM schema_migrations WHERE version = '$version';" | tr -d '[:space:]')

    if [ -z "$is_applied" ]; then
        log_warning "Migration $version is not applied, cannot rollback"
        return 0
    fi

    # Production safety check
    if [ "$ENVIRONMENT" = "production" ]; then
        log_warning "You are about to ROLLBACK a migration in PRODUCTION!"
        log_warning "This may result in DATA LOSS. Ensure you have a recent backup."
        log_info "Backup command: pg_dump -U $DB_USER -h $DB_HOST -d $DB_NAME > backup_$(date +%Y%m%d_%H%M%S).sql"
        echo ""
        read -p "Type 'YES I HAVE A BACKUP' to continue: " confirm
        if [ "$confirm" != "YES I HAVE A BACKUP" ]; then
            log_info "Rollback cancelled"
            exit 0
        fi
    fi

    log_info "Rolling back migration $version..."

    # Execute rollback
    if PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$actual_file"; then
        log_success "Migration $version rolled back successfully"
    else
        log_error "Rollback of migration $version failed"
        exit 1
    fi
}

function rollback_to_version() {
    local target_version=$1

    # Get all applied migrations
    local applied_migrations=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT version FROM schema_migrations ORDER BY version DESC;" | tr -d '[:space:]')

    local migrations_to_rollback=()

    for version in $applied_migrations; do
        # Rollback all versions greater than target
        if [ "$version" -gt "$target_version" ]; then
            migrations_to_rollback+=("$version")
        fi
    done

    if [ ${#migrations_to_rollback[@]} -eq 0 ]; then
        log_info "Already at version $target_version or lower"
        return 0
    fi

    log_info "Will rollback ${#migrations_to_rollback[@]} migration(s): ${migrations_to_rollback[*]}"

    for version in "${migrations_to_rollback[@]}"; do
        rollback_migration "$version"
    done

    log_success "Rolled back to version $target_version"
}

function dry_run() {
    local direction=$1
    local version=$2
    local migration_file="$PROJECT_ROOT/$MIGRATIONS_DIR/${version}_*_${direction}.sql"
    local actual_file=$(ls $migration_file 2>/dev/null | head -1)

    if [ ! -f "$actual_file" ]; then
        log_error "Migration file not found: ${version}_*_${direction}.sql"
        exit 1
    fi

    echo "========================================="
    echo "DRY RUN: Migration $version ($direction)"
    echo "========================================="
    echo ""
    echo "File: $actual_file"
    echo ""
    echo "SQL Content:"
    echo "-------------"
    cat "$actual_file"
    echo ""
    echo "========================================="
    echo "This was a dry run. No changes were made."
    echo "========================================="
}

# Main script

if [ $# -lt 1 ]; then
    print_usage
    exit 1
fi

COMMAND=$1

case "$COMMAND" in
    status)
        check_prerequisites
        get_migration_status
        ;;

    up)
        if [ $# -lt 2 ]; then
            log_error "Missing version number"
            print_usage
            exit 1
        fi
        VERSION=$2
        check_prerequisites
        ensure_schema_migrations_table
        apply_migration "$VERSION"
        ;;

    down)
        if [ $# -lt 2 ]; then
            log_error "Missing version number"
            print_usage
            exit 1
        fi
        VERSION=$2
        check_prerequisites
        ensure_schema_migrations_table
        rollback_migration "$VERSION"
        ;;

    rollback-to)
        if [ $# -lt 2 ]; then
            log_error "Missing version number"
            print_usage
            exit 1
        fi
        VERSION=$2
        check_prerequisites
        ensure_schema_migrations_table
        rollback_to_version "$VERSION"
        ;;

    dry-run)
        if [ $# -lt 3 ]; then
            log_error "Missing direction or version"
            echo "Usage: $0 dry-run <up|down> <version>"
            exit 1
        fi
        DIRECTION=$2
        VERSION=$3
        dry_run "$DIRECTION" "$VERSION"
        ;;

    *)
        log_error "Unknown command: $COMMAND"
        print_usage
        exit 1
        ;;
esac
