# Workflow Spec: System-Level Non-Functional Requirement (NFR) Evaluation Portal (Issue #15)

**GitHub Issue:** [#15 — do the non-functional requirement evaluation](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/15)  
**Status:** `/loop-me` Grilling Rounds 1–3 Settled (`Q1`–`Q16` Resolved, 2026-11-25) — Ready for User Confirmation & Implementation  
**Depends On:**
- [post_test_questionnaire.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/post_test_questionnaire.md) (`AdminPortalTabs`, `evaluation_metrics` CSUQ/NASA-TLX live usability bridge, `@media print` & IEEEtran Paper Export conventions)
- [language_switching_and_localization.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/language_switching_and_localization.md) (Bilingual `en`/`th` UI chrome, `EN | TH` toggle, `i18n/en.ts`, `i18n/th.ts`, `i18n/format.ts`, English-Only Realism Invariant)
- [ui_typography_readability.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/ui_typography_readability.md) (`rem`-only Semantic Typography Scale, `12px` `text-xs` hard floor, `text-log` monospace token, single-row Navbar at `>= 1280px`, `frontend/tests/typography.test.ts`)
- [docs/ADVISOR_PRESENTATION_PROGRESS_TASK.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/docs/ADVISOR_PRESENTATION_PROGRESS_TASK.md) (Section 3.2 10-Dimension NFR Matrix, Section 5 Security Review, Section 6 AI Model Criteria)
- [docs/SCENARIO_GENERATION_AND_FALLBACK_CONTRACT.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/docs/SCENARIO_GENERATION_AND_FALLBACK_CONTRACT.md) & [docs/FAILURE_AND_SESSION_STATUS_CONTRACT.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/docs/FAILURE_AND_SESSION_STATUS_CONTRACT.md) (RQ4 fallback reliability & zero `system_failure` semantics)
- [docs/adr/0005-openapi-code-first-with-dedicated-dtos.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/docs/adr/0005-openapi-code-first-with-dedicated-dtos.md) (Dedicated `dto/` structs and Swag OpenAPI generation)

**Scope Boundary with Issue #13:**
- **Issue #13 (completed spec)**: Subjective learner evaluation (Post-Test CSUQ v3 + Raw NASA-TLX + Open-Ended Impression Feedback, `/admin/evaluations`, `/admin/evaluations/statistics`).
- **Issue #15 (this spec)**: Admin-only, objective **System-Level Non-Functional Requirement (NFR) Evaluation** aligned with **ISO/IEC 25010** (`[R10]`) and the 10-attribute engineering matrix in `docs/ADVISOR_PRESENTATION_PROGRESS_TASK.md`, consuming live runtime telemetry, PostgreSQL integrity/reproducibility/safety/usability probes, safe in-app concurrency probes, and reproducible CLI/CI benchmark snapshots.

---

## Objective

Give platform **Administrators (exclusively)** a live, honest, reproducible, and publication-ready **Non-Functional Requirement (NFR) Evaluation Workbench** at `/admin/nfr` (4th tab in `AdminPortalTabs`) backed by a 3-layer real measurement architecture:

1. **Layer 1 — Passive Runtime Telemetry (Zero-DB-Write Overhead)**:
   - Lightweight Gin HTTP telemetry middleware (`middlewares/telemetry_middleware.go` + `services/nfr_telemetry.go`) capturing latency (`p50`, `p95`, `p99`, `mean`, `max`), throughput, and status class counts (`2xx`/`4xx`/`5xx`) inside a bounded lock-protected in-memory ring buffer (`10,000` samples), normalized by Gin `c.FullPath()`, separating `non_llm_api` routes from `POST /api/simulation/start` (`llm_simulation_start`), storing zero PII/bodies/IPs.
   - LLM generation telemetry hook in `services/llm_service.go` recording real call duration, retry attempts, provider/model, and outcome (`ai_success`, `fallback`, `error`).
   - Upgraded `GET /api/health` performing an actual `SELECT 1` PostgreSQL ping (2-second timeout) and returning `db_status`, `db_latency_ms`, and `uptime_seconds` while staying 100% compatible with existing Docker/CI `"healthy"` checks.
2. **Layer 2 — Live DB, Security Posture & Content-Safety Audit Probes**:
   - Deterministic pure-Go evaluator (`services/nfr_evaluator.go` + `services/nfr_probes.go`) auditing all **10 ISO/IEC 25010 dimensions** (`Performance`, `Availability`, `Scalability`, `Reliability`, `Security`, `Data Integrity`, `Reproducibility`, `Usability`, `Maintainability`, `Safety`) against explicit numeric thresholds (`nfr-rubric-v1`), strictly distinguishing `not_measured` / `insufficient_data` from `fail`.
   - Read-only in-process `httptest` Security Probe verifying Auth (`401`), RBAC (`403` for `learner`/`specialist` on `/api/admin/*`), and safe malformed/SQLi input handling without touching DB records or polluting live user telemetry.
   - Live Usability bridge pulling real CSUQ v3 + Raw NASA-TLX cohort means from `evaluation_metrics` (Issue #13) alongside objective Exam/Simulation completion rates and median completion times.
3. **Layer 3 — Reproducible Benchmarks, Immutable History & Paper-Ready Export**:
   - Safe in-app Concurrency Probe (`POST /api/admin/nfr/probe/scalability`, capped at `20` workers / `200` read-only requests / `15s` timeout, mutex-guarded, telemetry-isolated) for instant live advisor demos.
   - Reproducible external CLI runner (`scripts/run_nfr_benchmark.py`) executing multi-level HTTP load curves (`10 / 25 / 50` concurrent users), external security tests, and automated CI/test/coverage gate collection (`go test -cover ./...`, `npm test`, `tsc -b`, evidence/theory validators, optional `govulncheck`/`npm audit`), supporting both `--push` (`POST /api/admin/nfr/runs`) and `--out report.json` + UI **"Import Benchmark / CI JSON"** upload.
   - Persistent `nfr_evaluation_runs` PostgreSQL table (`init.sql` + idempotent `models/migrate.go`) enabling Run History selection, side-by-side **Baseline Delta Comparison (`Run B vs Run A`)**, per-run CSV/JSON export, and one-click **Copy LaTeX (`IEEEtran + booktabs`)**, **Copy Markdown**, **Auto-Generated Engineering NFR Paragraph (`EN` default / `TH` thesis toggle)**, collapsible **ISO/IEC 25010 `[R10]` Methodology & Formula Guide**, and clean `@media print` PDF layout.

---

## 1. Loop Profile

- **Loop Name**: System-Level Non-Functional Requirement (NFR) Evaluation & Reproducible Benchmark Loop
- **Frequency / Cadence**:
  - Passive Runtime: continuous in-memory capture on every API request and LLM scenario generation call.
  - Admin On-Demand: whenever an Admin visits `/admin/nfr`, clicks **Refresh Live**, **Reset Telemetry Window**, **Run Safe Concurrency Probe**, **Capture Snapshot**, **Import Benchmark JSON**, compares runs, or exports Paper/Thesis artifacts.
  - Benchmark / Pre-Defense / Pre-Release: running `scripts/run_nfr_benchmark.py` against local Docker or staging/production environments to record immutable `nfr_evaluation_runs`.
  - CI Verification: executed on every `go test ./...`, `cd frontend && npm test`, `npm run build`, OpenAPI docs sync, and `scripts/test_integration.py` run.
- **Primary Actors**:
  - **Database Schema & Model (`init.sql`, `models/migrate.go`, `models/nfr_evaluation_run.go`)**: Adds the single `nfr_evaluation_runs` table idempotently.
  - **Runtime Telemetry Collector & Middleware (`services/nfr_telemetry.go`, `middlewares/telemetry_middleware.go`, `services/llm_service.go`, `main.go`)**: Bounded ring buffer, LLM call recorder, probe-isolation flag, and DB-backed `/api/health`.
  - **NFR Evaluation Engine & Probes (`services/nfr_thresholds.go`, `services/nfr_probes.go`, `services/nfr_evaluator.go` + unit tests)**: All 10 ISO/IEC 25010 evaluators, overall verdict rollup, safe concurrency probe, read-only security `httptest` probe, DB integrity/reproducibility/safety queries, usability bridge, and zero-secret exposure guard.
  - **Admin NFR Controller & DTOs (`dto/nfr.go`, `controllers/nfr_controller.go`, `main.go`)**: Admin-only `/api/admin/nfr/*` endpoints (`RequireAdmin()`), Swagger annotations, and UTF-8 BOM CSV / JSON export.
  - **CLI Benchmark & CI Collector (`scripts/run_nfr_benchmark.py`)**: Multi-level concurrency load runner, external security probe, and CI/test/coverage aggregator emitting schema-valid `nfr_evaluation_runs` payloads.
  - **Admin NFR Workbench UI (`AdminPortalTabs.tsx`, `App.tsx`, `frontend/src/pages/AdminNfrPage.tsx`, `frontend/src/lib/nfrReport.ts`, `frontend/src/i18n/{en,th,types,format}.ts`)**: 4th Admin Portal tab, 6-section print-ready single-page report, 3 Recharts visualizations, Run History & Baseline Delta comparison, JSON import, and Paper-Ready LaTeX/Markdown/Narrative generators.
  - **Academic Provenance (`docs/REFERENCES.md`, `docs/references.bib`)**: Registers permanent reference **`R10`** (ISO/IEC 25010:2023 & ISO/IEC 25010:2011).
- **Trigger**:
  - Continuous passive capture via Gin middleware / LLM wrapper; event-driven on Admin UI actions (`/admin/nfr`), CLI benchmark runs, and CI pipeline executions.
- **Checkpoint**:
  - **Runtime**: None (100% autonomous).
  - **Implementation Sign-Off (1 Push-Right Visual Review Checkpoint — `Q12=A`)**: After all Go unit tests, Node frontend tests (`nfrReport.test.ts`, `i18n.test.ts`, `typography.test.ts`), OpenAPI Swagger sync, `scripts/test_integration.py`, and `scripts/run_nfr_benchmark.py` pass, present a single compact **Visual Review Brief** containing EN & TH screenshots of `/admin/nfr` showing:
    1. Top Control Bar, Executive KPI Strip, and 10-Dimension ISO/IEC 25010 Scorecard Grid (`Pass`/`Warn`/`Fail`/`Pending`)
    2. Telemetry & Reliability Recharts visualizations (Endpoint Latency `p50/p95/p99`, Scalability Concurrency Curve, AI vs Static Fallback Reliability)
    3. Run History & Baseline Delta Comparison view (with captured snapshot + concurrency probe / imported benchmark run)
    4. Paper-Ready Engineering Export Panel (LaTeX/Markdown preview, EN/TH auto-generated paragraph, collapsible `[R10]` Methodology Guide, and `@media print` preview)
    Wait for user approval before closing Issue #15.

---

## 2. Settled Design Decisions (`Q1`–`Q16`)

| Q# | Decision | Settled Rule |
| --- | --- | --- |
| `Q1` | NFR Taxonomy & Dimensions | **Full 10-Dimension ISO/IEC 25010 + Advisor Matrix (`1A`)**: `Performance`, `Availability`, `Scalability`, `Reliability`, `Security`, `Data Integrity`, `Reproducibility`, `Usability` (live bridge from #13 + objective task completion/time), `Maintainability`, and `Safety`; register ISO/IEC 25010 as `[R10]`. |
| `Q2` | Measurement Architecture | **3-Layer Hybrid Real Measurement (`2A`)**: Passive HTTP/LLM runtime telemetry + live DB/posture/security probes + reproducible concurrency/CI benchmark snapshots; zero mocked or hardcoded pass claims. |
| `Q3` | Storage & Persistence | **In-Memory Ring Buffer + 1 New PostgreSQL Table `nfr_evaluation_runs` (`3A`)**: Zero DB write overhead per HTTP request (Neon/Render friendly), immutable persisted run snapshots via `init.sql` + idempotent `models/migrate.go`. |
| `Q4` | Navigation & RBAC | **4th Tab in `AdminPortalTabs` (`/admin/nfr`) + Strict `RequireAdmin()` (`4A`)**: Zero new Top Navbar items (`>= 1280px` single-row preserved); all endpoints under `/api/admin/nfr/*` strictly return `403` for `learner`/`specialist` and `401` for unauthenticated callers. |
| `Q5` | Paper-Ready Reporting | **Operational NFR Scorecard + Complete Paper Export Suite (`5A`)**: `Download JSON / CSV (UTF-8 BOM)`, `Copy LaTeX (IEEEtran + booktabs)`, `Copy Markdown`, Auto-Generated NFR Evaluation Paragraph (`EN` default / `TH` thesis toggle), Collapsible `[R10]` Methodology Guide, and `@media print` PDF stylesheet. |
| `Q6` | Localization | **Bilingual UI Chrome (`EN`/`TH`) + English Technical Identifiers (`6A`)**: Labels, descriptions, dimension explanations, and status text follow active `EN | TH` locale; routes, symbols (`p50/p95/p99`, `RPS`, `ISO/IEC 25010`) stay English; Paper Copy defaults to Academic EN with TH thesis toggle. |
| `Q7` | Scalability Execution | **Dual-Mode (`7A`)**: Safe In-App Concurrency Probe button (`POST /api/admin/nfr/probe/scalability`) for live demos AND external CLI `scripts/run_nfr_benchmark.py` (`10 / 25 / 50` concurrent users pushing to `POST /api/admin/nfr/runs`) for publication benchmarks. |
| `Q8` | Security, Safety & Maintainability Sources | **Live In-Process Probes + Ingested CI/Test Gate Report (`8A`)**: Security (Runtime Posture + read-only `httptest` Auth/RBAC/Input probe + optional CVE scans), Safety (100% live DB content/secret/English-only/evidence bounds checks), Maintainability (real test/coverage/gate metrics ingested via CLI or JSON upload; honestly `not_measured` until ingested). |
| `Q9` | Numeric Acceptance Thresholds | **Canonical `nfr-rubric-v1` + Honest `not_measured` / `insufficient_data` Separation (`9A`)**: Explicit numeric Pass/Warn/Fail cutoffs for all 10 dimensions; optional env overrides (`NFR_P95_TARGET_MS`, `NFR_AVAILABILITY_TARGET_PCT`, etc.); unmeasured/zero-sample items never masquerade as `fail` or `pass`. |
| `Q10` | Runtime Telemetry & `/api/health` | **Bounded Ring Buffer (`10,000` samples) + LLM Recorder + Real DB-Ping `/api/health` (`10A`)**: Route normalization via `c.FullPath()`, separate `non_llm_api` vs `llm_simulation_start`, zero PII, and `GET /api/health` performing `SELECT 1` (`2s` timeout) returning `200 healthy` / `503 degraded`. |
| `Q11` | Run History & Comparison | **Full Run Lifecycle (`11A`)**: Live view on load, one-click Capture Snapshot (`run_label`/`environment`), Run History table, side-by-side **Compare 2 Runs (Delta)**, export/LaTeX/Print from live OR any selected `run_id`, and `DELETE /api/admin/nfr/runs/:runId` for cleanup. |
| `Q12` | Provenance, Hygiene & Checkpoint | **`[R10]` ISO/IEC 25010 + OpenAPI DTO Sync + Zero-Secret Exposure + 1 Visual Review Checkpoint (`12A`)**: Never leak raw secret/DSN/API key values; update Swagger docs (`ADR-0005`); present EN/TH visual brief after all tests pass before closing #15. |
| `Q13` | Overall Verdict Rollup | **Transparent Separate Rollup (`13A`)**: Displays `X Pass / Y Warn / Z Fail / W Pending`; Overall Verdict priority `fail` → `partial_pending` → `pass_with_warnings` → `all_passed`; `measured_pass_rate = passed / measured` computed exclusively over measured dimensions. |
| `Q14` | Probe Safety, Isolation & Reset | **Hard Caps + Read-Only Probes + Telemetry Isolation + Window Reset (`14A`)**: Concurrency presets (`5/50`, `10/100`, `20/200`) hard-capped at `20` workers / `200` reqs / `15s` + mutex (`409`); read-only `httptest` security probe; internal probe isolation header excluded from live passive telemetry; `POST /api/admin/nfr/telemetry/reset` resets in-memory window only. |
| `Q15` | Benchmark Import & CI Wiring | **CLI `--push` + `--out` + UI `"Import Benchmark / CI JSON"` (`<= 1 MB`) + CI Gate (`15A`)**: Supports both live push and offline JSON import; `.github/workflows/ci.yml` executes Go/Frontend NFR unit tests and `scripts/test_integration.py` NFR assertions automatically. |
| `Q16` | Page Layout & Recharts | **Structured Single-Page Print-Friendly Report (`16A`)**: 6 ordered sections (Control & Run Bar → Executive KPI Strip → 10-Dimension ISO/IEC 25010 Scorecard Grid → 3 Recharts Charts → Run History & Baseline Delta Table → Paper-Ready Export & `[R10]` Methodology Panel). |

---

## 3. Database & Data Model Contract

Add **one single table** (`nfr_evaluation_runs`) to `init.sql` (Section `1.7 Add analytics tables`, directly after `help_click_analytics`) AND mirror it idempotently in `models/migrate.go` (`ddl` slice) so existing PostgreSQL volumes upgrade cleanly on startup without destructive resets.

### 3.1 PostgreSQL DDL (`init.sql` & `models/migrate.go`)

```sql
CREATE TABLE IF NOT EXISTS nfr_evaluation_runs (
  run_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  triggered_by UUID REFERENCES users(user_id) ON DELETE SET NULL,
  run_label VARCHAR(120) NOT NULL DEFAULT 'Live NFR Snapshot',
  environment VARCHAR(64) NOT NULL DEFAULT 'local',
  source VARCHAR(32) NOT NULL CHECK (source IN ('live_snapshot', 'in_app_probe', 'cli_benchmark', 'imported_json')),
  rubric_version VARCHAR(32) NOT NULL DEFAULT 'nfr-rubric-v1',
  overall_status VARCHAR(32) NOT NULL CHECK (overall_status IN ('all_passed', 'pass_with_warnings', 'partial_pending', 'fail')),
  passed_count INT NOT NULL DEFAULT 0,
  warn_count INT NOT NULL DEFAULT 0,
  fail_count INT NOT NULL DEFAULT 0,
  not_measured_count INT NOT NULL DEFAULT 0,
  measured_pass_rate FLOAT NOT NULL DEFAULT 0,
  non_llm_p95_ms FLOAT,
  availability_pct FLOAT,
  report_payload JSONB NOT NULL,
  notes TEXT,
  created_at TIMESTAMP DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_nfr_evaluation_runs_created_at ON nfr_evaluation_runs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_nfr_evaluation_runs_source ON nfr_evaluation_runs(source);
```

### 3.2 Go GORM Model (`models/nfr_evaluation_run.go`)

```go
package models

import "time"

type NFREvaluationRun struct {
	RunID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"run_id"`
	TriggeredBy      *string   `gorm:"type:uuid" json:"triggered_by,omitempty"`
	RunLabel         string    `gorm:"type:varchar(120);not null;default:'Live NFR Snapshot'" json:"run_label"`
	Environment      string    `gorm:"type:varchar(64);not null;default:'local'" json:"environment"`
	Source           string    `gorm:"type:varchar(32);not null" json:"source"` // live_snapshot | in_app_probe | cli_benchmark | imported_json
	RubricVersion    string    `gorm:"type:varchar(32);not null;default:'nfr-rubric-v1'" json:"rubric_version"`
	OverallStatus    string    `gorm:"type:varchar(32);not null" json:"overall_status"` // all_passed | pass_with_warnings | partial_pending | fail
	PassedCount      int       `gorm:"not null;default:0" json:"passed_count"`
	WarnCount        int       `gorm:"not null;default:0" json:"warn_count"`
	FailCount        int       `gorm:"not null;default:0" json:"fail_count"`
	NotMeasuredCount int       `gorm:"not null;default:0" json:"not_measured_count"`
	MeasuredPassRate float64   `gorm:"not null;default:0" json:"measured_pass_rate"`
	NonLLMP95Ms      *float64  `gorm:"column:non_llm_p95_ms" json:"non_llm_p95_ms,omitempty"`
	AvailabilityPct  *float64  `gorm:"column:availability_pct" json:"availability_pct,omitempty"`
	ReportPayload    any       `gorm:"column:report_payload;type:jsonb;serializer:json;not null" json:"report_payload"`
	Notes            string    `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time `gorm:"default:now()" json:"created_at"`
}

func (NFREvaluationRun) TableName() string { return "nfr_evaluation_runs" }
```

### 3.3 Persistence Invariants
- `report_payload` stores the complete canonical `dto.NFRReportResponse` JSON snapshot at capture/ingest time, making every historical run immutable and self-contained for Baseline Delta comparison and Paper Export even if code or thresholds change later.
- Zero per-request database writes occur during normal user API traffic.
- `report_payload` MUST NEVER contain `JWT_SECRET`, `LLM_API_KEY`, `DATABASE_URL`, `POSTGRES_PASSWORD`, bearer tokens, or user emails/passwords.

---

## 4. Runtime Telemetry Middleware, LLM Recorder & `/api/health` Upgrade

### 4.1 In-Memory Telemetry Collector (`services/nfr_telemetry.go`)
A concurrency-safe (`sync.RWMutex`) singleton collector initialized at server startup (`services.GlobalNFRTelemetry()`), with clean constructor `NewNFRTelemetryStore(capacity int)` for deterministic unit tests:

- **Process Lifecycle**:
  - `processStartedAt time.Time` (never reset; drives server uptime).
  - `windowStartedAt time.Time` (initialized to `processStartedAt`; updated on `POST /api/admin/nfr/telemetry/reset`).
- **Bounded HTTP Ring Buffer (`capacity = 10,000` samples)**:
  - Each `HTTPRequestSample` stores only: `Route string` (`"METHOD /api/..."` from `c.FullPath()`, falling back to `"<unmatched>"` for 404s so random URLs never cause unbounded map cardinality), `Category string` (`"non_llm_api"` or `"llm_simulation_start"`), `StatusCode int`, `DurationMs float64`, `RecordedAt time.Time`.
  - **Route Classification**:
    - `POST /api/simulation/start` → `Category = "llm_simulation_start"` (excluded from `non_llm_api` latency percentiles so external LLM roundtrips do not distort CRUD API `p95`/`p99`).
    - All other `/api/*` routes except internal probe requests → `Category = "non_llm_api"`.
    - `OPTIONS` preflight requests and requests carrying the internal probe isolation marker are excluded from passive user traffic stats.
  - **Window Aggregates**:
    - Tracks `total_requests`, `status_2xx`, `status_3xx`, `status_4xx`, `status_5xx`, `in_flight_current`, `peak_concurrent_requests`, and per-route latency slices (capped cleanly via ring-buffer window eviction or per-route ring buffer of `1,000` samples so memory stays `< 5 MB`).
  - **Percentile Math (Exact Linear Interpolation — R-7)**:
    - Sort copy of durations in ascending order; compute `p50`, `p95`, `p99`, `mean`, `min`, `max` using the exact R-7 quantile formula `idx = p * (n - 1)` (matching `services/evaluation_stats.go`). Return `nil` when `n == 0` (never `NaN` or fake `0 ms`).
- **LLM Generation Telemetry Ring Buffer (`capacity = 500` samples)**:
  - Recorded directly inside `services/llm_service.go` (`GenerateLLMScenarioWithContext` / `GenerateLLMScenario`) whenever an LLM attempt executes or falls back:
    - `DurationMs float64`, `Attempts int`, `Retries int`, `Outcome string` (`"ai_success"` | `"fallback"` | `"error"`), `FallbackReason string`, `Provider string` (hostname/scheme only, e.g., `api.openai.com` — NEVER API key), `Model string`, `RecordedAt time.Time`.
  - Also stores the **latest in-app scalability probe result** (`lastScalabilityProbe *dto.NFRScalabilityBenchmarkDTO`) in memory so subsequent `GET /api/admin/nfr/live` refreshes show the latest probe run until replaced or reset.

### 4.2 Telemetry Middleware & Probe Isolation (`middlewares/telemetry_middleware.go`)
- Registered on the Gin engine in `main.go` directly after CORS (before routes).
- **Internal Probe Isolation Marker**:
  - Define package-exported constant `InternalNFRProbeHeader = "X-SOC-Internal-NFR-Probe"` paired with a per-process random 32-byte hex secret (`services.InternalProbeToken()`), generated once at startup in memory and NEVER exposed via HTTP.
  - Only in-process `httptest` probes (`services/nfr_probes.go`) set `req.Header.Set(InternalNFRProbeHeader, services.InternalProbeToken())`.
  - Middleware skips recording any request whose header matches `services.InternalProbeToken()`. External callers spoofing the header cannot bypass telemetry!
- Increments `in_flight` atomic counter before `c.Next()` (updating `peak_concurrent_requests`), decrements on defer, measures `time.Since(start)` in fractional milliseconds, and records into `services.GlobalNFRTelemetry()`.

### 4.3 Upgraded `GET /api/health` (`main.go`)
Replace the static JSON handler with a real PostgreSQL connectivity check:
- Obtain `sqlDB, err := db.DB()`; run `sqlDB.PingContext(ctx)` with `2 * time.Second` timeout and measure `dbLatencyMs`.
- If healthy: return HTTP **`200 OK`**:
  ```json
  {
    "status": "healthy",
    "service": "soc-trainer-api",
    "db_status": "up",
    "db_latency_ms": 2.41,
    "uptime_seconds": 1842
  }
  ```
- If DB unreachable: return HTTP **`503 Service Unavailable`**:
  ```json
  {
    "status": "degraded",
    "service": "soc-trainer-api",
    "db_status": "down",
    "db_latency_ms": 2001.12,
    "uptime_seconds": 1842
  }
  ```
- Backward compatibility verified: `docker-compose.yml` (`wget -q -O /dev/null http://127.0.0.1:8080/api/health`) checks HTTP 200, and `.github/workflows/ci.yml` (`grep -q "healthy"`) matches `"status": "healthy"`.

---

## 5. Pure-Go NFR Evaluation Engine (`nfr-rubric-v1`) & Probes

### 5.1 Canonical Threshold Configuration (`services/nfr_thresholds.go`)
Defaults (`RubricVersion = "nfr-rubric-v1"`), overridable via optional environment variables for controlled deployment tuning:

| Env Override (Optional) | Default | Meaning |
| --- | --- | --- |
| `NFR_P95_TARGET_MS` | `300.0` | Non-LLM API `p95` Pass ceiling (`Warn`: `500.0 ms`) |
| `NFR_P99_TARGET_MS` | `800.0` | Non-LLM API `p99` Pass ceiling (`Warn`: `1200.0 ms`) |
| `NFR_LLM_P95_TARGET_MS` | `20000.0` | LLM generation `p95` Pass ceiling (`Warn`: `45000.0 ms`) |
| `NFR_DB_PING_TARGET_MS` | `50.0` | DB Ping Pass ceiling (`Warn`: `150.0 ms`) |
| `NFR_AVAILABILITY_TARGET_PCT` | `99.0` | Non-5xx HTTP availability Pass floor (`Warn`: `95.0%`) |
| `NFR_SCALABILITY_P95_MS` | `500.0` | Concurrency load `p95` Pass ceiling (`Warn`: `1000.0 ms`) |
| `NFR_SCALABILITY_ERR_PCT` | `1.0` | Concurrency load error rate Pass ceiling (`Warn`: `5.0%`) |
| `NFR_COVERAGE_TARGET_PCT` | `60.0` | Go unit test coverage Pass floor (`Warn`: `40.0%`) |

### 5.2 Exact 10-Dimension ISO/IEC 25010 Matrix (`services/nfr_evaluator.go`)

Every dimension produces a structured `dto.NFRDimensionResultDTO`:
- `id` (stable slug), `iso_characteristic` (ISO/IEC 25010 `[R10]` mapping), `status` (`"pass" | "warn" | "fail" | "not_measured" | "insufficient_data"`), `summary_metric`, `target_expression`, `evidence_source`, `checks []dto.NFRCheckItemDTO` (granular sub-checks with measured value, unit, threshold, and status).

| # | Dimension ID | ISO/IEC 25010 Mapping (`[R10]`) | Real Measured Sub-Checks & Sources | Pass / Warn / Fail / Pending Rules |
| --- | --- | --- | --- | --- |
| 1 | `performance` | Performance Efficiency (Time Behaviour) | 1. Non-LLM API `p50/p95/p99` ms (ring buffer)<br>2. DB Ping latency ms (`SELECT 1`)<br>3. LLM Generation `p50/p95` ms (LLM ring buffer — optional if 0 LLM calls in window) | **Pass**: DB ping `<= 50 ms` AND (if `non_llm_n > 0`: `p95 <= 300 ms` & `p99 <= 800 ms`) AND (if `llm_n > 0`: `llm_p95 <= 20,000 ms`).<br>**Warn**: Any metric in warn band (`p95 <= 500`, `p99 <= 1200`, `db <= 150`, `llm_p95 <= 45,000`).<br>**Fail**: Any metric breaches warn ceiling.<br>*(Note: Because `GET /api/admin/nfr/live` itself runs after middleware or DB ping always executes, `db_ping` is always measured; if `non_llm_n == 0` right after reset, sub-check marks `insufficient_data` cleanly.)* |
| 2 | `availability` | Reliability (Availability) | 1. DB Connectivity (`SELECT 1` status `up`/`down`)<br>2. HTTP Non-5xx Success Rate (`(total - status_5xx) / total * 100`)<br>3. Process Uptime (seconds) | **Fail**: DB `down` OR (when `total > 0` and Non-5xx `< 95.0%`).<br>**Warn**: Non-5xx `95.0–98.99%`.<br>**Pass**: DB `up` AND (`total == 0` OR Non-5xx `>= 99.0%`). |
| 3 | `scalability` | Performance Efficiency (Capacity) | Populated from latest **In-App Safe Concurrency Probe** OR **CLI Benchmark Scalability Curve** (`levels[]`: `concurrency`, `total_requests`, `throughput_rps`, `p50_ms`, `p95_ms`, `p99_ms`, `error_rate_pct`) + passive `peak_concurrent_requests`. | **Not Measured**: When no concurrency probe or CLI load benchmark has been attached (`levels` empty).<br>**Pass**: Max tested level has `concurrency >= 10` (or `>= 5` for `light` preset annotated as `Warn: light preset`), `error_rate_pct <= 1.0%`, `p95_ms <= 500 ms`.<br>**Warn**: `error_rate_pct <= 5.0%` OR `p95_ms <= 1000 ms` OR max tested concurrency `< 10`.<br>**Fail**: `error_rate_pct > 5.0%` OR `p95_ms > 1000 ms`. |
| 4 | `reliability` | Reliability (Fault Tolerance & Recoverability) | 1. HTTP `5xx` Error Rate (`status_5xx / total * 100`)<br>2. Simulation `system_failure` Rate (`COUNT(failure_class='system_failure') / COUNT(*)` in `simulation_sessions` — **RQ4**)<br>3. Valid Scenario Delivery Rate (`COUNT(validation_status='valid') / COUNT(*)` in `scenario_snapshots`)<br>4. Generation Breakdown (`ai_generated` vs `static_fallback` counts & fallback reasons — informational, never penalizes fallback itself since fallback IS the reliability mechanism!) | **Pass**: `5xx <= 1.0%` AND `system_failure_rate == 0.0%` AND (`snapshots == 0` OR `valid_delivery == 100.0%`).<br>**Warn**: `5xx <= 5.0%` OR `valid_delivery 95.0–99.99%`.<br>**Fail**: `5xx > 5.0%` OR `system_failure_rate > 0.0%` OR `valid_delivery < 95.0%`. |
| 5 | `security` | Security (Authenticity, Integrity, Confidentiality) | 1. **In-Process Read-Only `httptest` Probe (6 checks)**:<br>   - Unauthenticated `GET /api/admin/nfr/live` → `401`<br>   - Forged/invalid JWT on `GET /api/admin/nfr/live` → `401`<br>   - Valid `learner` JWT on `GET /api/admin/nfr/live` → `403`<br>   - Valid `specialist` JWT on `GET /api/admin/nfr/live` → `403`<br>   - Unauthenticated `GET /api/exams/session` → `401`<br>   - SQLi query param on `GET /api/learn/domains?lang=' OR '1'='1` → non-`500`<br>2. **Runtime Posture Checks (6 boolean checks — Zero Secret Values)**:<br>   - `JWT_SECRET` set, non-default (`!= "supersecretkey"` / `"secret_jwt_key_default"`), length `>= 32`<br>   - `ENABLE_SWAGGER != "true"`<br>   - `BYPASS_POST_TEST_THRESHOLD != "true"`<br>   - CORS allowlist contains no wildcard `*`<br>   - 100% of `users.password_hash` start with bcrypt prefix (`$2a$`, `$2b$`, `$2y$`)<br>   - `GIN_MODE == "release"`<br>3. **Optional Ingested DevSecOps Scan Counts**: `govulncheck_high_critical`, `npm_audit_high_critical`, `gitleaks_findings` (marked `not_measured` if not attached). | **Fail**: Any active `httptest` probe fails (`< 100%`), OR non-bcrypt password exists, OR CORS wildcard `*` exists, OR default/empty `JWT_SECRET` in `release` mode, OR attached High/Critical CVE/secret count `> 0`.<br>**Warn**: Active probes `100%` pass, but non-critical dev posture flags active (`ENABLE_SWAGGER=true`, `BYPASS_POST_TEST_THRESHOLD=true`, `GIN_MODE!=release`, or short dev `JWT_SECRET` when `GIN_MODE!=release`).<br>**Pass**: Active probes `100%` AND Posture `100%` AND (if attached) `0` High/Critical findings. |
| 6 | `data_integrity` | Functional Suitability / Reliability (Data Integrity) | Live SQL queries against PostgreSQL:<br>1. **Pre/Post Zero-Overlap Violations**: count of `(user_id, question_id)` appearing in both a user's `pre` and `post` `exam_session_questions` (target `= 0`)<br>2. **Orphaned Records**: `simulation_sessions` without valid `scenario_id` or `user_id`, or `completed` `exam_sessions` with `score IS NULL` (target `= 0`)<br>3. **Curriculum Seed Integrity**: active `questions` `>= 540`, `theory_subtopics == 28`, `theory_translations` (`locale='th'`) `>= 33`, active `scenarios >= 15` | **Pass**: Overlap `= 0`, Orphans `= 0`, Curriculum Seed Integrity `100%`.<br>**Warn**: Overlap `= 0`, Orphans `= 0`, but partial seed counts (`questions < 540` or `th < 33`).<br>**Fail**: Any Pre/Post overlap `> 0` OR any orphaned/corrupt record `> 0`. |
| 7 | `reproducibility` | Maintainability / Reliability (Auditability & Replayability) | Live SQL audit of `simulation_sessions` left-joined with `scenario_snapshots`:<br>1. **Snapshot Coverage %**: sessions having a linked `scenario_snapshots` row<br>2. **Complete Metadata %**: snapshots having non-empty `source_type`, `prompt_version`, `blueprint_version`, `rubric_version`, and `validation_status` | **Insufficient Data**: When `total_simulation_sessions == 0`.<br>**Pass**: Snapshot coverage `== 100.0%` AND Complete metadata `== 100.0%`.<br>**Warn**: `95.0–99.99%`.<br>**Fail**: `< 95.0%`. |
| 8 | `usability` | Interaction Capability (Usability — Bridge from #13) | Live DB queries from `evaluation_metrics`, `exam_sessions`, and `simulation_sessions`:<br>1. **CSUQ v3 Overall Mean** (`/7.0`, `n` respondents — reusing `services.ComputePaperStatistics` or `evaluation_metrics` query)<br>2. **Raw NASA-TLX Overall Workload** (`/100`)<br>3. **Objective Task Completion Rate & Median Time**: Exam completion rate (`completed / total`), Simulation completion rate (`completed / (completed + abandoned)`), median exam & simulation completion seconds | **Insufficient Data**: When `questionnaire_n == 0` (still displays objective exam/simulation completion metrics if present, without faking pass/fail).<br>**Pass**: `CSUQ Mean >= 5.0` AND `RTLX Mean <= 55.0`.<br>**Warn**: `CSUQ 4.0–4.99` OR `RTLX 55.01–70.0`.<br>**Fail**: `CSUQ < 4.0` OR `RTLX > 70.0`. |
| 9 | `maintainability` | Maintainability (Testability & Modifiability) | Populated from **CLI / CI Gate Report** attached via `POST /api/admin/nfr/runs` or UI JSON import:<br>1. `go_test_passed` (`bool`) & `go_coverage_pct` (`float64`)<br>2. `frontend_test_passed` (`bool`, includes `i18n.test.ts` + `typography.test.ts`) & `frontend_tsc_passed` (`bool`)<br>3. `evidence_gate_passed` (`bool`) & `theory_gate_passed` (`bool`)<br>4. `integration_harness_passed` (`*bool`, optional) | **Not Measured**: When no CI/test gate payload is attached.<br>**Pass**: All executed automated gates pass (`100%`) AND `go_coverage_pct >= 60.0%`.<br>**Warn**: All executed gates pass (`100%`) AND `go_coverage_pct 40.0–59.99%`.<br>**Fail**: Any automated gate fails OR `go_coverage_pct < 40.0%`. |
| 10 | `safety` | Safety / Freedom from Risk (Simulated Training Boundary) | 100% live DB inspection across `questions`, `scenarios`, and `scenario_snapshots`:<br>1. **English-Only Realism Invariant**: `0` rows containing Thai Unicode (`[\u0E00-\u0E7F]`) in `questions.question_text` or `scenarios.initial_logs`<br>2. **Zero Secret / API-Key Leakage**: `0` `scenario_snapshots` rows matching secret patterns (`sk-`, `Bearer `, `ghp_`, `gho_`, `JWT_SECRET`, active `LLM_API_KEY` if configured)<br>3. **Zero Banned Boilerplate**: `0` active `scenarios` containing banned legacy strings (`Power off all domain controllers immediately`, `Reformat all network switches in the subnet`, `Disable MFA enforcement for affected users`, `SIEM Alert Rule [CORR-9901]`)<br>4. **Evidence Ratio Bound**: `100%` of active `scenarios` have evidence union `<= 50%` of `initial_logs` | **Pass**: All 4 safety invariants `100%` satisfied (`0` violations).<br>**Fail**: Any violation `> 0`. |

### 5.3 Overall Verdict Rollup (`Q13=A`)
Given the 10 `NFRDimensionResultDTO` statuses:
- Count `passed_count` (`status == "pass"`), `warn_count` (`status == "warn"`), `fail_count` (`status == "fail"`), `not_measured_count` (`status == "not_measured" || status == "insufficient_data"`).
- `measured_count = passed_count + warn_count + fail_count`.
- `measured_pass_rate`: if `measured_count == 0` → `0.0`, else `round2(float64(passed_count) / float64(measured_count) * 100.0)`.
- `overall_status` precedence:
  1. `fail_count > 0` → `"fail"`
  2. `not_measured_count > 0` → `"partial_pending"`
  3. `warn_count > 0` → `"pass_with_warnings"`
  4. Otherwise → `"all_passed"`

### 5.4 Safe In-App Concurrency Probe (`services/nfr_probes.go` — `Q7=A`, `Q14=A`)
- Protected by a package/controller `atomic.Bool` mutex: if a probe is already executing, immediately return HTTP **`409 Conflict`** with `code: "NFR_PROBE_ALREADY_RUNNING"`.
- Supported presets (default `"standard"`):
  - `"light"`: `concurrency = 5`, `total_requests = 50`
  - `"standard"`: `concurrency = 10`, `total_requests = 100`
  - `"burst"`: `concurrency = 20`, `total_requests = 200`
- **Hard Server-Side Caps**: Any custom values exceeding `concurrency > 20` or `total_requests > 200` are rejected (`400 NFR_INVALID_PROBE_PARAMS`); overall execution context has hard timeout `15 * time.Second`.
- **Safe Read-Only Target Rotation**: Executes in-process `httptest.NewRecorder()` requests directly against the live Gin router (`router.ServeHTTP`) cycling across 3 read-only endpoints:
  1. `GET /api/health`
  2. `GET /api/exams/sample`
  3. `GET /api/learn/domains` (using the requesting Admin's `Authorization` Bearer header)
- Every probe request carries `X-SOC-Internal-NFR-Probe: <InternalProbeToken()>` so passive live telemetry is 100% isolated.
- Measures exact `p50_ms`, `p95_ms`, `p99_ms`, `mean_ms`, `max_ms`, `throughput_rps`, `error_rate_pct`, `duration_ms`, stores it as `lastScalabilityProbe`, and if `save_run: true`, automatically persists a new `nfr_evaluation_runs` row (`source = "in_app_probe"`).

### 5.5 Read-Only In-Process Security Probe (`services/nfr_probes.go` — `Q8=A`, `Q14=A`)
- Runs synchronously in `< 20 ms` during live NFR audit using `httptest.NewRecorder()` against the live Gin router (with internal probe isolation header):
  1. `unauth_admin_nfr_blocked`: `GET /api/admin/nfr/live` (no header) → expects `401`
  2. `forged_jwt_rejected`: `GET /api/admin/nfr/live` with `Bearer invalid.jwt.token` → expects `401`
  3. `learner_rbac_denied`: `GET /api/admin/nfr/live` with synthetic short-lived (`30s`) HMAC JWT `{user_id: "00000000-0000-0000-0000-000000000001", role: "learner"}` → expects `403`
  4. `specialist_rbac_denied`: `GET /api/admin/nfr/live` with synthetic short-lived (`30s`) HMAC JWT `{user_id: "00000000-0000-0000-0000-000000000002", role: "specialist"}` → expects `403`
  5. `unauth_exam_blocked`: `GET /api/exams/session` (no header) → expects `401`
  6. `sqli_query_param_safe`: `GET /api/learn/domains?lang=' OR '1'='1` with Admin token → expects status `< 500` (gracefully normalized to `"en"` or rejected cleanly, zero SQL crash)
- Zero DB writes, zero side effects, and zero impact on passive user telemetry.

---

## 6. Admin-Only REST API Contract (`dto/nfr.go`, `controllers/nfr_controller.go`, `main.go`)

All NFR endpoints are mounted inside the existing strictly Admin-only route group in `main.go`:
```go
admin := api.Group("/admin", middlewares.AuthMiddleware(), middlewares.RequireAdmin())
```
Any unauthenticated request receives `401 AUTH_UNAUTHORIZED`; any `learner` or `specialist` receives `403 AUTH_FORBIDDEN`.

### 6.1 Endpoint Summary

| Method & Path | Purpose | Request / Query | Response |
| --- | --- | --- | --- |
| `GET /api/admin/nfr/live` | Compute real-time 10-dimension NFR evaluation (passive telemetry + live DB/Security/Safety/Usability probes + latest scalability probe if present) | — | `200 OK` (`dto.NFRReportResponse`) |
| `POST /api/admin/nfr/runs` | Persist an immutable NFR snapshot: either capturing current live evaluation (`mode: "live_snapshot"`) OR ingesting a CLI/imported benchmark JSON payload (`source: "cli_benchmark" | "imported_json"`, max body `1 MB`) with optional scalability levels, CI/maintainability results, and CVE scan counts merged over live DB checks | `dto.CreateNFRRunRequest` | `201 Created` (`dto.NFRRunDetailResponse`) |
| `GET /api/admin/nfr/runs` | List historical `nfr_evaluation_runs` ordered by `created_at DESC` | `?limit=50` (default `50`, max `200`) | `200 OK` (`dto.NFRRunListResponse`) |
| `GET /api/admin/nfr/runs/:runId` | Retrieve full saved `dto.NFRReportResponse` for a specific historical run | `:runId` (UUID) | `200 OK` (`dto.NFRRunDetailResponse`) / `404 NFR_RUN_NOT_FOUND` |
| `DELETE /api/admin/nfr/runs/:runId` | Delete a saved run (e.g., cleanup of exploratory test runs) | `:runId` (UUID) | `200 OK` (`{"deleted": true, "run_id": "..."}`) / `404 NFR_RUN_NOT_FOUND` |
| `POST /api/admin/nfr/probe/scalability` | Execute Safe In-App Concurrency Probe (`light` \| `standard` \| `burst`), update live scalability cache, and optionally save a snapshot (`save_run: true`) | `dto.RunScalabilityProbeRequest` | `200 OK` (`dto.ScalabilityProbeResponse`) / `409 NFR_PROBE_ALREADY_RUNNING` / `400 NFR_INVALID_PROBE_PARAMS` |
| `POST /api/admin/nfr/telemetry/reset` | Reset the in-memory HTTP & LLM telemetry ring buffer and `window_started_at` (never deletes DB `nfr_evaluation_runs`) | — | `200 OK` (`dto.TelemetryResetResponse`) |
| `GET /api/admin/nfr/export` | Download live report or a specific saved run as UTF-8 BOM CSV (Excel/SPSS-safe with formula-injection neutralization) or JSON | `?format=csv|json` (default `json`) & `&run_id=<uuid>` (optional) | `200 OK` attachment (`nfr_evaluation_<runId|live>_YYYYMMDD.csv|json`) |

### 6.2 Canonical Response Schema (`dto/nfr.go`)
- `NFRReportResponse`:
  - `run_id *string` (`nil` for live view, UUID for persisted run)
  - `run_label string`, `environment string`, `source string`, `rubric_version string` (`"nfr-rubric-v1"`), `evaluated_at time.Time`
  - `summary NFRSummaryDTO` (`overall_status`, `total_dimensions: 10`, `passed_count`, `warn_count`, `fail_count`, `not_measured_count`, `measured_pass_rate`)
  - `runtime NFRRuntimeSnapshotDTO` (`process_started_at`, `window_started_at`, `uptime_seconds`, `gin_mode`, `total_requests`, `non_llm_requests`, `llm_requests`, `status_2xx`, `status_4xx`, `status_5xx`, `availability_pct`, `error_rate_5xx_pct`, `peak_concurrent_requests`, `db_status`, `db_latency_ms`, `non_llm_latency NFRPercentilesDTO`, `llm_latency NFRPercentilesDTO`, `endpoints []NFREndpointLatencyDTO`)
  - `dimensions []NFRDimensionResultDTO` (fixed canonical order 1–10)
  - `scalability *NFRScalabilityBenchmarkDTO` (`mode`, `executed_at`, `target_endpoints`, `levels []NFRScalabilityLevelDTO`)
  - `reliability_breakdown NFRReliabilityBreakdownDTO` (`total_sessions`, `ai_generated_count`, `static_fallback_count`, `system_failure_count`, `valid_snapshot_count`, `total_snapshot_count`, `fallback_reasons map[string]int`)
  - `maintainability_report *NFRMaintainabilityReportDTO` (`collected_at`, `git_commit`, `go_test_passed`, `go_coverage_pct`, `frontend_test_passed`, `frontend_tsc_passed`, `evidence_gate_passed`, `theory_gate_passed`, `integration_harness_passed`, `govulncheck_high_critical`, `npm_audit_high_critical`, `gitleaks_findings`)
  - `thresholds NFRThresholdsDTO` (explicit numeric targets active during evaluation, so historical runs stay self-documenting)

### 6.3 Payload Hygiene & Error Codes
- `POST /api/admin/nfr/runs` enforces `http.MaxBytesReader(w, r.Body, 1<<20)` (`1 MB`); oversized uploads return `413` (`code: "NFR_PAYLOAD_TOO_LARGE"`).
- Invalid JSON or out-of-range benchmark numbers (`concurrency <= 0`, `error_rate_pct < 0 || > 100`, `coverage < 0 || > 100`, invalid `source`) return `400` (`code: "NFR_INVALID_BENCHMARK_PAYLOAD"`).
- Add bilingual error messages to `frontend/src/i18n/en.ts`, `th.ts`, and `format.ts` for:
  - `NFR_PROBE_ALREADY_RUNNING`
  - `NFR_INVALID_PROBE_PARAMS`
  - `NFR_INVALID_BENCHMARK_PAYLOAD`
  - `NFR_PAYLOAD_TOO_LARGE`
  - `NFR_RUN_NOT_FOUND`

---

## 7. Reproducible CLI Benchmark & CI Gate Aggregator (`scripts/run_nfr_benchmark.py`)

A self-contained Python 3.11+ script (using only stdlib + `requests`, already in `requirements.txt`) designed for controlled experiments, pre-defense benchmarking, and offline JSON export:

- **CLI Arguments**:
  - `--api-url` (default `http://localhost:8080`)
  - `--email` / `--password` (Admin credentials; also reads `ADMIN_EMAIL` / `ADMIN_PASSWORD` from env)
  - `--label` (default `"CLI NFR Benchmark"`)
  - `--environment` (default `"local-docker"`)
  - `--levels` (default `"10,25,50"` — comma-separated concurrent worker counts)
  - `--requests-per-level` (default `"100"`)
  - `--include-ci` (runs local verification gates: `go test -cover ./...`, `npm test`, `npx tsc -b`, `validate_evidence_mapping.py --mapping scripts/evidence_mapping.json --require-complete`, `load_theory_to_db.py --validate-only --require-all-th`, and parses exact Go total coverage % + pass/fail booleans)
  - `--include-security-scans` (optional; if `govulncheck` / `npm audit --json` are available locally, parses High/Critical counts; omits cleanly if unavailable)
  - `--push` (POSTs the resulting `CreateNFRRunRequest` JSON to `POST /api/admin/nfr/runs` using the authenticated Admin JWT)
  - `--out <path>` (writes the schema-valid JSON report to disk for archiving or uploading via the UI `"Import Benchmark / CI JSON"` button)
- **Workload Execution**:
  - Authenticates once as Admin, executes read-only concurrency waves across `GET /api/health`, `GET /api/exams/sample`, and `GET /api/learn/domains` using `concurrent.futures.ThreadPoolExecutor`, computes exact R-7 `p50/p95/p99`, `mean`, `max`, `throughput_rps`, and `error_rate_pct` per concurrency level.
  - Prints a crisp terminal summary table showing `overall_status`, per-level RPS/`p95`, and saved `run_id` when `--push` is used.

---

## 8. Frontend UX, Recharts, Paper Generators & Bilingual Specification

### 8.1 Routing & Shared Admin Portal Sub-Tabs (`Q4=A`)
- Update `frontend/src/components/AdminPortalTabs.tsx`:
  - Expand `active` union to `"specialists" | "evaluations" | "statistics" | "nfr"`.
  - Add 4th tab:
    - `key: "nfr"`, `to: "/admin/nfr"`, `icon: Activity` (from `lucide-react`), `label: t("admin.portalTabs.nfr")` (`"NFR Evaluation"` / `"การประเมิน NFR"`).
- Update `frontend/src/App.tsx`:
  - Add route `admin/nfr` → `<AdminNfrPage />`.
- Top Navbar (`Layout.tsx`) requires **zero changes** (`location.pathname.startsWith("/admin")` already highlights `Admin Portal` and preserves single-row `>= 1280px` width).

### 8.2 Admin NFR Evaluation Page (`frontend/src/pages/AdminNfrPage.tsx` — `Q16=A`)
Strict Admin guard (`userRole !== "admin"` renders the standard Access Denied card matching `AdminEvaluationsPage` / `AdminStatisticsPage`). Strictly adheres to the `rem`-only Semantic Typography Scale (`text-base` prose, `text-sm` UI controls, `text-log` monospace metrics/tables, `text-xs` 12px hard floor).

Structured into **6 Print-Friendly Report Sections**:

1. **Section 1 — Control & Run Bar (`print:hidden`)**:
   - Title + Subtitle + ISO/IEC 25010 `[R10]` Rubric badge (`nfr-rubric-v1`).
   - **Active View Selector**: Dropdown toggling between **`Live System View`** and any persisted **`nfr_evaluation_runs`** snapshot (`label • environment • date • verdict`), plus optional **`Compare Baseline`** dropdown selecting a second run for side-by-side Delta comparison.
   - **Action Buttons**:
     - **`Refresh Live`** (`GET /api/admin/nfr/live`)
     - **`Capture Snapshot`** (inline/modal form for `run_label`, `environment`, `notes` → `POST /api/admin/nfr/runs`)
     - **`Run Safe Concurrency Probe`** (preset selector `Light 5/50` | `Standard 10/100` | `Burst 20/200` + `Save as Run` checkbox → `POST /api/admin/nfr/probe/scalability`)
     - **`Import Benchmark / CI JSON`** (`<input type="file" accept=".json">` validating `<= 1 MB` → `POST /api/admin/nfr/runs`)
     - **`Reset Telemetry Window`** (confirmed action → `POST /api/admin/nfr/telemetry/reset`)
     - **`Print / Save PDF`** (`window.print()`)

2. **Section 2 — Executive KPI Strip (6 Cards)**:
   - **Overall Verdict Badge** (`All Passed` emerald / `Pass with Warnings` amber / `Partial — Awaiting Data` sky / `Failed` red) + `Measured Pass Rate (X / Y)`
   - **Dimension Rollup**: `Pass / Warn / Fail / Pending` counts
   - **Non-LLM API `p95`** (`XX.X ms`, target `<= 300 ms`) + sample count `n`
   - **Availability & DB Health** (`99.X%` Non-5xx + `DB UP • X.X ms`)
   - **Scenario Pipeline Reliability (RQ4)** (`100%` valid delivery, `0%` system failures, AI vs Fallback split)
   - **Server Uptime & Telemetry Window** (`Xd Xh Xm`, total requests in window)
   - *(When a Baseline Comparison Run is active, displays a clean **Delta Comparison Strip** showing signed deltas: `Δ Non-LLM p95`, `Δ Availability`, `Δ Passed Dimensions`, `Δ Peak RPS`.)*

3. **Section 3 — 10-Dimension ISO/IEC 25010 Scorecard Grid**:
   - Toolbar with `Expand All` / `Collapse All` and filter pills (`All (10)`, `Pass`, `Warn`, `Fail`, `Pending`).
   - 2-column responsive grid (`grid-cols-1 lg:grid-cols-2`) displaying all 10 dimensions in canonical order.
   - Each card shows: Dimension Name, ISO/IEC 25010 Characteristic subtitle, Status Pill (`Pass` / `Warn` / `Fail` / `Not Measured` / `Insufficient Data`), primary summary metric (`font-mono`), target rule, and expandable sub-checks table (`Check`, `Measured`, `Target`, `Status`) with zero secret values exposed.

4. **Section 4 — Visual Telemetry & Reliability Charts (`recharts`)**:
   - **Chart 1 — API Latency Percentiles by Endpoint (`BarChart` horizontal/grouped)**: Displays top active normalized routes with `p50`, `p95`, and `p99` ms bars plus a dashed reference line at `300 ms` (`NFR_P95_TARGET_MS`), with honest empty state if `0` requests in window.
   - **Chart 2 — Scalability Concurrency Curve (`ComposedChart` / `LineChart` + `BarChart`)**: Plots `Concurrency Level` (`5 / 10 / 20 / 25 / 50`) on X-axis against `Throughput (RPS)` and `p95 Latency (ms)` (with `500 ms` threshold reference line). Displays an actionable empty state with a direct `"Run Safe Concurrency Probe"` button when not yet measured.
   - **Chart 3 — Simulation Pipeline Reliability & Generation Split (`BarChart` + breakdown pills)**: Compares `AI Generated`, `Static Fallback`, `Valid Snapshots`, and `System Failures`, alongside top `fallback_reasons` counts.

5. **Section 5 — Run History & Baseline Comparison Table (`print:hidden` controls, summary printable)**:
   - Displays all saved `nfr_evaluation_runs` (`Timestamp`, `Label`, `Environment`, `Source` badge, `Overall Verdict`, `Pass/Warn/Fail/Pending`, `Non-LLM p95`, `Availability %`, and actions: `View`, `Set Baseline`, `Delete` with confirmation).

6. **Section 6 — Paper-Ready Engineering Export & `[R10]` ISO/IEC 25010 Methodology Panel (`frontend/src/lib/nfrReport.ts`)**:
   - **Paper Output Language Toggle**: `Academic EN (IEEE)` *(default)* | `Thai Thesis Summary (TH)`.
   - **Export Toolbar (`print:hidden`)**:
     - **`Copy LaTeX (IEEEtran + booktabs)`** — generates a publication-ready `table*` (`\begin{table*}[t]`) NFR Evaluation Matrix covering all 10 ISO/IEC 25010 dimensions, numerical targets, measured results, evidence source, and verdict (properly LaTeX-escaping `%`, `_`, `&`, `#`, `<=`, `>=`, and never outputting `NaN`/`undefined`), plus optional Scalability Curve LaTeX sub-table when benchmark data exists.
     - **`Copy Markdown`** — clean GFM NFR Measurement Matrix + Scalability Table.
     - **`Download CSV (UTF-8 BOM)`** & **`Download JSON`** — authenticated `axios` blob download from `GET /api/admin/nfr/export?format=...&run_id=...`.
   - **Auto-Generated Engineering Evaluation Paragraph**:
     - Dynamically synthesizes an honest academic paragraph citing `[R10]` (ISO/IEC 25010), active `run_label`/`environment`/`evaluated_at`, measured pass counts, Non-LLM `p50/p95/p99`, Availability %, RQ4 scenario delivery/fallback reliability, Security/Safety pass counts, Scalability throughput (if measured), and explicitly flagging any `not_measured` / `insufficient_data` dimensions without overclaiming. One-click **`Copy Paragraph`** button.
   - **Collapsible ISO/IEC 25010 `[R10]` Methodology & Threshold Guide (`EN` / `TH`)**:
     - Documents the 3-layer measurement methodology, R-7 percentile interpolation formula, `nfr-rubric-v1` Pass/Warn/Fail cutoffs, probe telemetry isolation, zero-secret posture policy, and honest `Not Measured` reporting standard so the Admin can directly defend the methodology to advisors and reviewers.

---

## 9. Academic Provenance Registration (`AGENTS.md` Compliance)

Register permanent reference **`R10`** in `docs/REFERENCES.md` and `docs/references.bib` before implementation sign-off:

- **`R10`**: International Organization for Standardization & International Electrotechnical Commission. (2023). *Systems and software engineering — Systems and software Quality Requirements and Evaluation (SQuaRE) — Product quality model* (ISO/IEC 25010:2023; superseding ISO/IEC 25010:2011). ISO/IEC.
- **BibTeX keys**: `iso25010_2023`, `iso25010_2011`
- **Mapping**: Grounds the 10-dimension Non-Functional Requirement evaluation taxonomy, metric definitions, and IEEEtran NFR Measurement Matrix on `/admin/nfr` (Issue #15).

---

## 10. Verification Plan & Acceptance Criteria

### 10.1 Automated Checks
1. **Go Backend Compilation & Unit Tests**:
   - `go build -o NUL main.go` (or `/dev/null` on Linux/CI)
   - `go test ./...` including:
     - `services/nfr_telemetry_test.go`: ring buffer wrap-around (`> capacity`), exact R-7 `p50/p95/p99` interpolation, `non_llm_api` vs `llm_simulation_start` separation, zero-sample `nil` handling, `ResetWindow()` preserving `processStartedAt`, and internal probe isolation skipping middleware capture.
     - `services/nfr_evaluator_test.go`: all 10 dimension threshold transitions (`pass`, `warn`, `fail`, `not_measured`, `insufficient_data`), overall verdict precedence (`Q13=A`), `measured_pass_rate`, zero-secret redaction check, English-only / secret-leak / boilerplate safety scanner unit tests, and `NaN`/`Inf`-free JSON marshaling.
     - `controllers/nfr_controller_test.go`: strict Admin RBAC (`401` unauthenticated, `403` `learner`/`specialist`, `200` `admin` across all `/api/admin/nfr/*` routes), live audit execution, capture snapshot + list + get + delete lifecycle, safe concurrency probe mutex (`409` on concurrent trigger) & hard cap rejection (`400` for `> 20` workers or `> 200` reqs), `1 MB` payload limit (`413`), CSV UTF-8 BOM + formula-injection neutralization (`=`, `+`, `-`, `@` prefixed), telemetry reset, and upgraded `/api/health` (`200` healthy / `503` degraded).
2. **OpenAPI Documentation Sync (`ADR-0005`)**:
   - Regenerate `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml` with `dto/nfr.go` and `/api/admin/nfr/*` Swag annotations.
3. **Frontend Typecheck, Unit Tests, i18n Parity, Typography Guard & Build**:
   - `cd frontend && npm test`:
     - New `frontend/tests/nfrReport.test.ts`: validates IEEEtran LaTeX escaping & table structure, Markdown table rendering, EN/TH auto-generated NFR paragraph (including honest handling of `not_measured` / `insufficient_data`, zero `NaN`/`undefined`), and baseline delta calculations.
     - Existing `frontend/tests/i18n.test.ts`: confirms 100% EN/TH key & `{{param}}` parity across all `admin.nfr.*` and error keys.
     - Existing `frontend/tests/typography.test.ts`: confirms zero arbitrary `text-[...px]` and zero sub-`12px` classes across `AdminNfrPage.tsx` and updated `AdminPortalTabs.tsx`.
   - `cd frontend && npx tsc -b && npm run build`
4. **Live Stack Integration Harness (`scripts/test_integration.py`) & CLI Benchmark (`scripts/run_nfr_benchmark.py`)**:
   - Verify `GET /api/health` returns `200` with `status == "healthy"`, `db_status == "up"`, and numeric `db_latency_ms`.
   - Verify non-admin (`learner`) receives `403` on `GET /api/admin/nfr/live`, `POST /api/admin/nfr/runs`, `POST /api/admin/nfr/probe/scalability`, `POST /api/admin/nfr/telemetry/reset`, and `GET /api/admin/nfr/export`.
   - Verify Admin retrieves `GET /api/admin/nfr/live` (`200`, 10 dimensions, `data_integrity` & `safety` & `security` probes executed cleanly, zero secret substrings leaked).
   - Verify Admin executes `POST /api/admin/nfr/probe/scalability` (`preset: "light", save_run: true`) → `200`, creates persisted run, visible in `GET /api/admin/nfr/runs`, retrievable via `GET /api/admin/nfr/runs/:runId`, exportable as UTF-8 BOM CSV & JSON, and deletable via `DELETE /api/admin/nfr/runs/:runId`.
   - Run `python scripts/run_nfr_benchmark.py --levels 5,10 --requests-per-level 30 --include-ci --out /tmp/nfr_test.json --push` (or local scratch path) and confirm `201 Created` with populated `scalability` and `maintainability` dimensions.

### 10.2 Push-Right Human Checkpoint (`Q12=A`)
Present a single **Visual Review Brief** after all automated tests pass, containing EN & TH screenshots of `/admin/nfr`:
- Top Control Bar, Executive KPI Strip, and 10-Dimension ISO/IEC 25010 Scorecard Grid
- Recharts Telemetry & Reliability Charts (Endpoint Latency, Scalability Curve, Simulation Pipeline Reliability)
- Run History & Baseline Delta Comparison view
- Paper-Ready Engineering Export Panel (LaTeX/Markdown/EN+TH Academic Paragraph, `[R10]` Methodology Guide, and Print/PDF preview)

Close Issue #15 upon user sign-off.
