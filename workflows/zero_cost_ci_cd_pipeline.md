# Workflow: Zero-Cost CI/CD Pipeline & DevSecOps Automation

## Objective
Establish an enterprise-grade, 100% free CI/CD and DevSecOps deployment pipeline for the SOC Trainer platform using GitHub Actions, Cloudflare Pages, Render Web Service, and Neon PostgreSQL. The pipeline enforces code quality, blue-team security scanning, and functional test harness execution before releasing updates with zero manual intervention and zero infrastructure cost.

---

## 1. Loop Profile

- **Loop Name**: Zero-Cost CI/CD & DevSecOps Release Loop
- **Frequency / Cadence**: 
  - **CI (Verification Loop)**: Event-driven on every Pull Request targeting `main`.
  - **CD (Delivery Loop)**: Event-driven on every Merge/Push to `main`.
  - **Keep-Alive (Liveness Loop)**: Scheduled cron every 14 minutes to prevent Render Free Tier cold-start dormancy.
- **Primary Actors**:
  - **Developer**: Authors code, opens PRs, reviews CI briefs.
  - **CI Engine (GitHub Actions)**: Automated gatekeeper validating linting, security posture, and test suites.
  - **Deployment Targets**: Cloudflare Pages (Frontend), Render (Go Backend API), Neon (Serverless PostgreSQL).

---

## 2. Triggers

- **Trigger A (Pull Request Verification)**:
  - Event: `pull_request` targeting `branch: main` (types: `opened`, `synchronize`, `reopened`).
  - Action: Runs Lint, DevSecOps Security Scan, Service Container Integration Harness, and Multi-stage Build verification. **No deployment occurs.**
- **Trigger B (Production Deployment)**:
  - Event: `push` directly or merged into `branch: main`.
  - Action: Verifies CI gate and immediately triggers CD jobs to deploy Frontend to Cloudflare Pages and Backend to Render via Webhook.
- **Trigger C (Backend Keep-Alive Cron)**:
  - Schedule: `cron: '*/14 * * * *'` (every 14 minutes).
  - Action: Lightweight HTTP GET to `https://<render-subdomain>.onrender.com/api` (or health endpoint) to prevent container hibernation.

---

## 3. Checkpoints & Push-Right Discipline

- **Push Right**: All validation (static analysis, dependency vulnerability scanning, secrets sniffing, Postgres service provisioning, exam seeding, and 30-question zero-overlap integration testing) executes **autonomously** before requesting human attention.
- **Checkpoint**: Pull Request Review / Merge Gate.
  - A developer or reviewer only checks the PR after GitHub Actions reports **All Checks Passed (Green)**.
  - Pull Request cannot be merged if any Security Gate or Integration Test fails.
- **Brief**:
  - GitHub Actions Workflow Summary table showing:
    1. Go & TypeScript Lint status
    2. Gitleaks / Secret Scan status (0 secrets detected)
    3. `govulncheck` & `npm audit` report (0 critical/high CVEs)
    4. Integration test suite result (Exam Pre-Test distribution 4/7/5/8/6 & Zero Overlap verified)
    5. Cloudflare Pages Preview URL (for live UI review on PR)

---

## 4. Pipeline Stages & Architecture

```
[ Developer PR / Push ]
          │
          ├──> 1. Lint & Code Quality (Go Vet, golangci-lint, TypeScript tsc)
          ├──> 2. DevSecOps Security Gate (Gitleaks, govulncheck, npm audit)
          ├──> 3. Database Test Harness (Postgres 16 + pgvector container -> Seed -> test_integration.py)
          └──> 4. Build Verification (Docker build backend, Vite build frontend)
                    │
                    ▼ (Only on Merge to `main`)
          ┌───────────────────────────┐
          │  Continuous Deployment    │
          ├─────────────┬─────────────┤
          │             │             │
          ▼             ▼             ▼
    Cloudflare Pages  Render API   Neon DB
    (Frontend SPA)   (Go Backend) (PostgreSQL)
```

### Stage 1: Lint & Code Quality (`job: lint`)
- **Go Backend**:
  - Environment: `actions/setup-go@v5` (Go 1.23+)
  - Command: `go vet ./...`
