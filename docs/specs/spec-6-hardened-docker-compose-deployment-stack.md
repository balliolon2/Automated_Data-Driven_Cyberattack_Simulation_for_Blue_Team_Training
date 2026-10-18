---
title: "Spec 6: Hardened, Project-Correct Docker Compose Deployment Stack"
labels: ["ready-for-agent"]
status: "ready-for-agent"
---

# Spec 6: Hardened, Project-Correct Docker Compose Deployment Stack

## Problem Statement

Anyone who runs the platform from a clean checkout cannot trust what they get. The Compose stack ships publicly known default credentials (database, JWT, pgAdmin), runs every service as root with no isolation, publishes the database to the host, and installs packages over the network at start-up. Worse, the "one-shot" seeding container reports success even when it loaded nothing, so a Learner can reach an empty or half-seeded environment, and a Specialist's résumé/certificate upload fails because the backend filesystem is now read-only. Regenerated exam content contains raw markdown artifacts, so Learners see corrupted question text and options. A Deployer has no single, honest signal that the environment is actually ready, and a Security Reviewer cannot point to a defensible baseline.

## Solution

A hardened, fail-closed, project-correct Compose deployment stack for the SOC Trainer, in which:

1. Secrets are mandatory and supplied only through the environment; the example template ships blank so the stack refuses to start until real values are set.
2. Every service runs with least privilege (non-root, read-only root filesystem where possible, dropped capabilities, no new privileges) while still allowing the artifacts the app must persist.
3. The database is reachable only by the services that need it; only the frontend is exposed on the host, with API and pgAdmin bound to loopback.
4. Seeding runs automatically as a build-time-pinned, non-root one-shot step whose failure is loud, so the backend only starts against a verified-seeded database.
5. The optional pgAdmin tool never blocks the core stack when unused.
6. Exam content is emitted clean (no stray markdown or embedded newlines) so the seeded question pool renders correctly to Learners.

## User Stories

1. As a Deployer, I want the stack to refuse to start when required secrets are missing, so that I never accidentally run a public environment with default credentials.
2. As a Deployer, I want a single command to bring up the whole environment from a clean checkout, so that setup is repeatable and unsurprising.
3. As a Deployer, I want the example environment file to contain no usable secrets, so that copying it cannot silently produce an insecure deployment.
4. As a Deployer, I want to know that seeding finished successfully before the backend accepts traffic, so that I never serve an empty question pool.
5. As a Deployer, I want a failed seeding step to fail the stack loudly, so that I can fix it instead of debugging a phantom "healthy" service.
6. As a Deployer, I want per-service health checks and resource limits, so that one runaway service cannot exhaust the host or hide a failure.
7. As a Deployer, I want bounded log files, so that long-running environments do not fill the disk.
8. As a Security Reviewer, I want the database to be reachable only from the services that need it, so that a compromised front end cannot talk to the database directly.
9. As a Security Reviewer, I want the API and the database admin tool bound to loopback, so that they are not exposed to the network by default.
10. As a Security Reviewer, I want the backend to run as an unprivileged user with a read-only filesystem and dropped capabilities, so that a backend compromise has minimal blast radius.
11. As a Security Reviewer, I want the seeding step to run as a non-root user with pinned dependencies installed at build time, so that it does not fetch and execute unpinned code as root.
12. As a Security Reviewer, I want the repository's real environment file to remain untracked, so that secrets never enter version control.
13. As a Learner, I want the pre-test to return the full 30-question pool distributed across the five domains, so that my Proficiency Score is measured against the intended blueprint.
14. As a Learner, I want question stems, options, and explanations to render as clean text, so that I am not distracted or misled by stray markup.
15. As a Learner, I want session recovery to keep working after a refresh, so that I can resume an interrupted attempt.
16. As a Specialist, I want to upload my résumé and certificates successfully, so that my application can be reviewed.
17. As a Specialist, I want my uploaded artifacts to survive a container restart, so that my application is not lost between sessions.
18. As a Specialist, I want the review console to read the same uploaded artifacts, so that I can trust the review experience.
19. As an Admin, I want to start the optional database admin tool only when I need it, so that the core stack stays minimal.
20. As an Admin, I want the optional tool to never block the core stack when I am not using it, so that bringing up the platform never fails because of an unused component.
21. As a Maintainer, I want re-running seeding to be idempotent and non-destructive, so that Learners' sessions and research data are never wiped.
22. As a Maintainer, I want the seeder to report what it inserted, updated, and skipped, so that I can verify content changes.
23. As a Maintainer, I want the parser to normalize exam source text, so that regenerating the question pool does not reintroduce markup defects.
24. As a Researcher, I want seeding never to delete Simulation Sessions, exam sessions, or research data, so that attrition and selection-bias analysis remains valid.

