#!/usr/bin/env bash
# ==============================================================================
# SOC Trainer - Automated Deployment & Management Harness (install.sh)
# Specification Contract: docs/INSTALL_SH_SPECIFICATION.md
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper output functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Generate cryptographically secure random string (hex)
generate_random_secret() {
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex 32
    else
        head -c 32 /dev/urandom | xxd -p -c 32 2>/dev/null || cat /proc/sys/kernel/random/uuid | tr -d '-'
    fi
}

# 1. Preflight Checks
cmd_check() {
    log_info "Running Preflight and Environment Checks..."
    local errors=0

    # Check Docker
    if command -v docker >/dev/null 2>&1; then
        local docker_ver
        docker_ver=$(docker --version)
        log_success "Docker Engine: Found ($docker_ver)"
    else
        log_error "Docker Engine: Not found. Please install Docker."
        errors=$((errors + 1))
    fi

    # Check Docker Compose
    if docker compose version >/dev/null 2>&1; then
        local compose_ver
        compose_ver=$(docker compose version)
        log_success "Docker Compose Plugin: Found ($compose_ver)"
    elif command -v docker-compose >/dev/null 2>&1; then
        local compose_ver
        compose_ver=$(docker-compose --version)
        log_success "Docker Compose Standalone: Found ($compose_ver)"
    else
        log_error "Docker Compose: Not found."
        errors=$((errors + 1))
    fi

    # Check .env file
    if [ -f ".env" ]; then
        log_success "Environment File (.env): Present"
    else
        log_warn "Environment File (.env): Not found (will be created during install)"
    fi

    # Check Ports (5432, 8080, 80)
    for port in 5432 8080 80; do
        if command -v nc >/dev/null 2>&1; then
            if nc -z localhost "$port" 2>/dev/null; then
                log_warn "Port $port is currently in use on host."
            else
                log_info "Port $port: Available"
            fi
        fi
    done

    if [ "$errors" -eq 0 ]; then
        log_success "Preflight Checks Passed!"
        return 0
    else
        log_error "Preflight Checks Failed with $errors error(s)."
        return 1
    fi
}

# 2. Setup .env file
setup_env() {
    local mode="${1:-local}"
    log_info "Configuring environment for mode: $mode"

    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            cp .env.example .env
        else
            cat <<EOF > .env
POSTGRES_USER=admin
POSTGRES_PASSWORD=password
POSTGRES_DB=cyber_sim
JWT_SECRET=$(generate_random_secret)
GIN_MODE=release
DEPLOYMENT_MODE=$mode
LLM_MODE=static
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4o-mini
LLM_API_KEY=
PROFICIENCY_THRESHOLD=70.0
MAX_SCENARIOS_CAP=10
SIMULATION_TIME_LIMIT_MINS=120
EOF
        fi

        # Generate a fresh JWT Secret
        local new_jwt
        new_jwt=$(generate_random_secret)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|JWT_SECRET=.*|JWT_SECRET=$new_jwt|g" .env
            sed -i '' "s|DEPLOYMENT_MODE=.*|DEPLOYMENT_MODE=$mode|g" .env
        else
            sed -i "s|JWT_SECRET=.*|JWT_SECRET=$new_jwt|g" .env
            sed -i "s|DEPLOYMENT_MODE=.*|DEPLOYMENT_MODE=$mode|g" .env
        fi

        # Secure permissions on Unix
        chmod 600 .env 2>/dev/null || true
        log_success "Created secure .env configuration file."
    else
        log_info "Existing .env found. Preserving current configuration."
    fi
}

# 3. Wait for PostgreSQL Readiness
wait_for_db() {
    log_info "Waiting for PostgreSQL database container to be healthy..."
    local retries=30
    local count=0

    while [ "$count" -lt "$retries" ]; do
        if docker compose exec -T db pg_isready -U admin -d cyber_sim >/dev/null 2>&1; then
            log_success "Database is ready and accepting connections!"
            return 0
        fi
        sleep 2
        count=$((count + 1))
        echo -n "."
    done

    echo ""
    log_error "Timeout waiting for Database container."
    return 1
}

# 4. Safe Database Seeding (Idempotent)
cmd_seed() {
    log_info "Executing Safe Idempotent Database Seeding..."
    
    # Run parsing and loading scripts
    if [ -f "scripts/parse_exams.py" ]; then
        log_info "Parsing CompTIA Security+ examination source files..."
        python3 scripts/parse_exams.py || python scripts/parse_exams.py
    fi

    if [ -f "scripts/load_exams_to_db.py" ]; then
        log_info "Seeding Security Domains & Assessment Questions..."
        python3 scripts/load_exams_to_db.py || python scripts/load_exams_to_db.py
    fi

    if [ -f "scripts/load_scenarios_to_db.py" ]; then
        log_info "Seeding Static Scenario Pool (Safe Upsert)..."
        python3 scripts/load_scenarios_to_db.py || python scripts/load_scenarios_to_db.py
    fi

    log_success "Database seeding process completed successfully."
}

