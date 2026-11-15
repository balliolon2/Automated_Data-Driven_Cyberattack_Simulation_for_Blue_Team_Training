---
title: "Spec 9: Domain Overviews, Sub-Topics, and Formative Sample Exam Questions (Issue #30)"
labels: ["ready-for-agent"]
status: "ready-for-agent"
---

# Spec 9: Domain Overviews, Sub-Topics, and Formative Sample Exam Questions

## Problem Statement

Following Spec 8, the Learning Theory Hub currently presents each of the five CompTIA Security+ (SY0-701) domains as a single summary page. While useful as an introductory overview, a single summary does not give Learners deep, structured coverage of individual exam objectives (`1.1` through `5.6`) or an immediate way to verify comprehension right after reading a concept. Without topic-level reinforcement, Learners must either jump straight from a broad overview into scored Simulation Sessions or guess whether they understood subtle distinctions (such as control categories, risk formulas, incident response phases, or chain-of-custody procedures). At the same time, introducing practice questions inside the learning hub risks leaking Post-Test questions unless every viewed sample question is systematically tracked and excluded from future diagnostic exams.

## Solution

Upgrade the Learning Theory Hub into a structured two-level curriculum hierarchy (**5 Domain Overviews → 28 SY0-701 Sub-Topics**) paired with locked **Formative Sample Questions** and strict **Pre-Test ∪ Learn ∪ Post-Test Zero-Overlap** isolation:

1. **Domain Overviews + 28 CompTIA Security+ (SY0-701) Sub-Topics:** Retain the existing five domain documents as high-level **Domain Overviews** linking conceptually to their nested Sub-topics, and author **28 granular Sub-topic modules** aligned 1:1 with official SY0-701 objectives (`1.1–1.4`, `2.1–2.5`, `3.1–3.4`, `4.1–4.9`, `5.1–5.6`). Every Sub-topic is grounded strictly in trustworthy public web sources (Wikipedia, NIST, CISA, MITRE ATT&CK, OWASP, CMU ISO, Stanford UIT, Microsoft Learn — never textbooks) with clickable `https://` links and provenance registration.
2. **2 Locked Formative Sample Questions per Sub-Topic (with Domain Fallback):** Classify all 540 multiple-choice exam-bank questions by `sub_topic_id`. When a Learner opens a Sub-topic for the first time, deterministically draw and persist **2 unused questions** (`is_fallback = false`) excluding the Learner's Pre-Test items and previously assigned Learn sample items; if fewer than 2 unused items remain in that `sub_topic_id`, fill remaining slots from unused items in the same parent `domain_id` (`is_fallback = true`) and surface a `Domain Related Fallback` indicator.
3. **Immediate Graded Feedback & Attempt-Gated Sub-Topic Completion:** Answers submitted on sample questions (`A–D`) are graded immediately against the bank, displaying right/wrong status and the full explanation while locking the first submitted attempt (`submitted_answer`, `is_correct`) to preserve clean formative telemetry and keep `proficiency_score` unchanged (ADR-0004). Attempting both (`2/2`) sample questions unlocks **"Mark Sub-topic as Completed"** regardless of correctness.
4. **Automatic Parent Domain Completion & 28/28 Stage 2 Gate:** Domain Overviews remain informational without a manual domain completion button; completing all Sub-topics in a domain automatically marks that parent domain completed, while direct API attempts to mark an incomplete domain complete are rejected. Starting Simulation Sessions requires completing all **28/28 Sub-topics**.
5. **Strict Zero-Overlap Post-Test Isolation:** Updating `StartPostTest` ensures every question assigned as a Learn sample question for a Learner is permanently excluded alongside Pre-Test questions when generating the Post-Test.
6. **Hierarchical Sidebar UI & Research Telemetry:** Present an expandable Nested Sidebar Tree with URL deep-linking (`/learn?domain=...&topic=...`), sequential `Previous / Next Sub-topic` navigation, Specialist/Admin read-only sample previews, and full formative telemetry inside the research analytics summary.

## User Stories

