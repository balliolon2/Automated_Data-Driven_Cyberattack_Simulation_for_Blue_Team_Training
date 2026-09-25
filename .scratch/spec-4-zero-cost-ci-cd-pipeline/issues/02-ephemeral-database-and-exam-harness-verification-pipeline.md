# 02: Ephemeral Database & Exam Harness Verification Pipeline

**What to build:** Implement an automated integration testing and build verification job inside GitHub Actions. The job provisions an isolated PostgreSQL 16 container with `pgvector` enabled as a GitHub Actions service container, executes `init.sql` schema definitions, loads the CompTIA Security+ 5-domain exam questions using `scripts/load_exams_to_db.py`, boots the Go backend in the runner environment, and executes the complete Python integration verification test (`scripts/test_integration.py`). This verifies end-to-end that user registration, login, pre-test session generation (4/7/5/8/6 question distribution), session recovery, and zero question overlap remain 100% functional. Additionally, it verifies multi-stage Docker image build for the Go backend and production bundle build for the Vite frontend.

**Blocked by:** 01: Automated Code Quality & DevSecOps Scanning Gate

**Status:** completed

- [x] A GitHub Actions job spins up a service container with image `pgvector/pgvector:pg16` on port `5432`.
- [x] Database schema is initialized cleanly by applying `init.sql` directly into the service container.
- [x] CompTIA Security+ questions from `parsed_questions.json` and security domains are successfully seeded into PostgreSQL via `python scripts/load_exams_to_db.py`.
- [x] Go backend starts up in the background connected to the service container via `DATABASE_URL` with health check verified.
- [x] Python integration test `scripts/test_integration.py` runs and verifies user registration, login, pre-test generation (30 questions across domains 1-5), and zero question overlap with post-test.
- [x] Backend Docker image builds cleanly using `docker build -t soc-trainer-backend:ci .`.
- [x] Frontend production bundle builds cleanly using `npm run build` inside `frontend/`.
- [x] The job fails and blocks the PR if any test assertion or build step fails.
