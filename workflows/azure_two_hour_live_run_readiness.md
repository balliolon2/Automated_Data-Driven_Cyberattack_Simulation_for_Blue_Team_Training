# Workflow Spec: Azure 60-User Concurrency Hardening & Dec 1 Live 2-Hour Run Readiness (Issue #74)

**GitHub Issue:** [#74 — chore(release): end-to-end 2-hour live-run readiness, timed dry-run, and presentation checklist (Dec 1)](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/74)  
**Status:** `/loop-me` Grilling Rounds 1–3 Settled (`Q1`, `Q6`, `Q13`–`Q15` Resolved, 2026-11-27) — Ready for Implementation  
**Blocked By:** `#70` (`compact_learning_and_two_hour_budget.md`), `#71` (`bilingual_exam_questions_th.md`), `#72` (`highschool_beginner_onboarding_hook.md`), `#73` (`questionnaire_visualizations.md`).  
**Target Hard Deadline:** **Tuesday, December 1, 2026 — Afternoon** (Azure Cloud VM + Caddy Auto-HTTPS, **≤ 60 concurrent high school learners**).

---

## Objective

Harden the platform runtime, Docker Compose configuration, and operational verification so that up to **60 concurrent Thai high school learners** on an **Azure Cloud VM with Caddy HTTPS** can smoothly register, take the bilingual Pre-Test, complete the 5 Concise Domain Essentials, play the Adaptive Scenario Loop (with live LLM API + instant concurrency overflow fallback), complete the bilingual Post-Test, submit the Questionnaire, and inspect the live Visualizations — all within **≤ 120 minutes** with zero 5xx crashes, zero DB connection exhaustion, and zero LLM queue stalls.

---

## 1. Loop Profile

- **Loop Name**: Azure 60-Learner Concurrency Hardening & Pre-Flight Release Gate
- **Trigger**:
  - Runtime: Continuous protection during the live 60-learner classroom session.
  - Pre-Flight Verification: Executed once all blockers (`#70`–`#73`) merge, prior to Tuesday Dec 1 morning deployment on Azure.
- **Checkpoint**:
  - **Checkpoint 2 (Pre-Flight Visual, Concurrency & Timed Dry-Run Sign-Off — `Q13`)**: Present a single decision-ready Release Brief containing:
    1. EN & TH screenshots across all upgraded flows (`#70`–`#73`),
    2. Green results across `go test ./...`, `cd frontend && npm test && npm run tsc && npm run build`, `python scripts/translate_questions_th.py --validate-only`, `python scripts/load_theory_to_db.py --validate-only`, Swagger sync, and `python scripts/test_integration.py`,
    3. 60-concurrent-user load/burst verification results (zero 5xx, bounded DB pool, verified LLM concurrency shedding),
    4. End-to-end timed dry-run confirmation (`<= 120 min`), and
    5. Ready-to-execute Azure deployment & clean-slate commands from `docs/HIGHSCHOOL_2HR_SESSION_RUNBOOK.md`.

---

## 2. Settled Design Decisions

| Q# | Decision | Settled Rule |
| --- | --- | --- |
| `Q6` | Hosting & Concurrency Target | **Azure Cloud VM + Caddy Auto-HTTPS** (`install.sh --mode cloud --domain <azure-domain>`), supporting up to **60 concurrent students** with `LLM_API_KEY` active. |
| `Q14` | PostgreSQL Connection Pooling | Configure explicit `sql.DB` pool bounds in `main.go`: `SetMaxOpenConns(50)` (env `DB_MAX_OPEN_CONNS`, default `50`), `SetMaxIdleConns(20)` (env `DB_MAX_IDLE_CONNS`, default `20`), `SetConnMaxLifetime(5 * time.Minute)` so 60 simultaneous users never exhaust Postgres's 100-connection ceiling. |
| `Q14` | LLM Concurrency Cap & Instant Fallback | Add non-blocking concurrency semaphore (`SIMULATION_LLM_MAX_CONCURRENCY`, default `8`) around runtime `GenerateLLMScenario` in `controllers/simulation_controller.go` / `services/llm_service.go`. If all `8` slots are occupied during a classroom surge, extra callers skip the LLM call immediately and fall back cleanly to the 15 bilingual static scenarios with `fallback_reason = "llm_concurrency_shed"` — zero 20–30s stalls and zero `429` cascades. |
| `Q14` | Compose Limits & Env Parity | Wire `MAX_SCENARIOS_CAP` (default `15`), `SIMULATION_LLM_MAX_CONCURRENCY` (default `8`), `DB_MAX_OPEN_CONNS` (default `50`), `DB_MAX_IDLE_CONNS` (default `20`), `VITE_DEFAULT_LOCALE` (default `th`), and overridable container limits (`BACKEND_CPU_LIMIT:-2.0`, `BACKEND_MEM_LIMIT:-1g`, `DB_CPU_LIMIT:-2.0`, `DB_MEM_LIMIT:-2g`) in `docker-compose.yml` and `.env.example`. |
| `Q15` | Facilitator & Azure Pre-Flight Runbook | Include Azure VM NSG (`80/443/22`), clean-slate DB check, smoke check, and minute-by-minute 120-minute session runbook in `docs/HIGHSCHOOL_2HR_SESSION_RUNBOOK.md`. |

---

## 3. Technical Hardening Specification

### 3.1 Bounded PostgreSQL Pool (`main.go`)
After `gorm.Open(...)`, obtain `sqlDB, err := db.DB()` and configure:
- `DB_MAX_OPEN_CONNS`: parsed int (`1..90`, default `50`) → `sqlDB.SetMaxOpenConns(maxOpen)`
- `DB_MAX_IDLE_CONNS`: parsed int (`1..maxOpen`, default `20`) → `sqlDB.SetMaxIdleConns(maxIdle)`
- `sqlDB.SetConnMaxLifetime(5 * time.Minute)`
- `sqlDB.SetConnMaxIdleTime(2 * time.Minute)`
- Log configured pool values at startup (zero secrets logged).

### 3.2 Non-Blocking LLM Concurrency Shedding (`services/llm_service.go`, `controllers/simulation_controller.go`)
- Add package-level helper in `services/llm_service.go`:
  - `ResolveLLMMaxConcurrency() int`: reads `SIMULATION_LLM_MAX_CONCURRENCY` (`1..64`, default `8`; `0` disables live LLM cleanly for pure static fallback mode).
  - `TryAcquireLLMSlot() (release func(), acquired bool)`: uses a lazy-initialized buffered `chan struct{}` semaphore with non-blocking `select { case sem <- struct{}{}: ... default: return nil, false }`.
- In `controllers/simulation_controller.go` (`StartSimulation`):
  - Before invoking `services.GenerateLLMScenario(...)`, call `release, ok := services.TryAcquireLLMSlot()`.
  - If `!ok`: set `fallbackReason = "llm_concurrency_shed"` and jump immediately into the existing 4-tier static bilingual scenario fallback selector (0 ms wait!).
  - If `ok`: `defer release()` and proceed with normal LLM generation + fallback on error.
- Unit-test `TryAcquireLLMSlot` concurrency cap, release, and `llm_concurrency_shed` fallback behavior in `services/llm_service_test.go` (race-detector clean with `go test -race`).

### 3.3 `docker-compose.yml` & `.env.example` Updates
- **`docker-compose.yml`**:
  - `db.deploy.resources.limits`: `cpus: "${DB_CPU_LIMIT:-2.0}"`, `memory: "${DB_MEM_LIMIT:-2g}"`.
  - `backend.environment`: add `MAX_SCENARIOS_CAP: ${MAX_SCENARIOS_CAP:-15}`, `SIMULATION_LLM_MAX_CONCURRENCY: ${SIMULATION_LLM_MAX_CONCURRENCY:-8}`, `DB_MAX_OPEN_CONNS: ${DB_MAX_OPEN_CONNS:-50}`, `DB_MAX_IDLE_CONNS: ${DB_MAX_IDLE_CONNS:-20}`.
  - `backend.deploy.resources.limits`: `cpus: "${BACKEND_CPU_LIMIT:-2.0}"`, `memory: "${BACKEND_MEM_LIMIT:-1g}"`.
  - `frontend.build.args`: `VITE_DEFAULT_LOCALE: ${VITE_DEFAULT_LOCALE:-th}`.
- **`.env.example`**:
  - Document `MAX_SCENARIOS_CAP=15`, `SIMULATION_LLM_MAX_CONCURRENCY=8`, `DB_MAX_OPEN_CONNS=50`, `DB_MAX_IDLE_CONNS=20`, `VITE_DEFAULT_LOCALE=th`, and optional container limit overrides (`BACKEND_CPU_LIMIT`, `BACKEND_MEM_LIMIT`, `DB_CPU_LIMIT`, `DB_MEM_LIMIT`).
- Verify `docker compose config` resolves cleanly for BOTH local HTTP and Cloud Caddy TLS overlays (`docker-compose.yml:docker-compose.cloud.yml`)!

### 3.4 60-Concurrent-User Burst Verification & Azure Pre-Flight Checklist
- Verify `scripts/run_nfr_benchmark.py` (or dedicated smoke run at `60` concurrent virtual learners hitting `/api/health`, `/api/exams/sample`, authenticated `/api/learn/domains`, `/api/simulation/status`) completes with **0% HTTP 5xx** and healthy p95 latency.
- Document in `docs/HIGHSCHOOL_2HR_SESSION_RUNBOOK.md`:
  - Recommended Azure VM sizing for 60 concurrent users: **≥ 2 vCPU / 4 GB RAM** (e.g., `Standard_B2ms` or `Standard_D2s_v5`) with Azure NSG allowing inbound TCP `80`, `443`, and restricted `22`.
  - One-command clean Azure install/update using `./install.sh --mode cloud --domain <domain> --acme-email <email>` (from Spec 16).
  - Pre-session clean-slate check (confirming 0 dry-run/test learner accounts pollute real cohort stats before students log in, while retaining `5` domains, `540` bilingual questions, `28` subtopics, `33` Thai theory docs, and `15` scenarios).

---

## 4. Verification & Acceptance Criteria

- [ ] `go test -race ./...` passes, including new DB pool helper tests and `TryAcquireLLMSlot` concurrency-shedding tests.
- [ ] `docker compose config` passes cleanly on both Local and Cloud Caddy TLS profiles with the new env variables and limits.
- [ ] Full regression harness passes: `go build`, `cd frontend && npm test && npm run tsc && npm run build`, `python scripts/translate_questions_th.py --validate-only`, `python scripts/load_theory_to_db.py --validate-only`, Swagger sync, and `python scripts/test_integration.py`.
- [ ] Timed end-to-end novice dry-run verified **≤ 120 minutes** from Landing → Pre-Test → 5 Domain Essentials → Scenario Loop → Post-Test → Questionnaire → Admin Visualizations.
- [ ] 60-user concurrency burst check passes with 0% 5xx errors and clean fallback behavior.
- [ ] `docs/HIGHSCHOOL_2HR_SESSION_RUNBOOK.md` finalized and Checkpoint 2 Release Brief approved before closing `#70`–`#74`.
