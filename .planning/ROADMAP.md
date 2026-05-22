# Project Roadmap

## Phases
- [ ] **Phase 1: Foundation & Mock Interfaces** - Establish auth, data security, scalable infra, and core SOC dashboards
- [ ] **Phase 2: AI Scenarios & RAG Mentoring** - Automate scenario generation and provide sub-2s cited AI feedback
- [ ] **Phase 3: Interactive Simulation & Scoring** - Enable mock tool workflows and score learner actions
- [ ] **Phase 4: Adaptive Routing & Analytics** - Tailor learning paths and visualize skill gaps

## Phase Details

### Phase 1: Foundation & Mock Interfaces
**Goal**: Users can securely log in and view the fundamental SOC workspace interfaces
**Mode:** mvp
**Depends on**: None
**Requirements**: AUTH-01, AUTH-02, SIM-01, SIM-02, SCEN-03
**Success Criteria**:
  1. User can authenticate as a Learner, Expert, or Admin with role-based access.
  2. User can view the Splunk-mimic and TheHive-mimic dashboards.
  3. System database is established with encryption for user data at rest.
  4. Infrastructure deploys successfully on GCP with autoscaling configuration.
**Plans**: TBD
**UI hint**: yes

### Phase 2: AI Scenarios & RAG Mentoring
**Goal**: Experts can approve system-generated scenarios, and learners can query theory-backed AI guidance
**Mode:** mvp
**Depends on**: Phase 1
**Requirements**: SCEN-01, SCEN-02, EVAL-03, EVAL-04
**Success Criteria**:
  1. Background system successfully generates structured scenario data in under 30 seconds.
  2. Expert user can review, approve, or reject generated scenarios via an administrative interface.
  3. User receives RAG-based answers to simulation queries within 2 seconds.
  4. User can verify that every AI mentor response cites a specific CompTIA domain or MITRE technique.
**Plans**: TBD
**UI hint**: yes

### Phase 3: Interactive Simulation & Scoring
**Goal**: Learners can actively triage incidents through the mock tools and earn points for correct actions
**Mode:** mvp
**Depends on**: Phase 2
**Requirements**: SIM-03, EVAL-02
**Success Criteria**:
  1. User can construct and execute search queries that return relevant mock logs in the SIEM UI.
  2. User can take triage and escalation actions in the SOAR UI.
  3. System automatically records every user action and assigns a correctness score and points in real-time.
**Plans**: TBD
**UI hint**: yes

### Phase 4: Adaptive Routing & Analytics
**Goal**: Learners start with appropriate scenarios based on their baseline and see personalized improvement plans
**Mode:** mvp
**Depends on**: Phase 3
**Requirements**: EVAL-01, EVAL-05
**Success Criteria**:
  1. New user is prompted to complete a domain-based pre-test.
  2. User only has access to scenarios matching their unlocked proficiency domains.
  3. After completing a scenario, user sees a visual breakdown of their skill gaps.
  4. User receives actionable recommendations for specific domains to practice next.
**Plans**: TBD
**UI hint**: yes

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Mock Interfaces | 0/0 | Not started | - |
| 2. AI Scenarios & RAG Mentoring | 0/0 | Not started | - |
| 3. Interactive Simulation & Scoring | 0/0 | Not started | - |
| 4. Adaptive Routing & Analytics | 0/0 | Not started | - |