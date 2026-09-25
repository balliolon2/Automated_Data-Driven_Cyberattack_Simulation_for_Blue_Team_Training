---
title: "Spec 4: Zero-Cost CI/CD Pipeline & DevSecOps Automation"
labels: ["ready-for-agent"]
status: "ready-for-agent"
---

# Spec 4: Zero-Cost CI/CD Pipeline & DevSecOps Automation

## Problem Statement

Currently, the project lacks an automated continuous integration and continuous delivery (CI/CD) system. Code quality verification, unit/integration testing against PostgreSQL and pgvector, and security audits must be executed manually on developer machines. Without automated pipeline guardrails, regression bugs, broken exam distribution logic, uncommitted environment secrets, and vulnerable dependencies can easily be merged into the codebase. Furthermore, there is no automated, zero-cost mechanism to deliver verified code to a public staging/production environment or to prevent free-tier hosting cold starts.

## Solution

Implement an enterprise-grade, 100% free CI/CD and DevSecOps automation pipeline using GitHub Actions, Cloudflare Pages, Render Web Service, and Neon Serverless PostgreSQL:

1. **Automated CI Quality & Security Gates**: Every Pull Request targeted at `main` triggers automated Go/TypeScript linting, static analysis, secret detection (Gitleaks), dependency vulnerability auditing (`govulncheck`, `npm audit`), and Docker build verification.
2. **Ephemeral Service Container Test Harness**: GitHub Actions spins up an isolated `pgvector/pgvector:pg16` container, executes `init.sql`, seeds exam domains, launches the Go backend, and runs `test_integration.py` to assert exam distribution (4/7/5/8/6) and zero-question-overlap integrity.
3. **Automated Zero-Cost Continuous Delivery (CD)**: On successful merge to `main`, the pipeline triggers backend deployment on Render via a secure deploy hook and deploys frontend production bundles to Cloudflare Pages.
4. **Cold-Start Liveness Worker**: A lightweight scheduled workflow pings the backend service every 14 minutes to prevent Render free-tier hibernation.

---

## User Stories

1. As a software developer, I want my Pull Request to automatically trigger linting checks for Go and TypeScript, so that formatting and syntax errors are caught before code review.
2. As a security engineer, I want secret scanning (Gitleaks) to run on every commit in a PR, so that database passwords, JWT secrets, or cloud tokens are never leaked into the repository.
3. As a blue-team platform maintainer, I want `govulncheck` to scan all Go dependencies in CI, so that known CVE vulnerabilities in backend packages are flagged immediately.
4. As a frontend engineer, I want `npm audit` to run during CI, so that high or critical security vulnerabilities in npm packages block the pipeline.
5. As an exam module maintainer, I want CI to spin up a PostgreSQL service container with pgvector and run `scripts/load_exams_to_db.py`, so that schema validity and seeding integrity are validated on every PR.
6. As a QA engineer, I want CI to launch the Go backend and execute `scripts/test_integration.py`, so that user registration, login, pre-test session generation, and zero question overlap between pre and post tests are verified end-to-end.
7. As a project reviewer, I want GitHub branch protection checks to display a unified green status summary, so that I can merge pull requests with confidence.
8. As a DevOps engineer, I want a merge to `main` to trigger the Render Deploy Hook, so that the production Go backend updates automatically without manual dashboard intervention.
9. As a frontend developer, I want production React builds to deploy automatically to Cloudflare Pages upon merge, so that users always access the latest UI release.
10. As a platform evaluator, I want the backend API on Render to stay warm via an automated 14-minute keep-alive ping, so that first-time requests do not experience 30-second cold-start delays.
11. As a contributor, I want clear CI log outputs and concise failure explanations, so that I can quickly fix build or test failures.
12. As a repository owner, I want the entire pipeline to operate strictly within free tier quotas, so that continuous delivery incurs zero monthly infrastructure expense.

---

## Implementation Decisions

### Pipeline Architecture & Workflow Decomposition
The automation is decomposed into three discrete, single-responsibility GitHub Actions workflows:

