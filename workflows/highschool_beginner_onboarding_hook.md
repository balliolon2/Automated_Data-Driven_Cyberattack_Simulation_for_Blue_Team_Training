# Workflow Spec: High School Beginner Onboarding, Interactive Hook & Runbook (Issue #72)

**GitHub Issue:** [#72 — feat(ux): beginner-friendly onboarding, interactive hook, and presentation plan for high school learners](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/72)  
**Status:** `/loop-me` Grilling Rounds 1–3 Settled (`Q6`, `Q11`, `Q13`, `Q15` Resolved, 2026-11-27) — Ready for Implementation  
**Depends On:**
- `workflows/ui_typography_readability.md` (`rem`-only Semantic Typography Scale, `12px` `text-xs` hard floor, `frontend/tests/typography.test.ts`).
- `workflows/language_switching_and_localization.md` (Bilingual `EN | TH` dictionaries with 100% key/param parity).
- `workflows/compact_learning_and_two_hour_budget.md` (`#70`) and `workflows/bilingual_exam_questions_th.md` (`#71`) for accurate 2-hour roadmap numbers and bilingual sample quiz.

**Explicit Scope Boundary (`Q6`, `Q15`):**
- **Out of Scope**: Building an HTML slide deck (the presenter is preparing their own slides).
- **In Scope**: 4 high-impact in-app beginner UX upgrades + 1 concise Markdown Presenter Hook & 2-Hour Runbook (`docs/HIGHSCHOOL_2HR_SESSION_RUNBOOK.md`) to feed the user's slides and classroom facilitation.

---

## Objective

Make the first 10 minutes feel exciting, welcoming, and crystal-clear for Thai high school students with **zero prior cybersecurity knowledge** so they want to play as a Junior Cyber Detective / SOC Tier 1 Analyst and never stall out on unfamiliar jargon or exam anxiety.

---

## 1. Loop Profile

- **Loop Name**: Novice First-Impression, Guided First-Mission & Facilitator Hook Loop
- **Trigger**:
  - Guest visits `/` (`LandingPage`) → starts Pre-Test (`ExamPage`) → enters 1st Simulation Session (`SimulationPage`).
  - Presenter opens `docs/HIGHSCHOOL_2HR_SESSION_RUNBOOK.md` before Tuesday Dec 1 afternoon.
- **Checkpoint**:
  - **Checkpoint 2 (Pre-Flight Visual Sign-Off — `Q13`)**: Review EN & TH screenshots of the new Landing Page Detective Preview + Mission Map, Pre-Test Reassurance Banner, and First-Scenario Quick-Guide/Glossary Drawer before closing `#72`.

---

## 2. Settled Design Decisions

| Q# | Decision | Settled Rule |
| --- | --- | --- |
| `Q6`, `Q15` | Presenter Deliverable | Do **not** generate an HTML slide deck. Deliver `docs/HIGHSCHOOL_2HR_SESSION_RUNBOOK.md` containing high-school hook talking points/analogies (ready to drop into user's slides), an interactive projector warm-up case, a minute-by-minute 120-minute schedule, and Azure classroom troubleshooting. |
| `Q11` | Default Guest Locale | Support `VITE_DEFAULT_LOCALE` (default `"th"` in `frontend/Dockerfile` & `docker-compose.yml`). When `localStorage.getItem("cybersim_lang")` is empty, default to `VITE_DEFAULT_LOCALE`; explicit `EN | TH` clicks and logged-in `users.preferred_language` still take full precedence. |
| `Q11` | Landing Interactive Hook | Reframe `LandingPage.tsx` Hero & Bento around a **"2-Hour Junior Cyber Detective Mission"** with a 4-stage visual roadmap (`20m Skill Scan → 20m Quick Brief → 40m Live Investigation → 20m Level-Up Check`), and make the Right-Column Workbench default to a **30-Second Interactive Cyber Detective Mini-Case** (guess Real Threat vs. Safe Routine from 3 friendly log lines with instant feedback), keeping the bilingual Sample Exam as Tab 2. |
| `Q11` | Pre-Test Reassurance Banner | Display a friendly, anxiety-reducing warm-up card on Pre-Test (`ExamPage.tsx` when `mode === "pre"`): *"Not a graded school exam — 0% is completely fine! This quick scan just maps your starting level so you can watch yourself level up after training."* |
| `Q11` | First-Scenario Quick-Guide & Glossary | Add a collapsible **"30-Second How to Play & SOC Glossary"** drawer on `SimulationPage.tsx` explaining Evidence Log selection, TP/FP Decision, and 3-phase Response Actions + plain-language definitions of 10 beginner terms (`SIEM`, `KQL`, `IOC`, `True Positive`, `False Positive`, `Containment`, `Eradication`, `Recovery`, `Brute Force`, `Phishing`) — **zero spoilers**, auto-expanded ONLY on a learner's **1st Scenario** (`completed_scenarios === 0`), collapsed on Scenario 2+. |

---

## 3. Frontend Specification

### 3.1 Configurable Guest Default Locale (`frontend/src/i18n/index.tsx`, `frontend/Dockerfile`, `docker-compose.yml`)
- In `frontend/src/i18n/index.tsx`:
  - Read `const DEFAULT_GUEST_LOCALE: Locale = normalizeLocale(import.meta.env.VITE_DEFAULT_LOCALE ?? "th");`
  - Update `readStoredLocale()`: if `window.localStorage.getItem(STORAGE_KEY)` is `null`/empty, return `DEFAULT_GUEST_LOCALE`; otherwise return `normalizeLocale(stored)`.
- In `frontend/Dockerfile`:
  - Add `ARG VITE_DEFAULT_LOCALE=th` and `ENV VITE_DEFAULT_LOCALE=$VITE_DEFAULT_LOCALE` before `RUN npm run build`.
- In `docker-compose.yml`:
  - Pass `args: { VITE_DEFAULT_LOCALE: ${VITE_DEFAULT_LOCALE:-th} }` under `frontend.build`.
- Ensure existing `frontend/tests/i18n.test.ts` remains 100% green (`normalizeLocale` itself unchanged; only empty guest storage resolves to default).

### 3.2 Beginner-Friendly Landing Page (`frontend/src/pages/LandingPage.tsx`, `frontend/src/lib/landingMiniCase.ts`)
- **Zero-Login 30-Second Detective Mini-Case (`landingMiniCase.ts`)**:
  - 2 rotatable beginner-friendly mini cases (1 True Positive — e.g., midnight login brute-force + suspicious zip download from overseas IP; 1 False Positive — e.g., scheduled school/office backup script matching approved maintenance ticket) with bilingual (`EN`/`TH`) friendly briefing, 3 clean English SIEM log lines, two clear buttons (**"Real Attack (True Positive)"** vs **"Safe Routine (False Positive)"**), and instant celebratory/encouraging feedback explaining *which* log line proved the verdict.
  - Default active workbench tab = `"detective"` (`Mini Cyber Case — 30s`), Tab 2 = `"quiz"` (`Sample Question — Bilingual`).
- **4-Stage 2-Hour Mission Roadmap**:
  - Update the Bento curriculum section to clearly show the 4-stage 2-hour journey (`01 Pre-Test Skill Scan ~20m` → `02 5 Quick Domain Briefs ~20m` → `03 Live Cyber Scenarios ~40m` → `04 Post-Test Level-Up & Survey ~25m`) with clear, encouraging Thai & English copy.

### 3.3 Pre-Test Anxiety-Free Warm-Up Banner (`frontend/src/pages/ExamPage.tsx`)
- Rendered cleanly when `mode === "pre"` (both before clicking Start and as a compact dismissible reminder banner at the top of the active Pre-Test):
  - Reassures beginners that scoring low on Pre-Test is **completely normal and expected**, encourages honest answering without stress, and highlights the 1-click `ดูต้นฉบับ EN` button and `~20 min` target pace.

### 3.4 First-Scenario Quick-Guide & SOC Glossary (`frontend/src/components/SimulationBeginnerGuide.tsx`, `frontend/src/pages/SimulationPage.tsx`)
- Pure bilingual component (`EN`/`TH`) with two tabs inside a sleek Graphite collapsible card positioned cleanly above the 3-tab Simulation workspace (`Scenario / Detect & Analyze / Respond`):
  - **Tab 1 — How to Play in 30 Seconds**:
    1. **Read the Story (อ่านเรื่องย่อ)**: Check who is involved (`Host`, `User`, `IP`) — no need to memorize every word.
    2. **Hunt Evidence Logs (เลือก Log หลักฐาน)**: Click the Quick-Filter chips, check the boxes on suspicious OR benign-proof log rows (don't select random noise!).
    3. **Judge & Respond (ตัดสิน TP/FP และเลือก Playbook)**: Decide if it's a real attack (TP) or safe routine (FP), then pick sensible Containment / Eradication / Recovery actions.
  - **Tab 2 — Beginner SOC Glossary (10 Terms)**: Plain Thai + English one-liners for `SIEM`, `KQL`, `IOC`, `True Positive`, `False Positive`, `Containment`, `Eradication`, `Recovery`, `Brute Force`, `Phishing`.
- **Auto-Expand Rule**:
  - On `SimulationPage` mount, inspect `completed_scenarios` (from existing `GET /api/simulation/status`, fail-open).
  - If `completed_scenarios === 0` and user hasn't manually collapsed it for this session (`localStorage` key `cybersim_guide_collapsed:<userId>`), open expanded by default; when `completed_scenarios >= 1`, start collapsed as a sleek 1-line bar `"Need a quick refresher? Open 30s Guide & SOC Glossary"`.
- **Zero Spoilers Invariant**: Guide content is 100% static/generic and NEVER reads scenario answers, evidence indices, or `is_true_positive`.

---

## 4. Presenter Hook & 2-Hour Runbook (`docs/HIGHSCHOOL_2HR_SESSION_RUNBOOK.md`)

Single focused Thai Markdown document covering:
1. **10-Minute Opening Hook Blueprint (for User's Slides)**:
   - 3 relatable Thai teen cyber analogies (IG/Roblox account takeover, fake parcel/scholarship SMS phishing, school Wi-Fi attack) mapping directly to SOC Tier 1 work.
   - 2-minute live projector icebreaker ("Spot the Hacker" 3-line log vote with hands raised before anyone logs in).
   - How to frame the Pre-Test so students feel excited rather than anxious ("เหมือนวัดพลังตัวละคร Level 1 ก่อนลงดันเจี้ยน!").
2. **Minute-by-Minute 120-Minute Classroom Timeline** (`0:00–2:00` with exact facilitator cues at minutes `0:10`, `0:30`, `0:50`, `1:30`, `1:50`, `1:55`).
3. **Azure Pre-Flight & Classroom Troubleshooting Cheat Sheet** (registration surge tips, password reset/new account handling, browser refresh recovery, static fallback transparency, and live Admin Visualization reveal at `1:58`).

---

## 5. Verification & Acceptance Criteria

- [ ] Guest visit with empty `localStorage` defaults cleanly to `TH` (via `VITE_DEFAULT_LOCALE=th`), while `EN | TH` toggle switches cleanly without reload.
- [ ] `LandingPage` Detective Mini-Case plays interactively without login in both `EN` and `TH`, with zero answer spoilers before clicking.
- [ ] `ExamPage` (`mode === "pre"`) renders the warm-up reassurance banner cleanly; `mode === "post"` does not show Pre-Test anxiety copy.
- [ ] `SimulationBeginnerGuide` auto-expands only when `completed_scenarios === 0`, collapses cleanly on Scenario 2+, and contains zero scenario-specific hints.
- [ ] `docs/HIGHSCHOOL_2HR_SESSION_RUNBOOK.md` authored cleanly in Thai.
- [ ] `cd frontend && npm test && npm run tsc && npm run build` passes cleanly (100% i18n parity + Typography Guard green + unit tests for `landingMiniCase.ts`).
