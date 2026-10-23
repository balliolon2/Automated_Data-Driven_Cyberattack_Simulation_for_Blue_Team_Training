# Workflow: Evidence Log Selection & Finding Scoring (Issue #11)

## Objective

Replace the brittle "the filtered view must contain every evidence log at once" discovery mechanic with an explicit learner action: **click a log row to mark it as evidence**. A key finding is credited when the learner selects any one of that finding's authored evidence logs. Selections are persisted server-side for refresh recovery and Specialist review. Correctness is revealed only after submission.

This workflow exists because the current implementation can award **zero** Detect & Analysis points even when the learner has visibly identified the correct log:

- `SimulationPage.tsx` infers discovery from the transient KQL-filtered result set.
- `scripts/generate_scenarios.py` assigns `evidence_log_indices` by positional chunking (`indices 15..N-16` split across findings), not by what the log actually proves.
- A learner who reads the right log but does not construct a query returning every chunked index scores nothing.

The loop this workflow makes real: **Learner Scenario Triage → evidence selection → scoring → result review**, repeated across the 1–10 scenario training loop.

---

## 1. Loop Profile

- **Loop Name**: Learner Scenario Triage Loop (Detect & Analyze sub-loop)
- **Frequency / Cadence**: Once per simulation session, repeated across the scenario training loop.
- **Primary Actors**:
  - **Learner**: investigates telemetry, selects evidence logs, classifies TP/FP, chooses response actions, submits.
  - **Evaluator (`services/evaluator.go`)**: converts the selected log set into finding credit and penalty.
  - **Scenario Authoring Pipeline**: assigns semantically meaningful evidence logs to every finding. See `workflows/scenario_evidence_authoring.md`.
  - **Specialist**: reviews selected evidence, matched findings, missed findings, and noise picks after completion. See `workflows/specialist_scenario_analysis.md`.

---

## 2. Trigger

- **Event**: a simulation session is `in_progress` and the learner opens **2. Detect & Analyze**.
- **Recovery event**: `GET /api/simulation/session` after a refresh/drop returns the previously selected log set.

No schedule is involved. This learner loop uses no AI. The only human checkpoint is the existing final submission.

---

## 3. Canonical Terms

| Term | Meaning |
|---|---|
| **Evidence log** | A row in `scenarios.initial_logs` that proves a key finding. Identified by its 0-based index in the scenario's original `initial_logs` array. |
| **Selected log** | A log the learner explicitly toggled on. A claim of relevance, not a TP/FP verdict. |
| **Noise log** | Any log not present in any key finding's evidence set. Selecting one is a false lead. |
| **Finding credit** | Full points for a key finding, awarded when the selected set intersects that finding's evidence set. No partial credit within a finding. |
| **Evidence penalty** | `−2` per selected noise log, applied once per unique log. |
| **Benign context log** | A log that establishes a False Positive verdict: approved change, HR travel confirmation, scheduled maintenance, authorized scanner range, etc. Required evidence for FP scenarios. |

---

## 4. Data Contract

### 4.1 Scenario `expected_outcomes.key_findings[]`

```json
{
  "id": "d1_f1",
  "description": "Identified push fatigue indicators",
  "domain_id": "domain1",
  "points": 10,
  "evidence_log_indices": [16, 17, 18],
  "verdict_support": "malicious",
  "explanation": "Three denied MFA pushes followed by an allowed push..."
}
```

Rules:

1. `evidence_log_indices` is **authored**. Positional chunking is forbidden.
2. Each finding has at least one evidence index. The count is scenario-dependent: a simple finding may need one log; a complex or False Positive finding may need several logs to establish the explanation. There is no per-finding cap.
3. Every index satisfies `0 <= index < len(initial_logs)`.
4. Duplicate indices within one finding are invalid.
5. Overlap between findings is allowed freely. A selected log that appears in multiple evidence sets credits every finding it supports and is never noise.
6. The union of all findings' evidence indices must be **≤ 50% of `len(initial_logs)`** (`floor(len * 0.5)`). This prevents tagging most logs as evidence and makes click-everything unprofitable.
7. For `is_true_positive == false`, at least one finding must be marked as verdict-supporting (`verdict_support: "benign"`) and its evidence must include the benign context log(s), not merely the alert that opened the case.
8. For `is_true_positive == true`, at least one finding must be marked `verdict_support: "malicious"`.
9. `points` defaults to `10` if absent; points must be a positive integer. Non-positive or fractional points are an authoring error.
10. `verdict_support` is stored on the finding (`malicious`, `benign`, or `supporting`) so the evaluator, seed validator, and Specialist review can verify that the scenario's verdict is actually provable from the evidence set.
11. **JSONB decoding**: GORM returns JSONB arrays as `[]interface{}` of `float64`. The evaluator must accept `[]int`, `[]float64`, or `[]interface{}` and reject non-integer or non-numeric values with a logged warning; never silently truncate a float.

### 4.2 Client-facing scenario payload

- **Remove** `sanitized_findings` and all `evidence_log_indices` from every learner-facing response. Shipping the answer key to the browser allows inspection and trivial cheating.
- Expose only `total_findings` (integer) for UI copy. Do not expose finding IDs, descriptions, evidence indices, or verdict-support markers before submission.

`ScenarioResponse` change:

```go
type ScenarioResponse struct {
    ScenarioID   string `json:"scenario_id"`
    Title        string `json:"title"`
    Description  string `json:"description"`
    DomainID     string `json:"domain_id"`
    Difficulty   int    `json:"difficulty"`
    InitialLogs  any    `json:"initial_logs"`
    PlaybookSteps any   `json:"playbook_steps"`
    TotalFindings int   `json:"total_findings"`
    // SanitizedFindings removed
}
```

### 4.3 Log identity

- Identity is the 0-based index in the scenario snapshot's original `initial_logs` array.
- The frontend attaches `__log_index` to each hydrated log object.
- KQL filtering and table sorting never change identity.
- `__`-prefixed metadata is excluded from table columns and from plain-text KQL search.

### 4.4 Session actions

New enum value `select_evidence`:

```sql
-- init.sql CREATE TYPE action_type
CREATE TYPE action_type AS ENUM (
  'view_log', 'triage_alert', 'query_rag', 'escalate',
  'respond', 'help_click', 'submit_decision', 'select_evidence'
);
```

Existing databases receive an idempotent migration executed outside a transaction:

```sql
ALTER TYPE action_type ADD VALUE IF NOT EXISTS 'select_evidence';
```

The `init.sql` change only affects fresh databases; the `ALTER TYPE` covers existing deployments. Do not wrap the `ALTER TYPE` in the same transaction as table migrations.

Action payloads:

```json
{ "operation": "toggle", "log_index": 17, "selected": true }
```

```json
{ "operation": "clear" }
```

- Append-only. The current set is derived by replaying `select_evidence` actions in `step_order`.
- Toggle replay: add when `selected == true`, remove when `false`.
- Clear replay: empty the set.
- `Points` is `0`; `IsCorrect` is `NULL`.
- Replaying the same toggle twice is harmless; the final set is idempotent.
- Malformed payloads are skipped with a server warning; they must not fail the whole replay.

**Step-order serialization**: `step_order` allocation must be serialized per session so rapid toggles cannot collide on `UNIQUE (session_id, step_order)`.

1. `BEGIN`.
2. `SELECT session_id FROM simulation_sessions WHERE session_id = ? FOR UPDATE`.
3. `SELECT COALESCE(MAX(step_order), 0)`.
4. Insert the action with `MAX + 1`.
5. `COMMIT`.
6. On unique violation, retry once. If it still fails, return `500` with a retryable error.

### 4.5 DTO changes

`models.SubmitScenarioInput`:

```go
type SubmitScenarioInput struct {
    SessionID          string   `json:"session_id" binding:"required"`
    IsTruePositive     bool     `json:"is_true_positive"`
    TPFPSelected       bool     `json:"tp_fp_selected" binding:"required"`
    SelectedActions    []string `json:"selected_actions" binding:"required"`
    SelectedLogIndices []int    `json:"selected_log_indices"`
    // DiscoveredFindings removed from the new path
}
```

