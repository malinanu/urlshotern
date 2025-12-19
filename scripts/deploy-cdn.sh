#!/bin/bash

#######################################
# CloudFlare CDN Deployment Script
# Automates CDN setup and configuration
#######################################

set -e  # Exit on error
set -u  # Exit on undefined variable

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CLOUDFLARE_CONFIG="$PROJECT_ROOT/configs/cloudflare.yaml"
WORKERS_DIR="$PROJECT_ROOT/cloudflare/workers"

# Load environment variables
if [ -f "$PROJECT_ROOT/.env" ]; then
    set -a
    source "$PROJECT_ROOT/.env"
    set +a
fi

# Required environment variables
: "${CLOUDFLARE_API_TOKEN:?CLOUDFLARE_API_TOKEN must be set}"
: "${CLOUDFLARE_ZONE_ID:?CLOUDFLARE_ZONE_ID must be set}"
: "${CLOUDFLARE_ACCOUNT_ID:?CLOUDFLARE_ACCOUNT_ID must be set}"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    log_info "Checking dependencies..."

    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed"
        exit 1
    fi

    if ! command -v wrangler &> /dev/null; then
        log_warn "wrangler not found, installing..."
        npm install -g @cloudflare/wrangler
    fi

    log_info "All dependencies satisfied"
}

configure_dns() {
    log_info "Configuring DNS records..."

    # Get origin IP
    ORIGIN_IP="${ORIGIN_IP:-$(curl -s ifconfig.me)}"

    # Create DNS record for apex domain
    curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data "{
            \"type\": \"A\",
            \"name\": \"trunc.site\",
            \"content\": \"${ORIGIN_IP}\",
            \"ttl\": 1,
            \"proxied\": true
        }" | jq '.'

    # Create DNS record for www subdomain
    curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data "{
            \"type\": \"A\",
            \"name\": \"www\",
            \"content\": \"${ORIGIN_IP}\",
            \"ttl\": 1,
            \"proxied\": true
        }" | jq '.'

    log_info "DNS records configured"
}

configure_ssl() {
    log_info "Configuring SSL/TLS..."

    curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/ssl" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{"value":"full_strict"}' | jq '.'

    curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/min_tls_version" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{"value":"1.2"}' | jq '.'

    curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/tls_1_3" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{"value":"on"}' | jq '.'

    curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/always_use_https" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{"value":"on"}' | jq '.'

    log_info "SSL/TLS configured"
}

configure_performance() {
    log_info "Configuring performance optimizations..."

    # Enable Auto Minify
    curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/minify" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{"value":{"css":"on","html":"on","js":"on"}}' | jq '.'

    # Enable Brotli
    curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/brotli" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{"value":"on"}' | jq '.'

    # Enable HTTP/2
    curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/http2" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{"value":"on"}' | jq '.'

    # Enable HTTP/3
    curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/http3" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{"value":"on"}' | jq '.'

    # Enable Early Hints
    curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/early_hints" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{"value":"on"}' | jq '.'

    log_info "Performance optimizations configured"
}

configure_caching() {
    log_info "Configuring caching rules..."

    # Create page rule for static assets
    curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/pagerules" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{
            "targets": [
                {
                    "target": "url",
                    "constraint": {
                        "operator": "matches",
                        "value": "*trunc.site/_next/static/*"
                    }
                }
            ],
            "actions": [
                {
                    "id": "cache_level",
                    "value": "cache_everything"
                },
                {
                    "id": "edge_cache_ttl",
                    "value": 31536000
                }
            ],
            "priority": 1,
            "status": "active"
        }' | jq '.'

    log_info "Caching rules configured"
}

create_kv_namespace() {
    log_info "Creating KV namespace for URL mappings..."

    cd "$WORKERS_DIR"

    # Create KV namespace
    KV_ID=$(wrangler kv:namespace create "URL_MAPPINGS" | grep -oP 'id = "\K[^"]+')

    if [ -n "$KV_ID" ]; then
        log_info "KV namespace created: $KV_ID"
        echo "URL_MAPPINGS_KV_ID=$KV_ID" >> "$PROJECT_ROOT/.env.cloudflare"
    else
        log_error "Failed to create KV namespace"
        exit 1
    fi

    cd "$PROJECT_ROOT"
}

deploy_workers() {
    log_info "Deploying CloudFlare Workers..."

    cd "$WORKERS_DIR"

    # Deploy redirect analytics worker
    log_info "Deploying redirect-analytics worker..."
    wrangler publish redirect-analytics.js --env production

    # Deploy A/B testing worker
    log_info "Deploying ab-testing worker..."
    wrangler publish ab-testing.js --env production

    cd "$PROJECT_ROOT"

    log_info "Workers deployed successfully"
}

configure_security() {
    log_info "Configuring security settings..."

    # Enable WAF
    curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/waf" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{"value":"on"}' | jq '.'

    # Configure rate limiting
    curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/rate_limits" \
        -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
        -H "Content-Type: application/json" \
        --data '{
            "match": {
                "request": {
                    "url": "*/api/*"
                }
            },
            "threshold": 100,
            "period": 600,
            "action": {
                "mode": "challenge",
                "timeout": 86400
            }
        }' | jq '.'

    log_info "Security settings configured"
}

verify_deployment() {
    log_info "Verifying deployment..."

    # Check DNS propagation
    log_info "Checking DNS..."
    dig trunc.site +short

    # Check SSL certificate
    log_info "Checking SSL..."
    curl -I https://trunc.site 2>&1 | grep -i "HTTP/"

    # Check edge cache
    log_info "Testing cache..."
    curl -I https://trunc.site 2>&1 | grep -i "cf-cache-status"

    log_info "Deployment verification complete"
}

print_summary() {
    echo ""
    echo "======================================"
    echo "  CDN Deployment Complete!"
    echo "======================================"
    echo ""
    echo "Configuration Summary:"
    echo "  Domain: trunc.site"
    echo "  CDN Provider: CloudFlare"
    echo "  SSL/TLS: Full (Strict)"
    echo "  Workers: Deployed"
    echo "  KV Namespace: Created"
    echo ""
    echo "Next Steps:"
    echo "  1. Update DNS nameservers to CloudFlare"
    echo "  2. Wait for DNS propagation (up to 48 hours)"
    echo "  3. Verify HTTPS is working"
    echo "  4. Monitor CDN analytics dashboard"
    echo ""
    echo "Useful Commands:"
    echo "  View logs: wrangler tail redirect-analytics"
    echo "  Purge cache: curl -X POST https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/purge_cache -H 'Authorization: Bearer ${CLOUDFLARE_API_TOKEN}' -d '{\"purge_everything\":true}'"
    echo ""
}

# Main execution
main() {
    log_info "Starting CDN deployment..."

    check_dependencies
    configure_dns
    configure_ssl
    configure_performance
    configure_caching
    create_kv_namespace
    deploy_workers
    configure_security
    verify_deployment
    print_summary

    log_info "CDN deployment completed successfully!"
}

# Run main function
main "$@"
