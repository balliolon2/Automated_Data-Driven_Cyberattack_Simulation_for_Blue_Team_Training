#!/usr/bin/env bash
# ==============================================================================
# SOC Trainer - Linux / macOS Automated Deployment & Operations Harness (install.sh)
# Specification: workflows/cross_platform_setup.md (Issue #64)
# Bash 3.2+ (macOS stock) & Bash 4+/5+ compatible.
# ==============================================================================

set -euo pipefail

CALLER_PWD="$(pwd)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

RESET_TRUNCATE_SQL="TRUNCATE TABLE session_actions, simulation_sessions, exam_session_questions, exam_sessions, user_skill_profiles, scenario_snapshots, evaluation_metrics, user_theory_progress, user_theory_subtopic_progress, user_theory_sample_questions, comment_upvotes, thread_upvotes, thread_comments, analysis_threads, user_notifications, specialist_applications, nfr_evaluation_runs CASCADE;"

if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED='' ; GREEN='' ; YELLOW='' ; BLUE='' ; NC=''
fi

log_info()    { printf '%b[INFO]%b %s\n' "$BLUE" "$NC" "$1"; }
log_success() { printf '%b[SUCCESS]%b %s\n' "$GREEN" "$NC" "$1"; }
log_warn()    { printf '%b[WARN]%b %s\n' "$YELLOW" "$NC" "$1" >&2; }
log_error()   { printf '%b[ERROR]%b %s\n' "$RED" "$NC" "$1" >&2; }

fail_stage() {
    local stage="$1"
    local code="${2:-1}"
    local cause="${3:-Unexpected error}"
    local diag="${4:-}"
    local resume="${5:-}"
    printf '\n========================================================\n' >&2
    log_error "Failure Stage : $stage"
    log_error "Exit Code     : $code"
    log_error "Likely Cause  : $cause"
    [[ -n "$diag" ]]   && printf 'Diagnostic    : %s\n' "$diag" >&2
    [[ -n "$resume" ]] && printf 'Resume Step   : %s\n' "$resume" >&2
    printf '========================================================\n' >&2
    exit "$code"
}

SUBCOMMAND=""
RESTORE_FILE=""
LOGS_SERVICE=""
MODE=""
NON_INTERACTIVE=false
DOMAIN=""
ACME_EMAIL=""
AUTO_DEPS=false
CONFIGURE_FIREWALL=false
IMAGE_TAG=""
VERIFY_E2E=false
WITH_PGADMIN=false
SKIP_DNS_CHECK=false
NO_PULL=false
NO_UPLOADS=false
REMOVE_VOLUMES=false
DOCKER_BIN=(docker)

