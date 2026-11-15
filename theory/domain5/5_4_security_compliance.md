# 5.4 Summarize elements of effective security compliance

## 1. Core Security+ Theory

Security [Regulatory Compliance](https://en.wikipedia.org/wiki/Regulatory_compliance) ensures controls satisfy statutory, contractual, and industry-specific obligations ([GDPR](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)):

- **Compliance Reporting & Monitoring:** Combining continuous automated compliance checks (drift detection, asset patching SLAs) with internal board reporting, external regulatory filings, **due diligence**, **due care**, signed policy **acknowledgments**, and formal executive **attestations**.
- **Privacy Frameworks & Legal Considerations:**
  - **GDPR / Data Privacy Mandates:** Distinguishes **Data Subjects** (identifiable individuals), **Data Controllers**, and **Data Processors**; enforces lawful processing basis, **Right to be Forgotten (erasure)**, breach notification timelines (`72 hours` under GDPR), **Data Inventory**, and defensible retention limits.
  - **Sector Standards:** **PCI-DSS** (payment cardholder data), **HIPAA** (health information), **FERPA** (education), and **SOX** (financial controls).
- **Consequences of Non-Compliance:** Regulatory **fines** (percentage of global turnover), **legal sanctions** and injunctions, contractual penalties, **loss of operating licenses/merchant capability**, and severe **reputational damage** leading to customer attrition.

## 2. SOC & Blue-Team Operational Practice

Audit readiness depends on immutable log retention and fast evidence retrieval during regulatory inquiries. SIEM retention tiers, tamper-evident audit logging, and PII masking controls verify compliance continuously without waiting for yearly external reviews:

```kql
AuditLogs
| where TimeGenerated > ago(30d)
| summarize EventVolume = count() by Category, Result
| order by EventVolume desc
```

## 3. References (Web Sources)

- [Wikipedia — Regulatory compliance](https://en.wikipedia.org/wiki/Regulatory_compliance)
- [Wikipedia — General Data Protection Regulation](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)
- [NIST SP 800-53 Rev. 5 — Audit and Accountability (AU)](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [NIST — Cybersecurity Framework (CSF 2.0)](https://www.nist.gov/cyberframework)