1. As a Learner who has not completed the Pre-Test, I want all Sub-topic modules and Domain Overviews to remain locked behind the Stage 1 Gate, so that my baseline proficiency is measured cleanly first.
2. As a Learner who opens the Learning Theory Hub after finishing the Pre-Test, I want to see the five domains rendered in an expandable nested sidebar tree, so that I can grasp the full 28-sub-topic structure at a glance.
3. As a Learner, I want each Domain entry in the sidebar tree to show my Proficiency Score, any Recommended Focus badge, and my Sub-topic completion progress (e.g., `2/4`), so that I know where to focus.
4. As a Learner visiting `/learn` without query parameters, I want the page to open on the Domain Overview of my primary Recommended Focus domain, so that I start where remediation matters most.
5. As a Learner viewing a Domain Overview, I want to read the high-level domain synthesis alongside an interactive Sub-topic checklist roadmap, so that I can jump directly into any nested Sub-topic without cluttering the overview with a premature completion button.
6. As a Learner, I want clickable URL deep-links (`/learn?domain=domain1&topic=1.2` or `topic=overview`), so that I can bookmark, refresh, or return directly from Evaluation or Simulation Results to the exact topic I need.
7. As a Learner studying Domain 1, I want dedicated Sub-topic modules covering `1.1` Security Controls, `1.2` Fundamental Security Concepts, `1.3` Change Management, and `1.4` Cryptographic Solutions, so that I can master Domain 1 objective by objective.
8. As a Learner studying Domain 2, I want dedicated Sub-topic modules covering `2.1` Threat Actors & Intelligence, `2.2` Threat Vectors & Attack Surfaces, `2.3` Types of Vulnerabilities, `2.4` Indicators of Malicious Activity, and `2.5` Enterprise Mitigation Techniques, so that I can connect threats to blue-team detection.
9. As a Learner studying Domain 3, I want dedicated Sub-topic modules covering `3.1` Security Architecture Models, `3.2` Enterprise Infrastructure Security, `3.3` Data Protection Strategies, and `3.4` Resilience & Recovery, so that I understand defensive architecture principles.
10. As a Learner studying Domain 4, I want dedicated Sub-topic modules covering `4.1` through `4.9` (Computing Resource Security, Asset Management, Vulnerability Management, Alerting & Monitoring, Modifying Enterprise Capabilities, IAM, Automation/SOAR, Incident Response, and Digital Forensics), so that every operational SY0-701 objective has concrete SOC guidance.
11. As a Learner studying Domain 5, I want dedicated Sub-topic modules covering `5.1` through `5.6` (Governance, Risk Management, Third-Party Risk, Compliance, Audits & Assessments, and Security Awareness), so that governance and compliance objectives are thoroughly covered.
12. As a Learner reading any Sub-topic module, I want direct clickable `https://` links to authoritative public websites (such as Wikipedia, NIST, CISA, MITRE ATT&CK, OWASP, CMU ISO, Stanford UIT, and Microsoft Learn), so that I can verify and explore concepts at primary sources.
13. As a Learner opening a Sub-topic module, I want to see two multiple-choice formative sample questions directly below the reading material, so that I can immediately apply what I just read.
14. As a Learner, I want the same two sample questions assigned to me for a Sub-topic to remain fixed across page refreshes and logins, so that refreshing the browser never swaps in new questions or drains the question bank.
15. As a Learner, I want none of my Pre-Test questions or previously assigned sample questions from other Sub-topics to reappear in the current Sub-topic's sample questions, so that every sample question provides fresh practice.
16. As a Learner viewing a sparse Sub-topic whose dedicated unused pool has fewer than two questions left, I want the remaining slot filled automatically from unused questions in the same parent domain and clearly tagged as `Domain Related Fallback`, so that I always receive two valid formative questions without breaking exclusivity.
17. As a Learner before submitting a sample question, I want the correct answer and explanation hidden from the API payload, so that I genuinely test my understanding first.
18. As a Learner, I want to submit each of the two sample questions independently and receive immediate right/wrong feedback, the highlighted correct option, and the full explanation, so that I learn why an option is right or wrong.
19. As a Learner, I want my first submitted answer on a sample question to be locked once submitted (while keeping the question and explanation visible for review), so that analytics reflect my true first-attempt comprehension.
20. As a Learner, I want the **Mark Sub-topic as Completed** button to stay disabled with a clear progress prompt until I have attempted both (`2/2`) sample questions for that Sub-topic, so that every Learner completes the formative check before progressing.
21. As a Learner who answered one or both sample questions incorrectly, I want **Mark Sub-topic as Completed** to still unlock once both questions have been attempted and their explanations reviewed, so that learning mistakes inform me rather than deadlocking my progress.
22. As a Learner, I want finishing the final remaining Sub-topic in a domain to automatically mark that parent domain complete and notify me, so that domain completion stays synchronized without extra clicks.
23. As a Learner, I want **Previous Sub-topic** and **Next Sub-topic** navigation buttons that step smoothly across domain boundaries (`1.4 → 2.1`, etc.), so that I can progress sequentially through all 28 modules.
24. As a Learner, I want my Sub-topic sample attempts and completions never to alter my `proficiency_score`, so that Skill Profile scores remain strictly governed by Pre-Test, Simulation, and Post-Test evaluations.
25. As a Learner who has completed fewer than all 28 Sub-topics (including legacy accounts that only marked domains complete prior to Sub-topics), I want starting a Simulation Session blocked with clear Sub-topic progress counts (`X/28`), so that every participant completes the full curriculum before entering the Scenario Loop.
26. As a Learner who has completed all 28 Sub-topics, I want an immediate **Proceed to Adaptive Simulation** call-to-action on `/learn` and unlocked Simulation access, so that I transition seamlessly into scenarios.
27. As a Learner taking the Post-Test after finishing Simulations, I want zero overlap between my Pre-Test questions, all sample questions I saw in `/learn`, and my Post-Test questions, so that my Post-Test score accurately measures knowledge transfer rather than question memorization.
28. As a Specialist or Admin opening any Sub-topic on `/learn`, I want to preview the Markdown curriculum and interact with read-only sample questions (graded statelessly with explanations shown) without writing Learner exclusion rows or progress telemetry, so that staff accounts never pollute study datasets.
29. As a Researcher inspecting the analytics summary endpoint, I want aggregate telemetry showing Sub-topics completed (`X/28`), total sample questions assigned and attempted, first-attempt sample accuracy percentage, and per-domain breakdowns, so that engagement with formative checks can be correlated with Post-Test learning gains.
30. As a Maintainer, I want all 540 bank questions validated against their parent domains and all 28 Sub-topic Markdown files embedded in the application binary alongside database seeding, so that deployments fail fast on misclassified content while serving reliably offline or in tests.

