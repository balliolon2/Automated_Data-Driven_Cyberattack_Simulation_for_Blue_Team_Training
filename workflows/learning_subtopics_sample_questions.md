# Workflow Spec: Domain Overviews, Sub-Topics, and Formative Sample Exam Questions (Issue #30)

**GitHub Issue:** [#30 — feat(learn): Domain Overviews, Sub-Topics, and Formative Sample Exam Questions](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/30)
**Status:** Interview Settled (`/loop-me` Rounds 1–3 Complete) — Ready for `/to-spec` & Implementation
**Extends:** Spec 8 ([`workflows/learning_theory_hub.md`](learning_theory_hub.md))

---

## 1. Goal & Educational Rationale

While Spec 8 established the two-stage Theory Gate (`Pre-Test → Learn → Simulation → Post-Test`) backed by 5 domain-level Markdown modules, each module currently serves as a high-level summary. Learners benefit from granular, structured study aligned 1:1 with the **28 official CompTIA Security+ (SY0-701) exam objectives** (`1.1`–`5.6`), paired with **formative self-check sample questions** drawn directly from unused items in the 540-question bank.

This workflow deepens `/learn` into a hierarchical **Domain Overview → 28 Sub-Topics → 2 Formative Sample Questions** experience while enforcing strict **research integrity**:
1. Every sample question seen during `/learn` is permanently locked per `(user_id, sub_topic_id)` and **excluded from that learner's `Post-Test`** alongside `Pre-Test` questions (**Zero Overlap across Pre-Test ∪ Learn ∪ Post-Test**).
2. Sample question attempts do **not** modify `user_skill_profiles.proficiency_score` (`ADR-0004` proficiency invariance).
3. All Sub-topic theory prose is grounded exclusively in **trustworthy public websites** (Wikipedia, NIST CSRC, CISA, MITRE ATT&CK, OWASP, CMU ISO, Stanford UIT, Microsoft Learn — never textbooks) with clickable `https://` links and registered provenance in `docs/REFERENCES.md` + `docs/references.bib`.

---

## 2. Canonical Taxonomy: 28 CompTIA Security+ (SY0-701) Sub-Topics

The 5 existing domain files (`theory/domain1_security_concepts.md` through `theory/domain5_security_management.md`) remain **Domain Overviews** in `rag_documents`. All **28 Sub-topics** live as Markdown files under `theory/domain{1..5}/` and follow the official SY0-701 blueprint:

| Domain (`domain_id`) | Sub-Topic (`sub_topic_id`) | Order | Official SY0-701 Objective Title | Markdown File Path (`theory/...`) |
| --- | --- | --- | --- | --- |
| `domain1` | `1.1` | 1 | Compare and contrast various types of security controls | `domain1/1_1_security_controls.md` |
| `domain1` | `1.2` | 2 | Summarize fundamental security concepts | `domain1/1_2_fundamental_security_concepts.md` |
| `domain1` | `1.3` | 3 | Explain change management best practices | `domain1/1_3_change_management.md` |
| `domain1` | `1.4` | 4 | Explain the importance of using appropriate cryptographic solutions | `domain1/1_4_cryptographic_solutions.md` |
| `domain2` | `2.1` | 1 | Compare and contrast threat actors and intelligence sources | `domain2/2_1_threat_actors_and_motivations.md` |
| `domain2` | `2.2` | 2 | Analyze common threat vectors and attack surfaces | `domain2/2_2_threat_vectors_and_attack_surfaces.md` |
| `domain2` | `2.3` | 3 | Explain various types of vulnerabilities | `domain2/2_3_types_of_vulnerabilities.md` |
| `domain2` | `2.4` | 4 | Analyze indicators of malicious activity | `domain2/2_4_indicators_of_malicious_activity.md` |
| `domain2` | `2.5` | 5 | Explain mitigation techniques used to secure the enterprise | `domain2/2_5_mitigation_techniques.md` |
| `domain3` | `3.1` | 1 | Compare and contrast security architecture models | `domain3/3_1_security_architecture_models.md` |
| `domain3` | `3.2` | 2 | Apply security principles to secure enterprise infrastructure | `domain3/3_2_enterprise_infrastructure_security.md` |
| `domain3` | `3.3` | 3 | Compare and contrast concepts and strategies to protect data | `domain3/3_3_data_protection_strategies.md` |
| `domain3` | `3.4` | 4 | Explain the importance of resilience and recovery in security architecture | `domain3/3_4_resilience_and_recovery.md` |
| `domain4` | `4.1` | 1 | Apply common security techniques to computing resources | `domain4/4_1_computing_resource_security.md` |
| `domain4` | `4.2` | 2 | Explain security implications of hardware, software, and data asset management | `domain4/4_2_asset_management.md` |
| `domain4` | `4.3` | 3 | Explain various activities associated with vulnerability management | `domain4/4_3_vulnerability_management.md` |
| `domain4` | `4.4` | 4 | Explain security alerting and monitoring concepts and tools | `domain4/4_4_security_alerting_and_monitoring.md` |
| `domain4` | `4.5` | 5 | Modify enterprise capabilities to enhance security | `domain4/4_5_modifying_enterprise_capabilities.md` |
| `domain4` | `4.6` | 6 | Implement and maintain identity and access management | `domain4/4_6_identity_and_access_management.md` |
| `domain4` | `4.7` | 7 | Explain the importance of automation and orchestration | `domain4/4_7_automation_and_orchestration.md` |
| `domain4` | `4.8` | 8 | Explain appropriate incident response activities | `domain4/4_8_incident_response_activities.md` |
| `domain4` | `4.9` | 9 | Use data sources to support an investigation and digital forensics | `domain4/4_9_data_sources_and_digital_forensics.md` |
| `domain5` | `5.1` | 1 | Summarize elements of effective security governance | `domain5/5_1_security_governance.md` |
| `domain5` | `5.2` | 2 | Explain elements of the risk management process | `domain5/5_2_risk_management_process.md` |
| `domain5` | `5.3` | 3 | Explain third-party risk assessment and management processes | `domain5/5_3_third_party_risk_management.md` |
| `domain5` | `5.4` | 4 | Summarize elements of effective security compliance | `domain5/5_4_security_compliance.md` |
| `domain5` | `5.5` | 5 | Explain types and purposes of audits and assessments | `domain5/5_5_audits_and_assessments.md` |
| `domain5` | `5.6` | 6 | Implement security awareness practices | `domain5/5_6_security_awareness_practices.md` |

**Total:** `4 + 5 + 4 + 9 + 6 = 28 Sub-topics`.

---

## 3. Content Architecture, Storage & Provenance

### 3.1 Markdown Structure (28 Sub-Topic Files)
Each Sub-topic Markdown file in `theory/domain{1..5}/*.md` must contain:
1. `# <sub_topic_id> <Objective Title>`
2. `## 1. Core Security+ Theory` — concise, factual conceptual breakdown aligned with SY0-701 terminology.
3. `## 2. SOC & Blue-Team Operational Practice` — applied blue-team telemetry, KQL/SIEM log patterns, MITRE ATT&CK technique context, or defensive procedures relevant to the sub-topic.
4. `## 3. References (Web Sources)` — direct clickable `https://` links to authoritative public websites only (Wikipedia, NIST, CISA, MITRE ATT&CK, OWASP, CMU ISO, Stanford UIT, Microsoft Learn). No textbook page citations or unverified URLs.

### 3.2 Embedded Go FS + PostgreSQL Seeding
- `main.go` continues embedding `//go:embed theory`, which embeds nested `theory/domain{1..5}/*.md` subdirectories recursively.
- `scripts/load_theory_to_db.py` is extended to:
  1. Keep upserting the 5 **Domain Overviews** into `rag_documents` (`doc_type = 'theory'`, `domain_id = domain1..domain5`).
  2. Upsert all **28 Sub-topics** into `theory_subtopics` (`ON CONFLICT (sub_topic_id) DO UPDATE`), validating required Markdown headers and `https://` `source_url`.
- `services/theory_service.go` maintains canonical ordering `OrderedTheorySubtopics` (28 entries) and loads Sub-topics from PostgreSQL (`theory_subtopics`) with transparent fallback to the embedded `theory/` FS.
- All web references used in Sub-topic files must be registered in [`docs/REFERENCES.md`](../docs/REFERENCES.md) and [`docs/references.bib`](../docs/references.bib) before seeding.

---

## 4. Database Schema (`init.sql` & `models/migrate.go`)

Both `init.sql` and `models/migrate.go` define the schema changes cleanly:

### 4.1 Add `sub_topic_id` to `questions`
```sql
ALTER TABLE questions ADD COLUMN IF NOT EXISTS sub_topic_id VARCHAR(16);
CREATE INDEX IF NOT EXISTS idx_questions_domain_subtopic
  ON questions(domain_id, sub_topic_id) WHERE is_active = true;
```

### 4.2 `theory_subtopics`
```sql
CREATE TABLE IF NOT EXISTS theory_subtopics (
  sub_topic_id VARCHAR(16) PRIMARY KEY,
  domain_id VARCHAR REFERENCES security_domains(domain_id) ON DELETE CASCADE NOT NULL,
  order_index INT NOT NULL,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  source_url TEXT,
  updated_at TIMESTAMP DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_theory_subtopics_domain ON theory_subtopics(domain_id, order_index);
```

### 4.3 `user_theory_subtopic_progress`
```sql
CREATE TABLE IF NOT EXISTS user_theory_subtopic_progress (
  user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
  sub_topic_id VARCHAR(16) REFERENCES theory_subtopics(sub_topic_id) ON DELETE CASCADE,
  completed_at TIMESTAMP DEFAULT now(),
  PRIMARY KEY (user_id, sub_topic_id)
);
CREATE INDEX IF NOT EXISTS idx_user_subtopic_progress_user ON user_theory_subtopic_progress(user_id);
```

### 4.4 `user_theory_sample_questions`
```sql
CREATE TABLE IF NOT EXISTS user_theory_sample_questions (
  user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
  sub_topic_id VARCHAR(16) REFERENCES theory_subtopics(sub_topic_id) ON DELETE CASCADE,
  question_id UUID REFERENCES questions(question_id) ON DELETE CASCADE,
  order_index INT NOT NULL,
  is_fallback BOOLEAN NOT NULL DEFAULT false,
  submitted_answer VARCHAR(10),
  is_correct BOOLEAN,
  assigned_at TIMESTAMP DEFAULT now(),
  answered_at TIMESTAMP,
  PRIMARY KEY (user_id, sub_topic_id, question_id),
  UNIQUE (user_id, question_id)
);
CREATE INDEX IF NOT EXISTS idx_user_theory_sample_q_user ON user_theory_sample_questions(user_id);
```
> **Why `UNIQUE (user_id, question_id)`?** Guarantees that a fallback draw in one sub-topic can never assign the same `question_id` twice to the same user in another sub-topic.

---

## 5. Question Classification & Pool Safety (`540` Question Bank)

### 5.1 Tagging All 540 Questions with `sub_topic_id`
- Extend `parsed_questions.json` / `scripts/parse_exams.py` (with an explicit deterministic sub-topic mapping script/table reviewed for accuracy across all 540 questions) so every question carries `"sub_topic_id"` (`"1.1"`–`"5.6"`) matching its parent `domain_id`.
- `scripts/load_exams_to_db.py`:
  - Matches existing rows by `(domain_id, normalized question_text)` and executes `UPDATE ... SET sub_topic_id = %s` so **existing `question_id` UUIDs are preserved** across re-seeds.
  - Sets `sub_topic_id` on `INSERT` for new deployments.
  - Fails fast if any active question in `parsed_questions.json` has an empty or mismatched `sub_topic_id` (e.g., `domain1` question with `2.1`).

### 5.2 Sample Pool Math & Zero-Overlap Post-Test Guarantee
Audit verified `540` active `multiple_choice` (`A–D`) questions across Domains 1–5 (`110 / 110 / 90 / 150 / 80`):
- **Total consumed before Post-Test (Pre-Test + 28 × 2 Learn Samples)**:
  - Domain 1: `4 pre + 8 learn = 12` → **98 remaining** (needs 4 post)
  - Domain 2: `7 pre + 10 learn = 17` → **93 remaining** (needs 7 post)
  - Domain 3: `5 pre + 8 learn = 13` → **77 remaining** (needs 5 post)
  - Domain 4: `8 pre + 18 learn = 26` → **124 remaining** (needs 8 post)
  - Domain 5: `6 pre + 12 learn = 18` → **62 remaining** (needs 6 post)

### 5.3 Sample Allocation & Fallback Algorithm (`GetSubTopicDetail`)
Because empirical inspection showed a few SY0-701 objectives (`5.6`, `4.8`) have fewer dedicated chapter questions (~3–7 items), opening a Sub-topic (`learner` role) runs inside a transaction with concurrency safety:
1. Check existing rows in `user_theory_sample_questions` for `(user_id, sub_topic_id)`. If present, return them ordered by `order_index`.
2. Build exclusion subquery for `user_id`:
   - `Pre-Test`: `exam_sessions (user_id, exam_type = 'pre') ⋈ exam_session_questions`
   - `Learn Seen`: `user_theory_sample_questions (user_id)`
3. **Primary Draw:** Select up to `2` active questions where `domain_id = subtopic.domain_id AND sub_topic_id = :subTopicId AND question_id NOT IN (exclusion)` (`ORDER BY RANDOM() LIMIT 2`), tagged `is_fallback = false`.
4. **Domain Fallback Draw:** If fewer than `2` questions were found, draw `2 - count` active questions where `domain_id = subtopic.domain_id AND question_id NOT IN (exclusion ∪ primary_drawn)` (`ORDER BY RANDOM() LIMIT ...`), tagged `is_fallback = true`.
5. Insert assigned questions immediately (`ON CONFLICT DO NOTHING`, re-reading committed rows on concurrent request) so browser refreshes preserve the exact 2 questions.

### 5.4 Post-Test Zero Overlap (`controllers/exam_controller.go` → `StartPostTest`)
`StartPostTest` updates `excludeSubquery` so eligible questions satisfy:
```sql
domain_id = ? AND is_active = true
AND question_id NOT IN (
  SELECT esq.question_id
  FROM exam_session_questions esq
  JOIN exam_sessions es ON es.session_id = esq.session_id
  WHERE es.user_id = ? AND es.exam_type = 'pre'
  UNION
  SELECT utsq.question_id
  FROM user_theory_sample_questions utsq
  WHERE utsq.user_id = ?
)
```

---

## 6. State Transitions, Gate Enforcement & API Contracts

All `/api/learn/*` routes require `middlewares.AuthMiddleware()`.

### 6.1 Progression State Machine
```
Pre-Test Done ──▶ [GET /learn/domains] returns 5 Domain Overviews + 28 Sub-topic metadata
                       │
                       ▼
                 Open Sub-topic [GET /learn/subtopics/:subTopicId]
                 ──▶ Locks 2 unused sample questions for Learner (Preview 2 for Staff)
                       │
                       ▼
                 Attempt Question 1 & Question 2
                 [POST /learn/subtopics/:subTopicId/sample-questions/:questionId/attempt]
                 ──▶ Immediate graded feedback + explanation; locks 1st submitted answer
                       │
                       ▼  (when attempted_count == assigned_count [2/2])
                 Unlock & Click "Mark Sub-topic as Completed"
                 [POST /learn/subtopics/:subTopicId/complete]
                       ├─ If all Sub-topics in parent domain done ─▶ auto-insert user_theory_progress(domain_id)
                       └─ If 28/28 Sub-topics done ─────────────────▶ Stage 2 Gate Satisfied (unlocks Simulation)
```

### 6.2 `GET /api/learn/domains`
Returns the 5 Domain Overviews (with existing proficiency/recommended_focus fields) **plus** nested Sub-topic tree metadata (without loading heavy sample question payloads upfront):
```json
{
  "unlocked": true,
  "theory_completed": false,
  "completed_count": 1,
  "total_domains": 5,
  "completed_subtopic_count": 7,
  "total_subtopics": 28,
  "completed_domains": ["domain1"],
  "completed_subtopics": ["1.1", "1.2", "1.3", "1.4", "2.1", "2.2", "2.3"],
  "domains": [
    {
      "domain_id": "domain1",
      "domain_name": "General Security Concepts",
      "title": "Domain 1: General Security Concepts",
      "content": "# Domain 1 Overview...",
      "source_url": "https://csrc.nist.gov/pubs/sp/800/53/r5/final",
      "proficiency_score": 71.4,
      "threshold": 70.0,
      "passed_threshold": true,
      "recommended_focus": false,
      "is_completed": true,
      "completed_at": "2026-11-15T10:30:00Z",
      "completed_subtopics_in_domain": 4,
      "total_subtopics_in_domain": 4,
      "subtopics": [
        {
          "sub_topic_id": "1.1",
          "domain_id": "domain1",
          "order_index": 1,
          "title": "Compare and contrast various types of security controls",
          "is_completed": true,
          "completed_at": "2026-11-15T10:25:00Z",
          "sample_attempted_count": 2,
          "sample_total_count": 2
        }
      ]
    }
  ]
}
```
> **Note:** Omit full Sub-topic `content` from `GET /api/learn/domains` to keep tree hydration lightweight; fetch Sub-topic prose + sample questions via `GET /api/learn/subtopics/:subTopicId`.

### 6.3 `GET /api/learn/subtopics/:subTopicId`
Returns the Sub-topic Markdown `content`, clickable `source_url`, completion status, `previous_sub_topic_id` / `next_sub_topic_id` navigation pointers, and **2 sample questions**.
- Validates `subTopicId` (`400 Bad Request` if unknown) and Stage 1 Pre-Test gate (`403 pre_test_required` for Learners without Pre-Test).
- **Learner Payload Hygiene:**
  - While `submitted_answer == null`: **omit `correct_answer` and `explanation`** from the JSON payload so learners cannot inspect answers via DevTools before submitting.
  - Once `submitted_answer != null`: return `submitted_answer`, `is_correct`, `correct_answer`, `explanation`, `answered_at`.
- **Specialist / Admin Preview (`role != 'learner'`):**
  - Read-only preview: deterministically selects up to 2 active questions (`ORDER BY question_id ASC LIMIT 2` matching `sub_topic_id`, fallback to `domain_id`) **without** inserting rows into `user_theory_sample_questions` or `user_theory_subtopic_progress`. Returns `"preview_mode": true`.

### 6.4 `POST /api/learn/subtopics/:subTopicId/sample-questions/:questionId/attempt`
- Request Body: `{ "answer": "A" }` (`"A"|"B"|"C"|"D"` normalized uppercase; `400 Bad Request` otherwise).
- **Learner Behavior (Immediate Graded Lock):**
  - Verify `(user_id, sub_topic_id, question_id)` exists in `user_theory_sample_questions` (`404 Not Found` if not assigned to this user/subtopic).
  - If `submitted_answer IS NULL`: evaluate against `questions.correct_answer`, store `submitted_answer`, `is_correct`, `answered_at = now()`.
  - If `submitted_answer IS NOT NULL`: do **not** overwrite existing `submitted_answer` / `is_correct` (idempotent lock returning first attempt + `"already_submitted": true`).
- **Specialist / Admin Behavior (`preview_mode`):**
  - Looks up `question_id` against `questions`, grades answer statelessly, writes **nothing** to PostgreSQL, returns `explanation` + `correct_answer` with `"preview_mode": true`.
- **Response `200 OK`:**
  ```json
  {
    "sub_topic_id": "1.2",
    "question_id": "uuid",
    "submitted_answer": "B",
    "is_correct": true,
    "correct_answer": "B",
    "explanation": "Non-repudiation prevents a party from denying...",
    "attempted_count": 2,
    "required_count": 2,
    "ready_to_complete": true,
    "already_submitted": false,
    "preview_mode": false
  }
  ```

### 6.5 `POST /api/learn/subtopics/:subTopicId/complete`
- **Validation (Learner):**
  - Count rows in `user_theory_sample_questions` for `(user_id, sub_topic_id)` where `submitted_answer IS NOT NULL`.
  - If `attempted < assigned` (or `assigned == 0` because Sub-topic was never opened): return `400 Bad Request`:
    ```json
    {
      "error": "Please attempt all assigned sample questions before marking this sub-topic as completed.",
      "reason": "sample_questions_incomplete",
      "attempted_count": 1,
      "required_count": 2
    }
    ```
- **Behavior:**
  - Idempotently inserts `(user_id, sub_topic_id, now())` into `user_theory_subtopic_progress`.
  - Checks if all Sub-topics in `subtopic.domain_id` exist in `user_theory_subtopic_progress`. If all done, idempotently inserts `(user_id, domain_id, now())` into `user_theory_progress` and returns `parent_domain_completed: true`.
  - For Specialist/Admin preview: returns `200 OK` (`preview_mode: true`) without altering Learner completion tables.

### 6.6 `POST /api/learn/domains/:domainId/complete` (Anti-Bypass Guard)
- Retained for backward compatibility but hardened:
  - For Learners, queries `user_theory_subtopic_progress` for `domainId`.
  - If any Sub-topic in `domainId` is incomplete, returns `403 Forbidden`:
    ```json
    {
      "error": "All sub-topics in this domain must be completed before marking the domain complete.",
      "reason": "subtopics_incomplete",
      "missing_subtopics": ["1.3", "1.4"]
    }
    ```
  - Only succeeds when all Sub-topics in `domainId` are completed.

### 6.7 Stage 2 Simulation Gate & Research Analytics
1. **`GET /api/simulation/status`**:
   - Computes `theory_completed = (completed_subtopic_count == 28)`.
   - Keeps `completed_theory_domains` and adds `completed_subtopic_count` (`int`), `total_subtopics` (`28`), `completed_theory_subtopics` (`[]string`).
2. **`POST /api/simulation/start`**:
   - Enforces `completed_subtopic_count == 28` (unless `BYPASS_THEORY_GATE=true` or caller is `specialist`/`admin` or already has an `in_progress` simulation session).
   - Pre-existing learner accounts with Spec 8 domain rows but `< 28` sub-topics completed must finish the 28 Sub-topics before starting new simulation scenarios (`403 theory_incomplete` with `completed_subtopic_count`, `total_subtopics: 28`).
3. **`GET /api/analytics/research-summary`**:
   - Keeps top-level `theory_completed` (`bool`, true when 28/28 done), `completed_theory_domains`, `theory_completed_at`.
   - Adds `learn_subtopic_telemetry`:
     ```json
     {
       "completed_subtopics_count": 28,
       "total_subtopics": 28,
       "completed_subtopics": ["1.1", "..."],
       "sample_questions_assigned": 56,
       "sample_questions_attempted": 56,
       "sample_questions_correct": 44,
       "sample_accuracy_percent": 78.57,
       "domain_breakdown": {
         "domain1": { "completed_subtopics": 4, "total_subtopics": 4, "sample_attempted": 8, "sample_correct": 6 }
       }
     }
     ```

---

## 7. Frontend UI & UX Contract (`LearnTheoryPage.tsx`)

- **Dark Mode Graphite Theme** (`bg-graphite-950` / `bg-graphite-900`, `text-graphite-50`, `font-mono` for telemetry/logs/codes).
- **Progress Banner**:
  - Shows overall Sub-topic progress (`X / 28 Sub-topics Completed`) + domain completion pills (`Y / 5 Domains`).
  - Once `28 / 28`, highlights Stage 2 Gate unlocked with **"Proceed to Adaptive Simulation"** CTA (`/simulation`).
- **Expandable Nested Sidebar Tree**:
  - Top level: 5 Domains (`domain1`..`domain5`) showing proficiency score badge, `Recommended Focus` pill when `< 70%`, and `completed/total` sub-topic counter (e.g., `2/4`).
  - Expanding a Domain shows:
    - **Domain Overview** item (`topic=overview`).
    - Nested **Sub-topic** items (`1.1`..`1.4`, etc.) with status indicators (`✓` completed, `●` active, `○` unstarted).
- **URL Deep-Linking**:
  - `/learn?domain=domain1&topic=overview` or `/learn?domain=domain1&topic=1.2`.
  - Default selection on initial visit without params: first `recommended_focus` domain (`Overview`).
- **Domain Overview Panel**:
  - Renders Overview Markdown, clickable web Reference link, and a **Sub-topic Roadmap Grid** linking straight into each Sub-topic in that domain.
  - **No manual "Mark Domain as Completed" button** (completion happens automatically when all nested Sub-topics finish).
- **Sub-topic Panel**:
  - Renders Sub-topic Markdown + clickable web References + Sub-topic status badge (`Completed` vs `In Progress`).
  - **Formative Practice Section ("Check Your Understanding — Sample Exam Questions")**:
    - Renders the 2 assigned `multiple_choice` (`A–D`) cards.
    - Displays subtle **`Domain Related Fallback`** badge when `is_fallback === true`.
    - Each card has independent `A/B/C/D` radio selection + **"Submit Answer"** button -> immediately renders right/wrong status, highlights the correct choice, shows `explanation`, and locks the answer.
  - **Completion Bar & Sequential Navigation**:
    - **"Mark Sub-topic as Completed"** button — disabled with hint (`"Attempt 2/2 sample questions above to unlock"`) until both answers are submitted (`preview_mode` staff see `"Preview Mode (Read-Only)"` indicator).
    - Clicking **"Mark Sub-topic as Completed"** marks the Sub-topic done, shows toast notification if the parent Domain just completed (`Domain X Completed!`) or if `28/28` unlocked Simulation, and keeps **"Previous Sub-topic"** / **"Next Sub-topic"** buttons ready for smooth sequential reading across all 28 Sub-topics (`1.4 → 2.1`, etc.).

---

## 8. Staged Implementation & Verification Plan

1. **Checkpoint 1 — Curriculum & Classification First (Push-Right)**:
   - Register all new public web sources in `docs/REFERENCES.md` and `docs/references.bib`.
   - Author the 28 Sub-topic Markdown files (`theory/domain{1..5}/*.md`) with verified clickable `https://` links.
   - Classify all 540 questions in `parsed_questions.json` with verified `sub_topic_id` (`1.1`–`5.6`), checking per-subtopic counts cleanly.
2. **Checkpoint 2 — Database & Backend API**:
   - Update `init.sql`, `models/`, `models/migrate.go`, `scripts/load_exams_to_db.py`, `scripts/load_theory_to_db.py`.
   - Update `services/theory_service.go`, `controllers/learn_controller.go`, `controllers/exam_controller.go` (`StartPostTest` exclusion), `controllers/simulation_controller.go` (Stage 2 28-subtopic gate + `GetResearchSummary`), `dto/`, and `main.go` routes.
   - Unit tests: Sub-topic draw + Domain fallback, `UNIQUE` non-duplication, `StartPostTest` exclusion of `Pre-Test ∪ Learn Samples`, anti-bypass `POST /learn/domains/:id/complete` 403, Stage 2 gate at `27/28` vs `28/28`, and immediate graded lock idempotency.
3. **Checkpoint 3 — Frontend UI & E2E Integration**:
   - Upgrade `LearnTheoryPage.tsx` with nested sidebar tree, deep-linking, inline 2-question formative quiz with locked explanations, and `Previous/Next` navigation.
   - Verify `go build -o /dev/null main.go`, `go test ./...`, `npm run tsc`, `npm run build` in `frontend/`.
   - Extend `scripts/test_integration.py` to exercise the entire `Pre-Test → 28 Sub-topics × 2 Sample Questions → Domain Auto-Completion → Simulation Gate Unlock → Zero-Overlap Post-Test → Research Summary` pipeline end-to-end against PostgreSQL.
