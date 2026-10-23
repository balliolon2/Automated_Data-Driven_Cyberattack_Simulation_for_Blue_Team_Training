---
title: "Spec 7: Evidence Log Selection and Finding Scoring"
labels: ["ready-for-agent"]
status: "ready-for-agent"
---

# Spec 7: Evidence Log Selection and Finding Scoring

## Problem Statement

A Learner investigates a Scenario, finds the log that clearly proves the incident, and still receives zero points for the Detect & Analyze Key Findings. The current implementation only credits a Key Finding when one transient KQL-filtered result set happens to contain every evidence index assigned to that finding. Seeing the right log is not enough; the Learner must also construct a query that returns the entire assigned range at once. The evidence indices themselves were assigned by positional chunking rather than by what each log proves, so the requirement is both invisible and arbitrary.

The Learner has no way to show their reasoning by marking a specific log as evidence. They cannot tell which selected rows earned credit or why they lost points, and a Specialist reviewing the attempt cannot see what the Learner considered important. False Positive Scenarios are especially unfair: the benign explanation (an approved change, HR travel, a scheduled maintenance window) may sit in a different query result than the alert, so a correct investigation can score nothing.

The result is an Evaluation Result that does not reflect investigation skill, a Detect & Analyze component that feels broken, and a review console that cannot coach the exact mistake.

## Solution

Make evidence selection an explicit Learner action in Detect & Analyze:

1. Every log row is clickable. Clicking toggles it into a visible **Selected Evidence** set; clicking again removes it.
2. A Key Finding is credited in full when the Learner selects any one of that finding's authored Evidence Logs. There is no partial credit inside a finding, and selecting several logs from the same finding does not double-credit.
3. Selecting a log that proves no Key Finding is a false lead and costs 2 points. The findings component floors at 0.
4. Evidence Logs are authored per Key Finding, including the benign context logs that prove a False Positive. Positional chunking is removed.
5. Selections persist server-side across refresh and filter changes. Submission uses the server-recorded selection, not a client-supplied answer.
6. Correctness is revealed only on the result page. During the Scenario the Learner sees selected state and a count, never green/red validation.
7. The Learner-facing Scenario payload no longer ships evidence indices. Only the number of Key Findings is exposed.
8. Specialists see the selected evidence, the findings it matched or missed, and every false lead with its penalty.
9. Scenario authoring becomes a gated pipeline: an LLM drafts evidence mappings, a deterministic validator checks them, a human approves a compact brief, and CI validates the committed seed.

## User Stories

1. As a Learner, I want to click a log row to mark it as evidence, so that I can show which telemetry proves my conclusion.
2. As a Learner, I want to click a selected log again to deselect it, so that I can correct a mistaken selection without reloading.
3. As a Learner, I want selected rows to look clearly selected, so that I can see the evidence set I am building.
4. As a Learner, I want to select evidence from multiple different KQL queries, so that I am not forced to craft one query that returns every relevant row.
5. As a Learner, I want my selection to survive a page refresh or a dropped connection, so that an interrupted investigation is not lost.
6. As a Learner, I want previously selected rows to stay selected when I change the KQL filter, so that filtering does not silently erase my work.
7. As a Learner, I want a persistent Selected Evidence tray, so that I can review everything I have marked without re-running queries.
8. As a Learner, I want to remove an item from the tray directly, so that I can prune the evidence set quickly.
9. As a Learner, I want a Clear All action, so that I can restart my evidence selection without clicking every row.
10. As a Learner, I want to know that unrelated logs cost 2 points, so that I do not click every row hoping for partial credit.
11. As a Learner, I want no correctness feedback while I am selecting, so that the exercise is not reduced to trial-and-error.
12. As a Learner, I want a clear warning when I submit with no evidence selected, so that I understand why my findings score is zero.
13. As a Learner, I want submitting with zero evidence to remain possible, so that I am not blocked from completing an attempt.
14. As a Learner, I want a Key Finding to be credited when I select any one decisive log, so that finding it is enough regardless of which supporting row I clicked first.
15. As a Learner, I want selecting more logs from the same finding not to change the credit, so that the score rewards evidence quality rather than click volume.
16. As a Learner, I want a log that genuinely supports two findings to count for both, so that I am not punished for a correct shared indicator.
17. As a Learner, I want my TP/FP Decision to remain a separate judgment from evidence selection, so that a wrong classification does not erase correct investigative work.
18. As a Learner, I want to see the total number of Key Findings without seeing which logs prove them, so that I know the scope of the task without being handed the answer.
19. As a Learner, I want the evidence indices removed from the browser payload, so that inspecting network traffic does not reveal the answer key.
20. As a Learner, I want a per-finding result showing found versus missed, so that I can see which parts of my investigation succeeded.
21. As a Learner, I want each found finding to show the evidence row that proved it, so that I can connect my click to the credit.
22. As a Learner, I want each missed finding to show the evidence row I failed to select, so that I can learn what a decisive log looked like.
23. As a Learner, I want every false lead listed with its 2-point penalty, so that I can see exactly where the score was lost.
24. As a Learner, I want a net findings score and a gross-versus-penalty breakdown, so that the result is explainable rather than a single opaque number.
25. As a Learner, I want my evidence identity to remain stable when I sort the table, so that sorting never changes which log I selected.
26. As a Learner, I want to toggle evidence with the keyboard, so that I can investigate without switching to the mouse.
27. As a Learner in a False Positive Scenario, I want the benign explanation log to be selectable evidence, so that proving the alert is authorized is a scored skill.
28. As a Learner, I want my old completed sessions to keep their original result, so that a scoring change does not rewrite my history.
29. As a Learner, I want the result page to degrade gracefully for an old session, so that a missing new field does not break the page.