## Implementation Decisions

### Domain Model & Curriculum Structure

- **Domain Overview vs. Theory Sub-Topic**: The five existing domain modules remain **Domain Overviews** stored in `rag_documents`. Twenty-eight **Theory Sub-Topics** corresponding 1:1 to CompTIA Security+ (SY0-701) objectives (`1.1–1.4`, `2.1–2.5`, `3.1–3.4`, `4.1–4.9`, `5.1–5.6`) are stored under nested subdirectories inside the embedded theory directory and seeded into a new `theory_subtopics` relation (`sub_topic_id`, `domain_id`, `order_index`, `title`, `content`, `source_url`, `updated_at`).
- **Web-Only Source Governance**: Every Sub-topic document must follow a uniform three-part Markdown layout (`Core Security+ Theory`, `SOC & Blue-Team Operational Practice`, `References (Web Sources)`), cite only trustworthy public web domains via clickable `https://` links, and be registered in the project's canonical reference registry and BibTeX bibliography before seeding.

### Question Classification, Pool Isolation & Allocation

- **Bank Classification (`questions.sub_topic_id`)**: Add nullable/indexed `sub_topic_id VARCHAR(16)` to `questions` and classify all 540 `multiple_choice` (`A–D`) questions in the exam pipeline so every question maps cleanly to a valid `sub_topic_id` matching its `domain_id`. Existing question UUIDs must be preserved on re-seed via normalized-text `UPDATE` matching rather than delete-and-reinsert.
- **Persistent Per-Learner Sample Lock (`user_theory_sample_questions`)**: When a Learner first opens a Sub-topic, deterministically draw and immediately persist **2** unused active questions inside a concurrency-safe transaction (`PRIMARY KEY (user_id, sub_topic_id, question_id)`, plus `UNIQUE (user_id, question_id)` so no question is ever assigned twice to the same learner across Sub-topics):
  - **Primary Draw**: unused active questions matching both `domain_id` and `sub_topic_id` (`NOT IN` Pre-Test questions and previously assigned Learn sample questions), flagged `is_fallback = false`.
  - **Domain Fallback Draw**: if fewer than 2 unused questions remain in that `sub_topic_id`, fill the remaining slot(s) from unused active questions in the parent `domain_id`, flagged `is_fallback = true`.
