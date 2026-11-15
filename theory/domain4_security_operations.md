# Domain 4: Security Operations

> **Domain ID:** `domain4` | **CompTIA Security+ (SY0-701) Exam Weight:** 28% | **Pre/Post-Test Quota:** 8 / 30 Questions

Domain 4 carries the highest weight in both the Security+ exam and day-to-day SOC Tier 1 workflows. It covers security monitoring, SIEM/SOAR telemetry analysis, vulnerability management, digital forensics, and the end-to-end Incident Response lifecycle.

---

## 1. Core Security+ Theory & Exam Concepts

### 1.1 Security Monitoring, SIEM & SOAR
- **SIEM (Security Information and Event Management):** Aggregates, normalizes, and correlates logs from endpoints, firewalls, identity providers, and cloud platforms in real time [R5]. Enables rule-based alerting, anomaly detection, and KQL/query-based threat hunting.
- **SOAR (Security Orchestration, Automation, and Response):** Connects SIEM alerts to automated playbooks for enrichment (threat intel lookups, GeoIP) and repetitive containment actions.
- **Log Aggregation & Time Synchronization:** Centralized syslog/beats collectors paired with **NTP** ensure accurate cross-system event timelines during forensic correlation.

### 1.2 Endpoint, Network & Email Security Operations
- **EDR / XDR:** Continuously monitors process trees, memory injection, registry edits, and network sockets; supports remote host isolation.
- **IDS vs. IPS:** Intrusion Detection Systems passively alert on signatures/anomalies; Intrusion Prevention Systems sit inline to actively drop malicious packets.
- **DNS Filtering, DLP & Email Security:** Blocks C2 domains via sinkholing, prevents unauthorized data exfiltration, and enforces SPF, DKIM, and DMARC authentication.

### 1.3 Vulnerability Management & Patch Operations
- **Discovery & Assessment:** Credentialed vs. non-credentialed vulnerability scans, CVSS v3/v4 severity scoring, and asset criticality prioritization.
- **False Positives vs. False Negatives:** A *False Positive* flags a secure system as vulnerable or malicious; a *False Negative* misses an actual compromise or vulnerability (the highest operational risk).
- **Remediation & Validation:** Prioritized patching, compensating controls when patching is delayed, and rescanning to confirm closure.

### 1.4 The NIST Incident Response Lifecycle (SP 800-61 Rev. 2) [R3]
1. **Preparation:** Building playbooks, deploying SIEM/EDR sensors, configuring centralized logging, and training SOC analysts.
2. **Detection and Analysis:** Triaging alerts, querying telemetry logs, determining scope, validating True vs. False Positives, and collecting Evidence Logs.
3. **Containment, Eradication, and Recovery:** Isolating affected assets, eliminating malware/persistence/root causes, and safely restoring production services.
4. **Post-Incident Activity (Lessons Learned):** Conducting blameless post-mortems, tuning noisy detection rules, and updating playbooks.

### 1.5 Digital Forensics & Evidence Preservation
- **Order of Volatility (RFC 3227):** Collect most volatile evidence first: CPU registers/cache $\rightarrow$ RAM/memory dump $\rightarrow$ running processes & network connections $\rightarrow$ temporary files $\rightarrow$ disk images $\rightarrow$ archival backups.
- **Legal Hold & Chain of Custody:** Documenting who collected, accessed, transferred, and stored forensic artifacts (with cryptographic hashes like SHA-256) to guarantee admissibility and integrity.

---

## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping

### 2.1 Reading Telemetry Logs in Detect & Analyze
Every simulated SOC log typically contains standardized fields: `timestamp`, `source`, `host`, `severity`, `src_ip`, `dst_ip`, `user`, and `message`. Analysts reconstruct attack chains by pivoting across four phases:
1. **Initial Signal:** High-severity EDR/IDS/Auth alert identifying the affected `host` or `user`.
2. **Pivot & Correlate:** Filter by `host` or `src_ip` around the alert timestamp to see what happened immediately before and after.
3. **Confirm Scope:** Look for lateral movement, persistence creation, or outbound data transfer.
4. **Check Context:** Search for benign explanations (change tickets, scheduled jobs, authorized scanners) before concluding.

