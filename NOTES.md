# Domain Notes & Terminology (NOTES.md)

## Core Roles & Terminology

- **Learner (คนเรียนรู้)**: The primary training user undergoing pre-test, scenario simulation loops, and post-test.
- **Specialist (ผู้เชี่ยวชาญ)**: An approved domain expert who reviews learner scenario answers, provides deep technical analysis, and publishes analysis threads (กระทู้).
- **Admin**: Platform administrator who verifies and approves Specialist candidate applications (evaluating resumes, certificates, and credentials) and oversees platform operations.
- **Nickname (ชื่อเล่น)**: Public display name for users across discussions, leaderboards, and reviews, preserving privacy while allowing friendly identification.
- **Specialist Application (การสมัครเป็น Specialist)**: Onboarding process requiring extra qualification artifacts (resume file/URL, professional certificates, security experience summary).
- **Analysis Thread (กระทู้การวิเคราะห์)**: A post authored by a Specialist breaking down scenario solutions, common learner mistakes, triage techniques, and key finding analysis.
- **Discussion / Comment (คอมเมนต์ / ตอบกระทู้)**: Collaborative replies and questions on an analysis thread between Learners, Specialists, and Admins.

## Settled Decisions (Round 1)

1. **Role Architecture**:
   - Roles: `learner`, `specialist`, `admin` (migrating `expert` to `specialist`).
   - Registration flow: Candidates register as `learner` by default. If applying as Specialist, an application record is created in `specialist_applications` with status `pending`.
   - Admin approves the application to promote the user role to `specialist`.
   - Existing Learners can also submit an application to upgrade to `specialist`.

2. **Credentials & Artifacts**:
   - Hybrid mode: Supports uploading resume and certificate files (PDF/Image) stored on server + optional URL links (LinkedIn, portfolio, Credly).

3. **Learner Answers Visibility**:
   - Specialists have a review dashboard to inspect completed simulation sessions filtered by Scenario, Domain, or Score.
   - Learner identity is displayed via **Nickname** to protect email privacy.
   - Specialist views: TP/FP triage decision, discovered key findings, response actions executed, log query attempts, and final score.

4. **Analysis Threads (กระทู้) & Discussion**:
   - Authoring: Restricted to `specialist` and `admin` roles.
   - Scope: Can be linked directly to a `scenario_id` or written as general blue-team analysis.
   - Interaction: 2-level comment hierarchy (top-level comment + direct reply). All roles (`learner`, `specialist`, `admin`) can participate in discussions.

5. **Nickname**:
   - Mandatory for all users at registration.
   - Must be unique across the platform.
   - Serves as the public display name in community discussions and learner review records.

## Settled Decisions (Round 2)

6. **Admin Checkpoint & Rejection Handling**:
   - Rejection requires an Admin to enter an explanation note (`rejection_reason`).
   - The applicant sees the `rejected` status with the reason on their Profile.
   - The applicant can re-apply by updating their profile details and re-uploading documents without creating a new account.

7. **Artifact Storage & File Security**:
   - Resumes and certificates stored under protected storage `/uploads/applications/{application_id}/`.
   - Accessible only to Admins and the application owner via JWT-authenticated endpoint `GET /api/admin/applications/:id/files/:file_type`.
   - Prevent public static directory exposure and enforce UUID naming.

8. **Linking Scenario Results to Analysis Threads**:
   - `ScenarioResultPage` embeds a "Specialist Insights & Discussions" section below evaluation metrics.
   - Directly links to threads tagged with that `scenario_id` so learners can immediately review expert analysis and common mistakes.

9. **Role Badges & Thread Pinning**:
   - Visual badges next to Nickname: `[Specialist]`, `[Admin]`, `[Learner]`.
   - Thread authors (`specialist`) and `admin` can pin key takeaway comments to the top of the thread.

10. **In-App Notification Loop**:
    - Navbar includes an in-app notification indicator (bell) showing:
      - Specialist application approval / rejection status updates.
      - Replies to user comments in analysis threads.

## Settled Decisions (Round 3)

11. **Rich Content & Scenario Log Embedding**:
    - Markdown editor with syntax highlighting for KQL queries, syslog/JSON dumps.
    - Quick-embed widget allowing Specialists to inject scenario summaries (alert details, key findings checklist, expected actions) into their thread.

12. **Moderation & Permissions**:
    - Authors can edit and delete their own threads and comments.
    - Thread authors (Specialists) can hide or delete inappropriate comments on their threads.
    - Admins have master moderation privileges (edit, delete, pin, lock) across all community content.

13. **Existing Account Migration**:
    - Migration script backfills `nickname` using email prefix (with random suffix if duplicate).
    - Users can update their nickname anytime in their Profile.

14. **Helpful / Upvoting System**:
    - 1-upvote per user on threads and comments to highlight quality analyses.

## Settled Decisions (CI/CD & DevSecOps Pipeline)

15. **Zero-Cost Hosting Architecture**:
    - Frontend SPA: Cloudflare Pages (Free tier, fast edge CDN, unlimited bandwidth).
    - Backend Go API: Render.com (Free Web Service tier, Docker runtime).
    - Database: Neon Serverless PostgreSQL (Free tier with pgvector extension support).

16. **Pipeline Stages & DevSecOps Gates**:
    - **Trigger Gate**: PRs to `main` trigger CI only (Linting, DevSecOps scanning, Test Harness, Build). Direct push/merge to `main` triggers CD.
    - **DevSecOps Scans**: Gitleaks (secret sniffing) + `govulncheck` (Go CVEs) + `npm audit` (Frontend dependencies).
    - **Integration Verification**: GitHub Actions spins up `pgvector/pgvector:pg16` service container, initializes `init.sql`, runs `load_exams_to_db.py`, launches Go backend, and runs `test_integration.py` to assert exam 30-question distribution & zero-overlap integrity.

17. **Continuous Delivery & Cold-Start Optimization**:
    - **Backend Deploy**: Triggered via Render Deploy Hook Webhook upon green merge to `main`.
    - **Frontend Deploy**: Automated deploy to Cloudflare Pages.
    - **Keep-Alive Worker**: Scheduled cron every 14 minutes pings the health endpoint to prevent Render Free Tier container hibernation.

## Settled Decisions (Issue #1: Adaptive Scenario Playbook & Anti-Predictability Loop)

