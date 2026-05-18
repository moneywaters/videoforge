#!/bin/bash
#
# Neon Database Setup Script for VideoForge
# This script provisions Neon databases for all 11 VideoForge services
#
# Usage:
#   ./scripts/neon-setup.sh [--api-key <key>] [--region <region>] [--dry-run]
#
# Requirements:
#   - NEON_API_KEY environment variable (or pass via --api-key)
#   - curl
#   - jq (for JSON parsing)
#

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="$PROJECT_ROOT/.env.neon"
LOG_FILE="$PROJECT_ROOT/logs/neon-setup.log"

# Neon API configuration
NEON_API_BASE="https://console.neon.tech/api/v2"
DEFAULT_REGION="aws-us-east-1"
DEFAULT_PG_VERSION=16

# Service databases to create (11 services)
DATABASES=(
    "videoforge_user"
    "videoforge_brief"
    "videoforge_video"
    "videoforge_campaign"
    "videoforge_shopify"
    "videoforge_performance"
    "videoforge_payout"
    "videoforge_notification"
    "videoforge_admin"
    "videoforge_ai_support"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo -e "${timestamp} [${level}] ${message}" | tee -a "$LOG_FILE"
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

# Print usage information
usage() {
    cat << EOF
Neon Database Setup Script for VideoForge

Usage:
    $0 [OPTIONS]

Options:
    --api-key <key>       Neon API key (or set NEON_API_KEY env var)
    --region <region>     AWS region (default: aws-us-east-1)
    --dry-run            Show what would be done without making changes
    -h, --help          Show this help message

Environment Variables:
    NEON_API_KEY         Neon API key (required)

Example:
    NEON_API_KEY=napi_xxx $0 --region aws-us-west-2

EOF
}

# Parse command line arguments
parse_args() {
    API_KEY="${NEON_API_KEY:-}"
    REGION="${DEFAULT_REGION}"
    DRY_RUN=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --api-key)
                API_KEY="$2"
                shift 2
                ;;
            --region)
                REGION="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN=true
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

    # Validate API key
    if [[ -z "$API_KEY" ]]; then
        log_error "NEON_API_KEY is required. Set it via environment or --api-key flag."
        usage
        exit 1
    fi
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

# Check if project exists
check_project_exists() {
    local project_name="$1"

    log_info "Checking if project '$project_name' exists..."

    local response
    response=$(neon_api_request GET "/projects" || true)

    if echo "$response" | jq -e ".projects[] | select(.name == \"$project_name\")" > /dev/null 2>&1; then
        local project_id
        project_id=$(echo "$response" | jq -r ".projects[] | select(.name == \"$project_name\") | .id")
        log_info "Found existing project: $project_id"
        echo "$project_id"
        return 0
    fi

    return 1
}

# Create Neon project
create_project() {
    local project_name="$1"

    log_info "Creating project '$project_name'..."

    local data
    data=$(jq -n \
        --arg name "$project_name" \
        --arg region "$REGION" \
        --arg pg_version "$DEFAULT_PG_VERSION" \
        '{
            project: {
                name: $name,
                region_id: $region,
                pg_version: $pg_version
            }
        }')

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY-RUN] Would create project: $data"
        echo "dry-run-project-id"
        return 0
    fi

    local response
    response=$(neon_api_request POST "/projects" "$data")

    if echo "$response" | jq -e ".project" > /dev/null 2>&1; then
        local project_id
        project_id=$(echo "$response" | jq -r ".project.id")
        log_success "Project created: $project_id"
        echo "$project_id"
        return 0
    else
        log_error "Failed to create project: $response"
        return 1
    fi
}

# Get default branch for project
get_default_branch() {
    local project_id="$1"

    log_info "Getting default branch for project $project_id..."

    local response
    response=$(neon_api_request GET "/projects/${project_id}/branches")

    if echo "$response" | jq -e ".branches[0]" > /dev/null 2>&1; then
        local branch_id
        branch_id=$(echo "$response" | jq -r ".branches[0].id")
        log_info "Default branch: $branch_id"
        echo "$branch_id"
        return 0
    else
        log_error "Failed to get branches: $response"
        return 1
    fi
}

# Check if database exists
check_database_exists() {
    local project_id="$1"
    local branch_id="$2"
    local db_name="$3"

    local response
    response=$(neon_api_request GET "/projects/${project_id}/branches/${branch_id}/databases" || true)

    if echo "$response" | jq -e ".databases[] | select(.name == \"$db_name\")" > /dev/null 2>&1; then
        return 0
    fi

    return 1
}

