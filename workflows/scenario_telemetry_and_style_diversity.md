# Workflow: Scenario Telemetry, Archetype & Briefing Style Diversity (Issue #25)

## Objective

Eliminate the repetitive structural, telemetry, and narrative patterns across Blue Team simulation scenarios (GitHub Issue #25) where learners encountered nearly identical scenarios and SIEM log tables on every run because:

1. **Hardcoded 5-Log Sandwich in Runtime AI Scenarios (`buildScenarioLogs`)**: Every AI scenario requested only a single `attack_log_message` from the LLM and wrapped it inside a fixed 45-log array where indices `15..19` were always the exact same 5 rows (`firewall` port 443 allowed → `edr` attack message → `ids` malicious payload signature → `web_app` HTTP 500 POST anomaly → `siem_correlator` `[CORR-9901]`), flanked by 15 identical `ad_controller` Kerberos TGT logs (`0..14`) and 25 identical `Routine network telemetry event` logs (`20..44`).
2. **Contiguous Evidence Clustering & Monotone Noise in Static Seed Scenarios (`scripts/generate_scenarios.py`)**: All 15 static scenarios generated 15–25 identical `info` noise rows via `for i in range(15)` at the top and bottom, placing all `warning`/`high`/`critical` evidence rows in a single contiguous block in the middle (`15..20`). Learners could scroll to index 15 or click `Filter: High Severity Alerts` to spot the entire evidence block immediately.
3. **Uniform Plot Selection & True Positive Bias**: AI scenario generation received only the broad domain name (`domain1`–`domain5`) without sub-topic rotation, recent session history exclusion, or True Positive / False Positive cadence control, resulting in repetitive web/C2 True Positive plots.
4. **Identical Narrative Footer & Broken SIEM Shortcuts (`SimulationPage.tsx`)**: Every AI scenario appended the same `INVESTIGATION CLUES` block, and the frontend's `Host:` shortcut button queried `scenario.initial_logs.find(l => l.host)?.host`, which selected `logs[0].host` (`WS-WORKSTATION1`, a noise host) instead of the incident's actual target host, alongside hardcoded `Firewall Logs` and `High Severity Alerts` buttons regardless of the scenario's log sources.

The loop this workflow makes real: **Session start → history-aware Sub-Topic/Archetype/TP-FP selection → multi-event incident chain + contextual noise synthesis → chronological interleaving & anti-repetition validation → context-aware SIEM UI investigation.**

---

## 1. Loop Profile

- **Loop Name**: Scenario Telemetry, Archetype & Briefing Style Diversity Loop
- **Frequency / Cadence**:
  - Every new adaptive simulation session (`POST /api/simulation/start`)
  - Every session recovery (`GET /api/simulation/session`)
  - Every static seed regeneration & validation (`scripts/generate_scenarios.py`, `scripts/validate_evidence_mapping.py`)
  - Backend startup quality gate (`EnsureScenarioEvidenceQualityFromSeed`)
- **Primary Actors**:
  - **Archetype & Cadence Selector (`controllers/simulation_controller.go`, `services/llm_service.go`)**: Inspects the learner's prior simulation sessions and selects the next CompTIA SY0-701 Sub-topic (`1.1`–`5.6`), Scenario Archetype, Briefing Style, and target TP/FP verdict.
  - **Hybrid Telemetry Synthesizer (`services/llm_service.go`, `services/scenario_evidence.go`)**: Combines the LLM's 4–8 contextual `incident_logs` with environment-matched background noise packs (`28–46` total logs, including `2–4` decoy warning/medium logs) and interleaves them chronologically.
  - **Static Seed Pipeline & Validator (`scripts/generate_scenarios.py`, `scripts/validate_evidence_mapping.py`)**: Builds interleaved, non-repetitive telemetry for all 15 static scenarios, syncs `scripts/evidence_mapping.json` and `scenarios_seed.json`, and enforces anti-repetition constraints.
  - **SIEM Investigation UI (`frontend/src/pages/SimulationPage.tsx`)**: Renders varied SOC briefing narratives and dynamic, context-aware SIEM filter shortcuts.
- **Trigger**: A learner starts or resumes a simulation session, or a developer/CI runs the seed and validation pipeline.
- **Checkpoint**: None at runtime (fully autonomous). Static seed changes remain governed by the `evidence_mapping.json` hash contract in `workflows/scenario_evidence_authoring.md`.

---

## 2. AI Scenario Generation & Telemetry Synthesis

### 2.1 Expanded `aiMicroScenario` & `aiIncidentLog` Schema (`services/llm_service.go`)

Replace the single-line `attack_log_message` bottleneck with a structured **4–8 event `incident_logs` chain** while keeping `AttackLogMessage` as a backward-compatible fallback for existing unit tests:

```go
type aiIncidentLog struct {
	Source            string `json:"source"`
	Severity          string `json:"severity"` // info | low | warning | medium | high | critical
	Host              string `json:"host"`
	SrcIP             string `json:"src_ip,omitempty"`
	DstIP             string `json:"dst_ip,omitempty"`
	User              string `json:"user,omitempty"`
	Message           string `json:"message"`
	RelativeOffsetSec int    `json:"relative_offset_sec,omitempty"`
}

type aiMicroScenario struct {
	Title             string                   `json:"title"`
	Description       string                   `json:"description"`
	SubTopicID        string                   `json:"sub_topic_id,omitempty"`
	Archetype         string                   `json:"archetype,omitempty"`
	BriefingStyle     string                   `json:"briefing_style,omitempty"`
	IsTruePositive    bool                     `json:"is_true_positive"`
	TpFpExplanation   string                   `json:"tp_fp_explanation"`
	AttackerIP        string                   `json:"attacker_ip"`
	TargetHost        string                   `json:"target_host"`
	PrimaryUser       string                   `json:"primary_user,omitempty"`
	IncidentLogs      []aiIncidentLog          `json:"incident_logs,omitempty"`
	AttackLogMessage  string                   `json:"attack_log_message,omitempty"`
	KeyFinding        string                   `json:"key_finding,omitempty"`
	Findings          []aiFinding              `json:"findings,omitempty"`
	ContainmentAction string                   `json:"containment_action,omitempty"`
	EradicationAction string                   `json:"eradication_action,omitempty"`
	RecoveryAction    string                   `json:"recovery_action,omitempty"`
	ContainmentSteps  []aiResponseActionOption `json:"containment_steps,omitempty"`
	EradicationSteps  []aiResponseActionOption `json:"eradication_steps,omitempty"`
	RecoverySteps     []aiResponseActionOption `json:"recovery_steps,omitempty"`
}
```

### 2.2 Incident Event Chain Contract (`incident_logs`)

1. **Event Count**: Each AI scenario (and each fallback template in `GenerateFallbackScenario`) provides **4 to 8 contextual `incident_logs`** representing the chronological steps of the incident (e.g., initial access/trigger → authentication/policy event → execution/lateral movement or benign authorization context → secondary telemetry).
2. **Source & Severity Diversity**:
   - `source` values must match the scenario's technical domain and archetype (e.g., `okta`, `azure_ad`, `ad_controller`, `aws_cloudtrail`, `s3_access`, `k8s_audit`, `edr`, `sysmon`, `file_integrity`, `dns_server`, `vpn_gateway`, `proxy_gateway`, `nginx`, `postgresql`, `dlp_sensor`, `email_gateway`, `change_management`, `hr_portal`, `itsm_ticketing`, `nids`, `firewall`).
   - Not every incident log may be `high` or `critical`: multi-stage chains should include supporting `info` or `warning` telemetry (such as an initial login, ACL read, DNS lookup, or change ticket) alongside higher-severity alerts.
   - False Positive scenarios (`is_true_positive: false`) MUST include at least one benign context log inside `incident_logs` (e.g., `change_management`, `hr_portal`, `itsm_ticketing`, or verified benign operational telemetry) and MUST NOT emit fake `"IDS Alert: Malicious attack payload signature matched"` or `"SIEM Alert Rule [CORR-9901]"` messages.
3. **Finding Evidence Binding**:
   - Every `findings[].evidence_log_messages` entry must match the exact `message` of one of the scenario's `incident_logs`.
   - Backward compatibility: if a legacy test invokes `synthesizeFullScenario` with `IncidentLogs` empty and `AttackLogMessage` populated, synthesize a contextual 4-event chain varied by `weakestDomain` rather than the banned `CORR-9901` 5-row block.

### 2.3 Environment Noise Packs & Decoy Warnings (`services/scenario_evidence.go`)

Instead of wrapping every scenario in 15 `ad_controller` Kerberos TGT rows + 25 `Routine network telemetry event` rows, `buildScenarioLogs` synthesizes background telemetry as follows:

1. **Variable Total Log Count (`28..46` rows)**:
   - Deterministically compute `targetTotalLogs = 28 + int(seedHash % 19)` (range `28` to `46` inclusive) using an FNV-1a hash of `ai.Title + "|" + ai.TargetHost + "|" + ai.AttackerIP`.
2. **6 Contextual Environment Noise Packs**:
   - Select 2–3 complementary noise packs based on the scenario's `domain_id`, `sub_topic_id`, and `incident_logs` sources:
     1. **Identity & Access Pack (`identity_directory`)**: `ad_controller`, `okta`, `vpn_gateway`, `sso_idp` (MFA challenge completions, group membership syncs, Kerberos service ticket renewals, OAuth token refreshes, password expiry reminders across varied corporate users and workstations).
     2. **Endpoint & Host Pack (`endpoint_workstation`)**: `edr`, `sysmon`, `windows_event`, `file_integrity` (signed binary launches, scheduled Defender signature updates, service control manager state checks, temp file cleanups, driver integrity checks).
     3. **Cloud & Infrastructure Pack (`cloud_infrastructure`)**: `aws_cloudtrail`, `s3_access`, `k8s_audit`, `iam_service` (`AssumeRole` by CI/CD runners, S3 lifecycle bucket transitions, EKS liveness probe events, KMS key rotation checks, CloudWatch metric deliveries).
     4. **Network & Perimeter Pack (`network_perimeter`)**: `firewall`, `nids`, `dns_server`, `proxy_gateway`, `dhcp` (internal DNS A/AAAA lookups, TLS 1.3 outbound CDN connections, DHCP lease renewals, NTP syncs, routine East-West microservice health checks).
     5. **Application & Data Pack (`application_data`)**: `nginx`, `web_app`, `postgresql`, `dlp_sensor`, `email_gateway` (HTTP 200/204 REST API calls, read-replica checkpoints, SPF/DKIM pass inbound emails, DLP policy pass scans on outbound attachments).
     6. **Governance & Operations Pack (`governance_itsm`)**: `change_management`, `itsm_ticketing`, `compliance_scanner`, `backup_service` (nightly incremental backup completions, CIS benchmark posture checks, closed routine helpdesk tickets, patch window heartbeats).
3. **Unique Noise Messages (Zero Repetition Loops)**:
   - Every noise row generated within a scenario is parameterized with distinct realistic hostnames, internal IPs, usernames, endpoints, or request IDs so **no single `message` string appears more than 2 times in a scenario**.
4. **Mandatory Decoy Warning/Medium Logs (`2..4` rows per scenario)**:
   - Every scenario's background noise MUST include **2 to 4 non-evidence decoy logs** with `severity: "warning"` or `"medium"` (`severity_level: 2` or `3`) on unrelated hosts/users, such as:
     - Single mistyped password followed by normal login on an unrelated workstation
     - Routine perimeter firewall drop on external internet background scanner noise (unrelated port/IP)
     - Transient upstream API or database connection pool latency spike (`HTTP 429` or `slow query 1250ms`)
     - TLS certificate 14-day renewal warning or disk usage 82% threshold notice
     - Endpoint agent quarantine of a benign EICAR test file in IT sandbox VLAN
   - None of these decoy warning logs are referenced in `findings[].evidence_log_indices`.

### 2.4 Chronological Interleaving & Index Resolution

1. **Interleaved Placement (No `15..19` Contiguous Block)**:
   - Given `M` incident logs (`4 <= M <= 8`) and `N` noise logs (`total = M + N` in `28..46`), deterministically assign each incident log a target slot index distributed across the timeline (between `10%` and `90%` of the array, with at least `1` to `5` noise logs separating distinct incident stages so at least one pair of incident logs is non-adjacent).
2. **Monotonic Timestamps**:
   - Assign monotonically increasing RFC3339 timestamps (`baseTime` + deterministic jittered step of `4..35` seconds per row) across all `M + N` rows after slot assignment so the table's default chronological order preserves the interleaved positions.
3. **Post-Interleave Evidence Index Mapping**:
   - Run `ensureBenignContextLog` (if needed for False Positive scenarios, inserting the benign context log into an interleaved slot rather than always appending at the end) **before** resolving `findings[].evidence_log_messages` to 0-based indices via `mapAIFindingEvidence`.

---

## 3. Sub-Topic / Archetype Rotation, History Exclusion & TP/FP Cadence

### 3.1 Selection Context (`ScenarioDiversityContext`)

Before calling `GenerateLLMScenario` in `controllers/simulation_controller.go`:

```go
type ScenarioDiversityContext struct {
	TargetSubTopicID      string
	TargetSubTopicTitle   string
	TargetArchetype       string
	TargetBriefingStyle   string
	TargetIsTruePositive  bool
	RecentTitles          []string
	RecentTargetHosts     []string
	RecentAttackerIPs     []string
	RecentSubTopicIDs     []string
}
```

1. **History Query**:
   - Query the learner's last `10` `simulation_sessions` (ordered by `started_at DESC`) and load their corresponding `scenarios` / `scenario_snapshots` to extract:
     - `RecentTitles`
     - `RecentTargetHosts` and `RecentAttackerIPs`
     - Recently covered `SubTopicID`s (or inferred sub-topics)
     - Recent `IsTruePositive` sequence
2. **Sub-Topic & Archetype Rotation**:
   - Fetch the canonical SY0-701 Sub-topics for `weakestDomain` via `services.SubtopicsForDomain(weakestDomain)` (`1.1`–`1.4`, `2.1`–`2.5`, `3.1`–`3.4`, `4.1`–`4.9`, `5.1`–`5.6`).
   - Pick the first sub-topic in that domain not yet seen by the learner (or least recently seen, seeded by `userID + completedCount`).
   - Map the chosen sub-topic and session index to a distinct **Operational Archetype** (e.g., *PKI Certificate Spoofing*, *IAM Privilege Creep*, *Phishing Macro to LOLBin*, *SQLi/API Data Dump*, *Cloud Storage ACL Drift*, *Zero-Trust VLAN Bypass*, *DNS/HTTPS Covert Exfiltration*, *SIEM False Alarm from Approved Change / DR Exercise*, *DLP Unencrypted PII Transfer*, *Supply-Chain Dependency Tampering*).
3. **Controlled TP/FP Cadence (~70% TP / ~30% FP)**:
   - Compute `TargetIsTruePositive` deterministically from the learner's completed session history:
     - If the last 2 sessions were both `is_true_positive == true`, the next session targets `false` (False Positive).
     - If the last session was `is_true_positive == false`, the next session targets `true` (never more than 1–2 False Positives in a row).
     - For a learner's very first session (`completedCount == 0`), target `true` unless seeded otherwise, yielding a balanced ~70% TP / ~30% FP mix across a training loop.
4. **Briefing Style Rotation (5 Styles)**:
   - Rotate across 5 SOC narrative formats based on session count:
     1. `siem_escalation_ticket`: Structured Tier-1 SIEM queue alert summary with rule trigger context and affected asset metadata.
     2. `shift_handover_brief`: Narrative handover note from the outgoing SOC shift analyst describing an unfolding anomaly and initial observations.
     3. `threat_hunting_lead`: Proactive threat-hunt hypothesis triggered by anomalous baseline telemetry across endpoints or cloud logs.
     4. `helpdesk_user_report`: Escalated IT service desk ticket where an employee or system owner reported unusual behavior alongside system telemetry.
     5. `cloud_guardrail_alert`: Automated CSPM / IAM / DLP compliance posture notification flagging a high-risk configuration or data flow event.
5. **Prompt Enforcement**:
   - `GenerateLLMScenario` injects `TargetSubTopicID`, `TargetSubTopicTitle`, `TargetArchetype`, `TargetBriefingStyle`, `TargetIsTruePositive`, and the `DO NOT REPEAT` list (`RecentTitles`, `RecentTargetHosts`, `RecentAttackerIPs`) into the user prompt.
   - Stop appending the repetitive hardcoded `\n\nINVESTIGATION CLUES:\n- Target Host: ...\n- Suspected Attacker IP: ...\n- Alert Vector: ...` block in `synthesizeFullScenarioWithErrors`; instead, embed a style-matched indicator summary only when the description does not already mention `TargetHost`.

---

## 4. Context-Aware SIEM Shortcuts on the Frontend (`SimulationPage.tsx`)

Replace the buggy static shortcut bar in `frontend/src/pages/SimulationPage.tsx` (which called `scenario.initial_logs.find(l => l.host)?.host` and grabbed `WS-WORKSTATION1` at index 0) with a deterministic helper `buildInvestigationShortcuts(scenario)`:

1. **Primary Host Shortcut**:
   - Extract the target host mentioned in `scenario.description` (or, if none is matched in the description text, select the host associated with the highest-severity log entry in `initial_logs` rather than `initial_logs[0]`).
   - Emit button: `Host: <targetHost>` → queries `<targetHost>`.
2. **Primary Actor / Network / Account Indicator Shortcut**:
   - Extract the primary IP (`src_ip` / `dst_ip`) or `user` referenced in `scenario.description` (or from the highest-severity log entry).
   - Emit button: `IP: <ip>` or `User: <user>` → queries `<ip>` or `user == "<user>"`.
3. **Contextual Telemetry Source Shortcut (No More Hardcoded `Firewall Logs`)**:
   - Inspect the distinct `source` values present in `scenario.initial_logs` and pick the primary incident-relevant source (e.g., `okta`, `aws_cloudtrail`, `dns_server`, `edr`, `postgresql`, `email_gateway`, `vpn_gateway`, `change_management`, `firewall`).
   - Emit button: `Source: <source>` → queries `source == "<source>"`.
4. **Contextual Secondary Filter Shortcut**:
   - Pick a second distinct relevant `source` (e.g., `source == "change_management"`, `source == "ad_controller"`, `source == "proxy"`) or a severity filter (`severity == "high"` / `severity == "warning"`) that matches the scenario's actual telemetry distribution.

---

## 5. Static Seed Re-Authoring, Anti-Repetition Gate & Legacy DB Archival

### 5.1 Upgrading the 15 Static Scenarios (`scripts/generate_scenarios.py`)

Refactor `scripts/generate_scenarios.py` so all 15 static scenarios across `domain1`–`domain5`:

1. **Eliminate `for i in range(15)` Identical Message Loops**:
   - Replace every single-message noise loop with rich, realistic multi-source background telemetry (`28–42` total logs per scenario) featuring varied users, hosts, IPs, and operational messages.
2. **Add Decoy Warning/Medium Noise (`2–4` per scenario)**:
   - Include realistic `warning`/`medium` background events that are NOT part of any finding's evidence set so filtering by severity alone never isolates only the answer rows.
3. **Chronological Interleaving**:
   - Interleave each scenario's evidence logs across the timeline (with noise before, between, and after evidence stages) and assign monotonically increasing timestamps.
4. **Synchronize `evidence_mapping.json`, `evidence_mapping_report.md` & `scenarios_seed.json`**:
   - Compute the exact 0-based `evidence_log_indices` after interleaving for every finding (`d1_f1`..`d5_f5`), update `scripts/evidence_mapping.json` (with fresh `mapping_sha256` and approval block) and `scripts/evidence_mapping_report.md`, and regenerate `scenarios_seed.json`.

### 5.2 Anti-Repetition Validation Gate (`validate_evidence_mapping.py` & `services/llm_service.go`)

Enforce anti-pattern rules in **both** Python (`scripts/validate_evidence_mapping.py`) and Go (`ValidateScenarioCandidate`):

1. **Banned Legacy Boilerplate Log Substrings**:
   - Reject any scenario whose `initial_logs` contain any of:
     - `"siem alert rule [corr-9901]: multi-stage incident correlation triggered"`
     - `"routine network telemetry event "`
     - `"kerberos ticket tgt requested for user account user"`
2. **Duplicate Message Cap**:
   - Reject any scenario where the exact same normalized `message` string is repeated more than **3 times** across `initial_logs`.
3. **Anti-Contiguous-Block & Decoy Check (Static Seed Validator)**:
   - In `scripts/validate_evidence_mapping.py`, verify that every scenario has `>= 2` non-evidence logs with `warning` or `medium` severity, and that no single `message` repeats `> 3` times.

### 5.3 Startup Legacy AI Scenario Archival (`EnsureScenarioEvidenceQualityFromSeed`)

Extend `shouldArchiveScenarioEvidence` in `services/scenario_evidence.go`:

- Any active **non-seed** scenario in PostgreSQL (`status != 'archived'` and title not in `scenarios_seed.json`) that fails `ValidateScenarioCandidate` OR contains any banned legacy boilerplate log substring (`CORR-9901`, `Routine network telemetry event`, `Kerberos ticket TGT requested for user account user`) is automatically marked `status = 'archived'` on startup.
- Existing `simulation_sessions` and `scenario_snapshots` are never modified or re-scored.

---

## 6. Verification Criteria (Definition of Done)

1. **Go Unit & Integration Tests (`go test ./...`)**:
   - `buildScenarioLogs` produces variable-length (`28–46`) interleaved log arrays with 6 environment noise packs, `2–4` decoy warning/medium logs, zero `CORR-9901` / `Routine network telemetry event` boilerplate strings, and max message repetition `<= 2`.
   - `synthesizeFullScenarioWithErrors` accurately maps multi-event `incident_logs` to interleaved 0-based `evidence_log_indices` for both True Positive and False Positive scenarios.
   - `GenerateFallbackScenario` for `domain1`..`domain5` produces distinct multi-event telemetry chains, passes `ValidateScenarioCandidate`, and contains zero banned boilerplate strings.
   - `ValidateScenarioCandidate` rejects scenarios containing `CORR-9901` boilerplate or `> 3` identical log messages.
   - `EnsureScenarioEvidenceQualityFromSeed` archives legacy AI scenarios containing the old `CORR-9901` 45-log sandwich while preserving valid scenarios and static seeds.
   - Sub-topic / Archetype / TP-FP cadence selector rotates across SY0-701 sub-topics and targets ~30% False Positive scenarios across sequential sessions.
2. **Python Seed & Validator Tests**:
   - `python scripts/generate_scenarios.py --mapping scripts/evidence_mapping.json` succeeds.
   - `python scripts/validate_evidence_mapping.py --mapping scripts/evidence_mapping.json` passes with zero errors.
   - `python scripts/validate_evidence_mapping.py --seed scenarios_seed.json` passes with zero errors.
   - `python -m pytest scripts/test_evidence_authoring.py` passes (including new tests for duplicate message cap and banned legacy log boilerplate).
3. **Frontend TypeScript & Build Verification**:
   - `npm --prefix frontend run build` succeeds with zero TypeScript errors, and `SimulationPage.tsx` renders dynamic, context-aware SIEM shortcuts matching the scenario's target host, actor IP/user, and actual log sources.
