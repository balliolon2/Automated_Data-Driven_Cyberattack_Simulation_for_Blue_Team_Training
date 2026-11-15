# Workflow: Learning Theory Hub & Pre-Simulation Knowledge Gate (Issue #26)

## Objective

Provide a dedicated **Learning Theory Hub** (`/learn`) covering all 5 CompTIA Security+ (SY0-701) domains so learners can study foundational security concepts, SIEM telemetry indicators, True/False Positive triage criteria, and 3-phase incident response playbook principles before entering the Adaptive Scenario Loop.

The workflow enforces a strict, data-driven progression for learners:

$$\text{Pre-Test (30 Questions)} \xrightarrow{\text{unlocks}} \text{Learn Theory (5 Domains)} \xrightarrow{\text{5/5 completed}} \text{Adaptive Simulation Loop} \longrightarrow \text{Post-Test}$$

Once unlocked by completing the Pre-Test, the Learning Theory Hub remains permanently accessible for remediation throughout the Scenario Loop (including direct deep-links from `/evaluation` and `/simulation/result/:id`).

---

## 1. Loop Profile

- **Loop Name**: Learner Theory Study & Remediation Loop
- **Frequency / Cadence**:
  - **Initial Mandatory Pass**: Once after completing the Pre-Test (`5/5` domains marked completed required to unlock the first Simulation Scenario).
  - **Ongoing Remediation**: On-demand between simulation scenarios whenever a learner wants to review a weak domain (`< 70%` proficiency) or a missed finding/action.
- **Primary Actors**:
  - **Learner**: completes Pre-Test, studies all 5 domain theory modules, marks each domain as completed, and revisits weak domains during the Scenario Loop.
  - **Specialist / Admin**: can browse and inspect all 5 domain theory modules at any time without a Pre-Test requirement.
  - **Seed Pipeline (`scripts/load_theory_to_db.py`)**: validates and ingests the 5 authored Markdown theory files (`theory/domain1_*.md` – `theory/domain5_*.md`) into PostgreSQL.

---

## 2. Trigger & Checkpoints

- **Unlock Trigger**: Completion of the 30-question Pre-Test (`POST /api/exams/submit-answer` where `completed == true`, creating the 5 `user_skill_profiles` rows).
- **Content Authoring Checkpoint (Push-Right)**:
  - Before seeding or wiring the UI, all 5 domain Markdown files (`theory/domain1_*.md` to `theory/domain5_*.md`) and their source registrations in `docs/REFERENCES.md` + `docs/references.bib` are compiled so the human owner can inspect the curriculum.
- **Learner Completion Checkpoints**:
  - One explicit **"Mark Domain as Completed"** checkpoint at the end of each of the 5 domain theory modules (`domain1` through `domain5`).
  - Completing all 5 domains (`5/5`) unlocks `POST /api/simulation/start`.

---

## 3. Canonical Terms

| Term | Meaning |
|---|---|
| **Theory Module (บทเรียนทฤษฎีรายโดเมน)** | The structured English curriculum document for one of the 5 CompTIA Security+ domains (`domain1`–`domain5`), combining exam-prep theory and Blue Team operational guidance. |
| **Theory Gate (เกตปลดล็อกการเรียนทฤษฎี)** | The two-stage access rule: (1) Pre-Test completion unlocks `/learn`, and (2) completing all 5 Theory Modules unlocks `POST /api/simulation/start`. |
| **Domain Theory Completion (การเรียนจบรายโดเมน)** | A persisted record in `user_theory_progress` indicating that a learner has finished studying a specific domain's Theory Module. |
| **Recommended Focus Domain (โดเมนที่แนะนำให้เน้น)** | Any domain where the learner's current `proficiency_score < 70.0%` (or the lowest-scoring domain), highlighted on `/learn` and `/evaluation`. |

---

## 4. Curriculum Authoring & Source Provenance Contract

### 4.1 Source Registry (`docs/REFERENCES.md` & `docs/references.bib`)

Every external reference used to compile the 5 domain theory files must be registered in `docs/REFERENCES.md` (with permanent `R-ID` keys `R2`, `R3`, …) and `docs/references.bib` **before** running the ingestion script:

