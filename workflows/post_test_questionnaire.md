# Workflow Spec: Post-Test Evaluation Questionnaire & Paper-Ready Statistical Reporting (Issue #13)

**GitHub Issue:** [#13 — Questionair for evaluate how users are impressive this website](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/13)  
**Status:** `/loop-me` Grilling Rounds 1–3 Settled (`Q1`–`Q16` Resolved, 2026-11-25) — Ready for User Confirmation & Implementation  
**Depends On:**
- [exam_qa_review.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/exam_qa_review.md) (Post-Test completion routing to `/exams/review/:sessionId`, `GET /api/exams/history`, Pre→Post baseline selection)
- [language_switching_and_localization.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/language_switching_and_localization.md) (Bilingual `en`/`th` UI chrome, `EN | TH` toggle, `i18n/en.ts`, `i18n/th.ts`, `i18n/errors.ts`)
- [ui_typography_readability.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/ui_typography_readability.md) (`rem`-only Semantic Typography Scale, `12px` `text-xs` hard floor, single-row Navbar at `>= 1280px`, `frontend/tests/typography.test.ts`)
- [docs/RESEARCH_EVALUATION_PLAN.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/docs/RESEARCH_EVALUATION_PLAN.md) & [docs/RESEARCH_PROTOCOL.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/docs/RESEARCH_PROTOCOL.md) (Statistical analysis candidates, participant-flow reporting, privacy/pseudonymization, pilot caveats)
- [docs/adr/0005-openapi-code-first-with-dedicated-dtos.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/docs/adr/0005-openapi-code-first-with-dedicated-dtos.md) (Dedicated `dto/` structs and Swag OpenAPI generation)

**Scope Boundary with Issue #15:**
- **Issue #13 (this spec)**: Post-Test Learner Evaluation Questionnaire (CSUQ v3 + Raw NASA-TLX + Open-Ended Impression Feedback), Admin Evaluations Dashboard & CSV/JSON Export, and Admin Paper-Ready Statistical Results & Methodology Page.
- **Issue #15 (separate)**: System-level Non-Functional Requirement (NFR) evaluation dashboard for Admins (API latency, uptime/availability, security checks, fallback reliability, scalability).

---

## Objective

Close the subjective usability, cognitive workload, and platform impression evaluation loop after learners complete the **Post-Test**, turning the pre-existing, previously unused `evaluation_metrics` table in `init.sql` (**Zero-DDL — no `init.sql` changes required**) into an end-to-end research pipeline:

1. **Learner Post-Test Questionnaire Loop**: After completing the Post-Test, learners view their score, Pre→Post Learning Gain, and Q&A review on `/exams/review/:sessionId` alongside a prominent non-blocking CTA banner inviting them to `/questionnaire` (with persistent reminders on `/evaluation` and `/profile` until submitted). Learners complete a bilingual (`EN`/`TH`) 3-step guided questionnaire (**CSUQ v3 16 items** + **Raw NASA-TLX 6 dimensions** + **3 structured open-ended impression prompts**) with `localStorage` auto-draft persistence. Upon submission, the Go backend authoritatively snapshots their latest Pre-Test score, Post-Test score, `improvement_pct`, and `user_skill_profiles` into `evaluation_metrics` (strictly once per `learner` account).
2. **Admin Evaluation Operations & Export Loop**: Admins navigate via shared Admin Portal sub-tabs (`/admin/specialists` ↔ `/admin/evaluations` ↔ `/admin/evaluations/statistics`) to inspect aggregate KPI cards, Recharts subscale/workload bar charts, pseudonymized individual responses + qualitative impressions, and download UTF-8 BOM CSV or JSON datasets ready for Excel/SPSS/Python.
3. **Admin Paper-Ready Statistical Reporting Loop**: On `/admin/evaluations/statistics`, a dedicated pure-Go statistical engine computes publication-grade descriptive statistics (`n`, `Mean`, `Sample SD`, Student-$t$ `95% CI`, `Median`, `IQR`, `Min–Max`), **Cronbach's $\alpha$** internal consistency, **Paired $t$-test + Cohen's $d_z$**, **Wilcoxon signed-rank test + rank-biserial $r_{rb}$**, and **Spearman's $\rho$** correlations—paired with one-click **Copy LaTeX (`IEEEtran`)**, **Copy Markdown**, an **Auto-Generated Academic Results Paragraph** (`EN` default + `TH` thesis option), a **Statistical Methodology & Formula Guide** citing `[R8]`/`[R9]`, honest **Small-$n$ Guardrails**, and a clean **Print / Save as PDF** stylesheet.

---

## 1. Loop Profile

- **Loop Name**: Post-Test Usability, Workload & Paper-Ready Statistical Evaluation Loop
- **Frequency / Cadence**:
  - Learner Runtime: triggered once per `learner` after completing the 30-question Post-Test (plus revisit in read-only summary mode afterward).
  - Admin Runtime: on-demand inspection, statistical computation, paper-table copy, print/PDF export, and CSV/JSON dataset download.
  - Verification: executed on every `go test ./...`, `cd frontend && npm test`, `npm run build`, OpenAPI docs sync, and `scripts/test_integration.py` run.
- **Primary Actors**:
  - **Database Model (`models/evaluation_metric.go`)**: Maps existing `evaluation_metrics` table (Zero-DDL).
  - **Pure-Go Scoring & Statistical Engine (`services/evaluation_scoring.go`, `services/evaluation_stats.go`, + unit tests)**: Deterministic CSUQ v3 & Raw NASA-TLX validation/scoring, Student-$t$ CDF/quantile via regularized incomplete beta function, Cronbach's $\alpha$, Paired $t$-test & Cohen's $d_z$, Wilcoxon signed-rank & $r_{rb}$, and tie-safe Spearman $\rho$.
  - **Evaluation Controller & DTOs (`controllers/evaluation_controller.go`, `dto/evaluation.go`, `main.go`)**: Learner eligibility/status/submit endpoints, Admin operational dashboard + CSV/JSON export, and Admin paper-ready statistics endpoint; additive fields on `GET /api/analytics/research-summary`.
  - **Learner Questionnaire UI (`frontend/src/pages/QuestionnairePage.tsx`, `frontend/src/lib/questionnaire.ts`)**: 3-step guided stepper, bilingual CSUQ/TLX catalogs, `localStorage` draft sync, Preview Mode (`specialist`/`admin`), Ineligible lock state, and Read-Only Submitted Confirmation state.
  - **Post-Test Prompt Banners (`ExamReviewPage.tsx`, `EvaluationPage.tsx`, `ProfilePage.tsx`)**: Non-blocking CTA banners flipping to `Submitted ✓` once completed.
  - **Admin Portal Tabs, Operational Dashboard & Paper Statistics Page (`AdminPortalTabs.tsx`, `AdminEvaluationsPage.tsx`, `AdminStatisticsPage.tsx`, `frontend/src/lib/paperStatistics.ts`)**: Shared tabs preserving single-row Navbar, Recharts visualizations, pseudonymized table, UTF-8 BOM CSV/JSON download, LaTeX/Markdown/Narrative generators, Methodology & Formula explainer, and `@media print` stylesheet.
  - **Academic Provenance (`docs/REFERENCES.md`, `docs/references.bib`)**: Registers `R8` (CSUQ v3) and `R9` (NASA-TLX / RTLX).
- **Trigger**:
  - Event-driven on Post-Test completion (`/exams/review/:sessionId` banner), direct visit to `/questionnaire`, or Admin access to `/admin/evaluations` and `/admin/evaluations/statistics`.
- **Checkpoint**:
  - **Runtime**: None (100% autonomous).
  - **Implementation Sign-Off (1 Push-Right Visual Review Checkpoint — `Q12=A`)**: After all Go unit tests, Node frontend tests (questionnaire validation, paper generators, i18n parity, typography guard), OpenAPI Swagger sync, and `scripts/test_integration.py` pass, present a single compact **Visual Review Brief** containing EN & TH screenshots of:
    1. Post-Test Review page (`/exams/review/:postSessionId`) showing the non-blocking Questionnaire CTA banner (before & after submission)
    2. `/questionnaire` across Steps 1 (CSUQ), 2 (NASA-TLX), 3 (Open Feedback), and Read-Only Submitted Confirmation
    3. `/admin/evaluations` (KPI cards, Recharts CSUQ & TLX bar charts, pseudonymized table + expanded notes, export buttons)
    4. `/admin/evaluations/statistics` (Paper-ready tables, Copy LaTeX/Markdown controls, Auto-generated EN/TH academic paragraph, Statistical Methodology & Formula Guide, and Print preview)
    Wait for user approval before closing Issue #13.

---

## 2. Settled Design Decisions (`Q1`–`Q16`)

| Q# | Decision | Settled Rule |
| --- | --- | --- |
| `Q1` | Survey Instruments | **Standard Research Bundle (`1A`)**: CSUQ v3 (16 items) + Raw NASA-TLX (6 dimensions) + Open-Ended Qualitative Impression Feedback (`notes`). |
| `Q2` | Trigger & Post-Test Flow | **Review First + Prominent Banner + Dedicated `/questionnaire` (`2A`)**: Post-Test continues routing to `/exams/review/:sessionId`; non-blocking CTA banner shown on Post-Test review, `/evaluation`, and `/profile` until submitted. |
| `Q3` | Eligibility & Cardinality | **1 Submission per Learner, Post-Test Gated (`3A`)**: Requires `completed` Post-Test (`403` otherwise), strictly 1 submission per `user_id` (`409` on duplicate), switches to read-only submitted confirmation afterward. |
| `Q4` | Scope Boundary with Issue #15 | **Learner + Admin Evaluations Dashboard & Export (`4A`)**: Issue #13 includes learner submission, Admin aggregate/pseudonymized dashboard, CSV/JSON export, and Paper Statistics; Issue #15 stays focused on system-level NFR benchmarks. |
| `Q5` | Localization | **Full Bilingual `EN` / `TH` (`5A`)**: Questionnaire items, anchors, instructions, and Admin UI follow active `EN | TH` locale; submitted `locale` recorded in metadata while numeric scores remain language-agnostic. |
| `Q6` | Database Schema & Snapshot | **Zero-DDL Server Auto-Snapshot (`6A`)**: Reuses existing `evaluation_metrics` table in `init.sql` untouched; backend authoritatively snapshots latest Pre-Test score, Post-Test score, `improvement_pct`, and `user_skill_profiles`. |
| `Q7` | CSUQ Scale & `N/A` Policy | **Ascending `1–7` (Higher = Better) + `N/A` Allowed (`7A`)**: `1 = Strongly Disagree` → `7 = Strongly Agree`; every item required (`1..7` or `N/A`); `N/A` excluded from subscale/overall means per Lewis (1995/2018); all-16 `N/A` rejected. |
| `Q8` | NASA-TLX Variant & Polarity | **Raw NASA-TLX (RTLX), `0–100` step `5`, Intuitive UI Performance (`8A`)**: No 15 pairwise comparisons; UI shows `Performance` as `0 = Failure` → `100 = Perfect`, backend stores `performance_rating` and computes `performance_workload = 100 - performance_rating` for `overall_workload = mean(6 workloads)`. |
| `Q9` | Questionnaire UX & Validation | **3-Step Guided Stepper + `localStorage` Auto-Draft (`9A`)**: Step 1 CSUQ → Step 2 NASA-TLX → Step 3 Open Feedback & Submit; survives refresh & `EN|TH` toggle; numeric items mandatory, open feedback optional (max 2,000 chars per prompt). |
| `Q10` | Qualitative Feedback (`notes`) | **3 Focused Optional Prompts (`10A`)**: (1) Most Impressive on this platform, (2) Areas to Improve / Friction, (3) Additional Comments on Scenarios / Learning Hub; stored both as structured JSONB and readable sectioned `notes TEXT`. |
| `Q11` | Admin Portal & Export | **Shared Admin Portal Sub-Tabs + Recharts + UTF-8 BOM CSV/JSON (`11A`)**: Zero Top Navbar bloat; KPI cards, CSUQ & TLX bar charts (`recharts`), pseudonymized table with expandable notes, and UTF-8 BOM CSV + JSON export. |
| `Q12` | Role Guard, `research-summary`, Provenance & Checkpoint | **Strict `learner` Submit / Non-Learner Preview + Additive Analytics + `R8`/`R9` + 1 Visual Checkpoint (`12A`)**: `specialist`/`admin` see interactive Preview Mode on `/questionnaire` (submit disabled); additive fields on `GET /api/analytics/research-summary`; register `R8` (CSUQ) & `R9` (NASA-TLX) in `docs/REFERENCES.md` & `docs/references.bib`; 1 Push-Right Visual Review Brief before closing #13. |
| `Q13` | Paper Statistics Access & Route | **Admin Only + Dedicated `/admin/evaluations/statistics` (`13A`)**: 3rd tab in `AdminPortalTabs` (`Specialist Applications` \| `Post-Test Evaluations` \| `Paper Statistics`), strictly `RequireAdmin()`. |
| `Q14` | Statistical Computation Scope | **Full IEEE Paper Suite in Pure Go (`14A`)**: Descriptive stats (`n`, `M`, `SD`, `95% CI`, `Median`, `IQR`, `Min–Max`, bands) for CSUQ subscales + Q1–Q16 & NASA-TLX 6 dims + RTLX; Cronbach's $\alpha$; Pre→Post Paired $t$-test + Cohen's $d_z$ + Wilcoxon signed-rank + $r_{rb}$; Spearman $\rho$ correlations (Gain↔CSUQ, Gain↔TLX, CSUQ↔TLX). |
| `Q15` | Paper-Ready Output Tools | **Complete 4-Tool Suite (`15A`)**: `Copy LaTeX (IEEEtran + booktabs)`, `Copy Markdown`, Auto-Generated Academic Results Paragraph, Collapsible Statistical Methodology & Formula Guide (with `[R8]`/`[R9]`), and clean `@media print` Print/PDF view. |
| `Q16` | Copy Language & Small-$n$ Guardrails | **Academic EN Default + TH Thesis Option + Honest Small-$n$ Gates (`16A`)**: UI & Methodology bilingual (`EN`/`TH`), Copy-to-Paper defaults to Academic English with toggle for Thai thesis summary; strict gates (`n < 2` no SD/CI/$\alpha$, `n < 3` no Paired $t$/Cohen's $d_z$/Spearman, `n < 5` no Wilcoxon, `n < 30` Pilot Sample Caveat banner). |

---

## 3. Database & Data Model Contract (Zero-DDL)

No SQL migration or `init.sql` modification is required. The existing `evaluation_metrics` table in `init.sql` (`lines 260–271`) is used verbatim:

```sql
CREATE TABLE evaluation_metrics (
  eval_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(user_id),
  pre_score FLOAT NOT NULL,
  post_score FLOAT NOT NULL,
  improvement_pct FLOAT,
  csuq_scores JSONB,
  nasa_tlx_scores JSONB,
  skill_gap_summary JSONB,
  eval_date TIMESTAMP DEFAULT now(),
  notes TEXT
);
```

### 3.1 Go GORM Model (`models/evaluation_metric.go`)
```go
type EvaluationMetric struct {
	EvalID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"eval_id"`
	UserID          string    `gorm:"type:uuid;not null" json:"user_id"`
	PreScore        float64   `gorm:"not null" json:"pre_score"`
	PostScore       float64   `gorm:"not null" json:"post_score"`
	ImprovementPct  float64   `json:"improvement_pct"`
	CSUQScores      any       `gorm:"column:csuq_scores;type:jsonb;serializer:json" json:"csuq_scores"`
	NASATLXScores   any       `gorm:"column:nasa_tlx_scores;type:jsonb;serializer:json" json:"nasa_tlx_scores"`
	SkillGapSummary any       `gorm:"column:skill_gap_summary;type:jsonb;serializer:json" json:"skill_gap_summary"`
	EvalDate        time.Time `gorm:"default:now()" json:"eval_date"`
	Notes           string    `gorm:"type:text" json:"notes"`
}

func (EvaluationMetric) TableName() string { return "evaluation_metrics" }
```

### 3.2 Concurrency & Cardinality Invariant (Zero-DDL Duplicate Prevention)
Inside `POST /api/evaluation/questionnaire`, execute inside a PostgreSQL transaction:
1. Acquire a transaction-scoped advisory lock keyed by user: `SELECT pg_advisory_xact_lock(hashtext('eval_metric:' || ?))`.
2. Re-check `SELECT COUNT(*) FROM evaluation_metrics WHERE user_id = ?`. If `> 0`, roll back and return `409 Conflict` (`code: "QUESTIONNAIRE_ALREADY_SUBMITTED"`).
3. Insert the single `EvaluationMetric` row and commit.

### 3.3 Canonical JSONB Schemas

#### A. `csuq_scores` JSONB
```json
{
  "schema_version": "csuq-v3-asc-7",
  "scale_min": 1,
  "scale_max": 7,
  "direction": "higher_is_better",
  "locale": "th",
  "items": {
    "q1": 6, "q2": 7, "q3": 6, "q4": 6, "q5": 7, "q6": 6,
    "q7": null, "q8": 6, "q9": 6, "q10": 6, "q11": 7, "q12": 6,
    "q13": 7, "q14": 6, "q15": 6,
    "q16": 7
  },
  "subscales": {
    "sysuse": 6.33,
    "infoqual": 6.20,
    "interqual": 6.33,
    "overall": 6.33
  },
  "valid_counts": {
    "sysuse": 6,
    "infoqual": 5,
    "interqual": 3,
    "overall": 15
  }
}
```
- **Validation**:
  - Exactly keys `q1`..`q16` accepted (missing key or unknown key → `400 QUESTIONNAIRE_INVALID_CSUQ`).
  - Each value is either `null` (`N/A`) or integer `1 <= v <= 7`.
  - At least 1 item across `q1..q16` must be non-null (`valid_counts.overall >= 1`), otherwise `400 QUESTIONNAIRE_ALL_NA`.
- **Academic Provenance (`AGENTS.md`)**: Registers **`R8` (CSUQ v3 — Lewis, 1995; 2018)** and **`R9` (NASA-TLX / RTLX — Hart & Staveland, 1988; Hart, 2006)** in `docs/REFERENCES.md` and `docs/references.bib`.
- **Scoring & Stats**: Subscale formulas (`[R8]`):
  - `sysuse` = arithmetic mean of non-null `q1..q6` (`null` if all 6 are `N/A`), rounded to 2 decimals.
  - `infoqual` = arithmetic mean of non-null `q7..q12` (`null` if all 6 are `N/A`), rounded to 2 decimals.
  - `interqual` = arithmetic mean of non-null `q13..q15` (`null` if all 3 are `N/A`), rounded to 2 decimals.
  - `overall` = arithmetic mean of non-null `q1..q16`, rounded to 2 decimals.

#### B. `nasa_tlx_scores` JSONB
```json
{
  "schema_version": "nasa-tlx-rtlx-v1",
  "variant": "raw_tlx",
  "scale_min": 0,
  "scale_max": 100,
  "step": 5,
  "dimensions": {
    "mental_demand": 60,
    "physical_demand": 10,
    "temporal_demand": 45,
    "performance_rating": 85,
    "performance_workload": 15,
    "effort": 55,
    "frustration": 20
  },
  "overall_workload": 34.17
}
```
- **Validation**:
  - All 6 inputs (`mental_demand`, `physical_demand`, `temporal_demand`, `performance_rating`, `effort`, `frustration`) are required pointers (so `0` is explicitly accepted, while omitted/`nil` → `400 QUESTIONNAIRE_INVALID_TLX`).
  - Each integer must satisfy `0 <= v <= 100` and `v % 5 == 0`.
- **Scoring Formula (Raw NASA-TLX, Hart 2006 — `[R9]`)**:
  - `performance_workload = 100 - performance_rating` (maps intuitive UI `0 = Failure .. 100 = Perfect` into standard RTLX workload polarity where higher = greater perceived workload/poorer performance).
  - `overall_workload = round((mental_demand + physical_demand + temporal_demand + performance_workload + effort + frustration) / 6.0, 2)`.

#### C. `skill_gap_summary` JSONB & `notes` TEXT
```json
{
  "snapshot_version": "eval-snapshot-v1",
  "submitted_locale": "th",
  "pre_session_id": "uuid",
  "post_session_id": "uuid",
  "completed_scenarios": 4,
  "average_proficiency": 78.4,
  "domains": [
    { "domain_id": "domain1", "proficiency_score": 80.0, "passed": true },
    { "domain_id": "domain2", "proficiency_score": 74.5, "passed": true },
    { "domain_id": "domain3", "proficiency_score": 82.0, "passed": true },
    { "domain_id": "domain4", "proficiency_score": 75.5, "passed": true },
    { "domain_id": "domain5", "proficiency_score": 80.0, "passed": true }
  ],
  "qualitative_feedback": {
    "most_impressive": "SIEM log investigation felt realistic...",
    "areas_to_improve": "Would like more cloud IAM scenarios...",
    "additional_comments": "Bilingual explanations helped a lot."
  }
}
```
- **Pre/Post Snapshot**:
  - Queries user's latest completed `post` exam (`WHERE user_id = ? AND exam_type = 'post' AND status = 'completed' ORDER BY completed_at DESC LIMIT 1`). Missing → `403 QUESTIONNAIRE_POST_TEST_REQUIRED`.
  - Queries user's latest completed `pre` exam (`ORDER BY completed_at DESC LIMIT 1`); falls back to `pre_score = 0.0` only if absent in synthetic test data.
  - `improvement_pct = math.Round((postScore - preScore) * 100) / 100` (signed percentage-point delta matching `GetResearchSummary` and `ExamReviewPage`).
- **Qualitative Feedback & `notes` Serialization**:
  - Each of the 3 optional strings (`most_impressive`, `areas_to_improve`, `additional_comments`) is trimmed and capped at `2000` Unicode runes (`400 QUESTIONNAIRE_FEEDBACK_TOO_LONG` if exceeded).
  - Server builds `notes` by joining non-empty sections with double newlines:
    - `[Most Impressive]\n<text>`
    - `[Areas to Improve]\n<text>`
    - `[Additional Comments]\n<text>`
  - If all 3 are empty, `notes = ""`.

---

## 4. Canonical Bilingual Survey Catalog (CSUQ v3 & Raw NASA-TLX)

### 4.1 CSUQ Version 3 (16 Items, Scale `1 = Strongly Disagree / ไม่เห็นด้วยอย่างยิ่ง` → `7 = Strongly Agree / เห็นด้วยอย่างยิ่ง`, plus `N/A / ไม่เกี่ยวข้อง`)

| Key | Subscale | English (`en`) | Thai (`th`) |
| --- | --- | --- | --- |
| `q1` | `SYSUSE` | Overall, I am satisfied with how easy it is to use this system. | โดยรวมแล้ว ฉันพึงพอใจกับความง่ายในการใช้งานระบบนี้ |
| `q2` | `SYSUSE` | It was simple to use this system. | ระบบนี้ใช้งานได้ง่ายและไม่ซับซ้อน |
| `q3` | `SYSUSE` | I was able to complete the tasks and scenarios quickly using this system. | ฉันสามารถทำภารกิจและสถานการณ์จำลอง (Scenarios) ได้อย่างรวดเร็วผ่านระบบนี้ |
| `q4` | `SYSUSE` | I felt comfortable using this system. | ฉันรู้สึกสะดวกสบายและไม่ติดขัดในการใช้งานระบบนี้ |
| `q5` | `SYSUSE` | It was easy to learn to use this system. | การเรียนรู้วิธีใช้งานระบบนี้ทำได้ง่าย |
| `q6` | `SYSUSE` | I believe I could become productive quickly using this system. | ฉันเชื่อว่าฉันสามารถใช้งานระบบนี้ให้เกิดประสิทธิภาพได้อย่างรวดเร็ว |
| `q7` | `INFOQUAL` | The system gave error messages that clearly told me how to fix problems. | ระบบแสดงข้อความแจ้งเตือนข้อผิดพลาด (Error Messages) ที่บอกวิธีแก้ไขปัญหาได้อย่างชัดเจน |
| `q8` | `INFOQUAL` | Whenever I made a mistake using the system, I could recover easily and quickly. | เมื่อฉันทำผิดพลาดระหว่างใช้งานระบบ ฉันสามารถแก้ไขหรือย้อนกลับได้ง่ายและรวดเร็ว |
| `q9` | `INFOQUAL` | The information (such as online help, on-screen messages, and other documentation) provided with this system was clear. | ข้อมูลที่ระบบแสดง (เช่น คำแนะนำบนหน้าจอ บทเรียน และคำอธิบายผล) มีความชัดเจน |
| `q10` | `INFOQUAL` | It was easy to find the information I needed. | ฉันสามารถค้นหาข้อมูลที่ต้องการภายในระบบได้ง่าย |
| `q11` | `INFOQUAL` | The information was effective in helping me complete the tasks and scenarios. | ข้อมูลในระบบช่วยให้ฉันทำภารกิจและสถานการณ์จำลองสำเร็จได้อย่างมีประสิทธิภาพ |
| `q12` | `INFOQUAL` | The organization of information on the system screens was clear. | การจัดวางโครงสร้างข้อมูลบนหน้าจอของระบบมีความเป็นระเบียบและชัดเจน |
| `q13` | `INTERQUAL` | The interface of this system was pleasant. | ส่วนติดต่อผู้ใช้ (Interface) ของระบบนี้มีความสวยงามและน่าใช้งาน |
| `q14` | `INTERQUAL` | I liked using the interface of this system. | ฉันชอบใช้งานส่วนติดต่อผู้ใช้ (Interface) ของระบบนี้ |
| `q15` | `INTERQUAL` | This system has all the functions and capabilities I expect it to have. | ระบบนี้มีฟังก์ชันและความสามารถครบถ้วนตามที่ฉันคาดหวัง |
| `q16` | `OVERALL` | Overall, I am satisfied with this system. | โดยรวมแล้ว ฉันพึงพอใจกับแพลตฟอร์มนี้ |

### 4.2 Raw NASA-TLX (6 Dimensions, Scale `0–100`, Step `5`)

| Key | Label (`EN` / `TH`) | Question (`EN` / `TH`) | Left Anchor (`0`) | Right Anchor (`100`) | Workload Transform |
| --- | --- | --- | --- | --- | --- |
| `mental_demand` | Mental Demand / ภาระทางความคิด | How mentally demanding was the training? / การฝึกอบรมนี้ต้องใช้ความคิดและการวิเคราะห์มากเพียงใด? | Very Low / น้อยมาก | Very High / สูงมาก | `value` |
| `physical_demand` | Physical Demand / ภาระทางร่างกาย | How physically demanding was the training? / การฝึกอบรมนี้ทำให้เกิดความเหนื่อยล้าทางร่างกายมากเพียงใด? | Very Low / น้อยมาก | Very High / สูงมาก | `value` |
| `temporal_demand` | Temporal Demand / แรงกดดันด้านเวลา | How hurried or rushed was the pace of the tasks? / คุณรู้สึกเร่งรีบหรือถูกกดดันเรื่องเวลามากเพียงใด? | Very Low / น้อยมาก | Very High / สูงมาก | `value` |
| `performance_rating` | Performance / ประสิทธิภาพของตนเอง | How successful were you in accomplishing what you were asked to do? / คุณคิดว่าตนเองทำภารกิจและการทดสอบได้สำเร็จมากน้อยเพียงใด? | Failure (Poor) / ไม่สำเร็จ | Perfect (Excellent) / ดีเยี่ยม | `100 - value` |
| `effort` | Effort / ความพยายามที่ใช้ | How hard did you have to work to accomplish your level of performance? / คุณต้องใช้ความพยายามมากเพียงใดเพื่อให้ได้ผลลัพธ์ระดับนี้? | Very Low / น้อยมาก | Very High / สูงมาก | `value` |
| `frustration` | Frustration / ความหงุดหงิดหรือเครียด | How insecure, discouraged, irritated, stressed, and annoyed were you? / คุณรู้สึกสับสน หงุดหงิด เครียด หรือท้อแท้มากเพียงใดระหว่างใช้งาน? | Very Low / น้อยมาก | Very High / สูงมาก | `value` |

---

## 5. Pure-Go Statistical Engine Specification (`services/evaluation_stats.go`)

All publication statistics run deterministically in pure Go with zero external runtime dependencies (no Python/R subprocesses), strictly scoped to `role = 'learner'` rows in `evaluation_metrics` (joined with `users`).

### 5.1 Mathematical & Guardrail Rules
1. **Descriptive Statistics (`DescriptiveStats`)** — computed for CSUQ (`sysuse`, `infoqual`, `interqual`, `overall`, and each individual item `q1..q16`), NASA-TLX (`mental_demand`, `physical_demand`, `temporal_demand`, `performance_rating`, `performance_workload`, `effort`, `frustration`, `overall_workload`), and Exam scores (`pre_score`, `post_score`, `improvement_pct`):
   - `n` (valid non-null observations), `mean`, `median`, `q1`, `q3`, `iqr` (linear interpolation percentile `R-7`, matching NumPy/SciPy default), `min`, `max`.
   - When `n >= 2`: sample standard deviation $s = \sqrt{\frac{1}{n-1}\sum (x_i - \bar{x})^2}$, standard error $\text{SE} = s / \sqrt{n}$, and **two-tailed Student-$t$ 95% Confidence Interval** $\bar{x} \pm t_{0.975,\,n-1}\cdot\text{SE}$ (computed via regularized incomplete beta function $I_x(a,b)$ + bisection accurate to `1e-9`).
   - When `n < 2`: `sd = null`, `ci_low = null`, `ci_high = null`, `status = "insufficient_n"`.
   - **Interpretive Bands**:
     - CSUQ (1–7): `>= 6.00` (`excellent`), `5.00–5.99` (`good`), `4.00–4.99` (`moderate`), `< 4.00` (`needs_improvement`).
     - NASA-TLX Workload (0–100): `< 30.0` (`low`), `30.0–49.99` (`moderate`), `50.0–79.99` (`high`), `>= 80.0` (`very_high`).
2. **Internal Consistency Reliability — Cronbach's $\alpha$ (`CronbachAlphaResult`)**:
   - Computed for: `csuq_overall` (`q1..q16`, $k=16$), `csuq_sysuse` (`q1..q6`, $k=6$), `csuq_infoqual` (`q7..q12`, $k=6$), `csuq_interqual` (`q13..q15`, $k=3$), and `nasa_tlx_overall` (6 workload dims, $k=6$).
   - Uses listwise complete cases over the $k$ items; reports `complete_cases_n` and `excluded_na_cases`.
   - Requires `complete_cases_n >= 2` and `total_variance > 0`; otherwise `alpha = null` with `status = "insufficient_n"` or `"zero_variance"`.
   - Formula: $\alpha = \frac{k}{k-1}\left(1 - \frac{\sum_{j=1}^k s_j^2}{s_{\text{total}}^2}\right)$.
   - Bands: `>= 0.90` (`excellent`), `>= 0.80` (`good`), `>= 0.70` (`acceptable`), `>= 0.60` (`questionable`), `>= 0.50` (`poor`), `< 0.50` (`unacceptable`).
3. **Pre→Post Learning Gain Inference (`LearningGainInference`)**:
   - Paired differences $d_i = \text{post}_i - \text{pre}_i$.
   - **Paired $t$-test & Cohen's $d_z$**:
     - Guard: requires `n >= 3` (`status = "insufficient_n"` if `< 3`) and `sd_d > 0` (`status = "zero_variance"` if `0`).
     - $t = \bar{d} / (s_d / \sqrt{n})$, $df = n - 1$, exact two-tailed $p = I_{\frac{df}{df+t^2}}(df/2, 1/2)$, 95% CI of $\bar{d}$, **Cohen's $d_z = \bar{d} / s_d$** (`< 0.20` negligible, `0.20–0.49` small, `0.50–0.79` medium, `>= 0.80` large).
   - **Wilcoxon Signed-Rank Test**:
     - Drops zero differences ($d_i = 0$); records `zero_diff_count` and `effective_n`.
     - Guard: requires `n >= 5` AND `effective_n >= 5` (`status = "insufficient_n"` otherwise).
     - Computes average ranks for ties, $W^+$, $W^-$, $W = \min(W^+, W^-)$, and **rank-biserial correlation** $r_{rb} = \frac{W^+ - W^-}{W^+ + W^-}$.
     - Two-tailed $p$-value: exact DP distribution when `effective_n <= 20` and no ties; tie-corrected normal approximation with continuity correction otherwise.
4. **Spearman's Rank Correlation (`SpearmanCorrelationResult`)**:
   - Computed for 3 pairs:
     1. `learning_gain_vs_csuq_overall`
     2. `learning_gain_vs_nasa_tlx_overall`
     3. `csuq_overall_vs_nasa_tlx_overall`
   - Guard: requires pairwise complete `n >= 3` (`status = "insufficient_n"` if `< 3`) and non-zero variance in both rank vectors (`status = "zero_variance"` otherwise).
   - Average-rank assignment for ties; Pearson $r$ on ranks = exact Spearman $\rho$; two-tailed $p$-value via Student-$t$ ($t = \rho\sqrt{(n-2)/(1-\rho^2)}$, $df = n - 2$; $p = 0$ if $|\rho| = 1$).
   - Strength bands: `|ρ| < 0.10` negligible, `0.10–0.29` weak, `0.30–0.49` moderate, `>= 0.50` strong.
5. **Participant Flow & Pilot Sample Caveat**:
   - Reports learner cohort funnel: `total_learners`, `completed_pre_test_learners`, `completed_post_test_learners`, `questionnaire_respondents`, `response_rate_pct`.
   - Sets `is_pilot_sample = questionnaire_respondents < 30` (renders Pilot Sample Caveat banner on `/admin/evaluations/statistics` and injects pilot phrasing into generated paper prose).
6. **Zero `NaN` / `Inf` JSON Safety**:
   - All `float64` outputs are sanitized so `math.IsNaN` or `math.IsInf` can never break JSON marshaling.

---

## 6. Backend API Specification (`controllers/evaluation_controller.go`, `dto/evaluation.go`, `main.go`)

### 6.1 Route Registration (`main.go`)
```go
evaluationController := controllers.NewEvaluationController(db)

// Protected learner evaluation questionnaire routes
evaluation := api.Group("/evaluation", middlewares.AuthMiddleware())
{
	evaluation.GET("/questionnaire/status", evaluationController.GetQuestionnaireStatus)
	evaluation.POST("/questionnaire", evaluationController.SubmitQuestionnaire)
}

// Inside existing Admin-only group: api.Group("/admin", middlewares.AuthMiddleware(), middlewares.RequireAdmin())
admin.GET("/evaluations", evaluationController.GetAdminEvaluations)
admin.GET("/evaluations/statistics", evaluationController.GetAdminStatistics)
admin.GET("/evaluations/export", evaluationController.ExportAdminEvaluations)
```

### 6.2 `GET /api/evaluation/questionnaire/status`
- **Auth**: Any authenticated user (`learner`, `specialist`, `admin`).
- **Response (`200 OK`, `dto.QuestionnaireStatusResponse`)**:
  ```json
  {
    "role": "learner",
    "is_preview_only": false,
    "has_completed_post_test": true,
    "latest_post_session_id": "uuid-or-null",
    "eligible": true,
    "ineligible_reason": "",
    "submitted": false,
    "submission": null
  }
  ```
  - When `role != "learner"`: `is_preview_only = true`, `eligible = false`, `ineligible_reason = "preview_only_non_learner"`.
  - When `role == "learner"` and no completed Post-Test: `eligible = false`, `ineligible_reason = "post_test_required"`.
  - When already submitted: `submitted = true`, `eligible = false`, `ineligible_reason = "already_submitted"`, and `submission` contains the learner's read-only `EvaluationMetricSummaryDTO` (`eval_id`, `pre_score`, `post_score`, `improvement_pct`, `csuq_scores`, `nasa_tlx_scores`, `qualitative_feedback`, `notes`, `submitted_locale`, `eval_date`).

### 6.3 `POST /api/evaluation/questionnaire`
- **Request (`dto.SubmitQuestionnaireRequest`)**:
  ```json
  {
    "locale": "th",
    "csuq": {
      "q1": 6, "q2": 7, "q3": 6, "q4": 6, "q5": 7, "q6": 6,
      "q7": null, "q8": 6, "q9": 6, "q10": 6, "q11": 7, "q12": 6,
      "q13": 7, "q14": 6, "q15": 6, "q16": 7
    },
    "nasa_tlx": {
      "mental_demand": 60,
      "physical_demand": 10,
      "temporal_demand": 45,
      "performance_rating": 85,
      "effort": 55,
      "frustration": 20
    },
    "qualitative_feedback": {
      "most_impressive": "...",
      "areas_to_improve": "...",
      "additional_comments": "..."
    }
  }
  ```
- **Guards & Error Codes** (all registered in `frontend/src/i18n/errors.ts` in EN & TH):
  - `401 AUTH_UNAUTHORIZED` — missing/invalid JWT.
  - `403 QUESTIONNAIRE_LEARNER_ONLY` — caller role is not `learner` (`specialist`/`admin` cannot submit real records).
  - `403 QUESTIONNAIRE_POST_TEST_REQUIRED` — caller has no completed `post` exam session.
  - `409 QUESTIONNAIRE_ALREADY_SUBMITTED` — row already exists for `user_id`.
  - `400 QUESTIONNAIRE_INVALID_CSUQ` / `QUESTIONNAIRE_ALL_NA` / `QUESTIONNAIRE_INVALID_TLX` / `QUESTIONNAIRE_FEEDBACK_TOO_LONG` — validation failures.
- **Success (`201 Created`)**: Returns `dto.SubmitQuestionnaireResponse` containing the persisted `EvaluationMetricSummaryDTO`.

### 6.4 `GET /api/admin/evaluations` (Admin Only)
- Queries all `evaluation_metrics` joined with `users` where `users.role = 'learner'`, ordered by `eval_date DESC`.
- Assigns deterministic pseudonymized `participant_code`: `fmt.Sprintf("P-%s", strings.ToUpper(userID[:8]))` alongside `nickname` (never returns `email` or `password_hash`!).
- Returns `dto.AdminEvaluationsResponse`:
  - `cohort`: `total_learners`, `completed_pre_test_learners`, `completed_post_test_learners`, `questionnaire_respondents`, `response_rate_pct`
  - `summary`: `csuq_subscales` (mean & SD for `sysuse`, `infoqual`, `interqual`, `overall`), `nasa_tlx_dimensions` (mean & SD for all 6 workload dims + `overall_workload`), `learning_gain` (mean pre, mean post, mean improvement)
  - `responses`: slice of pseudonymized `AdminEvaluationRowDTO` (`eval_id`, `participant_code`, `nickname`, `locale`, `eval_date`, `pre_score`, `post_score`, `improvement_pct`, `csuq_scores`, `nasa_tlx_scores`, `qualitative_feedback`, `notes`).

### 6.5 `GET /api/admin/evaluations/statistics` (Admin Only)
- Runs `services.BuildPaperStatisticsReport(...)` over all `role = 'learner'` rows in `evaluation_metrics` + cohort funnel counts.
- Returns `dto.AdminPaperStatisticsResponse`:
  - `generated_at`, `protocol_version`, `csuq_schema_version`, `nasa_tlx_schema_version`, `is_pilot_sample` (`n < 30`)
  - `cohort`: funnel metrics
  - `csuq_subscales`: `[]DescriptiveStatDTO` (`sysuse`, `infoqual`, `interqual`, `overall`)
  - `csuq_items`: `[]DescriptiveStatDTO` (`q1`..`q16`)
  - `nasa_tlx`: `[]DescriptiveStatDTO` (`mental_demand`, `physical_demand`, `temporal_demand`, `performance_rating`, `performance_workload`, `effort`, `frustration`, `overall_workload`)
  - `reliability`: `[]CronbachAlphaDTO` (`csuq_overall`, `csuq_sysuse`, `csuq_infoqual`, `csuq_interqual`, `nasa_tlx_overall`)
  - `learning_gain`: `LearningGainStatisticsDTO` (`pre`, `post`, `difference`, `paired_t_test`, `wilcoxon`)
  - `correlations`: `[]SpearmanCorrelationDTO` (3 pairs)

### 6.6 `GET /api/admin/evaluations/export?format=csv|json` (Admin Only)
- Default `format=csv`; invalid format → `400 EVALUATION_INVALID_EXPORT_FORMAT`.
- `format=csv`: writes UTF-8 BOM (`\xEF\xBB\xBF`) + RFC 4180 CSV with headers:
  `participant_code,nickname,eval_date,locale,pre_score,post_score,improvement_pct,csuq_q1,...,csuq_q16,csuq_sysuse,csuq_infoqual,csuq_interqual,csuq_overall,tlx_mental_demand,tlx_physical_demand,tlx_temporal_demand,tlx_performance_rating,tlx_performance_workload,tlx_effort,tlx_frustration,tlx_overall_workload,feedback_most_impressive,feedback_areas_to_improve,feedback_additional_comments,notes`
  - **CSV Formula-Injection Guard**: Prefix any free-text cell beginning with `=`, `+`, `-`, `@`, `\t`, or `\r` with a leading single quote `'` before CSV encoding so exported Excel files remain safe.
- `format=json`: sets `Content-Disposition: attachment; filename="soc_trainer_evaluations_YYYYMMDD.json"` and outputs structured JSON containing `cohort`, `statistics`, and pseudonymized `responses`.

### 6.7 Additive Update to `GET /api/analytics/research-summary`
Add optional fields to `dto.ResearchSummaryResponse` (zero breaking changes):
- `questionnaire_submitted bool`
- `questionnaire_submitted_at *time.Time` (`omitempty`)
- `csuq_overall *float64` (`omitempty`)
- `nasa_tlx_overall *float64` (`omitempty`)

---

## 7. Frontend Specification

### 7.1 Routes & Navigation (`App.tsx`, `Layout.tsx`, `AdminPortalTabs.tsx`)
- Add routes in `App.tsx`:
  - `/questionnaire` → `<QuestionnairePage />`
  - `/admin/evaluations` → `<AdminEvaluationsPage />`
  - `/admin/evaluations/statistics` → `<AdminStatisticsPage />`
- Update `Layout.tsx`:
  - Highlight `Admin Portal` nav link whenever `location.pathname.startsWith("/admin")` (keeps single-row Navbar at `>= 1280px` with zero new top-level links).
- Create shared `frontend/src/components/AdminPortalTabs.tsx` rendered at the top of `AdminSpecialistsPage`, `AdminEvaluationsPage`, and `AdminStatisticsPage`:
  - Tab 1: `Specialist Applications` (`/admin/specialists`)
  - Tab 2: `Post-Test Evaluations` (`/admin/evaluations`)
  - Tab 3: `Paper Statistics` (`/admin/evaluations/statistics`)
  - Hidden in `@media print` (`print:hidden`).

### 7.2 Post-Test Non-Blocking Entry Banners
- **`ExamReviewPage.tsx`** (only when `session.exam_type === "post"`):
  - Fetch `GET /api/evaluation/questionnaire/status` (fail-open).
  - If `eligible && !submitted`: render a high-visibility Graphite/Emerald prompt banner directly beneath the Header / Pre→Post Learning Gain Card: *"Finished your Post-Test? Share your platform impression & experience (2–3 min) →"* linking to `/questionnaire`, plus a matching CTA in the bottom action bar.
  - If `submitted`: render a compact emerald confirmation pill `"Post-Test Evaluation Submitted ✓ — View Summary"` linking to `/questionnaire`.
- **`EvaluationPage.tsx` & `ProfilePage.tsx`**:
  - When `examHistory?.latest_post_session_id` exists, fetch `GET /api/evaluation/questionnaire/status` (fail-open).
  - If `eligible && !submitted`: render a compact reminder card prompting `/questionnaire`.
  - If `submitted`: show a compact `"Evaluation Submitted ✓"` link to `/questionnaire`.

### 7.3 Learner Questionnaire Page (`frontend/src/pages/QuestionnairePage.tsx`)
- **Guards & States**:
  1. Unauthenticated → redirect `/login`.
  2. `is_preview_only === true` (`specialist`/`admin`) → show clear Amber **Preview Mode Banner** (*"Preview Mode — Only learners who complete the Post-Test can submit research evaluations. Form submission is disabled."*) and allow clicking through all 3 steps without calling `POST`.
  3. `!eligible && !submitted && !is_preview_only` (`post_test_required`) → clean locked card explaining that completing the Post-Test unlocks this evaluation, with CTA to `/evaluation` or `/post-test`.
  4. `submitted === true` → **Read-Only Submitted Confirmation View** showing thank-you message, submission timestamp, learner's CSUQ subscale scores (`SYSUSE`, `INFOQUAL`, `INTERQUAL`, `OVERALL` out of 7.0), Raw NASA-TLX overall workload (`/100`) & 6 dimension bars, submitted qualitative feedback, and CTAs back to `/evaluation` and `/profile`.
- **3-Step Guided Stepper (`!submitted`)**:
  - Top Progress Stepper (`1. Usability (CSUQ)` → `2. Workload (NASA-TLX)` → `3. Impressions & Submit`) with completion counter (`e.g., 12 / 16 answered`).
  - **Auto-Draft Persistence**: Keyed by `cybersim_questionnaire_draft:${userId}` in `localStorage`; saves on every input change, restores on refresh and `EN | TH` language switch, and clears immediately upon `201 Created`.
  - **Step 1 — CSUQ v3 (16 items)**:
    - Grouped cleanly into 3 labeled cards (`System Usefulness Q1–Q6`, `Information Quality Q7–Q12`, `Interface Quality & Overall Q13–Q16`) with scale legend (`1 = Strongly Disagree ... 7 = Strongly Agree`, `N/A = Not Applicable`).
    - Every item displays statement at `text-base leading-relaxed`, interactive `1..7` pill buttons (`min-h-[40px]`, keyboard-accessible `role="radiogroup"`) + distinct `N/A` button.
    - `"Next: Workload (NASA-TLX) →"` enabled only when all 16 items are answered and at least 1 item is numeric (inline helper message showing unanswered item count if clicked/disabled).
  - **Step 2 — Raw NASA-TLX (6 dimensions)**:
    - 6 stacked cards (`Mental Demand`, `Physical Demand`, `Temporal Demand`, `Performance`, `Effort`, `Frustration`), each showing dimension title (`text-base font-semibold`), prompt (`text-sm text-graphite-300`), explicit left/right anchor labels (`text-xs`), native range slider (`min=0`, `max=100`, `step=5`) with live monospace value badge (`XX / 100`), plus a subtle note on `Performance` clarifying `0 = Failure → 100 = Perfect`.
    - Tracks explicit touch/confirmation per dimension (default value initialized to `50` with an optional quick-confirm or touch indicator so learners consciously review all 6 sliders before clicking `"Next: Impressions →"`).
  - **Step 3 — Qualitative Impressions & Final Review**:
    - 3 optional textareas (`Most Impressive`, `Areas to Improve`, `Additional Comments`) at `text-base leading-relaxed` with live character counter (`0 / 2,000`).
    - Compact live preview card summarizing their Step 1 CSUQ average & Step 2 NASA-TLX workload before clicking `"Submit Evaluation"`.
    - Upon `201 Created`, clears `localStorage` draft and transitions smoothly into the Read-Only Submitted Confirmation View.

### 7.4 Admin Evaluations Operational Dashboard (`frontend/src/pages/AdminEvaluationsPage.tsx`)
- Access guard: `userRole !== "admin"` → standard Access Denied state (matching `AdminSpecialistsPage`).
- Renders `<AdminPortalTabs active="evaluations" />`.
- **Top Action Bar**: Title + subtitle + **`Export CSV (Excel / SPSS)`** button + **`Export JSON`** button (authenticated `axios` blob download with filename) + quick link button **`Open Paper-Ready Statistics →`**.
- **KPI Summary Cards (5 metrics)**:
  1. Respondents (`n`) & Response Rate (`n / completed_post_test_learners`)
  2. Mean CSUQ Overall (`M ± SD / 7.0`)
  3. Mean Raw NASA-TLX Workload (`M ± SD / 100`)
  4. Mean Pre→Post Learning Gain (`+X.X%`)
  5. Qualitative Feedback Count (respondents with non-empty `notes`)
- **Charts Row (`recharts` `ResponsiveContainer` + `BarChart`)**:
  - Left Card: **CSUQ Subscale Scores (1–7 Scale)** (`SYSUSE`, `INFOQUAL`, `INTERQUAL`, `OVERALL` with reference line at midpoint `4.0`).
  - Right Card: **Raw NASA-TLX Workload Profile (0–100 Scale)** (6 workload dimensions + Overall RTLX).
- **Pseudonymized Participant Table**:
  - Search/filter by nickname/participant code or locale (`All` | `EN` | `TH`).
  - Columns: Participant (`P-XXXXXXXX` + `nickname`), Pre→Post (`XX% → YY%`, signed delta badge), CSUQ Overall (`X.XX`), NASA-TLX (`XX.X`), Locale badge, Submitted Date, and Expand chevron.
  - Expanding a row reveals per-subscale CSUQ breakdown, 6 NASA-TLX values, and all 3 qualitative feedback sections (`Most Impressive`, `Areas to Improve`, `Additional Comments`).

### 7.5 Admin Paper-Ready Statistical Results & Methodology Page (`frontend/src/pages/AdminStatisticsPage.tsx` + `frontend/src/lib/paperStatistics.ts`)
- Access guard: Admin only; renders `<AdminPortalTabs active="statistics" />` (hidden in `@media print`).
- **Top Toolbar (`print:hidden`)**:
  - Language toggle for generated paper prose: `Academic EN (IEEE)` *(default)* | `Thai Thesis Summary (TH)`.
  - **`Print / Save PDF`** button (`window.print()`).
  - **`Copy All LaTeX Tables`** button (copies IEEEtran `booktabs` bundle to clipboard with `"Copied ✓"` toast).
- **Pilot Sample Caveat Banner** (shown when `is_pilot_sample === true`, i.e., `n < 30`):
  - Amber alert card reminding the researcher that current sample size (`n < 30`) is a pilot cohort—report as pilot findings and verify normality assumptions before making confirmatory claims (aligned with `docs/RESEARCH_EVALUATION_PLAN.md`).
- **Section 1 — Participant Flow & Sample Overview**:
  - Compact funnel table (`Total Learners` → `Pre-Test Completed` → `Post-Test Completed` → `Questionnaire Respondents (n)` → `Response Rate %`).
- **Section 2 — Auto-Generated Academic Results Paragraph**:
  - Publication-ready paragraph dynamically populated with real calculated values (`N`, CSUQ `M`, `SD`, `95% CI`, Cronbach's $\alpha$, RTLX `M`, `SD`, `95% CI`, Pre→Post Paired $t$-test / Wilcoxon / Cohen's $d_z$, Spearman $\rho$), gracefully omitting or annotating any test gated by small $n$ (`Never outputs NaN or fake p-values`), with one-click **`Copy Paragraph`** button.
- **Section 3 — Table I: CSUQ System Usability & Reliability (Subscales + Collapsible Q1–Q16 Item Table)**:
  - Columns: Construct / Item, `n`, `Mean`, `SD`, `95% CI`, `Median (IQR)`, `Min–Max`, `Cronbach's α`, `Band`.
  - Per-table **`Copy LaTeX (IEEEtran)`** and **`Copy Markdown`** buttons.
  - Collapsible sub-table for item-level **Q1–Q16 Descriptive Breakdown** (also copyable as LaTeX/Markdown).
- **Section 4 — Table II: Raw NASA-TLX Cognitive Workload Profile**:
  - Columns: Dimension (showing both `Performance (Self-Rating)` and `Performance (Workload = 100 - Rating)` transparently + `Overall RTLX`), `n`, `Mean`, `SD`, `95% CI`, `Median (IQR)`, `Min–Max`, `Band`, plus NASA-TLX Cronbach's $\alpha$.
  - Per-table **`Copy LaTeX (IEEEtran)`** and **`Copy Markdown`** buttons.
- **Section 5 — Table III: Pre→Post Learning Gain Inference & Usability/Workload Correlations**:
  - Part A: Pre-Test, Post-Test, $\Delta$ Gain, **Paired $t$-test** (`t`, `df`, `p`, `95% CI`, **Cohen's $d_z$** + magnitude), and **Wilcoxon signed-rank** (`W`, `n_eff`, `p`, **$r_{rb}$**).
  - Part B: **Spearman Rank Correlations ($\rho$, $p$, $n$, band)** for `Learning Gain ↔ CSUQ Overall`, `Learning Gain ↔ NASA-TLX Overall`, and `CSUQ Overall ↔ NASA-TLX Overall`.
  - Small-$n$ cells clearly display `"N/A (requires n ≥ 3)"` / `"N/A (requires n ≥ 5)"` / `"Zero variance"`.
  - Per-table **`Copy LaTeX (IEEEtran)`** and **`Copy Markdown`** buttons.
- **Section 6 — Statistical Methodology & Formula Guide (Collapsible Cards, Bilingual EN/TH)**:
  - Explains every statistical method used so the user can directly write or defend their paper's Methodology section:
    1. **CSUQ v3 Scoring & N/A Exclusion (`[R8]` Lewis, 1995; 2018)** — 1–7 ascending Likert mapping, subscale composition, N/A omission rule.
    2. **Raw NASA-TLX Scoring & Performance Polarity Inversion (`[R9]` Hart & Staveland, 1988; Hart, 2006)** — unweighted 6-dimension workload average, `100 - performance_rating` transformation, interpretive load thresholds.
    3. **Descriptive Measures & Small-Sample 95% CI** — Bessel-corrected sample SD ($n-1$), `R-7` linear quartiles/IQR, Student-$t$ 95% CI via exact incomplete beta quantile.
    4. **Internal Consistency (Cronbach's $\alpha$)** — exact formula, listwise complete-case reporting, George & Mallery (2003) reliability thresholds.
    5. **Paired Pre→Post Hypothesis Testing & Effect Sizes** — Paired $t$-test + Cohen's $d_z$ (parametric) vs. Wilcoxon signed-rank + rank-biserial $r_{rb}$ (non-parametric), exact DP vs. continuity/tie-corrected normal approximation, and selection rationale (`docs/RESEARCH_EVALUATION_PLAN.md`).
    6. **Spearman's Rank Correlation ($\rho$)** — monotonic rank-based association suited to bounded ordinal Likert/workload scales.
    7. **Validity & Small-Sample Guardrails** — minimum-$n$ thresholds, `n < 30` Pilot Study warning, non-responder/selection-bias note.
- **Print Stylesheet (`frontend/src/index.css` `@media print`)**:
  - Forces clean light-mode print output (`background: #ffffff`, `color: #111827`, crisp borders), hides `header`, `footer`, `.print-hidden`, buttons/copy toolbars, and ensures tables avoid awkward page-break splits.

---

## 8. Academic Provenance Registration (`AGENTS.md` Compliance)

Register two permanent entries in `docs/REFERENCES.md` and `docs/references.bib` before implementation sign-off:
- **`R8`**: Lewis, J. R. (1995). IBM computer usability satisfaction questionnaires: Psychometric evaluation and instructions for use. *International Journal of Human-Computer Interaction*, 7(1), 57–78; and Lewis, J. R. (2018). Measuring perceived usability: The CSUQ, SUS, and UMUX. *International Journal of Human-Computer Interaction*, 34(12), 1148–1156. (BibTeX keys: `lewis1995csuq`, `lewis2018measuring`).
- **`R9`**: Hart, S. G., & Staveland, L. E. (1988). Development of NASA-TLX (Task Load Index): Results of empirical and theoretical research. *Advances in Psychology*, 52, 139–183; and Hart, S. G. (2006). NASA-Task Load Index (NASA-TLX); 20 years later. *Proceedings of the Human Factors and Ergonomics Society Annual Meeting*, 50(9), 904–908. (BibTeX keys: `hart1988nasatlx`, `hart2006nasatlx`).

---

## 9. Verification Plan & Acceptance Criteria

### 9.1 Automated Checks
1. **Go Backend Compilation & Unit Tests**:
   - `go build -o NUL main.go` (or `/dev/null` on Linux/CI)
   - `go test ./...` including:
     - `services/evaluation_scoring_test.go`: CSUQ ascending 1–7 subscale math, N/A exclusion, all-16 N/A rejection, invalid range rejection, NASA-TLX step-5 validation, `performance_workload = 100 - performance_rating`, `overall_workload`, qualitative `notes` formatting + 2,000-char limit.
     - `services/evaluation_stats_test.go`: Table-driven tests verifying `DescriptiveStats` (mean, sample SD, Student-$t$ 95% CI, R-7 median/IQR), **Cronbach's $\alpha$**, **Paired $t$-test + Cohen's $d_z$**, **Wilcoxon signed-rank (exact small-$n$ and tie-corrected normal approx) + $r_{rb}$**, and **Spearman $\rho$** against pre-verified SciPy/textbook reference numbers (`1e-4` tolerance), PLUS all small-$n$ guardrails (`n=0, 1, 2, 3, 4, 5`), zero-variance cases, and zero `NaN`/`Inf` JSON marshaling.
     - `controllers/evaluation_controller_test.go`: Eligibility status (`learner` before/after Post-Test, `specialist`/`admin` preview-only), submit success + server auto-snapshot + duplicate `409`, `403` for non-learner and missing Post-Test, Admin RBAC (`403` for learner/specialist on `/api/admin/evaluations*`), CSV UTF-8 BOM + formula-injection neutralization, and additive fields in `GET /api/analytics/research-summary`.
2. **OpenAPI Documentation Sync (`ADR-0005`)**:
   - Regenerate `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml` with new `/api/evaluation/*` and `/api/admin/evaluations*` DTOs and annotations.
3. **Frontend Typecheck, Unit Tests, i18n Parity, Typography Guard & Build**:
   - `cd frontend && npm test`:
     - New `frontend/tests/questionnaire.test.ts`: validates 16-item CSUQ + 6-dim TLX catalog integrity, draft serialization/validation, and preview calculations.
     - New `frontend/tests/paperStatistics.test.ts`: validates IEEEtran LaTeX output escaping/structure, Markdown table generation, EN/TH academic paragraph rendering, and small-$n$ `"N/A"` formatting (zero `NaN`/`undefined`).
     - Existing `frontend/tests/i18n.test.ts`: confirms 100% EN/TH key & `{{param}}` parity across all new keys and error codes.
     - Existing `frontend/tests/typography.test.ts`: confirms zero arbitrary `text-[...px]` and zero sub-`12px` classes across all new/modified pages & components.
   - `cd frontend && npx tsc -b && npm run build`
4. **Live Stack Integration Harness (`scripts/test_integration.py`)**:
   - Before Post-Test: assert `GET /api/evaluation/questionnaire/status` returns `eligible: false`, `ineligible_reason: "post_test_required"`, and `POST /api/evaluation/questionnaire` returns `403 QUESTIONNAIRE_POST_TEST_REQUIRED`.
   - After Post-Test: assert status returns `eligible: true`; submit valid CSUQ (with at least 1 `N/A`) + Raw NASA-TLX + qualitative feedback → assert `201`, verify server auto-populated `pre_score`, `post_score`, `improvement_pct`, `performance_workload == 100 - performance_rating`, and `GET /api/analytics/research-summary` reflects `questionnaire_submitted == true`.
   - Re-submit attempt → assert `409 QUESTIONNAIRE_ALREADY_SUBMITTED`.
   - Non-admin accessing `/api/admin/evaluations`, `/api/admin/evaluations/statistics`, `/api/admin/evaluations/export` → assert `403`.
   - Admin accessing all 3 admin endpoints → assert `200`, valid JSON statistics payload (with honest small-$n$ guardrails), and UTF-8 BOM CSV header & row presence.

### 9.2 Push-Right Human Checkpoint (`Q12=A`)
Present a single **Visual Review Brief** after all automated tests pass, containing EN & TH screenshots of:
- Post-Test Review page (`/exams/review/:postSessionId`) CTA banner (pending & submitted states)
- `/questionnaire` (Steps 1–3 + Read-Only Submitted Confirmation)
- `/admin/evaluations` (KPI cards, Recharts charts, pseudonymized table, export buttons)
- `/admin/evaluations/statistics` (Paper-ready tables, Copy LaTeX/Markdown & Academic EN/TH Paragraph, Methodology Guide, and Print preview)

Close Issue #13 upon user sign-off.
