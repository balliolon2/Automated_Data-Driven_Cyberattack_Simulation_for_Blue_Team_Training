# Workflow: Adaptive Scenario Playbook Synthesis & Anti-Predictability Loop (Issue #1)

## Objective
Eliminate predictable response option patterns in Blue Team simulation scenarios (GitHub Issue #1) where learners could achieve 100% accuracy without reading telemetry because:
1. Every AI/fallback scenario used the exact same 3 hardcoded, absurdly short distractors (`"Power off all domain controllers immediately"`, `"Reformat all network switches in the subnet"`, `"Disable MFA enforcement for affected users"`).
2. Correct options were always placed at index `0` (`c1`, `e1`, `r1`) and were noticeably longer/more specific than distractors.
3. Several static scenarios in `scenarios_seed.json` had only 1–2 options per phase or lacked distractors entirely.

This workflow specifies the end-to-end loop for synthesizing, validating, remediating, and deterministically shuffling realistic, context-aware playbook response options across all scenario sources (LLM-generated, Fallback templates, Static seed pool, and existing Database records).

---

## 1. Loop Profile

- **Loop Name**: Scenario Playbook Generation, Remediation & Presentation Loop
- **Frequency / Cadence**: Per simulation session start (`POST /api/simulation/start`), session recovery (`GET /api/simulation/session`), and system startup/seed synchronization.
- **Primary Actors**:
  - **AI Scenario Synthesizer (`services/llm_service.go`)**: Generates context-specific correct actions and plausible technical distractors.
  - **Scenario Validator & Auto-Remediator (`services/llm_service.go` / `main.go`)**: Enforces option density, repairs legacy/incomplete playbook phases, and bans boilerplate strings.
  - **Simulation Controller (`controllers/simulation_controller.go`)**: Sanitizes answer keys and deterministically shuffles options per `session_id`.
  - **Learner**: Evaluates realistic incident response trade-offs during Phase 2 (Containment), Phase 3 (Eradication), and Phase 4 (Recovery).

---

## 2. Triggers

1. **Dynamic Scenario Request (`POST /api/simulation/start`)**: Learner requests a new adaptive training scenario targeting their weakest CompTIA Security+ domain.
2. **Session Hydration (`GET /api/simulation/session`)**: Learner refreshes or resumes an active `in_progress` simulation session.
3. **Database Seed / Startup Migration**: Backend boots or runs scenario seeding (`scripts/generate_scenarios.py` -> `scenarios_seed.json`), scanning existing `scenarios` rows for legacy 2-option or boilerplate patterns.

---

## 3. Structural & Schema Specifications

### 3.1 Playbook Option Contract (Per Phase)
Every scenario's `playbook_steps` JSONB object must contain the three keys `containment`, `eradication`, and `recovery`. Each phase array MUST satisfy:
- **Total Options**: `3` to `4` options per phase.
- **Correct Options**: `1` to `2` options with `"is_correct": true` and `"points": 5`.
- **Distractor Options**: At least `2` options with `"is_correct": false` and `"points": -3` (or `-2`).
- **Contextual & Length Parity**:
  - Both correct and distractor labels must be 12–35 words long, written in professional SOC operational language, and reference concrete scenario entities (`target_host`, `attacker_ip`, affected user/service, or relevant security controls).
  - Distractors must represent realistic SOC missteps (e.g., premature action that destroys volatile forensic memory when forensics is required, misconfigured firewall scope such as blocking internal gateway/DNS instead of attacker IP, applying a control from the wrong IR phase, or restarting a service without removing persistence).
- **Banned Boilerplate Labels**:
  - `"Power off all domain controllers immediately"`
  - `"Reformat all network switches in the subnet"`
  - `"Disable MFA enforcement for affected users"`

### 3.2 Expanded `aiMicroScenario` Schema (`services/llm_service.go`)
Update `aiMicroScenario` to accept structured phase options or explicit contextual distractors from the LLM:

```go
type aiPlaybookOption struct {
	Label       string `json:"label"`
	IsCorrect   bool   `json:"is_correct"`
	Explanation string `json:"explanation"`
}

type aiMicroScenario struct {
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	IsTruePositive    bool               `json:"is_true_positive"`
	TpFpExplanation   string             `json:"tp_fp_explanation"`
	AttackerIP        string             `json:"attacker_ip"`
	TargetHost        string             `json:"target_host"`
	AttackLogMessage  string             `json:"attack_log_message"`
	KeyFinding        string             `json:"key_finding"`
	ContainmentAction string             `json:"containment_action,omitempty"`
	EradicationAction string             `json:"eradication_action,omitempty"`
	RecoveryAction    string             `json:"recovery_action,omitempty"`
	ContainmentSteps  []aiPlaybookOption `json:"containment_steps,omitempty"`
	EradicationSteps  []aiPlaybookOption `json:"eradication_steps,omitempty"`
	RecoverySteps     []aiPlaybookOption `json:"recovery_steps,omitempty"`
}
```

### 3.3 LLM Prompt Specification (`GenerateLLMScenario`)
Update `userPrompt` in `services/llm_service.go` to instruct the model:
- Generate `containment_steps`, `eradication_steps`, and `recovery_steps`, each containing **3 options** (1 correct action with `is_correct: true` and 2 plausible technical distractors with `is_correct: false`).
- Enforce **strict style & length parity**: Every option (`is_correct: true` and `is_correct: false`) must be a detailed, single-sentence SOC operational procedure (15–30 words) referencing specific entities from the scenario (`target_host`, `attacker_ip`, affected accounts, or protocols).
- Never use generic or absurd distractors; distractors must test genuine CompTIA Security+ 701 analytical judgment (e.g., isolating the wrong network segment, flushing volatile evidence before capture, weakening authentication controls, or restoring from an unverified snapshot prior to patching).
- If `is_true_positive` is `false` (False Positive scenario), `is_correct: true` options must focus on verifying authorization/change tickets, tuning SIEM correlation thresholds to suppress benign baseline noise, and documenting the benign finding without disrupting production hosts, while `is_correct: false` distractors represent unnecessary host quarantine, credential revocation, or emergency firewall blocks.

---

## 4. Pipeline Execution

### Step 1: Contextual Backfill & Fallback Synthesis (`services/llm_service.go`)
1. When `synthesizeFullScenario(ai, weakestDomain, domainName)` runs:
   - For each phase (`containment`, `eradication`, `recovery`), inspect `ai.ContainmentSteps`, `ai.EradicationSteps`, and `ai.RecoverySteps`.
   - If the LLM returned legacy single-string fields (`ContainmentAction`, etc.) or fewer than 3 options in any phase, invoke `buildPhaseStepsWithDistractors(phase, weakestDomain, ai)` to populate 1 correct option (`+5` pts) and 2 domain-specific, entity-interpolated distractors (`-3` pts each).
2. **Contextual Distractor Generator (`buildContextualDistractors`)**:
   - Uses parameterized templates per domain (`domain1`..`domain5`) and phase (`containment`, `eradication`, `recovery`), interpolating `ai.TargetHost` and `ai.AttackerIP`.
   - Example interpolated distractors for `containment` on `TargetHost = "SRV-WEB01"`, `AttackerIP = "185.220.101.45"`:
     - Distractor A: *"Flush all volatile RAM caches and system event logs on SRV-WEB01 immediately to clear active session handles from 185.220.101.45 before isolating the interface."* (`is_correct: false`, `-3` pts — destroys forensic evidence).
     - Distractor B: *"Reroute inbound traffic from 185.220.101.45 to the internal corporate VLAN standby interface on SRV-WEB01 while keeping existing administrative sessions active."* (`is_correct: false`, `-3` pts — exposes internal VLAN to lateral movement).
   - Assigns sequential IDs (`c1`, `c2`, `c3`, `e1`, `e2`, `e3`, `r1`, `r2`, `r3`).

### Step 2: Quality & Anti-Predictability Validation Gate (`ValidateScenarioCandidate`)
Extend `ValidateScenarioCandidate(s *models.Scenario) error` in `services/llm_service.go`:
1. Verify `playbook_steps` contains `containment`, `eradication`, and `recovery`.
2. For each phase:
   - Must contain at least **3 options**.
   - Must contain at least **1 option** with `is_correct == true`.
   - Must contain at least **2 options** with `is_correct == false`.
   - No duplicate `label` strings within the scenario.
   - No option `label` may match the banned boilerplate list (`"Power off all domain controllers immediately"`, `"Reformat all network switches in the subnet"`, `"Disable MFA enforcement for affected users"`).

### Step 3: Static Seed & Database Auto-Remediation
1. **Static Seed (`scripts/generate_scenarios.py` & `scenarios_seed.json`)**:
   - Update all 15 static scenarios across Domains 1–5 (both True Positive and False Positive scenarios) so every phase (`containment`, `eradication`, `recovery`) has 3–4 realistic, domain-aligned options with at least 2 plausible distractors matching the length and technical tone of the correct options.
   - Regenerate `scenarios_seed.json`.
2. **Database Auto-Remediation (`EnsureScenarioPlaybookQuality(db *gorm.DB)`)**:
   - Executed at backend startup (and lazily if a scenario is fetched in `StartScenario` / `GetActiveSession`).
   - Queries all rows in `scenarios`. If any row has a phase with `< 3` options, `< 1` correct option, `< 2` incorrect options, or contains any banned boilerplate label, it automatically upgrades `playbook_steps` using the updated `scenarios_seed.json` definition (if title matches) or `UpgradeScenarioPlaybookSteps(&scenario)` (interpolating the scenario's host/IP clues) and saves the updated `playbook_steps` to PostgreSQL.

### Step 4: Deterministic Session-Scoped Option Shuffling (`controllers/simulation_controller.go`)
1. Update `sanitizeScenario(s models.Scenario, sessionID string) ScenarioResponse` and `sanitizePlaybookSteps(steps any, sessionID string) any`.
2. For each phase (`containment`, `eradication`, `recovery`):
   - Extract sanitized option items (`{"id": id, "label": label}`).
   - Compute a deterministic 64-bit FNV-1a / SHA-256 seed from `fmt.Sprintf("%s:%s", sessionID, phase)`.
   - Shuffle `sanitizedActions` in-place using a deterministic Fisher-Yates shuffle driven by that seed.
   - If `sessionID` is empty (fallback), seed with `fmt.Sprintf("%s:%s", s.ScenarioID, phase)`.
3. **Result**:
   - Options appear in a randomized order for every session (correct options are distributed across positions 1, 2, 3, 4 instead of always position 1).
   - Refreshing the browser (`GET /api/simulation/session`) produces the exact same option order for that session.
   - Scoring in `services.ScoreSubmission` matches by option `id` (`selectedSet[aID]`), remaining 100% invariant to display order.

---

## 5. Verification Criteria (Definition of Done)

An implementation of this workflow is complete when:
1. **Unit & Evaluator Tests (`go test ./...`)**:
   - `ValidateScenarioCandidate` rejects scenarios with `< 3` options per phase, missing distractors, or banned boilerplate labels.
   - `GenerateFallbackScenario` for all 5 domains passes `ValidateScenarioCandidate` and produces 3+ contextually distinct options per phase with zero banned boilerplate labels.
   - `sanitizePlaybookSteps` with a given `sessionID` produces a deterministic order across repeated calls and varies order across different `sessionID` values.
2. **Static Seed Verification**:
   - Every scenario in `scenarios_seed.json` has `>= 3` options in `containment`, `eradication`, and `recovery`, with at least 1 correct and 2 incorrect options per phase, and comparable label lengths.
3. **Database Remediation Verification**:
   - Existing scenarios in PostgreSQL containing `"Power off all domain controllers immediately"` are automatically upgraded on startup/load so no learner ever encounters the legacy boilerplate options again.
