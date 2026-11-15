# 4.8 Explain appropriate incident response activities

## 1. Core Security+ Theory

Structured [Computer Security Incident Management](https://en.wikipedia.org/wiki/Computer_security_incident_management) minimizes breach cost and guides calm coordination using the [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final) & [Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final) lifecycle alongside institutional runbooks like [CMU ISO Incident Response](https://www.cmu.edu/iso/governance/procedures/incident-response.html):

1. **Preparation:** Assembling CSIRT roles, communication matrices, legal/PR escalations, jump-kits, centralized log retention, and practicing **tabletop exercises / attack simulations** before an incident hits.
2. **Detection & Analysis:** Monitoring alerts, correlating timeline telemetry across hosts/networks, eliminating False Positives, determining scope, and declaring severity.
3. **Containment, Eradication & Recovery:**
   - **Containment:** Short-term network/host isolation or credential revocation stops lateral spread without immediately wiping volatile evidence.
   - **Eradication:** Removing adversary artifacts—terminating malicious processes, deleting scheduled persistence, closing initial exploit paths, and rotating all potentially exposed secrets.
   - **Recovery:** Restoring verified clean backups, rebuilding compromised hosts from golden images, tightening firewall/EDR monitoring, and confirming normal operations.
4. **Post-Incident Activity (Lessons Learned):** Conducting blame-free root-cause analysis (`RCA`), updating SIEM detections and SOPs, and fulfilling regulatory/external breach reporting within statutory windows.

## 2. SOC & Blue-Team Operational Practice

During incident triage and reporting ([Stanford UIT Report Incident](https://uit.stanford.edu/security/report-incident)), analysts distinguish confirmed malicious compromise (`True Positive` requiring immediate containment) from authorized penetration tests or change-ticket maintenance (`False Positive` requiring no disruption):

```kql
union DeviceProcessEvents, SigninLogs
| where TimeGenerated between (datetime(2026-11-01T00:00:00Z) .. datetime(2026-11-01T06:00:00Z))
| summarize EventCount = count(), Earliest = min(TimeGenerated), Latest = max(TimeGenerated) by Type
```

## 3. References (Web Sources)

- [NIST SP 800-61 Rev. 2 — Computer Security Incident Handling Guide](https://csrc.nist.gov/pubs/sp/800/61/r2/final)
- [NIST SP 800-61 Rev. 3 — Incident Response Recommendations](https://csrc.nist.gov/pubs/sp/800/61/r3/final)
- [Wikipedia — Computer security incident management](https://en.wikipedia.org/wiki/Computer_security_incident_management)
- [Carnegie Mellon University ISO — Computer Security Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)
- [Stanford University UIT — Incident Reporting & Response](https://uit.stanford.edu/security/report-incident)
