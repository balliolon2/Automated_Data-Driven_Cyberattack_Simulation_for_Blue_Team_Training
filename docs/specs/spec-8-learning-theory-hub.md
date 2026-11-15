---
title: "Spec 8: Learning Theory Hub and Pre-Simulation Knowledge Gate (Issue #26)"
labels: ["ready-for-agent"]
status: "ready-for-agent"
---

# Spec 8: Learning Theory Hub and Pre-Simulation Knowledge Gate

## Problem Statement

A Learner completes the 30-question Diagnostic Pre-Test, sees their initial Domain Breakdown with one or more domains below the 70% Proficiency Threshold, and is immediately sent into the Scenario Loop with no structured way to study the underlying cybersecurity theory first. When the Learner lacks foundational knowledge in a CompTIA Security+ domain — or does not yet know how theoretical security concepts translate into SIEM telemetry logs, True Positive versus False Positive triage criteria, and three-phase Incident Response actions — they are forced to learn purely by trial and error inside scored Simulation Sessions.

Furthermore, when a Learner misses a Key Finding or selects an incorrect Response Action during a Simulation Session, there is no reference curriculum on the platform where they can review that domain's principles before attempting the next Scenario. From a research perspective, the one-group pretest–posttest behavioral design also benefits from an explicit, verifiable instructional stage between baseline assessment and practical simulation so that every participant receives standardized, authoritative domain theory across all five CompTIA Security+ (SY0-701) domains before practical evaluation begins.

## Solution

Introduce a dedicated **Learning Theory Hub** covering all five CompTIA Security+ (SY0-701) domains, governed by a two-stage **Theory Gate** in the Learner journey:

1. **Stage 1 Gate (Pre-Test Unlocks Theory):** A Learner must first complete the 30-question Diagnostic Pre-Test to establish their baseline skill profile. Until the Pre-Test is completed, the Learning Theory Hub is locked for Learners. Once the Pre-Test is completed, the Learning Theory Hub unlocks permanently and can be revisited at any time. Specialists and Admins can access the Learning Theory Hub at any time without taking the Pre-Test.
2. **Five-Domain Actionable Curriculum:** Each of the five domains provides a comprehensive English **Theory Module** synthesized from authoritative, registered public sources (CompTIA Security+ SY0-701 objectives, NIST incident handling and security control guidelines, MITRE ATT&CK, OWASP, and computer security reference articles). Every module combines:
   - Core CompTIA Security+ theoretical concepts and exam-aligned definitions,
   - SOC telemetry, log indicators, KQL search patterns, and MITRE ATT&CK mappings for Detect & Analyze,
   - True Positive versus False Positive triage guidance (including Benign Context Logs) and proportional Containment, Eradication, and Recovery Response Action principles, and
   - Key takeaways and source citations.
3. **Skill-Profile Awareness Without Scoring Drift:** The Learning Theory Hub displays the Learner's current Proficiency Score alongside each domain and highlights any domain below the 70% Proficiency Threshold (or the lowest-scoring domain) as a **Recommended Focus Domain**. Studying or completing Theory Modules never directly alters a Learner's Proficiency Score, preserving the integrity of the evaluation formula.
4. **Stage 2 Gate (Completing All Five Domains Unlocks Simulation):** Each domain module includes an explicit **Mark Domain as Completed** action. A Learner must complete all five domains (`5/5`) before starting their first Simulation Session. Once all five domains are marked completed, starting adaptive Scenarios is unlocked.
5. **Contextual Remediation Links & Research Telemetry:** Both the Skill Evaluation page and the Scenario Result page provide direct deep-links into the relevant domain's Theory Module so Learners can remediate weak domains between Scenarios. Completion status and timestamps are exposed in the simulation status and research analytics summary endpoints.

## User Stories