30. As a Specialist, I want to see every log the Learner selected, so that I can review their investigative reasoning rather than only their final score.
31. As a Specialist, I want to see which selected logs matched a Key Finding and which were false leads, so that I can coach the exact misunderstanding.
32. As a Specialist, I want to see the acceptable evidence rows for missed findings, so that I can explain what the Learner overlooked.
33. As a Specialist, I want the evidence penalty shown separately, so that I can distinguish weak investigation from reckless clicking.
34. As a Specialist, I want evidence toggles in the action timeline, so that I can reconstruct how the Learner's evidence set evolved.
35. As a Specialist, I want to insert an Evidence Review into an Analysis Thread, so that my debrief uses the same evidence language as the attempt.
36. As a Specialist, I want legacy submissions to remain reviewable, so that the review console does not break for sessions created before this change.

37. As a Scenario Author, I want every Key Finding to have authored Evidence Logs, so that a Learner is never asked to find proof that does not exist.
38. As a Scenario Author, I want the number of Evidence Logs to vary by Scenario, so that a complex False Positive can require several corroborating logs without a rigid cap.
39. As a Scenario Author, I want a False Positive finding to include a benign context log, so that the verdict is provable from telemetry.
40. As a Scenario Author, I want an LLM to draft the mappings from the actual log messages, so that authoring 15 Scenarios does not become manual indexing.
41. As a Scenario Author, I want the LLM to return log messages rather than invented array positions, so that mappings cannot silently drift when the log set changes.
42. As a Scenario Author, I want a validator to reject empty, duplicated, out-of-range, or over-broad evidence, so that an invalid Scenario never reaches a Learner.
43. As a Scenario Author, I want the evidence union capped at half the logs, so that selecting every row cannot earn full credit.
44. As a Scenario Author, I want a compact review brief with log snippets and rationales, so that I can approve the mappings quickly instead of reading raw JSON.
45. As a Scenario Author, I want approval to be blocking, so that a plausible-looking but semantically wrong mapping cannot be seeded.
46. As a Scenario Author, I want any edit after approval to invalidate the approval hash, so that the approved artifact and the seeded artifact cannot diverge.
47. As a Scenario Author, I want the mapping, report, and seed committed together, so that the approval is auditable.
48. As a Scenario Author, I want CI to validate both the mapping and the final seed, so that a hand-edited seed cannot bypass the gate.
49. As a Scenario Author, I want AI-generated Scenarios to return one to four findings with exact evidence messages, so that dynamic content follows the same scoring contract.
50. As a Scenario Author, I want the backend to synthesize a benign context log when a False Positive generation omits one, so that dynamic FP Scenarios are never unsolvable.

