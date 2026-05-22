# Automated Data-Driven Cyberattack Simulation for Continuous Blue Team Workforce Training

## What This Is

A web-based cyberattack simulation platform that trains Blue Team learners (SOC Analyst Tier 1) through realistic SOC Analyst workflows. It uses Data-Driven + AI (RAG) to generate adaptive, non-repetitive scenarios and provides theory-backed feedback to track skill gaps.

## Core Value

Trains Blue Team learners through realistic, adaptive SOC Analyst workflows using AI-generated scenarios and theory-backed feedback, effectively tracking and bridging skill gaps.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

(None yet — ship to validate)

### Active

<!-- Current scope. Building toward these. -->

- [ ] User can take pre-test → unlock scenarios based on domain proficiency
- [ ] During simulation: UI mimics Splunk/TheHive; user can triage, query RAG, escalate, respond
- [ ] System logs every session_action with correctness flag + points
- [ ] RAG retrieves from rag_documents + exam MD files to generate contextual feedback
- [ ] Post-session: Show skill_gap visualization + recommend domains for practice
- [ ] Response time <2s for RAG queries (pgvector HNSW index)
- [ ] Scenario generation <30s (async job queue)
- [ ] Support 50 concurrent learners (GCP autoscaling config)
- [ ] All user data encrypted at rest; audit logs for expert review

### Out of Scope

<!-- Explicit boundaries. Includes reasoning to prevent re-adding. -->

- Real network traffic ingestion — simulation only.
- Storage of raw user credentials beyond auth hash — security guardrail.

## Context

- **Framework Alignment:** CompTIA Security+ 701 (5 Domains) + MITRE ATT&CK
- **Target User:** SOC Analyst Tier 1 (Beginner level)
- **Tech Stack:** React/Vite + TailwindCSS (Frontend), Node.js/Go + PostgreSQL (Backend), GCP (Cloud), pgvector + RAG (AI)
- **Database Schema:** Defined in DBML format with: users (roles: learner/expert/admin), security_domains, mitre_techniques, questions + test_attempts, scenarios + simulation_sessions + session_actions, rag_documents + feedback_logs, user_skill_profiles + evaluation_metrics + help_click_analytics.
- **Examination Assets:** `/examination/` folder contains 5 Markdown files for CompTIA domains.

## Constraints

- **Simulation Constraints**: No real network traffic ingestion (simulation only).
- **Data Privacy**: No storage of raw user credentials beyond auth hash. All user data encrypted at rest.
- **AI Feedback Constraints**: RAG responses must cite source (domain_id or mitre_ref).
- **Content Approval**: All AI-generated content requires expert approval flag before `scenario_status=active`.

## Key Decisions

<!-- Decisions that constrain future work. Add throughout project lifecycle. -->

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| RAG citations required | Ensures feedback is theory-backed and traceable to course material | — Pending |
| Expert approval for AI content | Maintains quality and accuracy of training scenarios before learner exposure | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-22 after initialization*