1. As a Learner who has not yet taken the Pre-Test, I want the Learning Theory Hub to tell me that the Diagnostic Pre-Test is required first, so that my baseline proficiency is measured before I study the curriculum.
2. As a Learner on the locked Learning Theory Hub, I want a direct action to start the Pre-Test, so that I can immediately unlock the learning materials.
3. As a Learner who has just finished the Pre-Test, I want to view my initial Domain Breakdown on the Skill Evaluation page first, so that I know which domains are below the 70% Proficiency Threshold before I begin reading.
4. As a Learner on the Skill Evaluation page who has not yet finished reading all five domains, I want the primary call-to-action to guide me to the Learning Theory Hub with my current completion count (`X/5`), so that I know what is required to unlock Simulation Sessions.
5. As a Learner, I want a permanent navigation link to the Learning Theory Hub in the top navigation bar, so that I can return to the theory materials at any point after completing the Pre-Test.
6. As a Learner, I want to see all five CompTIA Security+ domains listed clearly in the Learning Theory Hub (`General Security Concepts`, `Threats, Vulnerabilities, and Mitigations`, `Security Architecture`, `Security Operations`, and `Security Management`), so that I can navigate between domains effortlessly.
7. As a Learner, I want to see my current Proficiency Score percentage next to each domain in the Learning Theory Hub, so that I can connect the reading material to my skill profile.
8. As a Learner, I want domains where my Proficiency Score is below the 70% Proficiency Threshold to be badged as Recommended Focus Domains, so that I know where to pay the closest attention.
9. As a Learner whose domains are all at or above 70%, I want my lowest-scoring domain highlighted as the Recommended Focus Domain, so that I still know which domain the adaptive Scenario Loop will target first.
10. As a Learner, I want an overall progress indicator showing how many of the five domains I have completed (`0/5` through `5/5`), so that I can track my progress toward unlocking the Scenario Loop.
11. As a Learner studying Domain 1 (`General Security Concepts`), I want clear English coverage of the CIA Triad, AAA, non-repudiation, Zero Trust, cryptographic principles, and security control categories, so that I understand foundational security theory.
12. As a Learner studying Domain 2 (`Threats, Vulnerabilities, and Mitigations`), I want clear English coverage of threat actors, attack vectors, phishing/social engineering, application and network vulnerabilities, indicators of compromise, and mitigation techniques, so that I can recognize real-world attack patterns.
13. As a Learner studying Domain 3 (`Security Architecture`), I want clear English coverage of network segmentation, DMZs, cloud security models, defense-in-depth, secure protocols, host hardening, and resilience concepts, so that I understand how secure infrastructure is designed and defended.
14. As a Learner studying Domain 4 (`Security Operations`), I want clear English coverage of SIEM telemetry fields, KQL filtering strategies, endpoint and network monitoring, vulnerability management, and the incident response lifecycle, so that I can investigate operational alerts effectively.
15. As a Learner studying Domain 5 (`Security Program Management and Oversight`), I want clear English coverage of security governance, policies and standards, risk management calculations, third-party risk, compliance frameworks (GDPR/PDPA, PCI-DSS, ISO 27001), and incident reporting/chain of custody, so that I understand organizational oversight.
16. As a Learner, I want every domain's Theory Module to include concrete examples of SIEM logs, event fields, and KQL search patterns, so that I know what Evidence Logs to look for during Detect & Analyze.
17. As a Learner, I want every domain's Theory Module to explain how to distinguish a True Positive from a False Positive — including how Benign Context Logs (such as change tickets, HR travel approvals, and scheduled scans) prove an authorized event — so that I can make accurate TP/FP Decisions.
18. As a Learner, I want every domain's Theory Module to explain the difference between proportional Containment, Eradication, and Recovery Response Actions versus destructive or premature distractors, so that I avoid penalty traps in the playbook phases.
19. As a Learner, I want every domain's Theory Module to map concepts to relevant MITRE ATT&CK tactics and techniques, so that I can connect theoretical knowledge to attacker behavior.
20. As a Learner, I want telemetry examples, KQL queries, and MITRE technique identifiers rendered in a high-contrast monospace font on a dark graphite surface, so that the reading experience matches the operational aesthetic of the platform.
21. As a Learner, I want a Mark Domain as Completed button at the end of each domain's Theory Module, so that I can record that I have finished studying that domain.
22. As a Learner, I want my domain completion status to persist across page refreshes, logouts, and devices, so that I never have to re-mark domains I already completed.
23. As a Learner, I want marking a domain as completed to be idempotent, so that clicking the button multiple times or re-reading a completed domain never causes an error or resets my progress.
24. As a Learner, I want to step directly to the next or previous domain from the bottom of the reader panel, so that I can study all five domains in sequence without scrolling back to the top.
25. As a Learner who has completed fewer than five domains, I want any attempt to start a Simulation Session to redirect or instruct me to finish the remaining Theory Modules first, so that I do not enter the Scenario Loop unprepared.
26. As a Learner who has just completed the fifth and final Theory Module (`5/5`), I want an immediate Start Adaptive Simulation call-to-action on the Learning Theory Hub, so that I can enter the Scenario Loop right away.
27. As a Learner on the Skill Evaluation page, I want a Study Theory link on each domain row that opens the Learning Theory Hub directly to that specific domain, so that I can jump straight to the topic I need to review.
28. As a Learner on the Scenario Result page after completing a Simulation Session, I want a Review Domain Theory action that opens the Learning Theory Hub directly to the Scenario's domain, so that I can remediate any missed Key Findings or Response Actions before starting the next Scenario.
29. As a Learner, I want reading or completing Theory Modules not to change my Proficiency Score, so that my Proficiency Score accurately reflects my assessment and simulation performance.
30. As a Specialist, I want to view all five domain Theory Modules at any time without being required to take a Learner Pre-Test, so that I can reference the platform's curriculum when reviewing Learner submissions or writing Analysis Threads.
31. As an Admin, I want to view all five domain Theory Modules at any time without being required to take a Learner Pre-Test, so that I can audit the curriculum content.
32. As a Content Author, I want all five domain Theory Modules authored as structured Markdown documents in the repository, so that curriculum changes are version-controlled and human-reviewable.
33. As a Content Author, I want every external reference used in the Theory Modules registered in the project's canonical reference registry and BibTeX bibliography before ingestion, so that academic provenance and citation standards are strictly maintained.
34. As a Researcher, I want the simulation status and research summary endpoints to report whether a participant has completed the five Theory Modules, which domains they completed, and when they finished, so that instructional completion can be verified in the experimental dataset.
35. As a Maintainer, I want the Theory Modules seeded automatically during container startup and backed by an embedded fallback in the application binary, so that neither fresh deployments nor isolated test runs can fail due to missing theory rows.
36. As a Maintainer, I want a configuration flag to bypass the Theory Gate when running isolated simulation tests that intentionally skip the theory step, so that targeted test suites remain flexible while the default gate stays fail-closed.

