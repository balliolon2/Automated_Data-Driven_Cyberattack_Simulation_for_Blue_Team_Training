# Feature Research

**Domain:** Cyberattack Simulation & Blue Team Training
**Researched:** 2026-05-22
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Interactive SOC UI (SIEM/SOAR mimic) | A SOC platform must feel like a SOC. Users expect Splunk-like query capability and TheHive-like case management. | HIGH | Core to the simulation loop. Requires mocking standard tools. |
| Scenario-Based Training | Learning happens by doing. Generic questions aren't enough; users need concrete incidents (e.g., phishing, ransomware) to triage. | MEDIUM | Can start static and move to AI-generated later. |
| Action Scoring & Feedback | Learners need to know if they took the right actions. Correctness flags + points confirm they are learning. | MEDIUM | Requires defining specific success criteria per scenario step. |
| Skill Gap Tracking (Framework Mapping) | Training must map to industry standards (CompTIA, MITRE ATT&CK) for certifications and real-world relevance. | LOW | Handled via DB schema (`mitre_techniques`, `security_domains`). |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| AI-Generated Adaptive Scenarios | Stops content from becoming stale. Generates non-repetitive incidents tailored to learner's skill level. | HIGH | Requires async job queue (<30s generation). Core project differentiator. |
| Contextual RAG-Based Mentoring | Replaces generic hints with theory-backed, cited feedback (CompTIA/MITRE) instantly during scenarios. | HIGH | Needs pgvector HNSW index. Must respond in <2s with citations. |
| Expert-in-the-Loop Quality Control | Solves AI hallucination concerns. Ensures all AI-generated training material is vetted by an expert before deployment. | LOW | Simple DB state flag (`scenario_status=active`), but critical for trust. |
| Proficiency-Based Pre-tests | Adapts the learning path immediately. Unlocks relevant scenarios without forcing users through things they already know. | MEDIUM | Connects `test_attempts` to dynamic scenario unlocking. |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Real Network Traffic Ingestion & Live VMs | "True realism" requires real machines, live malware, and active SIEM ingestions. | Immensely expensive, complex orchestration, high security risk. Out of scope for this web-based Tier 1 simulation. | Data-driven mock logs presented in the simulated UI. |
| Red vs. Blue (Multiplayer) | Sounds fun and competitive (e.g., standard cyber ranges). | Too complex to balance. Requires synchronous players. Detracts from fundamental Tier 1 SOC training. | Asynchronous, system-generated solo training ranges. |
| Bring Your Own Tools (BYOT) | Let users connect their own EDR or SIEM. | Integration nightmare, impossible to standardize scoring or RAG feedback. | Built-in standard interfaces mimicking common tools (Splunk). |

## Feature Dependencies

```text
[AI-Generated Scenarios]
    └──requires──> [Expert Approval Workflow]

[Contextual RAG Feedback]
    └──requires──> [Domain Knowledge Base (rag_documents)]

[Proficiency-Based Pre-tests]
    └──requires──> [Skill Gap Tracking (Framework Mapping)]

[Interactive SOC UI]
    └──enhances──> [Action Scoring & Feedback]
```

### Dependency Notes

- **[AI-Generated Scenarios] requires [Expert Approval Workflow]:** Project constraint dictates all AI content requires an expert approval flag before becoming active.
- **[Contextual RAG Feedback] requires [Domain Knowledge Base]:** The pgvector database must be loaded with CompTIA/MITRE exam MD files to ensure citations are valid.
- **[Proficiency-Based Pre-tests] requires [Skill Gap Tracking]:** Pre-tests need a mapped framework to accurately unlock corresponding domain scenarios.

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [ ] Interactive SOC UI (SIEM/SOAR mimic) — Essential for the user to perform the "triage, escalate, respond" workflow.
- [ ] Static or Semi-Automated Scenarios — Need at least one functioning scenario type to prove the UI and scoring works.
- [ ] Contextual RAG Feedback (Manual queries) — Validates the AI mentoring differentiator (must hit the <2s requirement with citations).
- [ ] Basic Action Scoring — Required to prove we can track correctness and points.

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] AI-Generated Adaptive Scenarios — Full async generation (<30s) and expert approval pipeline.
- [ ] Proficiency-Based Pre-tests — To tailor the learning paths.
- [ ] Skill Gap Visualization Dashboard — Post-session visual feedback.

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] Custom Tool Mimicry — Adding mimics for other tools beyond Splunk/TheHive (e.g., CrowdStrike, Sentinel).
- [ ] Team-based Collaborative Response — Multiplayer blue-teaming.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Interactive SOC UI | HIGH | HIGH | P1 |
| Contextual RAG Mentoring | HIGH | HIGH | P1 |
| Basic Action Scoring | HIGH | LOW | P1 |
| Expert Approval Workflow | HIGH | LOW | P1 |
| AI-Generated Scenarios | HIGH | HIGH | P2 |
| Proficiency Pre-tests | MEDIUM | MEDIUM | P2 |
| Skill Gap Dashboard | HIGH | MEDIUM | P2 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | Traditional Cyber Ranges (RangeForce, HackTheBox) | Standard LMS (Pluralsight, Coursera) | Our Approach |
|---------|---------------------------------------------------|--------------------------------------|--------------|
| **Environment** | Expensive Live VMs, Real Tooling | Static Videos, Multiple Choice | Simulated Web UI (Fast, cheap, safe) |
| **Scenarios** | Hardcoded, finite supply | Static Case Studies | AI-Generated, Infinite & Adaptive |
| **Mentoring** | Mentors or Walkthrough PDFs | None | In-platform RAG AI referencing CompTIA/MITRE |
| **Scoring** | Flag capture (CTF style) | Quiz Scores | Granular session action logging & skill gap tracking |

## Sources

- `PROJECT.md` Requirements and Constraints.
- Market analysis of existing platforms like RangeForce and LetsDefend.
- Standard Blue Team (Tier 1 SOC) training requirements.

---
*Feature research for: Cyberattack Simulation & Blue Team Training*
*Researched: 2026-05-22*