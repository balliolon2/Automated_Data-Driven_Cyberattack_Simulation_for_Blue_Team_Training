## Parent

- #16 — Spec 7: Evidence Log Selection and Finding Scoring

## What to build

A Scenario Author can declare which Evidence Logs prove each Key Finding for a single Scenario, validate that mapping, and see it applied to the generated seed without changing the other Scenarios. The mapping records the Scenario title, finding ID, evidence log indices, a rationale, and verdict support (`malicious`, `benign`, or `supporting`). The validator rejects structural mistakes before anything is seeded.

This is the expand phase: Scenarios that do not yet have an authored mapping keep their current evidence behavior, so the application stays green while the authoring pipeline is built.

## Acceptance criteria

- [ ] A mapping for one Scenario can be validated and applied; the generated seed contains those Evidence Logs and verdict support for that Scenario.
- [ ] The validator rejects a missing or extra Scenario, a missing or extra Key Finding, empty evidence, duplicate indices, out-of-range indices, an empty rationale, invalid verdict support, an evidence union above 50% of logs, and non-positive points. Each error names the Scenario and Key Finding.
- [ ] Scenarios without an authored mapping keep their existing evidence behavior, and the application still starts and seeds.
- [ ] The validator is a pure function with a thin CLI and table-driven tests covering valid and invalid fixtures.
- [ ] No Learner-facing behavior changes in this ticket.

## Blocked by

- None (can start immediately)
