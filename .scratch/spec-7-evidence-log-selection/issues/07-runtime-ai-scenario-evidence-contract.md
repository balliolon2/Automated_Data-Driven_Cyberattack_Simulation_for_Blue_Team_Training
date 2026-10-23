## Parent

- #16 — Spec 7: Evidence Log Selection and Finding Scoring

## What to build

Dynamically generated Scenarios follow the same evidence contract. The LLM returns one to four Key Findings, each with exact evidence log messages and verdict support. The backend maps messages to indices, validates the evidence union and verdict support, and guarantees a benign context log for False Positive Scenarios, synthesizing one when the LLM omits it. It retries once with validation errors, then falls back to a single valid finding.

Invalid legacy False Positive rows without a benign context log are archived rather than served.

## Acceptance criteria

- [ ] A generated True Positive Scenario has one to four Key Findings, each mapped to valid evidence indices.
- [ ] A generated False Positive Scenario includes a benign context log; when the LLM omits one, the backend synthesizes it and uses it as evidence.
- [ ] Unmappable messages trigger one retry and then a valid single-finding fallback.
- [ ] Validation rejects more than four findings, empty or invalid evidence, an evidence union above 50%, a True Positive without malicious support, and a False Positive without benign support or a benign context log.
- [ ] Legacy AI rows with the existing attack-log evidence remain valid; invalid False Positive rows are archived.
- [ ] The generated Scenario is scored and reviewed through the same evidence path as a static Scenario.

## Blocked by

- #21
