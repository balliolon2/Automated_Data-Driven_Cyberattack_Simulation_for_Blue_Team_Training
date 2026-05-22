# Project Research Summary

**Project:** Automated Data-Driven Cyberattack Simulation for Continuous Blue Team Workforce Training
**Domain:** Cyberattack Simulation & Training Platform
**Researched:** 2026-05-22
**Confidence:** HIGH

## Executive Summary

The project is an automated, data-driven cyberattack simulation platform tailored for Blue Team (SOC) workforce training. Experts in this domain typically build training systems using expensive, complex live VMs and real network traffic. However, this platform takes an alternative approach: it uses a lightweight web-based UI mimicking real enterprise tools (Splunk, TheHive) combined with mock data to provide a safe, scalable, and low-cost environment focused on analytical workflows.

The recommended technical approach leverages a React SPA for the mock SOC interface, backed by a high-concurrency Node.js API. For its core differentiators—AI mentoring and scenario generation—it relies on PostgreSQL with pgvector for low-latency (<2s) Retrieval-Augmented Generation (RAG) and BullMQ for asynchronous (<30s) scenario creation. Crucially, an "expert-in-the-loop" workflow is enforced, requiring manual approval of all AI-generated scenarios before activation.

The major risks lie in AI hallucinations breaking the logical sequence of an attack simulation and the platform accidentally incentivizing "gamified" guessing rather than teaching real SOC skills. To mitigate these, the system must employ strict programmatic constraints on AI outputs, enforce detailed logging and scoring of specific analytical actions (like Splunk search queries), and ensure the UI strictly mimics dense, tabular enterprise software rather than looking like a modern consumer app.

## Key Findings

### Recommended Stack

**Core technologies:**
- **React + Vite + Tailwind CSS:** SPA Frontend — Dictated by constraints; provides rapid UI prototyping essential for mimicking SIEM/SOAR dashboards.
- **Node.js + Hono (or Fastify):** API Backend — Ensures high performance and low overhead, crucial for hitting the strict <2s RAG response time target.
- **PostgreSQL + pgvector + Drizzle ORM:** Database & Vector Store — Relational schema needed for system state, while native HNSW indexing enables fast vector searches without needing a separate vector DB.
- **BullMQ + Redis:** Job Queue — Best-in-class robust queue to handle the <30s async scenario generation without blocking the main API.
- **GCP Cloud Run + Cloud SQL:** Infrastructure — Scales instantly to meet the 50 concurrent learners constraint and scales to zero when unused.

### Expected Features

**Must have (table stakes):**
- Interactive SOC UI (SIEM/SOAR mimic) — users expect this to triage incidents.
- Scenario-Based Training — generic questions are insufficient; requires concrete incidents.
- Action Scoring & Feedback — learners need to know if they took the right actions.
- Skill Gap Tracking — maps training to industry standards (CompTIA, MITRE).

**Should have (competitive):**
- AI-Generated Adaptive Scenarios — differentiator to prevent content staleness.
- Contextual RAG-Based Mentoring — differentiator providing cited, instant feedback.
- Expert-in-the-Loop Quality Control — solves AI hallucination and trust concerns.
- Proficiency-Based Pre-tests — tailors the learning path dynamically.

**Defer (v2+):**
- Custom Tool Mimicry — beyond the core Splunk/TheHive mimics.
- Team-based Collaborative Response — multi-player complexity distracts from core Tier 1 training.

### Architecture Approach

The architecture separates the high-fidelity UI mimics from heavy AI operations using a robust API and background worker pattern. The presentation layer handles complex virtual workspaces without heavy client-side state.

**Major components:**
1. **Virtual Workspace (SIEM/Case Mock)** — Mimics Splunk and TheHive for log analysis.
2. **Session Manager & Scoring Engine** — Maintains the state machine of the simulation session and evaluates user actions for correctness.
3. **Scenario Builder & RAG Workers** — Async queues (BullMQ) handle LLM scenario synthesis, while direct pgvector integrations drive sub-2-second cited feedback.

### Critical Pitfalls

