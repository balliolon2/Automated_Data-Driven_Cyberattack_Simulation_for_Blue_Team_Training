# Workflow Spec: 4-Beat Incident Storytelling & Reader-Friendly Scenario Narrative

**Status:** `/loop-me` Grilling Rounds 1–3 Settled (`Q1`–`Q13` Resolved, 2026-11-23) — Ready for User Review Before Implementation  
**Depends On:**
- [scenario_telemetry_and_style_diversity.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/scenario_telemetry_and_style_diversity.md) (5 Briefing Styles & dynamic SIEM shortcut context)
- [language_switching_and_localization.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/language_switching_and_localization.md) (Hybrid Bilingual EN/TH & `translations` JSONB contract)

---

## Objective

Eliminate abrupt, jargon-heavy incident descriptions in Blue Team simulations where learners felt dropped into an unexplained technical wall without context. Every scenario's **Incident Narrative (`description` and `translations.th.description`)** MUST follow a smooth **4-beat narrative arc** written as friendly, reader-accessible prose without visual beat badges:

1. **Who you are (`role_context`)**: Your role as an organizational **SOC Analyst** and the workplace/sector context.
2. **What you are doing right now (`current_activity`)**: Your active duty or routine operational task when the alert arises.
3. **What happened / What was detected (`what_happened`)**: What anomalous activity was observed, described in approachable human language with brief plain-language explanations for specialized infrastructure while preserving key Host/User/IP entities.
4. **What could happen next (`potential_impact`)**: **Balanced Dual-Stakes** consequence framing explaining both what harm could occur if the alert is a real intrusion *and* what business disruption could happen if legitimate work is wrongly blocked — preserving full True Positive / False Positive suspense.

The two loops this workflow makes real:
1. **Scenario Authoring & Synthesis Loop**: Every newly generated AI scenario, all 15 static seed scenarios, and all 5 deterministic fallback templates emit structured 4-beat `story` metadata (EN + TH), pass validation gates, and deterministically compose smooth 3-paragraph narratives into `description` and `translations.th.description`.
2. **Legacy Local Upgrade & Display Loop**: Loading any existing local simulation session (in-progress or completed result) whose underlying scenario or snapshot lacks structured `story` beats automatically triggers a one-time bilingual rewrite (`UpgradeScenarioStoryBilingual`), overlays updated narrative text onto legacy snapshots without altering `initial_logs`, `playbook_steps`, or scoring, and renders comfortable, badge-free paragraphs in the UI.

---

## 1. Loop Profile

- **Loop Name**: 4-Beat Incident Storytelling Narrative Loop
- **Frequency / Cadence**:
  - Runtime generation: every `POST /api/simulation/start` (AI dynamic or deterministic fallback).
  - Runtime recovery & review: every `GET /api/simulation/session` and `GET /api/simulation/result/:sessionId` when a legacy local scenario lacks `translations.{en,th}.story`.
  - Authoring & CI: every seed regeneration (`scripts/generate_scenarios.py`, `scripts/scenario_translations_th.json`, `scenarios_seed.json`) and validation run (`go test ./...`, `scripts/validate_evidence_mapping.py`).
- **Primary Actors**:
  - **Prompt & Story Composer (`services/llm_service.go`)**: Enforces 4-beat `story` fields on `aiMicroScenario`, `GenerateFallbackScenario`, and `UpgradeScenarioStoryBilingual`, and composes clean 3-paragraph narratives.
  - **Lazy Legacy Story Upgrader (`controllers/simulation_controller.go`)**: Upgrades legacy scenarios/snapshots on read (`GetActiveSession` / `GetResult`) in both `en` and `th` while preserving `initial_logs`, `playbook_steps`, `expected_outcomes`, and `status`.
  - **Static Seed Authoring (`scripts/generate_scenarios.py`, `scripts/scenario_translations_th.json`)**: Authors 4-beat EN + TH stories across all 15 static scenarios without invalidating `scripts/evidence_mapping.json` hashes.
  - **Narrative Renderer (`frontend/src/pages/SimulationPage.tsx`, `frontend/src/pages/ScenarioResultPage.tsx`)**: Displays comfortably spaced multi-paragraph prose with zero numbered tags or beat badges.
