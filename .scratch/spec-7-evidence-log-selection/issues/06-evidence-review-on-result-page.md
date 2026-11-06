## Parent

- #16 — Spec 7: Evidence Log Selection and Finding Scoring

## What to build

The result page explains the evidence score. Each Key Finding shows found or missed with earned versus available points. Found findings show the matched Evidence Log rows; missed findings show the acceptable Evidence Logs that would have proven them. Every selected false lead is listed with its 2-point penalty. The panel shows gross finding credit, total penalty, and net findings points.

Legacy sessions without the new fields render the old findings view without the evidence panel.

## Acceptance criteria

- [ ] Found findings show the matched log rows with timestamp, source, severity, and message.
- [ ] Missed findings show the acceptable log rows that would have proven them.
- [ ] Every noise selection is listed with the 2-point penalty; gross credit, total penalty, and net findings points are visible.
- [ ] No correctness feedback appears during the Scenario; it appears only on the result page.
- [ ] Old sessions render without errors and without the evidence panel.
- [ ] The TypeScript build passes, and the manual checklist covers found, missed, noise, zero-selection, and legacy cases.

## Blocked by

- #21

## Manual checklist (T6)

- [ ] **Found case:** submit with one evidence log from a finding; result page shows `Found`, earned/points, and the matched row (index, timestamp, source, severity, message).
- [ ] **Missed case:** submit without that finding's evidence; result page shows `Missed` and every acceptable evidence row that would have proven it.
- [ ] **Noise case:** select a log outside every evidence set; Evidence Review lists it as "Not evidence for any Key Finding — 2-point penalty" and the summary shows gross credit, `-2` penalty, net findings points.
- [ ] **Zero selection:** submit with no evidence; Evidence Review shows "Selected evidence: 0 logs", zero/zero/zero summary, and a zero-selection notice; no crash.
- [ ] **Legacy session:** open a completed `rubric-v1` session whose result lacks `selected_log_indices`/`noise_log_indices`/per-finding evidence arrays; old findings section renders and the Evidence Review panel is hidden.
- [ ] **Shared evidence:** one selected log proving two findings lists both finding descriptions under "Proved".
- [ ] **Theme/monospace:** telemetry rows render in `font-mono` on the graphite dark theme.
- [ ] **No pre-submit reveal:** Scenario page still shows no correctness; only the result page reveals it.
