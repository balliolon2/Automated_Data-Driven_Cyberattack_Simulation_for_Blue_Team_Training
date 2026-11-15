# 5.2 Explain elements of the risk management process

## 1. Core Security+ Theory

[Risk Management](https://en.wikipedia.org/wiki/Risk_management) evaluates how threats exploit vulnerabilities to impact business assets, guiding rational security investment ([NIST SP 800-30 Rev. 1](https://csrc.nist.gov/pubs/sp/800/30/r1/final), [Business Impact Analysis](https://en.wikipedia.org/wiki/Business_impact_analysis)):

- **Risk Identification & Assessment Cadence:** Continuous monitoring, recurring annual assessments, ad-hoc trigger reviews (post-breach or M&A), and one-time project assessments.
- **Qualitative vs. Quantitative Risk Analysis:**
  - **Qualitative:** Subjective probability/impact ratings (`Low / Medium / High`) via heat maps and expert judgment.
  - **Quantitative:** Monetary calculation using **Asset Value (`AV`)**, **Exposure Factor (`EF`)**, **Single Loss Expectancy (`SLE = AV × EF`)**, **Annualized Rate of Occurrence (`ARO`)**, and **Annualized Loss Expectancy (`ALE = SLE × ARO`)**. Annualized control cost is justified when it is lower than expected ALE reduction.
- **Risk Appetite, Tolerance & Tracking:** Documented in a central **Risk Register** assigning a **Risk Owner**, Key Risk Indicators (`KRIs`), risk thresholds, and risk posture (**expansionary**, **neutral**, **conservative**).
- **Risk Treatment Strategies:** **Mitigate** (deploy controls), **Transfer** (cyber insurance, SLA outsourcing), **Accept** (document formal risk exception), or **Avoid** (retire risky service).
- **[Business Impact Analysis (BIA)](https://en.wikipedia.org/wiki/Business_impact_analysis):** Defines mission-essential functions alongside **RTO** (Recovery Time Objective), **RPO** (Recovery Point Objective), **MTBF**, and **MTTR**.

## 2. SOC & Blue-Team Operational Practice

SOC prioritization ties CVSS exploit intelligence and SIEM asset criticality straight back to **BIA** tiers and **Risk Registers** so critical payment or identity hosts receive immediate response before low-impact test assets:

```kql
SecurityAlert
| where AlertSeverity in ("High", "Critical")
| summarize AlertCount = count(), LatestTime = max(TimeGenerated) by CompromisedEntity, AlertName
| order by LatestTime desc
```

## 3. References (Web Sources)

- [Wikipedia — Risk management](https://en.wikipedia.org/wiki/Risk_management)
- [Wikipedia — Business impact analysis](https://en.wikipedia.org/wiki/Business_impact_analysis)
- [NIST SP 800-30 Rev. 1 — Guide for Conducting Risk Assessments](https://csrc.nist.gov/pubs/sp/800/30/r1/final)
- [NIST — Cybersecurity Framework (CSF 2.0)](https://www.nist.gov/cyberframework)
