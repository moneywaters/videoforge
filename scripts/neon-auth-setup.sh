#!/bin/bash
#
# Neon Auth Setup Script for VideoForge
# This script configures Neon Auth for serverless authentication
#
# Usage:
#   ./scripts/neon-auth-setup.sh [--api-key <key>] [--project-id <id>] [--branch-id <id>] [--providers google,github,email]
#
# Requirements:
#   - NEON_API_KEY environment variable (or pass via --api-key)
#   - NEON_PROJECT_ID (or pass via --project-id)
#   - NEON_BRANCH_ID (or pass via --branch-id)
#   - curl
#   - jq (for JSON parsing)
#
# Supported Providers:
#   - google: OAuth with Google
#   - github: OAuth with GitHub
#   - email: Email/password authentication
#

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_ROOT/.env.neon-auth"

# Neon API configuration
NEON_API_BASE="https://console.neon.tech/api/v2"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${timestamp} [${level}] ${message}"
}

log_info() {
    log "INFO" "$1"
}

log_warn() {
    log "${YELLOW}WARN${NC}" "$1"
}

log_error() {
    log "${RED}ERROR${NC}" "$1"
}

log_success() {
    log "${GREEN}SUCCESS${NC}" "$1"
}

log_step() {
    log "${CYAN}STEP${NC}" "$1"
}

# Print usage information
usage() {
    cat <<EOF
Neon Auth Setup Script for VideoForge

Usage:
    $0 [OPTIONS]

Options:
    --api-key <key>       Neon API key (or set NEON_API_KEY env var)
    --project-id <id>      Neon project ID (or set NEON_PROJECT_ID env var)
    --branch-id <id>       Neon branch ID (or set NEON_BRANCH_ID env var)
    --providers <list>     Comma-separated list of providers (default: email)
                         Supported: google,github,email
    --enable-oauth        Enable OAuth providers (requires client IDs/secrets)
    --generate-secret     Generate a cookie secret
    -h, --help           Show this help message

Environment Variables:
    NEON_API_KEY          Neon API key (required)
    NEON_PROJECT_ID       Neon project ID (required)
    NEON_BRANCH_ID       Neon branch ID (required)
    NEON_OAUTH_CLIENT_ID     Google OAuth client ID (optional)
    NEON_OAUTH_CLIENT_SECRET Google OAuth client secret (optional)
    NEON_GITHUB_CLIENT_ID   GitHub OAuth client ID (optional)
    NEON_GITHUB_CLIENT_SECRET GitHub OAuth client secret (optional)

Example:
    # Setup with default email authentication
    NEON_API_KEY=napi_xxx NEON_PROJECT_ID=proj_xxx NEON_BRANCH_ID=brn_xxx $0

    # Setup with additional OAuth providers
    NEON_API_KEY=napi_xxx $0 --project-id proj_xxx --branch-id brn_xxx --providers google,github,email

EOF
}

# Parse command line arguments
parse_args() {
    API_KEY="${NEON_API_KEY:-}"
    PROJECT_ID="${NEON_PROJECT_ID:-}"
    BRANCH_ID="${NEON_BRANCH_ID:-}"
    PROVIDERS="email"
    OAUTH_ENABLED=false
    GENERATE_SECRET=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --api-key)
                API_KEY="$2"
                shift 2
                ;;
            --project-id)
                PROJECT_ID="$2"
                shift 2
                ;;
            --branch-id)
                BRANCH_ID="$2"
                shift 2
                ;;
            --providers)
                PROVIDERS="$2"
                shift 2
                ;;
            --enable-oauth)
                OAUTH_ENABLED=true
                shift
                ;;
            --generate-secret)
                GENERATE_SECRET=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    # Validate required parameters
    if [[ -z "$API_KEY" ]]; then
        log_error "NEON_API_KEY is required. Set it via environment or --api-key flag."
        usage
        exit 1
    fi

    if [[ -z "$PROJECT_ID" ]]; then
        log_error "NEON_PROJECT_ID is required. Set it via environment or --project-id flag."
        usage
        exit 1
    fi

    if [[ -z "$BRANCH_ID" ]]; then
        log_error "NEON_BRANCH_ID is required. Set it via environment or --branch-id flag."
        usage
        exit 1
    fi

    # Export for use in other functions
    export API_KEY PROJECT_ID BRANCH_ID
}

# Generate a random cookie secret
generate_cookie_secret() {
    local secret
    secret=$(openssl rand -base64 32 2>/dev/null || head -c 32 /dev/urandom | base64)
    echo "$secret"
}

# Make API request to Neon
neon_api_request() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"

    local url="${NEON_API_BASE}${endpoint}"
    local headers=(
        -H "Authorization: Bearer ${API_KEY}"
        -H "Content-Type: application/json"
    )

    if [[ -n "$data" ]]; then
        headers+=(-d "$data")
    fi

    curl -s -X "$method" "${headers[@]}" "$url"
}

