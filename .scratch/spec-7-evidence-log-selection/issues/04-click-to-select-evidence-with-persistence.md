## Parent

- #16 — Spec 7: Evidence Log Selection and Finding Scoring

## What to build

A Learner can mark any log row as evidence. Clicking the row, its checkbox, or Space/Enter toggles selection. Selected rows are visually distinct and also appear in a persistent Selected Evidence tray with individual removal and Clear All controls. Every toggle is recorded server-side; the selection is returned with the Simulation Session and restored after refresh, filter changes, and tab switches.

The existing scoring path remains intact in this ticket. Correctness is never shown during selection.

## Acceptance criteria

- [ ] Row click, checkbox, Space, and Enter toggle the same log.
- [ ] Selected rows are visually distinct and the tray lists them with source, severity, timestamp, and message.
- [ ] Individual removal and Clear All work and are recorded server-side.
- [ ] Refresh and session recovery restore the selection; KQL filtering and sorting do not change evidence identity or erase selection.
- [ ] Invalid indices are rejected; index zero is accepted; a completed Simulation Session cannot be changed.
- [ ] Rapid toggles do not collide on step order.
- [ ] The existing scoring and result behavior still works while this slice is in place.

## Blocked by

- None (can start immediately)
