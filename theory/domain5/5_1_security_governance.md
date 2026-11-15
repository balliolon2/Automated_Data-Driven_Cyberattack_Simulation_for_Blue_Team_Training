# 5.1 Summarize elements of effective security governance

## 1. Core Security+ Theory

[Information Security Governance](https://en.wikipedia.org/wiki/Information_security_governance) establishes the leadership accountability, formal rule hierarchy, and data stewardship ensuring security strategy supports organizational objectives ([NIST Cybersecurity Framework 2.0](https://www.nist.gov/cyberframework)):

- **Governance Hierarchy:**
  - **Policies:** Mandatory high-level management intent (Acceptable Use Policy `AUP`, Information Security Policy, Change Management, Business Continuity/DR, SDLC, Onboarding/Offboarding).
  - **Standards:** Mandatory measurable technical baselines (minimum AES-256 encryption, 14-character MFA-backed passwords).
  - **Procedures:** Step-by-step operational checklists (`SOPs`, incident handling playbooks).
  - **Guidelines:** Recommended non-mandatory best practices adaptable to context.
- **Governance Structures & External Considerations:** Steering **boards and committees**, centralized vs. decentralized CISO authority, aligned with regulatory, legal, industry (`PCI-DSS`, `HIPAA`), and jurisdictional data sovereignty requirements.
- **Data Governance Roles & Responsibilities:**
  - **Data Owner:** Senior executive accountable for data classification, retention, and access authorization decisions.
  - **Data Controller:** Entity determining *why* and *how* personal data is processed ([GDPR](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)).
  - **Data Processor:** Third-party processing data strictly on behalf of the controller.
  - **Data Custodian / Steward:** Technical administrator executing backups, encryption, and day-to-day access maintenance.

## 2. SOC & Blue-Team Operational Practice

Governance directly dictates SOC escalation authority and acceptable-use investigations. When DLP flags external file sharing or unauthorized remote tools, policy owners and data stewards authorize remediation steps aligned with organizational privacy mandates and [CMU ISO Governance](https://www.cmu.edu/iso/governance/procedures/incident-response.html):

```kql
AuditLogs
| where Category == "RoleManagement" or OperationName has "policy"
| project TimeGenerated, InitiatedBy, OperationName, TargetResources
| order by TimeGenerated desc
```

## 3. References (Web Sources)

- [Wikipedia — Information security governance](https://en.wikipedia.org/wiki/Information_security_governance)
- [NIST — Cybersecurity Framework (CSF 2.0)](https://www.nist.gov/cyberframework)
- [Wikipedia — General Data Protection Regulation](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)
- [Carnegie Mellon University ISO — Incident Response Governance](https://www.cmu.edu/iso/governance/procedures/incident-response.html)
