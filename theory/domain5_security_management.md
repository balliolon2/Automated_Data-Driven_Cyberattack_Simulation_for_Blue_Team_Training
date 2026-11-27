# Domain 5: Security Program Management and Oversight (3-Minute Essential)

> **Domain 5** | **Exam Weight:** 20% (`6 / 30` Pre/Post-Test Questions) | **Reading Time:** ~3 min

If technical analysts fight fires, Domain 5 writes the fire safety code, calculates insurance budgets, runs evacuation drills, and ensures customer data complies with privacy law.

---

## 1. Core Security+ Theory & Exam Concepts

- **Governance Hierarchy** ([NIST CSF 2.0](https://www.nist.gov/cyberframework)):
  - *Policies* (mandatory top rules: *what* to protect) → *Standards* (exact technical specs like AES-256) → *Procedures / Playbooks* (step-by-step *how* to respond) → *Guidelines* (recommended tips).
  - *Data Roles:* **Data Owner** (accountable executive), **Data Custodian** (admin applying backups/ACLs).
- **Risk Management & Continuity Math** ([NIST SP 800-30](https://csrc.nist.gov/pubs/sp/800/30/r1/final), [Wikipedia: Risk Management](https://en.wikipedia.org/wiki/Risk_management), [Wikipedia: Business Continuity](https://en.wikipedia.org/wiki/Business_continuity)):
  - **4 Risk Treatments:** *Mitigate* (add controls), *Transfer* (cyber insurance / MSSP), *Accept* (executive sign-off), *Avoid* (stop the risky activity).
  - **Formulas:** `SLE = AV × EF` (cost of 1 incident) | `ALE = SLE × ARO` (expected yearly loss).
  - **Continuity Targets:** *RTO* = max tolerable downtime; *RPO* = max tolerable data loss window (sets backup frequency).
- **Privacy, Compliance & Third-Party Risk** ([Wikipedia: GDPR](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)):
  - Privacy laws (GDPR / PDPA) mandate protecting PII and fast breach notification (often 72 hours); standards like **PCI-DSS** protect payment cards and **ISO 27001 / SOC 2** verify controls via independent audits.
  - Check **vendors/contractors** with NDAs, SLAs, right-to-audit clauses, and tightly scoped support windows.

---

## 2. SOC Telemetry, KQL & MITRE ATT&CK

Domain 5 alerts catch audit-log tampering ([T1070](https://attack.mitre.org/techniques/T1070/)), PII uploads to Shadow IT cloud drives ([T1567](https://attack.mitre.org/techniques/T1567/)), and misused vendor VPN accounts ([T1199](https://attack.mitre.org/techniques/T1199/)):

```kql
source == "dlp" or event_id == "1102" or user contains "vendor"
```

- **Watch for:** `EventID 1102` (Windows Security audit log cleared!), bulk PII files going to personal cloud storage, or a contractor accessing databases without a ticket.

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Guide

| True Positive (Real Attack) | False Positive (Safe Routine / Benign Context Log) |
|---|---|
| Vendor account dumps HR PII tables outside any support ticket, or logs are wiped (`1102`) | Vendor remote session matching an **Approved Maintenance/Support Ticket** & SLA window |
| Sensitive PII uploaded to personal Google Drive/Dropbox (Shadow IT) | Authorized **Audit Evidence Export** by the Data Custodian for a compliance audit |

- **Containment:** Suspend the vendor VPN/account immediately, block unsanctioned cloud destinations in DLP/proxy, and preserve remaining logs offline.
- **Eradication:** Rotate shared vendor secrets, audit all queried tables, and enforce time-bound Just-in-Time access.
- **Recovery:** Re-enable vetted access under strict logging, assess PII breach scope for legal notification, and update governance SOPs.

---

## 4. Key Takeaways & References

- **RTO** is downtime limit, **RPO** is data-loss limit, and `EventID 1102` (cleared audit log) is always an immediate red flag!
- Query Reference: [Microsoft Learn: KQL](https://learn.microsoft.com/en-us/kusto/query/)