51. As a Maintainer, I want the evidence scoring logic to live in the pure evaluation module, so that the rubric is testable without HTTP or database setup.
52. As a Maintainer, I want the selection to be an append-only session action, so that the final set can be replayed and audited.
53. As a Maintainer, I want toggle and clear actions to be idempotent, so that retries and double-clicks do not corrupt the selection.
54. As a Maintainer, I want step-order allocation serialized per session, so that rapid toggles cannot collide.
55. As a Maintainer, I want submission to replay server-recorded actions, so that a forged client payload cannot award credit.
56. As a Maintainer, I want a new rubric version for new sessions, so that old and new scoring remain distinguishable in research data.
57. As a Maintainer, I want completed sessions to remain immutable, so that the Evaluation Result is a stable research record.
58. As a Maintainer, I want the evidence indices removed from the Learner payload, so that the answer key is not trivially discoverable.
59. As a Maintainer, I want the seed pipeline to have no positional-chunking fallback, so that the old failure mode cannot return unnoticed.
60. As a Maintainer, I want a high-seam integration test for select, recover, submit, and result, so that the persistence contract is verified end to end.

## Implementation Decisions

### Scoring model

- **Evidence Log**: a row in a Scenario's telemetry that proves a Key Finding, identified by its 0-based index in the original log array. This is a new domain term, distinct from Key Finding.
- **Finding credit**: a Key Finding earns its full configured points when the selected evidence set intersects that finding's authored Evidence Logs. No partial credit inside a finding; selecting multiple logs from the same finding does not stack.
- **Noise**: a selected log that belongs to no Key Finding's evidence set. Each unique noise log costs 2 points. The penalty is applied once per unique log.
- **Net findings score**: `max(0, sum of credited finding points - 2 × number of noise logs)`. The findings component cannot go below zero.
- **Shared evidence**: overlap between findings is allowed. A selected log that appears in two findings' evidence sets credits both and is not noise.
- **Domain attribution**: finding credit goes to each finding's domain; the evidence penalty goes to the Scenario's primary domain. Per-domain earned points floor at zero. Domain percentages are diagnostic and may not sum to the overall score.
- **TP/FP independence**: the TP/FP Decision remains a separate component. Finding credit does not require a correct verdict.
- **Overall score**: the existing raw-point model is retained. Total earned is TP/FP points plus net findings points plus response points; total maximum is the sum of positive maxima; the result is clamped to 0–100. No re-normalization to 20/40/40 is introduced.
- **Rubric version**: new Simulation Sessions use `rubric-v2-evidence-click`; completed sessions keep their recorded version.

### Evidence data contract

- Each Key Finding carries an authored `evidence_log_indices` array and a `verdict_support` marker (`malicious`, `benign`, or `supporting`).
- Evidence counts are Scenario-dependent. There is no per-finding cap, but the union of all findings' evidence indices must not exceed 50% of the Scenario's log count.
- A True Positive Scenario must have at least one `malicious` verdict-supporting finding. A False Positive Scenario must have at least one `benign` verdict-supporting finding whose evidence includes a benign context log, not merely the alert that opened the case.
- Points are positive integers. Missing points default to 10; fractional or non-positive points are an authoring error.
- JSONB evidence arrays arrive as generic numeric arrays and must be decoded defensively into integer indices. Invalid indices are logged and treated as unusable, never silently truncated.
- The Learner-facing Scenario payload exposes only `total_findings`; it removes the previous sanitized findings array and all evidence indices, finding descriptions, and verdict markers.

### Selection persistence and audit

- A new session action type, `select_evidence`, is appended to the existing action enum. Existing databases receive an idempotent enum migration; the change is applied outside table migrations.
- Toggle payload: operation `toggle`, the log index, and the selected boolean. Clear payload: operation `clear`.
- The final selection is derived by replaying these actions in step order. Toggle adds or removes; clear empties the set. Malformed payloads are skipped with a warning.
- Step-order allocation is serialized per Simulation Session with a row lock; toggle and submit use the same lock. A unique-violation retry is allowed once.
- Selection is not stored in a new session column. The append-only actions are the source of truth.
- Submission replays the server-recorded actions. A client-supplied selection is accepted only for reconciliation; a mismatch is logged and the server set wins.
- The final selection and its net findings points are recorded in the submit-decision action payload for audit.

### API contracts

