# Workflow Spec: App-Wide Semantic Typography Scale & Readability Guard (Issue #38)

**GitHub Issue:** [#38 — [UI] fonts is too small](https://github.com/balliolon2/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/issues/38)  
**Status:** `/loop-me` Grilling Rounds 1–2 Settled (`Q1`–`Q12` Resolved, 2026-11-23) — Ready for User Confirmation Before Implementation  
**Depends On:**
- [language_switching_and_localization.md](file:///c:/Users/bond/Documents/Automated_Data-Driven_Cyberattack_Simulation_for_Blue_Team_Training/workflows/language_switching_and_localization.md) (`Geist` + `IBM Plex Sans Thai` bilingual rendering across `en`/`th`)

---

## Objective

Eliminate microscopic, low-legibility typography across the entire CyberSim/SOC frontend—where an audit of `frontend/src/**/*.tsx` found **81% of all 635 font-size classes at `<= 12px`** (including **195 sub-12px arbitrary pixel classes** from `text-[8px]` to `text-[11px]` and only 16 occurrences of `text-base`)—by establishing a **Semantic Typography Scale** built 100% on `rem` units, improving bilingual Thai/English line-height and monospace fallback parity, cleaning up muted contrast/truncation regressions, and locking a deterministic **Automated Typography Guard** into `npm test` and GitHub Actions CI so sub-floor or arbitrary-pixel typography can never regress.

The two loops this workflow makes real:
1. **App-Wide Typography Calibration Loop**: Establish canonical `rem`-based semantic tokens (`tailwind.config.js` + `index.css`), refactor all 16 pages and 10 shared components to the role-based scale, apply targeted anti-truncation/layout and contrast hygiene fixes, and pass a single push-right visual review checkpoint comparing Before/After screenshots in both `EN` and `TH`.
2. **Continuous Typography Regression Guard Loop**: Every local `npm test` run and every PR/push in GitHub Actions CI (`lint-and-audit`) executes `frontend/tests/typography.test.ts`, automatically failing if any arbitrary pixel font size (`text-[...px]`), sub-12px `rem`/`em`/inline font size, invalid `text-graphite-550` class, or missing Thai monospace fallback / `text-log` token is introduced.

---

## 1. Loop Profile

- **Loop Name**: App-Wide Semantic Typography & Readability Guard Loop
- **Frequency / Cadence**:
  - One-time calibration pass across all 26 active TSX files (`16` pages + `10` shared components), `tailwind.config.js`, `index.css`, `package.json`, `.github/workflows/ci.yml`, and `AGENTS.md`.
  - Continuous CI/test enforcement on every subsequent frontend edit, PR, and merge to `main`.
- **Primary Actors**:
  - **Design Token Layer (`frontend/tailwind.config.js`, `frontend/src/index.css`)**: Defines `rem`-only scale including named `text-log` (`0.8125rem` / 13px), adds `"IBM Plex Sans Thai"` fallback to `fontFamily.mono`, and sets base `body` defaults (`text-base leading-relaxed`).
  - **Shared UI Components (`frontend/src/components/*.tsx`, 10 files)**: Standardizes typography, anti-truncation wrapping, proportional icons/checkboxes, and contrast across `Layout`, `DomainBreakdown`, `LogTable`, `KqlSearchBar`, `MarkdownView`, `ResponsePanel`, `RoleBadge`, `CommentSection`, `NotificationBell`, and `ScenarioLoadingState`.
  - **Application Pages (`frontend/src/pages/*.tsx`, 16 files)**: Upgrades all prose, UI controls, headings, cards, badges, and SIEM/evidence views to the Semantic Typography Scale.
  - **Automated Typography Guard (`frontend/tests/typography.test.ts`, `frontend/package.json`, `.github/workflows/ci.yml`)**: Static AST/regex verification suite ensuring zero sub-floor or arbitrary-`px` regressions.
- **Trigger**:
  - Calibration pass: Issue #38 implementation start.
  - Guard loop: event-driven on `npm test` locally and on every `pull_request` / `push` to `main` via `.github/workflows/ci.yml`.
- **Checkpoint**:
  - **1 Push-Right Visual Review Checkpoint (`Q12=A`)**: After completing all 26 files and passing `npm test`, `npx tsc -b`, `npm run lint`, and `npm run build`, present a single compact **Visual Review Brief** containing Before/After screenshots (`LandingPage`, `ScenarioResultPage`, `SimulationPage`, `LearnTheoryPage`, `ExamPage`) in both `EN` and `TH`, plus the font-class delta summary, for user sign-off prior to closing Issue #38.
  - **Ongoing CI Guard**: None (100% autonomous).

---

## 2. Settled Design Decisions (`Q1`–`Q12`)

| Q# | Decision | Settled Rule |
| --- | --- | --- |
| `Q1` | Scope | **App-wide (`1A`)**: All 16 pages + all 10 shared components + config/CSS/tests/docs. |
| `Q2` | Scaling Mechanism | **Semantic Typography Scale (`2A`)**: Role-based component/page typography refactor eliminating `text-[8px]`–`text-[11px]` at the source (no global root `html { font-size }` hack). |
| `Q3` | User Size Switcher | **Fixed Well-Calibrated Default (`3A`)**: No in-app `A-/A+` navbar toggle; 100% `rem`-based scale respects native browser zoom (`Ctrl + / -`). |
| `Q4` | SIEM Log Density | **Readable Compact (`4A`)**: `text-log` (`0.8125rem` / 13px) for dense log rows/messages and code blocks, `text-sm` (14px) for `KqlSearchBar` input & primary evidence callouts, `text-xs` (12px) for table headers/metadata/pills. |
| `Q5` | Thai / Bilingual Parity | **Unified Scale + Thai-Friendly Line-Height & Mono Fallback (`5A`)**: Same font sizes across `en`/`th` (zero layout jump on toggle), comfortable `leading-relaxed` / `leading-snug` so Thai vowels/tone marks never collide, plus `"IBM Plex Sans Thai"` fallback in `fontFamily.mono`. |
| `Q6` | Regression Loop | **Automated Typography Guard (`6A`)**: `frontend/tests/typography.test.ts` wired into `npm test` and GitHub Actions CI (`lint-and-audit`). |
| `Q7` | Exact Scale Tiers | **Balanced Modern Web Scale (`7A`)**: `16px` primary reading prose, `14px` standard UI/controls, `13px` dense SIEM logs/code, `12px` hard floor (zero sub-12px text anywhere). |
| `Q8` | Tailwind Token Architecture | **Standard Tailwind `rem` + 1 Named `rem` Token (`8A`)**: Keep standard Tailwind `text-xs`..`text-6xl` plus `fontSize.log: ['0.8125rem', { lineHeight: '1.5' }]` in `tailwind.config.js`; zero arbitrary `text-[...px]` anywhere. |
| `Q9` | Layout & Anti-Truncation | **Targeted Responsive & Anti-Truncation Polish (`9A`)**: Allow wrapping on long domain names in `DomainBreakdown`, maintain single-row Navbar at `>= 1280px` across all roles (`en`/`th`), size checkboxes/icons proportionally (`h-4 w-4`), and adjust `LogTable` container height (`max-h-[560px]`). |
| `Q10` | Guard Strictness | **Strict Floor + Zero Arbitrary `px` (`10A`)**: Forbid all `text-[<N>px]`, sub-`0.75rem` arbitrary sizes, `text-2xs`, sub-12px inline `fontSize`, invalid `text-graphite-550`, and enforce `tailwind.config.js` `mono` Thai fallback & `text-log` token. |
| `Q11` | Contrast Hygiene | **Fix Invalid Class + Lift Muted Readable Copy (`11A`)**: Replace invalid `text-graphite-550` with `text-graphite-400`; lift readable explanatory/hint/caption copy from `text-graphite-500` to `text-graphite-400` while preserving Graphite dark mode aesthetic. |
| `Q12` | Human Checkpoint | **1 Push-Right Visual Review Checkpoint (`12A`)**: Show compact Before/After screenshot Brief across 5 flagship screens (`EN` + `TH`) after all automated checks pass. |

---

## 3. Canonical Semantic Typography Scale & Tokens

### 3.1 `frontend/tailwind.config.js` Updates
Add `"IBM Plex Sans Thai"` fallback to `fontFamily.mono` and register the named `rem` token `log` under `theme.extend.fontSize`:

```js
fontFamily: {
  sans: ['Geist', '"IBM Plex Sans Thai"', 'sans-serif'],
  mono: ['"Geist Mono"', '"IBM Plex Sans Thai"', 'monospace'],
},
fontSize: {
  log: ['0.8125rem', { lineHeight: '1.5' }], // 13px at 16px root, 100% rem-scalable
},
```

### 3.2 `frontend/src/index.css` Updates
Update `@layer base` `body` so unclassed body copy defaults to readable `16px` with Thai-friendly line-height:

```css
@layer base {
  body {
    @apply bg-graphite-950 text-graphite-50 font-sans text-base leading-relaxed antialiased;
    overflow-x: hidden;
  }
}
```

### 3.3 Semantic Role Mapping Table (Single Source of Truth)

| Semantic Role | Tailwind Class | Computed Size (16px root) | Line Height | Where Used |
| --- | --- | --- | --- | --- |
| **Hero Display** | `text-4xl sm:text-5xl lg:text-6xl` | `36px–60px` | `leading-tight` | `LandingPage` hero headline |
| **Page Title (`h1`)** | `text-2xl sm:text-3xl` | `24px–30px` | `leading-tight` | Top heading on `ScenarioResultPage`, `EvaluationPage`, `SimulationPage`, `DiscussionsFeedPage`, `LearnTheoryPage`, `ProfilePage`, `AdminSpecialistsPage`, `SpecialistReviewPage`, `ThreadDetailPage`, `ThreadEditorPage`, `SimulationCompletePage` |
| **Section Heading (`h2`)** | `text-xl` | `20px` | `leading-snug` | Major page sections, `MarkdownView` `h2`, modal titles, Auth card titles (`LoginPage`, `RegisterPage`) |
| **Card / Sub-Section Heading (`h3`)** | `text-lg` | `18px` | `leading-snug` | Card titles (`Triage classification`, `Detect & Analysis Findings`, `Updated Domain Proficiency`, `Incident Narrative`, `MarkdownView` `h3`, finding titles) |
| **Primary Reading Prose** | `text-base` | `16px` (`1rem`) | `leading-relaxed` | Scenario incident narratives, Triage/Finding/Playbook explanations, Exam questions & options (`ExamPage`), `LandingPage` hero subtitle & sample quiz question/options, `MarkdownView` paragraphs/lists/blockquotes, Discussion thread body & comment prose |
| **Standard UI & Controls** | `text-sm` | `14px` (`0.875rem`) | `leading-normal` / `leading-snug` | Top Navbar links/nickname/buttons, Primary & secondary buttons, Form labels/inputs/textareas/selects, Playbook response option cards, `DomainBreakdown` domain names & scores, Sidebar theory sub-topic links, Tabs, Alerts/banners, KQL input bar, Evidence callout summaries |
| **Dense SIEM Telemetry & Code** | `text-log` | `13px` (`0.8125rem`) | `1.5` (built-in) | `LogTable` data cells (`timestamp`, `source`, `severity`, `host`, `message`), `MarkdownView` fenced code blocks / inline `code` / markdown table cells, compact matched-evidence log blocks (`ScenarioResultPage`, `SpecialistReviewPage`), KQL quick-filter chips |
| **Hard Floor (Micro Metadata & Badges)** | `text-xs` | `12px` (`0.75rem`) | `leading-normal` | `RoleBadge`, Domain/Difficulty/Status pills, `EN \| TH` toggle buttons, `LogTable` column headers & interaction hint, Timestamps, Field helper text, Intentional English-only i18n notices, Footer brand & copyright, Notification badge counts |
| **Forbidden** | `text-[<N>px]`, `text-[<0.75rem]`, `text-2xs`, inline `< 12px` | `< 12px` or fixed `px` | — | **Strictly banned everywhere in `frontend/src/`** |

### 3.4 `MarkdownView.tsx` Heading & Content Hierarchy
Because Markdown body prose moves from `text-sm` (14px) → `text-base` (16px), every Markdown element steps up proportionally to preserve clear visual hierarchy:
- `h1`: `text-2xl font-bold` (24px)
- `h2`: `text-xl font-bold` (20px)
- `h3`: `text-lg font-semibold` (18px)
- `h4`: `text-base font-semibold` (16px)
- Paragraphs (`p`), unordered/ordered list items (`li`), blockquotes (`blockquote`): `text-base leading-relaxed` (16px)
- Fenced code blocks (`pre`), inline `code`, and Markdown table body cells (`td`): `text-log font-mono` (13px)
- Markdown table headers (`th`): `text-xs font-semibold uppercase tracking-wider` or `text-sm font-semibold` (14px recommended for readability)
- Code block language header pill: `text-xs font-mono` (12px, replaces `text-[10px]`)

---

## 4. App-Wide Migration & Layout Polish Plan (26 Active TSX Files)

### 4.1 Shared Components (`frontend/src/components/`, 10 files)

| File | Key Typography Upgrades | Targeted Layout / Contrast Polish (`Q9A`, `Q11A`) |
| --- | --- | --- |
| `Layout.tsx` | Brand `text-lg`; Nav links `text-sm`; Nickname & Logout/Login buttons `text-sm`; `EN \| TH` toggle & inline role badge `text-xs` (replaces `text-[11px]`/`text-[10px]`); Footer `text-xs`–`text-sm`. | Verify single-row Navbar at `>= 1280px` (`xl`) across `learner`, `specialist`, and `admin` roles in both `EN` and `TH`. Use `gap-3 xl:gap-5` on nav links so Thai admin nav never wraps awkwardly. Lift footer copyright from `text-graphite-500` → `text-graphite-400`. |
| `DomainBreakdown.tsx` | Card heading `text-base`–`text-lg`; Domain row label `text-sm`; Score percentage `text-base font-bold`; Threshold `/ 70%` `text-xs`; `Study Theory` link `text-xs`. | **Remove single-line `truncate`** on domain name span; use `leading-snug break-words` with `items-start` row alignment so `"Security Program Management and Oversight"` and bilingual Thai domain names display in full without `...` clipping. Bump status icon `w-3.5 h-3.5` → `w-4 h-4`. Lift `/ 70%` from `text-graphite-500` → `text-graphite-400`. |
| `LogTable.tsx` | Header count `text-sm`; Interaction hint `text-xs` (replaces `text-[10px]`); Table headers `text-xs`; Table rows & `message` column `text-log` (`0.8125rem` / 13px). | Increase scroll container `max-h-[500px]` → `max-h-[560px]` to preserve visible row count; bump checkbox `h-3.5 w-3.5` → `h-4 w-4`; lift interaction hint & empty state from `text-graphite-500` → `text-graphite-400`. |
| `KqlSearchBar.tsx` | Search input `text-sm`; Quick-filter chips & operator hints `text-xs`–`text-log` (replaces `text-[10px]`). | Ensure comfortable input height (`py-2.5`) and lift hint text contrast to `text-graphite-400`. |
| `MarkdownView.tsx` | Full step-up per Section 3.4 (`p`/`li`/`blockquote` `text-base`, `h1–h4` `text-2xl..text-base`, `code`/`pre`/`td` `text-log`, code lang bar `text-xs`). | Add slightly generous vertical rhythm (`my-2`, `leading-relaxed`) so Thai diacritics and multi-line lists read effortlessly in `/learn` and `/discussions`. |
| `ResponsePanel.tsx` | Phase headings `text-base`; Option action labels `text-sm leading-relaxed`; Phase instructions/counters `text-sm` / `text-xs`. | Ensure checkbox/selection cards have comfortable vertical padding (`p-3.5`) for multi-line Thai playbook options. |
| `RoleBadge.tsx` | All 4 role badges upgraded from `text-[9px]` → `text-xs` (`px-2 py-0.5`). | Keep `whitespace-nowrap shrink-0` so badges never split across lines next to nicknames. |
| `CommentSection.tsx` | Comment prose `text-base leading-relaxed`; Author/actions/reply buttons/textarea `text-sm`; Timestamps/vote counters/badges `text-xs` (eliminates all `text-[10px]`/`text-[11px]`). | Maintain clean 2-level reply indentation without cramping comment text. |
| `NotificationBell.tsx` | Dropdown heading `text-sm`; Notification title `text-sm`; Notification message `text-xs`–`text-sm`; Timestamp & unread badge `text-xs` (eliminates `text-[9px]`–`text-[11px]`). | Widen dropdown slightly if needed (`w-80 sm:w-96`) so Thai notification messages breathe comfortably. |
| `ScenarioLoadingState.tsx` | Loading headings `text-xl`–`text-2xl`; Status prose `text-sm`–`text-base`; Step pills/telemetry flavor `text-xs` (eliminates `text-[10px]`/`text-[11px]`). | Lift muted progress labels from `text-graphite-500` → `text-graphite-400`. |

### 4.2 Pages (`frontend/src/pages/`, 16 files)

| Priority Group | Files | Target Typography Upgrades & Polish |
| --- | --- | --- |
| **Flagship Pages (Shown in Issue #38 & Core Flow)** | `LandingPage.tsx`, `ScenarioResultPage.tsx`, `SimulationPage.tsx`, `ExamPage.tsx`, `LearnTheoryPage.tsx` | - **`LandingPage.tsx`**: Hero subtitle `text-base sm:text-lg leading-relaxed`; CTA buttons `text-sm sm:text-base`; Sample quiz question & options `text-base` (`triage-quiz.md` tab `text-sm`, header metadata `CompTIA Security+ 701 Question` / `Difficulty` `text-xs`–`text-sm`, action buttons `text-sm`); Curriculum section heading `text-2xl`, card titles `text-lg`, card descriptions `text-sm sm:text-base`; **fix invalid `text-graphite-550` → `text-graphite-400`**; eliminate all `text-[8px]`–`text-[11px]`.<br>- **`ScenarioResultPage.tsx`**: Page title `text-3xl`, subtitle `text-sm sm:text-base`; Score display `text-5xl`; Card titles (`Triage classification`, `Detect & Analysis Findings`, `Incident Response Playbook`) `text-lg`; Triage explanation & finding rationales `text-base leading-relaxed`; Finding titles `text-base font-semibold`; Matched evidence log rows `text-log font-mono`; Domain/status/score badges `text-xs` (eliminates all 27 `text-[8px]`–`text-[11px]` instances).<br>- **`SimulationPage.tsx`**: Scenario title `text-2xl sm:text-3xl`; Incident Narrative paragraphs `text-base leading-relaxed`; Phase step tabs & SIEM shortcut buttons `text-sm` / `text-xs`; TP/FP verdict cards `text-base` heading + `text-sm` description; Selected evidence summary chips `text-xs`–`text-log`; eliminate all 24 `text-[10px]`–`text-[11px]` instances.<br>- **`ExamPage.tsx`**: Question stem `text-lg` / `text-base` (`leading-relaxed`); Answer options `text-base`; Progress bar & domain label `text-sm`; English-only TH notice `text-xs` (already 0 sub-12px, ensure full consistency with `16px` prose / `14px` controls).<br>- **`LearnTheoryPage.tsx`**: Domain/subtopic sidebar items `text-sm`; Main theory rendered via upgraded `MarkdownView`; Sample question stem & options `text-base`; Metadata pills/counters `text-xs` (eliminates all 25 `text-[10px]`–`text-[11px]` instances). |
| **Assessment & Profile Pages** | `EvaluationPage.tsx`, `SimulationCompletePage.tsx`, `ProfilePage.tsx`, `LoginPage.tsx`, `RegisterPage.tsx`, `ExamRedirectPage.tsx` | - Upgrade all primary descriptions/status messages to `text-base`, form labels/inputs/buttons/cards to `text-sm`, badges/timestamps/hints to `text-xs` (zero sub-12px).<br>- Lift helper text and secondary labels from `text-graphite-500` → `text-graphite-400`. |
| **Community, Specialist & Admin Pages** | `DiscussionsFeedPage.tsx`, `ThreadDetailPage.tsx`, `ThreadEditorPage.tsx`, `SpecialistReviewPage.tsx`, `AdminSpecialistsPage.tsx` | - Thread titles `text-lg` (feed) / `text-2xl sm:text-3xl` (detail); Thread excerpts & review descriptions `text-sm`–`text-base`; Specialist review evidence logs `text-log font-mono`; Form editors & filter dropdowns `text-sm`; Status/domain/role badges & timestamps `text-xs` (eliminates all 35 sub-12px instances in `SpecialistReviewPage` and all sub-12px across discussions/admin). |

---

## 5. Automated Typography Guard & CI Integration (`Q6A`, `Q10A`)

### 5.1 New Test Suite: `frontend/tests/typography.test.ts`
Using Node's built-in `node:test` + `node:assert/strict` (matching `tests/i18n.test.ts` and `tests/siemShortcuts.test.ts`), recursively scan `frontend/src/**/*.{tsx,ts,css}` (excluding dead unused `App.css` if deleted, or removing unused `frontend/src/App.css` cleanly since it has zero imports) and assert:

1. **Zero Arbitrary Pixel Font Sizes (`rem`-only enforcement)**:
   - Regex `/\btext-\[\d+(?:\.\d+)?px\]/g` MUST have **0 matches** across all `frontend/src/**/*.{tsx,ts,css}`.
2. **Zero Sub-Floor `rem`/`em` or Non-Standard Tiny Classes**:
   - Regex `/\btext-\[(?:0?\.\d+)(?:rem|em)\]/g`: parse numeric value; fail if `< 0.75`.
   - Regex `/\btext-(?:2xs|3xs)\b/g`: MUST have **0 matches**.
3. **Zero Sub-12px Inline JS/CSS `fontSize`**:
   - Match inline JS `fontSize\s*:\s*['"]?(\d+(?:\.\d+)?)(px|rem)?['"]?` and CSS `font-size\s*:\s*(\d+(?:\.\d+)?)(px|rem)` — fail if `px < 12` (or bare numeric `< 12`) or `rem < 0.75`.
4. **Zero Invalid Palette Typography Classes**:
   - Regex `/\btext-graphite-550\b/g` MUST have **0 matches**.
5. **Design Token & Bilingual Mono Fallback Contract (`frontend/tailwind.config.js`)**:
   - Assert `tailwind.config.js` includes `"IBM Plex Sans Thai"` inside `fontFamily.mono`.
   - Assert `tailwind.config.js` defines `fontSize.log` with a `rem` value `>= 0.75rem` (specifically `0.8125rem`).

### 5.2 Wiring into `package.json`, CI, and `AGENTS.md`
- **`frontend/package.json`**:
  ```json
  "test": "node --test tests/i18n.test.ts tests/siemShortcuts.test.ts tests/typography.test.ts"
  ```
- **`.github/workflows/ci.yml`**:
  - Update step display name from `Frontend Unit Tests (i18n & SIEM shortcuts)` → `Frontend Unit Tests (i18n, SIEM shortcuts & typography guard)` (it already executes `cd frontend && npm test`).
- **`AGENTS.md`**:
  - Document the Semantic Typography Scale (`16px` prose / `14px` UI / `text-log` `13px` SIEM telemetry / `12px` hard floor, `rem`-only, zero `text-[...px]`) under React Frontend Coding Standards.
- **Unused Scaffold Cleanup**:
  - Delete unused `frontend/src/App.css` (zero imports verified) or confirm zero references so the CSS tree stays clean.

---

## 6. Push-Right Visual Verification Checkpoint (`Q12A`)

Before declaring Issue #38 complete, run the full verification sequence:
1. **Automated Gates (must all pass green)**:
   - `cd frontend && npm test` (runs `i18n.test.ts`, `siemShortcuts.test.ts`, `typography.test.ts`)
   - `cd frontend && npx tsc -b`
   - `cd frontend && npm run lint`
   - `cd frontend && npm run build`
2. **Human Visual Review Brief**:
   - Present a compact, decision-ready **Brief** containing:
     - **Typography Audit Delta**: Before (`195` sub-12px, `320` `text-xs`, `67` `text-sm`, `16` `text-base`) vs. After (`0` sub-12px, `0` `text-[...px]`, role-balanced counts).
     - **Visual Evidence (Screenshots)**: Desktop (`1440x900` or `1720x940`) and Laptop (`1280x800`) renders in both **`EN`** and **`TH`** covering the 5 flagship views:
       1. `LandingPage` (hero + `triage-quiz.md` + curriculum cards)
       2. `ScenarioResultPage` (score + `DomainBreakdown` without clipping + triage explanation + Detect & Analysis findings + matched evidence)
       3. `SimulationPage` (Incident Narrative + `KqlSearchBar` + `LogTable` + `ResponsePanel`)
       4. `LearnTheoryPage` (sidebar + `MarkdownView` headings/prose/tables/code + sample question)
       5. `ExamPage` (question stem + options + header + Thai English-only notice)
     - **Overflow Spot-Check**: Confirmation of single-row Navbar at `1280px` in `TH` mode and unwrapped Domain names in `DomainBreakdown`.
3. **Definition of Done**:
   - All automated gates pass + user approves the Visual Review Brief (or requests targeted tweaks which are applied and re-verified) → ready to close Issue #38.