`models.FindingResult`:

```go
type FindingResult struct {
    ID                 string `json:"id"`
    Description        string `json:"description"`
    DomainID           string `json:"domain_id"`
    Found              bool   `json:"found"`
    Points             int    `json:"points"`
    Earned             int    `json:"earned"`
    EvidenceLogIndices []int  `json:"evidence_log_indices"`
    MatchedLogIndices  []int  `json:"matched_log_indices"`
    Explanation        string `json:"explanation"`
}
```

`models.ScenarioResult` additions:

```go
EvidencePenalty    float64 `json:"evidence_penalty"`
NoiseLogIndices    []int   `json:"noise_log_indices"`
SelectedLogIndices []int   `json:"selected_log_indices"`
```

`dto.SubmitScenarioRequest` mirrors `SubmitScenarioInput`. `discovered_findings` is removed from the documented API and ignored if sent.

---

## 5. Scoring

Let:

- `F` = the scenario's key findings.
- `E_f` = authored `evidence_log_indices` of finding `f`.
- `S` = the final selected index set recorded server-side.
- `points_f` = configured points for finding `f` (default 10).
- `noise(S) = S \ ⋃ E_f`.

```text
credited_f = points_f                  if S ∩ E_f ≠ ∅
           = 0                         otherwise

gross         = Σ credited_f
EvidencePenalty = 2 × |noise(S)|
FindingsPoints  = max(0, gross − EvidencePenalty)
```

- No partial credit inside a finding: one matching evidence log earns the full finding.
- Extra evidence logs from the same finding do not double-credit.
- A log shared by multiple findings credits each; it is penalized only if it belongs to none.
- `FindingsDetail[].Earned` is the finding credit before the global penalty. `Found` is `Earned > 0`.
- `FindingsDetail[].MatchedLogIndices = sorted(S ∩ E_f)`.
- `NoiseLogIndices = sorted(noise(S))`.
- `SelectedLogIndices = sorted(S)`.
- `FindingsPoints` is the net findings score after penalty, floored at 0.

### 5.1 Domain attribution

- Finding credit goes to `finding.domain_id`.
- The evidence penalty is deducted from `scenario.domain_id`.
- Each domain's earned points are floored at 0 before percentage calculation.
- TP/FP max points (20) stay attributed to `scenario.domain_id`.
- Response action scoring is unchanged.
- Domain scores are diagnostic and may not sum to `TotalScore` when a penalty is attributed to a single domain. Do not re-derive `TotalScore` by averaging domain percentages.

### 5.2 Total score

The existing raw-point model is retained:

```text
totalEarned = TPFPPoints + FindingsPoints + ResponsePoints
totalMax    = 20 + Σ(points_f) + Σ(positive response points)
TotalScore  = round(clamp(totalEarned, 0, totalMax) / totalMax × 100, 2)
```

No re-normalization to 20/40/40 is introduced; the nominal labels in the constants remain descriptive.

### 5.3 Verdict independence

TP/FP classification remains an independent 20-point component. Finding credit does **not** require a correct verdict. A learner who selects the exculpatory logs but still says "True Positive" earns the finding credit and loses the classification points. This measures two separate skills.

### 5.4 Worked example

Scenario: 2 findings × 10 points; `E_1 = {15,16}`, `E_2 = {20}`; 32 logs total.

| Selected `S` | Gross | Noise | Penalty | FindingsPoints |
|---|---:|---:|---:|---:|
| `{15}` | 10 | 0 | 0 | 10 |
| `{15,20}` | 20 | 0 | 0 | 20 |
| `{15,17}` | 10 | 1 (`17`) | 2 | 8 |
| `{15,20,17,18}` | 20 | 2 | 4 | 16 |
| all 32 logs | 20 | 29 | 58 | 0 |

---

## 6. Pipeline

### Step 1: Hydrate with stable indices

