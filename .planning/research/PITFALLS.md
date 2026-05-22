# Pitfalls Research

**Domain:** Cyberattack Simulation Platform (AI/RAG-driven for Blue Team)
**Researched:** 2026-05-22
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: AI Hallucinations Breaking Scenario Logic

**What goes wrong:**
The AI generates a scenario where the sequence of events is technically impossible (e.g., data exfiltration happens before initial access, or internal IPs are used as external C2 servers), or the AI evaluates correct learner actions as incorrect due to misunderstanding the context.

**Why it happens:**
Over-reliance on the LLM's internal knowledge base without strict programmatic constraints, or unstructured prompt generation that lacks rigid SIEM telemetry schemas.

**How to avoid:**
Implement strict RAG constraints. The AI should assemble pre-validated scenario components and randomize the IOCs (IPs, hashes) rather than inventing the attack narrative from scratch. Enforce the Key Decision from PROJECT.md: *All AI-generated content requires expert approval flag before `scenario_status=active`*.

**Warning signs:**
- Test learners consistently flag scenarios as "confusing" or "unfair".
- Generated SIEM logs show timestamps out of chronological order.
- Feedback logs show AI referencing generic security advice instead of the provided examination Markdown files.

**Phase to address:**
Phase 2/3 (Scenario Generation / RAG Implementation)

---

### Pitfall 2: Teaching the Platform, Not the Job (Gamification Flaws)

**What goes wrong:**
Learners achieve high scores by guessing or brute-forcing the UI (e.g., clicking all escalation options until one works) instead of demonstrating actual SOC analysis skills. They learn to beat the simulation but fail in real environments.

**Why it happens:**
Scoring is heavily weighted towards simple completion or clicking predefined buttons, rather than analyzing query construction and analytical workflow.

**How to avoid:**
Log and evaluate every `session_action`. Penalize random clicking or blind guessing. Require learners to "show their work" by inputting specific SPL (Splunk Search Processing Language) or evidence artifacts before allowing an escalation.

**Warning signs:**
- High completion rates but low time-in-scenario.
- Learners repeatedly clicking multiple options rapidly in the session logs.
- High domain scores in the platform but poor performance in the CompTIA Security+ pre-test/post-test.

**Phase to address:**
Phase 3 (Simulation Session Engine & Scoring)

---

### Pitfall 3: The "Video Game" UI Trap

**What goes wrong:**
The simulation interface looks too polished, gamified, and modern (e.g., dark mode neon graphics, oversized buttons), failing to prepare the learner for the dense, tabular, text-heavy interfaces of actual SIEMs like Splunk, TheHive, or Microsoft Sentinel.

**Why it happens:**
Frontend developers build a standard modern web app instead of treating the UI as an intentional emulation of enterprise security software.

**How to avoid:**
Strict adherence to the "Mimics Splunk/TheHive" requirement. Prioritize data density, log tables, advanced search bars, and tabular data over visual flair. 

**Warning signs:**
- UI components look like a consumer app dashboard.
- Lack of keyboard shortcuts.
- Logs are abstracted away into simple summaries rather than showing raw JSON/Syslog formats.

**Phase to address:**
Phase 1/2 (Frontend Development & UI/UX Design)

---

### Pitfall 4: Overwhelming the Tier 1 Learner

**What goes wrong:**
Scenarios designed for beginners quickly spiral into advanced threat hunting exercises (e.g., requiring memory forensics or reverse engineering) that are completely out of scope for a Tier 1 SOC Analyst or CompTIA Security+ 701 candidate.

**Why it happens:**
Subject matter experts (or the AI) suffer from the "curse of knowledge" and fail to calibrate the difficulty to true beginners.

**How to avoid:**
Strict mapping of scenarios to specific CompTIA Security+ 701 domains and MITRE ATT&CK techniques. Ensure every scenario has clear, visible "breadcrumbs" (e.g., a glaring IDS alert to start the investigation).