show_usage() {
    cat <<'EOF'
SOC Trainer - Linux / macOS Setup & Operations CLI (install.sh)

Usage:
  ./install.sh [command] [flags]

Commands (opens interactive menu when omitted on TTY):
  install            Preflight, env setup, containerized seed, startup & non-polluting smoke verify
  check              Read-only preflight checks (OS, Docker/Compose, systemd/group, CRLF, RAM/disk, ports, .env perm)
  status             Show container health, endpoints, active mode, redacted LLM/TLS state
  seed               Run idempotent containerized seed (domains, 540 questions, theory, 15 scenarios)
  check-seed         Read-only PostgreSQL count check across core tables
  update             Auto-backup -> optional git pull (--no-pull) -> rebuild/pull & re-seed -> smoke verify
  backup             Save gzipped SQL + uploads archive in backups/ (keeps 10 newest; --no-uploads)
  restore <file>     Guarded DB (+ matching uploads) restore after safety backup and RESTORE confirmation
  logs [service]     Stream Docker Compose logs
  uninstall          Stop/remove containers cleanly (preserves data volumes; --volumes needs DELETE DATA)
  reset-dev-data     Local-only wipe of learner/session runtime tables & re-seed (BLOCKED in Cloud mode)

Flags:
  --mode <local|cloud>       Deployment mode (local or cloud)
  --non-interactive          Run without interactive prompts using ENV / *_FILE / generated secrets
  --domain <hostname>        Public domain for Cloud Auto-HTTPS via Caddy (Let's Encrypt 80/443)
  --acme-email <email>       Optional ACME contact email for TLS notifications
  --auto-deps                Auto-install missing prerequisites (git, curl, Docker Engine/Compose)
  --configure-firewall       Explicitly add ufw rules (OpenSSH, 80/tcp, 443/tcp)
  --image-tag <tag>          Pull & deploy pinned versioned images instead of building from source
  --verify-e2e               Also execute scripts/test_integration.py (Local mode ONLY)
  --with-pgadmin             Start opt-in pgAdmin tools profile on 127.0.0.1:5050
  --skip-dns-check           Skip advisory DNS A-record check before Cloud HTTPS startup
  --no-pull                  Skip git pull during update
  --no-uploads               Skip uploads volume archive during backup
  --volumes                  With uninstall: delete Docker volumes after typed DELETE DATA confirmation
  --help, -h                 Show this help message
EOF
}

resolve_caller_path() {
    local raw="$1"
    [[ -n "$raw" ]] || return 0
    if [[ "$raw" == /* ]]; then
        printf '%s' "$raw"
    elif [[ -e "${CALLER_PWD}/${raw}" ]]; then
        printf '%s' "${CALLER_PWD}/${raw}"
    else
        printf '%s' "${SCRIPT_DIR}/${raw}"
    fi
}

parse_args() {
    local pos=()
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --help|-h)
                show_usage; exit 0 ;;
            --mode=*)
                MODE="${1#*=}"; shift ;;
            --mode)
                [[ $# -ge 2 ]] || fail_stage "Parse Arguments" 2 "Missing value for --mode" "./install.sh --help" ""
                MODE="$2"; shift 2 ;;
            --non-interactive)
                NON_INTERACTIVE=true; shift ;;
            --domain=*)
                DOMAIN="${1#*=}"; shift ;;
            --domain)
                [[ $# -ge 2 ]] || fail_stage "Parse Arguments" 2 "Missing value for --domain" "./install.sh --help" ""
                DOMAIN="$2"; shift 2 ;;
            --acme-email=*)
                ACME_EMAIL="${1#*=}"; shift ;;
            --acme-email)
                [[ $# -ge 2 ]] || fail_stage "Parse Arguments" 2 "Missing value for --acme-email" "./install.sh --help" ""
                ACME_EMAIL="$2"; shift 2 ;;
            --auto-deps)
                AUTO_DEPS=true; shift ;;
            --configure-firewall)
                CONFIGURE_FIREWALL=true; shift ;;
            --image-tag=*)
                IMAGE_TAG="${1#*=}"; shift ;;
            --image-tag)
                [[ $# -ge 2 ]] || fail_stage "Parse Arguments" 2 "Missing value for --image-tag" "./install.sh --help" ""
                IMAGE_TAG="$2"; shift 2 ;;
            --verify-e2e)
                VERIFY_E2E=true; shift ;;
            --with-pgadmin)
                WITH_PGADMIN=true; shift ;;
            --skip-dns-check)
                SKIP_DNS_CHECK=true; shift ;;
            --no-pull)
                NO_PULL=true; shift ;;
            --no-uploads)
                NO_UPLOADS=true; shift ;;
            --volumes)
                REMOVE_VOLUMES=true; shift ;;
            -*)
                fail_stage "Parse Arguments" 2 "Unknown flag: $1" "./install.sh --help" "Run './install.sh --help' for valid options." ;;
            *)
                pos+=("$1"); shift ;;
        esac
    done

    if [[ ${#pos[@]} -gt 0 ]]; then
        SUBCOMMAND="${pos[0]}"
        if [[ ${#pos[@]} -gt 1 ]]; then
            [[ "$SUBCOMMAND" == "restore" ]] && RESTORE_FILE="${pos[1]}"
            [[ "$SUBCOMMAND" == "logs" ]]    && LOGS_SERVICE="${pos[1]}"
        fi
    fi
}

run_root() {
    if [[ "$(id -u)" -eq 0 ]]; then
        "$@"
    elif command -v sudo >/dev/null 2>&1; then
        sudo "$@"
    else
        return 1
    fi
}

prompt_line() {
    local msg="$1"
    local ans=""
    if [[ -r /dev/tty ]]; then
        printf '%s' "$msg" > /dev/tty
        read -r ans < /dev/tty || ans=""
    else
        printf '%s' "$msg"
        read -r ans || ans=""
    fi
    printf '%s' "$ans"
}

prompt_secret() {
    local msg="$1"
    local val=""
    if [[ -r /dev/tty ]]; then
        printf '%s' "$msg" > /dev/tty
        read -rs val < /dev/tty || val=""
        printf '\n' > /dev/tty
    else
        printf '%s' "$msg"
        read -rs val || val=""
        printf '\n'
    fi
    printf '%s' "$val"
}

generate_hex_secret() {
    local bytes="${1:-32}"
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex "$bytes"
    elif [[ -r /dev/urandom ]]; then
        od -An -tx1 -N "$bytes" /dev/urandom | tr -d ' \n'
    else
        fail_stage "Secret Generation" 1 "Neither openssl nor /dev/urandom available for CSPRNG secret generation." "" ""
    fi
}

generate_safe_password() {
    local len="${1:-24}"
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -base64 "$((len * 2))" | tr -dc 'A-Za-z0-9_-' | head -c "$len"
    else
        od -An -tx1 -N "$len" /dev/urandom | tr -dc 'A-Za-z0-9' | head -c "$len"
    fi
}

json_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\n'/\\n}"
    s="${s//$'\r'/\\r}"
    s="${s//$'\t'/\\t}"
    printf '%s' "$s"
}

read_env_value() {
    local key="$1"
    local env_file="${SCRIPT_DIR}/.env"
    [[ -f "$env_file" ]] || return 0
    local line
    line="$(grep -E "^[[:space:]]*${key}=" "$env_file" | tail -n 1 || true)"
    [[ -n "$line" ]] || return 0
    local val="${line#*=}"
    val="${val%$'\r'}"
    if [[ ( "$val" == \"*\" && "$val" == *\" ) || ( "$val" == \'*\' && "$val" == *\' ) ]]; then
        val="${val:1:${#val}-2}"
    fi
    printf '%s' "$val"
}

get_pg_config() {
    CFG_PG_USER="$(read_env_value "POSTGRES_USER")"; [[ -z "$CFG_PG_USER" ]] && CFG_PG_USER="admin"
    CFG_PG_DB="$(read_env_value "POSTGRES_DB")"; [[ -z "$CFG_PG_DB" ]] && CFG_PG_DB="cyber_sim"
    if [[ ! "$CFG_PG_USER" =~ ^[A-Za-z0-9_]+$ || ! "$CFG_PG_DB" =~ ^[A-Za-z0-9_]+$ ]]; then
        fail_stage "Database Config" 2 "POSTGRES_USER/POSTGRES_DB contain invalid identifier characters." "" "Use alphanumeric/underscore database identifiers."
    fi
}

resolve_secret_setting() {
    local key="$1"
    local env_val="${!key-}"
    if [[ -n "$env_val" ]]; then
        printf '%s' "$env_val"
        return 0
    fi

    local file_var="${key}_FILE"
    local file_raw="${!file_var-}"
    if [[ -n "$file_raw" ]]; then
        local file_path
        file_path="$(resolve_caller_path "$file_raw")"
        [[ -f "$file_path" && -r "$file_path" ]] || fail_stage "Environment Setup" 1 "${file_var} points to unreadable file: $file_path" "ls -l '$file_path'" "Ensure secret file exists and is readable."
        local fval
        fval="$(tr -d '\r\n' < "$file_path")"
        [[ -n "$fval" ]] || fail_stage "Environment Setup" 1 "${file_var} is empty: $file_path" "" "Write a non-empty secret value to $file_path."
        printf '%s' "$fval"
        return 0
    fi

    local existing_val
    existing_val="$(read_env_value "$key")"
    printf '%s' "$existing_val"
}

resolve_existing_or_default() {
    local key="$1"
    local def_val="$2"
    local env_val="${!key-}"
    if [[ -n "$env_val" ]]; then
        printf '%s' "$env_val"
        return 0
    fi
    local ex
    ex="$(read_env_value "$key")"
    if [[ -n "$ex" ]]; then
        printf '%s' "$ex"
    else
        printf '%s' "$def_val"
    fi
}

# Pure POSIX / Bash 3.2+ env merger with pre-locked temp file & cleanup trap (zero declare -A)
write_merged_env() {
    local env_file="${SCRIPT_DIR}/.env"
    local tpl_file="${SCRIPT_DIR}/.env.example"
    local old_umask
    old_umask="$(umask)"
    umask 077
    local tmp_kv tmp_out
    tmp_kv="$(mktemp "${SCRIPT_DIR}/.env.tmp.kv.XXXXXX")"
    tmp_out="$(mktemp "${SCRIPT_DIR}/.env.tmp.out.XXXXXX")"
    umask "$old_umask"
    trap 'rm -f "${tmp_kv:-}" "${tmp_out:-}"' RETURN

    chmod 600 "$tmp_kv" "$tmp_out"
    cat > "$tmp_kv"

    local src_file=""
    if [[ -f "$env_file" ]]; then
        src_file="$env_file"
    elif [[ -f "$tpl_file" ]]; then
        src_file="$tpl_file"
    fi

    local seen_keys="|"
    if [[ -n "$src_file" ]]; then
        while IFS= read -r raw_line || [[ -n "$raw_line" ]]; do
            raw_line="${raw_line%$'\r'}"
            if [[ "$raw_line" =~ ^[[:space:]]*$ || "$raw_line" =~ ^[[:space:]]*# || "$raw_line" != *=* ]]; then
                printf '%s\n' "$raw_line" >> "$tmp_out"
                continue
            fi
            local lk="${raw_line%%=*}"
            lk="${lk//[[:space:]]/}"
            local match_line=""
            match_line="$(grep -E "^${lk}=" "$tmp_kv" | tail -n 1 || true)"
            if [[ -n "$match_line" ]]; then
                printf '%s\n' "$match_line" >> "$tmp_out"
                seen_keys="${seen_keys}${lk}|"
            else
                printf '%s\n' "$raw_line" >> "$tmp_out"
            fi
        done < "$src_file"
    fi

    while IFS= read -r kv_line || [[ -n "$kv_line" ]]; do
        [[ -n "$kv_line" ]] || continue
        local k="${kv_line%%=*}"
        case "$seen_keys" in
            *"|${k}|"*) ;;
            *) printf '%s\n' "$kv_line" >> "$tmp_out" ;;
        esac
    done < "$tmp_kv"

    chmod 600 "$tmp_out"
    mv "$tmp_out" "$env_file"
    chmod 600 "$env_file"
    rm -f "$tmp_kv"
    log_success "Saved and locked .env permissions (chmod 600)."
}

configure_docker_cmd() {
    DOCKER_BIN=(docker)
    if ! command -v docker >/dev/null 2>&1; then
        return 1
    fi
    if docker info >/dev/null 2>&1; then
        return 0
    fi
    if [[ "$(id -u)" -ne 0 ]] && command -v sudo >/dev/null 2>&1 && sudo docker info >/dev/null 2>&1; then
        DOCKER_BIN=(sudo docker)
        return 0
    fi
    return 1
}

ensure_compose_env() {
    export COMPOSE_PATH_SEPARATOR=":"
    local cf
    cf="$(read_env_value "COMPOSE_FILE")"
    if [[ -n "$cf" ]]; then
        export COMPOSE_FILE="$cf"
    else
        unset COMPOSE_FILE || true
    fi
    if [[ -n "$IMAGE_TAG" ]]; then
        export IMAGE_TAG="$IMAGE_TAG"
    else
        unset IMAGE_TAG || true
    fi
}

run_compose() {
    ensure_compose_env
    local args=(compose)
    if [[ "$WITH_PGADMIN" == true ]]; then
        args+=(--profile tools)
    fi
    if [[ "${DOCKER_BIN[0]}" == "sudo" ]]; then
        local env_pass=("COMPOSE_PATH_SEPARATOR=${COMPOSE_PATH_SEPARATOR}")
        [[ -n "${COMPOSE_FILE:-}" ]] && env_pass+=("COMPOSE_FILE=${COMPOSE_FILE}")
        [[ -n "${IMAGE_TAG:-}" ]]    && env_pass+=("IMAGE_TAG=${IMAGE_TAG}")
        sudo env "${env_pass[@]}" docker "${args[@]}" "$@"
    else
        docker "${args[@]}" "$@"
    fi
}

install_docker_linux() {
    local os_name
    os_name="$(uname -s)"
    if [[ "$os_name" == "Darwin" ]]; then
        fail_stage "Preflight (Docker Desktop)" 1 "Docker Desktop is required on macOS." "open -a Docker" "Install Docker Desktop for Mac, launch it until ready, and re-run ./install.sh."
    fi

    log_info "Installing Docker Engine and Docker Compose plugin..."
    if command -v apt-get >/dev/null 2>&1; then
        if command -v curl >/dev/null 2>&1; then
            curl -fsSL https://get.docker.com | run_root sh
        else
            run_root apt-get update -y
            run_root apt-get install -y docker.io docker-compose-plugin || run_root apt-get install -y docker.io docker-compose-v2
        fi
    elif command -v dnf >/dev/null 2>&1; then
        if grep -qi "amazon linux" /etc/os-release 2>/dev/null; then
            run_root dnf install -y docker docker-compose-plugin || run_root dnf install -y docker
        elif command -v curl >/dev/null 2>&1; then
            curl -fsSL https://get.docker.com | run_root sh || run_root dnf install -y docker docker-compose-plugin
        else
            run_root dnf install -y docker docker-compose-plugin
        fi
    elif command -v yum >/dev/null 2>&1; then
        run_root yum install -y docker || curl -fsSL https://get.docker.com | run_root sh
    else
        fail_stage "Docker Install" 1 "Unsupported Linux package manager for automated Docker install." "uname -a; cat /etc/os-release" "Install Docker Engine + Compose v2 manually and re-run."
    fi

    if command -v systemctl >/dev/null 2>&1 && [[ -d /run/systemd/system ]]; then
        run_root systemctl enable --now docker || true
    fi
    if [[ "$(id -u)" -ne 0 ]] && getent group docker >/dev/null 2>&1; then
        run_root usermod -aG docker "$(whoami)" || true
        log_warn "Added $(whoami) to 'docker' group (applies cleanly on next shell; using sudo socket in this run if needed)."
    fi
}

check_port_in_use() {
    local port="$1"
    if command -v ss >/dev/null 2>&1; then
        ss -ltn "sport = :$port" 2>/dev/null | grep -q ":$port"
    elif command -v lsof >/dev/null 2>&1; then
        lsof -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
    elif command -v nc >/dev/null 2>&1; then
        nc -z 127.0.0.1 "$port" >/dev/null 2>&1
    else
        return 1
    fi
}

offer_low_ram_swap() {
    [[ "$(uname -s)" == "Linux" ]] || return 0
    [[ -r /proc/meminfo ]] || return 0
    local mem_kb swap_kb mem_mb swap_mb
    mem_kb="$(awk '/^MemTotal:/ {print $2}' /proc/meminfo)"
    swap_kb="$(awk '/^SwapTotal:/ {print $2}' /proc/meminfo)"
    mem_mb=$((mem_kb / 1024))
    swap_mb=$((swap_kb / 1024))
    log_info "Memory: ${mem_mb} MB RAM | ${swap_mb} MB Swap"

    if [[ "$mem_mb" -lt 2048 && "$swap_mb" -eq 0 ]]; then
        log_warn "Host has < 2 GB RAM (${mem_mb} MB) and 0 swap. Source builds on Cloud VMs may hit OOM."
        local create_swap="$AUTO_DEPS"
        if [[ "$create_swap" != true && "$NON_INTERACTIVE" != true ]]; then
            local ans
            ans="$(prompt_line "Create a persistent 2 GB swapfile (/swapfile) now? [Y/n]: ")"
            [[ -z "$ans" || "$ans" =~ ^[Yy] ]] && create_swap=true
        fi
        if [[ "$create_swap" == true ]]; then
            if [[ -f /swapfile ]]; then
                log_warn "/swapfile already exists; activating swap..."
                run_root swapon /swapfile || true
                return 0
            fi
            run_root fallocate -l 2G /swapfile 2>/dev/null || run_root dd if=/dev/zero of=/swapfile bs=1M count=2048 status=none
            run_root chmod 600 /swapfile
            run_root mkswap /swapfile
            run_root swapon /swapfile
            if ! grep -q '^/swapfile ' /etc/fstab 2>/dev/null; then
                printf '/swapfile none swap sw 0 0\n' | run_root tee -a /etc/fstab >/dev/null
            fi
            log_success "Configured persistent 2 GB swapfile."
        fi
    fi
}

audit_env_permissions() {
    local env_file="${SCRIPT_DIR}/.env"
    local mode_str=""
    if stat -c '%a' "$env_file" >/dev/null 2>&1; then
        mode_str="$(stat -c '%a' "$env_file")"
    elif stat -f '%Lp' "$env_file" >/dev/null 2>&1; then
        mode_str="$(stat -f '%Lp' "$env_file")"
    fi
    if [[ "$mode_str" == "600" || "$mode_str" == "400" ]]; then
        log_success ".env Permissions: ${mode_str} (owner-only)"
    elif [[ -n "$mode_str" ]]; then
        log_warn ".env Permissions are ${mode_str}; tightening with chmod 600 recommended."
    fi
}

cmd_check() {
    local allow_auto="${1:-false}"
    log_info "Running Linux / macOS Preflight Checks..."
    local errors=0

    local os_s arch
    os_s="$(uname -s)"
    arch="$(uname -m)"
    if [[ -f /etc/os-release ]]; then
        local distro
        distro="$(awk -F= '/^PRETTY_NAME=/ {gsub(/"/,"",$2); print $2}' /etc/os-release)"
        log_info "OS: ${distro:-$os_s} ($arch)"
    else
        log_info "OS: $os_s ($arch)"
    fi

    for shf in "$SCRIPT_DIR/install.sh" "$SCRIPT_DIR/scripts/bootstrap.sh"; do
        if [[ -f "$shf" ]] && grep -q $'\r' "$shf" 2>/dev/null; then
            log_error "Shell script contains CRLF line endings: $shf"
            errors=$((errors + 1))
        fi
    done

    for tool in git curl; do
        if command -v "$tool" >/dev/null 2>&1; then
            log_success "Prerequisite '$tool': Found"
        else
            log_warn "Prerequisite '$tool': Missing on PATH"
        fi
    done

    if [[ "$os_s" == "Linux" ]]; then
        if command -v systemctl >/dev/null 2>&1 && [[ -d /run/systemd/system ]]; then
            local d_state
            d_state="$(systemctl is-active docker 2>/dev/null || echo inactive)"
            log_info "systemd docker unit: $d_state"
        fi
        if [[ "$(id -u)" -ne 0 ]] && command -v id >/dev/null 2>&1; then
            if id -nG 2>/dev/null | tr ' ' '\n' | grep -qx "docker"; then
                log_success "Docker Group Membership: Present"
            else
                log_info "Docker Group Membership: User not in 'docker' group (sudo fallback used if permitted)"
            fi
        fi
    fi

    local avail_kb avail_gb
    avail_kb="$(df -Pk "$SCRIPT_DIR" | awk 'NR==2 {print $4}')"
    avail_gb="$((avail_kb / 1024 / 1024))"
    log_info "Available Disk Space: ~${avail_gb} GB"
    [[ "$avail_gb" -lt 4 ]] && log_warn "Low free disk (< 4 GB); container images/builds need sufficient storage."

    if [[ -r /proc/meminfo ]]; then
        local mem_mb
        mem_mb="$(awk '/^MemTotal:/ {printf "%d", $2/1024}' /proc/meminfo)"
        log_info "System RAM: ${mem_mb} MB"
    fi

    if ! command -v docker >/dev/null 2>&1; then
        if [[ "$allow_auto" == true ]]; then
            local do_docker="$AUTO_DEPS"
            if [[ "$do_docker" != true && "$NON_INTERACTIVE" != true ]]; then
                local ans
                ans="$(prompt_line "Docker Engine is not installed. Install Docker now? [Y/n]: ")"
                [[ -z "$ans" || "$ans" =~ ^[Yy] ]] && do_docker=true
            fi
            if [[ "$do_docker" == true ]]; then
                install_docker_linux
            else
                log_error "Docker Engine not installed."
                errors=$((errors + 1))
            fi
        else
            log_error "Docker Engine not found."
            errors=$((errors + 1))
        fi
    fi

    if command -v docker >/dev/null 2>&1; then
        log_success "Docker CLI: $(docker --version)"
        if configure_docker_cmd; then
            log_success "Docker Daemon: Running"
            if run_compose version >/dev/null 2>&1; then
                log_success "Docker Compose: $(run_compose version)"
            else
                log_error "Docker Compose plugin v2 missing."
                errors=$((errors + 1))
            fi
        else
            log_error "Docker daemon is not reachable (start Docker service/Desktop or check sudo/docker group)."
            errors=$((errors + 1))
        fi
    fi

    if [[ -d "${SCRIPT_DIR}/.env" ]]; then
        log_error ".env is a directory rather than a regular file!"
        errors=$((errors + 1))
    elif [[ -f "${SCRIPT_DIR}/.env" ]]; then
        log_success "Environment File (.env): Present"
        audit_env_permissions
    else
        log_info "Environment File (.env): Will be created during install"
    fi

    for req in docker-compose.yml docker-compose.cloud.yml Caddyfile.template .env.example init.sql; do
        if [[ ! -f "${SCRIPT_DIR}/${req}" ]]; then
            log_error "Missing required repository file: $req"
            errors=$((errors + 1))
        fi
    done

    for p in 80 443 8080 5432; do
        if check_port_in_use "$p"; then
            log_warn "Port $p is currently in use on host (normal if stack is already running)."
        else
            log_info "Port $p : Available"
        fi
    done

    if [[ "$errors" -gt 0 ]]; then
        fail_stage "Preflight Checks" 1 "$errors preflight check(s) failed." "docker info; docker compose version" "Resolve reported issues and re-run ./install.sh install."
    fi
    log_success "All Preflight Checks Passed!"
}

render_caddyfile() {
    local domain_name="$1"
    local acme_email="$2"
    local tpl_file="${SCRIPT_DIR}/Caddyfile.template"
    [[ -f "$tpl_file" ]] || fail_stage "Cloud HTTPS Config" 1 "Missing Caddyfile.template" "" ""

    local out_file="${SCRIPT_DIR}/Caddyfile"
    : > "$out_file"
    if [[ -n "$acme_email" ]]; then
        printf '{\n    email %s\n}\n\n' "$acme_email" > "$out_file"
    fi
    local tpl_body
    tpl_body="$(cat "$tpl_file")"
    tpl_body="${tpl_body//__ACME_EMAIL_BLOCK__/}"
    tpl_body="${tpl_body//__DOMAIN__/$domain_name}"
    printf '%s\n' "${tpl_body%$'\n'}" >> "$out_file"
    log_success "Rendered Caddyfile for $domain_name (Auto-HTTPS)."
}

check_cloud_dns() {
    local domain_name="$1"
    [[ "$SKIP_DNS_CHECK" == true ]] && { log_info "Skipping DNS A-record check (--skip-dns-check)."; return 0; }
    log_info "Checking DNS A-record for $domain_name..."
    local resolved=""
    if command -v getent >/dev/null 2>&1; then
        resolved="$(getent ahostsv4 "$domain_name" 2>/dev/null | awk '{print $1}' | sort -u | tr '\n' ' ' || true)"
    elif command -v dig >/dev/null 2>&1; then
        resolved="$(dig +short A "$domain_name" 2>/dev/null | tr '\n' ' ' || true)"
    fi

    if [[ -z "$resolved" ]]; then
        log_warn "Domain '$domain_name' has no resolved IPv4 A-record yet. Let's Encrypt requires public DNS to match this VM."
        if [[ "$NON_INTERACTIVE" != true ]]; then
            local ans
            ans="$(prompt_line "Continue anyway? [y/N]: ")"
            [[ "$ans" =~ ^[Yy] ]] || fail_stage "DNS Verification" 1 "Paused waiting for DNS A-record propagation for $domain_name." "dig +short A $domain_name" "Point DNS to this VM's public IP and re-run."
        fi
        return 0
    fi

    local pub_ip=""
    if command -v curl >/dev/null 2>&1; then
        pub_ip="$(curl -fsS --max-time 4 https://api.ipify.org 2>/dev/null || true)"
    fi
    if [[ -n "$pub_ip" && "$resolved" != *"$pub_ip"* ]]; then
        log_warn "DNS for $domain_name resolves to [$resolved], while host public IP is $pub_ip."
        if [[ "$NON_INTERACTIVE" != true ]]; then
            local ans
            ans="$(prompt_line "Continue with Caddy TLS setup anyway? [y/N]: ")"
            [[ "$ans" =~ ^[Yy] ]] || fail_stage "DNS Verification" 1 "DNS A-record mismatch for $domain_name." "dig +short A $domain_name" "Update DNS or rerun with --skip-dns-check."
        fi
    else
        log_success "DNS resolution OK ($domain_name -> $resolved)."
    fi
}

audit_cloud_firewall() {
    local is_cloud="$1"
    local allow_tls="$2"
    printf '\n'
    log_info "Firewall & Network Exposure Check:"
    if [[ "$is_cloud" == true ]]; then
        printf '  Required Cloud Inbound Rules : TCP 80 (HTTP), TCP 443 (HTTPS when domain TLS active)\n'
        printf '  Must Stay Private / Loopback : TCP 5432 (PostgreSQL), TCP 8080 (Go API), TCP 5050 (pgAdmin)\n'
    else
        printf '  Local Stack Ports            : Web http://localhost:80 | API http://127.0.0.1:8080\n'
    fi

    if [[ "$CONFIGURE_FIREWALL" != true && "$is_cloud" == true && "$NON_INTERACTIVE" != true ]]; then
        local ans
        ans="$(prompt_line "Apply ufw firewall rules (OpenSSH + 80/443) now? [y/N]: ")"
        [[ "$ans" =~ ^[Yy] ]] && CONFIGURE_FIREWALL=true
    fi

    if [[ "$CONFIGURE_FIREWALL" == true ]]; then
        if command -v ufw >/dev/null 2>&1; then
            run_root ufw allow OpenSSH || run_root ufw allow 22/tcp
            run_root ufw allow 80/tcp
            [[ "$allow_tls" == true ]] && run_root ufw allow 443/tcp
            run_root ufw --force enable || true
            log_success "Configured ufw rules (OpenSSH, HTTP/HTTPS)."
        else
            log_warn "ufw not detected; configure your Cloud Security Group / host firewall for 80/443 manually."
        fi
    else
        log_info "Host firewall untouched (pass --configure-firewall to apply ufw rules explicitly)."
    fi
}

setup_env_config() {
    local existing_mode
    existing_mode="$(read_env_value "APP_ENVIRONMENT")"

    if [[ -z "$MODE" ]]; then
        if [[ "$existing_mode" == "local" || "$existing_mode" == "cloud" ]]; then
            MODE="$existing_mode"
        elif [[ "$NON_INTERACTIVE" == true ]]; then
            MODE="local"
        else
            printf '\nSelect Deployment Mode:\n  1) local  (Workstation / localhost HTTP :80) [Default]\n  2) cloud  (Public Cloud VM with Domain Auto-HTTPS or Public IP)\n'
            local mpick
            mpick="$(prompt_line "Choice [1/2]: ")"
            [[ "$mpick" == "2" || "$mpick" == "cloud" ]] && MODE="cloud" || MODE="local"
        fi
    fi

    if [[ "$MODE" != "local" && "$MODE" != "cloud" ]]; then
        fail_stage "Environment Setup" 2 "Invalid mode '$MODE'. Expected 'local' or 'cloud'." "./install.sh --help" ""
    fi

    if [[ "$VERIFY_E2E" == true && "$MODE" == "cloud" ]]; then
        fail_stage "Safety Guard" 1 "--verify-e2e is strictly prohibited in Cloud mode to protect research data." "" "Run without --verify-e2e."
    fi

    local existing_domain domain_val acme_val frontend_url compose_file
    existing_domain="$(read_env_value "DOMAIN")"
    domain_val="$DOMAIN"
    [[ -z "$domain_val" && "$MODE" == "cloud" ]] && domain_val="$existing_domain"
    acme_val="$ACME_EMAIL"
    [[ -z "$acme_val" ]] && acme_val="$(read_env_value "ACME_EMAIL")"
    frontend_url="$(read_env_value "FRONTEND_URL")"
    [[ -z "$frontend_url" ]] && frontend_url="http://localhost"
    compose_file="docker-compose.yml"
    TLS_ACTIVE=false

    if [[ "$MODE" == "cloud" ]]; then
        offer_low_ram_swap
        if [[ -z "$domain_val" && "$NON_INTERACTIVE" != true ]]; then
            domain_val="$(prompt_line "Cloud Domain Name for Auto-HTTPS (leave blank for IP-only HTTP :80): ")"
        fi
        if [[ -n "$domain_val" ]]; then
            if [[ "$domain_val" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ || "$domain_val" == *:* ]]; then
                fail_stage "Cloud Domain Validation" 2 "DOMAIN '$domain_val' looks like an IP address. Omit --domain for IP-only mode." "./install.sh install --mode cloud" ""
            fi
            if [[ -z "$acme_val" && "$NON_INTERACTIVE" != true ]]; then
                acme_val="$(prompt_line "ACME email for Let's Encrypt expiry notifications (optional, Enter to skip): ")"
            fi
            check_cloud_dns "$domain_val"
            render_caddyfile "$domain_val" "$acme_val"
            frontend_url="https://${domain_val}"
            compose_file="docker-compose.yml:docker-compose.cloud.yml"
            TLS_ACTIVE=true
        else
            log_warn "Cloud mode without Domain: exposing HTTP :80 without TLS. Add --domain later for HTTPS."
            if [[ "$NON_INTERACTIVE" != true && "$frontend_url" == "http://localhost" ]]; then
                local pub_ip=""
                command -v curl >/dev/null 2>&1 && pub_ip="$(curl -fsS --max-time 3 https://api.ipify.org 2>/dev/null || true)"
                local sug="${pub_ip:+http://$pub_ip}"
                [[ -z "$sug" ]] && sug="http://localhost"
                local ip_in
                ip_in="$(prompt_line "Public IP / Origin for CORS FRONTEND_URL [$sug]: ")"
                frontend_url="${ip_in:-$sug}"
            fi
        fi
    else
        if [[ -n "$existing_domain" && "$frontend_url" == "https://${existing_domain}" ]]; then
            frontend_url="http://localhost"
        elif [[ "$frontend_url" == https://* ]]; then
            frontend_url="http://localhost"
        fi
        domain_val=""
        acme_val=""
        compose_file="docker-compose.yml"
    fi

    local pg_user pg_db pg_pass jwt_secret admin_email admin_nick admin_pass generated_admin llm_key llm_url llm_model
    pg_user="$(resolve_secret_setting "POSTGRES_USER")"; [[ -z "$pg_user" ]] && pg_user="admin"
    pg_db="$(resolve_secret_setting "POSTGRES_DB")"; [[ -z "$pg_db" ]] && pg_db="cyber_sim"
    if [[ ! "$pg_user" =~ ^[A-Za-z0-9_]+$ || ! "$pg_db" =~ ^[A-Za-z0-9_]+$ ]]; then
        fail_stage "Environment Setup" 2 "POSTGRES_USER/POSTGRES_DB must match [A-Za-z0-9_]+." "" ""
    fi

    pg_pass="$(resolve_secret_setting "POSTGRES_PASSWORD")"
    if [[ -z "$pg_pass" ]]; then
        if [[ "$NON_INTERACTIVE" != true ]]; then
            pg_pass="$(prompt_secret "PostgreSQL Password (Enter to auto-generate strong secret): ")"
        fi
        [[ -z "$pg_pass" ]] && pg_pass="$(generate_safe_password 24)"
    fi

    jwt_secret="$(resolve_secret_setting "JWT_SECRET")"
    [[ -z "$jwt_secret" ]] && jwt_secret="$(generate_hex_secret 32)"

    admin_email="$(resolve_secret_setting "ADMIN_EMAIL")"
    if [[ -z "$admin_email" ]]; then
        if [[ "$NON_INTERACTIVE" != true ]]; then
            admin_email="$(prompt_line "Initial Admin Email [admin@soc.local]: ")"
        fi
        [[ -z "$admin_email" ]] && admin_email="admin@soc.local"
    fi

    admin_nick="$(resolve_secret_setting "ADMIN_NICKNAME")"; [[ -z "$admin_nick" ]] && admin_nick="admin"

    admin_pass="$(resolve_secret_setting "ADMIN_PASSWORD")"
    generated_admin=""
    if [[ -z "$admin_pass" ]]; then
        if [[ "$NON_INTERACTIVE" != true ]]; then
            admin_pass="$(prompt_secret "Initial Admin Password (Enter to auto-generate strong password): ")"
        fi
        if [[ -z "$admin_pass" ]]; then
            admin_pass="$(generate_safe_password 18)"
            generated_admin="$admin_pass"
        fi
    fi

    llm_key="$(resolve_secret_setting "LLM_API_KEY")"
    llm_url="$(resolve_existing_or_default "LLM_BASE_URL" "https://api.openai.com/v1")"
    llm_model="$(resolve_existing_or_default "LLM_MODEL" "gpt-4o-mini")"

    if [[ "$NON_INTERACTIVE" != true && -z "$llm_key" ]]; then
        local use_llm
        use_llm="$(prompt_line "Configure Live OpenAI-compatible LLM API Key? [y/N - Enter uses Static Fallback]: ")"
        if [[ "$use_llm" =~ ^[Yy] ]]; then
            local in_u in_m
            in_u="$(prompt_line "LLM Base URL [$llm_url]: ")"; [[ -n "$in_u" ]] && llm_url="$in_u"
            in_m="$(prompt_line "LLM Model [$llm_model]: ")"; [[ -n "$in_m" ]] && llm_model="$in_m"
            llm_key="$(prompt_secret "LLM API Key (hidden input): ")"
        fi
    fi

    local gin_mode swg_flag b_post b_theory req_email_verify
    gin_mode="$(resolve_existing_or_default "GIN_MODE" "release")"
    swg_flag="$(resolve_existing_or_default "ENABLE_SWAGGER" "false")"
    b_post="$(resolve_existing_or_default "BYPASS_POST_TEST_THRESHOLD" "false")"
    b_theory="$(resolve_existing_or_default "BYPASS_THEORY_GATE" "false")"
    req_email_verify="$(resolve_existing_or_default "REQUIRE_EMAIL_VERIFICATION" "true")"

    write_merged_env <<EOF
POSTGRES_USER=${pg_user}
POSTGRES_PASSWORD=${pg_pass}
POSTGRES_DB=${pg_db}
JWT_SECRET=${jwt_secret}
ADMIN_EMAIL=${admin_email}
ADMIN_PASSWORD=${admin_pass}
ADMIN_NICKNAME=${admin_nick}
APP_ENVIRONMENT=${MODE}
GIN_MODE=${gin_mode}
ENABLE_SWAGGER=${swg_flag}
BYPASS_POST_TEST_THRESHOLD=${b_post}
BYPASS_THEORY_GATE=${b_theory}
REQUIRE_EMAIL_VERIFICATION=${req_email_verify}
APP_BASE_URL=${frontend_url}
FRONTEND_URL=${frontend_url}
DOMAIN=${domain_val}
ACME_EMAIL=${acme_val}
COMPOSE_PATH_SEPARATOR=:
COMPOSE_FILE=${compose_file}
LLM_API_KEY=${llm_key}
LLM_BASE_URL=${llm_url}
LLM_MODEL=${llm_model}
EOF

    CFG_PG_USER="$pg_user"
    CFG_PG_DB="$pg_db"
    CFG_FRONTEND_URL="$frontend_url"
    CFG_DOMAIN="$domain_val"
    CFG_ADMIN_EMAIL="$admin_email"
    CFG_ADMIN_PASS="$admin_pass"
    CFG_GEN_ADMIN="$generated_admin"
    CFG_LLM_READY=$([[ -n "$llm_key" ]] && echo true || echo false)
}

wait_container_healthy() {
    local cname="$1"
    local timeout="${2:-120}"
    local elapsed=0
    log_info "Waiting for $cname to report healthy..."
    while [[ "$elapsed" -lt "$timeout" ]]; do
        local st
        st="$("${DOCKER_BIN[@]}" inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$cname" 2>/dev/null || echo "")"
        if [[ "$st" == "healthy" ]]; then
            log_success "$cname is healthy."
            return 0
        fi
        sleep 2
        elapsed=$((elapsed + 2))
        printf '.'
    done
    printf '\n'
    fail_stage "Container Readiness ($cname)" 1 "Timed out after ${timeout}s waiting for $cname." "docker compose logs --tail=80" "Inspect container logs and re-run ./install.sh."
}

cmd_seed() {
    configure_docker_cmd || fail_stage "Docker Runtime" 1 "Docker daemon is not running." "docker info" "Start Docker and retry."
    log_info "Executing idempotent containerized database seeding..."
    if [[ -n "$IMAGE_TAG" ]]; then
        run_compose pull seed || fail_stage "Pull Seeder Image" $? "Failed to pull tagged seed image (${IMAGE_TAG})." "docker compose pull seed" ""
        run_compose up --no-build --no-deps --force-recreate --exit-code-from seed seed || \
            fail_stage "Database Seeding" $? "Containerized seed failed cleanly." "docker compose logs seed" "Run ./install.sh seed."
    else
        run_compose up --build --no-deps --force-recreate --exit-code-from seed seed || \
            fail_stage "Database Seeding" $? "Containerized seed failed cleanly before starting backend." "docker compose logs seed" "Resolve content/DB error and run ./install.sh seed."
    fi
    log_success "Containerized seeding completed cleanly."
}

db_count() {
    local pg_user="$1" pg_db="$2" sql="$3"
    local val=""
    val="$(run_compose exec -T db psql -U "$pg_user" -d "$pg_db" -t -A -c "$sql" 2>/dev/null | tr -d '[:space:]')" || val="-1"
    [[ "$val" =~ ^[0-9]+$ ]] && printf '%s' "$val" || printf -- '-1'
}

cmd_check_seed() {
    configure_docker_cmd || fail_stage "Docker Runtime" 1 "Docker daemon is not running." "docker info" ""
    get_pg_config

    log_info "Checking PostgreSQL seeded counts and integrity..."
    local d q st th sc u s
    d="$(db_count "$CFG_PG_USER" "$CFG_PG_DB" "SELECT COUNT(*) FROM security_domains;")"
    q="$(db_count "$CFG_PG_USER" "$CFG_PG_DB" "SELECT COUNT(*) FROM questions;")"
    st="$(db_count "$CFG_PG_USER" "$CFG_PG_DB" "SELECT COUNT(*) FROM theory_subtopics;")"
    th="$(db_count "$CFG_PG_USER" "$CFG_PG_DB" "SELECT COUNT(*) FROM theory_translations WHERE locale = 'th';")"
    sc="$(db_count "$CFG_PG_USER" "$CFG_PG_DB" "SELECT COUNT(*) FROM scenarios;")"
    u="$(db_count "$CFG_PG_USER" "$CFG_PG_DB" "SELECT COUNT(*) FROM users;")"
    s="$(db_count "$CFG_PG_USER" "$CFG_PG_DB" "SELECT COUNT(*) FROM simulation_sessions;")"

    printf '  security_domains    : %s (expected >= 5)\n' "$d"
    printf '  questions           : %s (expected >= 540)\n' "$q"
    printf '  theory_subtopics    : %s (expected >= 28)\n' "$st"
    printf '  theory_translations : %s (expected >= 33)\n' "$th"
    printf '  scenarios           : %s (expected >= 15)\n' "$sc"
    printf '  users (runtime)     : %s\n' "$u"
    printf '  simulation_sessions : %s\n' "$s"

    if [[ "$d" -lt 5 || "$q" -lt 540 || "$st" -lt 28 || "$th" -lt 33 || "$sc" -lt 15 ]]; then
        fail_stage "Seed Verification" 1 "Seed counts incomplete or DB unreachable." "./install.sh seed" "Verify db is running, run './install.sh seed', and re-verify."
    fi
    log_success "All seed integrity assertions passed!"
}

http_get_body() {
    local url="$1"
    if command -v curl >/dev/null 2>&1; then
        curl -fsS --max-time 6 "$url"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$url"
    else
        return 1
    fi
}

run_smoke_verify() {
    log_info "Running non-polluting post-install smoke checks..."
    cmd_check_seed

    local health_ok=false hbody=""
    for _ in $(seq 1 15); do
        if hbody="$(http_get_body "http://127.0.0.1:8080/api/health" 2>/dev/null)" && \
           [[ "$hbody" == *'"status":"healthy"'* && "$hbody" == *'"db_status":"up"'* ]]; then
            health_ok=true
            break
        fi
        sleep 2
    done
    [[ "$health_ok" == true ]] || fail_stage "Backend Health Smoke Check" 1 "GET /api/health did not report healthy + db_status=up." "docker compose logs backend" "Inspect backend logs."
    log_success "Backend Health (/api/health): HEALTHY (db_status=up)"

    if ! run_compose exec -T frontend wget -q -O /dev/null http://127.0.0.1/ >/dev/null 2>&1; then
        fail_stage "Frontend Smoke Check" 1 "Frontend HTTP server check failed." "docker compose logs frontend" "Inspect frontend logs."
    fi
    if [[ "${TLS_ACTIVE:-false}" == true && -n "${CFG_DOMAIN:-}" ]]; then
        "${DOCKER_BIN[@]}" exec cyber_sim_caddy caddy version >/dev/null 2>&1 || \
            fail_stage "Caddy TLS Proxy Check" 1 "cyber_sim_caddy container check failed." "docker compose logs caddy" "Inspect Caddy logs."
        if command -v curl >/dev/null 2>&1 && curl -kfsS --max-time 6 "https://${CFG_DOMAIN}/" >/dev/null 2>&1; then
            log_success "Cloud HTTPS Endpoint (https://${CFG_DOMAIN}/): READY"
        else
            log_warn "HTTPS endpoint https://${CFG_DOMAIN}/ did not answer yet (DNS/ACME issuance may still be in progress)."
        fi
    else
        log_success "Frontend Web UI: READY"
    fi

    if [[ -n "${CFG_ADMIN_EMAIL:-}" && -n "${CFG_ADMIN_PASS:-}" ]] && command -v curl >/dev/null 2>&1; then
        local payload resp esc_email esc_pass
        esc_email="$(json_escape "$CFG_ADMIN_EMAIL")"
        esc_pass="$(json_escape "$CFG_ADMIN_PASS")"
        payload="{\"email\":\"${esc_email}\",\"password\":\"${esc_pass}\"}"
        # Pass payload via stdin (@-) so passwords never appear in /proc/cmdline
        if resp="$(printf '%s' "$payload" | curl -fsS --max-time 8 -H 'Content-Type: application/json' --data-binary @- http://127.0.0.1:8080/api/login 2>/dev/null)" && [[ "$resp" == *'"token"'* ]]; then
            log_success "Admin Authentication Smoke Check: VERIFIED (${CFG_ADMIN_EMAIL})"
        else
            fail_stage "Admin Auth Smoke Check" 1 "Admin login check failed for ${CFG_ADMIN_EMAIL}." "docker compose logs backend" "Verify ADMIN_EMAIL / ADMIN_PASSWORD in .env."
        fi
    fi

    if [[ "$VERIFY_E2E" == true ]]; then
        [[ "${MODE:-local}" == "cloud" ]] && fail_stage "Safety Guard" 1 "--verify-e2e is forbidden in Cloud mode." "" ""
        if command -v python3 >/dev/null 2>&1; then
            python3 -u "${SCRIPT_DIR}/scripts/test_integration.py" "http://127.0.0.1:8080/api"
        elif command -v python >/dev/null 2>&1; then
            python -u "${SCRIPT_DIR}/scripts/test_integration.py" "http://127.0.0.1:8080/api"
        else
            fail_stage "E2E Verification" 1 "python3 is required for optional --verify-e2e." "" "Install python3 or omit --verify-e2e."
        fi
    fi
}

cmd_backup() {
    configure_docker_cmd || fail_stage "Docker Runtime" 1 "Docker daemon is not running." "docker info" ""
    wait_container_healthy "cyber_sim_db" 60
    get_pg_config

    local bdir="${SCRIPT_DIR}/backups"
    mkdir -p "$bdir"
    local ts
    ts="$(date +"%Y%m%d_%H%M%S")"
    local sql_gz="${bdir}/cyber_sim_${ts}.sql.gz"
    local tmp_gz="${sql_gz}.tmp"

    log_info "Backing up PostgreSQL '$CFG_PG_DB' to $sql_gz..."
    if ! run_compose exec -T db pg_dump --no-owner --no-privileges -U "$CFG_PG_USER" "$CFG_PG_DB" | gzip -c > "$tmp_gz" || [[ ! -s "$tmp_gz" ]]; then
        rm -f "$tmp_gz"
        fail_stage "Database Backup" 1 "pg_dump backup failed." "docker compose logs db" "Verify database health."
    fi
    mv "$tmp_gz" "$sql_gz"
    log_success "Created DB backup: $sql_gz"

    if [[ "$NO_UPLOADS" != true ]] && "${DOCKER_BIN[@]}" volume inspect soc-trainer_uploads >/dev/null 2>&1; then
        local up_gz="${bdir}/cyber_sim_uploads_${ts}.tar.gz"
        # shellcheck disable=SC2016
        "${DOCKER_BIN[@]}" run --rm -v soc-trainer_uploads:/data:ro -v "${bdir}:/backup" alpine:latest sh -c 'tar -czf "/backup/$1" -C /data .' _ "cyber_sim_uploads_${ts}.tar.gz" || \
            log_warn "Uploads archive failed cleanly; DB backup remains intact."
        [[ -f "$up_gz" ]] && log_success "Created uploads archive: $up_gz"
    fi

    local count=0
    while IFS= read -r f; do
        count=$((count + 1))
        if [[ "$count" -gt 10 ]]; then
            local base ts_old
            base="$(basename "$f")"
            if [[ "$base" =~ ^cyber_sim_([0-9]{8}_[0-9]{6})\.sql\.gz$ ]]; then
                ts_old="${BASH_REMATCH[1]}"
                rm -f "${bdir}/cyber_sim_uploads_${ts_old}.tar.gz"
            fi
            rm -f "$f"
        fi
    done < <(ls -1t "${bdir}"/cyber_sim_[0-9]*_[0-9]*.sql.gz 2>/dev/null || true)
}

cmd_restore() {
    local bfile
    bfile="$(resolve_caller_path "$RESTORE_FILE")"
    [[ -n "$bfile" && -f "$bfile" && -s "$bfile" ]] || \
        fail_stage "Restore Validation" 2 "Specify an existing non-empty .sql.gz or .sql backup file." "./install.sh restore backups/cyber_sim_<timestamp>.sql.gz" ""
    configure_docker_cmd || fail_stage "Docker Runtime" 1 "Docker daemon is not running." "docker info" ""
    get_pg_config

    [[ "$NON_INTERACTIVE" == true ]] && fail_stage "Restore Guard" 1 "Destructive restore requires typing 'RESTORE' interactively (blocked in --non-interactive)." "" ""
    log_warn "Restore will overwrite database '$CFG_PG_DB' using $bfile"
    local typed
    typed="$(prompt_line "Type RESTORE to confirm: ")"
    [[ "$typed" == "RESTORE" ]] || { log_info "Restore cancelled."; return 0; }

    log_info "Taking pre-restore safety backup..."
    cmd_backup

    log_info "Stopping application containers..."
    local cf
    cf="$(read_env_value "COMPOSE_FILE")"
    if [[ "$cf" == *cloud* ]]; then
        run_compose stop backend frontend caddy || true
    else
        run_compose stop backend frontend || true
    fi
    wait_container_healthy "cyber_sim_db" 60

    log_info "Recreating clean database '$CFG_PG_DB'..."
    run_compose exec -T db psql -U "$CFG_PG_USER" -d postgres -v ON_ERROR_STOP=1 -c \
        "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${CFG_PG_DB}' AND pid <> pg_backend_pid(); DROP DATABASE IF EXISTS \"${CFG_PG_DB}\"; CREATE DATABASE \"${CFG_PG_DB}\" OWNER \"${CFG_PG_USER}\";" || \
        fail_stage "Database Recreate" 1 "Could not drop/recreate '$CFG_PG_DB'." "docker compose logs db" ""

    log_info "Restoring PostgreSQL backup stream..."
    if [[ "$bfile" == *.gz ]]; then
        gzip -dc "$bfile" | run_compose exec -T db psql -v ON_ERROR_STOP=1 -U "$CFG_PG_USER" -d "$CFG_PG_DB"
    else
        run_compose exec -T db psql -v ON_ERROR_STOP=1 -U "$CFG_PG_USER" -d "$CFG_PG_DB" < "$bfile"
    fi
    log_success "Database restored cleanly."

    local bn bdir
    bn="$(basename "$bfile")"
    bdir="$(cd "$(dirname "$bfile")" && pwd)"
    if [[ "$bn" =~ ^cyber_sim_([0-9]{8}_[0-9]{6})\.sql\.gz$ && -f "${bdir}/cyber_sim_uploads_${BASH_REMATCH[1]}.tar.gz" ]]; then
        local up_base="cyber_sim_uploads_${BASH_REMATCH[1]}.tar.gz"
        log_info "Restoring matching uploads archive: ${bdir}/${up_base}"
        # shellcheck disable=SC2016
        "${DOCKER_BIN[@]}" run --rm -v soc-trainer_uploads:/data -v "${bdir}:/backup:ro" alpine:latest sh -c \
            'find /data -mindepth 1 -delete && tar -xzf "/backup/$1" -C /data && chown -R 65534:65534 /data' _ "$up_base"
        log_success "Uploads archive restored cleanly."
    fi

    cmd_seed
    run_compose up -d --remove-orphans
    wait_container_healthy "cyber_sim_backend" 120
    log_success "Restore sequence completed and application restarted."
}

cmd_install() {
    cmd_check true
    setup_env_config

    if [[ "$MODE" == "cloud" ]]; then
        audit_cloud_firewall true "$TLS_ACTIVE"
    else
        audit_cloud_firewall false false
    fi

    log_info "Starting PostgreSQL database..."
    run_compose up -d db || fail_stage "Start Database" $? "Failed to start db service." "docker compose logs db" "Verify .env secrets and Docker daemon."
    wait_container_healthy "cyber_sim_db" 120

    cmd_seed

    log_info "Starting application services..."
    if [[ -n "$IMAGE_TAG" ]]; then
        run_compose pull backend frontend || fail_stage "Pull Application Images" $? "Failed to pull tagged images (${IMAGE_TAG})." "docker compose pull" ""
        run_compose up -d --no-build --remove-orphans || fail_stage "Start Application Stack" $? "Failed to start application services." "docker compose logs" ""
    else
        run_compose up -d --build --remove-orphans || fail_stage "Start Application Stack" $? "Failed to build/start application services." "docker compose logs" "Inspect error and re-run ./install.sh install."
    fi

    wait_container_healthy "cyber_sim_backend" 120
    wait_container_healthy "cyber_sim_frontend" 90
    if [[ "$TLS_ACTIVE" == true ]]; then
        log_info "Caddy Auto-HTTPS active on :80/:443 for ${CFG_DOMAIN}."
    fi

    run_smoke_verify

    printf '\n========================================================\n'
    log_success "SOC Trainer Installed & Verified (${MODE} mode)"
    printf '========================================================\n'
    printf '  Web Frontend     : %s\n' "$CFG_FRONTEND_URL"
    printf '  Backend API      : http://127.0.0.1:8080/api\n'
    printf '  Health Endpoint  : http://127.0.0.1:8080/api/health (HEALTHY / DB UP)\n'
    if [[ "$CFG_LLM_READY" == true ]]; then
        printf '  LLM Mode         : Live AI Configured (key redacted)\n'
    else
        printf '  LLM Mode         : Static Fallback Mode (key omitted)\n'
    fi
    printf '  Secrets (.env)   : Locked chmod 600 (DB/JWT/LLM secrets redacted)\n'
    if [[ -n "${CFG_GEN_ADMIN:-}" && "$NON_INTERACTIVE" != true ]]; then
        printf '\n  Initial Admin Login (shown ONCE — please save & change password):\n'
        printf '    Email          : %s\n' "$CFG_ADMIN_EMAIL"
        printf '    Password       : %s\n' "$CFG_GEN_ADMIN"
    fi
    printf '========================================================\n'
}

cmd_update() {
    cmd_check false
    log_info "Starting safe update pipeline (auto-backup -> optional pull -> rebuild/pull -> seed -> health verify)..."
    cmd_backup

    if [[ "$NO_PULL" != true && -d "${SCRIPT_DIR}/.git" ]] && command -v git >/dev/null 2>&1; then
        if [[ -z "$(git status --porcelain)" ]]; then
            log_info "Pulling latest updates (git pull --ff-only)..."
            git pull --ff-only || log_warn "git pull --ff-only did not advance; continuing with current working tree."
        else
            log_warn "Local modifications present; skipping git pull (use --no-pull or commit/stash)."
        fi
    fi

    run_compose up -d db
    wait_container_healthy "cyber_sim_db" 120
    cmd_seed

    if [[ -n "$IMAGE_TAG" ]]; then
        run_compose pull backend frontend || fail_stage "Pull Application Images" $? "Failed to pull tagged images (${IMAGE_TAG})." "docker compose pull" "Pre-update backup saved in backups/."
        run_compose up -d --no-build --remove-orphans || fail_stage "Update Application Stack" $? "Update start failed." "docker compose logs" ""
    else
        run_compose up -d --build --remove-orphans || fail_stage "Update Application Stack" $? "Update build/start failed." "docker compose logs" "Pre-update backup saved in backups/."
    fi

    wait_container_healthy "cyber_sim_backend" 120
    wait_container_healthy "cyber_sim_frontend" 90
    CFG_ADMIN_EMAIL="$(read_env_value "ADMIN_EMAIL")"
    CFG_ADMIN_PASS="$(read_env_value "ADMIN_PASSWORD")"
    CFG_DOMAIN="$(read_env_value "DOMAIN")"
    MODE="$(read_env_value "APP_ENVIRONMENT")"; [[ -z "$MODE" ]] && MODE="local"
    TLS_ACTIVE=$([[ "$MODE" == "cloud" && -n "$CFG_DOMAIN" ]] && echo true || echo false)
    run_smoke_verify
    log_success "Stack updated cleanly!"
}

cmd_status() {
    configure_docker_cmd || fail_stage "Docker Runtime" 1 "Docker daemon is not running." "docker info" ""
    local env_m url_m llm_k dom_m
    env_m="$(read_env_value "APP_ENVIRONMENT")"; [[ -z "$env_m" ]] && env_m="unconfigured"
    url_m="$(read_env_value "FRONTEND_URL")"; [[ -z "$url_m" ]] && url_m="http://localhost"
    llm_k="$(read_env_value "LLM_API_KEY")"
    dom_m="$(read_env_value "DOMAIN")"

    printf 'SOC Trainer Deployment Status:\n'
    printf '  Mode / Env   : %s\n' "$env_m"
    printf '  Frontend URL : %s\n' "$url_m"
    if [[ -n "$dom_m" ]]; then
        printf '  Cloud TLS    : Enabled (%s)\n' "$dom_m"
    else
        printf '  Cloud TLS    : Disabled (HTTP)\n'
    fi
    if [[ -n "$llm_k" ]]; then
        printf '  LLM Mode     : Configured (key redacted)\n'
    else
        printf '  LLM Mode     : Static Fallback\n'
    fi
    printf '\n'
    run_compose ps
}

cmd_uninstall() {
    configure_docker_cmd || fail_stage "Docker Runtime" 1 "Docker daemon is not running." "docker info" ""
    local env_m
    env_m="$(read_env_value "APP_ENVIRONMENT")"
    if [[ "$REMOVE_VOLUMES" == true ]]; then
        if [[ "$env_m" == "cloud" ]]; then
            log_warn "CRITICAL: Cloud environment detected! Deleting Docker volumes erases research data & uploads permanently."
        else
            log_warn "Deleting Docker volumes permanently erases PostgreSQL data & uploads."
        fi
        [[ "$NON_INTERACTIVE" == true ]] && fail_stage "Uninstall Guard" 1 "Deleting Docker volumes requires interactive typed confirmation ('DELETE DATA')." "" ""
        local typed
        typed="$(prompt_line "Type 'DELETE DATA' to confirm permanent volume deletion: ")"
        if [[ "$typed" != "DELETE DATA" ]]; then
            log_info "Volume deletion cancelled. Running safe 'docker compose down' (keeping volumes)..."
            run_compose down --remove-orphans
            return 0
        fi
        run_compose down -v --remove-orphans
        log_success "Containers and persistent volumes removed."
        return 0
    fi
    run_compose down --remove-orphans
    log_success "Containers stopped and removed cleanly (persistent volumes preserved)."
}

cmd_reset_dev_data() {
    local env_m dep_m
    env_m="$(read_env_value "APP_ENVIRONMENT")"
    dep_m="$(read_env_value "DEPLOYMENT_MODE")"
    if [[ "$MODE" == "cloud" || "$env_m" == "cloud" || "$dep_m" == "cloud" ]]; then
        fail_stage "Reset Dev Data Guard" 1 "ABORTED: reset-dev-data is strictly prohibited in CLOUD environment." "" ""
    fi
    if [[ "$NON_INTERACTIVE" == true ]]; then
        fail_stage "Reset Dev Data Guard" 1 "Destructive reset requires interactive confirmation ('RESET')." "" ""
    fi

    configure_docker_cmd || fail_stage "Docker Runtime" 1 "Docker daemon is not running." "docker info" ""
    get_pg_config
    log_warn "=== DANGER ZONE: RESET DEV DATA ==="
    log_warn "This truncates learner sessions, questionnaires, notifications, threads, and NFR runs."
    local confirm
    confirm="$(prompt_line "Type 'RESET' to confirm: ")"
    [[ "$confirm" == "RESET" ]] || { log_info "Operation cancelled."; return 0; }

    run_compose exec -T db psql -U "$CFG_PG_USER" -d "$CFG_PG_DB" -v ON_ERROR_STOP=1 -c "$RESET_TRUNCATE_SQL"
    log_success "Runtime tables truncated. Re-running idempotent seed..."
    cmd_seed
}

show_interactive_menu() {
    while true; do
        printf '\n===============================================\n'
        printf ' SOC Trainer — Linux Setup & Operations Menu\n'
        printf '===============================================\n'
        printf '  1) Install / Setup (Local Mode)\n'
        printf '  2) Install / Setup (Cloud Mode + Auto-HTTPS)\n'
        printf '  3) Preflight Check (check)\n'
        printf '  4) Service Status (status)\n'
        printf '  5) Verify Seed Data (check-seed)\n'
        printf '  6) Backup Database + Uploads (backup)\n'
        printf '  7) Restore from Backup (restore)\n'
        printf '  8) Safe Update (update)\n'
        printf '  9) View Container Logs (logs)\n'
        printf ' 10) Stop / Uninstall Stack (uninstall)\n'
        printf ' 11) Reset Dev Runtime Data (local only)\n'
        printf '  0) Exit\n'
        local sel
        sel="$(prompt_line "Select option [0-11]: ")"
        case "$sel" in
            1) MODE="local"; cmd_install ;;
            2) MODE="cloud"; cmd_install ;;
            3) cmd_check false ;;
            4) cmd_status ;;
            5) cmd_check_seed ;;
            6) cmd_backup ;;
            7) RESTORE_FILE="$(prompt_line "Path to .sql.gz backup file: ")"; cmd_restore ;;
            8) cmd_update ;;
            9)
                configure_docker_cmd || continue
                local svc
                svc="$(prompt_line "Service name (Enter for all): ")"
                if [[ -n "$svc" ]]; then run_compose logs --tail=100 "$svc"; else run_compose logs --tail=80; fi
                ;;
            10) cmd_uninstall ;;
            11) cmd_reset_dev_data ;;
            0) return 0 ;;
            *) log_warn "Invalid selection: $sel" ;;
        esac
    done
}

main() {
    parse_args "$@"
    if [[ -z "$SUBCOMMAND" ]]; then
        if [[ "$NON_INTERACTIVE" == true || ( ! -t 0 && ! -r /dev/tty ) ]]; then
            NON_INTERACTIVE=true
            SUBCOMMAND="install"
        else
            show_interactive_menu
            exit 0
        fi
    fi

    case "$SUBCOMMAND" in
        install)        cmd_install ;;
        check)          cmd_check false ;;
        status)         cmd_status ;;
        seed)           cmd_seed ;;
        check-seed)     cmd_check_seed ;;
        update)         cmd_update ;;
        backup)         cmd_backup ;;
        restore)        cmd_restore ;;
        logs)
            configure_docker_cmd || fail_stage "Docker Runtime" 1 "Docker daemon is not running." "" ""
            if [[ -n "$LOGS_SERVICE" ]]; then run_compose logs -f "$LOGS_SERVICE"; else run_compose logs -f; fi
            ;;
        uninstall)      cmd_uninstall ;;
        reset-dev-data) cmd_reset_dev_data ;;
        *)
            show_usage
            fail_stage "Command Dispatch" 2 "Unknown subcommand: $SUBCOMMAND" "./install.sh --help" ""
            ;;
    esac
}

main "$@"
