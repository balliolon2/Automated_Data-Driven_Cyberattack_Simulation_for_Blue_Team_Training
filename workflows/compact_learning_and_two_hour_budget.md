# Workflow Spec: Compact Learning Hub & 2-Hour End-to-End Training Budget (Issue #70)

**GitHub Issue:** [#70 — feat(learn,sim): condense Learning Hub topics and calibrate end-to-end 2-hour learner budget](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/70)  
**Status:** `/loop-me` Grilling Rounds 1–3 Settled (`Q1`–`Q3`, `Q7`–`Q8`, `Q13`–`Q14` Resolved, 2026-11-27) — Ready for Implementation  
**Depends On / Amends:**
- Amends `workflows/learning_subtopics_sample_questions.md` & `NOTES.md` Decision #38: Stage 2 **Theory Gate** transitions from `28/28` Theory Sub-topics to **`5/5` Concise Domain Essentials**, retaining all 28 Sub-topics as Optional Deep Dive (zero DB table deletion, zero NFR seed regression).
- Preserves `docs/SCORING_AND_PROFICIENCY_CONTRACT.md` (`WeightOld = 0.70`, `WeightNew = 0.30`, `ProficiencyThreshold = 70.0%`, `proficiency-v1`) and `docs/adr/0004-consolidated-evaluation-engine.md`.
- Preserves `docs/adr/0005-openapi-code-first-with-dedicated-dtos.md` (Swagger/OpenAPI sync).

---

## Objective

Guarantee that a zero-knowledge high school learner can complete the entire **Pre-Test → Learning Theory Hub → Adaptive Scenario Loop → Post-Test → Questionnaire** flow within **≤ 120 minutes** (target buffer: **105–115 minutes**):

| Stage | Target Time | Mechanism |
| --- | :---: | --- |
| Live Hook, Registration & Onboarding | `10 min` | Presenter hook + beginner Landing & Pre-Test warm-up (`#72`) |
| Pre-Test (30 questions, `4/7/5/8/6`) | `20 min` | Bilingual Thai+EN (`#71`); 30-min hard auto-submit timer + 20-min pacing hint |
| Learning Theory Hub (`/learn`) | `15–20 min` | **5 Concise Domain Essentials** (~3 min each) satisfy Stage 2 Theory Gate (`5/5`) |
| Adaptive Scenario Loop (`/simulation`) | `30–40 min` | ~1–2 min/Scenario; `MAX_SCENARIOS_CAP = 15` (default), `alpha = 0.30`, `Threshold = 70%` |
| Post-Test (30 questions, `4/7/5/8/6`, Zero-Overlap) | `20 min` | Bilingual Thai+EN (`#71`); 30-min hard auto-submit timer + 20-min pacing hint |
| Post-Test Questionnaire (`/questionnaire`) | `5 min` | CSUQ v3 + Raw NASA-TLX + Personal Visual Summary (`#73`) |
| **Total End-to-End** | **100–115 min** | **Comfortably ≤ 120 min (2 hours)** |

---

## 1. Loop Profile

- **Loop Name**: 2-Hour Compact Learning & Adaptive Pacing Loop
- **Trigger**:
  - Learner enters `/learn` after completing Pre-Test, marks 5/5 Domain Essentials completed to unlock `/simulation`, and runs up to `MAX_SCENARIOS_CAP` Scenarios until all 5 domains pass `>= 70%` or cap is reached.
- **Checkpoint**:
  - **Checkpoint 1 (Content & Tone Gate — Push-Right, `Q13`)**: Author **Domain 1 Essentials (EN + TH)** first and present a compact Brief for beginner tone/length verification before writing Domains 2–5 (`theory/domain{2..5}_*.md` and `theory/th/domain{2..5}_*.md`).
  - **Checkpoint 2 (Pre-Flight Visual & Dry-Run Gate — `Q13`)**: Present `/learn` and `/simulation` EN/TH screenshots + full verification suite results before closing `#70`.

---

## 2. Settled Design Decisions

| Q# | Decision | Settled Rule |
| --- | --- | --- |
| `Q2`, `Q7` | Condensed Learning Content & Gate | Rewrite the 5 Domain Overviews (`theory/domain{1..5}_*.md` & `theory/th/domain{1..5}_*.md`) into beginner-friendly **3-Minute Domain Essentials** (`~350–500` words EN / `~1,800–2,600` chars TH each). Re-enable direct `POST /api/learn/domains/:domainId/complete` for `learner` without mandatory sample-quiz gating. Unlock Stage 2 **Theory Gate** at **`5/5` Domain Essentials** in `user_theory_progress`. Keep all 28 detailed Sub-topics (`1.1`–`5.6`) and their 2 formative sample questions accessible underneath as **Optional Deep Dive**. |
| `Q3`, `Q8` | 2-Hour Budget & `MaxScenarios` | Keep Pre-Test and Post-Test at **30 questions (`4/7/5/8/6`)** with 30-min hard auto-submit timer + visible **"Recommended pace: ~20 min"** pacing pill. Keep `WeightOld = 0.70`, `WeightNew = 0.30`, `ProficiencyThreshold = 70.0%`. Wire existing `.env.example` variable **`MAX_SCENARIOS_CAP`** with new default **`15`** across `docker-compose.yml`, `services/evaluator.go`, and `controllers/simulation_controller.go`; remove dead stale constants `WeightOld = 0.4, WeightNew = 0.6` in `controllers/simulation_controller.go`. |

---

## 3. Concise 5-Domain Essentials Content Contract

### 3.1 File Locations & Validation Rules
All 5 Domain Overview files (EN + TH) remain in their canonical paths so embedded Go fallback (`//go:embed theory`), `scripts/load_theory_to_db.py`, `rag_documents`, and `theory_translations` stay 100% compatible (NFR `Sub>=28, TH>=33` stays green):
- English: `theory/domain1_security_concepts.md` through `theory/domain5_security_management.md`
- Thai: `theory/th/domain1_security_concepts.md` through `theory/th/domain5_security_management.md`

### 3.2 Structure & Length Budget per Domain Essential
Each file MUST remain concise (**1,800–2,800 Thai chars** / **350–550 English words**, ~3 minutes reading time) and retain the canonical 4 top-level headings validated by `services/theory_service.go` and `scripts/load_theory_to_db.py` (relaxing only any excessive minimum-length threshold if present, while still asserting non-empty sections, clickable `https://` citations, and Thai runes in `theory/th/`):
1. `## 1. Core Security+ Theory` — High-school-friendly analogy + crisp bullet/table cheat sheet of the 4–6 highest-yield exam concepts for that domain.
2. `## 2. SOC Telemetry, KQL & Attack Indicators` — 3–5 concrete red-flag log signals + 1 beginner-readable KQL filter example.
3. `## 3. True Positive vs. False Positive Triage` — Compact 2-column table ("Real Threat / True Positive" vs "Safe Routine / False Positive").
4. `## 4. Incident Response Playbook Principles (Containment, Eradication, Recovery)` — 3 crisp bullets (1 per IR phase) + clickable authoritative `https://` references.

---

## 4. Backend Specification (Zero-DDL)

### 4.1 Stage 2 Theory Gate (`services/theory_service.go`, `controllers/learn_controller.go`, `controllers/simulation_controller.go`)
1. **Domain-Level Gate Helper (`services/theory_service.go`)**:
   - Add `IsTheoryDomainGateSatisfied(completedDomains []string) bool`: returns `true` iff all 5 canonical domains (`domain1`..`domain5`) are present.
   - Retain `IsTheorySubtopicGateSatisfied` for reference/analytics compatibility, but switch Stage 2 Gate enforcement in `evaluateStage2TheoryGate` (`controllers/simulation_controller.go`) to check `services.IsTheoryDomainGateSatisfied(completedDomains)`.
   - `getUserTheoryProgressSummary` (`controllers/simulation_controller.go`): computes `theoryCompleted = services.IsTheoryDomainGateSatisfied(completedDomains)`.
2. **Direct Domain Essential Completion (`POST /api/learn/domains/:domainId/complete`)**:
   - Update `LearnController.CompleteDomain` (`controllers/learn_controller.go`):
     - Verify Stage 1 access (completed Pre-Test for `learner`; preview pass-through for `specialist`/`admin` with zero DB writes and `preview_mode: true`).
     - Validate `domainId` against `services.OrderedTheoryDomains` (`404` otherwise).
     - Idempotently call `lc.Store.MarkTheoryCompleted(userID, domainID)` (removing the legacy `403 subtopics_incomplete` blocker).
     - Return `200 OK` with updated `completed_domains`, `completed_count`, `total_domains: 5`, `theory_completed: len(completedDomains) >= 5`, and `preview_mode`.
   - Keep `POST /api/learn/subtopics/:subTopicId/complete` working identically for learners who choose Optional Deep Dive (completing all sub-topics in a domain still idempotently marks the parent domain completed too).
3. **Stage 2 Error Payload**:
   - When `POST /api/simulation/start` rejects an incomplete Learner (`403 Forbidden`), return `code: "THEORY_INCOMPLETE"`, `reason: "theory_incomplete"`, `completed_count` (`0..4`), `total_domains: 5`, and `completed_domains`, guiding the Learner straight to `/learn`.

### 4.2 Configurable `MAX_SCENARIOS_CAP` (Default `15`) & Dead-Constant Cleanup
- **`services/evaluator.go`**:
  - Update `DefaultMaxScenarios = 15` (keep `MaxScenarios = 15` constant for backward compatibility).
  - Add `ResolveMaxScenariosCap() int`: parses `os.Getenv("MAX_SCENARIOS_CAP")`; returns parsed integer if `1 <= v <= 50`, else `15`.
- **`controllers/simulation_controller.go`**:
  - Delete stale unused constants `WeightOld = 0.4` and `WeightNew = 0.6` (`lines 34-35`), delegating uniformly to `services.WeightOld` (`0.70`) and `services.WeightNew` (`0.30`).
  - Replace hardcoded `MaxScenarios` references across `GetSimulationStatus`, `StartSimulation`, and `SubmitSimulation` with `services.ResolveMaxScenariosCap()`.
- **`.env.example` & `docker-compose.yml`**:
  - Update `MAX_SCENARIOS_CAP=15` in `.env.example` and pass `MAX_SCENARIOS_CAP: ${MAX_SCENARIOS_CAP:-15}` to `backend` in `docker-compose.yml`.

---

## 5. Frontend Specification

### 5.1 Learning Theory Hub (`frontend/src/pages/LearnTheoryPage.tsx`)
- **Primary Progress Header**: Display **"Essentials Completed: X / 5 (~3 min each)"** with clear visual checkmarks per domain (`domain1`–`domain5`) and a prominent **"Start Simulation →"** CTA as soon as `theory_completed === true` (`5/5`).
- **Domain Overview Workspace**:
  - Show estimated reading badge (`~3 min read / อ่าน ~3 นาที`).
  - Render a prominent primary action button at the bottom of the Domain Essential: **"Mark Domain Essential Completed"** (`POST /api/learn/domains/:domainId/complete`) with instant checkmark feedback and **"Next Domain →"** button (or **"Start Simulation →"** on the 5th domain).
  - Provide an optional secondary CTA: **"Optional Deep Dive: Explore X Sub-topics & Practice Questions"** expanding the nested sub-topic list cleanly without blocking progression.
- **Sidebar Tree**:
  - Keep Domain 1–5 top-level entries clean and scannable with Essential completion status badges; collapse each domain's child Sub-topics by default under an expandable **"Deep Dive (X)"** toggle so novices aren't overwhelmed by 28 links on arrival.
  - Deep-links `/learn?domain=domain1&topic=1.2` continue working cleanly (auto-expanding that domain's Deep Dive group when navigated from `ExamReviewPage`).

### 5.2 Exam Pacing Hint (`frontend/src/pages/ExamPage.tsx`)
- Preserve the existing 30-minute countdown timer (`EXAM_DURATION_SECONDS = 30 * 60`) and single-shot timeout auto-finalize.
- Add a calm helper pill beside the timer: **`Target pace: ~20 min` / `แนะนำใช้ ~20 นาที`** (`text-xs font-mono`, compliant with Typography Guard).

### 5.3 Simulation & Evaluation Status (`frontend/src/pages/EvaluationPage.tsx`, `SimulationPage.tsx`, `SimulationCompletePage.tsx`)
- Update Theory Gate progress indicators from `X / 28 Sub-topics` to **`X / 5 Domain Essentials`** (`completed_count / total_domains`), and display dynamic `max_scenarios` from `GET /api/simulation/status` (`X / 15`) everywhere rather than hardcoded `10`.

---

## 6. Verification & Acceptance Criteria

- [ ] Checkpoint 1 approved: Domain 1 Essentials (EN+TH) reviewed for beginner clarity and ~3-min length before updating Domains 2–5.
- [ ] `python scripts/load_theory_to_db.py --validate-only` passes across all 5 Domain Essentials + 28 Sub-topics + 33 Thai translations.
- [ ] `go test ./...` passes, including updated `controllers/learn_controller_test.go` (direct `CompleteDomain` succeeds for Learner without sub-topics, Staff remains `preview_mode`), `controllers/simulation_theory_gate_test.go` (4/5 blocked, 5/5 unlocked), and `services/evaluator_test.go` (`ResolveMaxScenariosCap` default 15 + env override).
- [ ] `cd frontend && npm test && npm run tsc && npm run build` passes (including `i18n.test.ts` and `typography.test.ts`).
- [ ] `scripts/test_integration.py` updated to verify 5/5 Domain Essentials unlock Simulation + `MAX_SCENARIOS_CAP` reporting, passing green against Docker stack.
- [ ] Swagger docs (`docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go`) regenerated cleanly.
