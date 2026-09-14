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
     - Discovered Key Findings (which were found, which were missed).
     - Response Actions taken (Containment / Eradication / Recovery checks).
     - Chronological action timeline with step orders and timestamps.
3. **Authoring Interface**:
   - Specialist clicks "Draft Analysis for Scenario".
   - Split-pane Markdown editor with Live Preview.
   - **Quick-Insert Template Bar**:
     - *Insert Scenario Baseline*: Injects scenario title, attack narrative, and MITRE technique.
     - *Insert Key Findings Table*: Injects the expected findings checklist for educational demonstration.
     - *Insert Common Pitfalls*: Structured template for common missteps observed in learner submissions.
   - Code block syntax highlighting for KQL queries, syslog snippets, and event IDs.

---

## 5. Checkpoint (Publishing)

- **Actor**: Specialist author.
- **Action**: Specialist reviews preview and clicks **[Publish Thread]**.
- **Validation**:
  - Title non-empty (5–200 characters).
  - Content non-empty (minimum 50 characters).
  - Scenario reference valid (if linked).

---

## 6. Outcomes & Downstream Integration

1. **Community Board Feed**:
   - Thread appears in `/discussions` feed sorted by latest / most upvoted.
   - Tagged with scenario badge, author `[Specialist]` badge, and domain badge.
2. **Scenario Result Page Integration**:
   - When any learner completes the linked scenario and views `ScenarioResultPage`:
   - System queries `GET /api/threads?scenario_id=:id`.
   - A dedicated **"Specialist Insights & Community Discussions"** panel appears directly below their evaluation score and domain breakdown.
   - Learners can click directly into the thread to see how an expert approaches the incident and compare their choices.
