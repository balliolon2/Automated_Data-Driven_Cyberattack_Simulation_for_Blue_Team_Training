# 1.3 Explain change management best practices

## 1. Core Security+ Theory

As documented in [Change Management](https://en.wikipedia.org/wiki/Change_management_(engineering)) and [Change Control](https://en.wikipedia.org/wiki/Change_control), formal change management prevents misconfigurations, unintended service outages, and unapproved security weakenings whenever hardware, software, firewall rules, or cloud IAM policies are updated:

- **Business Processes & Approval Flow:** Every non-trivial change starts with a **Change Request (CR)** detailing scope, justification, dependencies, and **Impact Analysis** (identifying blast radius and affected security zones). A **Change Advisory Board (CAB)** or peer reviewers authorize approved changes prior to execution; emergency patches follow an expedited emergency CAB track with mandatory post-hoc review.
- **Technical Safeguards:**
  - **Testing & Validation:** Replicating production configurations in a staging sandbox before deployment.
  - **Maintenance Windows:** Scheduling changes (including required service restarts or schema migrations) during pre-approved low-impact periods with stakeholder notification.
  - **Allowlists & Restricted Lists:** Enforcing deployment freezes during peak business windows and controlling which packages or binaries are permitted.
  - **Backout / Rollback Plan:** Every CR must document concrete steps and known-good snapshots/backups to revert cleanly if validation fails.
  - **Standard Operating Procedures (SOPs) & Version Control:** Updating architectural diagrams, runbooks, and version-controlled Infrastructure-as-Code (`IaC`) immediately after deployment to prevent configuration drift.
- **Legacy Systems & Exceptions:** When legacy applications prevent standard upgrades or restarts, risk owners document formal **exceptions** and apply compensating controls (segmentation or WAF rules).

## 2. SOC & Blue-Team Operational Practice

Change management records (`CHG-xxxx` tickets) are one of the highest-signal **Benign Context Logs** during SOC alert triage. When an SIEM rule flags a firewall policy modification, scheduled service restart, or GPO change, analysts verify whether the user account, target asset, and UTC timestamp match an approved CR in the maintenance window—preventing False Positive escalations while catching unauthorized **Impair Defenses** activity ([MITRE ATT&CK T1562](https://attack.mitre.org/techniques/T1562/)):

```kql
AuditLogs
| where OperationName has_any ("Update policy", "Modify firewall", "Install extension")
| project TimeGenerated, InitiatedBy, TargetResources, Result
| sort by TimeGenerated desc
```

## 3. References (Web Sources)

- [Wikipedia — Change management (engineering)](https://en.wikipedia.org/wiki/Change_management_(engineering))
- [Wikipedia — Change control](https://en.wikipedia.org/wiki/Change_control)
- [NIST SP 800-53 Rev. 5 — Configuration Management (CM)](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Impair Defenses (T1562)](https://attack.mitre.org/techniques/T1562/)
- [Microsoft Learn — KQL Overview](https://learn.microsoft.com/en-us/kusto/query/)
