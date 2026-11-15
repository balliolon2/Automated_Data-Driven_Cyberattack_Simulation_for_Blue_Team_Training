# 5.3 Explain third-party risk assessment and management processes

## 1. Core Security+ Theory

Modern organizations inherit cyber exposure from cloud vendors, SaaS providers, MSPs, and open-source suppliers ([Supply Chain Security](https://en.wikipedia.org/wiki/Supply_chain_security), [NIST SP 800-161 Rev. 1](https://csrc.nist.gov/pubs/sp/800/161/r1/final)):

- **Vendor Assessment & Due Diligence:** Reviewing vendor **SOC 2 Type II / ISO 27001** independent attestations, running security questionnaires, checking penetration testing summaries, evaluating **supply chain** fourth-party dependencies, and screening for **conflict of interest** prior to procurement.
- **Rules of Engagement & Contractual Safeguards:**
  - **NDA (Non-Disclosure Agreement):** Protects confidential intellectual property and vulnerability findings exchanged during negotiations.
  - **MOU / MOA (Memorandum of Understanding / Agreement):** Expresses bilateral non-binding intent (`MOU`) or formal operational commitment (`MOA`).
  - **MSA (Master Service Agreement) & SOW (Statement of Work):** Establishes umbrella legal terms (`MSA`) and task-specific deliverables/timelines (`SOW`).
  - **SLA (Service Level Agreement):** Quantifies uptime guarantees (`99.95%`), incident notification timelines, and remediation penalties.
  - **BPA (Business Partners Agreement) & Right-to-Audit:** Authorizes periodic compliance verification and security inspections.
- **Ongoing Vendor Monitoring:** Continuously tracking third-party CVE exposures, suspicious B2B federation tokens, and credential hygiene across partner connections.

## 2. SOC & Blue-Team Operational Practice

Trusted supplier channels represent an attractive foothold for **Trusted Relationship** abuse ([MITRE ATT&CK T1199](https://attack.mitre.org/techniques/T1199/)). Blue teams constrain vendor remote accounts with time-boxed access, conditional policies, and behavioral monitoring on B2B integration logs:

```kql
SigninLogs
| where UserType == "Guest" or AppDisplayName has "B2B"
| summarize Logins = count(), DistinctIPs = dcount(IPAddress) by UserPrincipalName, AppDisplayName
| order by Logins desc
```

## 3. References (Web Sources)

- [NIST SP 800-161 Rev. 1 — Cybersecurity Supply Chain Risk Management](https://csrc.nist.gov/pubs/sp/800/161/r1/final)
- [Wikipedia — Supply chain security](https://en.wikipedia.org/wiki/Supply_chain_security)
- [MITRE ATT&CK — Trusted Relationship (T1199)](https://attack.mitre.org/techniques/T1199/)
- [Stanford University UIT — Information Security](https://uit.stanford.edu/security)
