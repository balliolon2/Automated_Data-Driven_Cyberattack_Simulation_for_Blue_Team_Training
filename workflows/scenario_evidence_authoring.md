# Workflow: Scenario Evidence Authoring

## Objective

Give every key finding a small set of telemetry rows that genuinely prove it, so the learner-facing evidence-selection scoring in `workflows/evidence_log_selection_scoring.md` is fair and explainable.

This workflow covers both sources of scenarios:

1. **Static seed scenarios** (`scripts/generate_scenarios.py` → `scenarios_seed.json`): an LLM drafts evidence mappings; a deterministic validator checks them; the user approves one compact brief; only then are they seeded.
2. **Runtime AI-generated scenarios** (`services/llm_service.go`): the LLM returns findings with exact evidence log messages; the backend maps those messages to indices, validates them, and guarantees a benign context log for False Positive cases.

The loop this workflow makes real: **Scenario authored or changed → evidence drafted → validated → human-approved → seeded → scored.**

---

## 1. Loop Profile

- **Loop Name**: Scenario Evidence Authoring Loop
- **Frequency / Cadence**: Per new or changed scenario, and whenever the seed is regenerated.
- **Primary Actors**:
  - **Drafting Script (`scripts/draft_evidence_mapping.py`)**: asks the LLM for evidence mappings.
  - **Validator (`scripts/validate_evidence_mapping.py`)**: enforces structural and coverage rules.
  - **User (domain reviewer)**: approves or edits the mapping brief.
  - **Seed Pipeline (`scripts/generate_scenarios.py`, `scripts/load_scenarios_to_db.py`)**: applies approved mappings and upserts.
  - **CI**: refuses unapproved or invalid mappings and seeds.
- **Trigger**: a scenario is added or edited, a seed regeneration is requested, or CI validation fails.
- **Checkpoint**: one blocking user approval after all scenarios are drafted and validated.
- **Push right**: the LLM drafts the whole batch, the validator catches structural errors, and the user is asked once, late, with a decision-ready brief.
- **Brief**: a compact table of scenario → finding → chosen log rows and rationale, not the raw JSON.

No AI is used in the learner's Detect & Analyze loop. AI is used here only to draft mappings.

---

## 2. Artifacts

| Path | Purpose | Committed |
|---|---|---|
| `scripts/evidence_mapping.json` | Machine-readable mapping, approval metadata, prompt/model provenance | Yes |
| `scripts/evidence_mapping_report.md` | Human review brief | Yes |
| `scenarios_seed.json` | Final seed with evidence indices merged | Yes |
| `scripts/draft_evidence_mapping.py` | LLM drafting CLI | Yes |
| `scripts/validate_evidence_mapping.py` | Shared validator for mapping and seed | Yes |

### 2.1 `evidence_mapping.json`

```json
{
  "schema_version": 1,
  "generated_at": "2026-10-20T10:00:00Z",
  "model": "gpt-4o-mini",
  "prompt_version": "evidence-mapping-v1",
  "approval": {
    "approved": false,
    "approved_by": null,
    "approved_at": null,
    "mapping_sha256": null
  },
  "scenarios": [
    {
      "scenario_title": "MFA Push Fatigue on Executive Account",
      "is_true_positive": true,
      "findings": [
        {
          "finding_id": "d1_f1",
          "evidence_log_indices": [16, 17, 18],
          "rationale": "Indices 16-18 show repeated denied MFA pushes followed by an allowed push from the same user.",
          "verdict_support": "malicious"
        }
      ]
    }
  ]
}
```

Rules:

- `scenario_title` must match exactly one base scenario.
- `finding_id` must match exactly one finding in that scenario.
- `verdict_support` is one of `malicious`, `benign`, `supporting`.
- `approval.mapping_sha256` is a SHA-256 of the canonicalized `scenarios` array (sorted keys, no approval block). Any edit invalidates approval.

---

## 3. Input Contract

The drafting script reads base scenarios from `generate_scenarios.create_scenarios()` (or an equivalent base JSON) and gives the LLM, per scenario:

```json
{
  "title": "...",
  "description": "...",
  "is_true_positive": true,
  "tp_fp_explanation": "...",
  "key_findings": [
    { "id": "d1_f1", "description": "...", "domain_id": "domain1", "points": 10 }
  ],
  "initial_logs": [
    {
      "index": 0,
      "timestamp": "...",
      "source": "ad_controller",
      "severity": "info",
      "message": "...",
      "src_ip": "...",
      "dst_ip": "...",
      "user": "...",
      "host": "..."
    }
  ]
}
```

The `index` shown to the LLM is the authoritative 0-based position in `initial_logs`.

