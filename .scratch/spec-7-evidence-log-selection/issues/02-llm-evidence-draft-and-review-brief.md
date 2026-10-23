## Parent

- #16 — Spec 7: Evidence Log Selection and Finding Scoring

## What to build

One command drafts Evidence Logs for every Key Finding in all 15 static Scenarios. The LLM receives the Scenario's TP/FP verdict, finding descriptions, and indexed log messages; it returns indices, a rationale, and verdict support. The validator checks every draft; invalid responses are retried with the exact validation errors. The output is a committed mapping draft plus a compact human review brief showing chosen log snippets, rationale, verdict support, union percentage, shared-log warnings, and alert-only False Positive warnings.

No seed is changed in this ticket. The brief is the review artifact for the next ticket.

## Acceptance criteria

- [ ] One command covers all 15 Scenarios and produces a mapping draft and a human-readable review brief.
- [ ] Each report row shows the Scenario, TP/FP verdict, Key Finding, evidence log index, timestamp, source, severity, message snippet, rationale, verdict support, and warnings.
- [ ] Invalid LLM output triggers a retry with the validator errors. After the retry limit, the Scenario is marked failed and the command exits non-zero.
- [ ] No positional chunking or silent fallback is used.
- [ ] The draft is not seeded; the report is the only review surface.
- [ ] Tests use a fake LLM for valid output, invalid output, and retry exhaustion.

## Blocked by

- #17