# Create database
create_database() {
    local project_id="$1"
    local branch_id="$2"
    local db_name="$3"

    log_info "Creating database '$db_name'..."

    local data
    data=$(jq -n \
        --arg name "$db_name" \
        '{
            database: {
                name: $name,
                owner_name: "neondb_owner"
            }
        }')

    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY-RUN] Would create database: $db_name"
        return 0
    fi

    local response
    response=$(neon_api_request POST "/projects/${project_id}/branches/${branch_id}/databases" "$data")

    if echo "$response" | jq -e ".database" > /dev/null 2>&1; then
        log_success "Database created: $db_name"
        return 0
    else
        # Database might already exist, which is fine
        if echo "$response" | jq -e ".code == \"DbAlreadyExists\"" > /dev/null 2>&1; then
            log_warn "Database already exists: $db_name"
            return 0
        fi
        log_error "Failed to create database $db_name: $response"
        return 1
    fi
}

# Get connection string for database
get_connection_string() {
    local project_id="$1"
    local branch_id="$2"
    local db_name="$3"

    log_info "Getting connection string for '$db_name'..."

    local response
    response=$(neon_api_request GET "/projects/${project_id}/branches/${branch_id}")

    if echo "$response" | jq -e ".branch" > /dev/null 2>&1; then
        # Extract connection details from the branch response
        local host
        host=$(echo "$response" | jq -r ".branch .host")
        local password
        password=$(echo "$response" | jq -r ".branch .password")

        if [[ -z "$host" || "$host" == "null" ]]; then
            log_error "Failed to get connection details"
            return 1
        fi

        # Construct connection string (Neon uses PostgreSQL URL format)
        local conn_string="postgres://neondb_owner:${password}@${host}/${db_name}?sslmode=require"
        echo "$conn_string"
        return 0
    else
        log_error "Failed to get connection string: $response"
        return 1
    fi
}

# Save connection strings to env file
save_env_file() {
    local project_id="$1"
    local branch_id="$2"

    log_info "Saving connection strings to $ENV_FILE..."

    cat > "$ENV_FILE" << EOF
# Neon Configuration for VideoForge
# Generated by scripts/neon-setup.sh on $(date '+%Y-%m-%d %H:%M:%S')
# Project ID: ${project_id}
# Branch ID: ${branch_id}

# Neon API Configuration
NEON_API_KEY=${API_KEY}
NEON_PROJECT_ID=${project_id}
NEON_BRANCH_ID=${branch_id}

# Neon Database Connection Strings
# These URLs use Neon serverless PostgreSQL with SSL required
EOF

    for db_name in "${DATABASES[@]}"; do
        local conn_string
        conn_string=$(get_connection_string "$project_id" "$branch_id" "$db_name")

        if [[ -n "$conn_string" ]]; then
            local var_name="DATABASE_URL_${db_name#videoforge_}"
            var_name=$(echo "$var_name" | tr '[:lower:]' '[:upper:]')
            echo "${var_name}=${conn_string}" >> "$ENV_FILE"
            log_info "Saved connection string for $db_name"
        fi
    done

    # Add default fallback
    cat >> "$ENV_FILE" << EOF

# Default database URL (falls back to local postgres for development)
# DATABASE_URL=postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable
EOF

    log_success "Connection strings saved to $ENV_FILE"
}

# Main setup function
setup() {
    local project_name="videoforge"
    local project_id=""
    local branch_id=""

    # Ensure logs directory exists
    mkdir -p "$PROJECT_ROOT/logs"

    log_info "=========================================="
    log_info "Neon Database Setup for VideoForge"
    log_info "=========================================="

    # Check if project exists
    if check_project_exists "$project_name"; then
        project_id=$(check_project_exists "$project_name")
        branch_id=$(get_default_branch "$project_id")
    else
        # Create new project
        if [[ "$DRY_RUN" == "true" ]]; then
            project_id="dry-run-project-id"
        else
            project_id=$(create_project "$project_name") || exit 1
            branch_id=$(get_default_branch "$project_id") || exit 1
        fi
    fi

    # Create databases for each service
    log_info "Creating databases for ${#DATABASES[@]} services..."

    for db_name in "${DATABASES[@]}"; do
        if ! check_database_exists "$project_id" "$branch_id" "$db_name"; then
            create_database "$project_id" "$branch_id" "$db_name" || log_warn "Could not create $db_name"
        else
            log_info "Database already exists: $db_name"
        fi
    done

    # Save connection strings
    if [[ "$DRY_RUN" != "true" ]]; then
        save_env_file "$project_id" "$branch_id"
    else
        log_info "[DRY-RUN] Would save connection strings to $ENV_FILE"
    fi

    log_success "=========================================="
    log_success "Neon setup complete!"
    log_success "=========================================="
    log_info "Connection strings have been saved to: $ENV_FILE"
    log_info "View your databases at: https://console.neon.tech"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_warn "This was a dry run - no actual changes were made."
    fi
}

# Run main function
main() {
    parse_args "$@"
    setup
}

main "$@"