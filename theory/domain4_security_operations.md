# Domain 4: Security Operations

> **Domain ID:** `domain4` | **Domain Scope:** 28% Weight (8 / 30 Pre/Post-Test Questions) | **Sources:** Wikipedia, NIST CSRC, CMU ISO, Stanford UIT, MITRE ATT&CK

Domain 4 represents core Tier-1/Tier-2 SOC operations: security monitoring, SIEM/SOAR telemetry correlation, vulnerability operations, digital forensics, and end-to-end incident handling grounded in [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final), [NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final), the [CMU Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html), and Wikipedia security references.

---

## 1. Core Security+ Theory & Exam Concepts

### 1.1 Security Monitoring, SIEM & SOAR
Described in [Wikipedia: Security Operations Center (SOC)](https://en.wikipedia.org/wiki/Security_operations_center) and [Wikipedia: Security Information and Event Management (SIEM)](https://en.wikipedia.org/wiki/Security_information_and_event_management):
- **SIEM (Security Information and Event Management):** Combines real-time event monitoring (SEM) and historical log management/correlation (SIM) across endpoints, firewalls, identity providers, and cloud platforms to detect threats and support threat hunting.
- **SOAR (Security Orchestration, Automation, and Response):** Links SIEM alerts to repeatable playbooks for indicator enrichment (threat intelligence, GeoIP) and bounded containment actions.
- **Log Aggregation & Time Synchronization:** Centralized syslog/agent collectors synchronized via Network Time Protocol (NTP) preserve accurate cross-system timelines during incident reconstruction.

### 1.2 Endpoint, Network & Email Security Operations
Described in [Wikipedia: Intrusion Detection System](https://en.wikipedia.org/wiki/Intrusion_detection_system):
- **EDR / XDR:** Continuously monitors process trees, credential access, registry modifications, and network sockets, enabling targeted endpoint isolation.
- **IDS vs. IPS:** Intrusion Detection Systems passively inspect traffic and alert on signatures or anomalies; Intrusion Prevention Systems sit inline to actively drop malicious flows.
- **DNS Filtering & Email Authentication:** Sinkholes malicious C2 domains and validates SPF, DKIM, and DMARC records to curb spoofing ([CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)).

### 1.3 Vulnerability Management & Detection Accuracy
- **Scanning & Prioritization:** Credentialed vs. non-credentialed vulnerability scanning, CVSS severity scoring, and asset-criticality triage ([NIST CSRC Glossary](https://csrc.nist.gov/glossary)).
- **False Positives vs. False Negatives:** A *False Positive* flags benign operational activity or a patched host as malicious/vulnerable; a *False Negative* misses an active intrusion or real exposure (highest operational risk).
- **Remediation & Verification:** Coordinated patching, compensating controls during patch freezes, and post-remediation verification scans.

### 1.4 Incident Response Lifecycle (NIST & University Frameworks)
Aligned with [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final), [NIST SP 800-61 Rev. 3 / CSF 2.0 Profile](https://csrc.nist.gov/pubs/sp/800/61/r3/final), [Wikipedia: Computer Security Incident Management](https://en.wikipedia.org/wiki/Computer_security_incident_management), and the [CMU Computer Security Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html):
1. **Preparation:** Building response playbooks, deploying SIEM/EDR sensors, configuring centralized logging, and conducting tabletop drills ([Stanford University IT](https://uit.stanford.edu/security)).
2. **Detection and Analysis:** Triaging alerts, querying telemetry logs, determining scope, validating True vs. False Positives, and collecting Evidence Logs.
3. **Containment, Eradication, and Recovery:** Isolating affected assets, eliminating persistence/root causes, and restoring verified clean services.
4. **Post-Incident Activity (Lessons Learned):** Conducting structured post-incident reviews, tuning noisy detection rules, and updating security controls.

### 1.5 Digital Forensics & Evidence Preservation
Described in [Wikipedia: Digital Forensics](https://en.wikipedia.org/wiki/Digital_forensics):
- **Order of Volatility (RFC 3227):** Acquire fragile evidence first: CPU registers/cache → RAM/memory dump → running processes and network sockets → temporary files → disk images → archival backups.
- **Legal Hold & Chain of Custody:** Documenting who collected, transferred, and verified artifacts (paired with SHA-256 cryptographic hashes) preserves legal admissibility.

---

## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping

### 2.1 Reading Telemetry Logs in Detect & Analyze
Every simulated SOC log typically contains standardized fields: `timestamp`, `source`, `host`, `severity`, `src_ip`, `dst_ip`, `user`, and `message`. Analysts reconstruct attack chains by pivoting across four phases:
1. **Initial Signal:** High-severity EDR/IDS/Auth alert identifying the affected `host` or `user`.
2. **Pivot & Correlate:** Filter by `host` or `src_ip` around the alert timestamp to see what happened immediately before and after.
3. **Confirm Scope:** Look for lateral movement, persistence creation, or outbound data transfer.
4. **Check Context:** Search for benign explanations (change tickets, scheduled jobs, authorized scanners) before concluding.

### 2.2 Common KQL Investigation Patterns

Query syntax aligned with [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/):

```kql
severity == "high" or severity == "critical"
host == "WS-FIN-04"
source == "edr" and message contains "lsass"
src_ip == "185.220.101.42" or dst_ip == "185.220.101.42"
```

### 2.3 MITRE ATT&CK Mapping

All technique IDs link directly to the official [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/):

| Technique ID | Technique Name | Observable Telemetry Indicator |
|---|---|---|
| [T1003](https://attack.mitre.org/techniques/T1003/) | OS Credential Dumping | EDR alert for `procdump`, `mimikatz`, or unauthorized read handle to `lsass.exe` memory |
| [T1053](https://attack.mitre.org/techniques/T1053/) | Scheduled Task/Job | `EventID 4698` or Linux `crontab` modification creating an unexpected recurring binary execution |
| [T1110](https://attack.mitre.org/techniques/T1110/) | Brute Force | High-frequency `EventID 4625` failures from a single external IP followed by a `4624` logon success |
| [T1071](https://attack.mitre.org/techniques/T1071/) | Application Layer Protocol | Periodic fixed-interval HTTPS/DNS beaconing from an endpoint to an uncategorized external domain |

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide

### 3.1 Distinguishing True Positives (TP) from False Positives (FP)
- **True Positive Indicators:** LSASS memory access followed by lateral SMB authentication; newly created scheduled task pointing to `%TEMP%` or `/tmp`; high-volume failed logins transitioning into a successful login and privilege escalation.
- **False Positive Indicators (Benign Context Logs):**
  - **Approved IT Change / Patching:** Elevated PowerShell or service creation accompanied by an SCCM/Intune deployment or approved change-ticket log.
  - **Authorized Endpoint Maintenance:** Scheduled AV full-disk scan or backup agent reading system files during its configured maintenance window.
  - **Benign Operational Anomaly:** Routine developer build script or authorized monitoring probe documented in CMDB/ticketing logs.

### 3.2 3-Phase Playbook Response Principles
Guided by [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final), [CMU Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html), and [Stanford Incident Reporting](https://uit.stanford.edu/security/report-incident):
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

### Authoritative Web References
- **Wikipedia (Open Reference):**
  - [Security Operations Center (SOC)](https://en.wikipedia.org/wiki/Security_operations_center)
  - [Security Information and Event Management (SIEM)](https://en.wikipedia.org/wiki/Security_information_and_event_management)
  - [Intrusion Detection System](https://en.wikipedia.org/wiki/Intrusion_detection_system)
  - [Computer Security Incident Management](https://en.wikipedia.org/wiki/Computer_security_incident_management) & [Digital Forensics](https://en.wikipedia.org/wiki/Digital_forensics)
- **Government & Standards Agencies (NIST & CISA):**
  - [NIST SP 800-61 Rev. 2: Computer Security Incident Handling Guide](https://csrc.nist.gov/pubs/sp/800/61/r2/final)
  - [NIST SP 800-61 Rev. 3: Incident Response & CSF 2.0 Profile](https://csrc.nist.gov/pubs/sp/800/61/r3/final)
  - [CISA Phishing & Threat Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)
- **Universities & Open Knowledge Bases:**
  - [CMU Information Security Office: Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)
  - [Stanford University IT: Report an Incident & SISA Portal](https://uit.stanford.edu/security/report-incident)
  - [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/) (`T1003`, `T1053`, `T1110`, `T1071`)
  - [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/)
