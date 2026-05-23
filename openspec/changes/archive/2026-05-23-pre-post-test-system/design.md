## Context

The simulation platform requires a pre-test and post-test feature consisting of 100 questions. These questions need to be parsed from markdown files and persisted. To minimize costs, Gemini Pro will be invoked via Google Auth to evaluate question difficulty. A resilient exam session manager must randomize questions while handling user state persistence.

## Goals / Non-Goals

**Goals:**
- Design a parser that safely extracts questions, choices, explanations, and correct answers from markdown files.
- Design database schema to store questions, including domains and AI-assigned difficulty.
- Design an exam session state capable of saving progress and restoring it after interruptions.
- Incorporate Google Auth for server-to-server Gemini API requests to evaluate questions efficiently.

**Non-Goals:**
- Real-time multiplayer synchronization for the exam.
- Designing the frontend UI layout in this scope (only the APIs are covered).
- Support for other question formats (e.g. matching, essays).

## Decisions

- **Parser Script**: A standalone Python script will be used. It uses regex and structured text parsing to extract the components. This avoids complex parsing logic in Go.
- **AI Integration**: The Python script will utilize the `google-auth` and `google-generativeai` libraries. Google service account credentials will authenticate with Vertex AI / Gemini API to assign difficulty levels.
- **Database Schema**:
  - `Questions` table: id, text, option_a, option_b, option_c, option_d, correct_option, explanation, domain (1-5), difficulty (1-3).
  - `ExamSessions` table: id, user_id, exam_type (pre/post), status (in_progress, completed), created_at, updated_at.
  - `ExamSessionQuestions` table (join table): session_id, question_id, order_index, user_answer, is_correct.
- **Randomization Strategy**: The backend Go API will query the database, using `ORDER BY RANDOM() LIMIT X` for each of the 5 domains, fetching 12, 22, 18, 28, and 20 questions respectively. For post-tests, an additional filter `NOT IN (SELECT question_id FROM ExamSessionQuestions WHERE user_id = X AND exam_type = 'pre')` will ensure the questions are entirely new.

## Risks / Trade-offs

- **Risk: Python Script Authentication Failure** -> *Mitigation*: Ensure the Service Account JSON is correctly scoped and provided as an environment variable (`GOOGLE_APPLICATION_CREDENTIALS`).
- **Risk: Markdown Format Inconsistencies** -> *Mitigation*: The Python parser should be robust, utilizing well-tested regex patterns. The script will log malformed questions to standard error rather than crashing the entire process.
- **Risk: Database Performance with `ORDER BY RANDOM()`** -> *Mitigation*: Since the question pool is relatively small (thousands, not millions), `ORDER BY RANDOM()` in Postgres/MySQL is acceptable. If it grows, we can switch to application-level shuffling.