- **`[R2]`** CompTIA Security+ (SY0-701) Exam Objectives / Official Domain Blueprint
- **`[R3]`** NIST Special Publication 800-61 Rev. 2 (*Computer Security Incident Handling Guide*) & NIST SP 800-53 Rev. 5 (*Security and Privacy Controls*)
- **`[R4]`** MITRE ATT&CK® Enterprise Matrix (Tactics, Techniques, and Detection Telemetry)
- **`[R5]`** OWASP Top 10 & CISA / Wikipedia Computer Security Reference Articles

### 4.2 Repository Markdown Files (`theory/`)

Five canonical English Markdown files live under `theory/` at the repository root:

| Domain ID | Domain Name | File Path |
|---|---|---|
| `domain1` | General Security Concepts | `theory/domain1_security_concepts.md` |
| `domain2` | Threats, Vulnerabilities, and Mitigations | `theory/domain2_threats_vulnerabilities.md` |
| `domain3` | Security Architecture | `theory/domain3_security_architecture.md` |
| `domain4` | Security Operations | `theory/domain4_security_operations.md` |
| `domain5` | Security Program Management and Oversight | `theory/domain5_security_management.md` |

### 4.3 Required Structure Inside Every Domain Markdown File

Each file in `theory/domain{1..5}_*.md` MUST be written in **English** and include YAML-style or structured header metadata followed by 4 mandatory sections:

1. **Section 1 — Core Security+ Theory & Exam Concepts**:
   - Comprehensive explanation of the domain's foundational principles, terminology, frameworks, and architectural/governance concepts tested in the 30-question Pre-Test and Post-Test.
2. **Section 2 — SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping**:
   - Concrete examples of log sources (SIEM, Windows Event IDs / Sysmon, Linux auth/syslog, Firewall, EDR, CloudTrail/IAM), key fields to inspect, KQL query patterns, and mapped MITRE ATT&CK techniques relevant to that domain.
3. **Section 3 — Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide**:
   - How to distinguish **True Positives (TP)** from **False Positives (FP)** (including benign context logs such as change tickets, HR travel approvals, and authorized vulnerability scans).
   - How to select proportional **Containment**, **Eradication**, and **Recovery** actions while avoiding destructive or premature distractors (`-2` to `-3` penalty traps).
4. **Section 4 — Key Takeaways & References**:
   - Quick-review checklist before entering a simulation scenario, plus inline `[R2]`–`[R5]` citations.

---

## 5. Database & Seeding Contract

### 5.1 Schema (`init.sql` & `models/migrate.go`)

1. **Theory Content Storage (`rag_documents` + `theory_modules` / embedded fallback)**:
   - Store the 5 domain theory documents in `rag_documents` (existing table in `init.sql`: `doc_id`, `title`, `content`, `domain_id`, `source_url`, `mitre_ref`, `updated_at`) so both the `/api/learn` endpoint and RAG services share one canonical table in PostgreSQL.
2. **Learner Progress Table (`user_theory_progress`)**:
   - Added to `init.sql` and auto-migrated idempotently in `models.RunMigrations(db)`:
     ```sql
     CREATE TABLE IF NOT EXISTS user_theory_progress (
       user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
       domain_id VARCHAR REFERENCES security_domains(domain_id) ON DELETE CASCADE,
       completed_at TIMESTAMP DEFAULT now(),
       PRIMARY KEY (user_id, domain_id)
     );
     CREATE INDEX IF NOT EXISTS idx_user_theory_progress_user ON user_theory_progress(user_id);
     ```

### 5.2 Seeding Pipeline (`scripts/load_theory_to_db.py` & `docker-compose.yml`)

- `scripts/load_theory_to_db.py`:
  - Reads `theory/domain1_*.md` through `theory/domain5_*.md`.
  - Validates that all 5 domains exist, are non-empty, and contain the required sections.
  - Upserts each domain's theory module into `rag_documents` keyed by `(domain_id, title)`.
- `docker-compose.yml`:
  - Mounts `./theory:/app/theory:ro` into the `seed` container and runs `python /app/scripts/load_theory_to_db.py` after `load_exams_to_db.py`.