On `POST /api/simulation/start` and `GET /api/simulation/session`, the frontend maps:

```ts
scenario.initial_logs.map((log, index) => ({ ...log, __log_index: index }))
```

`LogTable` excludes keys starting with `__` when discovering columns. `KqlSearchBar` plain-text search must ignore `__`-prefixed metadata (use a JSON replacer or a sanitized copy).

### Step 2: Toggle evidence

`POST /api/simulation/select-log`

```json
{ "session_id": "uuid", "log_index": 17, "selected": true }
```

Validation:

1. Session exists, belongs to the authenticated user, and is `in_progress`.
2. `0 <= log_index < len(initial_logs)`.
3. Scenario snapshot exists and `initial_logs` is an array.
4. Append a `select_evidence` action using the serialized step-order procedure.
5. Return the replayed selection.

`log_index = 0` is valid. Do not use Gin's `binding:"required"` on a plain `int`; use `*int` or manual validation.

The UI updates optimistically. On failure, revert the toggle and show a non-blocking error. Previous selections remain when a KQL filter hides them.

### Step 3: Clear all

`POST /api/simulation/select-log/clear`

```json
{ "session_id": "uuid" }
```

Appends one `{"operation":"clear"}` action and returns an empty set. This keeps Clear all auditable without N requests.

### Step 4: Rehydrate

Both `GET /api/simulation/session` and `POST /api/simulation/start` return:

```json
{
  "session": { "...": "..." },
  "scenario": { "...": "...", "total_findings": 3 },
  "selected_log_indices": [15, 18]
}
```

New sessions return `[]`. The frontend restores selection before rendering the table.

### Step 5: Submit

`POST /api/simulation/submit`

```json
{
  "session_id": "uuid",
  "is_true_positive": true,
  "tp_fp_selected": true,
  "selected_actions": ["c1", "e2", "r1"],
  "selected_log_indices": [15, 18]
}
```

- The server replays `select_evidence` actions and uses that set as the source of truth.
- If the client payload differs, log a warning and score the replayed set. Do not fail the submission.
- Record the final set in the `submit_decision` payload:

```json
{
  "selected_log_indices": [15, 18],
  "client_selected_log_indices": [15, 18],
  "selection_mismatch": false,
  "findings_points": 18,
  "evidence_penalty": 2
}
```

- The `submit_decision` action's `Points` is `int(result.FindingsPoints)` (net after penalty).
- `triage_alert` and `respond` actions are unchanged.
- `SubmitScenario` must use the same per-session row lock as `select-log` when allocating `step_order`; a toggle racing a submit must not collide on `(session_id, step_order)`.

### Step 6: Evaluate and reveal

`services.ScoreSubmission` applies §5. `GET /api/simulation/result/:sessionId` returns the evidence review:

- Per finding: `found`, `earned`, `points`, all acceptable `evidence_log_indices`, `matched_log_indices`, explanation.
- `selected_log_indices`, `noise_log_indices`, `evidence_penalty`, and net `findings_points`.
- `scenario.initial_logs` remains available to the session owner so the result page can show the row text.
- `session.rubric_version` is `rubric-v2-evidence-click` for new sessions.

---

## 7. API Contract

### 7.1 `POST /api/simulation/select-log`

| Field | Type | Notes |
|---|---|---|
| `session_id` | UUID string | required |
| `log_index` | int | 0-based; 0 is valid |
| `selected` | bool | true = select, false = deselect |

Response `200`:

```json
{ "selected_log_indices": [15, 17], "selected_count": 2 }
```

Errors:

- `400`: invalid body, invalid index, scenario logs unavailable.
- `403`: session not owned by caller or not `in_progress`.
- `404`: session or scenario not found.
- `500`: step-order allocation failed after retry; client may retry.

### 7.2 `POST /api/simulation/select-log/clear`

Request `{ "session_id": "uuid" }`; response `{ "selected_log_indices": [], "selected_count": 0 }`.

### 7.3 `POST /api/simulation/submit`