- **Trigger**: Session start (`POST /api/simulation/start`), session load/language toggle (`GET /api/simulation/session`, `GET /api/simulation/result/:sessionId`), or seed build/test run.
- **Checkpoint**: None at runtime (100% autonomous). Spec review checkpoint completed via `/loop-me` prior to implementation.

---

## 2. Settled Narrative Design Rules (`Q1`–`Q13`)

### 2.1 Persona, Voice, and Readability Policy
1. **Single Anchored Role with Varied Duty (`Q1=A`)**: Every scenario places the learner as an organizational **SOC Analyst** (`นักวิเคราะห์ SOC`). Organization flavor (e.g., manufacturing plant, SaaS fintech, enterprise corporate network, logistics portal) and current duty shift across scenarios instead of changing job titles randomly.
2. **Second-Person Immersive Voice (`Q6=A`)**: Write directly to `"You"` (`"คุณ"`) in present/immediate narrative voice so the learner steps smoothly into the situation.
3. **Friendly Guided Prose + Entity Preservation (`Q7=A`)**:
   - Translate cold acronyms and deep domain concepts into natural functional descriptions on first mention (e.g., *"plant production data server (`DMZ-HIST01`)"* / *"เซิร์ฟเวอร์เก็บข้อมูลการผลิต (`DMZ-HIST01`)"* rather than *"DMZ Historian Gateway"*; *"equipment control commands targeting factory PLC (`PLC-MXR01`)"* / *"คำสั่งควบคุมเครื่องจักร (`PLC-MXR01`)"* rather than raw *"Modbus writes"* or *"tag-write history"*).
   - Remove awkward outgoing commands telling the learner what tool trace to run (e.g., stop saying *"pull the EDR process tree..."* inside the narrative).
   - **Canonical Entity Invariant**: Preserve **2–3 primary entities** (target hostname, actor username, and/or actor IP) verbatim in English prose in both `description` and `translations.th.description` so `frontend/src/lib/siemShortcuts.ts` (`buildInvestigationShortcuts`) extracts `Host:`, `User:`, `IP:` filter chips identically across EN/TH.
4. **Balanced Dual-Stakes Anti-Spoiler Rule (`Q3=A`)**:
   - Beat 4 (`potential_impact`) MUST describe *conditional stakes on both sides*: what security/operational damage could occur if the activity is genuinely malicious, paired with what workflow/production disruption could occur if legitimate authorized work is wrongly blocked.
   - True Positive and False Positive scenarios MUST read with identical urgency and never hint at the final TP/FP verdict in the narrative.

### 2.2 3-Paragraph Flow & Badge-Free Rendering (`Q2=C`, `Q8=A`, `Q13=A`)
Compose the 4 structured beats deterministically into **3 short paragraphs** separated by `\n\n`:

- **Paragraph 1 (Scene Setting)**: `role_context` + `" "` + `current_activity` (1–2 friendly sentences introducing who you are, shift/company context, and what you were working on).
- **Paragraph 2 (The Discovery)**: `what_happened` (2 clear sentences describing what triggered your attention and which primary systems/accounts/IPs are involved in plain terms).
- **Paragraph 3 (Stakes / Dual-Impact)**: `potential_impact` (1–2 sentences explaining what could go wrong either way and why you need to verify the logs first).

**UI Display Constraint**:
- Render the paragraphs smoothly (`space-y-3` on readable body font styling inside `Incident Narrative`) with **NO beat labels, step numbers, pill badges, or collapsible sections**.

### 2.3 Canonical Thai Tone Reference (`Q13=A`)
Approved example rewriting legacy scenario `70dc1a44-0cf3-4a46-a1fb-5911d2dad866`:

> ช่วงเปลี่ยนกะเช้าวันจันทร์ คุณปฏิบัติหน้าที่เป็น **SOC Analyst** ประจำโรงงานผลิตแห่งหนึ่ง และกำลังตรวจสรุปรับช่วงต่อจากทีมกะดึกร่วมกับแผงแจ้งเตือนความปลอดภัยประจำวัน
>
> ระบบตรวจพบความผิดปกติบนเซิร์ฟเวอร์เก็บข้อมูลการผลิต (`DMZ-HIST01`) ซึ่งปกติทำหน้าที่เพียงพักข้อมูลไว้ในโซนกลาง แต่จู่ๆ มีการล็อกอินเข้าใช้งานบัญชี `svc_scada_sync` จาก IP ภายนอก `45.134.26.88` แล้วเชื่อมต่อตรงเข้าไปสั่งงานเครื่องจักร (`PLC-MXR01`) ในโซนควบคุมโรงงานโดยไม่ผ่านจุดเชื่อมต่อที่กำหนดไว้ (`JUMP-OT01`)
>
> หากเป็นการลักลอบเจาะระบบจริง ผู้โจมตีอาจส่งคำสั่งปลอมไปรบกวนสายการผลิตหรือทำให้เครื่องจักรทำงานผิดพลาดได้ แต่หากเป็นงานซ่อมบำรุงหรือซิงก์ข้อมูลที่ได้รับอนุมัติล่วงหน้า การรีบตัดการเชื่อมต่อเครื่องนี้อาจทำให้การบันทึกข้อมูลและกระบวนการผลิตหยุดชะงักได้ คุณจึงต้องตรวจสอบ Log ก่อนตัดสินใจ

### 2.4 Mapping All 5 `BriefingStyles` to Natural Story Openers (`Q9=A`)
All 5 rotated `BriefingStyles` remain active and differentiate *how the scene begins* in Beat 1–2 while preserving the 4-beat flow:

| BriefingStyle | Beat 1 & Beat 2 Opening Scene |
| --- | --- |
| `siem_escalation_ticket` | You are on regular SOC monitoring shift working through the morning/night SIEM alert queue when a high-priority ticket flags unusual system behavior. |
| `shift_handover_brief` | You just started your shift and are reading the handover summary left by the outgoing analyst when an unresolved anomaly catches your eye. |
| `threat_hunting_lead` | You are running a scheduled weekly baseline health review across endpoints/network telemetry when a recurring pattern stands out from normal traffic. |
| `helpdesk_user_report` | You are assisting IT Support review an escalated ticket from an employee or system owner reporting unexpected behavior, and initial logs confirm strange activity. |
| `cloud_guardrail_alert` | You are monitoring daily cloud deployment and compliance dashboards alongside Dev/Infra operations when an automated guardrail flags a risky configuration or transfer. |

---

## 3. Data Schema & JSONB Contract (`Q10=A`, `Q11=A`)

### 3.1 Zero-DDL Storage inside Existing `scenarios.translations` JSONB
To preserve full compatibility with existing Docker PostgreSQL volumes (`scenarios` table & `scenario_snapshots` JSONB `scenario_data`) and respect `AGENTS.md` schema rules (`init.sql` untouched):

1. Define canonical `story` payload structure (identical keys in `en` and `th`):
   ```json
   {
     "role_context": "...",
     "current_activity": "...",
     "what_happened": "...",
     "potential_impact": "..."
   }
   ```
2. Store under `scenarios.translations`:
   ```json
   {
     "en": {
       "description": "<Paragraph 1>\n\n<Paragraph 2>\n\n<Paragraph 3>",
       "story": {
         "role_context": "...",
         "current_activity": "...",
         "what_happened": "...",
         "potential_impact": "..."
       }
     },
     "th": {
       "title": "...",
       "description": "<Thai Paragraph 1>\n\n<Thai Paragraph 2>\n\n<Thai Paragraph 3>",
       "story": {
         "role_context": "...",
         "current_activity": "...",
         "what_happened": "...",
         "potential_impact": "..."
       },
       "tp_fp_explanation": "...",
       "playbook_steps": { "...": { "label": "...", "explanation": "..." } },
       "findings": { "...": { "description": "...", "explanation": "..." } }
     }
   }
   ```