---

## 4. Authoring Rules

These are both prompt instructions and validator rules.

1. **Pick the minimum decisive set.** Choose the smallest set of logs that would convince a SOC analyst of the finding. One log is acceptable; several are acceptable when the finding needs corroboration.
2. **No positional chunking.** Never select a contiguous range merely because it sits between noise blocks.
3. **True Positive evidence** must directly support the malicious conclusion: the malicious command, the unauthorized access, the exfiltration connection, the persistence mechanism, the policy change, etc.
4. **False Positive evidence** must establish the benign/authorized explanation: approved change ticket, HR travel confirmation, scheduled maintenance window, authorized scanner range, known service account, etc. Do not select only the high-severity alert that opened the case.
5. **Do not select a log just because it is high severity.** Severity is a triage hint, not proof.
6. **Do not select generic noise.** Baseline Kerberos tickets, routine proxy traffic, and unrelated DHCP events are not evidence.
7. **Overlap is allowed.** A log may support more than one finding if it genuinely proves both. The report flags shared logs for the reviewer.
8. **Union cap.** The union of all findings' evidence indices must be **≤ 50% of `initial_logs`** (floor). This keeps click-everything from earning full credit.
9. **Verdict support is mandatory.** A True Positive scenario must have at least one finding marked `malicious`; a False Positive scenario must have at least one finding marked `benign`.
10. **Every finding gets evidence.** A finding with no decisive log is not scorable and must be rewritten or removed.
11. **Rationale is mandatory.** One sentence (≤ 300 characters) citing the log message and why it proves the finding.

---

## 5. LLM Drafting

### 5.1 CLI

```bash
python scripts/draft_evidence_mapping.py \
  --output scripts/evidence_mapping.json \
  --report scripts/evidence_mapping_report.md \
  --concurrency 3 \
  --max-retries 2
```

Optional: `--scenario "<title>"` to redraft one scenario without touching the rest.

Environment:

- `LLM_API_KEY` (required)
- `LLM_BASE_URL` (default `https://api.openai.com/v1`)
- `LLM_MODEL` (default `gpt-4o-mini`)

The script uses the OpenAI-compatible chat completions endpoint, `temperature: 0`, and strict JSON output. It strips Markdown fences if the model adds them.

### 5.2 Prompt contract

System:

> You are a CompTIA Security+ 701 blue-team scenario editor. Assign telemetry logs to key findings. Return JSON only.

User payload: the scenario JSON from §3 plus:

- the authoring rules from §4,
- the exact output schema,
- this instruction: *"Use only the provided log indices. Never invent an index. For a False Positive, the verdict-supporting finding must point to the log that proves the benign or authorized explanation."*

Output schema per scenario:

```json
{
  "scenario_title": "...",
  "findings": [
    {
      "finding_id": "d1_f1",
      "evidence_log_indices": [16, 17],
      "rationale": "...",
      "verdict_support": "malicious"
    }
  ]
}
```

### 5.3 Retry

1. Parse the JSON. On parse failure, retry with the parse error.
2. Run the validator for this scenario. On validation failure, retry with the exact error list and the available log index/message list.
3. After `--max-retries` (default 2), mark the scenario `failed`, keep the error in the report, and exit non-zero.
4. **Never fall back to positional chunking.** A failed scenario requires a redraft or a manual mapping.

---

## 6. Validation Gate

`python scripts/validate_evidence_mapping.py --mapping scripts/evidence_mapping.json`

The validator is shared by the drafting script, `generate_scenarios.py`, and CI. It exits `1` on any error. Approval and `failed`-scenario checks apply only in **mapping mode**; **seed mode** checks scenario content only, so a committed seed can be validated without the mapping file.

### 6.1 Errors

1. `schema_version` missing or unsupported.
2. Scenario set does not match the base set exactly (missing, extra, or duplicate title).
3. Finding set does not match the base set exactly (missing, extra, or duplicate ID).
4. `evidence_log_indices` is empty, not a list of integers, contains duplicates, or is out of range.
5. `rationale` is empty or longer than 300 characters.
6. `verdict_support` is missing or not one of `malicious`, `benign`, `supporting`.
7. True Positive scenario has no `malicious` finding.
8. False Positive scenario has no `benign` finding.
9. Union of evidence indices exceeds `floor(len(initial_logs) * 0.5)`.
10. A finding's evidence set equals every log in the scenario.
11. Finding `points` is ≤ 0.
12. Mapping approval is missing, `approved == false`, or `mapping_sha256` does not match the current scenarios content.
13. A mapping file is being applied while any scenario is marked `failed`.

