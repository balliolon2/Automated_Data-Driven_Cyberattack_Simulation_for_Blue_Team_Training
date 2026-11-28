# Workflow: Registration Email Verification Loop (Issue #69)

## Objective
Require email ownership verification before newly registered accounts can sign in to the SOC Trainer platform, delivering a **single dual-language (Thai + English) verification email** asynchronously via standard SMTP while keeping existing/seeded Admin accounts active and providing an `REQUIRE_EMAIL_VERIFICATION` env toggle for CI and controlled classroom sessions.

---

## 1. Loop Profile

- **Loop Name**: Registration Email Verification Loop
- **Frequency / Cadence**: Event-driven (upon `/register`, `/register-specialist`, resend, or verify action)
- **Primary Actors**:
  - **Registrant**: New `learner` or `specialist` applicant registering an account
  - **Admin**: Reviewer on `/admin/specialists` (sees Specialist applications only after email verification)
- **Checkpoints**: None for standard registration (autonomous user self-verification via inbox link). Specialist role elevation still passes through the Admin checkpoint in `workflows/specialist_onboarding_and_approval.md` once the applicant's email is verified.

---

## 2. Triggers

1. **Standard Learner Signup (`POST /api/register`)**: User submits email, password, nickname, and optional `preferred_language` on `/register`.
2. **Atomic Specialist Signup (`POST /api/register-specialist`)**: Applicant selects *"Apply as Specialist"* on `/register` and submits multipart registration credentials + Bio + Resume (required) + optional Certificate / LinkedIn / Portfolio URLs in one request.
3. **Resend Request (`POST /api/resend-verification`)**: Unverified user requests a new link from the `/register` pending card, `/login` error callout, or `/verify-email` expired-link card.
4. **Link Click (`GET /verify-email?token=...` -> `POST /api/verify-email`)**: User opens the link from their email and submits verification confirmation.

---

## 3. Database & Schema Contract (`init.sql` + `models/migrate.go` + `models/user.go`)

Add verification fields directly to `users` (kept strictly in sync between `init.sql` and idempotent `models/migrate.go`):

- `email_verified BOOLEAN NOT NULL DEFAULT false` (`models/migrate.go` uses `ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT true` so pre-existing rows backfill as verified cleanly without locking out prior learners/admins; initial Admin seeding explicitly sets `email_verified = true` and `email_verified_at = NOW()`).
- `email_verified_at TIMESTAMP NULL`
- `verification_token_hash VARCHAR(64)` (`json:"-"`, indexed cleanly via `idx_users_verification_token_hash` where `verification_token_hash IS NOT NULL`)
- `verification_sent_at TIMESTAMP NULL`
- `verification_expires_at TIMESTAMP NULL`

---

## 4. Configuration & Infrastructure Contract

| Env Var | Default | Behavior |
|---|---|---|
| `REQUIRE_EMAIL_VERIFICATION` | `true` | `true`: new `/register` & `/register-specialist` accounts start `email_verified = false` and require link click before `/login`. `false`: accounts auto-verify at signup (`email_verified = true`, `email_verified_at = NOW()`) for CI (`test_integration.py`) or emergency offline classroom workshops. |
| `APP_BASE_URL` | `http://localhost` | Origin used to construct `<APP_BASE_URL>/verify-email?token=<rawToken>`. Falls back to first non-wildcard entry in `CORS_ALLOWED_ORIGINS` or `http://localhost`. `install.sh` / `install.ps1` `--mode cloud --domain <domain>` auto-populates `APP_BASE_URL=https://<domain>` alongside `FRONTEND_URL`. |
| `SMTP_HOST` | `""` | Outbound SMTP hostname. When empty, skips network send safely and logs a non-blocking server notice (**never prints plaintext tokens or secrets**). |
| `SMTP_PORT` | `587` | SMTP server port (`587` STARTTLS, `465` implicit TLS, `1025` plaintext local Mailpit). |
| `SMTP_USER` / `SMTP_PASSWORD` | `""` | Optional SMTP authentication credentials (`PLAIN` / `LOGIN` over TLS). |
| `SMTP_FROM` | `SMTP_USER` or `no-reply@soc-trainer.local` | Sender header (`From: SOC Trainer <...>`). |
| `SMTP_TLS_MODE` | `starttls` | `starttls` \| `tls` \| `none` (allows `none` explicitly for local Mailpit dev container). |

