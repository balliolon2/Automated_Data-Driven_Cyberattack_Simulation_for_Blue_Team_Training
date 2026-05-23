## Why

To support a robust and fair assessment system for pre-test and post-test examinations, we need a mechanism that randomly selects 100 questions distributed across 5 specific domains, guaranteeing a unique set of questions for each user. Furthermore, the system must retain the examination session state to prevent data loss upon disconnection or page refresh, and it will incorporate Gemini Pro AI via Google Authentication to automatically evaluate and assign difficulty levels to parsed questions, providing a cost-effective and intelligent assessment solution.

## What Changes

- Python script pipeline to parse markdown-formatted question banks into structured JSON data (including questions, choices, answers, and explanations).
- Database schema updates to persist the parsed questions, their respective domains, and AI-evaluated difficulty levels.
- Implementation of Google Authentication to securely interact with the Gemini Pro API without incurring pay-as-you-go costs.
- Gemini Pro AI integration to analyze and assign a difficulty level to each parsed question before database ingestion.
- Exam session state management (e.g., storing the user's progress and current randomized question set) to allow resuming in case of network drops or browser refreshes.
- A randomization engine that generates a unique 100-question test per user per session (Domain 1: 12, Domain 2: 22, Domain 3: 18, Domain 4: 28, Domain 5: 20).
- Post-test generation will strictly randomize a completely new set of questions to prevent memorization from the pre-test.

## Capabilities

### New Capabilities
- `exam-parser`: Parses markdown question banks into structured JSON data.
- `ai-difficulty-evaluator`: Uses Gemini Pro via Google Auth to evaluate and assign difficulty to questions.
- `exam-session-manager`: Handles randomization (100 questions across 5 domains) and session persistence for users.

### Modified Capabilities
- 

## Impact

- **Database**: Introduction of new tables for `Questions`, `ExamSessions`, and `UserResponses`.
- **Backend**: New API endpoints for starting exams, submitting answers, and retrieving session state.
- **Data Pipeline**: Addition of a standalone Python preprocessing script.
- **Authentication**: Integration of Google Auth specifically for invoking the Gemini Pro API.
