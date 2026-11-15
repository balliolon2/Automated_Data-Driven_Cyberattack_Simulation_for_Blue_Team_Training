# 5.5 Explain types and purposes of audits and assessments

## 1. Core Security+ Theory

Independent verification validates that designed controls operate effectively in reality ([IT Security Audit](https://en.wikipedia.org/wiki/Information_technology_security_audit), [Penetration Test](https://en.wikipedia.org/wiki/Penetration_test), [NIST SP 800-115](https://csrc.nist.gov/pubs/sp/800/115/final)):

- **Internal vs. External Audits:**
  - **Internal Audits & Self-Assessments:** Conducted by internal audit committees or security teams to verify policy adherence and fix gaps proactively prior to formal scrutiny.
  - **External Audits & Regulatory Examinations:** Conducted by independent third-party CPAs/assessors (e.g., SOC 2, ISO 27001 certification bodies, PCI QSA) or government regulators to provide impartial attestation.
- **Penetration Testing & Offensive Environments ([NIST SP 800-115](https://csrc.nist.gov/pubs/sp/800/115/final)):**
  - **Unknown Environment (Black Box):** Simulates an external attacker starting with zero internal architecture knowledge.
  - **Partially Known Environment (Gray Box):** Provides limited credentials or diagrams to focus test budget on high-value internal tiers.
  - **Known Environment (White/Crystal Box):** Full source code, network maps, and elevated test access for exhaustive verification.
  - **Reconnaissance:** **Passive** (OSINT, DNS/WHOIS queries without touching target systems) vs. **Active** (port scanning, vulnerability probing).
  - **Physical Penetration Testing & Purple Teaming:** Testing badge cloning/tailgating alongside red-versus-blue joint exercises calibrating SIEM detections.

## 2. SOC & Blue-Team Operational Practice

Authorized penetration tests require documented **Rules of Engagement (`RoE`)**—approved tester IPs, scope windows, and emergency contacts—stored in Benign Context Logs so Blue Team triage can differentiate sanctioned tests from malicious **Network Discovery / Valid Accounts** ([MITRE ATT&CK T1078](https://attack.mitre.org/techniques/T1078/)):

```kql
SecurityAlert
| where TimeGenerated > ago(24h)
| project TimeGenerated, AlertName, CompromisedEntity, Entities, RemediationSteps
```

## 3. References (Web Sources)

- [Wikipedia — Information technology security audit](https://en.wikipedia.org/wiki/Information_technology_security_audit)
- [Wikipedia — Penetration test](https://en.wikipedia.org/wiki/Penetration_test)
- [NIST SP 800-115 — Technical Guide to Information Security Testing and Assessment](https://csrc.nist.gov/pubs/sp/800/115/final)
- [MITRE ATT&CK — Valid Accounts (T1078)](https://attack.mitre.org/techniques/T1078/)