- **Backend Embedded Fallback**:
  - Go backend also embeds `theory/*.md` (or loads from disk/DB) so unit/integration tests and newly initialized environments never fail with empty theory content even if `rag_documents` has not been seeded yet.

---

## 6. API Contract

### 6.1 `GET /api/learn/domains` (Authenticated: Bearer JWT)

Returns all 5 domain theory modules, the learner's current `proficiency_score` per domain, and completion status.

- **Access Gate**:
  - If the caller's role is `learner` and they have not completed the Pre-Test (`len(user_skill_profiles) == 0`), return `403 Forbidden`:
    ```json
    {
      "error": "Please complete the Pre-Test first to unlock Learning Theory.",
      "reason": "pre_test_required",
      "unlocked": false
    }
    ```
  - If the caller's role is `specialist` or `admin`, or if the `learner` has completed the Pre-Test (`len(user_skill_profiles) > 0`), return `200 OK`:
    ```json
    {
      "unlocked": true,
      "theory_completed": false,
      "completed_count": 2,
      "total_domains": 5,
      "completed_domains": ["domain1", "domain2"],
      "domains": [
        {
          "domain_id": "domain1",
          "domain_name": "General Security Concepts",
          "title": "Domain 1: General Security Concepts",
          "content": "# Domain 1: General Security Concepts\n...",
          "source_url": "docs/REFERENCES.md#R2-R5",
          "proficiency_score": 71.4,
          "threshold": 70.0,
          "passed_threshold": true,
          "recommended_focus": false,
          "is_completed": true,
          "completed_at": "2026-11-15T10:30:00Z"
        }
      ]
    }
    ```
  - Note: `recommended_focus` is `true` when `proficiency_score < 70.0` (or if all domains are `>= 70.0`, on the single lowest-scoring domain).

### 6.2 `POST /api/learn/domains/:domainId/complete` (Authenticated: Bearer JWT)

Marks a specific domain (`domain1`..`domain5`) as completed for the authenticated user.

- **Validation**:
  - `domainId` must be one of `domain1`, `domain2`, `domain3`, `domain4`, `domain5` (`400 Bad Request` otherwise).
  - For a `learner`, Pre-Test must already be completed (`403 Forbidden` with `reason: "pre_test_required"` if not).
- **Behavior**:
  - Idempotently inserts `(user_id, domain_id, now())` into `user_theory_progress` (`ON CONFLICT (user_id, domain_id) DO NOTHING`).
  - Does **not** modify `user_skill_profiles.proficiency_score`.
- **Response `200 OK`**:
  ```json
  {
    "domain_id": "domain1",
    "is_completed": true,
    "completed_domains": ["domain1", "domain2", "domain3", "domain4", "domain5"],
    "completed_count": 5,
    "total_domains": 5,
    "theory_completed": true
  }
  ```

### 6.3 Simulation Gate Integration (`GET /api/simulation/status` & `POST /api/simulation/start`)

1. **`GET /api/simulation/status`**:
   - Adds fields to `SimulationStatusResponse`:
     - `theory_completed`: `bool` (`len(completed_domains) >= 5`)
     - `completed_theory_domains`: `[]string`
2. **`POST /api/simulation/start`**:
   - When starting a new simulation session (and `os.Getenv("BYPASS_THEORY_GATE") != "true"`), verify that the learner has completed all 5 domains in `user_theory_progress` (`count == 5`).
   - If `count < 5` (and no active `in_progress` simulation session already exists), return `403 Forbidden`:
     ```json
     {
       "error": "Please complete all 5 Learning Theory domains before starting simulation scenarios.",
       "reason": "theory_incomplete",
       "completed_count": 2,
       "total_domains": 5,
       "completed_domains": ["domain1", "domain2"]
     }
     ```
3. **`GET /api/analytics/research-summary`**:
   - Includes `theory_completed` (`bool`), `completed_theory_domains` (`[]string`), and `theory_completed_at` (`*time.Time`, the max `completed_at` across the 5 domains when all 5 are done).

---

## 7. Frontend UI & Navigation Contract