### 2.2 Common KQL Investigation Patterns

```kql
severity == "high" or severity == "critical"
host == "WS-FIN-04"
source == "edr" and message contains "lsass"
src_ip == "185.220.101.42" or dst_ip == "185.220.101.42"
```

### 2.3 MITRE ATT&CK Mapping [R4]

| Technique ID | Technique Name | Observable Telemetry Indicator |
|---|---|---|
| T1003 | OS Credential Dumping | EDR alert for `procdump`, `mimikatz`, or unauthorized read handle to `lsass.exe` memory |
| T1053 | Scheduled Task/Job | `EventID 4698` or Linux `crontab` modification creating an unexpected recurring binary execution |
| T1110 | Brute Force | High-frequency `EventID 4625` failures from a single external IP followed by a `4624` logon success |
| T1071 | Application Layer Protocol | Periodic fixed-interval HTTPS/DNS beaconing from an endpoint to an uncategorized external domain |

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide

### 3.1 Distinguishing True Positives (TP) from False Positives (FP)
- **True Positive Indicators:** LSASS memory access followed by lateral SMB authentication; newly created scheduled task pointing to `%TEMP%` or `/tmp`; high-volume failed logins transitioning into a successful login and privilege escalation.
- **False Positive Indicators (Benign Context Logs):**
  - **Approved IT Change / Patching:** Elevated PowerShell or service creation accompanied by an SCCM/Intune deployment or approved change-ticket log.
  - **Authorized Endpoint Maintenance:** Scheduled AV full-disk scan or backup agent reading system files during its configured maintenance window.
  - **Benign Operational Anomaly:** Routine developer build script or authorized monitoring probe documented in CMDB/ticketing logs.

### 3.2 3-Phase Playbook Response Principles [R3]
- **Phase 1 — Containment:**
  - *Correct:* Network-isolate the affected endpoint via EDR (preserving volatile memory), block confirmed C2 IPs/domains at the firewall, and disable compromised local/domain accounts.
  - *Distractors to Avoid:* Immediately powering off the server before memory acquisition (destroys volatile RAM evidence) or disabling SIEM alerting to silence the queue.
- **Phase 2 — Eradication:**
  - *Correct:* Terminate malicious processes, delete malicious scheduled tasks/services/binaries, revoke compromised Kerberos tickets/sessions, and reset affected credentials.
  - *Distractors to Avoid:* Deleting log files during cleanup (destroys forensic trail) or leaving persistence tasks intact.
- **Phase 3 — Recovery:**
  - *Correct:* Rebuild or restore compromised hosts from known-good images/clean backups, verify EDR health and log ingestion, and place affected accounts/hosts under heightened SIEM monitoring.
  - *Distractors to Avoid:* Reconnecting an unverified host directly to production without verifying clean state.

---

## 4. Key Takeaways & References

- **Pivot, Don't Guess:** Never select log rows solely because `severity == "high"`. Verify which exact rows prove each Key Finding; every noise pick costs `-2` points.
- **Preserve Volatility:** In containment, prefer *network isolation* over *hard power-off* so forensic RAM evidence is preserved.
- **Verify FP Explanations:** Look explicitly for benign context logs (change tickets, scheduled maintenance, authorized tools) whenever an alert could have an operational explanation.

### References
- `[R2]` CompTIA, Inc. (2023). *CompTIA Security+ Certification Exam Objectives (Exam SY0-701)* — Domain 4.0 Security Operations.
- `[R3]` Cichonski, P., et al. (2012). *NIST SP 800-61 Rev. 2: Computer Security Incident Handling Guide*.
- `[R4]` The MITRE Corporation. (2024). *MITRE ATT&CK Enterprise Matrix* (`T1003`, `T1053`, `T1110`, `T1071`).
- `[R5]` Wikipedia contributors. (2026). *Security information and event management* and *Computer security incident management*.