18. **Playbook Option Density & Balance**:
    - Every incident response phase (`containment`, `eradication`, `recovery`) MUST contain **3 to 4 options**:
      - **1–2 Correct Actions** (`is_correct: true`, `+5` points each).
      - **2 Contextual Plausible Distractors** (`is_correct: false`, `-2` to `-3` points each).
    - **Length & Tone Parity**: Distractor labels must match the technical specificity, sentence structure, and length of correct actions (referencing the scenario's `target_host`, `attacker_ip`, affected accounts, or services) rather than short, obviously absurd actions.
    - **Forbidden Legacy Distractors**: Hardcoded boilerplate distractors (`"Power off all domain controllers immediately"`, `"Reformat all network switches in the subnet"`, `"Disable MFA enforcement for affected users"`) are banned across all scenarios.

19. **Contextual Distractor Synthesis & Fallback Pool**:
    - **LLM Generation**: `aiMicroScenario` schema and prompt are expanded so the LLM generates contextual distractors and explanations per phase alongside correct actions.
    - **Domain-Aware Contextual Backfill**: If the LLM omits distractors or when `GenerateFallbackScenario` is invoked, a domain-and-context-interpolated distractor generator synthesizes realistic distractors dynamically parameterized with the scenario's `target_host` and `attacker_ip`.
    - **False Positive Alignment**: For False Positive (`is_true_positive: false`) scenarios, correct playbook steps represent non-destructive verification, alert tuning, and baseline documentation, while distractors represent premature host isolation, account lockouts, or destructive remediation.

20. **Deterministic Session-Scoped Option Shuffling**:
    - `sanitizePlaybookSteps(steps, seed)` deterministically shuffles the option slice within each phase (`containment`, `eradication`, `recovery`) using a hash seeded by `session_id + ":" + phase`.
    - Ensures correct answers never predictably appear at index 0, while guaranteeing order stability across browser refreshes (`GET /api/simulation/session` and `POST /api/simulation/start`).

21. **Validation Gate & Legacy Scenario Auto-Remediation**:
    - **Validation Gate (`ValidateScenarioCandidate`)**: Enforces that every phase has `>= 3` options, at least `1` correct option, at least `2` incorrect options, unique option labels, and zero banned boilerplate strings.
    - **Static Seed & DB Auto-Upgrade**: `scripts/generate_scenarios.py` and `scenarios_seed.json` are updated so all 15 static scenarios have 3–4 balanced options per phase; on backend startup or scenario load, any legacy DB scenarios containing < 3 options per phase or banned boilerplate labels are automatically upgraded with contextual distractors.

## Settled Decisions (Deployment Hardening & Container Security)

22. **Fail-Closed Secrets**:
    - `docker-compose.yml` no longer ships default credentials. `POSTGRES_PASSWORD`, `JWT_SECRET`, and `PGADMIN_DEFAULT_PASSWORD` use the `${VAR:?}` form, so Compose aborts instead of silently starting with `password`/`secret_jwt_key_default`.
    - `.env` is the single source of local secrets (git-ignored); `.env.example` documents every required key. `ENABLE_SWAGGER` defaults to `false`.

23. **Least-Privilege Containers**:
    - **Backend**: runs as `nobody` (`65534:65534`), `read_only: true` root FS with a `tmpfs /tmp`, `cap_drop: [ALL]`, and `no-new-privileges`. It binds `8080` (>1024) so no capability is required.
    - **All services**: `security_opt: no-new-privileges:true`, bounded `json-file` logs (10MB x 3), and `deploy.resources.limits` (CPU/memory) to contain runaway processes.

24. **Network Segmentation & Minimal Host Exposure**:
    - Two bridges: `frontend_net` (frontend + backend) and `backend_net` (backend + db + seed + pgadmin). The database is never on `frontend_net`.
    - The Postgres port is **not** published to the host; the backend and the one-shot `seed` container reach it over `backend_net`. The API (`8080`) and pgAdmin (`5050`) bind to `127.0.0.1` only; only the frontend (`80`) is exposed on all interfaces.

25. **Automated One-Shot Seeding & Health Gates**:
   - A `seed` service (Python 3.11-slim) runs `parse_exams.py` -> `load_exams_to_db.py` -> `generate_scenarios.py` -> `load_scenarios_to_db.py`, mounting the repo read-only (`./:/app:ro`) and writing ephemeral artifacts to `/work`.
   - Start order is gated by health/completion: `db (healthy)` -> `seed (completed successfully)` -> `backend (healthy)` -> `frontend`. pgAdmin is opt-in via `--profile tools`.

## Canonical Terms (Issue #11)

- **Evidence log (log หลักฐาน)**: a row in a scenario's `initial_logs` that proves one key finding; identified by its 0-based index in the original array.
- **Selected log (log ที่ถูกเลือก)**: a log the learner explicitly toggled on. It is a claim of relevance, not a TP/FP verdict.
- **Noise log (log ขยะ/ไม่เกี่ยว)**: a log not present in any key finding's evidence set. Selecting one is a false lead.
- **Finding credit (คะแนน finding)**: full finding points awarded when the selected set intersects that finding's authored evidence set. No partial credit inside a finding.
- **Evidence penalty (โทษเลือกผิด)**: `−2` per unique selected noise log.
- **Benign context log (log ยืนยันว่า benign)**: a change ticket, HR travel confirmation, scheduled maintenance, or authorized scanner range that proves a False Positive verdict.
- **Evidence mapping report (รายงาน mapping)**: the compact human-review brief produced during scenario evidence authoring.
- **Evidence selection (การเลือกหลักฐาน)**: the learner's click-to-select action in Detect & Analyze, persisted as `select_evidence` session actions.

## Settled Decisions (Issue #11: Detect & Analyze Findings Cannot Get Score)

26. **Root Cause**:
   - Findings were credited only when the learner's transient KQL-filtered view contained **all** of a finding's evidence indices. Seeing the right log but not querying every index produced 0.
   - `assign_evidence_indices()` distributed indices by positional chunking (`15..N-16`), not by what each log proves.

27. **New Scoring Unit**:
   - A finding is credited in full when the learner selects **any one** of its authored evidence logs.
   - Selecting multiple logs from the same finding does not double-credit.
   - No partial credit within a finding.

28. **Noise Penalty**:
   - `−2` per unique selected log that belongs to no finding's evidence set.
   - Deducted from the findings component only, floored at 0.
   - Attributed to `scenario.domain_id`; finding credit remains attributed to each `finding.domain_id`.

29. **Evidence Authoring**:
   - `evidence_log_indices` is authored, never positionally chunked.
   - No per-finding count cap; the count depends on the scenario. False Positive findings may need several logs to establish the benign explanation.
   - The union of all evidence sets must be ≤ 50% of `initial_logs`.
   - Overlap between findings is allowed freely; a shared log credits every finding it supports.
   - True Positive scenarios require at least one `malicious` verdict-supporting finding; False Positive scenarios require at least one `benign` verdict-supporting finding with a benign context log.

30. **Authoring Workflow**:
   - Static seeds: LLM drafts mappings → deterministic validation → one blocking user approval of a compact report → apply and seed.
   - Runtime AI scenarios: 1–4 findings; exact `evidence_log_messages`; backend maps messages to indices; False Positive cases get a synthesized benign context log if the LLM omits one.
   - Mapping JSON, report, and seed are committed; CI validates both mapping and seed.
   - See `workflows/scenario_evidence_authoring.md`.

31. **Selection Persistence & Audit**:
   - New `select_evidence` action type with `toggle` and `clear` payloads.
   - `POST /api/simulation/select-log` and `POST /api/simulation/select-log/clear`.
   - `GET /api/simulation/session` returns `selected_log_indices`; selections survive refresh and filter changes.
   - Submit replays server-recorded actions as the source of truth; client/server mismatch is logged, and the server wins.
   - Step-order allocation is serialized per session to avoid `(session_id, step_order)` collisions.

32. **Feedback & Payload Hygiene**:
   - No correctness feedback during selection; only selected state and count.
   - Learner payload exposes only `total_findings`; all evidence indices, finding IDs, and descriptions are removed.
   - KQL query history and query count are audit-only and never scored.

33. **TP/FP Independence**:
   - TP/FP classification remains a separate 20-point component.
   - Finding credit does not require a correct verdict; evidence identification and classification are measured separately.

34. **Result & Specialist Review**:
   - Result page shows found/missed findings, matched evidence rows, missed acceptable evidence, noise picks, and the `−2` penalty.
   - Specialist review shows selected logs, matched/missed findings, noise picks, and net findings points.
   - Completed sessions are immutable and retain their original rubric version (`rubric-v1`); new sessions use `rubric-v2-evidence-click`.

## Proposed / Deferred Loops

- **Missed-Finding Remediation Loop**: after a scenario, use the missed finding (not only the weakest domain) to select or generate the next scenario targeting that specific technique. Deferred; revisit after Issue #11 ships.




## Canonical Terms (Issue #26)

- **Theory Module (บทเรียนทฤษฎีรายโดเมน)**: the structured English curriculum document for one of the 5 CompTIA Security+ SY0-701 domains (`domain1`–`domain5`), stored under `theory/domain{1..5}_*.md` and seeded into PostgreSQL.
- **Theory Gate (เกตปลดล็อกการเรียนทฤษฎี)**: the two-stage progression gate where (1) completing the 30-question Pre-Test unlocks `/learn`, and (2) completing all 5 Theory Modules (`5/5`) unlocks starting Simulation Scenarios (`POST /api/simulation/start`).
- **Domain Theory Completion (การเรียนจบรายโดเมน)**: a persisted record in `user_theory_progress (user_id, domain_id, completed_at)` indicating a learner has completed a domain's Theory Module.
- **Recommended Focus Domain (โดเมนที่แนะนำให้เน้น)**: any domain where the learner's `proficiency_score < 70.0%` (or the lowest-scoring domain), highlighted on `/learn` and `/evaluation`.

## Settled Decisions (Issue #26: Learning Theory Hub)

35. **Role in the Learner Loop & Two-Stage Gate**:
   - Progression order: `Pre-Test (30 questions)` → unlocks `Learn (/learn)` → completing all `5/5` domains in `Learn` unlocks `Adaptive Simulation (/simulation)` → `Post-Test`.
   - Once unlocked after the Pre-Test, `/learn` remains permanently accessible at any time (via Navbar, `/evaluation`, and `/simulation/result/:id`).
   - `specialist` and `admin` roles can view `/learn` at any time without taking the Pre-Test.

36. **Content Structure, Language & Provenance**:
   - Written in **English** across 5 Markdown files in `theory/`:
     - `theory/domain1_security_concepts.md`
     - `theory/domain2_threats_vulnerabilities.md`
     - `theory/domain3_security_architecture.md`
     - `theory/domain4_security_operations.md`
     - `theory/domain5_security_management.md`
   - Combines **CompTIA Security+ (SY0-701) Core Theory** (for Pre/Post-test alignment) and **SOC Blue-Team Operational Theory** (SIEM/EDR telemetry & KQL indicators, MITRE ATT&CK mapping, TP vs. FP triage with benign context logs, and 3-phase Containment/Eradication/Recovery playbook selection).
   - Compiled from authoritative public references (CompTIA SY0-701 Objectives, NIST SP 800-53 / 800-61r2, MITRE ATT&CK, OWASP, Wikipedia) and registered in `docs/REFERENCES.md` (`R2`–`R5`) and `docs/references.bib` before seeding.

37. **Storage, Progress Tracking & Scoring Invariance**:
   - Seeded into PostgreSQL (`rag_documents`) via `scripts/load_theory_to_db.py` (wired into `docker-compose.yml` `seed` service) with embedded Go fallback.
   - Domain-level completion (`5/5` checkpoints via **"Mark Domain as Completed"**) persisted in `user_theory_progress`.
   - Displays learner's current `proficiency_score` per domain and flags weak domains (`< 70%`) as **Recommended Focus**, without directly altering `proficiency_score`.
   - Exposed in `GET /api/simulation/status` and `GET /api/analytics/research-summary` (`theory_completed`, `completed_theory_domains`, `theory_completed_at`).
   - See `workflows/learning_theory_hub.md`.

## Canonical Terms (Issue #30)

- **Domain Overview (ภาพรวมโดเมน)**: the 5 high-level curriculum summaries retained in `theory/domain{1..5}_*.md` (`rag_documents`), acting as informational landing pages and sub-topic roadmaps for each Security+ domain.
- **Theory Sub-Topic (หัวข้อย่อยตามแนว SY0-701)**: one of the **28 official CompTIA Security+ (SY0-701) objective modules** (`1.1`–`5.6`), stored under `theory/domain{1..5}/*.md`, embedded in Go, and seeded into `theory_subtopics`.
- **Formative Sample Questions (ข้อสอบตัวอย่างระหว่างเรียน)**: the **2 unused `multiple_choice` questions** locked per `(user_id, sub_topic_id)` in `user_theory_sample_questions` upon first viewing a Sub-topic, with immediate graded feedback, locked first attempt, and domain-level fallback (`is_fallback`) when a Sub-topic pool is exhausted.
- **Learn Seen Pool (กลุ่มข้อสอบที่เคยเห็นในบทเรียน)**: all `question_id`s assigned to a learner in `user_theory_sample_questions`, which are permanently excluded alongside `Pre-Test` questions when generating `StartPostTest` (`Zero-Overlap: Pre-Test ∪ Learn ∪ Post-Test = ∅`).

## Settled Decisions (Issue #30: Domain Overviews, Sub-Topics & Sample Questions)

38. **Exact 28-Sub-Topic Granularity & Stage 2 Gate**:
    - 1:1 alignment with CompTIA Security+ (SY0-701): Domain 1 (`1.1`–`1.4`), Domain 2 (`2.1`–`2.5`), Domain 3 (`3.1`–`3.4`), Domain 4 (`4.1`–`4.9`), Domain 5 (`5.1`–`5.6`) = **28 Sub-topics**.
    - Domain Overviews are informational only (no manual `"Mark Domain Completed"` button); completing all Sub-topics in a domain automatically marks the parent domain completed in `user_theory_progress`, and legacy `POST /api/learn/domains/:id/complete` rejects (`403 subtopics_incomplete`) until all child Sub-topics are done.
    - Stage 2 Gate (`POST /api/simulation/start`) requires **28/28 Sub-topics completed** in `user_theory_subtopic_progress` (pre-existing learner accounts must complete the 28 Sub-topics; `specialist`/`admin` keep preview access without writing Learner completion rows).

39. **Formative Sample Question Allocation, Fallback & Zero Overlap**:
    - All 540 `examination/` questions (`parsed_questions.json`) are classified into `questions.sub_topic_id` (`1.1`–`5.6`) and safely upserted via `scripts/load_exams_to_db.py` preserving existing `question_id` UUIDs.
    - Opening a Sub-topic locks 2 unused questions for `(user_id, sub_topic_id)` excluding `Pre-Test ∪ Learn Seen`; if fewer than 2 remain in `sub_topic_id`, remaining slots draw from the parent `domain_id` (`is_fallback = true`). `UNIQUE(user_id, question_id)` prevents assigning the same question twice.
    - Submitting an answer immediately grades, reveals `explanation` + `correct_answer`, and locks the first submitted attempt (`submitted_answer`, `is_correct`) without altering `user_skill_profiles.proficiency_score` (`ADR-0004`).
    - Attempting both (`2/2`) sample questions unlocks `"Mark Sub-topic as Completed"` regardless of whether the answers were right or wrong.
    - `StartPostTest` excludes **both** `Pre-Test` questions and all rows in `user_theory_sample_questions` for that user.

40. **Web-Only Trustworthy Sourcing, UI Tree & Research Analytics**:
    - Sub-topic prose draws strictly from trustworthy public websites (Wikipedia, NIST, CISA, MITRE ATT&CK, OWASP, CMU ISO, Stanford UIT, Microsoft Learn — no textbooks) with clickable `https://` links and registration in `docs/REFERENCES.md` + `docs/references.bib`.
    - `/learn` uses an expandable **Nested Sidebar Tree** with deep-linking (`/learn?domain=domain1&topic=1.2`) and `Previous / Next Sub-topic` navigation.
    - `GET /api/analytics/research-summary` exposes `learn_subtopic_telemetry` (`completed_subtopics_count`, `sample_questions_assigned`, `sample_questions_attempted`, `sample_questions_correct`, `sample_accuracy_percent`, and `domain_breakdown`).
    - See `workflows/learning_subtopics_sample_questions.md`.


## Canonical Terms (Issue #25)

- **Incident Event Chain (ชุดล็อกเหตุการณ์หลัก)**: the **4 to 8 contextual telemetry log entries** (`incident_logs[]`) generated for a specific scenario's narrative, replacing the legacy hardcoded 5-row block (`firewall` → `edr` → `ids` → `web_app` → `siem_correlator` `[CORR-9901]`).
- **Environment Noise Pack (ชุดล็อกพื้นหลังตามสภาพแวดล้อม)**: a domain/archetype-matched generator producing diverse operational background telemetry (`identity_directory`, `endpoint_workstation`, `cloud_infrastructure`, `network_perimeter`, `application_data`, `governance_itsm`) without repetitive identical `for-loop` messages.
- **Decoy Warning Log (ล็อกแจ้งเตือนตัวหลอกที่ไม่มีพิษภัย)**: a non-evidence background noise log carrying `warning` or `medium` severity (e.g., transient auth retry, health-check timeout, routine scanner block, certificate expiry notice) included (`2–4` per scenario) so learners cannot solve Detect & Analyze merely by filtering `High/Critical` or `Warning`.
- **Chronological Interleaving (การคละลำดับเวลาตามจริง)**: dispersing the scenario's Incident Event Chain across a variable-length (`28–46` rows) timeline with background noise before, between, and after key events, eliminating the fixed `15..19` middle-block sandwich.
- **Scenario Archetype & Briefing Style (ต้นแบบสถานการณ์และสไตล์การบรีฟเคส)**: the combination of a CompTIA SY0-701 Sub-topic (`1.1`–`5.6`), an operational incident archetype, and one of 5 SOC briefing formats (`siem_escalation_ticket`, `shift_handover_brief`, `threat_hunting_lead`, `helpdesk_user_report`, `cloud_guardrail_alert`) rotated across a learner's sessions.
- **Context-Aware SIEM Shortcuts (ปุ่มลัด SIEM ตามบริบทโจทย์)**: dynamic investigation filter chips on `SimulationPage.tsx` derived from the scenario's actual target host, actor IP/user, and primary telemetry sources rather than `initial_logs[0].host` or hardcoded `Firewall Logs`.

## Settled Decisions (Issue #25: Scenario Telemetry, Archetype & Briefing Style Diversity)

41. **Multi-Event AI Incident Chain & Contextual Noise Synthesis (Hybrid)**:
    - `aiMicroScenario` (`services/llm_service.go`) is expanded with `incident_logs` (`4–8` structured log events `{source, severity, host, src_ip, dst_ip, user, message}`), `sub_topic_id`, `archetype`, and `briefing_style`, while retaining `attack_log_message` fallback compatibility for legacy unit tests.
    - The hardcoded 5-log block (`Perimeter Firewall: Connection request initiated...`, `IDS Alert: Malicious attack payload signature matched...`, `HTTP 500 Internal Server Error...`, `SIEM Alert Rule [CORR-9901]...`) and the fixed 45-log Kerberos/Routine sandwich in `buildScenarioLogs` (`services/scenario_evidence.go`) are retired.
    - `GenerateFallbackScenario` uses multi-event, domain-distinct incident chains rather than the legacy 5-row wrapper.

42. **Variable Log Count, Multi-Severity Decoy Noise & Chronological Interleaving**:
    - Total log count per scenario varies deterministically in the range **28–46 logs** (never fixed at 45).
    - Background noise is synthesized from 6 environment-matched Noise Packs with unique messages and realistic host/IP/user variety (no `for-loop` repeating the exact same message 15–25 times).
    - Every scenario includes **2–4 Decoy Warning/Medium noise logs** that are not part of any finding's evidence set, alongside informational/warning/high/critical incident logs, preventing severity-only clicking shortcuts.
    - Incident events are interleaved chronologically across the log array (with noise separating multi-stage events) rather than clustered contiguously at indices `15..19`, and `evidence_log_indices` are resolved after final timeline ordering.

43. **Sub-Topic / Archetype Rotation, History Exclusion & Controlled TP/FP Cadence**:
    - Before calling `GenerateLLMScenario`, the backend inspects the learner's recent `simulation_sessions` (titles, hosts, IPs, sub-topics, and verdicts) to select the next unvisited (or least-visited) SY0-701 Sub-topic (`1.1`–`5.6`) within the weakest domain and a matching Scenario Archetype + Briefing Style.
    - Recent scenario titles, hosts, and IPs are passed into the LLM prompt's exclusion list to prevent consecutive plot/entity repetition.
    - The backend deterministically targets a **~70% True Positive / ~30% False Positive cadence** (injecting a False Positive target roughly every 3rd–4th scenario and preventing >2 consecutive False Positives) and passes `target_is_true_positive` to the LLM prompt.

44. **Varied Narrative Briefing Styles & Dynamic SIEM Shortcuts**:
    - Replaces the uniform `INVESTIGATION CLUES:\n- Target Host: ...\n- Suspected Attacker IP: ...\n- Alert Vector: ...` footer with 5 realistic SOC briefing styles (`siem_escalation_ticket`, `shift_handover_brief`, `threat_hunting_lead`, `helpdesk_user_report`, `cloud_guardrail_alert`).
    - Fixes the `SimulationPage.tsx` `Host:` shortcut bug (which grabbed `initial_logs[0].host` = `WS-WORKSTATION1`) and replaces static `Firewall Logs` / `High Severity Alerts` buttons with 4 dynamic, context-aware shortcuts derived from the scenario's primary entities (`host`, `ip` or `user`, primary `source`, and relevant `severity`).

45. **Static Seed Re-Authoring, Anti-Repetition Gate & Legacy DB Archival**:
    - All 15 static scenarios in `scripts/generate_scenarios.py` are upgraded with interleaved timelines, diverse multi-source/multi-severity background noise (`28–46` rows), and zero repetitive `for-loop` message blocks; `scripts/evidence_mapping.json`, `scripts/evidence_mapping_report.md`, and `scenarios_seed.json` are updated in lockstep.
    - Both `scripts/validate_evidence_mapping.py` and `services.ValidateScenarioCandidate` enforce anti-repetition rules: banning legacy boilerplate log strings (`SIEM Alert Rule [CORR-9901]`, `Routine network telemetry event `, `Kerberos ticket TGT requested for user account user`) and capping identical `message` repetition (`<= 3` occurrences per scenario).
    - On startup (`EnsureScenarioEvidenceQualityFromSeed`), active non-seed scenarios in PostgreSQL that still contain the legacy `CORR-9901` / 45-log sandwich boilerplate are automatically marked `status = 'archived'` without altering completed `simulation_sessions` or `scenario_snapshots`.
    - See `workflows/scenario_telemetry_and_style_diversity.md`.

## Canonical Terms (Issue #39)

- **UI Locale (`en` / `th`) (ภาษาแสดงผลของระบบ)**: The user's active interface language (`en` default, `th` supported, extensible via locale codes) synced across `localStorage` (`cybersim_lang`) and `users.preferred_language`, without changing URL paths.
- **English-Only Realism Invariant (ข้อกำหนดคงภาษาอังกฤษเพื่อความสมจริง)**: The hard rule that CompTIA Security+ assessment items (Pre-Test, Post-Test, and `/learn` Formative Sample Questions), Simulation `initial_logs`, KQL syntax, IOCs, and MITRE ATT&CK IDs MUST remain 100% English (zero Thai Unicode `\u0E00–\u0E7F`) regardless of active UI Locale.
- **Hybrid Bilingual Terminology (คำศัพท์เทคนิคไทยควบคู่อังกฤษ)**: Thai prose localization style that retains standard CompTIA Security+ and SOC technical terms/acronyms (`SIEM`, `KQL`, `True Positive`, `False Positive`, `Containment`, `Eradication`, `Recovery`, `MITRE ATT&CK`) alongside natural Thai explanations.
- **Scenario Translation Payload (`translations.th`) (ข้อมูลคำแปลสถานการณ์สองภาษา)**: The structured JSONB translation map stored alongside canonical English fields in `scenarios` and immutable `scenario_snapshots`, covering `title`, `description`, `tp_fp_explanation`, `playbook_steps[action_id]`, and `findings[finding_id]` (sanitized pre-submission to prevent answer leaks).
- **Theory Translation Mirror (`theory/th/` & `theory_translations`) (คลังบทเรียนทฤษฎีภาษาไทยคู่ขนาน)**: The 33 Thai Markdown files mirroring `theory/` (5 Domain Overviews + 28 Sub-topics), embedded in Go and seeded into the dedicated `theory_translations` table so `rag_documents` vector embeddings remain untouched.
- **Additive Localization Contract (สัญญา API แบบเพิ่มฟิลด์ไม่กระทบของเดิม)**: Adding `code` + `params` to `dto.ErrorResponse` and `template_key` + `template_params` to `user_notifications` while keeping English `error`/`title`/`message` intact as backward-compatible fallbacks.

## Settled Decisions (Issue #39: Bilingual Language Switching & Localization)

46. **Locales, Strict Content Boundaries & Terminology**:
    - Supports `en` (default) and `th` with zero URL prefix changes (`/learn`, `/simulation` stay unchanged).
    - **Translated (`en`/`th`)**: Frontend UI chrome (all 16 pages & shared components), 33 Learning Theory Markdown files (5 Domain Overviews + 28 Sub-topics), Simulation Scenario non-log fields (Title, Narrative/Briefing, Playbook Options, Findings, Explanations across static & AI scenarios), API Errors, and In-App Notifications.
    - **Strictly English-Only**: Pre-Test, Post-Test, `/learn` Formative Sample Questions, Simulation `initial_logs`, KQL syntax/queries, IOCs, and MITRE IDs — accompanied by a subtle explanatory info notice when `locale === "th"`.
    - **Author-Language Preserved**: Analysis Threads, Comments, and Admin free-text rejection reasons are never auto-translated.
    - Uses Hybrid Bilingual Terminology (Thai prose + standard English CompTIA/SOC terms) and Gregorian (ค.ศ.) Thai dates (`th-TH-u-ca-gregory`) for UI timestamps while keeping SIEM timestamps RFC3339/ISO.

47. **Preference Persistence, Switcher UX & Zero-Dependency Frontend i18n**:
    - Persisted in `localStorage` (`cybersim_lang`) + `users.preferred_language VARCHAR(10) NOT NULL DEFAULT 'en'`.
    - Register inherits guest `localStorage` locale; Login adopts DB `users.preferred_language`; explicit toggle while logged in updates `localStorage` optimistically and persists via `PUT /api/user/language`.
    - Compact `EN | TH` toggle in Top Navbar on every route + synced selector on ProfilePage; switches instantly without reload and preserves all active Exam, Simulation, Learn, and Discussion state.
    - Zero-dependency typed TS dictionaries (`frontend/src/i18n/en.ts`, `th.ts`) enforced at compile time via `tsc -b`, paired with `IBM Plex Sans Thai` + `Geist` (`font-sans`) while keeping `Geist Mono` (`font-mono`) for SIEM/KQL.

48. **Multilingual Theory, Bilingual Single-Call AI Scenarios & Additive Backend Contracts**:
    - 33 Thai theory files live in `theory/th/`, embedded via `//go:embed`, validated & seeded into `theory_translations` (`entity_type`, `entity_id`, `locale`, `title`, `content`), and delivered via `GET /api/learn/domains?lang=` and `GET /api/learn/subtopics/:id?lang=`.
    - 15 static scenario translations live in `scripts/scenario_translations_th.json` merged into `scenarios_seed.json` without touching `initial_logs` or `evidence_mapping.json`.
    - Runtime AI scenarios use **Bilingual Single-Call** generation (`en` + `th` non-log fields, EN-only `incident_logs`) stored in `scenarios.translations` and immutable `scenario_snapshots`, delivered inline on Simulation/Result/Specialist Review endpoints for 0 ms toggling (with strict pre-submission sanitization omitting `tp_fp_explanation`, `findings`, and playbook `explanation`); `buildInvestigationShortcuts` always consumes canonical EN `title`/`description`.
    - `ValidateScenarioCandidate` enforces **Strict Reject & Fallback** if `initial_logs` contain Thai Unicode or `translations.th` is incomplete; `EnsureScenarioEvidenceQualityFromSeed` archives legacy active non-seed AI scenarios missing valid `translations.th` on startup.
    - Backend errors add `code` + `params` (preserving EN `error`); `user_notifications` adds `template_key` + `template_params` (preserving EN `title`/`message`).

49. **Push-Right Translation Checkpoint & 3-Gate CI Verification**:
    - Initial static translation uses **1 push-right Human Review Checkpoint**: author Pilot Batch (Domain 1 Theory [5 files] + 3 sample Static Scenarios) and present a compact Brief for tone/glossary approval before scaling to Domains 2–5 (28 files) and remaining 12 Static Scenarios.
    - Full 3-gate CI verification: (1) Frontend `tsc -b` + dictionary key/`{{param}}` parity check, (2) 33 TH theory files + 15 static seeds `translations.th` completeness check, (3) Go & Python strict English-only invariant tests (zero `\u0E00–\u0E7F` in exams, sample questions, and `initial_logs`).
    - See `workflows/language_switching_and_localization.md`.

## Canonical Terms (Scenario Storytelling Narrative)

- **4-Beat Incident Narrative (`story`)**: Structured narrative metadata (`role_context`, `current_activity`, `what_happened`, `potential_impact`) stored under `translations.{en,th}.story` inside existing JSONB without altering PostgreSQL DDL.
- **Friendly Guided Prose (ภาษาเล่าเรื่องที่เข้าใจง่าย)**: Explaining specialized components inline in plain terms while retaining 2–3 canonical English entities (`Host`, `User`, `IP`) so `siemShortcuts.ts` extracts SIEM shortcuts identically and removing abrupt analyst tool commands from the story.
- **Balanced Dual-Stakes Impact (ผลกระทบสมดุลสองด้าน)**: Conditional Beat 4 framing describing both security/operational harm if malicious *and* business disruption if legitimate work is mistakenly blocked, preventing True Positive / False Positive spoilers.

## Settled Decisions (Scenario Storytelling Narrative — `/loop-me` Q1–Q13)

50. **SOC Analyst Persona, Second-Person Voice & Smooth 3-Paragraph Flow**:
    - Every scenario frames the learner as an organizational **SOC Analyst (`นักวิเคราะห์ SOC`)** in second-person voice (`"You..."` / `"คุณ..."`), rotating workplace duty and trigger scene across all 5 `BriefingStyles`.
    - Structured `story` (4 beats) composes cleanly into **3 short badge-free paragraphs** (`[role_context + current_activity]\n\n[what_happened]\n\n[potential_impact]`) in both canonical `description` and `translations.th.description`.
    - UI displays spaced narrative paragraphs cleanly with zero step tags, numbered beat headers, or spoiler hints.

51. **Zero-DDL Coverage Across AI, Fallback, 15 Seeds & Lazy Legacy Local Upgrade**:
    - Stored inside existing `scenarios.translations.{en,th}.story` JSONB (`init.sql` unchanged).
    - Upgraded uniformly across runtime AI prompts, `GenerateFallbackScenario`, all 15 static seeds (`scripts/generate_scenarios.py`, `scripts/scenario_translations_th.json`, keeping `evidence_mapping.json` SHA-256 hashes valid), and lazy on-demand `UpgradeScenarioStoryBilingual` when existing local sessions/snapshots lack `translations.{en,th}.story` (overlaying narrative only onto legacy snapshots without altering `initial_logs`, `playbook_steps`, or scoring, failing open if `LLM_API_KEY` is unset).
    - See `workflows/scenario_storytelling_narrative.md`.

## Canonical Terms (Issue #38: UI Typography & Readability)

- **Semantic Typography Scale (สเกลตัวอักษรตามบทบาท)**: Role-based `rem` font-size hierarchy across the frontend (`text-base` 16px primary reading prose, `text-sm` 14px standard UI/controls, `text-log` 13px / `0.8125rem` dense SIEM logs & code, `text-xs` 12px absolute hard floor).
- **Hard Floor (`text-xs` / 12px / `0.75rem`)**: Minimum permissible font size anywhere in `frontend/src/`; all `text-[8px]`–`text-[11px]` and arbitrary `px` font-size classes are banned.
- **SIEM Log Token (`text-log`)**: Named Tailwind `rem` fontSize token (`0.8125rem`, line-height `1.5`) giving `LogTable`, code blocks, and evidence snippets readable compact density while scaling cleanly with browser zoom.
- **Automated Typography Guard (`frontend/tests/typography.test.ts`)**: Static unit test run in `npm test` and GitHub Actions CI that fails builds introducing arbitrary `text-[...px]`, sub-`0.75rem` font sizes, invalid `text-graphite-550`, or missing Thai monospace fallback / `text-log` token.

## Settled Decisions (Issue #38: App-Wide UI Typography & Readability — `/loop-me` Q1–Q12)

52. **App-Wide Semantic Typography Scale & `rem`-Only Architecture**:
    - Refactors all 16 pages and 10 shared components to the Balanced Modern Web Scale (`16px` reading prose, `14px` UI/controls, `13px` `text-log` SIEM telemetry/code, `12px` `text-xs` hard floor) with zero user font-size switcher in Navbar (native browser zoom preserved via 100% `rem` units).
    - Adds `fontSize.log: ['0.8125rem', { lineHeight: '1.5' }]` and appends `"IBM Plex Sans Thai"` fallback to `fontFamily.mono` in `frontend/tailwind.config.js`; sets `body` default to `text-base leading-relaxed` in `frontend/src/index.css`.
    - Keeps unified scale across `EN` and `TH` (no layout jump on toggle) with Thai-friendly `leading-relaxed` / `leading-snug`.

53. **Targeted Anti-Truncation, Contrast Hygiene & CI Typography Guard**:
    - Applies targeted layout polish where larger fonts could clip or crowd: un-truncates long domain names in `DomainBreakdown` (`leading-snug` wrap), preserves single-row Navbar at `>= 1280px` across all roles (`EN`/`TH`), sizes checkboxes/icons proportionally (`h-4 w-4`), and increases `LogTable` scroll height (`max-h-[560px]`).
    - Fixes invalid `text-graphite-550` (`LandingPage.tsx`) and lifts readable muted/hint copy from `text-graphite-500` → `text-graphite-400`.
    - Enforces regressions via `frontend/tests/typography.test.ts` wired into `npm test` & `.github/workflows/ci.yml`, with 1 Push-Right Visual Review Checkpoint (Before/After screenshots in `EN` & `TH` on 5 flagship screens) before closing Issue #38.
    - See `workflows/ui_typography_readability.md`.

## Canonical Terms (Issue #14: Post-Exam Q&A Review & Learning Alignment)

- **Exam Q&A Review (`/exams/review/:sessionId`)**: Dedicated, bookmarkable, refresh-safe post-completion review page for both Pre-Test and Post-Test showing overall score, domain scores, structured explanations, and direct Learning Hub links.
- **Smart Expand (กางเฉพาะข้อผิด/เว้นว่างอัตโนมัติ)**: Default card state on `ExamReviewPage` where `Incorrect` and `Unanswered` cards open automatically for immediate study while `Correct` cards start collapsed to 1-line summaries (toggleable individually or via `Expand All / Collapse All`).
- **Distractor Breakdown (การแยกอธิบายตัวเลือกที่ผิด)**: Deterministic client-side parsing of the 540/540 uniform `explanation` format into `Why <Correct> is Correct` + per-option incorrect rationales, pinning and highlighting the learner's mistaken pick (`Your pick`) first (with graceful raw fallback).
- **Pre→Post Learning Gain Card (การ์ดพัฒนาการ Pre→Post)**: Comparison banner shown on completed Post-Test reviews comparing overall % and all 5 Security Domains against the learner's latest completed Pre-Test (`ORDER BY completed_at DESC LIMIT 1`, aligned with `GetResearchSummary`) while strictly preserving `user_skill_profiles` Proficiency Invariance.
- **Explicit Exam Finalization (`POST /api/exams/sessions/:sessionId/finalize`)**: Idempotent completion endpoint that marks blank answers as `unanswered` (`user_answer = ""`, `is_correct = false`), records real `time_spent_seconds`, transitions the session to `completed`, and supports both confirmed partial submission and single-shot `00:00:00` auto-finalize.

## Settled Decisions (Issue #14: Post-Exam Q&A Review — `/loop-me` Q1–Q10)

54. **Dedicated Review Route, Revisit Entry Points & Strict Completion Guard (Zero-DDL)**:
    - Completing either Pre-Test or Post-Test redirects immediately to `/exams/review/:sessionId`; persistent revisit links live on `/evaluation` ("Exam Answer Reviews" card) and `/profile` ("Exam History & Reviews" section).
    - Backend adds `POST /api/exams/sessions/:sessionId/finalize`, `GET /api/exams/history`, and `GET /api/exams/sessions/:sessionId/review` with zero DB schema changes (`init.sql` untouched).
    - Strict access guard: learners can view only their own `completed` sessions (`403 EXAM_NOT_COMPLETED` for `in_progress`, `404` for foreign/missing), ensuring zero answer leakage.

55. **Learning-Aligned Review UX, Structured Explanations & Post-Test Gain**:
    - Default view: `All` filter + `Grouped by Security Domain` (`domain1`–`domain5`) + **Smart Expand** (`Incorrect`/`Unanswered` expanded, `Correct` collapsed), with toggle to `Exam Order (1–30)`, status & domain filters, and a color-coded `01–30` Jump Grid.
    - Deterministic `parseExamExplanation` utility splits explanations into main rationale + per-option distractor notes (pinning the learner's wrong pick with a `Your pick` badge) and every card deep-links directly to `/learn?domain=<id>&topic=<sub_topic_id>`.
    - Post-Test review renders an Overall + 5-Domain Pre→Post Learning Gain card against the latest completed Pre-Test; `FinalizeExamSession` computes `domain_scores` for both `pre` and `post` while keeping `user_skill_profiles` untouched on `post` (Proficiency Invariance).

56. **Real Time Tracking, Timeout Auto-Finalize & Visual Checkpoint**:
    - `ExamPage` records real `time_spent_seconds` and performs a single-shot auto-submit + finalize when the 30-minute timer reaches `00:00:00`, routing cleanly to `/exams/review/:sessionId`.
    - Verified via Go unit tests, Vitest parser/i18n/typography suites, `scripts/test_integration.py`, and 1 Push-Right Visual Review Checkpoint (EN/TH screenshots of Pre-Test Review, Post-Test Review + Gain Card, `/evaluation`, and `/profile`) before closing Issue #14.
    - See `workflows/exam_qa_review.md`.

## Canonical Terms (Issue #13: Post-Test Evaluation Questionnaire & Paper-Ready Statistics)

- **Post-Test Evaluation Questionnaire (`/questionnaire`)**: The bilingual (`EN`/`TH`) 3-step subjective evaluation completed once per `learner` after finishing the 30-question Post-Test, combining **CSUQ v3 (16 items)**, **Raw NASA-TLX (6 dimensions)**, and **3 optional qualitative impression prompts**, persisted into the existing `evaluation_metrics` table (Zero-DDL).
- **CSUQ v3 Ascending Scale (`1–7`, Higher = Better + `N/A`)**: Standardized 16-item usability instrument (`[R8]` Lewis, 1995; 2018) scored `1 = Strongly Disagree` → `7 = Strongly Agree` across `SYSUSE` (`Q1–Q6`), `INFOQUAL` (`Q7–Q12`), `INTERQUAL` (`Q13–Q15`), and `OVERALL` (`Q1–Q16`), excluding `N/A` items from subscale means while forbidding all-16 `N/A`.
- **Raw NASA-TLX with Polarity Normalization (RTLX `0–100`, step `5`)**: Unweighted 6-dimension cognitive workload instrument (`[R9]` Hart & Staveland, 1988; Hart, 2006); UI renders `Performance` intuitively (`0 = Failure` → `100 = Perfect`) while the backend stores both `performance_rating` and inverted `performance_workload = 100 - performance_rating` to compute `overall_workload = mean(6 workloads)`.
- **Server Auto-Snapshot (`evaluation_metrics`)**: Authoritative server-side capture of latest completed `pre_score`, `post_score`, `improvement_pct = post - pre`, and `user_skill_profiles` (`skill_gap_summary`) executed inside a `pg_advisory_xact_lock` transaction on submission so client scores are never trusted and duplicate rows per user are prevented without schema changes.
- **Paper-Ready Statistical Report (`/admin/evaluations/statistics`)**: Admin-only publication workbench powered by a pure-Go statistical engine computing descriptive stats (`n`, `M`, `SD`, Student-$t$ `95% CI`, `Median`, `IQR`, `Min–Max`), **Cronbach's $\alpha$**, **Paired $t$-test + Cohen's $d_z$**, **Wilcoxon signed-rank + $r_{rb}$**, and **Spearman $\rho$**, paired with `Copy LaTeX (IEEEtran)`, `Copy Markdown`, an Auto-Generated Academic Results Paragraph (`EN` default / `TH` thesis option), a bilingual Methodology & Formula Guide, Print/PDF mode, and Small-$n$ Guardrails (`n < 2/3/5/30`).

## Settled Decisions (Issue #13: Post-Test Evaluation Questionnaire & Paper Statistics — `/loop-me` Q1–Q16)

57. **Standard Research Bundle, Non-Blocking Flow & Zero-DDL Server Snapshot**:
    - Uses **CSUQ v3 (16 items, ascending `1–7` + `N/A`) + Raw NASA-TLX (6 dims, `0–100` step `5`, UI `Performance` inverted server-side) + 3 Optional Qualitative Impression Prompts** (stored as structured JSONB and readable sectioned `notes TEXT`), bilingual (`EN`/`TH`) with `localStorage` auto-draft persistence.
    - Finishing Post-Test still routes to `/exams/review/:sessionId` with a prominent non-blocking CTA banner to `/questionnaire` (plus persistent reminders on `/evaluation` and `/profile` until submitted).
    - Strictly 1 submission per `learner` gated on having a completed Post-Test (`403` before Post-Test, `409` on duplicate via `pg_advisory_xact_lock`); `specialist`/`admin` see interactive Preview Mode only; Zero-DDL using existing `evaluation_metrics` with server auto-snapshot + additive fields on `GET /api/analytics/research-summary`.
    - Registers `R8` (CSUQ v3 — Lewis, 1995; 2018) and `R9` (NASA-TLX / RTLX — Hart & Staveland, 1988; Hart, 2006) in `docs/REFERENCES.md` and `docs/references.bib`.

58. **Admin Evaluations Dashboard, Paper-Ready Statistics & Verification**:
    - Adds shared `AdminPortalTabs` (`Specialist Applications` | `Post-Test Evaluations` | `Paper Statistics`) across `/admin/specialists`, `/admin/evaluations`, and `/admin/evaluations/statistics` with zero Top Navbar bloat (keeping Issue #15 scoped strictly to technical NFR benchmarks).
    - `/admin/evaluations` provides KPI cards, Recharts CSUQ & NASA-TLX bar charts, a pseudonymized response table (`P-XXXXXXXX` + `nickname`, expandable qualitative notes), and UTF-8 BOM CSV (with formula-injection neutralization) + JSON export.
    - `/admin/evaluations/statistics` provides the full pure-Go IEEE statistical suite (`Descriptive + 95% CI`, `Cronbach's α`, `Paired t-test + Cohen's d_z`, `Wilcoxon + r_rb`, `Spearman ρ`), `Copy LaTeX (IEEEtran)` & `Copy Markdown`, Auto-Generated Academic Results Paragraph (`EN` default + `TH` thesis toggle), bilingual Statistical Methodology & Formula Guide, `@media print` PDF layout, and strict Small-$n$ Guardrails (`n < 2/3/5`, Pilot Caveat when `n < 30`).
    - Verified via Go unit tests (including SciPy/textbook-checked math), Node frontend tests (questionnaire, paper generators, i18n parity, typography guard), OpenAPI Swagger sync, `scripts/test_integration.py`, and 1 Push-Right Visual Review Checkpoint (EN & TH screenshots) before closing Issue #13.
    - See `workflows/post_test_questionnaire.md`.

## Canonical Terms (Issue #15: System-Level NFR Evaluation Portal)

- **NFR Evaluation Workbench (`/admin/nfr`)**: The Admin-only 4th tab inside `AdminPortalTabs` evaluating the platform across all **10 ISO/IEC 25010 (`[R10]`) dimensions** (`Performance`, `Availability`, `Scalability`, `Reliability`, `Security`, `Data Integrity`, `Reproducibility`, `Usability`, `Maintainability`, `Safety`), strictly separated from Learner subjective surveys (Issue #13) while bridging live CSUQ/NASA-TLX scores into `Usability`.
- **3-Layer Hybrid NFR Measurement (การวัดผล NFR จริง 3 ชั้น)**: Combines (1) Passive Runtime Telemetry (bounded in-memory ring buffer + LLM attempt/latency recorder + DB-backed `/api/health`), (2) Live DB/Security/Safety/Usability Audit Probes, and (3) Reproducible Concurrency & CI Benchmarks (Safe In-App Probe + `scripts/run_nfr_benchmark.py` / UI JSON Import).
- **NFR Rubric (`nfr-rubric-v1`)**: The canonical numeric `Pass` / `Warn` / `Fail` cutoff ruleset with explicit `not_measured` and `insufficient_data` states (never masquerading as fake pass or fail) and Transparent Separate Rollup (`overall_status`: `fail` → `partial_pending` → `pass_with_warnings` → `all_passed`, `measured_pass_rate = passed / measured`).
- **NFR Evaluation Run (`nfr_evaluation_runs`)**: An immutable PostgreSQL snapshot storing a complete `NFRReportResponse` JSONB payload (`live_snapshot`, `in_app_probe`, `cli_benchmark`, or `imported_json`) for Run History selection, side-by-side Baseline Delta Comparison, and reproducible Paper/Thesis exports across server restarts.
- **Internal Probe Telemetry Isolation (การแยกทราฟฟิกทดสอบออกจาก Telemetry จริง)**: Per-process secret header (`X-SOC-Internal-NFR-Probe`) injected exclusively by in-process `httptest` Security and Safe Concurrency Probes so synthetic admin audit traffic never pollutes real user HTTP latency/availability metrics.

## Settled Decisions (Issue #15: System-Level NFR Evaluation — `/loop-me` Q1–Q16)

59. **10-Dimension ISO/IEC 25010 Matrix, 3-Layer Measurement & Single-Table Persistence**:
    - Evaluates all 10 advisor/ISO/IEC 25010 dimensions (`Performance`, `Availability`, `Scalability`, `Reliability`, `Security`, `Data Integrity`, `Reproducibility`, `Usability`, `Maintainability`, `Safety`) and registers ISO/IEC 25010:2023/2011 as permanent reference **`[R10]`** in `docs/REFERENCES.md` & `docs/references.bib`.
    - Passive HTTP telemetry (`10,000`-sample bounded ring buffer, `c.FullPath()` route normalization, separating `non_llm_api` vs `POST /api/simulation/start`, zero PII) + LLM call recorder run in memory with zero per-request DB write overhead (Render/Neon free-tier friendly); `GET /api/health` upgraded to run a real 2s PostgreSQL `SELECT 1` ping (`200 healthy` / `503 degraded`).
    - Adds 1 PostgreSQL table `nfr_evaluation_runs` (`init.sql` + idempotent `models/migrate.go`) to store immutable evaluation snapshots (`live_snapshot`, `in_app_probe`, `cli_benchmark`, `imported_json`), enabling Run History, Baseline Delta Comparison, per-run export, and test-run deletion (`DELETE /api/admin/nfr/runs/:runId`).

60. **Dual-Mode Benchmarks, Probe Safety, Telemetry Isolation & Strict Admin RBAC**:
    - Mounted exclusively as the 4th tab in `AdminPortalTabs` (`/admin/nfr`, zero Top Navbar changes) and under `/api/admin/nfr/*` protected by `AuthMiddleware() + RequireAdmin()` (`403` for `learner`/`specialist`), enforcing a strict **Zero-Secret Exposure Rule** (boolean posture checks only, never leaking secrets/DSNs/keys).
    - **Scalability** supports both a Safe In-App Concurrency Probe (`POST /api/admin/nfr/probe/scalability`, presets `5/50`, `10/100`, `20/200`, hard-capped at `20` workers / `200` read-only requests / `15s` + mutex `409`) and external CLI `scripts/run_nfr_benchmark.py` (`10/25/50` concurrent users + optional CI/test/coverage & security scan collector supporting `--push` and `--out` + UI `"Import Benchmark / CI JSON"` upload `<= 1 MB`).
    - **Security** runs a read-only in-process `httptest` Auth/RBAC/Input probe (6 checks) + 6 runtime posture checks; **Safety** runs live DB checks (English-only invariant, zero secret leaks in snapshots, zero banned boilerplate, evidence `<= 50%`); **Usability** bridges live `#13` CSUQ/NASA-TLX scores + objective completion metrics; all in-process probes are excluded from live passive telemetry via an internal per-process secret header, and Admins can cleanly reset the in-memory telemetry window (`POST /api/admin/nfr/telemetry/reset`).

61. **Bilingual Single-Page Report, Recharts Visualizations, Paper Export Suite & Checkpoint**:
    - `/admin/nfr` uses a 6-section print-friendly single-page layout (`rem`-only typography, `EN`/`TH` chrome with English technical identifiers): Control & Run Bar → Executive KPI Strip (+ Baseline Delta) → 10-Dimension ISO/IEC 25010 Scorecard Grid → 3 Recharts Charts (Endpoint Latency `p50/p95/p99`, Scalability Curve, AI vs Static Fallback Reliability) → Run History Table → Paper-Ready Export & `[R10]` Methodology Panel (`Copy LaTeX IEEEtran`, `Copy Markdown`, UTF-8 BOM CSV / JSON download, EN/TH Auto-Generated NFR Paragraph, `@media print` PDF).
    - Verified via Go unit tests, OpenAPI Swagger sync (`dto/nfr.go`), Node frontend tests (`nfrReport.test.ts`, `i18n.test.ts`, `typography.test.ts`), `scripts/test_integration.py`, `scripts/run_nfr_benchmark.py`, and 1 Push-Right Visual Review Checkpoint (EN & TH screenshots) before closing Issue #15.
    - See `workflows/nfr_evaluation.md`.

## Canonical Terms (Cross-Platform Setup & Cloud Bootstrap)

- **Cross-Platform Installer Harness (`install.sh` & `install.ps1`)**: Symmetric Linux Bash and Windows PowerShell CLIs (paired with `install.bat` ExecutionPolicy bypass wrapper) orchestrating preflight checks, `.env` provisioning, containerized `db → seed → backend → frontend` startup, smoke verification, and `backup` / `restore` / `update` / `uninstall` without requiring host Python, Go, or Node.js.
- **Clean-VM Bootstrapper (`scripts/bootstrap.sh` & `scripts/bootstrap.ps1`)**: Standalone entrypoints for bare Linux/Windows hosts that install `git` if missing, clone or fast-forward this repository, and hand off directly to `install.sh` / `install.ps1`.
- **Cloud TLS Compose Overlay (`docker-compose.cloud.yml` + `Caddyfile`)**: Optional Compose override persisted via `COMPOSE_FILE` and `COMPOSE_PATH_SEPARATOR=:` in `.env` whenever a Cloud `DOMAIN` is configured, resetting `frontend` host ports (`!reset []`) so `caddy:2-alpine` terminates `:80/:443` with automatic Let's Encrypt TLS cleanly.
- **Secret Resolution Precedence (`ENV → <KEY>_FILE → .env → Prompt/CSPRNG`)**: Fail-closed credential lookup preserving existing `.env` secrets on re-runs, generating CSPRNG `POSTGRES_PASSWORD`/`JWT_SECRET` (no weak `"password"` fallback), locking `.env` file permissions (`chmod 600` / `icacls`), and suppressing secrets in stdout/logs.
- **Non-Polluting Smoke Verification**: Post-install readiness checks asserting `5` domains, `540` questions, `28` subtopics, `33` Thai theory docs, `15` scenarios, `/api/health` (`db_status=up`), frontend HTTP `200`, and Admin login verification without inserting synthetic test/learner records into research tables.

## Settled Decisions (Cross-Platform Setup & Cloud Bootstrap — `/loop-me` Q1–Q17)

62. **Unified Self-Hosted Docker Stack, Clean-Machine Bootstrapping & Full OS Parity**:
    - Standardizes Local and Cloud deployments on the Spec 6 Docker Compose stack across **Linux (`install.sh`: Ubuntu/Debian `apt`, RHEL/Alma/Rocky/Fedora `dnf`/`yum`, Amazon Linux 2023 + macOS Docker Desktop check)** and **Windows (`install.ps1` + `install.bat` wrapper: Windows 10/11 Docker Desktop WSL2 & Windows Server)**, plus bare-metal `scripts/bootstrap.{sh,ps1}` repo cloners.
    - Offers confirmed (`--auto-deps`) host dependency installation (`git`, Docker Engine/Compose on Linux; WSL2/`winget` Git + Docker Desktop with clear reboot resume guidance on Windows) and adds `.gitattributes` enforcing `eol=lf` on shell/Docker/Caddy scripts and `eol=crlf` on `.ps1`/`.bat`.
    - Supports full lifecycle parity (`install`, `check`, `status`, `seed`, `check-seed`, `update`, `backup`, `restore`, `logs`, `uninstall`, `reset-dev-data`), interactive menu when invoked bare, and `--non-interactive` flags.

63. **Cloud Auto-HTTPS (`Caddy`), Low-RAM Protection & Safe Firewall Posture**:
    - `--mode cloud` sets `APP_ENVIRONMENT=cloud`; providing a Domain runs optional DNS A-record check, generates `Caddyfile` + activates `docker-compose.cloud.yml` (`ports: !reset []` on `frontend`, Caddy `:80/:443`), and persists `FRONTEND_URL=https://<domain>` + `COMPOSE_FILE` + `COMPOSE_PATH_SEPARATOR=:` inside `.env`. IP-only Cloud mode stays on HTTP `:80` with an explicit TLS warning.
    - Source build (`--build`) is default, supplemented by a 2 GB Linux Cloud swap helper when RAM < 2 GB (no swap) and optional `--image-tag`; host firewall rules (`ufw` / Windows Firewall) are audited and listed cleanly by default and altered only on explicit `--configure-firewall` opt-in.

64. **Fail-Closed Secrets, 100% Containerized Seeding, Guarded Ops & Doc Consolidation**:
    - Removes host Python requirement by running seeding inside Docker `seed` cleanly; post-install smoke checks verify DB counts, health, HTTP, and Admin login without generating synthetic learner rows (`--verify-e2e` optional in `--mode local` only).
    - `backup` writes timestamped `.sql.gz` + `uploads` volume `.tar.gz` (retaining last 10 sets) using actual `.env` credentials; `update` auto-backs up before rebuild & safe re-seed; `restore` enforces typed `RESTORE` + pre-restore safety backup; `reset-dev-data` is strictly forbidden in Cloud mode; `uninstall --volumes` requires typed `DELETE DATA`.
    - Consolidates setup documentation into `workflows/cross_platform_setup.md`, updates `.env.example` & `README.md`, and marks legacy setup drafts superseded.
    - See `workflows/cross_platform_setup.md`.



