#!/usr/bin/env bash
# ==============================================================================
# SOC Trainer - Clean-Machine Linux / macOS Bootstrapper (scripts/bootstrap.sh)
# Ensures git is present, clones or fast-forwards the repository, and launches
# install.sh with all forwarded arguments.
# ==============================================================================

set -euo pipefail

REPO_URL="${SOC_REPO_URL:-https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training.git}"
BRANCH="${SOC_REPO_BRANCH:-main}"
TARGET_DIR="${SOC_INSTALL_DIR:-Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training}"

AUTO_DEPS=false
NON_INTERACTIVE=false
FORWARD_ARGS=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        --repo-url)
            [[ $# -ge 2 ]] || { echo "[ERROR] Missing value after --repo-url" >&2; exit 2; }
            REPO_URL="$2"; shift 2 ;;
        --branch)
            [[ $# -ge 2 ]] || { echo "[ERROR] Missing value after --branch" >&2; exit 2; }
            BRANCH="$2"; shift 2 ;;
        --dir)
            [[ $# -ge 2 ]] || { echo "[ERROR] Missing value after --dir" >&2; exit 2; }
            TARGET_DIR="$2"; shift 2 ;;
        --auto-deps)
            AUTO_DEPS=true
            FORWARD_ARGS+=("$1"); shift ;;
        --non-interactive)
            NON_INTERACTIVE=true
            FORWARD_ARGS+=("$1"); shift ;;
        *)
            FORWARD_ARGS+=("$1"); shift ;;
    esac
done

if [[ ${#FORWARD_ARGS[@]} -eq 0 ]]; then
    FORWARD_ARGS=("install")
fi

log_info()  { printf '[INFO] %s\n' "$1"; }
log_warn()  { printf '[WARN] %s\n' "$1" >&2; }
log_error() { printf '[ERROR] %s\n' "$1" >&2; }

run_root() {
    if [[ "$(id -u)" -eq 0 ]]; then
        "$@"
    elif command -v sudo >/dev/null 2>&1; then
        sudo "$@"
    else
        log_error "Root privileges or sudo are required to install missing system packages ($*)."
        exit 1
    fi
}

install_git_and_curl() {
    local os_name
    os_name="$(uname -s)"
    if [[ "$os_name" == "Darwin" ]]; then
        if ! command -v git >/dev/null 2>&1; then
            log_error "Git is required on macOS. Run 'xcode-select --install' or 'brew install git' and re-run."
            exit 1
        fi
        return 0
    fi

    if command -v apt-get >/dev/null 2>&1; then
        run_root apt-get update -y
        run_root apt-get install -y git curl ca-certificates
    elif command -v dnf >/dev/null 2>&1; then
        run_root dnf install -y git curl ca-certificates
    elif command -v yum >/dev/null 2>&1; then
        run_root yum install -y git curl ca-certificates
    else
        log_error "Unsupported package manager. Install git and curl manually and re-run."
        exit 1
    fi
}

log_info "SOC Trainer Linux / macOS Clean-Machine Bootstrapper"

if ! command -v git >/dev/null 2>&1 || ! command -v curl >/dev/null 2>&1; then
    do_install="$AUTO_DEPS"
    if [[ "$do_install" != true && "$NON_INTERACTIVE" != true && -r /dev/tty ]]; then
        printf "Missing required tool(s) (git/curl). Install now? [Y/n]: "
        read -r ans < /dev/tty || ans="y"
        if [[ -z "$ans" || "$ans" =~ ^[Yy] ]]; then
            do_install=true
        fi
    fi
    if [[ "$do_install" != true ]]; then
        log_error "git and curl are required. Re-run with --auto-deps or install them manually."
        exit 1
    fi
    install_git_and_curl
fi

if [[ -f "docker-compose.yml" && -f "install.sh" ]]; then
    TARGET_DIR="$(pwd)"
    log_info "Detected existing repository at $TARGET_DIR"
elif [[ -d "$TARGET_DIR/.git" ]]; then
    log_info "Updating existing clone in $TARGET_DIR (branch $BRANCH)..."
    git -C "$TARGET_DIR" pull --ff-only origin "$BRANCH" || log_warn "Could not fast-forward cleanly; continuing with existing tree."
else
    log_info "Cloning repository ($BRANCH) into $TARGET_DIR..."
    git clone --branch "$BRANCH" "$REPO_URL" "$TARGET_DIR"
fi

cd "$TARGET_DIR"
chmod +x install.sh
exec ./install.sh "${FORWARD_ARGS[@]}"
