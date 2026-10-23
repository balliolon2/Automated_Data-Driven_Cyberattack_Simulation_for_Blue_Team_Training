## Parent

- #16 — Spec 7: Evidence Log Selection and Finding Scoring

## What to build

The Specialist submission console shows the evidence trail. Summary rows add evidence-selected count, noise-selected count, and evidence penalty. The detail view lists every selected log, whether it matched a Key Finding or was a false lead, the acceptable Evidence Logs for missed findings, net findings points, and the penalty. The timeline includes toggle and clear actions.

A Specialist can insert the Evidence Review into an Analysis Thread. Legacy submissions fall back to the legacy discovered-findings view and are never re-scored.

## Acceptance criteria

- [ ] Summary rows show findings found and total, evidence selected, noise selected, and evidence penalty.
- [ ] The detail view lists selected logs with snippets, matched findings, missed findings with acceptable evidence, and noise picks with their penalty.
- [ ] The timeline shows every evidence toggle and clear action.
- [ ] The quick-insert Evidence Review uses the same data and is never exposed to Learners.
- [ ] Legacy submissions remain reviewable and are not re-scored.
- [ ] Backend counters and frontend rendering are covered by tests or the documented manual checklist.

## Blocked by

- #21