- New endpoint: toggle one log as selected or deselected. It validates session ownership, `in_progress` status, and the index range. Index zero is valid and must not be rejected by a required-field validator.
- New endpoint: clear all selected logs for the session in one auditable action.
- The start-session and get-session responses include the selected log indices alongside the session and Scenario.
- The submit request replaces `discovered_findings` with an optional `selected_log_indices` reconciliation field. The server ignores any legacy findings list.
- The result response adds, per finding, earned points, all acceptable evidence indices, and the matched indices; at the result level it adds the selected indices, the noise indices, and the evidence penalty.
- The learner-facing Scenario response adds `total_findings` and removes the sanitized findings payload.

### Learner experience

- The Detect & Analyze table gains a checkbox column and whole-row click selection. Space and Enter toggle the focused row.
- Selected rows use a distinct cyan/indigo treatment, not the green used for correctness after submission.
- A persistent Selected Evidence tray lists selected rows with timestamp, source, severity, and a truncated message. It supports individual removal and Clear All.
- The header counter says “Selected evidence: N logs”. It must not show found-versus-total findings during the Scenario.
- The interface explains that unrelated logs cost 2 points and that correctness is revealed after submission.
- The auto-discovery effect that inferred findings from the filtered result set is removed, along with the discovered-findings state.
- The final submission panel reports the number of selected logs, warns on zero selections, and does not block submission.
- On reload, the selected indices returned by the session endpoint restore the selection before the table renders.

### Result experience

- The result page shows each Key Finding as found or missed, with earned versus available points.
- Found findings show the matched evidence row; missed findings show the acceptable evidence rows that would have proven them.
- A new Evidence Review panel lists every selected log as either proving a finding or “not evidence for any Key Finding — 2-point penalty”.
- The panel shows gross credit, total penalty, and net findings points.
- Legacy sessions without the new fields render the old findings section without the evidence panel.

### Scenario authoring and generation

- Positional chunking of evidence indices is removed. Every Evidence Log is authored against the finding it actually proves.
- Static seed authoring uses an LLM drafting step, a deterministic validator, and one blocking human approval of a compact report. The mapping JSON, report, and final seed are committed together.
- The validator rejects missing or extra scenarios/findings, empty or duplicate or out-of-range indices, empty rationales, invalid verdict support, a union above 50%, non-positive points, missing approval, and an approval hash mismatch. It treats a shared log and an alert-only False Positive evidence set as warnings.
- The approval records the approver, timestamp, and a hash of the mapping. Any mapping edit invalidates the hash and blocks seeding.
- The seed pipeline applies the approved mapping, copies both evidence indices and verdict support into the Scenario's expected outcomes, and writes the seed. There is no positional fallback.
- CI validates the mapping and the final seed.
- Runtime AI generation returns one to four findings, each with exact evidence log messages copied from the generated logs. The backend maps messages to indices, retries once on failure, and then falls back to a single finding. For False Positive Scenarios, the backend synthesizes or attaches a benign context log if the LLM omitted one.
- Scenario validation rejects more than four findings, empty or invalid evidence, a union above 50%, a True Positive without malicious verdict support, and a False Positive without a benign context log.
- Legacy AI rows with one finding and the existing attack-log evidence remain valid. Invalid False Positive rows without a benign context log are archived rather than served.

### Specialist review

- Submission summary rows add evidence-selected count, noise-selected count, and evidence penalty alongside findings found and total findings.
- Submission detail adds the selected indices, per-finding evidence review, noise selections with log snippets, net findings points, and the evidence penalty.
- Summary counters replay the selection: findings found is the count of findings whose evidence intersects the selection; noise count is the selection outside every evidence set; penalty is twice the noise count.
- Legacy sessions fall back to the legacy discovered-findings payload. They are never re-scored.
- The review modal gains an Evidence Review section and a quick-insert template for Analysis Threads. The timeline includes toggle and clear actions.

### Compatibility and migration

- Completed sessions are immutable. Their stored Evaluation Result and rubric version are not recomputed.
- Legacy session reconstruction reads the old discovered-findings payload only when no selection actions exist.
- The removed sanitized-findings payload is a deliberate breaking frontend contract change, shipped in the same release as the new selection UI.
- Evaluation remains a pure backend module per ADR-0004; the frontend remains a presentation layer and does not reconstruct scores.
- Runtime AI generation remains provider-agnostic with the existing static fallback per ADR-0003. Evidence mapping now uses exact log messages instead of LLM-invented indices; provider selection, fallback behavior, and loading experience are unchanged.

## Testing Decisions

