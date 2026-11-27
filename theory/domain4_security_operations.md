# Domain 4: Security Operations (3-Minute Essential)

> **Domain 4** | **Exam Weight:** 28% (`8 / 30` Pre/Post-Test Questions — Largest Domain!) | **Reading Time:** ~3 min

Think of a hospital ER monitoring center: SIEM collects every heartbeat (logs), SOAR speeds up routine steps, and the SOC team decides which patient needs urgent emergency response and how to preserve evidence cleanly.

---

## 1. Core Security+ Theory & Exam Concepts

- **SOC Tools: SIEM, SOAR & EDR** ([Wikipedia: SIEM](https://en.wikipedia.org/wiki/Security_information_and_event_management)):
  - *SIEM:* Collects and correlates logs from firewalls, endpoints, and logins synced by **NTP** clocks.
  - *SOAR:* Automates repeatable lookups and containment steps via playbooks.
  - *EDR / XDR vs. IDS/IPS:* EDR watches endpoints (processes/memory); IDS passively alerts on network anomalies while inline **IPS** actively blocks them.
- **Vulnerability Operations:**
  - Prioritize patches using **CVSS** severity + asset criticality.
  - Know detection accuracy: *False Positive* = false alarm on safe activity; *False Negative* = missed real attack (far more dangerous!).
- **NIST Incident Response Lifecycle** ([NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final), [CMU IR Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)):
  1. *Preparation* (tools, sensors, playbooks) → 2. *Detection & Analysis* (triage, collect evidence logs) → 3. *Containment, Eradication & Recovery* → 4. *Post-Incident Activity (Lessons Learned)*.
- **Digital Forensics** ([Wikipedia: Digital Forensics](https://en.wikipedia.org/wiki/Digital_forensics)):
  - **Order of Volatility (RFC 3227):** Grab fragile data first: CPU cache → RAM → running processes/network connections → temp files → disk images → backups.
  - **Chain of Custody + SHA-256 hashes:** Prove nobody tampered with collected evidence.

---

## 2. SOC Telemetry, KQL & MITRE ATT&CK

Domain 4 represents real Tier-1 SOC triage: credential dumping ([T1003](https://attack.mitre.org/techniques/T1003/)), scheduled tasks ([T1053](https://attack.mitre.org/techniques/T1053/)), brute-force logins ([T1110](https://attack.mitre.org/techniques/T1110/)), and C2 beaconing ([T1071](https://attack.mitre.org/techniques/T1071/)):

```kql
severity == "high" or severity == "critical" or (source == "edr" and message contains "lsass")
```

- **4-Step SIEM Pivot:** Spot initial alert → filter around `host` or `src_ip` → check lateral movement or persistence → check benign maintenance tickets!

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Guide

| True Positive (Real Attack) | False Positive (Safe Routine / Benign Context Log) |
|---|---|
| `procdump`/unknown tool reads `lsass.exe` memory + SMB lateral hop | Scheduled endpoint antivirus scan or approved SCCM deployment job |
| Hundreds of `EventID 4625` failures followed by 1 success & new scheduled task | Authorized IT admin monitoring script documented in CMDB/change logs |

- **Containment:** Network-isolate the endpoint via EDR (never pull the power plug first, or you wipe RAM evidence!), block C2 IPs, lock compromised accounts.
- **Eradication:** Kill rogue tasks/services, clean binaries, revoke stolen Kerberos/session tokens, reset passwords.
- **Recovery:** Reconnect clean hosts, re-scan for persistence, and add new detection rules during Lessons Learned.

---

## 4. Key Takeaways & References

- Never power off a compromised host before capturing volatile **RAM**; use **EDR isolation** instead so evidence survives!
- Query Reference: [Microsoft Learn: KQL](https://learn.microsoft.com/en-us/kusto/query/)