- **Local Dev Mailpit Profile (`docker-compose.yml`)**: Opt-in `mailpit` (`axllent/mailpit`) under `profiles: ["dev-email"]` binding `127.0.0.1:8025:8025` (Web UI) and `127.0.0.1:1025:1025` (SMTP) so default production/CI stacks stay lean while local developers can inspect dual-language HTML emails visually.

---

## 5. Pipeline Execution

### Step 1 — Registration & Token Issuance
1. When `REQUIRE_EMAIL_VERIFICATION=true`:
   - Generate `32` CSPRNG bytes (`crypto/rand`) encoded with `base64.RawURLEncoding` as `<rawToken>`.
   - Compute lowercase hex `SHA-256(<rawToken>)` -> `verification_token_hash` (never persist raw tokens).
   - Set `email_verified = false`, `verification_sent_at = NOW()`, `verification_expires_at = NOW() + 24h`.
2. **Atomic Specialist Registration (`POST /api/register-specialist`)**:
   - Validate password/email/nickname uniqueness and multipart `bio` + `resume` (<= 10MB, `.pdf/.png/.jpg/.jpeg`) + optional `certificate` BEFORE creating DB records, or run file save + DB transaction with deterministic rollback cleanup (`os.RemoveAll(appDir)`) if DB creation fails.
   - Creates both `users` (`role = 'learner'`, `email_verified = false`) and `specialist_applications` (`status = 'pending'`) atomically.
3. Dispatch verification email **asynchronously** (goroutine with bounded `15s` dial/send timeout so slow or failing SMTP never fails registration HTTP `201`).
4. Return `201 Created` including `"requires_verification": true` (or `false` when bypassed via env flag) so the frontend branches cleanly without hardcoding server config.

### Step 2 — Dual-Language Async Email Template (`services/email_service.go`)
- Sends a single `multipart/alternative` (plain text + clean HTML) email containing **both Thai and English** in one message (independent of per-user UI locale toggle):
  - **Subject**: `[SOC Trainer] ยืนยันอีเมลเพื่อเปิดใช้งานบัญชี / Verify Your Email Address`
  - **Body**: Clear bilingual welcome + verification CTA button & raw fallback link (`<APP_BASE_URL>/verify-email?token=<rawToken>`) + explicit `24 ชั่วโมง / 24 hours` expiration notice. Never includes passwords or internal tokens outside the verification URL.

### Step 3 — Login Gate (`POST /api/login`)
1. Always validate email + `bcrypt` password **first** (`401 AUTH_INVALID_CREDENTIALS` on wrong password so account verification status is never leaked to unauthenticated callers).
2. If password succeeds, `REQUIRE_EMAIL_VERIFICATION=true`, and `!user.EmailVerified`:
   - Reject with `403 Forbidden`, `code: "AUTH_EMAIL_NOT_VERIFIED"`, `"error": "Pending email verification. Please check your inbox and verify your email before logging in."`, and `"email": user.Email`.
3. Do NOT issue a JWT until `email_verified == true`.

### Step 4 — Verification (`POST /api/verify-email`)
- Accepts `{ "token": "<rawToken>" }` (never mutates DB on `GET`, preventing email security scanners like Outlook SafeLinks from pre-consuming single-use tokens).
- Computes `SHA-256(token)` and looks up matching user:
  - No match -> `400 Bad Request` (`AUTH_VERIFICATION_TOKEN_INVALID`).
  - Expired (`NOW() > verification_expires_at`) -> `400 Bad Request` (`AUTH_VERIFICATION_TOKEN_EXPIRED`).
- Updates `email_verified = true`, `email_verified_at = NOW()`, clears `verification_token_hash = NULL`, `verification_sent_at = NULL`, `verification_expires_at = NULL` (strict single-use), and returns `200 OK` (`message`, `email`).

