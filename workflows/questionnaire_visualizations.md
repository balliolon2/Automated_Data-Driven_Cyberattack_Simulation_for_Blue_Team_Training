# Workflow Spec: Comprehensive Questionnaire & Learning Gain Visualizations (Issue #73)

**GitHub Issue:** [#73 — feat(evaluation): add comprehensive questionnaire and pre/post learning gain visualizations](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/73)  
**Status:** `/loop-me` Grilling Rounds 1–3 Settled (`Q1`, `Q5`, `Q12`–`Q13` Resolved, 2026-11-27) — Ready for Implementation  
**Depends On:**
- `workflows/post_test_questionnaire.md` (Issue #13 — `evaluation_metrics`, `/questionnaire`, `/admin/evaluations`, `/admin/evaluations/statistics`, `GET /api/admin/evaluations`, `GET /api/admin/evaluations/statistics`).
- `workflows/ui_typography_readability.md` (`rem`-only Semantic Typography Scale, `12px` `text-xs` hard floor, `frontend/tests/typography.test.ts`).
- `docs/adr/0005-openapi-code-first-with-dedicated-dtos.md` (Swagger/OpenAPI sync).

---

## Objective

Upgrade Questionnaire and Learning Gain visualizations across **all 3 evaluation surfaces** using `recharts` so researchers, advisors, and learners can visually interpret usability (CSUQ v3), cognitive workload (Raw NASA-TLX), cohort completion, and Pre→Post learning gains at a glance — with **1-click SVG download** and crisp `@media print` support for slides and thesis reports:

1. **Admin Operational Dashboard (`/admin/evaluations`)**: 4 rich interactive charts (**Cohort Funnel**, **CSUQ Stacked Likert Distribution `1–7` + `N/A`**, **NASA-TLX 6-Axis Radar Chart**, **Pre→Post Paired Learning Gain Chart**) computed cleanly from existing `GET /api/admin/evaluations` data.
2. **Admin Paper Statistics Workbench (`/admin/evaluations/statistics`)**: Publication Figures section (**Pre→Post Paired Comparison** + **3 Spearman Correlation Scatter Plots**) powered by an additive pseudonymized `participant_points` array in `GET /api/admin/evaluations/statistics`.
3. **Learner Submitted Confirmation View (`/questionnaire`)**: Personal visual achievement summary (**Pre→Post Score Jump Bar** + **Personal CSUQ & NASA-TLX Radar/Bars**).

---

## 1. Loop Profile

- **Loop Name**: Post-Test Questionnaire Visual Analytics & Presentation Export Loop
- **Trigger**:
  - Learner submits `/questionnaire` (renders personal summary charts immediately).
  - Admin opens `/admin/evaluations` or `/admin/evaluations/statistics` and optionally clicks **"Download SVG"** or **"Print / PDF"**.
- **Checkpoint**:
  - **Checkpoint 2 (Pre-Flight Visual Sign-Off — `Q13`)**: Present EN & TH screenshots of `/admin/evaluations`, `/admin/evaluations/statistics`, and `/questionnaire` submitted view (with non-zero cohort data + empty/small-`n` guard states) + verify downloaded `.svg` integrity before closing `#73`.

---

## 2. Settled Design Decisions

| Q# | Decision | Settled Rule |
| --- | --- | --- |
| `Q5`, `Q12` | 3-Surface Recharts Suite | Upgrade `/admin/evaluations` (4 charts), `/admin/evaluations/statistics` (Pre→Post Comparison + 3 Spearman Scatter Plots), and `/questionnaire` Submitted View (2 personal charts). |
| `Q12` | Additive Backend Data Contract | Add `participant_points` (`[]ParticipantPlotPointDTO`: `participant_code`, `pre_score`, `post_score`, `improvement_pct`, `csuq_overall`, `nasa_tlx_overall`) to `GET /api/admin/evaluations/statistics` (pseudonymized `P-XXXXXXXX`, zero PII, Zero-DDL, backward-compatible). |
| `Q12` | Slide/Paper Export | Wrap every Admin chart in a reusable `ChartExportCard` with a 1-click **"Download SVG"** button (cloning the rendered Recharts `<svg>` with clean standalone background + font styling) and print-safe `@media print` styling. |

---

## 3. Backend Additive Contract (Zero-DDL)

### 3.1 `dto/evaluation.go`
Add `ParticipantPlotPointDTO` and append field to `AdminPaperStatisticsResponse`:
```go
type ParticipantPlotPointDTO struct {
	ParticipantCode string   `json:"participant_code" example:"P-A1B2C3D4"`
	PreScore        float64  `json:"pre_score" example:"46.7"`
	PostScore       float64  `json:"post_score" example:"76.7"`
	ImprovementPct  float64  `json:"improvement_pct" example:"30.0"`
	CSUQOverall     *float64 `json:"csuq_overall,omitempty" example:"6.25"`
	NASATLXOverall  *float64 `json:"nasa_tlx_overall,omitempty" example:"35.0"`
}
```
Inside `AdminPaperStatisticsResponse`:
```go
ParticipantPoints []ParticipantPlotPointDTO `json:"participant_points"`
```

### 3.2 `services/evaluation_stats.go` & `controllers/evaluation_controller.go`
- Populate `ParticipantPoints` deterministically from the same `role = 'learner'` rows used to compute `LearningGain` and `Correlations`, ordered by `eval_date ASC` (so chart axes follow cohort order consistently), omitting `email` and raw `user_id` (using `P-XXXXXXXX` pseudonym only).
- Ensure JSON-safe finite numbers (`0` NaN/Inf).
- Add unit tests in `services/evaluation_stats_test.go` verifying `ParticipantPoints` length, pseudonym formatting, and value parity with `PairedTTest` / `Spearman`.
- Regenerate Swagger docs (`docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go`).

---

## 4. Frontend Architecture & Visualization Specification

### 4.1 Pure Chart Data Builders (`frontend/src/lib/evaluationCharts.ts`)
Pure TypeScript functions (100% unit-tested via `node --test` in `frontend/tests/evaluationCharts.test.ts`):
1. **`buildCohortFunnelData(cohort)`**: 4-step array (`Total Learners`, `Pre-Test Done`, `Post-Test Done`, `Survey Submitted`) with counts and `% of total`.
2. **`buildCSUQDistributionData(responses, mode: "subscales" | "items")`**: Computes 100% stacked percentage buckets across all valid responses:
   - `low` (`1–2`), `neutral` (`3–4`), `positive` (`5`), `strong` (`6–7`), `na` (`N/A`), plus mean score and `n` for tooltips.
   - Supports both compact **4-Subscale view** (`SYSUSE`, `INFOQUAL`, `INTERQUAL`, `OVERALL`) and toggleable **16-Item (`Q1–Q16`)** view.
3. **`buildTLXRadarData(summaryOrDimensions)`**: 6-axis array (`Mental`, `Physical`, `Temporal`, `Perf. Workload`, `Effort`, `Frustration`) normalized to `0–100` (with optional cohort benchmark/reference `50` ring).
4. **`buildPrePostPairedData(rowsOrPoints)`**: Per-participant (`P-XXXXXXXX`) `{ label, pre, post, delta }` series (capped/scrollable cleanly when `n > 20`, plus cohort `Pre Mean` vs `Post Mean` summary row).
5. **`buildSpearmanScatterSeries(points)`**: Produces 3 clean pairwise arrays (`gainVsCsuq`, `gainVsTlx`, `csuqVsTlx`), filtering out null points cleanly so Recharts `<ScatterChart>` never renders NaN points.
6. **`exportSvgElement(containerEl, filename)`**: Pure DOM helper locating the Recharts `.recharts-wrapper svg`, cloning it, setting explicit `xmlns`, `width`, `height`, and dark/light backdrop rectangle, serializing via `XMLSerializer`, and triggering a clean `.svg` Blob download.

### 4.2 Reusable Chart Wrapper (`frontend/src/components/ChartExportCard.tsx`)
- Renders card header (title, concise subtitle, optional view-mode toggle slot like `Subscales | Q1–Q16`), **1-click `"Download SVG"` button** (`print:hidden`), responsive Recharts container, and clean empty-state / small-`n` guardrail message when `n === 0` (or `n < 3` for correlation scatter).

### 4.3 Surface 1: `/admin/evaluations` (`AdminEvaluationsPage.tsx`)
Replace the legacy 2 flat bar charts with a 2×2 visualization grid wrapped in `ChartExportCard`:
1. **Cohort Participation Funnel** (horizontal Recharts `BarChart` showing drop-off & completion % across all 4 cohort stages).
2. **CSUQ v3 Stacked Likert Distribution (`1–7` + `N/A`)** (100% stacked horizontal `BarChart` with `Subscales` | `All 16 Items` toggle, showing exact response spread instead of hiding variance behind flat averages).
3. **Raw NASA-TLX 6-Dimension Workload Radar** (Recharts `RadarChart` + `PolarGrid` + `PolarAngleAxis` + `PolarRadiusAxis domain={[0, 100]}` highlighting cognitive workload shape across all 6 dimensions).
4. **Pre-Test vs. Post-Test Learning Gain Comparison** (grouped Recharts `BarChart` comparing Pre vs. Post scores per participant + mean reference lines, visually illustrating positive `Δ Score` at a glance).

### 4.4 Surface 2: `/admin/evaluations/statistics` (`AdminStatisticsPage.tsx`)
Add a dedicated **Visual Figures** section (placed directly above or alongside Table III, styled with `.paper-report-section` for crisp `@media print` PDF output):
1. **Figure 1 — Pre-Test vs. Post-Test Paired Cohort Chart** (paired Pre/Post bars + `Δ` delta tooltip for every respondent + cohort means).
2. **Figure 2 — Spearman Correlation Scatter Plots (3-column responsive grid)**:
   - Learning Gain (`Δ%`) vs. CSUQ Overall (`1–7`)
   - Learning Gain (`Δ%`) vs. NASA-TLX Workload (`0–100`)
   - CSUQ Overall (`1–7`) vs. NASA-TLX Workload (`0–100`)
   - Each card displays its live `ρ` and `p` badge from `data.correlations`, renders `<ScatterChart>` when `n >= 3`, and shows an honest small-`n` message when `n < 3`.

### 4.5 Surface 3: Learner Submitted View (`/questionnaire` — `QuestionnairePage.tsx`)
Upgrade the existing read-only Submitted Confirmation View with a motivating 2-column visual wrap-up:
1. **My Learning Progress Chart**: Clean Recharts `BarChart` comparing the learner's own `Pre-Test %` vs `Post-Test %` with signed `+Δ%` badge.
2. **My NASA-TLX Workload Radar & CSUQ Subscale Bars**: Compact Recharts `RadarChart` (6 NASA-TLX workload axes) + horizontal progress bars (`1–7` scale) for `SYSUSE`, `INFOQUAL`, `INTERQUAL`, `OVERALL`.

---

## 5. Verification & Acceptance Criteria

- [ ] Go unit tests (`go test ./...`) verify `ParticipantPoints` population, pseudonymization (zero emails/UUIDs), and finite numeric safety.
- [ ] Node unit tests (`frontend/tests/evaluationCharts.test.ts`, `i18n.test.ts`, `typography.test.ts`) pass 100%, verifying Likert bucket percentages sum accurately, `N/A` handling, radar dimension ordering, scatter null filtering, and `rem`-only typography compliance.
- [ ] `npm run tsc` and `npm run build` pass with zero TypeScript or Recharts type errors.
- [ ] Swagger docs regenerated cleanly for `ParticipantPlotPointDTO`.
- [ ] Live visual check (Checkpoint 2): EN & TH screenshots of all 3 pages + verified downloaded `.svg` opens cleanly in a browser/editor.
