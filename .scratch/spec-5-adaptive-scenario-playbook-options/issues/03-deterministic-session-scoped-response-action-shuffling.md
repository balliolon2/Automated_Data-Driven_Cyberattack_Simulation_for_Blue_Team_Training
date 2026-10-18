# 03: Deterministic Session-Scoped Response Action Shuffling & Scoring Invariance

**What to build:** Eliminate position bias (where the first checkbox in each phase is predictably the correct Response Action) by deterministically shuffling the sanitized Response Actions within each phase (`containment`, `eradication`, `recovery`) using a seed derived from the active Simulation Session (`session_id` + phase name) before returning the Scenario to the learner, ensuring stability across page refreshes and 100% scoring invariance.

**Blocked by:** None (can start immediately)

**Status:** completed

- [x] When sanitizing a Scenario for a Simulation Session, sensitive fields (`is_correct`, `points`, `explanation`, `domain_id`) are stripped and the remaining `{id, label}` Response Action options in each phase (`containment`, `eradication`, `recovery`) are deterministically shuffled using a hash seed derived from `session_id` and the phase name (falling back to `scenario_id` if `session_id` is empty).
- [x] Both `POST /api/simulation/start` (new session creation and active session resumption) and `GET /api/simulation/session` (session recovery on refresh) pass the active `session_id` into the Scenario sanitizer so option ordering never shifts when a learner refreshes the page.
- [x] Different `session_id` values distribute the correct Response Action across different option positions rather than always placing it at index 0.
- [x] Submitting selected Response Action IDs produces the exact same `Evaluation Result` points and `Domain Breakdown` percentages regardless of the shuffled order in which the options were presented.
- [x] Unit tests verify deterministic ordering for identical `session_id` inputs, positional distribution across distinct `session_id` inputs, answer-key stripping, and scoring invariance.
