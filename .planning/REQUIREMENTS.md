# Requirements: Automated Data-Driven Cyberattack Simulation

**Defined:** 2026-05-22
**Core Value:** Trains Blue Team learners through realistic, adaptive SOC Analyst workflows using AI-generated scenarios and theory-backed feedback, effectively tracking and bridging skill gaps.

## v1 Requirements

### Authentication & Authorization
- [ ] **AUTH-01**: User can log in as Learner, Expert, or Admin
- [ ] **AUTH-02**: All user data is encrypted at rest

### Simulation UI (Workspace)
- [ ] **SIM-01**: User can view a mock SIEM interface mimicking Splunk for log analysis
- [ ] **SIM-02**: User can view a mock ticketing interface mimicking TheHive for escalation
- [ ] **SIM-03**: User can execute search queries and actions (triage, investigate, respond) within the workspace

### Scenario Engine
- [ ] **SCEN-01**: System asynchronously generates AI cyberattack scenarios combining MITRE TTPs and CompTIA domains in under 30 seconds
- [ ] **SCEN-02**: Expert can review and approve/reject AI-generated content before it becomes active for learners
- [ ] **SCEN-03**: System supports 50 concurrent learners via GCP autoscaling

### Evaluation & Feedback
- [ ] **EVAL-01**: User can take a domain-based pre-test to unlock appropriate scenarios
- [ ] **EVAL-02**: System logs every session action with a correctness flag and assigned points
- [ ] **EVAL-03**: User can query for contextual RAG feedback during simulation with response time under 2 seconds
- [ ] **EVAL-04**: RAG feedback strictly cites source material (domain_id or mitre_ref) from course exams
- [ ] **EVAL-05**: User can view post-session skill gap visualization and domain practice recommendations

## v2 Requirements

### Analytics & Gamification
- **ANLY-01**: Leaderboards for learners based on scenario scores
- **ANLY-02**: Team-based simulation scenarios

## Out of Scope

| Feature | Reason |
|---------|--------|
| Live VM integration | Excluded due to cost, complexity, and explicit project constraints |
| Real network traffic ingestion | High security risk and operational overhead; mock logs required |
| Un-approved AI scenarios | Hallucination risk; expert approval is a mandatory quality gate |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| AUTH-01 | Phase 1 | Pending |
| AUTH-02 | Phase 1 | Pending |
| SIM-01 | Phase 1 | Pending |
| SIM-02 | Phase 1 | Pending |
| SIM-03 | Phase 3 | Pending |
| SCEN-01 | Phase 2 | Pending |
| SCEN-02 | Phase 2 | Pending |
| SCEN-03 | Phase 1 | Pending |
| EVAL-01 | Phase 4 | Pending |
| EVAL-02 | Phase 3 | Pending |
| EVAL-03 | Phase 2 | Pending |
| EVAL-04 | Phase 2 | Pending |
| EVAL-05 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 13 total
- Mapped to phases: 13
- Unmapped: 0 ✓

---
*Requirements defined: 2026-05-22*
*Last updated: 2026-05-22 after initial definition*