# Domain 5: Security Program Management and Oversight

> **Domain ID:** `domain5` | **CompTIA Security+ (SY0-701) Exam Weight:** 20% | **Pre/Post-Test Quota:** 6 / 30 Questions

This module covers security governance, risk management, compliance, third-party oversight, auditing, and post-incident accountability tested in Domain 5, connecting organizational oversight directly to SOC escalation and recovery decisions.

---

## 1. Core Security+ Theory & Exam Concepts

### 1.1 Governance Structure & Document Hierarchy
Security governance aligns cybersecurity controls with organizational mission, legal obligations, and board-level oversight [R2, R3]:
- **Policies:** High-level, mandatory management statements defining *what* must be protected (e.g., Acceptable Use Policy, Incident Response Policy, Data Classification Policy).
- **Standards:** Specific, mandatory baselines specifying *requirements* (e.g., minimum 14-character passwords, AES-256 encryption).
- **Procedures (SOPs / Playbooks):** Step-by-step operational instructions defining *how* staff execute tasks consistently.
- **Guidelines:** Recommended best practices that allow contextual flexibility.
- **Roles & Responsibilities:** Data Owner (accountable executive), Data Custodian (technical administrator), Data Controller vs. Data Processor (privacy roles under GDPR/PDPA).

### 1.2 Risk Management & Quantitative Metrics
- **Risk Assessment Steps:** Identify assets $\rightarrow$ identify threats & vulnerabilities $\rightarrow$ determine likelihood & impact $\rightarrow$ select risk treatment.
- **Risk Treatment Strategies:**
  1. *Mitigate (Reduce):* Implement controls to lower likelihood or impact.
  2. *Transfer (Share):* Purchase cyber insurance or outsource to an MSSP.
  3. *Accept:* Document formal management sign-off when mitigation cost exceeds potential loss.
  4. *Avoid:* Discontinue the risky activity or system entirely.
- **Quantitative Risk Formulas:**
  - `SLE` (Single Loss Expectancy) = `AV` (Asset Value) $\times$ `EF` (Exposure Factor)
  - `ALE` (Annualized Loss Expectancy) = `SLE` $\times$ `ARO` (Annualized Rate of Occurrence)
- **Business Continuity Metrics:**
  - `RTO` (Recovery Time Objective): Maximum acceptable downtime before business disruption becomes critical.
  - `RPO` (Recovery Point Objective): Maximum acceptable data loss measured in time (drives backup frequency).
  - `MTTR` / `MTBF`: Mean Time to Repair and Mean Time Between Failures.

### 1.3 Privacy, Compliance & Third-Party Risk
- **Data Protection & Regulatory Frameworks:**
  - *GDPR / PDPA:* Regulate lawful processing of Personal Data / PII, consent, data subject rights, and strict breach notification timelines (e.g., 72 hours).
  - *PCI-DSS:* Protects cardholder data environments (CDE) via encryption, segmentation, and vulnerability testing.
  - *ISO/IEC 27001:* Certifiable Information Security Management System (ISMS) standard emphasizing continual improvement.
- **Third-Party & Supply Chain Risk (TPRM):** Vendor due diligence, right-to-audit clauses, NDAs, SLAs (Service Level Agreements), and monitoring SaaS/vendor access logs.

### 1.4 Audits, Attestation & Post-Incident Governance
- **Internal vs. External Audits:** Independent verification of control effectiveness (SOC 2 Type I/II attestation reports, regulatory inspections).
- **Security Awareness & Tabletop Exercises:** Phishing simulations, role-based training for developers/admins, and incident-response tabletop drills.
- **Post-Incident Governance:** Preserving chain of custody, notifying legal/compliance/DPO stakeholders, and documenting Lessons Learned to close policy or control gaps [R3].

---

## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping

Domain 5 scenarios involve policy violations, unauthorized PII/regulated data access, shadow IT, third-party vendor account abuse, or compliance-driven incident escalation.

### 2.1 Key Log Sources & Fields
- **DLP (Data Loss Prevention) Logs:** `policy_matched` (`PII`, `PCI`, `Confidential`), `user`, `destination` (`USB`, `personal_cloud`, `external_email`), `action` (`BLOCK`/`AUDIT`), `file_count`.
- **Vendor / Third-Party VPN & PAM Logs:** `vendor_account`, `ticket_ref`, `session_duration`, `target_system`, `commands_executed`.
- **Audit & Compliance Logs:** Access to restricted HR/customer PII tables, retention policy tampering, or audit-log clearing (`EventID 1102`).

### 2.2 Common KQL Investigation Patterns

```kql
source == "dlp" and severity == "high"
message contains "PII" or message contains "exfiltration"
user contains "vendor" or event_id == "1102"
```

### 2.3 MITRE ATT&CK Mapping [R4]

| Technique ID | Technique Name | Observable Telemetry Indicator |
|---|---|---|
| T1070 | Indicator Removal | Security audit log cleared (`EventID 1102`), bash history truncated, or SIEM forwarding disabled |
| T1567 | Exfiltration Over Web Service | Bulk upload of customer/PII archives to unsanctioned cloud storage (Shadow IT) flagged by DLP/proxy |
| T1199 | Trusted Relationship | Third-party contractor/vendor VPN account accessing out-of-scope production databases outside an approved SLA ticket |
| T1530 | Data from Cloud Storage | Automated enumeration or download of restricted compliance/PII storage buckets by non-owner service accounts |

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide

### 3.1 Distinguishing True Positives (TP) from False Positives (FP)
- **True Positive Indicators:** Bulk export of unencrypted PII to an unapproved personal cloud drive; vendor account logging in without an active maintenance/support ticket; security audit logs deliberately cleared immediately after sensitive queries.
- **False Positive Indicators (Benign Context Logs):**
  - **Approved Vendor Support Ticket:** Third-party maintenance access backed by a valid change/vendor authorization ticket with matching scope and timeframe.
  - **Authorized Compliance / Audit Export:** Encrypted transfer of audit evidence performed by the designated Data Custodian or external auditor under a documented legal/audit request.
  - **DLP Test / Awareness Exercise:** Triggered DLP policy event originating from an authorized compliance testing account or scheduled phishing simulation.

### 3.2 3-Phase Playbook Response Principles [R3]
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
- **Protect Audit Integrity:** Clearing logs (`T1070`) or tampering with evidence breaks legal compliance and is always a high-priority True Positive indicator.

### References
- `[R2]` CompTIA, Inc. (2023). *CompTIA Security+ Certification Exam Objectives (Exam SY0-701)* — Domain 5.0 Security Program Management and Oversight.
- `[R3]` NIST. (2020). *NIST SP 800-53 Rev. 5* (RA, PL, PS, PT, SR families) & Cichonski, P., et al. (2012). *NIST SP 800-61 Rev. 2*.
- `[R4]` The MITRE Corporation. (2024). *MITRE ATT&CK Enterprise Matrix* (`T1070`, `T1567`, `T1199`, `T1530`).
- `[R5]` Wikipedia contributors. (2026). *Information security*, *Risk management*, and *General Data Protection Regulation*.
