# Workflow Spec: Post-Exam Q&A Review & Learning-Aligned Feedback (Issue #14)

**GitHub Issue:** [#14 — show question&answer tests after finish pre-test and post-test](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/14)  
**Status:** `/loop-me` Grilling Rounds 1–2 Settled (`Q1`–`Q10` Resolved, 2026-11-25) — Ready for User Confirmation & Implementation  
**Depends On:**
- [learning_subtopics_sample_questions.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/learning_subtopics_sample_questions.md) (`questions.sub_topic_id`, `/learn?domain=...&topic=...` deep links, `Pre ∪ Learn ∪ Post = ∅` Zero-Overlap policy)
- [language_switching_and_localization.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/language_switching_and_localization.md) (Bilingual `en`/`th` UI chrome + strictly English-only exam questions/options/explanations)
- [ui_typography_readability.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/ui_typography_readability.md) (`rem`-only Semantic Typography Scale, `12px` `text-xs` hard floor, `frontend/tests/typography.test.ts`)

---

## Objective

Close the assessment feedback loop after both **Pre-Test** and **Post-Test** so learners immediately receive a refresh-safe, **learning-aligned Q&A Review** instead of being silently redirected away from their exam answers. Every completed exam session becomes a structured diagnostic and mastery study artifact backed by existing PostgreSQL data (**Zero-DDL — no `init.sql` changes required**):

- **Immediate Post-Submission Redirect**: Completing either Pre-Test or Post-Test navigates directly to `/exams/review/:sessionId`.
- **Reliable Finalization & Time Tracking**: Submitting with unanswered items or hitting `00:00:00` explicitly finalizes the session (`POST /api/exams/sessions/:sessionId/finalize`), persists real `time_spent_seconds`, marks blank answers as `unanswered` (`is_correct = false`), and prevents stalled `in_progress` sessions.
- **Learning-Aligned Review Presentation**: Groups all 30 questions by CompTIA SY0-701 Security Domain by default (with toggle to Exam Order `1–30`), auto-expands incorrect/unanswered cards while keeping correct cards compact, parses the 540/540 structured explanations into **Why `<Correct>` is Correct** + **Distractor Breakdown** (highlighting the learner's mistaken pick first), and deep-links every question directly to `/learn?domain=<domain_id>&topic=<sub_topic_id>`.
- **Pre-Test → Post-Test Learning Gain**: Post-Test review displays an overall and 5-domain learning gain comparison against the learner's latest completed Pre-Test while strictly preserving **Proficiency Invariance** (`user_skill_profiles.proficiency_score` is never overwritten by Post-Test completion).
- **Persistent Revisit Loop**: Learners can revisit any completed Pre-Test or Post-Test Q&A review at any time from `/evaluation` and `/profile`.

The two loops this workflow makes real:
1. **Exam Completion & Immediate Diagnostic Loop**: Learner submits exam (all 30 answered, confirmed partial submission, or `00:00:00` auto-finalize) → backend finalizes atomically & records elapsed seconds → frontend navigates to `/exams/review/:sessionId` → learner inspects mistakes grouped by Security Domain, reads structured rationale + distractor breakdown, and jumps directly to weak SY0-701 Sub-topics in `/learn`.
2. **Longitudinal Study & Post-Test Mastery Loop**: Learner revisits completed Pre-Test or Post-Test Q&A reviews from `/evaluation` or `/profile` via `GET /api/exams/history` and `GET /api/exams/sessions/:sessionId/review`, comparing Pre-Test diagnostic weaknesses against Post-Test mastery gains while backend guards strictly block access to `in_progress` or foreign sessions.

---

## 1. Loop Profile

- **Loop Name**: Post-Exam Q&A Review & Learning Alignment Loop
- **Frequency / Cadence**:
  - Runtime: triggered once on every Pre-Test / Post-Test completion (manual submit or timeout auto-finalize), and on demand whenever a learner revisits exam reviews from `/evaluation` or `/profile`.
  - Verification: executed on every `go test`, `npm test`, `npm run build`, and `scripts/test_integration.py` run.
- **Primary Actors**:
  - **Backend Exam API (`controllers/exam_controller.go`, `dto/exam.go`, `main.go`)**: Adds `POST /api/exams/sessions/:sessionId/finalize`, `GET /api/exams/history`, and `GET /api/exams/sessions/:sessionId/review`; upgrades `FinalizeExamSession` to order results by `order_index ASC` and compute deterministic `domain_scores` for both `pre` and `post` without mutating `user_skill_profiles` on `post`.
  - **Exam Submission & Timer Client (`frontend/src/pages/ExamPage.tsx`)**: Tracks real elapsed seconds, handles partial submission & single-shot `00:00:00` auto-finalize, and routes straight to `/exams/review/:sessionId`.
  - **Explanation Parser Utility (`frontend/src/utils/examExplanation.ts`)**: Deterministically parses `"Correct Answer: X. ... Option Y is incorrect. ..."` into structured rationale + per-option distractor notes with graceful raw-text fallback.
  - **Exam Review Page (`frontend/src/pages/ExamReviewPage.tsx`)**: Renders summary metrics, Post-Test Pre→Post Learning Gain card, Domain Breakdown, color-coded `01–30` Jump Grid, Filter & Grouping Toolbar, Smart-Expanded Question Cards, and next-step CTAs.
  - **Revisit Surfaces (`frontend/src/pages/EvaluationPage.tsx`, `frontend/src/pages/ProfilePage.tsx`, `frontend/src/App.tsx`, `frontend/src/components/Layout.tsx`)**: Exposes persistent entry links to completed Pre-Test / Post-Test reviews and session history.
  - **Localization & Guard Suites (`frontend/src/i18n/en.ts`, `frontend/src/i18n/th.ts`, `frontend/src/i18n/errors.ts`, `controllers/exam_controller_test.go`, `frontend/tests/`, `scripts/test_integration.py`)**: Enforces 100% EN/TH UI parity, English-only exam content invariance, `rem`-only typography guard, and zero answer leakage.
- **Trigger**:
  - Runtime: event-driven on exam submission, `00:00:00` timer expiry, direct navigation to `/exams/review/:sessionId`, or revisit click from `/evaluation` / `/profile`.
- **Checkpoint**:
  - **Runtime**: None (100% autonomous).
  - **Implementation Sign-Off (1 Push-Right Visual Review Checkpoint — `Q10=A`)**: After all backend, frontend, i18n, unit, typography, and integration tests pass, present a single compact **Visual Review Brief** containing EN & TH screenshots of:
    1. Pre-Test Review page (`/exams/review/:preSessionId` — default Domain-grouped Smart-Expand view + expanded incorrect card with distractor highlight & `/learn` deep link)
    2. Post-Test Review page (`/exams/review/:postSessionId` — Pre→Post Learning Gain Card + Exam Order `1–30` view)
    3. Revisit entry surfaces (`/evaluation` Exam Answer Reviews card & `/profile` Exam History & Reviews section)
    Wait for user approval before closing Issue #14.

---

## 2. Settled Design Decisions (`Q1`–`Q10`)

| Q# | Decision | Settled Rule |
| --- | --- | --- |
| `Q1` | Post-Submission Routing & Revisit Entry Points | **Dedicated Page + Dual Revisit Links (`1A`)**: Redirect immediately to `/exams/review/:sessionId` after finishing Pre-Test or Post-Test; add revisit entry points on both `/evaluation` and `/profile`. |
| `Q2` | Learning Alignment & Card Anatomy | **Learning-Aligned Review Layout (`2A`)**: Top Summary Header + Domain Breakdown + Smart Filter/Grouping Toolbar + `01–30` color-coded Jump Grid + Question Cards with A–D visual diff, structured Explanation, SY0-701 Sub-topic badge, and direct `/learn?domain=...&topic=...` link. |
| `Q3` | Pre-Test vs. Post-Test Pedagogical Framing | **Immediate Full Q&A for Both + Post-Test Gain (`3A`)**: Full answers & explanations shown immediately after both Pre-Test (diagnostic framing, safe under `Pre ∪ Learn ∪ Post = ∅`) and Post-Test (mastery framing + Pre→Post Learning Gain comparison). |
| `Q4` | Unanswered / Timeout Submission Bug | **Explicit Finalize Endpoint (`4A`)**: Add `POST /api/exams/sessions/:sessionId/finalize` so partial and timed-out submissions finalize cleanly, mark blank items as `user_answer = ""` / `is_correct = false` ("Unanswered"), while preserving `POST /api/exams/submit-answer` backward compatibility. |
| `Q5` | Access Control & History Scope | **Learner Own Completed Sessions Only (`5A`)**: Strict `user_id` ownership check + `status = 'completed'` guard (`403 EXAM_NOT_COMPLETED` for `in_progress`, `404` for foreign/missing). Expose `GET /api/exams/history` and `GET /api/exams/sessions/:sessionId/review`. Specialist Console out of scope. |
| `Q6` | Default View, Grouping & Smart Expand | **Domain-Grouped + All + Smart Expand (`6A`)**: Default filter = `All`, default grouping = `Grouped by Security Domain` (`domain1`–`domain5`), Smart Expand auto-expands `Incorrect` & `Unanswered` cards while collapsing `Correct` cards to 1-line summaries (with per-card toggle + `Expand All / Collapse All` + switch to `Exam Order 1–30`). |
| `Q7` | Explanation Formatting | **Structured Parser + Graceful Fallback (`7A`)**: Split 540/540 explanations into `Why <Correct> is Correct` + `Why Other Options Are Incorrect` per-option list, highlighting & pinning the learner's wrong choice (`Your pick`) first; fall back cleanly to raw paragraph if a future item doesn't match the pattern. |
| `Q8` | Pre→Post Learning Gain & Proficiency Invariance | **Overall + 5-Domain Delta, Latest Completed Pre Baseline (`8A`)**: Compare Post-Test against the learner's latest completed Pre-Test (`ORDER BY completed_at DESC LIMIT 1`, matching `GetResearchSummary`) across overall % and all 5 domains. Compute `domain_scores` for both `pre` and `post`, but **never** overwrite `user_skill_profiles.proficiency_score` on `post`. |
| `Q9` | Timer & Auto-Finalize | **Real `time_spent_seconds` + Single-Shot Timeout Auto-Finalize (`9A`)**: Record real elapsed seconds on exam completion; when the 30-minute timer reaches `00:00:00`, auto-submit unsaved picks, call finalize, and route to `/exams/review/:sessionId` (guarded against double-trigger and stale sessions). |
| `Q10` | Entry Points, CTAs & Checkpoint | **`/evaluation` Card + `/profile` History + Context CTAs + 1 Visual Checkpoint (`10A`)**: Zero-DDL; full Go, Vitest, typography, and `test_integration.py` coverage + 1 Push-Right EN/TH Visual Review Brief before closing Issue #14. |

---

## 3. Database & Research Invariants (Zero-DDL)

No SQL migration or `init.sql` modification is required. All needed columns already exist:
- `exam_sessions`: `session_id`, `user_id`, `exam_type`, `status`, `score`, `total_questions`, `started_at`, `completed_at`, `time_spent_seconds`
- `exam_session_questions`: `session_id`, `question_id`, `order_index`, `user_answer`, `is_correct`
- `questions`: `question_id`, `domain_id`, `sub_topic_id`, `type`, `question_text`, `options`, `correct_answer`, `explanation`, `mitre_ref`

**Non-negotiable Invariants:**
1. **Zero Answer Leakage**: Neither `GET /api/exams/session` nor `GET /api/exams/sessions/:sessionId/review` may ever expose `correct_answer` or `explanation` while `session.status != 'completed'`.
2. **Proficiency Invariance**: `FinalizeExamSession` updates `user_skill_profiles` **strictly when `session.ExamType == "pre"`**. Finalizing a `post` session computes `domain_scores` for display/review only and never modifies `user_skill_profiles`.
3. **Zero-Overlap Preservation**: `BuildPostTestExclusionQuery` continues enforcing `Pre-Test ∪ Learn ∪ Post-Test = ∅` untouched.
4. **English-Only Exam Content**: `question_text`, `options`, `correct_answer`, and `explanation` remain strictly English in both `en` and `th` locales (with the subtle Thai notice banner when `locale === "th"`).

---

## 4. Backend API Specification (`controllers/exam_controller.go`, `dto/exam.go`, `main.go`)

### 4.1 Route Registration (`main.go`)
Add inside the JWT-protected `/api/exams` group (`middlewares.AuthMiddleware()`):
```go
exams.POST("/sessions/:sessionId/finalize", examController.FinalizeSession)
exams.GET("/history", examController.GetExamHistory)
exams.GET("/sessions/:sessionId/review", examController.GetExamReview)
```
Existing routes (`POST /pre-test`, `POST /post-test`, `GET /session`, `POST /submit-answer`) remain unchanged in path and backward-compatible in response shape.

### 4.2 Canonical Domain Order & Sub-Topic Title Resolution
- Domain scores are always returned in deterministic order: `domain1`, `domain2`, `domain3`, `domain4`, `domain5` (skipping a domain only if `total == 0`).
- Canonical English domain names reuse the standard CompTIA SY0-701 labels (`General Security Concepts`, `Threats, Vulnerabilities, and Mitigations`, `Security Architecture`, `Security Operations`, `Security Program Management and Oversight`), localized client-side via existing `resolveDomainName(domainId, domainName, locale)`.
- Each review item resolves `sub_topic_title` from `services.OrderedTheorySubtopics` (matching `question.SubTopicID`), yielding `""` if `sub_topic_id` is nil/unmapped—zero extra DB round-trips or hardcoded duplicates.
- Domain passing threshold reuses `PROFICIENCY_THRESHOLD` env var (default `70.0`).

### 4.3 `POST /api/exams/sessions/:sessionId/finalize`
- **Request Body (optional JSON)**:
  ```json
  { "time_spent_seconds": 842 }
  ```
- **Validation & Guards**:
  1. Extract authenticated `user_id` from JWT context (`401 AUTH_UNAUTHORIZED` if missing).
  2. Look up `exam_sessions` by `session_id = :sessionId AND user_id = userID`. If not found → `404` (`code: "EXAM_SESSION_NOT_FOUND"`).
  3. **Idempotency**: If `session.Status == "completed"`, do **not** re-write `completed_at` or `user_skill_profiles`; return `200 OK` with the finalized summary (`session_id`, `completed: true`, `score`, `correct_count`, `unanswered_count`, `domain_scores`, `results`).
  4. If `session.Status == "abandoned"` → `400` (`code: "EXAM_SESSION_NOT_FOUND"`).
  5. **Persist Elapsed Time**: If `input.TimeSpentSeconds > 0`, set `session.TimeSpentSeconds = max(existing, input.TimeSpentSeconds)` (clamped to `<= 7200`). If still nil/0, leave as existing or `0` (never negative).
  6. **Finalize Transaction**:
     - Update all `exam_session_questions` rows for `session_id` where `user_answer IS NULL OR user_answer = ''` to `user_answer = ''`, `is_correct = false`.
     - Execute shared `FinalizeExamSession(session, userID)`:
       - Count `correct_count` (`is_correct = true`).
       - Compute `score = math.Round((float64(correctCount) / float64(max(1, session.TotalQuestions))) * 100 * 100) / 100`.
       - Set `session.Status = "completed"`, `session.Score = &score`, `session.CompletedAt = &now`, save `session`.
       - Compute `domain_scores` across `domain1`..`domain5` for **both** `pre` and `post`.
       - **Only if `session.ExamType == "pre"`**: upsert `user_skill_profiles` for each domain with `proficiency_score = domain.Percentage`.
       - Query `detailed_results` joined with `questions` **explicitly ordered by `exam_session_questions.order_index ASC`**, selecting `exam_session_questions.order_index`, `user_answer`, `is_correct`.
- **Success Response (`200 OK`)**:
  ```json
  {
    "session_id": "uuid",
    "exam_type": "pre",
    "message": "Exam completed successfully!",
    "completed": true,
    "score": 73.33,
    "correct_count": 22,
    "unanswered_count": 2,
    "total_questions": 30,
    "time_spent_seconds": 842,
    "domain_scores": [ ... ],
    "results": [ ... ]
  }
  ```

### 4.4 `POST /api/exams/submit-answer` (Backward-Compatible Upgrades)
- Keep all existing fields (`completed`, `message`, `score`, `results`, `correct_count`, `domain_scores`) so `scripts/test_integration.py` passes without breakage.
- Additive upgrades when `unansweredCount == 0`:
  - Include `session_id: session.SessionID` and `domain_scores` for **both** `pre` and `post`.
  - Return `results` ordered by `order_index ASC` with `order_index` populated.

### 4.5 `GET /api/exams/history`
- Queries all `exam_sessions` for authenticated `user_id` where `status = 'completed'`, ordered by `completed_at DESC, started_at DESC`.
- Aggregates `correct_count` and `unanswered_count` per session in a single grouped SQL query (no N+1).
- **Success Response (`200 OK`)**:
  ```json
  {
    "latest_pre_session_id": "uuid-or-null",
    "latest_post_session_id": "uuid-or-null",
    "sessions": [
      {
        "session_id": "uuid",
        "exam_type": "post",
        "status": "completed",
        "score": 86.67,
        "correct_count": 26,
        "incorrect_count": 4,
        "unanswered_count": 0,
        "total_questions": 30,
        "started_at": "2026-11-25T00:00:00Z",
        "completed_at": "2026-11-25T00:18:12Z",
        "time_spent_seconds": 1092
      }
    ]
  }
  ```

### 4.6 `GET /api/exams/sessions/:sessionId/review`
- **Guards**:
  1. `401 AUTH_UNAUTHORIZED` if unauthenticated.
  2. Query `exam_sessions` where `session_id = :sessionId AND user_id = userID`. Missing/foreign → `404` (`code: "EXAM_SESSION_NOT_FOUND"`).
  3. If `session.Status != "completed"` → `403` (`code: "EXAM_NOT_COMPLETED"`, `error: "Exam session is not completed yet. Complete the exam before reviewing answers."`, `exam_type: session.ExamType`).
- **Pre-Test Comparison (`pre_comparison`)**:
  - Computed **only when `session.ExamType == "post"`**.
  - Finds user's latest completed `pre` session (`WHERE user_id = ? AND exam_type = 'pre' AND status = 'completed' ORDER BY completed_at DESC LIMIT 1`), matching `GetResearchSummary`.
  - If found, computes `pre_score`, `post_score`, `score_delta = round((post - pre) * 100) / 100`, `pre_correct_count`, `post_correct_count`, and `domain_deltas` across `domain1`..`domain5` (`pre_percentage`, `post_percentage`, `delta`).
  - If no completed `pre` session exists (edge case in tests), `pre_comparison` is `null`.
- **Success Response (`200 OK`)**:
  ```json
  {
    "session": {
      "session_id": "uuid",
      "exam_type": "post",
      "status": "completed",
      "score": 86.67,
      "correct_count": 26,
      "incorrect_count": 3,
      "unanswered_count": 1,
      "total_questions": 30,
      "started_at": "2026-11-25T00:00:00Z",
      "completed_at": "2026-11-25T00:18:12Z",
      "time_spent_seconds": 1092
    },
    "domain_scores": [
      {
        "domain_id": "domain1",
        "domain_name": "General Security Concepts",
        "total": 4,
        "correct": 4,
        "incorrect": 0,
        "unanswered": 0,
        "percentage": 100.0,
        "threshold": 70.0,
        "passed": true
      }
    ],
    "pre_comparison": {
      "pre_session_id": "uuid",
      "pre_score": 60.0,
      "post_score": 86.67,
      "score_delta": 26.67,
      "pre_correct_count": 18,
      "post_correct_count": 26,
      "domain_deltas": [
        {
          "domain_id": "domain1",
          "domain_name": "General Security Concepts",
          "pre_percentage": 50.0,
          "post_percentage": 100.0,
          "delta": 50.0
        }
      ]
    },
    "results": [
      {
        "order_index": 0,
        "question_id": "uuid",
        "domain_id": "domain4",
        "domain_name": "Security Operations",
        "sub_topic_id": "4.1",
        "sub_topic_title": "Apply common security techniques to computing resources",
        "type": "multiple_choice",
        "question_text": "XYZ Corp is designing...",
        "options": { "A": "21", "B": "80", "C": "443", "D": "25" },
        "user_answer": "B",
        "correct_answer": "C",
        "is_correct": false,
        "is_unanswered": false,
        "explanation": "Correct Answer: C. 443. ...",
        "mitre_ref": null
      }
    ]
  }
  ```

### 4.7 `ExamStore` Interface Extension
Extend `ExamStore` in `controllers/exam_controller.go` with clean mockable methods so `controllers/exam_controller_test.go` can test every new handler without requiring a live PostgreSQL instance:
- `FindSessionByIDAndUser(sessionID, userID string) (*models.ExamSession, error)`
- `MarkUnansweredAndFinalize(session *models.ExamSession, userID string) (int64, int64, []ExamDomainScore, []ExamDetailedResult, error)`
- `GetCompletedExamHistory(userID string) ([]dto.ExamHistoryItemDTO, error)`
- `GetExamSessionReviewData(session *models.ExamSession) (int64, int64, []ExamDomainScore, []ExamDetailedResult, error)`
- `GetLatestCompletedPreTestSession(userID string) (*models.ExamSession, error)`

---

## 5. Frontend Specification

### 5.1 Deterministic Explanation Parser (`frontend/src/utils/examExplanation.ts`)
Pure TypeScript utility with unit test coverage (`frontend/tests/examExplanation.test.ts`):
```ts
export interface DistractorExplanation {
  option: string; // "A" | "B" | "C" | "D"
  rationale: string;
}

export type ParsedExplanation =
  | {
      structured: true;
      correctOption: string;
      correctRationale: string;
      distractors: DistractorExplanation[];
    }
  | {
      structured: false;
      raw: string;
    };
```
**Parsing Rules:**
1. Trim input; if empty, return `{ structured: false, raw: "" }`.
2. Match distractor markers `Option\s+([A-D])\s+is\s+incorrect\.?\s*` globally (case-insensitive).
3. If at least 1 distractor marker is found:
   - Slice the prefix before the first `Option [A-D] is incorrect`.
   - Strip leading `Correct Answer:\s*([A-D])\.?\s*` prefix (capturing `correctOption`, defaulting to `correctAnswer` parameter).
   - Extract each distractor `{ option, rationale }` from the slice between consecutive markers.
   - Verify `correctRationale.length > 0` and every distractor has non-empty `rationale`.
   - Return `{ structured: true, correctOption, correctRationale, distractors }`.
4. Otherwise return `{ structured: false, raw: explanation.trim() }`.
5. **Distractor Ordering Helper (`orderDistractorsForLearner(distractors, userAnswer)`)**:
   - When the learner answered wrongly (`userAnswer` matches a distractor's `option`), place that option **first** (with a red `Your pick` badge in the UI) followed by the remaining distractors in `A → D` order.

### 5.2 Exam Submission & Timeout Auto-Finalize (`frontend/src/pages/ExamPage.tsx`)
1. **Elapsed Time Tracking**:
   - Define `const EXAM_DURATION_SECONDS = 30 * 60;`.
   - Keep `const previousTimeSpent = session?.time_spent_seconds ?? 0;` and `const currentRunElapsed = Math.max(1, EXAM_DURATION_SECONDS - timeLeft);` so `totalTimeSpent = previousTimeSpent + currentRunElapsed`.
   - Reset `timeLeft` to `EXAM_DURATION_SECONDS` when starting/loading a session and reset `autoFinalizeTriggeredRef.current = false`.
2. **Submit Flow (`submitAndFinalize(skipConfirm: boolean)`)**:
   - Guard against concurrent calls (`if (!session || submitting) return;`).
   - If `!skipConfirm` and `unanswered > 0`, show existing `window.confirm(...)`.
   - Set `submitting = true`.
   - Loop through `questions` where `answers[q.question_id]` is set and differs from `q.user_answer`, calling `POST /api/exams/submit-answer` (pass `time_spent_seconds: currentRunElapsed` on the final unsaved answer if `unanswered === 0`, `0` on earlier calls).
   - Track whether any `submit-answer` response returned `completed === true`.
   - If `!alreadyCompleted` (e.g., unanswered items remain, timeout occurred, or answers were already saved before refresh), call `POST /api/exams/sessions/${session.session_id}/finalize` with `{ time_spent_seconds: totalTimeSpent }`.
   - Navigate immediately to ``navigate(`/exams/review/${session.session_id}`, { replace: true })``.
3. **Single-Shot Timeout Auto-Finalize**:
   - Track `hasTickedRef` (becomes `true` once the interval ticks down from `> 0`) and `autoFinalizeTriggeredRef`.
   - When `timeLeft === 0 && session && !loading && !submitting && !autoFinalizeTriggeredRef.current`:
     - Set `autoFinalizeTriggeredRef.current = true`.
     - Display the timeout banner (`exam.autoSubmittingNotice`) and invoke `submitAndFinalize(true)` automatically.

### 5.3 Dedicated Exam Review Page (`frontend/src/pages/ExamReviewPage.tsx`)
Route: `/exams/review/:sessionId` (registered in `frontend/src/App.tsx`; `Layout.tsx` marks `Exam` nav link active on `/exams/review/*`).

**Page Structure (Top to Bottom — Graphite Dark Theme + Semantic Typography Scale):**
1. **Error / Guard States**:
   - `401` → redirect `/login`.
   - `403 EXAM_NOT_COMPLETED` → clean centered card explaining the exam is still in progress, with CTA button to resume `/pre-test` or `/post-test`.
   - `404` / network error → clean error card with CTA back to `/evaluation`.
2. **Header & Exam Summary Banner**:
   - Back link (`/evaluation`) + Exam Type badge (`PRE-TEST REVIEW` / `POST-TEST REVIEW`) + Completed timestamp (`formatDate`) + Time Spent (`mm:ss`).
   - Contextual framing subtitle:
     - Pre-Test: Diagnostic focus copy (*"Review your mistakes by CompTIA Security+ domain and jump straight to the matching sub-topics in the Learning Hub."*).
     - Post-Test: Mastery & growth copy (*"Review your Post-Test performance and compare your growth against your Pre-Test baseline."*).
   - 4 concise stat tiles: **Overall Score (`XX.X%`)**, **Correct (`N / 30`)**, **Incorrect (`M`)**, **Unanswered (`U`)** (or **Time Spent** when `U === 0`, showing both cleanly).
   - If `locale === "th"`, render the subtle English-only exam content notice banner (`exam.englishOnlyNotice`).
3. **Pre-Test → Post-Test Learning Gain Card (Post-Test Only, when `pre_comparison != null`)**:
   - Prominent overall comparison row: `Pre-Test XX.X%` → `Post-Test YY.Y%` + color-coded Delta Pill (`+26.7%` emerald if `> 0`, `0.0%` neutral, `-X.X%` amber/red if `< 0`).
   - Compact 5-Domain Delta Grid/Table: each row shows localized Domain Name (`resolveDomainName`), `Pre %`, `Post %`, visual dual/delta indicator, and signed `Δ %`.
4. **Domain Score Breakdown**:
   - Reuses `<DomainBreakdown />` (passing `domain_scores` mapped to `DomainStatus` with `studyDomainHref={(id) => \`/learn?domain=\${id}\`}` and showing `correct/total` alignment).
5. **Compact Question Jump Grid (`01`–`30`)**:
   - Single compact panel with 30 numbered monospace buttons (`text-xs font-mono`):
     - Emerald tint = Correct
     - Red tint = Incorrect
     - Amber/Graphite-dashed tint = Unanswered
   - Legend (`Correct`, `Incorrect`, `Unanswered`).
   - Clicking item `N`:
     - If current status/domain filter hides question `N`, resets the conflicting filter to `all` so the target card is guaranteed visible.
     - Ensures question `N` is added to `expandedIds`.
     - Smooth-scrolls to `#review-q-${orderIndex}` with brief focus ring highlight.
6. **Sticky / Prominent Filter & Alignment Toolbar**:
   - **Status Filter Pills** (with live counts): `All (30)` | `Incorrect (N)` | `Unanswered (U)` *(only shown if U > 0)* | `Correct (K)`.
   - **Domain Filter Select / Pills**: `All Domains` | `Domain 1`..`Domain 5`.
   - **Alignment Mode Segmented Toggle**: `By Domain` *(default)* | `Exam Order (1–30)`.
   - **Expand/Collapse All Button**: toggles expanding/collapsing all currently visible cards.
7. **Question Review List**:
   - **Default State on Load**:
     - `statusFilter = "all"`, `domainFilter = "all"`, `groupMode = "domain"`.
     - `expandedIds` initialized to **all `question_id`s where `!is_correct` (i.e., both `Incorrect` and `Unanswered`)**; `Correct` cards start collapsed.
   - **In `By Domain` Mode**:
     - Groups visible questions under `Domain 1` → `Domain 5` headers showing localized domain name, `correct / total` score pill, weak-domain indicator (`< 70%` amber badge), and quick `Study Domain →` link (`/learn?domain=<id>`).
   - **Question Review Card Anatomy**:
     - **Collapsed Row (for `Correct` cards by default)**: clickable full-width header displaying `#NN`, Sub-topic pill (`4.1`), Emerald `Correct` pill, 1-line truncated question stem (`text-sm`), and `Show Explanation` chevron.
     - **Expanded Card**:
       - Header: `#NN` (`font-mono`), Domain badge (in Exam Order mode), SY0-701 Sub-topic badge (`4.1 · Apply common security techniques...`), Status badge (`Correct` emerald / `Incorrect` red / `Unanswered` amber), and collapse toggle button.
       - Question Stem (`text-base text-graphite-100 leading-relaxed`).
       - Unanswered callout strip (if `is_unanswered`): subtle amber notice (*"You did not answer this question before submitting."*).
       - **4 Options Comparison (`A`, `B`, `C`, `D` sorted alphabetically)**:
         - Correct answer row: Emerald border/tint + `CheckCircle2` icon + `Correct Answer` badge (and `Your Answer` badge too if learner picked it).
         - Learner's wrong answer row: Red border/tint + `XCircle` icon + `Your Answer` badge.
         - Other unselected options: subtle graphite border/muted surface (`text-graphite-300`).
         - All option text rendered at `text-base leading-relaxed`.
       - **Structured Explanation Box**:
         - Parsed via `parseExamExplanation(q.explanation, q.correct_answer)`:
           - Section 1 (`Why Option <X> is Correct`): Emerald accent header + rationale prose (`text-base leading-relaxed`).
           - Section 2 (`Why Other Options Are Incorrect`): Ordered list of distractors via `orderDistractorsForLearner`; if a distractor matches `q.user_answer`, render with red left accent + `Your pick` badge at top of the distractor list; other distractors rendered cleanly below (`text-base leading-relaxed`).
         - Graceful fallback: if `!parsed.structured`, render single clean Explanation box with `parsed.raw`.
       - **Card Learning Footer**:
         - Direct button/link: **`Study Sub-Topic <sub_topic_id> in Learning Hub →`** linking to `/learn?domain=${q.domain_id}&topic=${q.sub_topic_id}` (or `/learn?domain=${q.domain_id}` if `sub_topic_id` is absent).
8. **Bottom Next-Step Action Bar**:
   - **Pre-Test**:
     - Primary CTA: `Go to Evaluation & Skill Gap` → `/evaluation`
     - Secondary CTA: `Go to Learning Hub` → `/learn`
   - **Post-Test**:
     - Primary CTA: `Back to Evaluation Summary` → `/evaluation`
     - Secondary CTA: `View Profile & Exam History` → `/profile`

### 5.4 Revisit Entry Surfaces (`/evaluation` and `/profile`)
1. **`EvaluationPage.tsx`**:
   - Fetch `GET /api/exams/history` alongside `/api/simulation/status` (fail-open so a history error never blocks `/evaluation`).
   - When `latest_pre_session_id` or `latest_post_session_id` exists, render an **"Exam Answer Reviews"** card above or below the Skill Gap card showing:
     - **Pre-Test Q&A Review** card/button (displays score `%`, `correct/30`, completed date, and `Review Q&A →` link to `/exams/review/${latestPre.session_id}`).
     - **Post-Test Q&A Review** card/button (shown when `latest_post_session_id` exists, displays score `%`, `correct/30`, delta vs Pre-Test if both exist, and `Review Q&A →` link to `/exams/review/${latestPost.session_id}`).
2. **`ProfilePage.tsx`**:
   - Fetch `GET /api/exams/history` on mount (fail-open).
   - Render an **"Exam History & Reviews"** section showing all completed Pre-Test and Post-Test attempts in reverse chronological order:
     - Columns/rows: Exam Type badge (`Pre-Test` / `Post-Test`), Completed Date (`formatDate`), Score (`XX.X%` & `correct/30`), Time Spent (`mm:ss`), and `Review Q&A →` button linking to `/exams/review/${item.session_id}`.
     - Empty state when no exams are completed yet, with a subtle link to `/pre-test`.

### 5.5 Localization (`en.ts`, `th.ts`, `errors.ts`)
- Add `examReview.*` and `profile.examHistory.*` / `evaluation.examReviews.*` keys with 100% EN/TH key & `{{param}}` parity.
- Register `EXAM_NOT_COMPLETED` in `frontend/src/i18n/errors.ts` (EN & TH).
- Add `exam.autoSubmittingNotice` for timeout auto-finalization.
- Keep `frontend/tests/i18n.test.ts` (or existing i18n parity check) and `frontend/tests/typography.test.ts` green.

---

## 6. Verification Plan & Acceptance Criteria

### 6.1 Automated Checks
1. **Go Backend Compilation & Unit Tests**:
   - `go build -o /dev/null main.go` (or Windows `go build -o NUL main.go`)
   - `go test ./...` including new tests in `controllers/exam_controller_test.go`:
     - `FinalizeSession`: finalizes partial exam (marks blank items unanswered + `is_correct = false`), records `time_spent_seconds`, is idempotent on repeat calls, updates `user_skill_profiles` on `pre` only (Proficiency Invariance on `post`).
     - `GetExamHistory`: returns user's completed sessions only, ordered `DESC`, with `latest_pre_session_id` and `latest_post_session_id`.
     - `GetExamReview`: returns `200` with ordered `order_index ASC` results, `domain_scores`, `sub_topic_title`, and `pre_comparison` on `post`; returns `403 EXAM_NOT_COMPLETED` on `in_progress`; returns `404` on foreign or non-existent session.
2. **Frontend Typecheck, Unit Tests, Typography Guard & Build**:
   - `cd frontend && npm test`:
     - New `frontend/tests/examExplanation.test.ts`: validates parser on real structured explanations (including all 540 items from `parsed_questions.json` or representative corpus + malformed fallback + learner wrong-pick ordering).
     - Existing `frontend/tests/typography.test.ts`: confirms zero arbitrary `text-[...px]` and zero sub-`12px` classes across all new/modified TSX files.
     - Existing i18n parity tests: confirms 100% EN/TH key parity.
   - `cd frontend && npx tsc -b && npm run build`
3. **End-to-End Integration Harness (`scripts/test_integration.py`)**:
   - While Pre-Test is `in_progress`, assert `GET /api/exams/sessions/{pre_id}/review` returns `403` (`EXAM_NOT_COMPLETED`).
   - After Pre-Test completes, assert `GET /api/exams/history` includes `pre_id` (`latest_pre_session_id`), and `GET /api/exams/sessions/{pre_id}/review` returns `200` with 30 results ordered `0..29`, non-empty `explanation`, `correct_answer`, `sub_topic_id`, and 5 `domain_scores`.
   - Test partial-exam finalize flow on a dedicated test user (answer 28/30, call `POST /api/exams/sessions/{id}/finalize`, verify `completed == true`, `unanswered_count == 2`, and review shows 2 `is_unanswered == true` items).
   - After Post-Test completes, assert `GET /api/exams/sessions/{post_id}/review` returns `200` with `pre_comparison` populated (`pre_score`, `post_score`, `score_delta`, 5 `domain_deltas`) and zero overlap with Pre-Test/Learn.

### 6.2 Push-Right Human Checkpoint (`Q10=A`)
Present a single **Visual Review Brief** after all automated tests pass, containing EN & TH screenshots of Pre-Test Review, Post-Test Review (+ Learning Gain Card), and `/evaluation` + `/profile` entry surfaces, plus test pass summary. Close Issue #14 upon user sign-off.
