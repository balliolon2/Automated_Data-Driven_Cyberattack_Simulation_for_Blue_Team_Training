# ADR-0002: Client-side Session Acknowledgment Tracking via LocalStorage

## Status
Accepted

## Context
When a user finishes a training scenario, their results are submitted, and they are redirected to the scenario evaluation results page (`/simulation/result/:id`). Under the hood, the backend `GET /api/simulation/session` endpoint retrieves this recently completed session to support session recovery and prevent page state loss during refresh.

However, when the user clicks "Next Training Scenario" to advance, they are routed back to `/simulation`. If `/simulation` detects the completed session, it automatically redirects the user back to the results page, trapping them in a redirection loop. We need to distinguish between a completed session that is being reviewed (which should route to results) and a completed session that has been acknowledged (which should allow starting the next scenario).

## Decision
Instead of altering the database schema for the `simulation_sessions` table to add an `acknowledged` column (which would require database migrations and updates to GORM models/queries), we track the acknowledgment state on the client side using `localStorage`.

1. When the user clicks the "Next Training Scenario" button on `ScenarioResultPage.tsx`, we save a flag in localStorage: `ack_session_<id> = "true"`.
2. When the `SimulationPage.tsx` loads and checks the active session, it inspects localStorage for the session ID. If `ack_session_<id>` is `"true"`, it bypasses the redirect and proceeds to create/start the next scenario.

## Consequences
- **Zero Database Overhead**: Avoids complex SQL migrations or changes to GORM structs/queries.
- **Robust Session Recovery**: Users can still refresh the page or resume their review, and the system will correctly load their evaluation results as long as they haven't explicitly moved forward.
- **Client-Side Autonomy**: Keeps the session recovery flow clean and self-contained within the frontend router logic.
