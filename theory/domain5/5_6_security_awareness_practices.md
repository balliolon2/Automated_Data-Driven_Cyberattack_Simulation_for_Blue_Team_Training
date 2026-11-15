# 5.6 Implement security awareness practices

## 1. Core Security+ Theory

Human decision-making remains both a primary attack target and a powerful early-warning sensor ([Security Awareness](https://en.wikipedia.org/wiki/Security_awareness), [NIST SP 800-50 Rev. 1](https://csrc.nist.gov/pubs/sp/800/50/r1/final), [CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)):

- **Program Lifecycle & Governance Cadence:**
  - **Initial & Recurring Training:** Mandatory onboarding instruction (covering AUP acknowledgment, data handling, reporting channels) reinforced via annual refresher modules and timely threat advisories.
  - **Role-Based Training:** Tailoring curricula to high-risk roles—**software developers** (OWASP secure coding, secret hygiene), **executives** (whaling/BEC verification), and **IT/privileged admins** (change control, PAM workflows).
- **Targeted Threat Awareness:**
  - **Phishing Campaigns & Simulations:** Controlled simulated phishing emails measuring click rates vs. **reporting rates** ([MITRE ATT&CK T1566](https://attack.mitre.org/techniques/T1566/)) to trigger constructive micro-training rather than punitive fear.
  - **Social Engineering, Password Hygiene & Insider Threats:** Recognizing urgency pretexts, avoiding password reuse via enterprise password managers, disabling risky removable USB/cables, and spotting anomalous or risky operational behavior.
- **Metrics & Culture:** Measuring phish-button report turnaround, declining repeat-click rates, and fostering a blame-free **report-first culture** ([Stanford UIT Report Incident](https://uit.stanford.edu/security/report-incident)).

## 2. SOC & Blue-Team Operational Practice

Fast user phishing submissions directly shrink SOC **Mean Time to Detect (`MTTD`)**, turning trained staff into distributed human sensors whose reports feed SOAR email-purge playbooks before additional employees interact with malicious links:

```kql
EmailEvents
| where LatestDeliveryAction == "Delivered" and ReportId != ""
| summarize ReportedCount = count() by SenderFromAddress, Subject
| order by ReportedCount desc
```

## 3. References (Web Sources)

- [Wikipedia — Security awareness](https://en.wikipedia.org/wiki/Security_awareness)
- [NIST SP 800-50 Rev. 1 — Building a Cybersecurity and Privacy Learning Program](https://csrc.nist.gov/pubs/sp/800/50/r1/final)
- [CISA — Phishing Threat & Response Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)
- [Stanford University UIT — Incident Reporting & Response](https://uit.stanford.edu/security/report-incident)
- [MITRE ATT&CK — Phishing (T1566)](https://attack.mitre.org/techniques/T1566/)
