# ADR-0004: Consolidated Backend Evaluation Engine and Immutable Result Snapshot

## Status
Accepted

## Context
In the original simulation flow, scoring logic was duplicated across two layers:
1. The Go backend (`scoreSubmission` in `controllers/simulation_controller.go`) computed score percentages and updated user proficiency profiles during `POST /api/simulation/submit`.
2. The React frontend (`reconstructResult` in `frontend/src/pages/ScenarioResultPage.tsx`) re-evaluated raw action telemetry to reconstruct points, findings status, response action accuracy, and per-domain breakdown for UI presentation.

This created architectural friction:
- **Leaked Seam**: Any changes to scoring rules, domain point weighting, or threshold formulas required synchronized code changes in both Go and TypeScript.
- **Drift Risk**: Discrepancies between frontend reconstruction and backend scoring could mislead learners regarding their actual domain proficiency progress.
- **Testing Burden**: Scoring invariants were difficult to test in isolation because backend calculation was tightly coupled with HTTP handlers and database transactions.

## Decision
1. **Dedicated Pure Go Evaluator (`services/evaluator.go`)**:
   Extract all scoring and evaluation business logic into a deep module with a clean interface (`services.Evaluator`). The evaluator accepts pure Go domain structs and returns an immutable `ScenarioResult` without dependencies on HTTP contexts or database transactions.

2. **Persisted Immutable Evaluation Snapshot**:
   When a simulation is submitted, the complete `ScenarioResult` is persisted directly within `simulation_sessions.skill_gap` (JSONB) in PostgreSQL. Subsequent calls to `GET /api/simulation/result/:sessionId` return this finalized evaluation DTO directly.

3. **Transparent Legacy Session Fallback**:
   For sessions created prior to this refactor where `skill_gap` only stored domain scores, the backend evaluator automatically reconstructs the full DTO on-the-fly from historical `session_actions`, guaranteeing zero regression for existing user data.

4. **Pure Presentation Frontend**:
   Eliminate `reconstructResult()` and redundant client-side calculation from `ScenarioResultPage.tsx`. The frontend binds directly to the backend Evaluation Result DTO.

## Consequences
- **High Locality**: All scoring algorithms, point allocations, and domain weighting logic reside in a single backend module.
- **Zero Scoring Drift**: Frontend renders the exact server-certified evaluation metrics.
- **Superior Testability**: The pure evaluator is 100% unit-testable via fast, in-memory table-driven Go tests (`services/evaluator_test.go`).
- **Code Reduction**: Eliminates over 105 lines of duplicate client-side evaluation code.