# Configure authentication providers
configure_providers() {
    local providers_json="$1"

    log_step "Configuring auth providers: $providers_json"

    local endpoint="/projects/${PROJECT_ID}/branches/${BRANCH_ID}/auth/providers"

    local data
    data=$(jq -n \
        --argjson providers "$providers_json" \
        '{
            providers: $providers
        }')

    local response
    response=$(neon_api_request POST "$endpoint" "$data")

    if echo "$response" | jq -e ".providers" > /dev/null 2>&1; then
        log_success "Providers configured successfully"
        echo "$response" | jq .
        return 0
    else
        # Some errors are acceptable (providers might already be configured)
        if echo "$response" | jq -e '.code == "ProviderAlreadyExists"' > /dev/null 2>&1; then
            log_warn "Providers already configured"
            return 0
        fi
        log_error "Failed to configure providers: $response"
        return 1
    fi
}

# Configure OAuth providers (Google, GitHub)
configure_oauth() {
    local provider="$1"
    local client_id="$2"
    local client_secret="$3"

    log_step "Configuring $provider OAuth provider..."

    local endpoint="/projects/${PROJECT_ID}/branches/${BRANCH_ID}/auth/providers/${provider}"

    local data
    data=$(jq -n \
        --arg client_id "$client_id" \
        --arg client_secret "$client_secret" \
        '{
            provider: {
                type: $provider,
                enabled: true,
                client_id: $client_id,
                client_secret: $client_secret
            }
        }')

    local response
    response=$(neon_api_request PUT "$endpoint" "$data")

    if echo "$response" | jq -e ".provider" > /dev/null 2>&1; then
        log_success "$provider OAuth provider configured"
        return 0
    else
        log_error "Failed to configure $provider: $response"
        return 1
    fi
}

# Get current auth configuration
get_auth_config() {
    log_step "Fetching current auth configuration..."

    local endpoint="/projects/${PROJECT_ID}/branches/${BRANCH_ID}/auth"

    local response
    response=$(neon_api_request GET "$endpoint")

    if echo "$response" | jq -e "." > /dev/null 2>&1; then
        log_success "Current auth configuration:"
        echo "$response" | jq .
    else
        log_error "Failed to get configuration: $response"
        return 1
    fi
}

# Save configuration to env file
save_env_file() {
    log_step "Saving configuration to $ENV_FILE..."

    local cookie_secret
    cookie_secret=$(generate_cookie_secret)

    cat > "$ENV_FILE" <<EOF
# Neon Auth Configuration for VideoForge
# Generated by scripts/neon-auth-setup.sh on $(date '+%Y-%m-%d %H:%M:%S')

# Neon Auth API Configuration
NEON_API_KEY=${API_KEY}
NEON_PROJECT_ID=${PROJECT_ID}
NEON_BRANCH_ID=${BRANCH_ID}
NEON_AUTH_COOKIE_SECRET=${cookie_secret}

# Auth Providers Enabled
# Providers: ${PROVIDERS}

# JWT Configuration
# JWT_EXPIRY=3600 (1 hour)
# REFRESH_TOKEN_EXPIRY=604800 (7 days)

# Session Configuration
# SESSION_COOKIE_NAME=neon_session
# SESSION_COOKIE_SECURE=true
# SESSION_COOKIE_SAMESITE=lax

EOF

    log_success "Configuration saved to $ENV_FILE"
    log_info "Add the following to your .env file:"
    log_info "  source $ENV_FILE"
}

# Main setup function
setup() {
    log_info "=========================================="
    log_info "Neon Auth Setup for VideoForge"
    log_info "=========================================="

    if [[ "$GENERATE_SECRET" == "true" ]]; then
        local secret
        secret=$(generate_cookie_secret)
        log_info "Generated cookie secret: $secret"
        return 0
    fi

    # Convert providers to JSON array
    local providers_json
    local IFS=','
    read -ra PROVIDER_ARRAY <<< "$PROVIDERS"
    unset IFS

    providers_json=$(printf '%s\n' "${PROVIDER_ARRAY[@]}" | jq -R . | jq -s .)

    log_info "Project ID: $PROJECT_ID"
    log_info "Branch ID: $BRANCH_ID"
    log_info "Providers: $PROVIDERS"
    echo

    # Configure providers
    configure_providers "$providers_json"

    # Configure OAuth if enabled
    if [[ "$OAUTH_ENABLED" == "true" ]]; then
        if [[ -n "${NEON_OAUTH_CLIENT_ID:-}" && -n "${NEON_OAUTH_CLIENT_SECRET:-}" ]]; then
            configure_oauth "google" "$NEON_OAUTH_CLIENT_ID" "$NEON_OAUTH_CLIENT_SECRET"
        fi
        if [[ -n "${NEON_GITHUB_CLIENT_ID:-}" && -n "${NEON_GITHUB_CLIENT_SECRET:-}" ]]; then
            configure_oauth "github" "$NEON_GITHUB_CLIENT_ID" "$NEON_GITHUB_CLIENT_SECRET"
        fi
    fi

    echo
    # Get current configuration
    get_auth_config

    echo
    # Save to env file
    save_env_file

    echo
    log_success "=========================================="
    log_success "Neon Auth setup complete!"
    log_success "=========================================="
    log_info "Configuration saved to: $ENV_FILE"
    log_info "View your auth settings at: https://console.neon.tech"
    log_info ""
    log_info "Next steps:"
    log_info "  1. Add NEON_AUTH_COOKIE_SECRET to your environment"
    log_info "  2. Update your application's auth configuration"
    log_info "  3. Restart your services"
}

# Run main function
main() {
    parse_args "$@"
    setup
}

main "$@"