3. Top-level `scenario.Description` ALWAYS stores the composed 3-paragraph English string (`translations.en.description`) so:
   - `buildInvestigationShortcuts` (`SimulationPage.tsx`) keeps scanning `scenario.title` + `scenario.description` cleanly.
   - `sanitizeScenarioTranslations` continues exposing only safe display fields (`title`, `description`, `playbook_steps[id].label` per locale) without exposing answer keys or inflating client payloads.

### 3.2 Deterministic Composition Helper (`ComposeStoryDescription`)
Backend helper `ComposeStoryDescription(story scenarioStory) string` (and matching Python helper in `scripts/generate_scenarios.py`):
- Joins `strings.TrimSpace(story.RoleContext)` + `" "` + `strings.TrimSpace(story.CurrentActivity)` as Paragraph 1, then appends `"\n\n"`, Paragraph 2 (`WhatHappened`), `"\n\n"`, Paragraph 3 (`PotentialImpact`).
- Used consistently by AI generation (`synthesizeFullScenarioWithErrors`), `GenerateFallbackScenario`, `UpgradeScenarioStoryBilingual`, and `scripts/generate_scenarios.py`.

---

## 4. Source-by-Source Upgrade Plan (`Q4=B`, `Q10=A`, `Q12=A`)

### 4.1 AI Scenario Synthesis (`services/llm_service.go`)
1. Expand `aiMicroScenario` with optional structured bilingual story objects (`Story *aiStoryBeats` and `StoryTh *aiStoryBeats`) while keeping `Description`/`DescriptionTh` fallback compatibility for legacy tests.
2. Update `systemPrompt`, `briefingStyleInstructions`, and `userPrompt` JSON template in `GenerateLLMScenarioWithContext` to instruct the LLM to author all 4 beats (`story` + `story_th`) following Second-Person Voice, Friendly Guided Prose, and Balanced Dual-Stakes rules while keeping primary `target_host`, `attacker_ip`, and `primary_user` in English text.
3. In `synthesizeFullScenarioWithErrors`: if `ai.Story` has all 4 non-empty beats, compose `scenario.Description = ComposeStoryDescription(ai.Story)` and store `translations.en.story`; if `ai.StoryTh` has all 4 non-empty beats containing Thai text, compose `th.description = ComposeStoryDescription(ai.StoryTh)` and store `translations.th.story`.

### 4.2 Deterministic Fallback Templates (`GenerateFallbackScenario` in `services/llm_service.go`)
Rewrite all 5 domain fallback templates (`domain1`–`domain5`) to populate complete 4-beat `Story` (EN) + `StoryTh` (TH) aligned with their respective `BriefingStyle` and dual-stakes framing, composing crisp 3-paragraph `Description` and `DescriptionTh`.

### 4.3 15 Static Seed Scenarios (`scripts/generate_scenarios.py`, `scripts/scenario_translations_th.json`, `scenarios_seed.json`)
1. Update every static scenario across `domain1`–`domain5` (keeping exact `title`, `initial_logs`, `findings`, `evidence_log_indices`, `verdict_support`, and `playbook_steps`) with 4-beat EN + TH `story` objects and composed 3-paragraph `description` / `translations.th.description`.
2. Regenerate `scenarios_seed.json` cleanly via `scripts/generate_scenarios.py`.
3. Verify `scripts/validate_evidence_mapping.py --mode seed` passes with **zero** SHA-256 hash mismatches against `scripts/evidence_mapping.json`.
4. Ensure backend startup seeding / remediation refreshes `description` & `translations` on existing static seed DB rows when `translations.en.story` or `translations.th.story` is missing.