# 5. Check Seed Validation (Read-Only)
cmd_check_seed() {
    log_info "Checking seed integrity and counts in database..."
    
    docker compose exec -T db psql -U admin -d cyber_sim -c "
        SELECT 'Security Domains' as content_type, COUNT(*) as count FROM security_domains
        UNION ALL
        SELECT 'Questions', COUNT(*) FROM questions
        UNION ALL
        SELECT 'Scenarios', COUNT(*) FROM scenarios
        UNION ALL
        SELECT 'Users (Runtime)', COUNT(*) FROM users
        UNION ALL
        SELECT 'Simulation Sessions (Runtime)', COUNT(*) FROM simulation_sessions;
    " || log_warn "Could not query database directly via docker compose."
}

# 6. Database Backup
cmd_backup() {
    local backup_dir="backups"
    mkdir -p "$backup_dir"
    local timestamp
    timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_file="$backup_dir/cyber_sim_backup_${timestamp}.sql"

    log_info "Creating database backup to $backup_file..."
    docker compose exec -T db pg_dump -U admin cyber_sim > "$backup_file"
    log_success "Backup successfully created: $backup_file ($(du -h "$backup_file" | cut -f1))"
}

# 7. Reset Dev Data (Local Only with explicit confirmation)
cmd_reset_dev_data() {
    log_warn "=== DANGER ZONE: RESET DEV DATA ==="
    log_warn "This will wipe all users, sessions, and reset scenario tables."
    
    if [ -f ".env" ]; then
        # shellcheck disable=SC1091
        source .env
    fi

    if [ "${DEPLOYMENT_MODE:-local}" == "cloud" ]; then
        log_error "ABORTED: reset-dev-data is strictly prohibited in CLOUD / PRODUCTION mode."
        exit 1
    fi

    read -p "Type 'RESET' to confirm destructive data reset: " confirm
    if [ "$confirm" == "RESET" ]; then
        log_info "Resetting development database..."
        docker compose exec -T db psql -U admin -d cyber_sim -c "
            TRUNCATE TABLE session_actions, simulation_sessions, exam_session_questions, exam_sessions, user_skill_profiles, scenario_snapshots, evaluation_metrics CASCADE;
        "
        log_success "Development runtime data reset. Now re-seeding questions and scenarios..."
        cmd_seed
    else
        log_info "Operation cancelled by user."
    fi
}

# 8. Main Install Flow
cmd_install() {
    local mode="local"
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --mode)
                mode="$2"
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done

    log_info "Starting SOC Trainer Installation Pipeline (Mode: $mode)..."
    
    cmd_check
    setup_env "$mode"

    log_info "Building and starting containers via Docker Compose..."
    docker compose up -d db

    wait_for_db

    cmd_seed

    log_info "Starting backend and frontend services..."
    docker compose up -d --build backend frontend

    log_info "Running Service Health Checks..."
    sleep 5

    echo ""
    echo "========================================================"
    log_success "SOC Trainer Platform Installed Successfully!"
    echo "========================================================"
    echo "  - Web Frontend:  http://localhost"
    echo "  - REST API:      http://localhost:8080/api"
    echo "  - Database Port: 5432 (cyber_sim)"
    echo "  - Deployment:    $mode mode"
    echo "========================================================"
}

# CLI Argument Dispatcher
case "${1:-install}" in
    install)
        shift || true
        cmd_install "$@"
        ;;
    check)
        cmd_check
        ;;
    seed)
        cmd_seed
        ;;
    check-seed)
        cmd_check_seed
        ;;
    backup)
        cmd_backup
        ;;
    reset-dev-data)
        cmd_reset_dev_data
        ;;
    uninstall)
        log_warn "Stopping and removing containers..."
        docker compose down
        log_success "Containers stopped."
        ;;
    update)
        log_info "Updating application stack..."
        cmd_backup || true
        docker compose pull || true
        docker compose up -d --build
        cmd_seed
        log_success "Stack updated successfully."
        ;;
    *)
        echo "Usage: $0 {install|check|seed|check-seed|update|backup|uninstall|reset-dev-data}"
        exit 1
        ;;
esac
