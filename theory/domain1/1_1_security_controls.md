# 1.1 Compare and contrast various types of security controls

## 1. Core Security+ Theory

Security controls are safeguards implemented to manage cyber and physical risk across confidentiality, integrity, and availability. Following [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final) and [Security Controls](https://en.wikipedia.org/wiki/Security_controls) taxonomies, controls are classified along two orthogonal dimensions:

1. **Control Categories (Implementation Mechanism):**
   - **Technical (Logical):** Implemented in software, hardware, or firmware (firewalls, multi-factor authentication, TLS/AES encryption, endpoint detection and response, access control lists).
   - **Managerial (Administrative):** Risk assessments, security governance frameworks, policy creation, vendor due diligence, and vulnerability management planning.
   - **Operational:** Day-to-day procedures executed primarily by people rather than systems alone (security awareness training, incident response execution, configuration review, visitor escorts).
   - **Physical:** Controls protecting physical facilities and assets (badge readers, biometric locks, bollards, mantraps, security guards, CCTV, Faraday cages).
2. **Control Types (Functional Intent):**
   - **Preventive:** Stop unauthorized activity before impact (least privilege ACLs, input validation, IPS blocking).
   - **Deterrent:** Discourage adversaries prior to attempting abuse (warning banners, visible cameras, legal prosecution notices).
   - **Detective:** Identify and alert on anomalous or unauthorized events during or after execution (SIEM correlation alerts, IDS sensors, file integrity monitoring, motion alarms).
   - **Corrective:** Restore affected systems or limit damage after detection (patching vulnerabilities, isolating compromised hosts, restoring clean backups).
   - **Compensating:** Provide an alternative safeguard when a primary control is temporarily infeasible (virtual patching at a WAF for an unpatchable legacy application).
   - **Directive:** Mandate compliant human or system behavior via explicit instruction (Acceptable Use Policies, restricted-area signage, SOP checklists).

## 2. SOC & Blue-Team Operational Practice

In Security Operations Center workflows, defense-in-depth correlates signals across control layers so failures in a **preventive** control are caught rapidly by **detective** telemetry and contained by **corrective** playbooks. For example, if an attacker bypasses preventive MFA prompts via token theft ([MITRE ATT&CK T1556](https://attack.mitre.org/techniques/T1556/)), SIEM rules detect impossible travel or anomalous MFA bursts (`Detective`) so responders revoke tokens and isolate the session (`Corrective`):

```kql
SigninLogs
| where ResultType != "0" or RiskLevelDuringSignIn in ("medium", "high")
| summarize FailedCount = count(), DistinctIPs = dcount(IPAddress) by UserPrincipalName, AppDisplayName
| where FailedCount >= 5 or DistinctIPs >= 3
```

## 3. References (Web Sources)

- [NIST SP 800-53 Rev. 5 — Security and Privacy Controls](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [Wikipedia — Security controls](https://en.wikipedia.org/wiki/Security_controls)
- [Wikipedia — Access control](https://en.wikipedia.org/wiki/Access_control)
- [MITRE ATT&CK — Modify Authentication Process (T1556)](https://attack.mitre.org/techniques/T1556/)
- [Microsoft Learn — KQL Overview](https://learn.microsoft.com/en-us/kusto/query/)