### Step 5 — Resend Flow (`POST /api/resend-verification`)
- Accepts `{ "email": "..." }`:
  - If email syntax is invalid -> `400 Bad Request` (`AUTH_INVALID_PAYLOAD`).
  - If user does not exist or is already verified (or `REQUIRE_EMAIL_VERIFICATION=false`) -> return neutral `200 OK` (`"If the account requires verification, a new verification email has been sent."`) to prevent email enumeration.
  - If user exists and `!email_verified` and `verification_sent_at` is within **60 seconds** -> return `429 Too Many Requests` (`code: "AUTH_VERIFICATION_COOLDOWN"`, `"retry_after_seconds": remaining`).
  - Otherwise overwrite `verification_token_hash`, `verification_sent_at = NOW()`, `verification_expires_at = NOW() + 24h` (invalidating any previous token), dispatch email asynchronously, and return `200 OK`.

### Step 6 — Admin Queue Gate (`controllers/admin_controller.go`)
- `GET /api/admin/applications` filters joined `users.email_verified = true` so pending Specialist applications submitted via `/register-specialist` remain hidden until the applicant confirms email ownership.
- `POST /api/admin/applications/:id/approve` additionally validates target user `email_verified == true`, returning `400` (`ADMIN_APPLICANT_EMAIL_UNVERIFIED`) if called prematurely.

---

## 6. Frontend UX Contract

- **`/register` (`RegisterPage.tsx`)**:
  - *"Apply as Specialist"* tab submits `POST /api/register-specialist` (atomic multipart signup + document upload) instead of chaining `/register` -> `/login` -> `/specialist/apply`.
  - When response returns `requires_verification: true`, transitions in-place to a bilingual **Pending Email Verification** state showing the registered email, 24-hour expiration notice, a 60s-cooldown **Resend Verification Email** button, and a **Go to Login** button.
  - When `requires_verification: false` (`REQUIRE_EMAIL_VERIFICATION=false`), maintains the instant Login / Specialist `/profile` redirect flow unchanged.
- **`/login` (`LoginPage.tsx`)**:
  - On `403 AUTH_EMAIL_NOT_VERIFIED`, displays a clear bilingual **Pending Email Verification** banner with an inline **Resend Verification Email** button calling `POST /api/resend-verification` for that email (with 60s countdown/cooldown feedback).
- **`/verify-email` (`VerifyEmailPage.tsx`)**:
  - Public route reading `?token=` and calling `POST /api/verify-email` once on mount (StrictMode-safe via `useRef` guard).
  - Renders bilingual **Success** (CTA to `/login`) or **Invalid / Expired Token** state with an inline email field + **Resend Verification Email** action.
- **Typography & Bilingual Parity (`AGENTS.md`, `i18n/en.ts`, `i18n/th.ts`)**:
  - Uses semantic `rem`-only typography (`text-base`, `text-sm`, `text-xs`), Graphite palette, and 100% EN/TH i18n key parity verified by `frontend/tests/i18n.test.ts` and `frontend/tests/typography.test.ts`.

---

## 7. Verification & CI Acceptance

1. **Go Tests**: Unit tests for `services/email_service.go` (CSPRNG + SHA-256 hashing, dual-language TH+EN MIME generation, non-blocking fallback without logging raw tokens, mock SMTP client), `controllers/auth_test.go` (`403 AUTH_EMAIL_NOT_VERIFIED` after valid password only, `POST /api/verify-email` single-use & expiry, `60s` resend cooldown, `REQUIRE_EMAIL_VERIFICATION=false` bypass), and atomic `register-specialist` + Admin unverified guard tests.
2. **Frontend & OpenAPI**: `docs/swagger.{json,yaml}` regenerated cleanly (`dto` request/response structs updated per ADR-0005); `npm run lint`, `npm test` (`i18n` + `typography`), and `npm run build` pass.
3. **CI / Integration Suites**: `.github/workflows/ci.yml` keeps existing E2E suites green via `REQUIRE_EMAIL_VERIFICATION=false`, paired with dedicated integration assertions exercising `REQUIRE_EMAIL_VERIFICATION=true` (`register` -> blocked `login` `403` -> verify / Mailpit check -> `login` `200`).

---

## 8. Out of Scope for Issue #69
- Outbound notification emails when Admin approves or rejects Specialist applications (already uses in-app notifications in `workflows/specialist_onboarding_and_approval.md`; can reuse `services/email_service.go` in a follow-up ticket).
- Self-service "Forgot Password" email reset loop or multi-factor email OTP codes.