1. **`ci.yml` (Continuous Integration & DevSecOps Gate)**:
   - **Triggers**: `pull_request` targeting `main`, `push` targeting `main`.
   - **Job 1 (`lint-and-audit`)**:
     - Go environment setup (`actions/setup-go@v5`) with `go vet ./...` and `govulncheck`.
     - Node environment setup (`actions/setup-node@v4`) with `npm ci`, `npm run lint`, `npx tsc -b`, and `npm audit --audit-level=high`.
     - Secret sniffing using `gitleaks/gitleaks-action@v2`.
   - **Job 2 (`integration-test`)**:
     - Requires `lint-and-audit` to succeed.
     - Runs on `ubuntu-latest` with a `services:` container running `pgvector/pgvector:pg16` on port `5432`.
     - Applies `init.sql` directly to the Postgres container.
     - Starts the Go backend executable in the background (`go run main.go &`) pointing to `DATABASE_URL`.
     - Executes Python test harness (`load_exams_to_db.py` followed by `test_integration.py`).
   - **Job 3 (`build-validation`)**:
     - Requires `lint-and-audit` to succeed.
     - Validates frontend production bundle (`npm run build`).
     - Validates backend Docker build (`docker build -t soc-trainer-api:check .`).

2. **`cd.yml` (Continuous Delivery)**:
   - **Triggers**: `push` on branch `main` (after CI completion).
   - **Backend Release**: Sends an authenticated HTTP POST request to Render Deploy Hook (`RENDER_DEPLOY_HOOK_URL`).
   - **Frontend Release**: Employs Cloudflare Pages deploy action or triggers Cloudflare Git integration targeting `dist/`.

3. **`keep_alive.yml` (Dormancy Prevention)**:
   - **Triggers**: `schedule` cron expression (`*/14 * * * *`).
   - **Action**: Lightweight `curl` request to the backend health endpoint (`GET /api` or `GET /api/health`).

### Environment Secrets & Variables
The pipeline relies on the following repository secrets stored in GitHub Settings:
- `RENDER_DEPLOY_HOOK_URL`: Render Web Service deploy trigger URL.
- `CLOUDFLARE_API_TOKEN`: Cloudflare Pages deployment token (if deploying via action).
- `CLOUDFLARE_ACCOUNT_ID`: Cloudflare account identifier.
- `VITE_API_URL`: Production backend URL injected into frontend environment configuration.

---

## Testing Decisions

### Seams & Verification
- **High-Level Test Seam**: GitHub Actions execution environment. The pipeline verifies itself by executing the complete workflow on a pull request.
- **Functional Integration Seam**: `scripts/test_integration.py`. This script interacts exclusively through external HTTP API boundaries (`/api/register`, `/api/login`, `/api/exams/pre-test`, `/api/exams/session`, `/api/exams/post-test`) and verifies:
  1. User authentication flows.
  2. Exactly 30 questions distributed across the 5 domains in required proportions (4/7/5/8/6).
  3. Session state recovery on simulated drops.
  4. Zero question overlap between pre-test and post-test sessions.
- **Security Seam**: CLI return codes of Gitleaks and `govulncheck`. Non-zero exit codes block downstream jobs.

---

## Out of Scope

- Paid Kubernetes cluster orchestration or enterprise cloud infrastructure (AWS EKS, GCP GKE).
- Blue-green zero-downtime deployment (Render free tier performs rolling container replacement with minor spin-up latency).
- Automated database schema rollback/down-migrations (migrations are additive through `init.sql`).
- Self-hosted GitHub Actions runners (all jobs run on standard GitHub-hosted `ubuntu-latest` free runners).

---

## Further Notes

- **Quotas & Resource Management**: Standard GitHub Actions free tier provides 2,000 Linux runner minutes/month for private repositories and unlimited for public repositories. The combined execution time of this optimized workflow is estimated at under 2.5 minutes per run.
- **Zero Configuration Drift**: The `init.sql` and `Dockerfile` currently in the repository root are utilized directly by CI without requiring secondary test configurations.
