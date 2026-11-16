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