- **Zero-Overlap Post-Test Exclusion**: `StartPostTest` excludes the union of all `question_id`s from the user's Pre-Test sessions and all `question_id`s persisted in `user_theory_sample_questions` for that user. Across domains (`110/110/90/150/80`), consuming Pre-Test (`4/7/5/8/6`) plus up to `28 × 2 = 56` Learn samples (`8/10/8/18/12`) always leaves between 62 and 124 unused questions per domain for Post-Test selection.

### Schema Shape

```sql
ALTER TABLE questions ADD COLUMN IF NOT EXISTS sub_topic_id VARCHAR(16);

CREATE TABLE IF NOT EXISTS theory_subtopics (
  sub_topic_id VARCHAR(16) PRIMARY KEY,
  domain_id VARCHAR REFERENCES security_domains(domain_id) ON DELETE CASCADE NOT NULL,
  order_index INT NOT NULL,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  source_url TEXT,
  updated_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_theory_subtopic_progress (
  user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
  sub_topic_id VARCHAR(16) REFERENCES theory_subtopics(sub_topic_id) ON DELETE CASCADE,
  completed_at TIMESTAMP DEFAULT now(),
  PRIMARY KEY (user_id, sub_topic_id)
);

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
```

### API Contracts, Anti-Bypass Gate & Role Behavior

All new and updated handlers use dedicated request/response DTOs and OpenAPI annotations per ADR-0005, and preserve ADR-0004 (`user_skill_profiles.proficiency_score` invariance):

- **`GET /api/learn/domains`**: Extends each domain DTO with its nested `subtopics` metadata array (`sub_topic_id`, `order_index`, `title`, `is_completed`, `completed_at`, `sample_attempted_count`, `sample_total_count`) and top-level `completed_subtopic_count` / `total_subtopics` (`28`) / `completed_subtopics`. Full Sub-topic Markdown bodies are omitted here for payload efficiency.
- **`GET /api/learn/subtopics/:subTopicId`**: Returns Sub-topic Markdown `content`, clickable `source_url`, completion status, `previous_sub_topic_id` / `next_sub_topic_id` navigation pointers, and the 2 assigned sample questions. For Learners, `correct_answer` and `explanation` are omitted until `submitted_answer` is non-null. For Specialists and Admins, returns deterministic read-only preview questions (`preview_mode: true`) without inserting rows into Learner tracking tables.
- **`POST /api/learn/subtopics/:subTopicId/sample-questions/:questionId/attempt`**: Accepts `{ "answer": "A"|"B"|"C"|"D" }`. For Learners, verifies question assignment, evaluates answer once on first submit (`submitted_answer`, `is_correct`, `answered_at`), locks subsequent attempts cleanly (`already_submitted: true` without mutating first-attempt correctness), and returns `explanation`, `correct_answer`, and `ready_to_complete`. For Specialists/Admins, grades statelessly (`preview_mode: true`) with zero database writes.
- **`POST /api/learn/subtopics/:subTopicId/complete`**: Verifies that all assigned sample questions for `(user_id, sub_topic_id)` have been attempted (`400 sample_questions_incomplete` otherwise), idempotently inserts into `user_theory_subtopic_progress`, and if all Sub-topics in the parent domain are completed, automatically records `(user_id, domain_id)` in `user_theory_progress`.
- **Anti-Bypass Guard on `POST /api/learn/domains/:domainId/complete`**: Rejects Learner requests with `403 subtopics_incomplete` (listing missing `sub_topic_id`s) unless every child Sub-topic in that domain is already completed.
- **Stage 2 Simulation Gate (`GET /api/simulation/status` & `POST /api/simulation/start`)**: Upgrades `theory_completed` evaluation for Learners from `5/5` domains to **`28/28` Sub-topics completed** (retaining `BYPASS_THEORY_GATE=true` and `specialist`/`admin` exemptions). Pre-existing Learner accounts without 28 completed Sub-topics are gated cleanly.
- **Research Telemetry (`GET /api/analytics/research-summary`)**: Retains existing fields and adds `learn_subtopic_telemetry` (`completed_subtopics_count`, `total_subtopics`, `completed_subtopics`, `sample_questions_assigned`, `sample_questions_attempted`, `sample_questions_correct`, `sample_accuracy_percent`, and per-domain `domain_breakdown`).