### 4.4 Lazy Legacy Local Story Upgrade (`services/llm_service.go`, `controllers/simulation_controller.go`) (`Q12=A`)
1. Add `HasCompleteStory(s *models.Scenario) bool` verifying both `translations.en.story` and `translations.th.story` contain all 4 non-empty beats (and `translations.th.story` contains Thai runes).
2. Add `UpgradeScenarioStoryBilingual(ctx context.Context, scenario *models.Scenario) (string, map[string]any, error)` using shared `resolveLLMClientConfig()`:
   - Sends canonical scenario context (`title`, `domain_id`, `description`, `target_host`/`primary_user`/`attacker_ip` extracted via `ExtractPrimaryScenarioEntities`, plus existing `playbook_steps` and `findings` only when `translations.th` also lacks full step/finding translations).
   - Outputs validated EN `story` + composed `description`, TH `story` + composed `translations.th.description`, plus any missing TH `title`/`tp_fp_explanation`/`playbook_steps`/`findings` in a single LLM call (keeping `TranslateScenarioToThai` unified cleanly rather than duplicating LLM translation plumbing!).
   - Validates via `ValidateScenarioTranslationsTh` and `ValidateStoryBeats` before persisting.
3. Update `controllers/simulation_controller.go`:
   - In `GetActiveSession` and `GetResult`: call `sc.ensureScenarioStoryAndTranslation(reqCtx, &scenario)` (replacing single-locale-only backfill check with complete EN+TH story check `!services.HasCompleteStory(&scenario)` — regardless of whether current request is `en` or `th`, gated by `LLM_API_KEY` fail-open and per-scenario lock `lockScenarioTranslation` with 120s bounded timeout).
   - Persist ONLY `description` and `translations` on `scenarios` row (NEVER mutate `initial_logs`, `playbook_steps`, `expected_outcomes`, `is_true_positive`, or `status`).
   - Update `sessionScenarioFromDB`: when an immutable `scenario_snapshots` record lacks complete `story` metadata (`!services.HasCompleteStory(&scenario)`), supplement BOTH `description` and `translations` from the `live` `scenarios` row if `HasCompleteStory(&live)` is true (while keeping snapshot `initial_logs`, `playbook_steps`, and `expected_outcomes` immutable!).

---

## 5. Verification Plan

1. **Go Unit Tests (`services/llm_service_test.go`, `controllers/simulation_controller_test.go`)**:
   - `ComposeStoryDescription` 3-paragraph assembly and `HasCompleteStory` checks.
   - `GenerateFallbackScenario` all 5 domains produce complete EN + TH 4-beat stories containing primary entities and Balanced Dual-Stakes phrasing, passing `ValidateScenarioCandidate`.
   - `UpgradeScenarioStoryBilingual` parses mock LLM story JSON cleanly, preserves `InitialLogs` 100% English, fails open when `LLM_API_KEY` is unset, and `ensureScenarioStoryAndTranslation` coalesces concurrent callers + overlays upgraded story onto snapshot sessions.
2. **Python Seed & Evidence Validation**:
   - `python scripts/validate_evidence_mapping.py --mode seed` passes cleanly with intact `evidence_mapping.json` hashes.
   - Verify all 15 static scenarios contain 4-beat `translations.en.story` and `translations.th.story`, 3-paragraph `description` / `translations.th.description`, and primary entities retained cleanly for SIEM shortcuts.
3. **Frontend Verification (`frontend/`)**:
   - `npm --prefix frontend test -- --run` (add/verify `siemShortcuts.test.ts` on 3-paragraph rewritten EN narratives, confirming `Host:`, `User:`, `IP:` shortcuts resolve identically).
   - `npm --prefix frontend run build` and `eslint` cleanly pass with 0 new warnings.
4. **Live Stack Verification**:
   - Rebuild Docker Compose (`backend`, `frontend`, `seeder` if seed updated) and verify against live session/API (`GET /api/simulation/session` & `GET /api/simulation/result/:sessionId` in both EN and TH + `python scripts/test_integration.py`).