1. **AI Hallucinations Breaking Scenario Logic** — Avoid by forcing the AI to assemble pre-validated components rather than writing from scratch, and requiring expert approval.
2. **Gamification Flaws** — Avoid by scoring detailed query construction and evidence logging rather than just tracking clicks. Do not teach users to "brute force" the UI.
3. **The "Video Game" UI Trap** — Avoid by prioritizing data density, tabular layouts, and raw logs (JSON/Syslog) over modern, polished consumer graphics.
4. **Overwhelming the Tier 1 Learner** — Avoid by strictly mapping scenarios to entry-level domains (CompTIA 701) and providing clear breadcrumbs.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Foundation & Core Platform UI
**Rationale:** Establishes the database schema, auth, and the core visual mimicry (SIEM/TheHive) which is required for any scenario to be played.
**Delivers:** DB schema, authentication, basic API, Splunk/TheHive UI mocks, static MVP scenarios.
**Addresses:** Interactive SOC UI, Basic Action Scoring.
**Avoids:** The "Video Game" UI Trap.

### Phase 2: RAG Pipeline & AI Scenario Engine
**Rationale:** AI features are the core differentiators but require the foundation to be testable. RAG provides the mentoring, and the builder supplies the content.
**Delivers:** Embedding pipeline, RAG retrieval (<2s), async scenario generation (<30s), Admin UI for expert approval.
**Uses:** Node.js, Vercel AI SDK, BullMQ, pgvector.
**Implements:** Scenario Builder, RAG & Feedback.
**Avoids:** AI Hallucinations, Synchronous LLM generation timeouts.

### Phase 3: Simulation Session & Scoring Engine
**Rationale:** With scenarios generated (or static) and UI ready, the active simulation logic can be linked to evaluate users in real-time.
**Delivers:** Session Manager (State Machine), precise log exposure rules, action evaluation, points calculation.
**Uses:** React (Frontend API Client), Node.js (Core Logic).
**Implements:** State Machine for Simulation Session.
**Avoids:** Client-Side Scenario Logic, Gamification Flaws (scoring actual `session_actions`).

### Phase 4: Adaptive Learning & Dashboards
**Rationale:** Adds advanced routing and visual polish based on the telemetry generated in Phase 3.
**Delivers:** Proficiency pre-tests, skill gap visualizations, dynamic scenario unlocking.
**Addresses:** Skill Gap Tracking, Proficiency-Based Pre-tests.
**Avoids:** Overwhelming the Tier 1 Learner.

### Phase Ordering Rationale

- Data and UI (Phase 1) must exist before any scenarios can be tested or played.
- AI Generation (Phase 2) is entirely backend-focused and can be developed to populate the database with pending scenarios.
- The Simulation Engine (Phase 3) bridges the UI and the Data by running the state machine.
- Dashboards (Phase 4) rely on the scoring data generated by completed sessions.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 2:** Structuring the exact JSON output schemas for the LLM to generate valid scenario states; robust handling of BullMQ on serverless platforms (Cloud Run).
- **Phase 3:** Designing the exact state machine rules to expose simulated logs progressively without leaking the full dataset to the client.

Phases with standard patterns (skip research-phase):
- **Phase 1:** Standard React/Tailwind frontend setup and PostgreSQL DBML/Drizzle schema creation.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | React/Vite/Tailwind and Node/Hono perfectly align with the strict performance requirements and constraints. |
| Features | HIGH | Clear delineation between necessary mock features and out-of-scope live network complexities. |
| Architecture | HIGH | Standard asynchronous worker patterns and pgvector RAG are well-proven for these exact constraints. |
| Pitfalls | HIGH | Deeply specific to the cybersecurity training domain, focusing on the danger of bad training habits and hallucinations. |

**Overall confidence:** HIGH

### Gaps to Address

- **Simulated Log Structures:** Need to define exactly what the mocked JSON/CSV log schemas look like for the MVP before the AI takes over generation.
- **LLM Prompt Engineering:** Need to formulate the strict system prompts and Zod validation schemas to ensure the AI adheres to the `<30s` scenario structure.
- **UI Interaction Limits:** How strictly should the Splunk search syntax (SPL) be validated? If too strict, users will fail on syntax; if too loose, they aren't learning.

## Sources

### Primary (HIGH confidence)
- `PROJECT.md` Constraints — Framework choices, AI approval rules, and response time metrics.
- STACK.md — Stack recommendations and performance rationales.
- FEATURES.md — Core features, differentiators, and MVP definition.
- ARCHITECTURE.md — Data flow, scaling, and async queuing patterns.
- PITFALLS.md — Security risks, gamification flaws, and AI hallucination mitigations.

---
*Research completed: 2026-05-22*
*Ready for roadmap: yes*