### 7.1 Routes & Navbar (`App.tsx` & `Layout.tsx`)

- Add route `/learn` rendering `LearnTheoryPage`.
- Add `{ path: "/learn", label: "Learn" }` to `Layout.tsx` `navLinks` between `Exam` and `Evaluation` (`Exam` | `Learn` | `Evaluation` | `Discussions`).

### 7.2 `LearnTheoryPage` (`/learn`)

- **Dark Mode Graphite Theme**: Uses `bg-graphite-950` / `bg-graphite-900`, `text-graphite-50`, and `font-mono` for log examples, KQL queries, MITRE IDs, and metrics.
- **Locked State (Pre-Test Not Completed)**:
  - Displays a clear lock card explaining that the 30-question Diagnostic Pre-Test must be completed first to establish baseline domain proficiency and unlock the Learning Theory Hub, with a primary button **"Take Pre-Test"** (`/pre-test`).
- **Unlocked State**:
  - **Top Progress Banner**:
    - Shows overall reading progress (`X / 5 Domains Completed` progress bar).
    - Once `5 / 5` is reached, displays an unlocked CTA button **"Start Adaptive Simulation"** (`/simulation`).
  - **Domain Sidebar / Tabs (`domain1` – `domain5`)**:
    - Supports deep-linking via `/learn?domain=domain4`.
    - Shows each domain's title, completion badge (`Completed` checkmark vs. `Unread`), current `Proficiency Score` (`XX%`), and a **"Recommended Focus"** badge if `proficiency_score < 70%`.
  - **Main Theory Reader Panel**:
    - Renders the selected domain's Markdown curriculum cleanly with formatted headings, tables, KQL/log code blocks (`font-mono`), and callouts.
    - Bottom action bar:
      - **"Mark Domain as Completed"** button (or **"Completed ✓ — Next Domain →"** once marked).
      - Quick navigation buttons (**Previous Domain** / **Next Domain**).

### 7.3 Integration with `EvaluationPage` (`/evaluation`) & `ScenarioResultPage` (`/simulation/result/:id`)

- **`EvaluationPage.tsx`**:
  - Each domain row or card includes a **"Study Theory"** link (`/learn?domain=domainX`).
  - Bottom primary CTA button:
    - If `theory_completed == false`: shows **"Study Learning Theory First (X/5 Completed)"** navigating to `/learn`.
    - If `theory_completed == true` and `needsTraining == true`: shows **"Start Recommended Scenarios"**.
- **`ScenarioResultPage.tsx`**:
  - Includes a **"Review Domain Theory"** button linking to `/learn?domain={scenario.domain_id}` so a learner who missed findings or playbook steps can immediately review that domain's theory before starting the next scenario.

---

## 8. Verification Criteria (Definition of Done)

1. **Provenance & Markdown Curriculum**:
   - `docs/REFERENCES.md` and `docs/references.bib` register the authoritative theory references (`R2`–`R5`).
   - All 5 Markdown files (`theory/domain1_*.md` – `theory/domain5_*.md`) exist in English and contain all 4 required sections.
   - `python scripts/load_theory_to_db.py` validates and seeds all 5 domains idempotently.
2. **Gate Enforcement**:
   - A newly registered learner (before Pre-Test) gets `403` (`reason: "pre_test_required"`) on `GET /api/learn/domains` and `POST /api/learn/domains/:domainId/complete`.
   - After completing Pre-Test, `GET /api/learn/domains` returns `200 OK` with all 5 domains and the learner's proficiency scores.
   - Before completing all 5 domains in `/learn`, `POST /api/simulation/start` returns `403` (`reason: "theory_incomplete"`).
   - After marking `domain1`..`domain5` completed, `POST /api/simulation/start` succeeds (`201 Created`).
   - Completing theory modules does **not** alter `user_skill_profiles.proficiency_score`.
3. **Build & Integration Tests**:
   - `go build -o /dev/null main.go` and `go test ./...` pass.
   - `cd frontend && npm run build` passes with zero TypeScript errors.
   - `scripts/test_integration.py` and `scripts/test_simulation_integration.py` pass end-to-end.
