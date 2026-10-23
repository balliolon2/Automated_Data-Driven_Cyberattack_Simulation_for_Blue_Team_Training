## Parent

- #16 — Spec 7: Evidence Log Selection and Finding Scoring

## What to build

The author approves the drafted mapping. Approval records the approver, timestamp, and a hash of the mapping; any later edit invalidates the approval. Applying the approved mapping writes Evidence Logs and verdict support into every static Scenario in the seed, and positional chunking is deleted. CI validates both the mapping and the final seed, so an unapproved, invalid, or hand-edited artifact cannot ship.

This is the contract phase of the authoring pipeline: every static Scenario now uses authored evidence, and there is no chunking fallback.

## Acceptance criteria

- [ ] An unapproved mapping cannot be applied; an approved mapping whose content changed is rejected.
- [ ] The generated seed contains authored Evidence Logs and verdict support for every Key Finding in all 15 static Scenarios.
- [ ] Positional chunking and its fallback are removed from the seed pipeline.
- [ ] Every False Positive Scenario has a benign verdict-supporting finding with a benign context log.
- [ ] CI fails on missing or extra findings, empty or duplicate or out-of-range evidence, an evidence union above 50%, missing verdict support, and a deliberately broken seed fixture.
- [ ] The mapping draft, review brief, and generated seed are committed together.

## Blocked by

- #18