## Implementation Decisions

### Domain model and glossary additions

- **Theory Module**: the structured English curriculum document for one of the five CompTIA Security+ (SY0-701) domains (`domain1` through `domain5`), combining exam-aligned theory and Blue Team operational guidance.
- **Theory Gate**: the two-stage progression rule enforcing that (1) Pre-Test completion unlocks the Learning Theory Hub for Learners, and (2) completing all five Theory Modules (`5/5`) unlocks starting Simulation Sessions.
- **Domain Theory Completion**: a persisted record indicating that a Learner has completed studying a specific domain's Theory Module, along with the completion timestamp.
- **Recommended Focus Domain**: any domain where the Learner's current Proficiency Score is below the 70% Proficiency Threshold (or, if all domains meet the threshold, the single domain with the lowest Proficiency Score).

### Curriculum authoring and provenance contract

- Five canonical English Markdown documents are authored in the repository, one per CompTIA Security+ (SY0-701) domain (`domain1` through `domain5`).
- Before parsing or database ingestion, all authoritative reference sources used to synthesize the curriculum — CompTIA Security+ SY0-701 Exam Objectives, NIST Special Publication 800-61 Rev. 2 and SP 800-53 Rev. 5, the MITRE ATT&CK Enterprise Matrix, OWASP Top 10, and public computer security reference articles — must be registered with permanent `R-ID` keys in the project's reference registry and matching entries in the BibTeX bibliography.
- Every domain's Theory Module must contain four structured sections in English:
  1. **Core Security+ Theory & Exam Concepts**: foundational principles, definitions, architectures, and governance concepts aligned with the Pre-Test and Post-Test blueprint.
  2. **SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping**: concrete log sources, field names, KQL search patterns, and MITRE ATT&CK techniques relevant to investigating incidents in that domain.
  3. **Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide**: criteria for distinguishing True Positives from False Positives (including Benign Context Logs) and selecting proportional Containment, Eradication, and Recovery Response Actions while avoiding destructive distractors.
  4. **Key Takeaways & References**: a pre-simulation review checklist and inline `[R-ID]` citations.