- **React Frontend**:
  - Environment: `actions/setup-node@v4` (Node 20+)
  - Command: `cd frontend && npm ci && npm run lint && npx tsc -b`

### Stage 2: DevSecOps Security Gate (`job: security-audit`)
- **Secret Scanning**:
  - Action: `gitleaks/gitleaks-action@v2` (detects accidental commits of JWT secrets, database credentials, or API keys).
- **Go Vulnerability Check**:
  - Command: `go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...`
- **Frontend Dependency Audit**:
  - Command: `cd frontend && npm audit --audit-level=high`

### Stage 3: Integration Test Harness with PostgreSQL (`job: integration-test`)
- **Service Container**:
  - Image: `pgvector/pgvector:pg16`
  - Ports: `5432:5432`
  - Env: `POSTGRES_DB=soc_trainer`, `POSTGRES_USER=postgres`, `POSTGRES_PASSWORD=postgres`
- **Database Init**:
  - Execute `init.sql` schema into service container:
    `PGPASSWORD=postgres psql -h localhost -U postgres -d soc_trainer -f init.sql`
- **Backend Launch**:
  - Run Go backend as background task connected to Postgres:
    `DATABASE_URL="postgres://postgres:postgres@localhost:5432/soc_trainer?sslmode=disable" JWT_SECRET="ci-test-secret" go run main.go &`
- **Python Test Harness**:
  - Install dependencies: `pip install requests psycopg2-binary`
  - Seed database: `python scripts/load_exams_to_db.py`
  - Execute integration test: `python scripts/test_integration.py` (verifies 30 questions, domain distribution, session recovery, and zero overlap).

### Stage 4: Build Verification (`job: build-check`)
- **Backend Docker**: Test Docker build `docker build -t soc-trainer-backend:test .`
- **Frontend Assets**: Test production bundle `cd frontend && npm run build`

### Stage 5: Continuous Deployment (`job: deploy` - on `main` push only)
- **Backend (Render Free Web Service)**:
  - Action: Call Render Deploy Hook URL via `curl`:
    `curl -X POST "${{ secrets.RENDER_DEPLOY_HOOK_URL }}"`
- **Frontend (Cloudflare Pages)**:
  - Option A: Native Cloudflare Git Integration (connects directly to repo `main` branch with root directory `frontend`, build command `npm run build`, output directory `dist`, env `VITE_API_URL`).
  - Option B: GitHub Actions deploy via `cloudflare/wrangler-action@v3` using `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID`.

---

## 5. Required Secrets Checklist (GitHub Repository Secrets)

| Secret Name | Purpose | Where to Obtain (Free) |
| :--- | :--- | :--- |
| `RENDER_DEPLOY_HOOK_URL` | Triggers immediate backend redeploy on Render | Render Dashboard -> Web Service -> Settings -> Deploy Hook |
| `CLOUDFLARE_API_TOKEN` | Deploys frontend build to Cloudflare Pages (if using Wrangler) | Cloudflare Dashboard -> My Profile -> API Tokens |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare Account Identifier | Cloudflare Dashboard -> Workers & Pages |
| `VITE_API_URL` | Production Backend API Base URL | e.g. `https://soc-trainer-api.onrender.com/api` |

---

## 6. Cold-Start Keep-Alive Specification

To mitigate Render's 15-minute inactivity spin-down on free tier:
- Workflow file: `.github/workflows/keep_alive.yml`
- Runs: `cron: '*/14 * * * *'` (or external uptime pinger like `UptimeRobot` free tier)
- Command: `curl -s -f https://<app-name>.onrender.com/api/health || true`
- Benefit: Guarantees snappy response times for learners and evaluators without incurring hosting costs.

---

## 7. Definition of Done

This workflow spec is fully realized when:
1. `.github/workflows/ci.yml` is created and passes cleanly on Pull Requests.
2. `.github/workflows/cd.yml` is created to notify Render and Cloudflare on merge to `main`.
3. `.github/workflows/keep_alive.yml` is active to maintain backend readiness.
4. The integration test harness executes and validates exam logic against a live `pgvector` container inside GitHub Actions.
