# Workflow Spec: Bilingual Hybrid Thai Exam Questions (All 540 Items) (Issue #71)

**GitHub Issue:** [#71 — feat(exams,i18n): add Thai localization support for Pre-Test and Post-Test questions](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/71)  
**Status:** `/loop-me` Grilling Rounds 1–3 Settled (`Q1`, `Q4=B`, `Q9`–`Q10`, `Q13` Resolved, 2026-11-27) — Ready for Implementation  
**Depends On / Amends:**
- Amends `workflows/language_switching_and_localization.md` & `NOTES.md` Decision #46: Narrows the **English-Only Realism Invariant** strictly to canonical English question columns (`questions.question_text`, `options`, `explanation`) and **SIEM `initial_logs`, KQL syntax, IOCs, and MITRE ATT&CK IDs**, enabling full Hybrid Bilingual Thai translations under `questions.translations.th`.
- Integrates with `workflows/exam_qa_review.md` (`parseExamExplanation` Distractor Breakdown on `/exams/review/:sessionId`) and `workflows/nfr_evaluation.md` (`services/nfr_evaluator.go` Safety probe).
- Follows `docs/adr/0005-openapi-code-first-with-dedicated-dtos.md`.

---

## Objective

Eliminate the English-only reading barrier for Thai high school novices across **all 540 CompTIA Security+ (SY0-701) exam questions** while preserving 100% canonical English ground truth, stable `question_id` UUIDs, Zero-Overlap (`Pre-Test ∪ Learn ∪ Post-Test = ∅`), and uniform 540-question random sampling (`4/7/5/8/6` domain blueprint, no difficulty cap):

1. Store verified **Hybrid Bilingual Thai** translations (`question_text`, `options.A..D`, structured `explanation`) in `questions.translations` JSONB (exact same pattern as `scenarios.translations`).
2. Render Thai seamlessly when UI Locale = `th` across **all 4 question surfaces** (**Pre-Test / Post-Test**, **Post-Exam Q&A Review**, **`/learn` Sample Questions**, and **Landing Sample Quiz**) with a 1-click per-question **"ดูต้นฉบับ EN / Show EN"** toggle.
3. Preserve structured **Distractor Breakdown** parsing (`Correct Rationale` + per-option incorrect notes with `Your pick` highlight) in both English and Thai.

---

## 1. Loop Profile

- **Loop Name**: Bilingual 540-Question Authoring, Validation & Delivery Loop
- **Trigger**:
  - Authoring/Seed: Offline batch translation (`scripts/translate_questions_th.py`) → deterministic validation → committed `scripts/question_translations_th.json` → idempotent DB seed (`scripts/load_exams_to_db.py`).
  - Runtime: Instant client-side `EN | TH` and per-question **"ดูต้นฉบับ EN"** toggling with zero extra network round-trips or runtime LLM calls.
- **Checkpoint**:
  - **Checkpoint 1 (Content & Tone Gate — Push-Right, `Q13`)**: Generate and present **10 sample Thai question translations** (2 per domain across `domain1`–`domain5`) in a compact Brief for tone/terminology approval before batch-translating all remaining 530 questions.
  - **Checkpoint 2 (Pre-Flight Visual & Verification Gate — `Q13`)**: Present EN/TH screenshots of `ExamPage` (with per-question Show EN toggle) and `ExamReviewPage` (Thai Distractor Breakdown) + full test suite green before closing `#71`.

---

## 2. Settled Design Decisions

| Q# | Decision | Settled Rule |
| --- | --- | --- |
| `Q4=B` | Translation Coverage | Translate **all 540/540 questions** (`parsed_questions.json`) into Hybrid Bilingual Thai (`question_text`, `options`, `explanation`). |
| `Q9` | Storage, Seed & UX Surfaces | Add `questions.translations JSONB NOT NULL DEFAULT '{}'::jsonb` (`init.sql` + `models/migrate.go`), seeded from `scripts/question_translations_th.json` via `scripts/load_exams_to_db.py` (preserving UUIDs and canonical EN columns untouched). Deliver inline across all 4 surfaces (Pre/Post Exam, Exam Review, Learn Sample Questions, Landing Sample Quiz) with strict pre-submission sanitization (`explanation` omitted before grading) + per-question 1-click **"ดูต้นฉบับ EN / Show EN"** toggle + Thai Distractor Breakdown parser. |
| `Q10` | Difficulty Sampling | Keep uniform random draw across **all 540 questions** (`difficulty` 1–5) per the `4/7/5/8/6` domain blueprint — **do NOT add `EXAM_MAX_DIFFICULTY`**. |

---

## 3. Database & Translation Data Contract

### 3.1 Schema Update (`init.sql` + `models/migrate.go` + `scripts/load_exams_to_db.py`)
Add to `questions` table:
```sql
translations JSONB NOT NULL DEFAULT '{}'::jsonb
```
Idempotent migration statement in `models/migrate.go` and `scripts/load_exams_to_db.py`:
```sql
ALTER TABLE questions ADD COLUMN IF NOT EXISTS translations JSONB NOT NULL DEFAULT '{}'::jsonb;
```

### 3.2 Canonical JSONB Payload (`questions.translations`)
```json
{
  "th": {
    "question_text": "บริษัท AlphaTech กำลังทำ Risk Analysis สำหรับระบบ Online Payment Gateway โดยคำนวณค่า Annual Rate of Occurrence (ARO) ของช่องโหว่หนึ่งได้เท่ากับ 2 และ Single Loss Expectancy (SLE) เท่ากับ $50,000 บริษัทควรคาดการณ์ความเสียหายต่อปี (ALE) เป็นเท่าใด?",
    "options": {
      "A": "$10,000",
      "B": "$100,000",
      "C": "$25,000",
      "D": "$1,000,000"
    },
    "explanation": "คำตอบที่ถูกต้อง: B. $100,000 ค่า Annualized Loss Expectancy (ALE) คำนวณจาก ARO คูณด้วย SLE ดังนั้น ALE = 2 × $50,000 = $100,000 ตัวเลือก A ไม่ถูกต้อง เพราะไม่ได้ใช้สูตรคำนวณ ALE อย่างถูกต้อง ตัวเลือก C ไม่ถูกต้อง เพราะเป็นเพียงครึ่งหนึ่งของ SLE ไม่ใช่ความเสียหายรายปี ตัวเลือก D ไม่ถูกต้อง เพราะสูงกว่าค่า ALE ที่คำนวณได้มาก"
  }
}
```

### 3.3 Committed Translation File (`scripts/question_translations_th.json`)
Stored inside `scripts/` (already mounted read-only at `/app/scripts:ro` in `docker-compose.yml` `seed` service — zero Compose volume changes!).
Each entry is keyed by `f"{domain_id}:{source_question_no}"` (e.g., `"domain5:506"`) and stores:
- `domain_id`, `source_question_no`, `en_sha256` (hash of canonical English `question_text` for drift detection)
- `th`: `{ "question_text": "...", "options": { "A": "...", ... }, "explanation": "..." }`

### 3.4 Authoring & Deterministic Validation Script (`scripts/translate_questions_th.py`)
Supports:
- `--pilot`: translates 10 representative questions (2 per domain) to stdout/report for Checkpoint 1 review.
- `--all`: resumable batch translation of all 540 questions using configured `LLM_API_KEY` / `LLM_BASE_URL` / `LLM_MODEL` from `.env` (writing checkpoints every batch so interruptions resume cleanly).
- `--validate-only`: strict zero-network deterministic CI gate asserting ALL of the following for **540/540** questions:
  1. Exact 540-question coverage matching `parsed_questions.json` (`en_sha256` matches).
  2. `th.question_text` and `th.explanation` are non-empty and contain Thai Unicode (`[\u0E00-\u0E7F]`).
  3. `th.options` has the **exact same keys** (`A`, `B`, `C`, `D`) as canonical EN `options`, and every option string is non-empty (numeric/port/acronym-only options like `"443"` or `"$100,000"` may remain identical to EN!).
  4. `th.explanation` starts with `คำตอบที่ถูกต้อง: <correct_answer>.` matching canonical `correct_answer` verbatim (zero answer-letter drift!) and contains `ตัวเลือก <K> ไม่ถูกต้อง` markers for every distractor key so `parseExamExplanation` parses 540/540 cleanly.
  5. Canonical English `parsed_questions.json` contains **0 Thai Unicode characters**.

### 3.5 Seeding (`scripts/load_exams_to_db.py`)
- Reads `scripts/question_translations_th.json` (adjacent to the script via `Path(__file__).resolve().parent / "question_translations_th.json"`, working identically on host and inside Docker `seed` where `working_dir = /work`).
- Runs validation checks before committing; upserts `translations = json.dumps({"th": entry["th"]})` onto each matched `questions` row while keeping `question_id` UUIDs and canonical English columns intact.

---

## 4. Backend API & Safety Contract

### 4.1 Model & DTO Updates (`models/exam.go`, `dto/exam.go`, `dto/learn.go`)
- Add `Translations any` (`gorm:"column:translations;type:jsonb;serializer:json" json:"translations,omitempty"`) to `models.Question`.
- Add `Translations any` (`json:"translations,omitempty"`) to:
  - `dto.QuestionDTO` (served by `POST /api/exams/pre-test`, `POST /api/exams/post-test`, `GET /api/exams/session`, `GET /api/exams/sample`)
  - `dto.DetailedExamResultDTO` (served by `POST /api/exams/sessions/:sessionId/finalize`, `GET /api/exams/sessions/:sessionId/review`, `POST /api/exams/submit-answer`)
  - `dto.LearnSampleQuestionDTO` & `dto.AttemptSampleQuestionResponse` (served by `GET /api/learn/subtopics/:subTopicId` and `POST /api/learn/subtopics/:subTopicId/sample-questions/:questionId/attempt`)
- **Strict Pre-Submission Sanitization Helper (`services.SanitizeQuestionTranslations(raw any, includeExplanation bool) map[string]any`)**:
  - Pre-submission (`ExamPage` active session, unattempted `LearnSampleQuestion`): **MUST strip `explanation`** from `translations.th` (returning only `question_text` and `options`) so learners can never inspect answers via DevTools!
  - Post-submission (`ExamReviewPage`, `FinalizeExamSession`, `AttemptSampleQuestion`, already-attempted `LearnSampleQuestion`, `GetSampleQuestion` on Landing): includes full `translations.th` including `explanation`.

### 4.2 NFR Safety Boundary Alignment (`services/nfr_evaluator.go` & `controllers/nfr_controller.go`)
- Keep `ThaiViolationInExams` checking **canonical English columns only** (`question_text`, `options`, `explanation` — NOT `translations`!) and `ThaiViolationInLogs` checking `scenarios.initial_logs` & `scenario_snapshots.initial_logs`.
- Update the descriptive Target label on `english_only_invariant` check in `services/nfr_evaluator.go` to `"0 Thai chars in canonical EN Exams & SIEM initial_logs"` so the NFR Safety dimension remains `PASS` while clearly documenting the revised invariant.

---

## 5. Frontend Specification

### 5.1 Shared Localization & Parser Utilities
- **`frontend/src/lib/questionLocale.ts`** (pure TS, unit-tested via `node --test`):
  - `resolveLocalizedQuestion(q, locale, forceOriginalEn)`: returns `{ questionText, options, explanation, hasThaiTranslation }`.
  - Gracefully falls back to canonical English if `translations?.th` is missing or `forceOriginalEn === true`.
- **`frontend/src/lib/examReview.ts`**:
  - Extend `parseExamExplanation(explanation, correctAnswer)` to deterministically parse both:
    - English: `Correct Answer: X. ... Option Y is incorrect. ...`
    - Thai: `คำตอบที่ถูกต้อง: X. ... ตัวเลือก Y ไม่ถูกต้อง ...`
  - Unit-test both EN and TH parsing (including `Your pick` distractor pinning and graceful raw fallback).

### 5.2 UI Behavior Across All 4 Question Surfaces
1. **`ExamPage.tsx` (Pre-Test & Post-Test)**:
   - Displays localized Thai when `locale === "th"` and English when `locale === "en"`.
   - Adds a compact button in the question header when `locale === "th"` and `hasThaiTranslation`: **"ดูต้นฉบับ EN"** ↔ **"กลับเป็นภาษาไทย"** (per-question toggle state reset when moving to next/previous question, never touching selected answers).
   - Replaces the old "Exam questions are English-only" alert banner with a compact helpful note: *"แสดงคำแปลภาษาไทยควบคู่ศัพท์เทคนิค — กดปุ่ม 'ดูต้นฉบับ EN' เพื่อเทียบภาษาอังกฤษได้ทุกข้อ"*.
2. **`ExamReviewPage.tsx`**:
   - Displays Thai question stems, options, and structured **Distractor Breakdown** in `TH` mode, with a per-card **"ดูต้นฉบับ EN"** toggle.
3. **`LearnTheoryPage.tsx` (Optional Deep Dive Sample Questions)**:
   - Displays Thai question stems, options, and post-attempt explanations in `TH` mode with per-question **"ดูต้นฉบับ EN"** toggle.
4. **`LandingPage.tsx` (Sample Exam Tab)**:
   - Displays Thai question stems, options, and post-submit explanation in `TH` mode with **"ดูต้นฉบับ EN"** toggle.

---

## 6. Verification & Acceptance Criteria

- [ ] Checkpoint 1 approved: 10 sample Thai translations (2 per domain) approved for high-school readability and Hybrid Bilingual terminology before translating remaining 530 questions.
- [ ] `python scripts/translate_questions_th.py --validate-only` passes **540/540** with 0 errors.
- [ ] `scripts/load_exams_to_db.py` seeds all 540 `translations` idempotently preserving existing `question_id` UUIDs.
- [ ] Go unit tests verify `SanitizeQuestionTranslations` omits `explanation` pre-submission and includes it post-submission, plus `Zero-Overlap` and NFR Safety `PASS`.
- [ ] Frontend unit tests verify `questionLocale.ts`, Thai + English `parseExamExplanation` in `examReview.test.ts`, `i18n.test.ts`, `typography.test.ts`, `npm run tsc`, and `npm run build`.
- [ ] `scripts/test_integration.py` verifies `translations.th` presence on Pre/Post/Review endpoints and absence of `explanation` leak during active exams.
- [ ] Swagger docs regenerated cleanly.