### 6.2 Warnings (do not fail)

- A log is evidence for more than one finding.
- A False Positive finding's evidence is entirely `high`/`critical` severity (possible alert-only mapping).
- A True Positive finding's evidence is entirely `info`/`low` severity (unlikely to prove malice).
- A finding's evidence spans unrelated log sources without a rationale that explains the correlation.
- Union is between 40% and 50% (close to the cap).

### 6.3 Seed mode

```bash
python scripts/validate_evidence_mapping.py --seed scenarios_seed.json
```

Checks every scenario's `expected_outcomes.key_findings[]` directly, without requiring the mapping file. CI runs this on every PR so a hand-edited seed cannot bypass the gate.

---

## 7. Human Review Brief

`scripts/evidence_mapping_report.md` is the checkpoint artifact. It must be readable in under two minutes per scenario and must not require opening the JSON.

Structure:

```markdown
# Evidence Mapping Review

Batch: 15 scenarios · Model: gpt-4o-mini · Prompt: evidence-mapping-v1
Mapping SHA-256: <hash>
Status: PENDING APPROVAL

## 1. <Scenario Title> — TRUE POSITIVE
Evidence union: 5 / 45 logs (11%)

| Finding | Evidence logs | Rationale | Support |
|---|---|---|---|
| d1_f1 — Identified push fatigue indicators | #16 10:04 edr/high — "MFA push denied for..."<br>#17 10:05 edr/high — "MFA push denied for..." | Three denied pushes followed by an allowed push prove fatigue. | malicious |

Flags: none

## 2. <Scenario Title> — FALSE POSITIVE
Evidence union: 3 / 45 logs (7%)

| Finding | Evidence logs | Rationale | Support |
|---|---|---|---|
| d2_f3 — Correlated scanner IP with change log | #21 09:12 change_mgmt/info — "CR-4012 approved..."<br>#22 09:14 firewall/info — "10.0.99.0/24..." | The approved change and scanner range prove authorized activity. | benign |

Flags: shared log #22 also supports d2_f1
```

Rules:

- Group by scenario; show verdict and union percentage.
- Include index, timestamp, source, severity, and a truncated message (≤ 140 characters) for each chosen log.
- Show the rationale and verdict support.
- Show all warnings.
- For False Positive scenarios, explicitly label the benign context log so the reviewer can verify it is not just the alert.
- Keep the raw mapping out of the brief; link to the JSON path instead.

### 7.1 Approval

The user reviews the brief and either:

- approves it, or
- requests a redraft of specific scenarios, or
- edits `evidence_mapping.json` directly.

Approval is recorded by:

```bash
python scripts/draft_evidence_mapping.py --approve --approved-by "<nickname>"
```

This sets:

```json
"approval": {
  "approved": true,
  "approved_by": "<nickname>",
  "approved_at": "<ISO-8601>",
  "mapping_sha256": "<hash>"
}
```

Any subsequent edit to `scenarios` invalidates the hash, and CI refuses to apply it. Do not seed an unapproved mapping.

---

## 8. Apply and Seed

1. `python scripts/generate_scenarios.py --mapping scripts/evidence_mapping.json`
   - loads base scenarios,
   - validates the mapping (including approval),
   - copies each finding's `evidence_log_indices` and `verdict_support` into `expected_outcomes.key_findings[]`,
   - writes root `scenarios_seed.json`.
2. `python scripts/load_scenarios_to_db.py`
   - upserts every scenario, including `expected_outcomes`.
3. **Delete `assign_evidence_indices()`** and any positional-chunking fallback. There is no automatic evidence assignment.

CI:

```bash
python scripts/validate_evidence_mapping.py --mapping scripts/evidence_mapping.json
python scripts/validate_evidence_mapping.py --seed scenarios_seed.json
```

Both must pass. The mapping, report, and seed are committed together so the approval is auditable.

---

## 9. Runtime AI-Generated Scenarios

Static seeds are authored ahead of time. Runtime AI scenarios are authored at generation time and need the same guarantees.

### 9.1 Output contract

Extend `aiMicroScenario` (`services/llm_service.go`):

```go
type aiFinding struct {
    Description         string   `json:"description"`
    DomainID            string   `json:"domain_id"`
    Points              float64  `json:"points"`
    EvidenceLogMessages []string `json:"evidence_log_messages"`
    VerdictSupport      string   `json:"verdict_support"` // malicious | benign | supporting
    Explanation         string   `json:"explanation"`
}

type aiMicroScenario struct {
    // existing fields...
    Findings []aiFinding `json:"findings,omitempty"`
}
```