**Warning signs:**
- Learners frequently abandoning scenarios without completing them.
- High volume of RAG/Help queries asking "Where do I even start?".

**Phase to address:**
Phase 2 (Content & Curriculum Design)

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Synchronous LLM calls for scenarios | Faster MVP development | Blocks UI, timeouts (>30s) | Never (Use async job queue) |
| Hardcoding MITRE IDs in UI code | Quick UI implementation | UI breaks when MITRE updates | MVP only, move to DB quickly |
| Storing simulated raw logs in DB | Easier to query for UI | Database bloat, slow queries | Never (Generate logs on the fly) |

## Integration Gotchas

Common mistakes when connecting to external services (LLMs, Vector DBs).

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| LLM (Feedback) | Accepting raw markdown/text output | Force Structured Outputs (JSON) and validate with Zod/Joi |
| pgvector (RAG) | Forgetting to add HNSW index | Explicitly build HNSW index to meet the <2s response requirement |
| LLM Prompting | Exposing system prompt to users | Implement strict guardrails against Prompt Injection ("Ignore previous instructions") |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Large Context Windows | LLM API costs spike, slow responses | Chunk examination MD files, only inject relevant chunks | >10 concurrent users |
| Missing Index on Session Logs | Slow post-session skill gap calc | Index `session_action` table on `user_id` and `domain_id` | >10k session actions |
| Heavy Client-side State | Browser tab freezes during sim | Keep simulation state in backend, stream updates to UI | Complex scenarios (1000+ logs) |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Prompt Injection by Learners | Learners manipulate AI to get max points | Sanitize user input before RAG, strict system prompt constraints |
| Leaking RAG Source Data | Users see the "answer key" | Ensure RAG retrieval does not directly output the expert answer key |
| Client-Side Scenario Logic | Users inspect browser network tab for answers | All correctness validation must happen server-side |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Blank Canvas Start | Users don't know what to do first | Always start scenarios with a simulated "Ticket" or "Alert" |
| Delayed Feedback | Users don't know they made a mistake until the end | Provide subtle real-time hints ("Are you sure this IP is internal?") |
| Unclear Skill Gaps | Users see a score but not how to improve | Map missed actions directly to CompTIA domains for targeted practice |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **RAG Integration:** Often missing relevance filtering — verify RAG rejects off-topic queries.
- [ ] **SIEM UI Mimicry:** Often missing text-selection copying — verify users can highlight and copy IP addresses from logs.
- [ ] **Async Scenario Generation:** Often missing progress indicators — verify UI shows "Generating..." while the job queue processes.
- [ ] **Skill Gap Visualization:** Often missing historical comparison — verify users can see improvement over multiple sessions.

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Bad AI Scenario goes Live | LOW | Admin flags scenario inactive, regenerates, invalidates scores |
| Prompt Injection Exploited | MEDIUM | Update system prompts, clear manipulated session logs |
| Slow RAG Responses (>2s) | MEDIUM | Re-index pgvector, reduce LLM context window size, add caching |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| The "Video Game" UI Trap | Phase 1 (Core Platform UI) | UI review against actual Splunk/TheHive screenshots |
| AI Hallucinations Breaking Logic | Phase 2 (Scenario Engine) | Expert review step implemented; test cases pass |
| Overwhelming Tier 1 Learner | Phase 2 (Scenario Engine) | CompTIA Domain mapping enforced in schema |
| Gamification Flaws | Phase 3 (Simulation & Scoring) | Anti-guessing logic implemented in `session_action` scoring |
| Client-Side Scenario Logic | Phase 3 (Simulation & Scoring) | Penetration testing of API endpoints for answer leaking |

## Sources

- Project Context (`PROJECT.md`) decisions and constraints.
- Established SOC simulation and cyber range best practices.
- AI/RAG implementation patterns for educational tech.
- General knowledge on SIEM UI/UX workflows.

---
*Pitfalls research for: Cyberattack Simulation Platform (AI/RAG-driven)*
*Researched: 2026-05-22*