---
title: "Spec 2: Specialist Submission Review Console & Scenario-Linked Analysis Threads"
labels: ["ready-for-agent"]
status: "ready-for-agent"
---

# Spec 2: Specialist Submission Review Console & Scenario-Linked Analysis Threads

## Problem Statement

While the platform captures comprehensive log interactions, alert triage classifications, response actions, and evaluation scores from learner simulation sessions, this data remains inaccessible to human subject-matter experts. Specialists cannot review learner investigation patterns to detect common misunderstandings or recurring analytical pitfalls. Concurrently, learners completing scenarios receive only automated scoring rubrics, without human context, expert reasoning, or real-world Blue Team guidance. There is also no dedicated medium within the application for specialists to write and publish in-depth case breakdowns (กระทู้) linked to specific simulation scenarios.

## Solution

1. Provide a Specialist Review Console (`/specialist/reviews`) allowing verified Specialists to browse and inspect completed simulation submissions filtered by scenario, domain, or low score thresholds, strictly preserving learner privacy by displaying only user Nicknames.
2. Build an Analysis Thread authoring engine for Specialists and Platform Admins featuring a rich Markdown editor, log/KQL query syntax highlighting, and a quick-insert template bar that injects scenario blueprint details (alert telemetry, key findings checklist, expected response actions).
3. Connect expert analyses directly into the learner workflow by displaying a "Specialist Insights & Community Discussions" panel on the `ScenarioResultPage` immediately following scenario completion, while also indexing threads in a central `/discussions` catalog.

## User Stories

1. As a specialist, I want to access a dedicated Review Console, so that I can inspect how learners are solving security simulation incidents.
2. As a specialist, I want to filter learner submissions by scenario, security domain, or score (e.g., sessions scoring < 70%), so that I can focus on areas where learners struggle the most.
3. As a specialist, I want to drill down into an individual learner's simulation session to examine their exact TP/FP triage choice, discovered key findings, response actions, and log query history, so that I can identify their analytical breakdown points.
4. As a learner, I want my identity in the specialist review console to be masked with my public Nickname, so that my personal email and profile remain private.
5. As a specialist, I want to draft a new Analysis Thread directly from a scenario or submission review, so that I can capture educational insights while the context is fresh.
6. As a specialist, I want to write my analysis using full Markdown (including headers, lists, and code blocks), so that I can present clear, structured investigation guides.
7. As a specialist, I want code syntax highlighting for KQL queries, syslog dumps, and JSON event logs, so that code snippets in my post are readable.
8. As a specialist, I want a quick-insert template button to pull in the scenario's alert narrative, key findings checklist, and recommended playbooks, so that I do not have to manually transcribe scenario baselines.
9. As a specialist, I want to publish my Analysis Thread with associated tags and an optional link to a specific `scenario_id`, so that learners practicing that scenario can easily locate it.
10. As a specialist, I want to edit or delete my own published analysis threads, so that I can keep my content up to date or correct mistakes.
11. As a learner completing a simulation scenario, I want to see a "Specialist Insights" section at the bottom of my `ScenarioResultPage`, so that I can immediately read expert writeups comparing different investigative approaches.
12. As a learner, I want to browse a central `/discussions` feed filterable by domain, scenario, or popularity, so that I can learn about scenarios I have not yet attempted or review historical case studies.
13. As a platform administrator, I want to edit, unpublish, or delete any thread that violates content guidelines, so that platform standards are upheld.

## Implementation Decisions

### Access Control & Privacy
- Access to the Review Console (`/specialist/reviews`) and submission inspection endpoints is restricted to users whose JWT claims have `role = 'specialist'` or `role = 'admin'`.
- All learner submissions returned by review APIs join against the `users` table to project only the user's `nickname` and tier, strictly excluding `email`, `user_id`, or personally identifiable information.

### Schema Specifications

- **Analysis Threads Table (`analysis_threads`)**:
  - `thread_id`: UUID PK, default `gen_random_uuid()`
  - `author_id`: UUID FK referencing `users(user_id)` with `ON DELETE CASCADE`
  - `scenario_id`: UUID FK referencing `scenarios(scenario_id)` with `ON DELETE SET NULL` (nullable for general security topics)
  - `title`: VARCHAR(255) NOT NULL
  - `content`: TEXT NOT NULL (stored as GitHub Flavored Markdown)
  - `tags`: VARCHAR[] (e.g., `{'brute-force', 'triage-error', 'domain4'}`)
  - `upvote_count`: INT DEFAULT 0
  - `view_count`: INT DEFAULT 0
  - `is_pinned`: BOOLEAN DEFAULT false
  - `is_locked`: BOOLEAN DEFAULT false
  - `created_at`: TIMESTAMP default `now()`
  - `updated_at`: TIMESTAMP default `now()`

### API Contracts

- `GET /api/specialist/submissions`
  - Authenticated (`specialist` or `admin`).
  - Query parameters: `scenario_id` (optional), `domain_id` (optional), `max_score` (optional, for finding low-score sessions), `page`, `page_size`.
  - Response: Paginated list of completed sessions with `session_id`, learner `nickname`, `scenario_title`, `final_score`, `tp_fp_correct`, `findings_found_count`, `total_findings_count`, and `completed_at`.
- `GET /api/specialist/submissions/:session_id`
  - Authenticated (`specialist` or `admin`).
  - Response: Comprehensive breakdown including the scenario baseline, learner triage choice (`triage_alert`), response actions taken (`respond`), discovered findings (`submit_decision`), and chronological action logs.
- `POST /api/threads`
  - Authenticated (`specialist` or `admin`).
  - Payload: `{"title": "string", "content": "string", "scenario_id": "uuid" (optional), "tags": ["string"]}`.
  - Response: 201 Created with the created thread object.
- `GET /api/threads`
  - Authenticated (Any role).
  - Query parameters: `scenario_id`, `domain_id`, `tag`, `sort` (`latest`, `upvotes`), `page`, `page_size`.
  - Response: Paginated list of threads with author nickname, role, title snippet, scenario reference, upvote count, comment count, and timestamps.
- `GET /api/threads/:id`
  - Authenticated (Any role). Increments `view_count` atomically.
  - Response: Complete thread document including author metadata (`nickname`, `role`), full markdown content, scenario reference summary, and timestamps.
- `PUT /api/threads/:id`
  - Authenticated (Author or `admin`).
  - Payload: `{"title": "string", "content": "string", "tags": ["string"]}`.
  - Response: 200 OK with updated thread.
- `DELETE /api/threads/:id`
  - Authenticated (Author or `admin`).
  - Response: 200 OK with deletion confirmation.

## Testing Decisions

- **Good Test Criteria**: Tests must verify behavior exclusively through the HTTP REST API boundary (Highest Seam), ensuring that non-specialists receive 403 Forbidden on review and authoring endpoints, that learner PII is absent from submission responses, and that threads are accurately queryable by `scenario_id`.
- **Tested Modules**: Specialist Submission Controller, Analysis Thread Controller, Role Authorization Middleware, and Result Page Discussion Integrator.
- **Prior Art**: Follows the existing simulation session evaluation tests in `controllers/simulation_controller.go` and `scripts/test_integration.py`.

## Out of Scope

- AI automated generation or grading of analysis threads.
- Video recording or real-time audio playback within threads.
- Direct private messaging between specialists and learners.

## Further Notes

- Threads linked to a `scenario_id` are automatically prioritized on the `ScenarioResultPage` when a learner finishes that scenario.
- Triage label: `ready-for-agent`
