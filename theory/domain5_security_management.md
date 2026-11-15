# Domain 5: Security Program Management and Oversight

> **Domain ID:** `domain5` | **Domain Scope:** 20% Weight (6 / 30 Pre/Post-Test Questions) | **Sources:** Wikipedia, NIST CSRC, NIST CSF 2.0, CMU ISO, Stanford UIT, MITRE ATT&CK

This module covers cybersecurity governance, risk management, compliance, third-party oversight, and post-incident accountability grounded in [NIST CSF 2.0](https://www.nist.gov/cyberframework), [NIST SP 800-30 Rev. 1](https://csrc.nist.gov/pubs/sp/800/30/r1/final), [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [CMU ISO Governance](https://www.cmu.edu/iso/governance/procedures/incident-response.html), and Wikipedia governance articles.

---

## 1. Core Security+ Theory & Exam Concepts

### 1.1 Governance Structure & Document Hierarchy
Guided by the **Govern** function of [NIST Cybersecurity Framework (CSF) 2.0](https://www.nist.gov/cyberframework) and [Stanford University Information Security](https://uit.stanford.edu/security):
- **Policies:** High-level, mandatory management statements defining *what* must be protected (e.g., Acceptable Use Policy, Incident Response Policy, Data Classification Policy).
- **Standards:** Specific, mandatory technical baselines specifying exact requirements (e.g., AES-256 encryption, MFA enforcement).
- **Procedures (SOPs / Playbooks):** Step-by-step operational instructions defining *how* SOC teams execute incident handling and evidence preservation consistently ([CMU Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)).
- **Guidelines:** Recommended best practices permitting contextual engineering judgment.
- **Roles & Responsibilities:** Data Owner (accountable executive), Data Custodian (technical administrator), and Data Controller vs. Data Processor under privacy statutes.

### 1.2 Risk Management & Continuity Metrics
Defined in [NIST SP 800-30 Rev. 1: Guide for Conducting Risk Assessments](https://csrc.nist.gov/pubs/sp/800/30/r1/final), [Wikipedia: Risk Management](https://en.wikipedia.org/wiki/Risk_management), and [Wikipedia: Business Continuity](https://en.wikipedia.org/wiki/Business_continuity):
- **Risk Assessment Workflow:** Identify assets → identify threats & vulnerabilities → assess likelihood & impact → select risk treatment.
- **Risk Treatment Options:**
  1. *Mitigate (Reduce):* Deploy technical/administrative controls to lower likelihood or impact.
  2. *Transfer (Share):* Purchase cyber insurance or contractually share operational duties with an MSSP.
  3. *Accept:* Record formal executive sign-off when control costs outweigh potential loss.
  4. *Avoid:* Discontinue the hazardous workflow or exposed legacy system entirely.
- **Quantitative Risk Formulas:**
  - `SLE` (Single Loss Expectancy) = `AV` (Asset Value) × `EF` (Exposure Factor)
  - `ALE` (Annualized Loss Expectancy) = `SLE` × `ARO` (Annualized Rate of Occurrence)
- **Business Continuity & Resilience Metrics:**
  - `RTO` (Recovery Time Objective): Maximum acceptable downtime before business impact becomes severe.
  - `RPO` (Recovery Point Objective): Maximum tolerable data loss window (governs backup and replication frequency).
  - `MTTR` / `MTBF`: Mean Time to Repair and Mean Time Between Failures.

### 1.3 Privacy, Compliance & Third-Party Risk
Described in [Wikipedia: General Data Protection Regulation (GDPR)](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation) and [Wikipedia: Data Loss Prevention](https://en.wikipedia.org/wiki/Data_loss_prevention_software):
- **Regulatory & Privacy Obligations:** Under [GDPR](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation) and comparable data protection laws (e.g., PDPA), organizations must protect Personal Data / PII, document lawful processing, and notify supervisory authorities within statutory timelines (e.g., 72 hours) following confirmed personal-data breaches.
- **Security Standards:** PCI-DSS (cardholder data protection) and ISO/IEC 27001 (ISMS certification).
- **Third-Party & Supply Chain Risk:** Vendor security due diligence, right-to-audit clauses, NDAs, SLAs, and auditing vendor VPN/SaaS sessions.

### 1.4 Audits, Exercises & Post-Incident Governance
- **Independent Audits & Attestations:** Internal and external audits (e.g., SOC 2 Type I/II) verify control operation across logging and access reviews ([NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final) CA/AU families).
- **Tabletop Exercises & Reporting:** Conducting recurring IR tabletop drills ([Stanford University IT](https://uit.stanford.edu/security)) and closing policy gaps in post-incident reviews ([NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final)).

---

## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping

Domain 5 scenarios involve policy violations, unauthorized PII/regulated data access, shadow IT, third-party vendor account abuse, or compliance-driven incident escalation.

### 2.1 Key Log Sources & Fields
- **DLP (Data Loss Prevention) Logs:** `policy_matched` (`PII`, `PCI`, `Confidential`), `user`, `destination` (`USB`, `personal_cloud`, `external_email`), `action` (`BLOCK`/`AUDIT`), `file_count`.
- **Vendor / Third-Party VPN & PAM Logs:** `vendor_account`, `ticket_ref`, `session_duration`, `target_system`, `commands_executed`.
- **Audit & Compliance Logs:** Access to restricted HR/customer PII tables, retention policy tampering, or audit-log clearing (`EventID 1102`).

### 2.2 Common KQL Investigation Patterns

Query syntax aligned with [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/):

```kql
source == "dlp" and severity == "high"
message contains "PII" or message contains "exfiltration"
user contains "vendor" or event_id == "1102"
```

### 2.3 MITRE ATT&CK Mapping

All technique IDs link directly to the official [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/):

| Technique ID | Technique Name | Observable Telemetry Indicator |
|---|---|---|
| [T1070](https://attack.mitre.org/techniques/T1070/) | Indicator Removal | Security audit log cleared (`EventID 1102`), bash history truncated, or SIEM forwarding disabled |
| [T1567](https://attack.mitre.org/techniques/T1567/) | Exfiltration Over Web Service | Bulk upload of customer/PII archives to unsanctioned cloud storage (Shadow IT) flagged by DLP/proxy |
| [T1199](https://attack.mitre.org/techniques/T1199/) | Trusted Relationship | Third-party contractor/vendor VPN account accessing out-of-scope production databases outside an approved SLA ticket |
| [T1530](https://attack.mitre.org/techniques/T1530/) | Data from Cloud Storage | Automated enumeration or download of restricted compliance/PII storage buckets by non-owner service accounts |

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide

### 3.1 Distinguishing True Positives (TP) from False Positives (FP)
- **True Positive Indicators:** Bulk export of unencrypted PII to an unapproved personal cloud drive; vendor account logging in without an active maintenance/support ticket; security audit logs deliberately cleared immediately after sensitive queries.
- **False Positive Indicators (Benign Context Logs):**
  - **Approved Vendor Support Ticket:** Third-party maintenance access backed by a valid change/vendor authorization ticket with matching scope and timeframe.
  - **Authorized Compliance / Audit Export:** Encrypted transfer of audit evidence performed by the designated Data Custodian or external auditor under a documented legal/audit request.
  - **DLP Test / Awareness Exercise:** Triggered DLP policy event originating from an authorized compliance testing account or scheduled phishing simulation.

### 3.2 3-Phase Playbook Response Principles
Guided by [NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final), [CMU Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html), and [Stanford Incident Reporting](https://uit.stanford.edu/security/report-incident):
- **Phase 1 — Containment:**
  - *Correct:* Suspend the violating vendor/insider access token, revoke OAuth grants to unsanctioned cloud apps, block the unauthorized external upload destination, and preserve intact audit logs.
  - *Distractors to Avoid:* Deleting employee accounts before preserving chain-of-custody logs or publicly disclosing unverified breach details on social media.
- **Phase 2 — Eradication:**
  - *Correct:* Enforce DLP blocking rules, remove unauthorized third-party integrations, close excess RBAC/vendor permissions (Least Privilege), and rotate exposed shared secrets.
  - *Distractors to Avoid:* Quietly suppressing the compliance alert without remediating the policy violation or notifying governance stakeholders.
- **Phase 3 — Recovery:**
  - *Correct:* Escalate confirmed PII/regulated data exposure to Legal, Compliance, and the Data Protection Officer (DPO) within statutory notification windows (e.g., 72-hour GDPR/PDPA rule), document Chain of Custody, and update vendor/DLP policies during Lessons Learned.
  - *Distractors to Avoid:* Resuming vendor remote access before re-signing SLA/security controls and verifying audit logging.

---

## 4. Key Takeaways & References

- **Governance Extends into Recovery:** Even after technical containment succeeds, Domain 5 Playbook steps frequently require **Chain of Custody**, **Compliance/DPO Escalation**, and **Policy/Lessons Learned Updates**.
- **Check Scope & Authorization:** Vendor and PII access alerts turn on whether an approved ticket, audit charter, or SLA authorization exists in the telemetry logs.
- **Protect Audit Integrity:** Clearing logs ([T1070](https://attack.mitre.org/techniques/T1070/)) or tampering with evidence breaks legal compliance and is always a high-priority True Positive indicator.

### Authoritative Web References
- **Wikipedia (Open Reference):**
  - [Information Security](https://en.wikipedia.org/wiki/Information_security) & [Risk Management](https://en.wikipedia.org/wiki/Risk_management)
  - [Business Continuity](https://en.wikipedia.org/wiki/Business_continuity) & [Data Loss Prevention (DLP)](https://en.wikipedia.org/wiki/Data_loss_prevention_software)
  - [General Data Protection Regulation (GDPR)](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)
- **Government & Standards Agencies (NIST):**
  - [NIST Cybersecurity Framework (CSF) 2.0](https://www.nist.gov/cyberframework)
  - [NIST SP 800-30 Rev. 1: Guide for Conducting Risk Assessments](https://csrc.nist.gov/pubs/sp/800/30/r1/final)
  - [NIST SP 800-53 Rev. 5: Security and Privacy Controls](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
  - [NIST SP 800-61 Rev. 3: Incident Response Recommendations](https://csrc.nist.gov/pubs/sp/800/61/r3/final)
- **Universities & Open Knowledge Bases:**
  - [CMU Information Security Office: Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)
  - [Stanford University Information Security (SISA)](https://uit.stanford.edu/security) & [Report an Incident](https://uit.stanford.edu/security/report-incident)
  - [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/) (`T1070`, `T1567`, `T1199`, `T1530`)
  - [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/)
