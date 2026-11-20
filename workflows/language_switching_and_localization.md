# Workflow Spec: Bilingual Language Switching & Localization (Issue #39)

**GitHub Issue:** [#39 — [Feature] Need language switching](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/39)
**Status:** Interview Settled (`/loop-me` Rounds 1–3 Complete, Q1–Q19 Resolved) — Ready for Implementation

---

## Objective

Enable seamless **English (`en`, default) ↔ Thai (`th`)** language switching across the SOC Trainer platform while strictly preserving **English-only realism** for CompTIA Security+ exam items and SIEM simulation telemetry logs:

- **Localized (`en` / `th`)**: All frontend UI chrome (16 pages & shared components), Learning Theory Hub curriculum (5 Domain Overviews + 28 SY0-701 Sub-topics = 33 Markdown documents), Simulation Scenario non-log fields (Title, Incident Narrative/Briefing, Playbook Response Action options, Key Findings descriptions, TP/FP & feedback explanations — across both the 15 static seeds and runtime AI scenarios), Backend API error messages, and In-App Notifications.
- **Strictly English-Only (Zero Translation Invariant)**: Pre-Test & Post-Test exams (`question_text`, `options`, `correct_answer`, `explanation`), Formative Sample Questions inside `/learn`, Simulation `initial_logs` (all log fields & messages), KQL query syntax/operators, IOCs (hostnames, IPs, usernames, log sources), and MITRE ATT&CK IDs.
- **Author-Language Preserved (No Auto-Translation)**: Community Analysis Threads, Comments, and Admin free-text rejection reasons stay in whatever language the author typed.

The two loops this workflow makes real:
1. **Runtime Localization Loop**: Guest or authenticated user selects `EN | TH` → instant zero-reload UI & content switch preserving all in-progress state → preference persists to `localStorage` and `users.preferred_language` → exams and SIEM telemetry remain strictly English with a subtle explanatory notice in Thai mode.
2. **Content Translation & Verification Maintenance Loop**: Author/update UI dictionaries, 33 Thai theory docs, 15 static scenario translations, or bilingual AI prompts → deterministic CI gates enforce 100% key/`{{param}}` parity, 100% scenario/theory translation completeness, and zero Thai Unicode leakage into exams or `initial_logs`.

---

## 1. Loop Profile

- **Loop Name**: Bilingual Language Switching & Localization Loop
- **Frequency / Cadence**:
  - Runtime: every user language toggle (`EN | TH`), Login/Register, `/learn` navigation, `/simulation` start/resume/result, Specialist Review, and Notification fetch.
  - Authoring & CI: initial pilot + full translation of static Theory (33 files) and Static Scenarios (15 scenarios), plus every subsequent PR adding UI copy, theory content, or scenario prompts.
- **Primary Actors**:
  - **Frontend i18n Engine (`frontend/src/i18n/`)**: Zero-dependency typed TypeScript dictionaries (`en.ts`, `th.ts`), `I18nProvider`, `useI18n()`, date formatter, and bilingual scenario selector.
  - **Preference & Contract Layer (`controllers/auth.go`, `controllers/notification_controller.go`, all controllers)**: Persists `users.preferred_language`, emits additive `code`/`params` on errors and `template_key`/`template_params` on notifications.
  - **Theory Localization Pipeline (`theory/th/`, `scripts/load_theory_to_db.py`, `services/theory_service.go`, `controllers/learn_controller.go`)**: Validates, embeds, seeds, and serves localized theory Markdown via `theory_translations`.
  - **Bilingual Scenario Pipeline (`scripts/scenario_translations_th.json`, `scripts/generate_scenarios.py`, `services/llm_service.go`, `services/scenario_evidence.go`, `controllers/simulation_controller.go`, `controllers/specialist_controller.go`)**: Generates, validates, sanitizes, and snapshots bilingual scenarios (`translations.th`) with strict English-only `initial_logs`.
- **Trigger**:
  - Runtime: event-driven (user clicks `EN | TH` toggle or Profile selector, logs in/registers, or loads `/learn` / `/simulation`).
  - Authoring/CI: event-driven (content authoring run or PR to `main`).
- **Checkpoint**:
  - **Runtime**: None (100% autonomous).
  - **Initial Static Content Authoring**: **1 push-right Human Review Checkpoint** after authoring the Pilot Batch (**Domain 1 Theory** [1 Overview + 4 Sub-topics = 5 files] + **3 sample Static Scenarios** from `scripts/scenario_translations_th.json`). Present a compact **Brief** (terminology glossary, excerpt of 1 Domain Overview, 1 Sub-topic, 1 True Positive & 1 False Positive scenario, and file links) for tone/terminology approval before scaling to Domains 2–5 (28 files) and the remaining 12 Static Scenarios.

---

## 2. Supported Locales, Content Boundary Matrix & Terminology Policy

### 2.1 Locales & Defaults
- Supported locale codes: `en` (canonical default), `th`.
- Schema and code structures use locale codes (`en`, `th`) rather than boolean flags so future languages can be added without refactoring.
- URL paths remain unchanged (`/learn`, `/simulation`, `/discussions/:id`, etc. — **no** `/th/...` URL prefix) so existing bookmarks, deep links, and `user_notifications.link_url` values never break.

### 2.2 Content Boundary Matrix

| Surface / Artifact | Behavior in `en` | Behavior in `th` | Invariant |
| --- | --- | --- | --- |
| Frontend UI Chrome (16 pages, Navbar, Footer, Modals, Badges, Form Labels, KQL Helper Chrome, Exam/Sim Navigation Chrome) | English | Thai | Live instant switch, zero reload, zero state loss |
| Security Domain Names (`domain1`–`domain5`) | Official SY0-701 EN names | Thai + EN in parentheses (mapped client-side by `domain_id`) | Backend `domain_name` stays canonical EN for analytics compatibility |
| Learning Theory Markdown (5 Overviews + 28 Sub-topics) | `theory/**/*.md` | `theory/th/**/*.md` via `theory_translations` | Preserves KQL code blocks, log snippets, and clickable `https://` citations |
| Simulation Scenario Non-Log Content (`title`, `description`, `playbook_steps`, `findings`, `tp_fp_explanation`) | Canonical EN fields | `translations.th` JSONB | Preserves IOCs (hosts, IPs, users, sources) and MITRE IDs verbatim in EN |
| API Error Messages & In-App Notifications | EN (`error` / `title` / `message`) | Frontend renders from `code`/`params` and `template_key`/`template_params` | Additive fields preserve legacy EN fields for backward compatibility |
| Pre-Test & Post-Test (`question_text`, `options`, `correct_answer`, `explanation`) | English | **Strictly English** + subtle TH info notice | Zero Thai Unicode (`\u0E00–\u0E7F`) |
| Formative Sample Questions in `/learn` (`question_text`, `options`, `correct_answer`, `explanation`) | English | **Strictly English** + subtle TH info notice | Zero Thai Unicode (`\u0E00–\u0E7F`) |
| Simulation SIEM Telemetry (`initial_logs`) & KQL Syntax | English | **Strictly English** + subtle TH info notice | Zero Thai Unicode (`\u0E00–\u0E7F`) |
| Community Content (Analysis Threads, Comments, Admin Rejection Reason text) | Author's original text | Author's original text (surrounding UI chrome translates) | Never auto-translated |

### 2.3 Hybrid Bilingual Terminology Policy (Thai Mode)
- Use natural Thai prose for UI instructions, explanations, narratives, and theory prose, while **retaining standard CompTIA Security+ and SOC technical terms in English**:
  - Acronyms & tool terms stay English: `SIEM`, `KQL`, `EDR`, `IDS/IPS`, `MITRE ATT&CK`, `MFA`, `SSO`, `IAM`, `DLP`, `PKI`, `TLS`, `DNS`, `VPN`, `Pre-Test`, `Post-Test`, `SOC`, `IOC`, `CVSS`.
  - Key operational concepts pair Thai clarity with English terms: e.g., `ยืนยันเหตุภัยคุกคามจริง (True Positive)`, `สัญญาณเตือนลวง (False Positive)`, `การควบคุมขอบเขต (Containment)`, `การกำจัดภัยคุกคาม (Eradication)`, `การฟื้นฟูระบบ (Recovery)`, `คะแนนความเชี่ยวชาญ (Proficiency Score)`, `หลักฐานสำคัญ (Key Finding)`.
  - Canonical bilingual domain display names (rendered by Frontend from `domain_id`):
    - `domain1`: `แนวคิดความปลอดภัยทั่วไป (General Security Concepts)`
    - `domain2`: `ภัยคุกคาม ช่องโหว่ และการลดความเสี่ยง (Threats, Vulnerabilities, and Mitigations)`
    - `domain3`: `สถาปัตยกรรมความปลอดภัย (Security Architecture)`
    - `domain4`: `ปฏิบัติการด้านความปลอดภัย (Security Operations)`
    - `domain5`: `การบริหารจัดการและกำกับดูแลความปลอดภัย (Security Program Management and Oversight)`
- **Date & Time Formatting**:
  - UI timestamps (notifications, discussions, profile, completion dates, review lists): formatted via `Intl.DateTimeFormat` using `en-US` for `en` and `th-TH-u-ca-gregory` for `th` (Thai month names with **Gregorian / ค.ศ.** year, e.g., `20 พ.ย. 2026`).
  - SIEM telemetry timestamps inside `LogTable` / `ScenarioResultPage` / `SpecialistReviewPage` remain verbatim ISO/RFC3339 Gregorian strings.

### 2.4 Intentional English-Only Notices (Thai Mode Only)
When `locale === "th"`, render a compact, non-intrusive subtle info notice (`bg-graphite-900/70 border border-graphite-800 text-graphite-300 text-xs`) on the three English-only training surfaces:
1. **`ExamPage.tsx`** (Pre-Test & Post-Test header): *"ข้อสอบคงเป็นภาษาอังกฤษตามมาตรฐาน CompTIA Security+ (SY0-701)"*
2. **`LearnTheoryPage.tsx`** (Formative Sample Questions section header): *"ข้อสอบตัวอย่างคงเป็นภาษาอังกฤษตามมาตรฐานข้อสอบ CompTIA Security+"*
3. **`SimulationPage.tsx`** (above `LogTable` in Detect & Analyze): *"SIEM Logs และคำสั่ง KQL คงเป็นภาษาอังกฤษเพื่อจำลองการใช้งานเครื่องมือ SOC จริง"*

When `locale === "en"`, these notices are omitted.

---

## 3. Database Schema & Migrations (`init.sql` & `models/migrate.go`)

Every schema change MUST be applied cleanly in **both** `init.sql` (fresh installations) and `models/migrate.go` (idempotent startup migrations for existing databases):

```sql
-- 1. User language preference (users)
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS preferred_language VARCHAR(10) NOT NULL DEFAULT 'en';

-- 2. Dedicated multilingual theory translations (keeps rag_documents vector(768) untouched)
CREATE TABLE IF NOT EXISTS theory_translations (
  entity_type VARCHAR(16) NOT NULL CHECK (entity_type IN ('domain_overview', 'subtopic')),
  entity_id   VARCHAR(32) NOT NULL,
  locale      VARCHAR(10) NOT NULL,
  title       VARCHAR(255) NOT NULL,
  content     TEXT NOT NULL,
  updated_at  TIMESTAMP DEFAULT now(),
  PRIMARY KEY (entity_type, entity_id, locale)
);
CREATE INDEX IF NOT EXISTS idx_theory_translations_locale
  ON theory_translations(locale, entity_type);

-- 3. Bilingual scenario translations JSONB (scenarios)
ALTER TABLE scenarios
  ADD COLUMN IF NOT EXISTS translations JSONB NOT NULL DEFAULT '{}'::jsonb;

-- 4. Structured notification localization metadata (user_notifications)
ALTER TABLE user_notifications
  ADD COLUMN IF NOT EXISTS template_key VARCHAR(64),
  ADD COLUMN IF NOT EXISTS template_params JSONB;
```

### Model Struct Updates
- `models.User`: add `PreferredLanguage string` (`gorm:"type:varchar(10);not null;default:'en'" json:"preferred_language"`).
- `models.TheoryTranslation`: maps to `theory_translations` (`EntityType`, `EntityID`, `Locale`, `Title`, `Content`, `UpdatedAt`).
- `models.Scenario`: add `Translations any` (`gorm:"type:jsonb;serializer:json;not null;default:'{}'" json:"translations,omitempty"`).
- `models.UserNotification`: add `TemplateKey *string` (`gorm:"type:varchar(64)" json:"template_key,omitempty"`) and `TemplateParams any` (`gorm:"type:jsonb;serializer:json" json:"template_params,omitempty"`).

---

## 4. Language Preference Persistence, Sync & CORS

### 4.1 API Endpoints & DTOs (`dto/auth.go`, `controllers/auth.go`, `main.go`)
- `POST /api/register` (`dto.RegisterRequest`): accepts optional `preferred_language` (`"en"` or `"th"`; defaults to `"en"` if omitted/empty; returns `400` with `code: "INVALID_LANGUAGE"` if invalid).
- `POST /api/login` (`dto.AuthResponse`): includes `preferred_language`.
- `GET /api/user/profile` (`dto.UserProfileResponse`): includes `preferred_language`.
- `PUT /api/user/language` (protected by `middlewares.AuthMiddleware()`):
  - Request (`dto.UpdateLanguageRequest`): `{"language": "en" | "th"}` (required).
  - Validates against allowlist `{"en", "th"}`; rejects invalid values with `400` (`code: "INVALID_LANGUAGE"`).
  - Updates `users.preferred_language` and returns `200` (`{"message": "Language preference updated", "preferred_language": "..."}`).
- CORS in `main.go`: add `"Accept-Language"` to `AllowHeaders`.

### 4.2 Precedence & Sync State Machine
1. **Unauthenticated (Guest)**:
   - Reads `localStorage.getItem("cybersim_lang")` on boot; defaults to `"en"` if missing or invalid.
   - Toggling `EN | TH` updates React Context + `localStorage` + `document.documentElement.lang` immediately.
2. **Registration**:
   - Frontend sends the current guest `locale` as `preferred_language` in `POST /api/register`, so the newly registered account immediately inherits the language the user chose on Landing/Register.
3. **Login (DB Wins for Multi-Device Consistency)**:
   - On successful `POST /api/login` (and on authenticated hydration via `GET /api/user/profile`), Frontend adopts `preferred_language` from the server response and syncs `localStorage.setItem("cybersim_lang", serverLang)`.
4. **Explicit Toggle While Logged In**:
   - Clicking `EN | TH` in the Top Navbar or changing Language Preference on `ProfilePage`:
     1. Optimistically updates React Context, `localStorage`, and `document.documentElement.lang` immediately (0 ms perceived latency).
     2. Fires `PUT /api/user/language` in the background (with JWT). If the API fails transiently, keep the local UI responsive and log a warning without disrupting active exam/simulation state.
5. **Logout**:
   - Removes auth keys (`token`, `userEmail`, `userNickname`, `userRole`, `userId`) while **preserving** `cybersim_lang` in `localStorage` so the Login screen stays in the user's current language.

---

## 5. Learning Theory Localization (`theory/th/` & `/api/learn/*`)

### 5.1 File Layout & Go Embedding
Keep canonical English files untouched at `theory/domain{1..5}_*.md` and `theory/domain{1..5}/*.md`. Mirror all 33 Thai translations under `theory/th/`:
- 5 Domain Overviews:
  - `theory/th/domain1_security_concepts.md`
  - `theory/th/domain2_threats_vulnerabilities.md`
  - `theory/th/domain3_security_architecture.md`
  - `theory/th/domain4_security_operations.md`
  - `theory/th/domain5_security_management.md`
- 28 Sub-topics:
  - `theory/th/domain1/1_1_security_controls.md` … `theory/th/domain5/5_6_security_awareness_practices.md` (1:1 identical filenames to English).

Update `main.go` embed directive:
```go
//go:embed theory/*.md theory/domain1/*.md theory/domain2/*.md theory/domain3/*.md theory/domain4/*.md theory/domain5/*.md theory/th/*.md theory/th/domain1/*.md theory/th/domain2/*.md theory/th/domain3/*.md theory/th/domain4/*.md theory/th/domain5/*.md
var theoryMarkdownFS embed.FS
```

### 5.2 Provenance & Content Rules for Thai Markdown
- Every Thai Markdown file is a faithful translation of its registered English counterpart and MUST preserve the exact same clickable `https://` reference links (no new unregistered external sources are introduced).
- Required structural invariants per Thai file (validated by `scripts/load_theory_to_db.py` and `services/theory_service.go`):
  1. Top-level `# <Title>` heading (used as the localized `title` stored in `theory_translations`).
  2. Numbered section headings:
     - Domain Overviews: `## 1.`, `## 2.`, `## 3.`, `## 4.`
     - Sub-topics: `## 1.`, `## 2.`, `## 3.`
  3. Contains KQL guidance (`kql`), TP/FP guidance (`true positive` and `false positive`, preserved via Q6A bilingual terminology), and clickable `](https://` reference links matching the EN file's reference URLs.
  4. Contains Thai script characters (`[\u0E00-\u0E7F]`) so untranslated English stubs cannot silently pass.
  5. Code blocks (KQL queries, SIEM log examples, CLI commands) stay strictly in English.

### 5.3 Seeding & Service Loading
- `scripts/load_theory_to_db.py`:
  - Continues upserting EN Domain Overviews into `rag_documents` and EN Sub-topics into `theory_subtopics`.
  - Reads, validates, and upserts all 33 `theory/th/**` files into `theory_translations` (`ON CONFLICT (entity_type, entity_id, locale) DO UPDATE SET title = EXCLUDED.title, content = EXCLUDED.content, updated_at = NOW()`).
  - Supports `--validate-only` flag for CI execution without requiring a live database.
- `services/theory_service.go`:
  - Adds locale-aware loaders `LoadTheoryModulesByLocale(db, locale)` and `LoadTheorySubtopicsByLocale(db, locale)`.
  - For `locale == "en"`: uses existing canonical loaders.
  - For `locale == "th"`: queries `theory_translations` (`locale = 'th'`), falls back to embedded `theory/th/` FS if DB row missing, and falls back to EN only if both are absent.

### 5.4 API Contract (`controllers/learn_controller.go`)
- `GET /api/learn/domains?lang=en|th` and `GET /api/learn/subtopics/:subTopicId?lang=en|th`:
  - Locale resolution priority: query param `?lang` → `Accept-Language` header → authenticated user's `users.preferred_language` → `"en"`.
  - Returns localized `title` and `content` in the existing DTO fields (`TheoryDomainDTO`, `TheorySubtopicSummaryDTO`, `LearnSubtopicDetailResponse`).
  - `LearnSubtopicDetailResponse.SampleQuestions` (`question_text`, `options`, `correct_answer`, `explanation`) MUST always return the canonical English rows from `questions` untouched.
- Frontend (`LearnTheoryPage.tsx`):
  - Passes `?lang=${locale}` on both endpoints and refetches automatically when `locale` changes, preserving the active `domain` / `topic` search params, sidebar expand state, and sample question UI state.

---

## 6. Bilingual Simulation Scenarios (Static Seeds, Runtime AI, Snapshots & Hygiene)

### 6.1 Canonical `translations` JSONB Schema
Canonical columns on `scenarios` (`title`, `description`, `tp_fp_explanation`, `playbook_steps`, `expected_outcomes`, `initial_logs`) remain **English** so seed matching (`title + domain_id`), session history exclusion, and `buildInvestigationShortcuts` remain deterministic and backward-compatible.

Localized non-log content lives in `scenarios.translations` (and is captured inside immutable `scenario_snapshots.scenario_data`):

```json
{
  "th": {
    "title": "การหลบเลี่ยง MFA ด้วยวิธี Push Fatigue",
    "description": "ระบบแจ้งเตือนพบผู้ใช้ 'ajones' ปฏิเสธการแจ้งเตือน MFA Push ติดต่อกันหลายครั้งภายในหนึ่งนาที ตามด้วยการเข้าสู่ระบบสำเร็จจาก External IP 198.51.100.42...",
    "tp_fp_explanation": "กรณีนี้คือ True Positive เนื่องจากการส่งคำขอ MFA Push ซ้ำหลายครั้งแล้วตามด้วยการล็อกอินสำเร็จจาก IP ภายนอก บ่งชี้ถึงการโจมตีแบบ MFA Push Fatigue...",
    "playbook_steps": {
      "d1_c1": {
        "label": "เพิกถอน Okta SSO tokens และ web sessions ที่ใช้งานอยู่ทั้งหมดของบัญชี ajones ทันที",
        "explanation": "ยุติเซสชันที่ผู้โจมตียึดครองไว้ได้ทันที"
      }
    },
    "findings": {
      "d1_f1": {
        "description": "ตรวจพบคำขอ MFA Push ที่ถูกปฏิเสธซ้ำหลายครั้งก่อนมีการล็อกอินสำเร็จ",
        "explanation": "ลำดับเหตุการณ์ MFA ที่ถูกปฏิเสธติดต่อกันหลายครั้งแล้วอนุมัติสำเร็จยืนยันรูปแบบ Push Fatigue"
      }
    }
  }
}
```

### 6.2 Pre-Submission Sanitization & Answer-Leak Hygiene
**CRITICAL INVARIANT**: Pre-submission endpoints (`POST /api/simulation/start`, `GET /api/simulation/session`) returning `dto.ScenarioResponse` MUST sanitize `translations` just like canonical fields:
- **Allowed in `ScenarioResponse.Translations["th"]`**:
  - `title`
  - `description`
  - `playbook_steps`: map of `action_id -> { "label": "..." }` (**NO `explanation`**, **NO `is_correct`**)
- **STRICTLY FORBIDDEN in `ScenarioResponse.Translations`**:
  - `tp_fp_explanation`
  - `findings` (neither descriptions, explanations, nor indices)
  - `playbook_steps[id].explanation`

Post-submission endpoints (`POST /api/simulation/submit`, `GET /api/simulation/result/:sessionId`, and `GET /api/specialist/submissions/:session_id`) include the **full** `translations` map from the session's immutable `scenario_snapshots` (falling back to `scenarios.translations` for pre-snapshot rows), enabling instant 0 ms `EN | TH` toggling on `SimulationPage`, `ScenarioResultPage`, and `SpecialistReviewPage` with **zero network refetch**.

### 6.3 Static Seed Pipeline (`scripts/scenario_translations_th.json` & `scripts/generate_scenarios.py`)
- Store Thai translations for all 15 static scenarios in a dedicated structured file `scripts/scenario_translations_th.json`, keyed by canonical English scenario `title`.
- `scripts/generate_scenarios.py` loads `scripts/scenario_translations_th.json` and attaches `translations: { "th": ... }` to each generated entry in `scenarios_seed.json` **without altering `initial_logs` or `evidence_log_indices`** (preserving `scripts/evidence_mapping.json` hashes intact).
- `scripts/load_scenarios_to_db.py` ensures the `translations` column exists and upserts `translations` JSONB into `scenarios` for all 15 static seeds.

### 6.4 Runtime AI Bilingual Single-Call Generation (`services/llm_service.go`)
- Expand `aiMicroScenario`, `aiFinding`, and `aiResponseActionOption` to request Thai paired fields in the **same single LLM call**:
  - `aiMicroScenario`: `TitleTh`, `DescriptionTh`, `TpFpExplanationTh`
  - `aiFinding`: `DescriptionTh`, `ExplanationTh`
  - `aiResponseActionOption`: `LabelTh`, `ExplanationTh`
  - `incident_logs`: **English only** (no `*_th` fields exist on `aiIncidentLog`).
- Prompt rules enforced in `GenerateLLMScenarioWithContext`:
  1. `incident_logs` and all IOCs (`host`, `src_ip`, `dst_ip`, `user`, `source`, `message`) MUST be **100% English** with zero Thai characters.
  2. Every `*_th` field MUST be written in natural Thai following Hybrid Bilingual Terminology (Q6A), keeping hostnames, IPs, usernames, log source names, and MITRE IDs verbatim in English.
  3. Match the assigned `BriefingStyle` tone in both `description` and `description_th`.
- `synthesizeFullScenarioWithErrors` maps each generated `action_id` and `finding_id` to its `*_th` label/description/explanation inside `scenario.Translations["th"]`.
- `GenerateFallbackScenario` and domain-aware contextual distractor backfill emit complete `translations.th` for every domain (`domain1`–`domain5`) and both True Positive & False Positive paths.

### 6.5 Strict Validation Gate & Legacy Archival (`ValidateScenarioCandidate` & Startup)
- **Strict Reject & Fallback (`ValidateScenarioCandidate`)**:
  1. **Zero-Thai-in-Logs Gate**: Reject any scenario if ANY string field inside `initial_logs` contains a Thai Unicode rune (`\u0E00`–`\u0E7F`).
  2. **Complete `translations.th` Gate**: Reject any scenario if `translations.th` is missing or empty for `title`, `description`, `tp_fp_explanation`, ANY `playbook_steps` `action_id` (`label` + `explanation`), or ANY `findings` `finding_id` (`description` + `explanation`), OR if `translations.th.title` / `translations.th.description` lack Thai characters (`\u0E00`–`\u0E7F`).
  3. When runtime LLM output fails validation, reject it and fall back cleanly to bilingual `GenerateFallbackScenario` or the bilingual Static Seed Pool (never persisting half-translated AI scenarios).
- **Startup Legacy Archival (`EnsureScenarioEvidenceQualityFromSeed`)**:
  - Any active **non-seed** scenario in `scenarios` (`status != 'archived'` and title not in `scenarios_seed.json`) that lacks valid `translations.th` (or fails `ValidateScenarioCandidate`) is automatically marked `status = 'archived'` on backend startup.
  - Completed `simulation_sessions` and `scenario_snapshots` remain strictly immutable; old snapshots lacking `translations.th` fall back gracefully to canonical EN in the UI.

### 6.6 Context-Aware SIEM Shortcuts Invariant (`frontend/src/lib/siemShortcuts.ts`)
- `buildInvestigationShortcuts` MUST always receive canonical English `scenario.title` and `scenario.description` (NOT the Thai translated fields), ensuring ASCII entity/keyword boosting produces identical, accurate shortcuts in both `en` and `th`.
- Only the shortcut button prefix labels (`Host:`, `IP:`, `User:`, `Source:`, `Filter:`) may be localized via UI chrome if desired; the underlying KQL `query` strings remain 100% English.

---

## 7. Additive Backend API Errors & In-App Notifications Contracts

### 7.1 Standardized API Error Contract (`dto/common.go` & Controllers)
Extend `dto.ErrorResponse` additively (preserving `error` in English so no existing integration test breaks):
```go
type ErrorResponse struct {
	Error   string         `json:"error" example:"Invalid request payload"`
	Message string         `json:"message,omitempty" example:"Detailed error description"`
	Code    string         `json:"code,omitempty" example:"BAD_REQUEST"`
	Params  map[string]any `json:"params,omitempty"`
}
```
All user-facing controller error responses include a canonical `code` (and `params` when interpolated), mapped by Frontend `t('errors.' + code, params)` with automatic fallback to `error`:
- **Auth & Profile**: `AUTH_UNAUTHORIZED`, `AUTH_FORBIDDEN`, `AUTH_INVALID_PAYLOAD`, `AUTH_INVALID_CREDENTIALS`, `AUTH_EMAIL_EXISTS`, `AUTH_NICKNAME_REQUIRED`, `AUTH_NICKNAME_TAKEN`, `AUTH_NICKNAME_INVALID`, `INVALID_LANGUAGE`, `USER_NOT_FOUND`.
- **Specialist & Admin**: `SPECIALIST_BIO_REQUIRED`, `SPECIALIST_RESUME_REQUIRED`, `SPECIALIST_FILE_TOO_LARGE`, `SPECIALIST_INVALID_FILE_TYPE`, `SPECIALIST_APPLICATION_EXISTS`, `ADMIN_APPLICATION_NOT_FOUND`, `ADMIN_REASON_REQUIRED`.
- **Exam**: `EXAM_PRETEST_REQUIRED`, `EXAM_POSTTEST_LOCKED`, `EXAM_SESSION_NOT_FOUND`, `EXAM_QUESTION_NOT_FOUND`, `EXAM_INVALID_ANSWER`.
- **Learn**: `LEARN_PRETEST_REQUIRED`, `LEARN_DOMAIN_NOT_FOUND`, `LEARN_SUBTOPIC_NOT_FOUND`, `LEARN_SUBTOPICS_INCOMPLETE`, `LEARN_SAMPLE_NOT_FOUND`, `LEARN_SAMPLE_INCOMPLETE`, `LEARN_SAMPLE_ALREADY_ANSWERED`.
- **Simulation**: `SIM_PRETEST_REQUIRED`, `SIM_THEORY_GATE_LOCKED` (params: `completed_subtopic_count`, `total_subtopics`), `SIM_SESSION_NOT_FOUND`, `SIM_ALREADY_SUBMITTED`, `SIM_TPFP_REQUIRED`, `SIM_INVALID_LOG_INDEX`, `SIM_NO_SCENARIO_AVAILABLE`.
- **Community & Notifications**: `THREAD_NOT_FOUND`, `THREAD_LOCKED`, `THREAD_FORBIDDEN`, `COMMENT_NOT_FOUND`, `COMMENT_FORBIDDEN`, `NOTIFICATION_FETCH_FAILED`, `INTERNAL_SERVER_ERROR`.

### 7.2 In-App Notifications (`user_notifications` & `NotificationBell.tsx`)
All 4 notification creation sites populate both legacy English `title`/`message` AND structured `template_key`/`template_params`:

| Event | `template_key` | `template_params` |
| --- | --- | --- |
| Admin approves Specialist application | `specialist_application_approved` | `{}` |
| Admin rejects Specialist application | `specialist_application_rejected` | `{"reason": "<admin free-text reason>"}` |
| Reply to user's comment | `comment_reply` | `{"actor": "<nickname>", "thread_title": "<title>"}` |
| New comment on Specialist's thread | `thread_comment` | `{"actor": "<nickname>", "thread_title": "<title>"}` |

- `dto.NotificationResponse` adds `TemplateKey *string` (`json:"template_key,omitempty"`) and `TemplateParams any` (`json:"template_params,omitempty"`).
- `NotificationBell.tsx` renders via `t('notifications.' + template_key + '.title', template_params)` and `t('notifications.' + template_key + '.message', template_params)` when `template_key` is present and recognized, falling back to `title` and `message` for pre-existing legacy rows.

---

## 8. Frontend i18n Architecture, Typography & UX

### 8.1 Zero-Dependency Typed i18n Engine (`frontend/src/i18n/`)
No new npm packages are added to `frontend/package.json`. Create:
- `frontend/src/i18n/types.ts`: `export type Locale = "en" | "th"`, `TranslationKey`, `TranslationParams`, and `TranslationDictionary` derived from `typeof en`.
- `frontend/src/i18n/en.ts`: canonical English dictionary organized by domain namespace (`common`, `nav`, `auth`, `landing`, `profile`, `exam`, `learn`, `evaluation`, `simulation`, `result`, `discussions`, `specialist`, `admin`, `notifications`, `domains`, `errors`).
- `frontend/src/i18n/th.ts`: Thai dictionary typed with `satisfies TranslationDictionary` (so `tsc -b` fails compilation immediately if any key is missing, extra, or mistyped).
- `frontend/src/i18n/format.ts`:
  - `interpolate(template, params)` supporting `{{name}}` placeholders.
  - `formatDate(iso, locale, options?)` using `Intl.DateTimeFormat(locale === "th" ? "th-TH-u-ca-gregory" : "en-US", ...)`.
  - `resolveScenarioLocale(scenario, locale)` returning localized `title`, `description`, `tp_fp_explanation`, playbook labels/explanations, and finding descriptions/explanations with automatic EN fallback while keeping `initial_logs` untouched.
  - `resolveApiError(err, t)` reading `err.response?.data?.code` + `params` and falling back to `err.response?.data?.error`.
- `frontend/src/i18n/index.tsx`: `I18nProvider` (mounted at root in `App.tsx`) + `useI18n()` hook exposing `{ locale, setLocale, syncLocaleFromUser, t, formatDate }`.
  - Sets `document.documentElement.lang = locale` inside `useEffect`.
  - Attaches an `axios` request interceptor setting `Accept-Language: <locale>` automatically on all API calls.

### 8.2 Typography (`frontend/index.html` & `frontend/tailwind.config.js`)
- Load **`IBM Plex Sans Thai`** (`400;500;600;700`) alongside `Geist` and `Geist Mono` in `frontend/index.html`.
- Update `frontend/tailwind.config.js`:
  ```js
  fontFamily: {
    sans: ['Geist', '"IBM Plex Sans Thai"', 'sans-serif'],
    mono: ['"Geist Mono"', 'monospace'],
  }
  ```
- Maintain `font-mono` (`Geist Mono`) exclusively for SIEM logs, KQL inputs, IP/host badges, and telemetry timestamps.

### 8.3 Switcher UX & Live State Preservation
- **Top Navbar (`Layout.tsx`)**: Place a compact segmented `EN | TH` button group (`aria-label` accessible, Graphite theme styling matching `bg-graphite-900 border-graphite-800`) visible on **every route** (both guest Landing/Login/Register and authenticated routes).
- **Profile Page (`ProfilePage.tsx`)**: Add a Language Preference card/selector synchronized with the same `useI18n().setLocale` handler.
- **Zero State Loss**: Switching `EN | TH` NEVER triggers `window.location.reload()` and NEVER unmounts page components. Active Exam answers, Simulation KQL query text, selected evidence log indices, TP/FP radio choice, checked Playbook checkboxes, and Discussion comment/thread editor drafts remain 100% intact across switches.

---

## 9. Content Translation Authoring Loop & Push-Right Checkpoint

To ensure Thai cybersecurity prose is accurate and matches the user's preferred terminology before translating all 33 theory docs + 15 static scenarios:

1. **Stage 1 — Pilot Batch Authoring**:
   - Translate **Domain 1 Theory** (5 files: `theory/th/domain1_security_concepts.md` + `theory/th/domain1/1_1_*.md` through `1_4_*.md`).
   - Translate **3 representative Static Scenarios** in `scripts/scenario_translations_th.json` (covering both True Positive and False Positive scenarios).
   - Run automated structural & zero-Thai-in-logs validators against the pilot batch.
2. **Stage 2 — Single Blocking Human Review Checkpoint (Brief)**:
   - Present a concise, decision-ready **Brief** containing:
     - Standardized TH/EN cybersecurity glossary table used across the pilot.
     - Short representative excerpt from 1 Domain Overview + 1 Sub-topic.
     - Short representative excerpt from 1 True Positive + 1 False Positive scenario (Title, Narrative, 2 Playbook options, TP/FP Explanation).
     - Clickable links to the pilot files for optional deep inspection.
   - Wait for user approval or tone adjustments.
3. **Stage 3 — Full Scale-Out**:
   - Incorporate any feedback into the glossary, translate remaining **28 Theory files** (Domains 2–5) and remaining **12 Static Scenarios**, run full validation suite, and regenerate `scenarios_seed.json`.

---

## 10. Verification Criteria & CI Guardrails (Definition of Done)

Every PR touching localization, UI copy, theory docs, or scenarios must pass all 3 automated gates alongside existing test suites:

1. **Gate 1 — Frontend TypeScript, i18n Dictionary Parity & Build**:
   - `npm --prefix frontend run build` (`tsc -b && vite build`) passes with zero errors.
   - Automated i18n parity check script (`scripts/validate_i18n.py` or Node check wired into CI) asserts:
     - 100% key equality between `frontend/src/i18n/en.ts` and `frontend/src/i18n/th.ts`.
     - Zero empty translation strings.
     - 100% identical `{{param}}` placeholder sets for every matching EN/TH key pair.
     - `frontend/src/lib/siemShortcuts.ts` tests confirm identical shortcuts when given EN title/description regardless of UI locale.
2. **Gate 2 — Theory & Static Seed Translation Parity**:
   - `python scripts/load_theory_to_db.py --validate-only` passes for all 33 EN files AND all 33 TH files (`theory/th/**`), checking required headers, KQL presence, TP/FP terms, Thai Unicode presence, and exact `https://` citation parity with EN counterparts.
   - `python -m pytest scripts/test_theory_seeding.py` passes (extended with `theory_translations` seeding & validation tests).
   - `python scripts/generate_scenarios.py --mapping scripts/evidence_mapping.json` succeeds without changing `evidence_mapping.json` hashes.
   - `python scripts/validate_evidence_mapping.py --seed scenarios_seed.json` verifies all 15 static seeds contain complete `translations.th` covering every `action_id` and `finding_id`, AND contain **zero** Thai Unicode characters (`[\u0E00-\u0E7F]`) inside `initial_logs`.
   - `python -m pytest scripts/test_evidence_authoring.py` passes.
3. **Gate 3 — Backend Go Tests, Strict English-Only Invariants & OpenAPI Sync**:
   - `go test ./...` passes, including new unit/controller tests verifying:
     - `PUT /api/user/language`, `POST /api/register` (`preferred_language`), `POST /api/login`, and `GET /api/user/profile` preference sync & validation (`400 INVALID_LANGUAGE`).
     - `GET /api/learn/domains?lang=th` & `GET /api/learn/subtopics/:id?lang=th` return Thai Markdown while `sample_questions` remain strictly English (zero Thai Unicode).
     - `POST /api/exams/pre-test` & `POST /api/exams/post-test` remain strictly English regardless of `Accept-Language: th` or `users.preferred_language = 'th'`.
     - `ValidateScenarioCandidate` rejects Thai characters inside `initial_logs` and rejects incomplete `translations.th`.
     - `StartScenario` / `GetActiveSession` sanitize `translations.th` (never exposing `tp_fp_explanation`, `findings`, or playbook `explanation`).
     - `EnsureScenarioEvidenceQualityFromSeed` archives legacy non-seed active AI scenarios missing valid `translations.th` while preserving completed sessions/snapshots.
     - Notification creation populates `template_key` and `template_params` alongside EN `title`/`message`.
   - Regenerate Swagger OpenAPI specs (`scripts/generate_docs.sh` / `generate_docs.ps1`) per `ADR-0005`.
   - Existing end-to-end integration scripts (`test_integration.py`, `test_spec1_integration.py`, `test_spec2_integration.py`, `test_spec3_integration.py`, `test_simulation_integration.py`) pass with zero regressions.