Keep the legacy `KeyFinding` field for backward compatibility. If `Findings` is empty, synthesize one finding from `KeyFinding`.

### 9.2 Prompt rules

- Return **1–4 findings**.
- Every finding must include `evidence_log_messages` copied **exactly** from the generated log messages.
- True Positive: at least one finding is `malicious`.
- False Positive: at least one finding is `benign` and points to a benign context log (approved change, HR travel, scheduled maintenance, authorized scan).
- Do not invent log messages.
- Prefer the minimum decisive set; union must stay ≤ 50% of logs.

### 9.3 Backend mapping

`synthesizeFullScenario`:

1. Build the log set. For False Positive scenarios, guarantee at least one benign context log:
   - If the LLM supplied a benign message that is not in the log set, append a log with `source: "change_management"` or `"hr_system"`, `severity: "info"`, and that message.
   - If it supplied none, synthesize: `CHANGE_MANAGEMENT: Approved change CR-<id> authorizes the observed activity on <target_host> during this window.`
2. Build an exact-message → index map. Fall back to trimmed, case-insensitive matching.
3. Map each finding's `evidence_log_messages` to indices. Reject missing, duplicate, or out-of-range mappings.
4. Validate: 1–4 findings, non-empty evidence, union ≤ 50%, required verdict support.
5. On failure, retry once with the validation errors and the list of available messages.
6. If still invalid, fall back to one finding:
   - True Positive: evidence = the attack-log block (`15..19` in the current 45-log layout).
   - False Positive: evidence = the benign context log index(es).
7. Store the resulting `key_findings` in `expected_outcomes`.

### 9.4 Scenario validation

Extend `ValidateScenarioCandidate` to reject:

- zero findings or more than 4,
- any finding with empty/invalid/duplicate evidence,
- union > 50% of logs,
- a True Positive with no `malicious` verdict support,
- a False Positive with no `benign` verdict support,
- a False Positive with no benign context log in the log set.

### 9.5 Startup remediation

`EnsureScenarioEvidenceQuality` (startup and lazy load):

- Static scenario title matches the approved seed → use the seed's mapping.
- AI row with one finding and evidence `15..19` → valid.
- AI row with no evidence, out-of-range evidence, or a False Positive with no benign context log → set `status='archived'` and do not serve it. The next `StartScenario` generates a fresh scenario.
- Completed sessions are never modified or re-scored.

---

## 10. Verification Criteria (Definition of Done)

### 10.1 Validator unit tests (`scripts/test_evidence_authoring.py`)

1. A valid mapping passes.
2. Missing finding, extra finding, and duplicate finding ID fail.
3. Empty evidence, duplicate index, negative index, and out-of-range index fail.
4. Union at 50% passes; union at 51% fails.
5. TP without `malicious` fails; FP without `benign` fails.
6. FP with only `high` severity evidence emits a warning, not an error.
7. Empty rationale and rationale > 300 chars fail.
8. Unapproved mapping fails; approved mapping with a changed hash fails.
9. Seed mode catches a hand-edited seed with a missing evidence array.

### 10.2 Drafting

1. A deliberate invalid LLM response triggers a retry with the validator errors.
2. After max retries, the scenario is marked `failed` and the command exits non-zero.
3. No code path calls `assign_evidence_indices` or chunks indices positionally.

### 10.3 Seed and DB

1. All 15 static scenarios pass mapping and seed validation.
2. Every scenario has at least one finding with evidence.
3. Every False Positive scenario has a benign context log in a `benign` finding.
4. `load_scenarios_to_db.py` succeeds; `expected_outcomes` is updated.
5. CI fails on an intentionally broken fixture.

### 10.4 Runtime AI

1. A generated True Positive maps its evidence messages to valid indices.
2. A generated False Positive includes a benign context log; if the LLM omits it, the synthesized log is used and marked `benign`.
3. An unmappable LLM response retries, then falls back to a single valid finding.
4. `ValidateScenarioCandidate` rejects 5 findings, empty evidence, and union > 50%.
5. Legacy AI rows with `15..19` still load; invalid FP rows are archived.

### 10.5 Build

```bash
go build -o /dev/null main.go
go test ./...
python scripts/validate_evidence_mapping.py --mapping scripts/evidence_mapping.json
python scripts/validate_evidence_mapping.py --seed scenarios_seed.json
python -m pytest scripts/test_evidence_authoring.py
```

---

## 11. Non-Goals

- No manual per-scenario UI for authoring in this change; the brief and JSON are the interface.
- No change to exam question ingestion.
- No re-scoring of completed sessions.
- No positional or heuristic evidence assignment as a fallback.
- No learner-facing exposure of the mapping.
