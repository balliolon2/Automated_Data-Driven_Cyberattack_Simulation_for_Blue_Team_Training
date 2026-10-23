# Workflow: Specialist Scenario Analysis

## Objective
Enable verified Specialists to inspect real learner simulation submissions (anonymized by Nickname), identify common triage errors and investigation pitfalls, and author authoritative Analysis Threads (กระทู้) linked to specific scenarios or broader Blue Team methodologies.

---

## 1. Loop Profile

- **Loop Name**: Specialist Analysis Authoring Loop
- **Frequency / Cadence**: Continuous / Expert-driven
- **Primary Actors**:
  - **Specialist**: Analyst with approved credentials authoring case breakdowns
  - **Learner**: Audience reviewing expert analysis after scenario attempts
  - **Admin**: Content supervisor

---

## 2. Trigger

- **Event**: Specialist navigates to the **Specialist Review Console** (`/specialist/reviews`) to inspect learner performance data or visits a Scenario details page to publish an instructional breakdown.

---

## 3. Data Schema & Requirements

### Analysis Threads Table (`analysis_threads`)
- `thread_id`: `UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `author_id`: `UUID REFERENCES users(user_id) ON DELETE CASCADE`
- `scenario_id`: `UUID REFERENCES scenarios(scenario_id) ON DELETE SET NULL` (nullable for general topics)
- `title`: `VARCHAR(255) NOT NULL`
- `content`: `TEXT NOT NULL` (Markdown format)
- `tags`: `VARCHAR[]` (e.g. `['credential-access', 'triage-pitfalls', 'false-positive']`)
- `upvote_count`: `INT DEFAULT 0`
- `view_count`: `INT DEFAULT 0`
- `is_pinned`: `BOOLEAN DEFAULT false`
- `is_locked`: `BOOLEAN DEFAULT false`
- `created_at`: `TIMESTAMP DEFAULT now()`
- `updated_at`: `TIMESTAMP DEFAULT now()`

### Thread Upvotes Table (`thread_upvotes`)
- `thread_id`: `UUID REFERENCES analysis_threads(thread_id) ON DELETE CASCADE`
- `user_id`: `UUID REFERENCES users(user_id) ON DELETE CASCADE`
- `created_at`: `TIMESTAMP DEFAULT now()`
- `PRIMARY KEY (thread_id, user_id)`

### 3.3 Evidence Review Data Contract (Issue #11)

No new table is required. The review console reads the learner's `select_evidence` session actions and the scenario's authored evidence mapping.

**Summary row (`SpecialistSubmissionSummary`) additions:**

```go
EvidenceSelectedCount int     `json:"evidence_selected_count"`
NoiseSelectedCount    int     `json:"noise_selected_count"`
EvidencePenalty       float64 `json:"evidence_penalty"`
FindingsFoundCount    int     `json:"findings_found_count"`
TotalFindingsCount    int     `json:"total_findings_count"`
```

**Detail (`SpecialistSubmissionDetail`) additions:**

```go
SelectedLogIndices []int               `json:"selected_log_indices"`
EvidenceFindings   []EvidenceFindingDTO `json:"evidence_findings"`
NoiseSelections    []EvidenceLogDTO     `json:"noise_selections"`
FindingsPoints     float64              `json:"findings_points"`
EvidencePenalty    float64              `json:"evidence_penalty"`
```

`EvidenceFindingDTO`:

```go
type EvidenceFindingDTO struct {
    ID                 string        `json:"id"`
    Description        string        `json:"description"`
    DomainID           string        `json:"domain_id"`
    Found              bool          `json:"found"`
    Points             int           `json:"points"`
    Earned             int           `json:"earned"`
    EvidenceLogIndices []int         `json:"evidence_log_indices"`
    MatchedLogIndices  []int         `json:"matched_log_indices"`
    EvidenceLogs       []EvidenceLogDTO `json:"evidence_logs"`
    Explanation        string        `json:"explanation"`
}
```

`EvidenceLogDTO`:

```go
type EvidenceLogDTO struct {
    Index     int    `json:"index"`
    Timestamp string `json:"timestamp"`
    Source    string `json:"source"`
    Severity  string `json:"severity"`
    Message   string `json:"message"`
}
```

Behavior:

- Replay `select_evidence` actions in `step_order` to rebuild the final selection set. Handle `{"operation":"toggle"}` and `{"operation":"clear"}` payloads.
- Compute summary counters from the replayed set `S`:
  - `TotalFindingsCount = len(key_findings)`.
  - `FindingsFoundCount = count(f where S ∩ E_f ≠ ∅)`.
  - `NoiseSelectedCount = |S \ ⋃ E_f|`.
  - `EvidencePenalty = 2 × NoiseSelectedCount`.
- For legacy sessions with no `select_evidence` actions, fall back to the `submit_decision.discovered_findings` payload and expose the legacy fields only. Never re-score legacy sessions.
- `ExpectedFindings` remains the full scenario `expected_outcomes` and is visible to Specialists only.
- The summary/detail endpoints must not expose evidence indices to Learners.

---

## 4. Pipeline Execution (Review & Data Aggregation)

1. **Submission Ingestion & Anonymization**:
   - Query completed records from `simulation_sessions` joined with `session_actions`, `scenarios`, and `users`.
   - Protect student privacy: Only display `users.nickname` (email and personal identity remain strictly hidden).
2. **Review Console Metrics**:
   - Filterable by:
     - **Scenario**: Specific incident scenario.
     - **Domain**: CompTIA Security+ Domains 1 through 5.
     - **Score Range**: Filter low scores (< 70%) to investigate struggle points.
   - Per-submission drill-down:
     - TP/FP Triage choice (`is_true_positive` vs learner's `user_choice`).
     - **Evidence Review** (new, Issue #11):
       - Selected evidence logs with index, timestamp, source, severity, and message.
       - Whether each selected log matched a key finding or was noise (`−2` evidence penalty).
       - Key findings found vs missed, including the acceptable evidence logs the learner failed to select.
       - Net `findings_points`, gross finding credit, and total `evidence_penalty`.
     - Response Actions taken (Containment / Eradication / Recovery checks).
     - Chronological action timeline with step orders and timestamps, including every `select_evidence` toggle and `clear` action.
3. **Authoring Interface**:
   - Specialist clicks "Draft Analysis for Scenario".
   - Split-pane Markdown editor with Live Preview.
   - **Quick-Insert Template Bar**:
     - *Insert Scenario Baseline*: Injects scenario title, attack narrative, and MITRE technique.
     - *Insert Key Findings Table*: Injects the expected findings checklist, including the evidence logs that prove each finding.
     - *Insert Evidence Review*: Injects a compact table of the learner's selected logs, matched findings, missed evidence, and noise picks with the `−2` penalty. Pulls from `EvidenceFindings` and `NoiseSelections`; never renders raw session payloads.
     - *Insert Common Pitfalls*: Structured template for common missteps observed in learner submissions, including selecting the high-severity alert as evidence in a False Positive case.
   - Code block syntax highlighting for KQL queries, syslog snippets, and event IDs.

---

## 5. Checkpoint (Publishing)

- **Actor**: Specialist author.
- **Action**: Specialist reviews preview and clicks **[Publish Thread]**.
- **Validation**:
  - Title non-empty (5–200 characters).
  - Content non-empty (minimum 50 characters).
  - Scenario reference valid (if linked).
  - If the author inserts an *Evidence Review*, every referenced finding/log index must exist in the linked scenario; otherwise the insert is replaced with a warning instead of leaking a broken reference.

---

## 6. Outcomes & Downstream Integration

1. **Community Board Feed**:
   - Thread appears in `/discussions` feed sorted by latest / most upvoted.
   - Tagged with scenario badge, author `[Specialist]` badge, and domain badge.
2. **Scenario Result Page Integration**:
   - When any learner completes the linked scenario and views `ScenarioResultPage`:
   - System queries `GET /api/threads?scenario_id=:id`.
   - A dedicated **"Specialist Insights & Community Discussions"** panel appears directly below their evaluation score, evidence review, and domain breakdown.
   - Learners can click directly into the thread to see how an expert approaches the incident and compare their choices, including which logs the expert treats as decisive evidence and which are red herrings.