### Storage and seeding architecture

- Theory Module documents are stored in the existing knowledge-base document table keyed by domain identifier and title, and are also embedded into the backend binary as a deterministic fallback so the API always serves all five domains even if database seeding has not yet run.
- A dedicated seeding script validates that all five domain Markdown documents exist, are non-empty, and contain the required sections, then idempotently upserts them into the database. The one-shot seed container runs this script automatically during stack initialization.
- Learner progress is stored in a dedicated `user_theory_progress` relation keyed by composite primary key `(user_id, domain_id)` with a `completed_at` timestamp. Both the initial schema script and the startup migration runner ensure this table exists idempotently.

### Gate enforcement and API contracts

- All new endpoints use dedicated request and response DTOs and include OpenAPI annotations per ADR-0005.
- **List Theory Modules endpoint (`GET /api/learn/domains`)**:
  - Requires authentication.
  - For a `learner` who has not completed the Pre-Test (no skill profile rows exist), returns `403 Forbidden` with reason `pre_test_required` and `unlocked: false`.
  - For `specialist` and `admin` roles, or for any `learner` who has completed the Pre-Test, returns `200 OK` with `unlocked: true`, `theory_completed` boolean (`completed_count == 5`), `completed_count`, `total_domains: 5`, `completed_domains` array, and the ordered list of five domain objects (each containing domain metadata, full Markdown content, source reference, current `proficiency_score`, `threshold`, `passed_threshold`, `recommended_focus`, `is_completed`, and `completed_at`).
- **Complete Domain Theory endpoint (`POST /api/learn/domains/:domainId/complete`)**:
  - Requires authentication.
  - Validates that `domainId` is one of `domain1` through `domain5` (`400 Bad Request` otherwise).
  - For a `learner` without a completed Pre-Test, returns `403 Forbidden` with reason `pre_test_required`.
  - Idempotently records `(user_id, domain_id, completed_at)` without modifying `user_skill_profiles.proficiency_score`, and returns the updated completion summary (`is_completed: true`, `completed_domains`, `completed_count`, `total_domains: 5`, `theory_completed`).
- **Simulation Gate (`POST /api/simulation/start`)**:
  - When a Learner attempts to start a new Simulation Session (and no active `in_progress` session is being resumed), checks whether all five domains are recorded as completed in `user_theory_progress` (unless the environment bypass flag `BYPASS_THEORY_GATE=true` is set).
  - If fewer than five domains are completed, returns `403 Forbidden` with reason `theory_incomplete`, along with `completed_count`, `total_domains: 5`, and `completed_domains`.
- **Status and Research Analytics exposure (`GET /api/simulation/status` and `GET /api/analytics/research-summary`)**:
  - Both endpoints include `theory_completed` (`bool`) and `completed_theory_domains` (`[]string`), and the research summary additionally includes `theory_completed_at` (the timestamp when the fifth domain was completed, if applicable).

### Frontend experience and navigation flow

