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