- `selected_log_indices` is accepted for reconciliation but never trusted over replay.
- `discovered_findings` is removed from the contract; if present, ignore it.
- Existing response keys remain; `result` gains the fields in §4.5.

### 7.4 `GET /api/simulation/result/:sessionId`

- Existing keys remain.
- `result.findings_detail[]` gains `earned`, `evidence_log_indices`, `matched_log_indices`.
- `result` gains `evidence_penalty`, `noise_log_indices`, `selected_log_indices`.
- `scenario` continues to include `initial_logs` for the owner.

---

## 8. UI Contract

### 8.1 Detect & Analyze panel

- A checkbox column and whole-row click both toggle selection.
- Keyboard: focusable row; Space/Enter toggles.
- Selected rows get a distinct selected style (cyan/indigo ring or left border + subtle background) independent of severity styling and not confusable with the green "correct" color used after submission.
- A persistent **Selected Evidence** tray lists selected rows (timestamp, source, severity, truncated message) with individual remove controls and **Clear all**.
- Header counter: `Selected evidence: N logs`. Do **not** show `found / total findings` during the scenario; it leaks progress.
- Copy above the table: *"Select the logs that prove each finding. Each unrelated log costs 2 points. Correctness is revealed after submission."*
- No green/red correctness styling before submit. The API must not return correctness data during the session.
- Selection persists across KQL filter changes, sorting, tab changes, and refresh.
- Removing the current auto-discovery `useEffect` and the `discoveredFindings` state is mandatory.

### 8.2 Final Submission panel

- Replace "Evidence Found: N indicators" with "Evidence Selected: N logs".
- Zero selections are allowed. Show a warning: *"No evidence selected — Detect & Analysis will score 0."* Do not block submission.
- Submission remains one action; there is no per-click checkpoint.

### 8.3 Result page

- Each finding shows `Earned / Points` and `found`/`missed`.
- Found findings show the matched evidence row(s), including timestamp, source, and message.
- Missed findings show the acceptable evidence row(s) that would have proven them.
- A new **Evidence Review** panel lists every selected log with either:
  - the finding it proved, or
  - `Not evidence for any key finding — −2 pts`.
- Display the total `evidence_penalty` and net `findings_points`.
- Legacy sessions missing the new fields render the old findings section without the evidence panel.

---

## 9. Authoring & Migration

1. **Static seed**: explicit per-finding mappings for all 15 scenarios, authored by an LLM, validated, and approved by the user. See `workflows/scenario_evidence_authoring.md`. Delete `assign_evidence_indices()`; no chunking fallback.
2. **Validation gate**: reject any scenario where a finding lacks evidence, an index is out of range/duplicated, points are invalid, or the evidence union exceeds 50% of logs. Run it in the seed pipeline and on scenario creation.
3. **AI-generated scenarios**: 1–4 findings per scenario. For False Positive scenarios, a benign context log is mandatory; if the LLM omits it, the backend synthesizes one (approved change, HR travel, scheduled maintenance) and uses it as evidence. Each finding carries exact `evidence_log_messages`; the backend maps messages to indices. Retry once, then fall back to a single finding. Details in `workflows/scenario_evidence_authoring.md`.
4. **Existing database rows**: re-seed static scenarios. Legacy AI rows with one finding and indices `15..19` remain valid. Legacy FP rows with no benign context log are archived and regenerated; do not silently present an unsolvable FP case.
5. **Completed sessions are immutable**: never re-score them. Their stored `SkillGap` and rubric version remain as recorded.
6. **Rubric version**: new sessions use `rubric-v2-evidence-click`; old sessions keep `rubric-v1`.
7. **CI gate**: validate `scenarios_seed.json` on every PR; fail when any active scenario has an empty, out-of-range, duplicate, or over-50% evidence set.

---

## 10. Specialist Integration

Update `workflows/specialist_scenario_analysis.md` and the review console:

- **Summary row**: `findings_found_count / total_findings_count`, `evidence_selected_count`, `noise_selected_count`, `evidence_penalty`.
- **Detail modal**: a new **Evidence Review** section showing:
  - each selected log row (index, timestamp, source, severity, message),
  - whether it matched a finding or was noise,
  - per-finding found/missed and the acceptable evidence rows,
  - net findings points and penalty.