- The top navigation bar includes a `Learn` link positioned between `Exam` and `Evaluation`.
- The Learning Theory Hub page supports selecting any of the five domains via tabs/sidebar and supports deep-linking to a specific domain via a `domain` query parameter.
- When locked (`pre_test_required`), the Learning Theory Hub displays an assessment-required card directing the Learner to the Pre-Test.
- When unlocked, the Learning Theory Hub displays:
  - A progress header (`X / 5 Domains Completed`) and, once `5 / 5` is reached, a primary **Start Adaptive Simulation** button.
  - Domain navigation tabs showing each domain's completion checkmark, current Proficiency Score, and a **Recommended Focus** badge where applicable.
  - A dark-mode graphite Markdown reader rendering headings, tables, callouts, and monospace telemetry/KQL blocks, with a **Mark Domain as Completed** button and previous/next domain navigation at the bottom.
- On the Skill Evaluation page:
  - Each domain entry provides a **Study Theory** link deep-linking to that domain on the Learning Theory Hub.
  - While `theory_completed` is `false`, the primary call-to-action directs the Learner to complete the Learning Theory Hub (`X/5 Domains Completed`) rather than starting a Simulation Session directly. Once `theory_completed` is `true`, the **Start Recommended Scenarios** button is unlocked.
- On the Scenario Result page:
  - A **Review Domain Theory** button deep-links to the completed Scenario's domain in the Learning Theory Hub so the Learner can review relevant theory before starting their next Scenario.

## Testing Decisions

- **What makes a good test:** Tests must assert observable behavior through public contracts — HTTP status codes, JSON response shapes, gate transitions, and proficiency score invariance — rather than internal implementation details.
- **Primary seam (existing, highest): the HTTP simulation integration flow.** Extend the end-to-end Python simulation integration suite (`scripts/test_simulation_integration.py`) to verify the complete progression:
  1. Calling the Learning Theory endpoints before completing the Pre-Test returns `403 Forbidden` (`pre_test_required`).
  2. Completing the 30-question Pre-Test unlocks `GET /api/learn/domains`, returning all five domains with non-empty content, learner proficiency scores, and `theory_completed: false`.
  3. Attempting `POST /api/simulation/start` before completing all five Theory Modules returns `403 Forbidden` (`theory_incomplete`).
  4. Completing `domain1` through `domain5` via `POST /api/learn/domains/:domainId/complete` is idempotent, transitions `theory_completed` to `true`, and leaves `proficiency_score` unchanged.
  5. After `5/5` Theory Modules are completed, `POST /api/simulation/start` succeeds (`201 Created`), and both `GET /api/simulation/status` and `GET /api/analytics/research-summary` report `theory_completed: true` with all five domains listed.
- **Curriculum and gate unit seam (existing Go test runner):** Add table-driven Go tests verifying that all five domain Theory Modules load with non-empty English content and required sections, that `recommended_focus` flags domains below the 70% threshold (or the lowest domain when all pass), and that the Theory Gate enforces both `pre_test_required` and `theory_incomplete` invariants cleanly.
- **Frontend verification:** Verify with the TypeScript compiler and production Vite build (`npm run build`) to guarantee zero type or bundling errors.

## Out of Scope

- Scored quizzes or assessments inside the Learning Theory Hub that alter `user_skill_profiles.proficiency_score`.
- Real-time LLM generation of theory content when a Learner opens the page.
- Forced minimum timer lockouts or scroll-position locks before a Learner can click **Mark Domain as Completed**.
- Changes to the 30-question Pre-Test / Post-Test question bank or the Scenario evaluation rubric (ADR-0004).
- Multi-language runtime translation toggles (curriculum content is standardized in English to match exam questions and SIEM telemetry).

## Further Notes

- **Workflow source of truth:** Detailed workflow and contract rules were established during the `/loop-me` session and recorded in `workflows/learning_theory_hub.md` and `NOTES.md`.
- **ADR alignment:** Respects ADR-0004 (Proficiency Score and Scenario evaluation remain purely in the backend evaluator without drift from reading activity) and ADR-0005 (code-first OpenAPI annotations with dedicated DTOs in `dto/`).
- **Originating issue:** #26, "New Feature: learning theory page".
