## 1. Database & Schema Updates

- [x] 1.1 Create `Questions` table schema with domain, difficulty, and correct_option fields
- [x] 1.2 Create `ExamSessions` table schema for persisting user test state
- [x] 1.3 Create `ExamSessionQuestions` junction table schema for storing randomized questions per session
- [x] 1.4 Apply database migrations

## 2. Python Markdown Parser & AI Integration

- [x] 2.1 Set up Python script environment (install `google-auth`, `google-generativeai`)
- [x] 2.2 Implement markdown parsing logic (regex extraction for questions, options, explanation)
- [x] 2.3 Implement Gemini API integration for evaluating question difficulty
- [x] 2.4 Write database insertion logic in Python to populate the `Questions` table from the JSON
- [x] 2.5 Run the script across all 5 domain markdown files

## 3. Go Backend - Exam Session APIs

- [x] 3.1 Implement `POST /api/exams/pre-test` to generate and persist a new randomized 100-question session (Domain 1: 12, Domain 2: 22, Domain 3: 18, Domain 4: 28, Domain 5: 20)
- [x] 3.2 Implement `POST /api/exams/post-test` to generate a new post-test session, filtering out questions present in the user's pre-test
- [x] 3.3 Implement `GET /api/exams/session` to fetch current active session for state recovery
- [x] 3.4 Implement `POST /api/exams/submit-answer` to update `ExamSessionQuestions` with user_answer

## 4. Integration & Testing

- [x] 4.1 Unit test the question randomization query logic
- [x] 4.2 Test session persistence by simulating disconnect/reconnect
- [x] 4.3 End-to-End test of the pre-test to post-test workflow ensuring zero question overlap