- **What makes a good test here:** it asserts externally observable behavior — the score a Learner receives for a given evidence set, the selection returned after a refresh, the review a Specialist sees — not private helpers, React state names, or SQL ordering. Tests should fail if the Learner-visible contract changes, not if the implementation is refactored.
- **Primary seam (existing, highest): the pure evaluation module.** It accepts a Scenario and a submission input and returns the full Evaluation Result. This is the same seam established by ADR-0004 and the existing evaluator tests. Table-driven cases cover: one evidence log credits a finding; no evidence credits zero; duplicate logs from the same finding do not stack; a noise log subtracts two and floors at zero; a shared log credits both findings; domain credit and penalty attribution; total-score clamping; and pre-penalty versus net fields.
- **End-to-end seam (existing): the HTTP simulation integration flow.** Extend the existing Python simulation integration script to cover start → select a log → refresh and recover the same selection → submit → result, and to assert the evidence review. This is the seam that proves persistence, ownership checks, action replay, and the API contract together. It uses the same running-stack harness CI already uses.
- **Authoring seam (new, pure): the evidence validator CLI.** This is a new seam, proposed at the highest point because content integrity cannot be observed through a Learner HTTP flow until an invalid Scenario has already shipped. It validates a mapping file and a final seed file with no database. Table-driven tests cover valid input, missing/extra findings, empty/duplicate/out-of-range indices, the 50% union boundary, missing verdict support, a False Positive without a benign context log, an unapproved mapping, and an approval-hash mismatch. Prior art: none in this repo; keep it a pure function plus a thin CLI.
- **Frontend verification:** no new frontend unit-test harness is introduced. The selection interaction is exercised through the existing HTTP seam, and the UI is verified with the TypeScript build plus a manual checklist: row click, checkbox, keyboard toggle, filter persistence, refresh restoration, zero-selection warning, and the result-page Evidence Review. A React test harness would be a new infrastructure seam disproportionate to this change.
- **Regression tests:** the existing evaluator tests must be rewritten from found-finding IDs to evidence indices; the existing controller shuffle-invariance test must continue to prove that display order does not affect score; the existing integration script must stop sending the legacy findings list.
- **Chosen seams:** two existing seams (pure evaluator, HTTP integration) plus one new pure authoring seam (validator). No new Go HTTP seam, no new frontend harness, and no new database abstraction are introduced.

## Out of Scope

- Changing the TP/FP Decision rubric or the Response Action rubric.
- Scoring KQL query count, query quality, or search history. These remain audit-only.
- Partial credit inside a single Key Finding.
- Re-scoring completed Simulation Sessions.
- A manual per-finding mapping UI for Learners (“which finding does this log prove?”). The system maps selected logs.
- A missed-finding-driven next-Scenario selection loop. It is recorded as a deferred loop in the project notes.
- A generic scenario-authoring UI; the mapping JSON and review brief are the authoring interface.
- A new frontend test runner or component-testing infrastructure.
- Exam question ingestion, pre-test, post-test, and adaptive domain selection changes.
- LLM provider selection, prompt infrastructure, or static-pool fallback changes beyond the evidence-message contract.

## Further Notes

- **Root cause evidence.** The Learner page inferred discovery from the transient filtered result set, so a finding was credited only when every assigned index appeared together. The seed generator distributed evidence indices by splitting the block between the noise sections, not by reading the log messages. This is why a Learner can see the answer and still score zero.
- **Workflow source of truth.** The grilling produced detailed workflow specs: the Learner scoring loop and the Scenario evidence authoring loop. The Specialist analysis workflow was updated to expose the evidence review. This spec is the implementation-facing synthesis of those decisions.
- **Glossary additions.** Add Evidence Log, Evidence Penalty, and Benign Context Log to the project glossary, and clarify that a Key Finding is proven by one or more Evidence Logs rather than discovered through a query window.
- **Scoring contract alignment.** The existing Scoring and Proficiency Contract defines the findings component as found points over total points. This spec refines that component to binary per-finding credit plus a noise penalty, recorded under a new rubric version. Old sessions are not recalculated.
- **ADR alignment.** ADR-0004 is respected: scoring stays in the pure evaluator and the frontend renders the server's Evaluation Result. ADR-0003 is respected and refined: scenarios remain LLM-generated with static fallback, while evidence is now returned as exact log messages mapped server-side rather than as LLM-invented array positions.
- **Originating issue:** #11, “Detect & Analysis Findings can not get score”.
