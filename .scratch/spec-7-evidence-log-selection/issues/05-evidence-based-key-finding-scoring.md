## Parent

- #16 — Spec 7: Evidence Log Selection and Finding Scoring

## What to build

Submission now scores Key Findings from the server-recorded Evidence Log selection. A Key Finding earns its full points when the selection intersects its authored Evidence Logs; selecting several logs from the same finding does not stack; a log shared by findings credits each; each unique selected log outside every evidence set costs 2 points; the findings component floors at zero. The Evaluation Result exposes earned points, matched evidence, selected indices, noise indices, and the penalty.

The Learner-facing payload stops shipping evidence indices. The old filtered-view auto-discovery is removed, and new Simulation Sessions use the new rubric version. Old Simulation Sessions keep their recorded Evaluation Result and are never re-scored.

## Acceptance criteria

- [ ] The pure evaluator credits a finding on evidence intersection, not on a found-finding ID list.
- [ ] No evidence credits zero; duplicate evidence from the same finding does not stack; shared evidence credits both findings; each unique noise log subtracts two and the findings component floors at zero.
- [ ] Domain credit and penalty attribution, plus total-score clamping, match the specification.
- [ ] Submission replays server-recorded actions; a forged client selection cannot award credit; a mismatch is logged and the server-recorded set wins.
- [ ] The Learner payload exposes only the total number of Key Findings, not evidence indices or finding details.
- [ ] The old filtered-view auto-discovery is removed and the old scoring input is no longer used for new sessions.
- [ ] New sessions record the new rubric version; old sessions are not re-scored and still render.
- [ ] Evaluator table tests and the end-to-end select → recover → submit → result integration test pass.

## Blocked by

- #19
- #20