### Frontend Learning Hub Architecture

- **Nested Sidebar Tree & Deep-Linking**: Replace the flat 5-domain tab list with an expandable Domain → Overview + nested Sub-topics tree displaying proficiency badges, Recommended Focus highlights, and Sub-topic completion counters (`X/Y`). Sync selection cleanly with `?domain=...&topic=...` URL query parameters (defaulting to the primary Recommended Focus domain's Overview).
- **Overview vs. Sub-Topic Workspace**: Domain Overviews render high-level theory plus a Sub-topic checklist grid without a manual domain completion button. Sub-topic views render Markdown theory, clickable web links, two independent sample question cards (highlighting `Domain Related Fallback` badges when `is_fallback` is true), immediate locked answer feedback + explanations, an attempt-gated **Mark Sub-topic as Completed** button, domain/gate unlock notifications, and cross-domain **Previous Sub-topic / Next Sub-topic** buttons.

## Testing Decisions

- **What makes a good test**: Assert observable public behavior — HTTP statuses, sanitized pre-submission vs. post-submission JSON payloads, question set disjointness, idempotent lock semantics, gate state transitions, and proficiency score invariance — rather than internal SQL structures or helper functions.
- **Seam 1 (Primary End-to-End Integration Seam — existing HTTP verification harness)**: Extend the existing Python API integration suites to exercise the full user journey against a live seeded PostgreSQL database: Pre-Test unlocks `/learn`; `GET /api/learn/domains` returns 5 Overviews and 28 Sub-topics; opening a Sub-topic locks 2 non-Pre-Test questions and omits answers before submission; submitting both attempts unlocks Sub-topic completion; completing all Sub-topics in a domain auto-completes the parent domain (and direct domain completion fails beforehand with `403`); `POST /api/simulation/start` rejects at `27/28` and succeeds at `28/28`; `StartPostTest` yields 30 questions with **zero overlap** against `Pre-Test ∪ Learn Sample Questions`; and `/api/analytics/research-summary` reports exact formative telemetry while `proficiency_score` remains unaffected by Learn attempts.
- **Seam 2 (Controller & Store Boundary — existing `LearnStore` / controller Go test suite)**: Test primary draw vs. domain fallback (`is_fallback = true`), cross-subtopic uniqueness, immediate first-attempt lock idempotency, anti-bypass `403 subtopics_incomplete`, Staff `preview_mode` zero-persistence behavior, `StartPostTest` SQL exclusion, and Stage 2 `27/28` vs. `28/28` gate checks cleanly without requiring live HTTP servers for every permutation.
- **Seam 3 (Curriculum & Question Bank Validation — existing Python validation suite)**: Assert that all 5 Domain Overviews and 28 Sub-topic Markdown files pass required section checks and contain clickable `https://` web URLs (no broken or missing files), and verify 100% of the 540 questions carry a valid `sub_topic_id` matching their parent `domain_id`.
- **Seam 4 (Frontend Build Seam)**: Verify zero TypeScript or bundler errors via `tsc` and production Vite build after UI integration.

## Out of Scope

- Grading formative Learn sample questions into `user_skill_profiles.proficiency_score` or modifying the Scenario evaluation rubric (ADR-0004).
- Dynamic/on-the-fly LLM generation of Sub-topic lessons or exam questions during Learner page loads.
- Support for non-MCQ interactive question types (`drag_drop`, `scenario_based` bank items — all 540 existing bank items are `multiple_choice` `A–D`).
- Unlimited answer-reset retries on Learner sample questions after first submission (`submitted_answer` is permanently locked on first attempt).
- Minimum reading timers or scroll-tracking locks.

## Further Notes

- **Originating Workflow Spec**: Settled decisions, full 28-item SY0-701 mapping, and canonical glossary terms are documented in `workflows/learning_subtopics_sample_questions.md` and `NOTES.md`.
- **Execution Push-Right Checkpoint**: Authoring the 28 web-grounded Sub-topic Markdown files, updating `docs/REFERENCES.md` + `docs/references.bib`, and auditing the 540-question `sub_topic_id` classification should be completed and reviewed as the first milestone before wiring database migrations, API handlers, and frontend components.
- **Originating Issue**: #30, "feat(learn): Domain Overviews, Sub-Topics, and Formative Sample Exam Questions".