## Implementation Decisions

- **Compose topology.** A single deployment definition with these services: database; a one-shot seeder; backend; frontend; and an optional database admin tool behind a profile. Start order is gated on health/completion, not mere dependency: database healthy -> seeder completed successfully -> backend healthy -> frontend.
- **Fail-closed secrets.** Database password, JWT secret, and the admin tool password are supplied via required-variable interpolation so the stack aborts when unset. The example template ships those values blank. The optional tool's secret must not be required (see below).
- **Environment contract.** Configured variables must match what the backend reads (database URL, JWT secret, initial admin email/password/nickname, Swagger toggle, frontend origin, LLM base URL/model/key). Swagger is disabled by default. The configured origin must satisfy the backend's CORS allow-list.
- **Least privilege.** The backend runs as a fixed unprivileged uid with a read-only root filesystem plus a tmpfs for temp files, dropped capabilities, and no-new-privileges. A writable named volume is mounted for persisted artifacts, and a one-shot ownership step grants that volume to the backend uid before the backend starts.
- **Specialist artifact persistence.** Specialist application artifacts remain stored under the backend's uploads area and are backed by the named volume so uploads succeed and survive restarts, preserving the existing "protected storage, authenticated access only" decision.
- **Network segmentation.** Two networks: a frontend-facing network (frontend + backend) and a data network (backend + database + seeder + admin tool). The database is not published to the host. The API and the optional admin tool bind to loopback; only the frontend is exposed on all interfaces.
- **Seeder.** Runs as a small dedicated image with pinned dependencies installed at build time, executing as a non-root user with dropped capabilities and a read-only mount of only the required source directories; transient generated artifacts are written to a scratch working directory rather than the repository.
- **Seed completion gate.** The exam and scenario loaders must exit non-zero on a missing input file or on any error (instead of printing and exiting zero), so the "seeder completed successfully" condition genuinely means content loaded. Existing idempotent, non-destructive upsert behavior and count reporting are preserved, per the Database Seeding Contract.
- **Optional tool isolation.** The database admin tool is opt-in via profile, binds to loopback, and its credential is not a required variable, so the core stack starts without it.
- **Content normalization.** The exam parser strips markdown emphasis markers and collapses embedded newlines/whitespace in question stems, options, and explanations before they are persisted, and the regenerated pool replaces the previously cleaned values.
- **Resource controls.** Each service declares CPU/memory limits and bounded JSON log rotation.

## Testing Decisions

- **What makes a good test here:** it asserts externally observable behavior — "from a clean environment, the documented command yields a ready platform that serves the real Learner/Specialist flows" — never the internal mechanics of Compose, scripts, or containers.
- **Single high seam (preferred, existing):** the composed stack exercised end-to-end by the existing Python integration harness over HTTP. This is the same seam CI already uses (Spec 4). Representative assertions: registration/login succeed; the pre-test returns exactly 30 questions across the five domains with the required distribution; session recovery works; a Specialist upload succeeds and the artifact persists across a restart; the seeded question text contains no markdown artifacts.
- **Static seam (supporting):** Compose configuration parses cleanly, and start-up is verified to fail when required secrets are absent (fail-closed behavior).
- **Not a seam for this work:** the Go unit seams in `services/` and `controllers/` are unaffected; no new Go seams are introduced.
- **Prior art:** Spec 4's CI integration harness and `scripts/test_integration.py`; the spec-level integration scripts under `scripts/`.

## Out of Scope

- The `install.sh` bootstrap CLI and its interactive `.env` creation flow (separate spec).
- Cloud-provider deployment specifics (Render, Cloudflare Pages, Neon), HTTPS/TLS termination, DNS, and firewalls.
- LLM provider selection or prompt changes.
- Database schema or migration changes.
- Frontend feature changes beyond rendering already-clean content.
- External secret managers (Docker Secrets, cloud KMS) and non-Docker runtimes (Kubernetes).
- Adding demo/seed learner data.

## Further Notes

- Aligns with the Database Seeding Contract (idempotent, non-destructive, reported counts, preserved runtime/research data) and the Deployment Bootstrap Plan (service startup order, health checks, cloud safety baseline).
- The integrity of the seed completion gate depends entirely on the loaders returning a non-zero exit code on failure — this is the single most important correctness detail.
- The read-only-backend hardening and the uploads volume are one decision: hardening is only acceptable because a dedicated writable volume preserves Specialist artifacts.
- The optional admin tool's credential must not be a required interpolation variable; Compose interpolates inactive-profile services, so requiring it would block the core stack.
