# Workflow Spec: Cross-Platform Setup & Cloud Bootstrap (`install.sh` + `install.ps1`)

**Status:** `/loop-me` Grilling Rounds 1–3 Settled (`Q1`–`Q17` Resolved, 2026-11-25) — Ready for `/to-spec` Seam Confirmation & Issue Publication  
**Depends On / Supersedes:**
- [Spec 6: Hardened Docker Compose Deployment Stack](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/docs/specs/spec-6-hardened-docker-compose-deployment-stack.md) (`docker-compose.yml`, fail-closed secrets, least-privilege containers, one-shot `seed`, optional `--profile tools`)
- [nfr_evaluation.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/nfr_evaluation.md) (`GET /api/health` DB-ping contract, `APP_ENVIRONMENT` tracking)
- *Supersedes legacy drafts*: [docs/INSTALL_SH_SPECIFICATION.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/docs/INSTALL_SH_SPECIFICATION.md) & [docs/DEPLOYMENT_BOOTSTRAP_PLAN.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/docs/DEPLOYMENT_BOOTSTRAP_PLAN.md) (outdated `install.sh` host-Python flow and open decisions resolved here)

---

## Objective

Provide an **end-to-end, fail-closed, cross-platform Setup & Lifecycle Suite** so deploying the platform from a **clean machine or Cloud VM** (even before installing `git` or `Docker`) succeeds cleanly in both **Linux/Cloud (`install.sh`)** and **Windows (`install.ps1` + `install.bat`)** environments, with optional zero-config **Auto-HTTPS (Caddy + Let's Encrypt)** for Cloud domains, zero host Python/Go/Node runtime requirements, full secret hygiene, non-polluting smoke verification, and safe database + upload volume `backup` / `restore` / `update` / `uninstall` operations.

---

## 1. Loop Profile

- **Loop Name**: Cross-Platform Clean-Machine Setup, Cloud TLS Bootstrap, & Ops Lifecycle Loop
- **Frequency / Cadence**:
  - **Initial Clean-Machine Setup**: `scripts/bootstrap.sh` (`curl ... | bash`) or `scripts/bootstrap.ps1` (`irm ... | iex`), OR repo-local `./install.sh` / `.\install.ps1` (`install.bat`).
  - **Routine Operations**: On-demand `check`, `status`, `logs`, `seed`, `check-seed`, `backup`, `restore`, `update`, `uninstall`, and `reset-dev-data` (Local-only).
- **Primary Actors & Deliverables**:
  - **Clean-VM Bootstrappers (`scripts/bootstrap.sh`, `scripts/bootstrap.ps1`)**: Lightweight entrypoints for bare machines that install `git` if missing, clone/update this repo, and hand off cleanly to `install.sh` / `install.ps1`.
  - **Linux / Cloud Installer (`install.sh`)**: POSIX/Bash 4+ setup & lifecycle CLI supporting Ubuntu/Debian (`apt`), RHEL/AlmaLinux/Rocky/Fedora (`dnf`/`yum`), Amazon Linux 2023, and macOS (Docker Desktop verification).
  - **Windows Installer & Wrapper (`install.ps1`, `install.bat`)**: PowerShell 5.1+/7+ setup & lifecycle CLI with full subcommand & menu parity, WSL2 & Docker Desktop orchestration, `winget` bootstrapper, Windows Defender Firewall opt-in rules, `icacls` `.env` lockdown, and `install.bat` ExecutionPolicy bypass wrapper.
  - **Cloud HTTPS Overlay (`docker-compose.cloud.yml`, `Caddyfile.template`)**: Optional Compose override adding `caddy:2-alpine` (`80/443`, automatic Let's Encrypt TLS) while clearing conflicting `frontend` host port bindings (`ports: !reset []`).
  - **Configuration & Cross-OS Hygiene (`.env.example`, `.gitattributes`, `README.md`, legacy doc supersede notes)**: Complete `.env.example` matching all backend/Compose/Cloud variables, `.gitattributes` enforcing `LF` on shell/Docker/Caddy files and `CRLF` on Windows `.ps1`/`.bat` scripts, and consolidated setup docs.
- **Trigger**:
  - Manual terminal invocation with interactive menu (when run with no args on a TTY), explicit CLI subcommands, or unattended Cloud-Init / CI automation (`--non-interactive`).
- **Checkpoint**:
  - **Pre-Destructive Interactive Checkpoints**:
    - Explicit opt-in prompt before installing host OS packages (`git`, `docker`, WSL2), creating swap, modifying firewall rules (`--configure-firewall`), or deleting Docker volumes (`uninstall --volumes` requires typed `DELETE DATA`).
    - Explicit typed `RESTORE` guard before `restore`; explicit typed `RESET` guard (and strict refusal when `APP_ENVIRONMENT=cloud`) before `reset-dev-data`.
  - **Implementation Acceptance Verification**: Smoke-tested end-to-end on both Windows (`install.ps1` / `install.bat` against live Docker Desktop) and Linux (`install.sh` syntax/static/E2E Compose & CLI tests) before closing the feature issue.

---

## 2. Settled Design Decisions (`Q1`–`Q17`)

| Q# | Decision | Settled Rule |
| --- | --- | --- |
| `Q1` | Target Runtime Architecture | **Unified Self-Hosted Docker Compose (`1A`)**: Same `db → seed → backend → frontend` Compose architecture across Local machines and Cloud VMs (hardened Spec 6 baseline). |
| `Q2` | Clean-Machine Prerequisites | **Smart Auto-Bootstrap (`2A`)**: Confirm (or `--auto-deps`) to install `git`/Docker Engine+Compose on Linux; check/install WSL2 + Docker Desktop via `winget` on Windows with clean pause & resume message if reboot is required. |
| `Q3` | Cloud Public Exposure & HTTPS | **Optional Caddy Auto-HTTPS Override (`3A`)**: `--mode local` uses HTTP `:80`; `--mode cloud` prompts Domain vs Public IP, activating `caddy` TLS (`:80/:443`) + setting `FRONTEND_URL=https://<domain>` when a domain is provided, or HTTP `:80` + clear warning for IP-only. |
| `Q4` | Secrets & Admin Provisioning | **Zero Weak Defaults (`4A`)**: CSPRNG-generated `POSTGRES_PASSWORD` & `JWT_SECRET`; `ADMIN_EMAIL` / hidden-or-generated `ADMIN_PASSWORD`; optional `LLM_API_KEY` (Static Fallback allowed); lock `.env` (`chmod 600` / `icacls`). |
| `Q5` | Seeding & Smoke Checks | **100% Containerized Seed + Non-Polluting Smoke Verification (`5A`)**: Seed strictly via Docker (`seed` service — zero host Python/Go/Node needed); verify DB counts, `/api/health`, frontend HTTP, and Admin login without creating synthetic learner data (`--verify-e2e` optional in `--mode local` only). |
| `Q6` | Build vs Pull Image Strategy | **Source Build Default + Low-RAM Cloud Swap + `--image-tag` (`6A`)**: `docker compose up -d --build` by default; offer 2 GB swap on Linux Cloud VMs when RAM < 2 GB without swap; optional `--image-tag` flag for versioned registry pulls. |
| `Q7` | Command & UX Parity | **Full Lifecycle Parity + Interactive Menu (`7A`)**: Identical `install`, `check`, `status`, `seed`, `check-seed`, `update`, `backup`, `restore`, `logs`, `uninstall`, `reset-dev-data` across Linux & Windows + interactive menu when called bare + `--non-interactive`. |
| `Q8` | Empty-Machine Entry & Windows ExecutionPolicy | **OS-Specific Bootstrappers + `install.bat` (`8A`)**: Linux (`apt`/`dnf`/`yum`/AL2023 + macOS Docker check) & Windows (`install.ps1` + `install.bat` bypassing ExecutionPolicy cleanly). |
| `Q9` | Cloud Proxy Port Isolation & Compose Persistence | **`ports: !reset []` + `COMPOSE_FILE` in `.env` (`9A`)**: Persist `COMPOSE_FILE=docker-compose.yml:docker-compose.cloud.yml` and `COMPOSE_PATH_SEPARATOR=:` in `.env`; render `Caddyfile` from domain/email inputs so all subcommands stay consistent. |
| `Q10` | Secret Precedence & Redacted Output | **Env → `_FILE` → Existing `.env` → Prompt/CSPRNG (`10A`)**: Suppress all secrets in stdout/logs (`POSTGRES_PASSWORD`, `JWT_SECRET`, `LLM_API_KEY` never printed); print auto-generated `ADMIN_PASSWORD` once on interactive TTY only (`--non-interactive` keeps silent in `.env`). |
| `Q11` | Cross-OS Line Endings | **`.gitattributes` + Preflight CRLF Guard (`11A`)**: Enforce `eol=lf` on `*.sh`, `Dockerfile*`, `Caddyfile*` and `eol=crlf` on `*.ps1`, `*.bat`; detect/warn on `\r` corruption in Linux bash preflight. |
| `Q12` | Backup, Restore & Update Lifecycle | **Gzipped SQL + Uploads Volume + Guarded Restore (`12A`)**: `backup` reads actual `.env` credentials (`POSTGRES_USER`, `POSTGRES_DB`), exports `.sql.gz` + `uploads` volume `.tar.gz`, retains latest 10 sets; `update` runs auto-backup → optional pull → rebuild & `--force-recreate` seed → health checks; `restore` requires typed `RESTORE` + pre-restore safety backup. |
| `Q13` | Cloud Firewall Handling | **Inspect & Recommend Default + Opt-In `--configure-firewall` (`13A`)**: Show port audit & Security Group checklist without silently modifying rules; apply `ufw` (`22,80,443`) / Windows Firewall (`80,443`) only on explicit flag or interactive confirmation. |
| `Q14` | Windows Runtime Model | **Docker Desktop + WSL2 Primary Target (`14A`)**: Native Linux container execution via Docker Desktop/WSL2 on Windows 10/11 & Server; check daemon status, attempt launch if stopped, and verify WSL2 prerequisites. |
| `Q15` | Repo-less Bootstrapper Scripts | **Standalone `scripts/bootstrap.sh` & `scripts/bootstrap.ps1` (`15A`)**: Ensure `git`, clone/fast-forward repo cleanly, and chain straight into `install.sh` / `install.ps1`. |
| `Q16` | Docs & `.env.example` Consolidation | **Single Source of Truth (`16A`)**: Update `.env.example` (including `APP_ENVIRONMENT`, `DOMAIN`, `ACME_EMAIL`, `COMPOSE_FILE`, `COMPOSE_PATH_SEPARATOR`, `BYPASS_THEORY_GATE`), modernize `README.md`, and mark old `docs/INSTALL_SH_SPECIFICATION.md` / `docs/DEPLOYMENT_BOOTSTRAP_PLAN.md` superseded. |
| `Q17` | Daily Host Scheduled Backups | **Manual + Automatic Pre-Update/Restore Only (`17A`)**: No unattended OS Cron / Windows Task Scheduler created automatically; document optional 1-line schedule examples in `README.md`. |

---

## 3. Canonical Terms (Added to `NOTES.md`)

- **Cross-Platform Installer Harness (`install.sh` & `install.ps1`)**: Symmetric Linux Bash and Windows PowerShell CLIs (with `install.bat` ExecutionPolicy wrapper) managing the entire Docker Compose setup and operations lifecycle without requiring Go, Node, or Python on the host.
- **Clean-VM Bootstrapper (`scripts/bootstrap.sh` & `scripts/bootstrap.ps1`)**: Standalone remote/bare-machine launcher scripts that ensure `git` is installed, clone or fast-forward the repository, and invoke the local installer harness.
- **Cloud TLS Compose Overlay (`docker-compose.cloud.yml` + `Caddyfile`)**: Opt-in Docker Compose layer persisted via `COMPOSE_FILE` and `COMPOSE_PATH_SEPARATOR=:` in `.env` when a public Cloud `DOMAIN` is configured, suppressing conflicting `frontend` host ports (`!reset []`) so `caddy` handles HTTP/HTTPS (`:80/:443`) and automatic Let's Encrypt certificates cleanly.
- **Secret Resolution Precedence (`ENV → <KEY>_FILE → .env → Prompt/CSPRNG`)**: Fail-closed configuration contract preserving existing `.env` values across re-runs, generating cryptographically random `POSTGRES_PASSWORD`, `JWT_SECRET`, and optional `ADMIN_PASSWORD`, locking `.env` file permissions (`chmod 600` / owner-only `icacls`), and never echoing DB/JWT/LLM secrets to console output or logs.
- **Non-Polluting Smoke Verification**: Post-startup validation checking container health, PostgreSQL seeded counts (`domains`, `540` questions, `28` subtopics, `33` TH theory docs, `15` scenarios), `/api/health` (`db_status=up`), frontend HTTP `200`, and Admin login without writing synthetic learner/research rows.

---

## 4. Detailed Architecture & CLI Contract

### 4.1 Supported CLI Subcommands & Flags (`install.sh` / `install.ps1`)

```text
Subcommands:
  install            Preflight, env setup, start db -> seed -> backend -> frontend (+ caddy if Cloud domain), smoke verify
  check              Read-only preflight (OS, Docker/Compose, WSL2/systemd, ports, RAM/disk, .env permissions, line endings)
  status             Show container health, endpoints, active mode, redacted LLM/TLS state
  seed               Run idempotent containerized seed (`docker compose up --build --no-deps --force-recreate --exit-code-from seed seed`)
  check-seed         Read-only PostgreSQL count check (`security_domains`, `questions`, `theory_subtopics`, `theory_translations`, `scenarios`, runtime counts)
  update             Auto-backup -> optional git pull (--no-pull to skip) -> build/pull -> db & seed -> up backend frontend -> smoke check
  backup             Create gzipped DB dump (.sql.gz) + uploads volume archive (.tar.gz) in ./backups/ (retains latest 10 sets; --no-uploads opt-out)
  restore <file>     Verify archive -> confirm RESTORE -> safety backup -> stop app -> restore DB (+ matching uploads if present) -> seed/restart
  logs [service]     Stream Docker Compose logs safely (optional service filter)
  uninstall          Stop/remove containers cleanly (preserves pgdata/uploads volumes by default; --volumes requires typed DELETE DATA)
  reset-dev-data     Local-only wipe of runtime tables + re-seed; STRICTLY REFUSED when APP_ENVIRONMENT=cloud

Common Flags:
  --mode <local|cloud>       Target deployment profile (default: prompt or local)
  --non-interactive          Zero interactive prompts; auto-generate missing secrets silently into locked .env
  --domain <host>            Public hostname enabling Cloud Caddy HTTPS overlay
  --acme-email <email>       Contact email for Let's Encrypt TLS certificates
  --auto-deps                Auto-install missing prerequisites (git, Docker Engine/Desktop)
  --configure-firewall       Explicitly apply host firewall rules (ufw 22/80/443 or Windows Firewall 80/443)
  --image-tag <tag>          Optional versioned image tag override
  --verify-e2e               Also run scripts/test_integration.py (allowed in --mode local only)
  --with-pgadmin             Include Compose `tools` profile (loopback pgAdmin :5050)
  --skip-dns-check           Bypass DNS A-record check prior to Caddy TLS startup
  --no-pull / --no-uploads   Opt out of git pull (update) or uploads archive (backup)
  --volumes                  With uninstall only: remove Docker volumes after typed confirmation
```

### 4.2 Cloud HTTPS Flow (`docker-compose.cloud.yml` & `Caddyfile`)

- Activated when `--mode cloud` and a valid non-IP `DOMAIN` is configured:
  1. Optional DNS A-record check (`dig`/`Resolve-DnsName`/socket lookup) comparing `DOMAIN` against host public IP (warnings allow abort/retry unless `--skip-dns-check`).
  2. Render `Caddyfile` dynamically from `Caddyfile.template`:
     - Include global `{ email <ACME_EMAIL> }` block only when `ACME_EMAIL` is non-empty.
     - Site block `<DOMAIN> { encode zstd gzip; reverse_proxy frontend:80 }` (plus safe security headers).
  3. Set in `.env`: `APP_ENVIRONMENT=cloud`, `FRONTEND_URL=https://<DOMAIN>`, `DOMAIN=<DOMAIN>`, `COMPOSE_PATH_SEPARATOR=:`, `COMPOSE_FILE=docker-compose.yml:docker-compose.cloud.yml`.
  4. `docker-compose.cloud.yml` overrides `frontend.ports: !reset []` and launches `caddy:2-alpine` (`cyber_sim_caddy`) on host ports `80:80` and `443:443` attached to `frontend_net`.

### 4.3 Non-Polluting Smoke Verification Contract

1. Verify `db` healthy, `seed` exit code `0`, `backend` healthy (`GET http://127.0.0.1:8080/api/health` returns `status=healthy` & `db_status=up`), and `frontend` / `caddy` serving HTTP `200` (with `--insecure`/DNS-pending warning if local cert still issuing).
2. Read-only seed assertions via `docker compose exec -T db psql`:
   - `security_domains == 5`
   - `questions == 540`
   - `theory_subtopics == 28`
   - `theory_translations >= 33`
   - `scenarios >= 15`
3. Verify Admin authentication (`POST /api/login` using `.env` `ADMIN_EMAIL` / `ADMIN_PASSWORD`) succeeds without creating synthetic users/exam/simulation rows.
4. Print concise redacted status summary (`Redacted Secrets: OK`, active `FRONTEND_URL`, `LLM` mode `Live AI` vs `Static Fallback`, Swagger state).