- **Timeline**: `select_evidence` actions appear with their toggle/clear payload and timestamp.
- **Backend**: `GetSubmissions` and `GetSubmissionDetail` replay `select_evidence`; for legacy sessions fall back to `submit_decision.discovered_findings`. `ExpectedFindings` remains available to Specialists only.

---

## 11. Verification Criteria (Definition of Done)

### 11.1 Evaluator unit tests

1. One selected index from `E_f` credits the full `points_f`.
2. No selected index from `E_f` credits 0 for that finding.
3. Two selected indices from the same `E_f` do not double-credit.
4. A noise selection subtracts 2 and floors `FindingsPoints` at 0.
5. A log in two `E_f` sets credits both findings and is not noise.
6. `EvidencePenalty == 2 × |noise|`; `NoiseLogIndices` and `SelectedLogIndices` are sorted.
7. Domain attribution: credit to finding domains; penalty to `scenario.domain_id`; per-domain floor at 0.
8. `TotalScore` clamps to `[0, 100]` and uses the existing raw-point denominator.
9. `FindingsDetail[].Earned` is the pre-penalty credit; `Found` is `Earned > 0`.

### 11.2 API and integration tests

1. Toggle → `GET /session` returns the same `selected_log_indices`.
2. Clear → returns `[]`; refresh stays `[]`.
3. Submit scores from replayed actions, not from a forged client payload.
4. Toggle after completion is rejected; invalid indices are rejected; index `0` is accepted.
5. Concurrent toggles do not collide on `(session_id, step_order)`.
6. Result exposes `evidence_penalty`, `noise_log_indices`, `selected_log_indices`, and per-finding evidence.
7. An old session reconstructed from `discovered_findings` still renders.

### 11.3 Frontend

1. Row click, checkbox, Space, and Enter all toggle.
2. Selection survives filtering, sorting, tab switches, and refresh.
3. No correctness styling or correctness data before submit.
4. Zero-selection warning appears; submission is not blocked.
5. Result page names missed evidence and each noise penalty.
6. `__log_index` is not shown as a table column and does not affect plain-text search.

### 11.4 Seeds and authoring

1. All 15 static scenarios pass evidence validation.
2. No scenario ships empty, out-of-range, duplicated, or over-50% evidence.
3. Every FP scenario has a benign context log in at least one verdict-supporting finding's evidence.
4. `scenarios_seed.json` is regenerated and `load_scenarios_to_db.py` succeeds.
5. CI rejects an intentionally broken fixture.

### 11.5 Build

```bash
go build -o /dev/null main.go
go test ./...
cd frontend && npm run build
python scripts/validate_evidence_mapping.py --mapping scripts/evidence_mapping.json
python scripts/test_simulation_integration.py   # requires the docker stack
```

---

## 12. Backward Compatibility

- **Old sessions**: `ReconstructFromActions` keeps reading `submit_decision.discovered_findings` when no `select_evidence` actions exist. They are never re-scored.
- **Old clients**: `selected_log_indices` is optional in the submit body; the server uses replayed actions. `discovered_findings` is ignored.
- **Removed payload**: `sanitized_findings` is a breaking frontend contract change, shipped in the same release as the new selection UI.
- **Schema**: `select_evidence` is appended to the enum; no table columns are added.
- **Rubric**: old sessions retain `rubric-v1`; new sessions use `rubric-v2-evidence-click`.

---

## 13. Non-Goals

- No change to the TP/FP or Response scoring rubric.
- No scoring of KQL query count, query quality, or search history (audit-only).
- No partial credit within a single finding.
- No per-finding user mapping ("which finding does this log prove?"); the system maps selected logs.
- No AI or human checkpoint inside the learner's Detect & Analyze interaction; the only checkpoint is final submission.
- No re-scoring of completed sessions.
- No missed-finding-driven next-scenario selection in this change; that loop is recorded in `NOTES.md` as a future proposal.
