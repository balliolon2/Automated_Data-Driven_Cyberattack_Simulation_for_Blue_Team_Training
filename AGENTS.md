# Agent Development & Harness Engineering Guide (AGENTS.md)

Welcome, AI Agent. This repository houses the **Automated Data-Driven Cyberattack Simulation for Blue Team Training** platform (SOC Trainer). This document serves as your engineering harness, detailing system design, database schemas, development workflows, testing suites, and coding constraints.

---

## 🛠️ Tech Stack & Architecture Overview

The codebase is split into a Go-based REST API backend and a Vite-backed React frontend.

```
├── controllers/          # Go backend controllers (Auth, Exam logic)
├── models/               # GORM database schemas & struct definitions
├── middlewares/          # JWT authentication middleware
├── frontend/             # React (Vite, TS, Tailwind CSS, Recharts)
├── examination/          # CompTIA Security+ markdown source exam files
├── scripts/              # Ingestion, seeding, and verification scripts
├── init.sql              # Core database schema (extensions, tables, enums)
└── docker-compose.yml    # Development and deployment orchestrator
```

### Port Mappings
- **Go Backend**: Running on port `:8080` (endpoints grouped under `/api`)
- **React Frontend**: Dev server runs on port `:5173` (mapped to port `:80` in production)
- **PostgreSQL**: Standard port `:5432`

---

## 🗄️ Database & Schema Specifications

The database runs **PostgreSQL 16** with the `vector` (pgvector) and `pgcrypto` extensions. Enums and relational integrity are strictly defined in `init.sql`.

### Core Tables & Models

#### 1. Users (`users` / `models.User`)
Tracks learner profile information, tiers, and roles.
- `UserID` (UUID, PK, Default: `gen_random_uuid()`)
- `Email` (VARCHAR, Unique, Not Null)
- `PasswordHash` (VARCHAR, Not Null, hidden in JSON responses)
- `Role` (Enum `role`: 'learner', 'expert', 'admin')
- `CurrentTier` (INT, Default: 1)
- `IsActive` (BOOLEAN, Default: true)

#### 2. Security Domains (`security_domains`)
Defines Kasetsart University's mapped CompTIA Security+ 701 categories:
- `Domain 1` (Security Concepts): 12 questions
- `Domain 2` (Threats, Vulnerabilities, and Mitigations): 22 questions
- `Domain 3` (Security Architecture): 18 questions
- `Domain 4` (Security Operations): 28 questions
- `Domain 5` (Security Management): 20 questions

#### 3. Questions (`questions` / `models.Question`)
Houses the ingested test questions mapping back to domains and MITRE techniques.
- `QuestionID` (UUID, PK)
- `DomainID` (VARCHAR, FK referencing `security_domains.domain_id`)
- `Type` (Enum `question_type`: 'multiple_choice', 'true_false', 'scenario_based', 'drag_drop')
- `QuestionText` (TEXT)
- `Options` (JSONB - maps to Go `any` with GORM JSON serializer)
- `CorrectAnswer` (VARCHAR)
- `Explanation` (TEXT)
- `MitreRef` (VARCHAR, FK referencing `mitre_techniques`)

#### 4. Exam Sessions (`exam_sessions` / `models.ExamSession`)
Coordinates the active status and scoring for test attempts.
- `SessionID` (UUID, PK)
- `UserID` (UUID, FK referencing `users`)
- `ExamType` (Enum `test_type`: 'pre', 'post')
- `Status` (Enum `session_status`: 'in_progress', 'completed', 'abandoned')
- `Score` (FLOAT)
- `TotalQuestions` (INT, Default: 100)

#### 5. Exam Session Questions (`exam_session_questions` / `models.ExamSessionQuestion`)
Maintains the mapping of questions assigned to specific user sessions to verify progress.
- `SessionID` (UUID, FK, composite PK)
- `QuestionID` (UUID, FK, composite PK)
- `OrderIndex` (INT)
- `UserAnswer` (VARCHAR, Nullable)
- `IsCorrect` (BOOLEAN, Nullable)

---

## 🏎️ The Verification & Ingestion Harness

Before submitting any code changes, verify your work against the test harness. 

### 1. Ingestion & Seeding
Raw questions are parsed from markdown and loaded into the database:
- `scripts/parse_exams.py`: Parses files inside `/examination` to generate `parsed_questions.json`.
- `scripts/load_exams_to_db.py`: Connects to PostgreSQL, seeds the `security_domains` table, and inserts parsed questions.

### 2. Integration Verification Suite
`scripts/test_integration.py` verifies the following end-to-end user flows:
- **Registration & Auth**: Checks `/api/register` and `/api/login` endpoints.
- **Pre-Test Initialization**: Requests `/api/exams/pre-test` and confirms it returns exactly 100 questions distributed across the 5 domains in the exact required counts (12/22/18/28/20).
- **Session Recovery**: Calls `/api/exams/session` to ensure exam state is retrieved during drops/refreshes.
- **Submission**: Commits answers to `/api/exams/submit-answer` and scores them.
- **Zero Question Overlap**: Initializes a `/api/exams/post-test` session and verifies that none of the questions used in the Pre-Test appear in the Post-Test.

### How to Run Ingestion and Tests
Ensure the docker stack is running, then execute the following on the host:
```bash
# Set up Python virtual environment (if needed) and install dependencies
pip install psycopg2-binary requests

# Seed the database
python scripts/load_exams_to_db.py

# Run verification tests
python scripts/test_integration.py
```

---

## 📐 Coding Standards & Critical Guidelines

### 1. Go Backend (Gin + GORM)
- **GORM Tags**: Ensure proper tags matching `init.sql`. Every foreign key relation must refer back to its parent model. Use standard GORM JSON serializers for complex fields like `options`.
- **JWT Protection**: Protect any learner actions using `middlewares.AuthMiddleware()`. Ensure claims extract `user_id` reliably.
- **Session Recovery**: When starting a new exam session, check if there is already an `in_progress` session of the same type. If so, return the existing session rather than creating a new one.

### 2. React Frontend
- **Dark Mode First**: The system implements a Graphite theme. Base pages should use deep backgrounds (`bg-graphite-900`) and high-contrast text (`text-graphite-50`).
- **Typography & Font Rules**: Use monospace font (`font-mono` / JetBrains Mono) for logging outputs, telemetry data, and CLI/Terminal emulation elements.
- **State Hydration**: Always invoke `GET /api/exams/session` on page refresh or component mount to hydrate active examination states.

### 3. Zero Overlap Policy
When querying questions for a `post` test, you must query all questions previously answered or generated in the user's `pre` test and exclude them from the random pool. Maintain this constraint at the database layer using SQL `NOT IN` or GORM queries.

---

## 🤖 Instructions for AI Agents
1. **Always Connect to the DB Safely**: Use environment variable `DATABASE_URL`. Do not hardcode credentials.
2. **Preserve Database Schema Integrity**: Never write GORM modifications that alter DB schemas outside of `init.sql` updates.
3. **Verify Local Compiles**:
   - Backend: Run `go build -o /dev/null main.go` to verify compilation.
   - Frontend: Verify TypeScript types inside `/frontend` (e.g., `npm run tsc` or `vite build`).
4. **Run Ingestion and Verification scripts** to prove your code didn't break core features before resolving tasks.
