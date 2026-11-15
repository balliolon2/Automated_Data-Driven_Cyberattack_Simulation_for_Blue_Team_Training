# 3.4 Explain the importance of resilience and recovery in security architecture

## 1. Core Security+ Theory

Resilience preserves operational availability against hardware failure, cyberattacks (especially ransomware), and natural disasters ([Business Continuity](https://en.wikipedia.org/wiki/Business_continuity)):

- **High Availability (`HA`) & Redundancy:** Eliminates single points of failure (`SPOF`) via active-active/active-passive server clustering, **load balancing**, multipath networking, **N+1** redundant power/cooling, and geographically dispersed availability zones.
- **Storage & Backup Strategies:**
  - **Full Backup:** Complete copy—slowest to capture/largest storage, fastest single-step restore.
  - **Incremental Backup:** Copies only data changed since the *last backup* (full or incremental)—fastest write, slower multi-step chain restore.
  - **Differential Backup:** Copies all data changed since the *last full backup*—balanced speed (`Full + latest Differential`).
  - **Snapshots, Replication & Immutable Off-site Backups (3-2-1 Rule):** Write-Once-Read-Many (**WORM / immutable**) object locks prevent ransomware from wiping backups ([CISA #StopRansomware](https://www.cisa.gov/stopransomware)).
- **Recovery Site Types:**
  - **Hot Site:** Pre-configured mirror facility with live/near-real-time data replication (lowest downtime, highest cost).
  - **Warm Site:** Infrastructure/network provisioned, requires recent backup restoration before cutover.
  - **Cold Site:** Empty conditioned facility space/power/cooling awaiting full hardware delivery and builds (slowest recovery, lowest cost).
- **Power Resiliency & Key Metrics:** **UPS** batteries handle short brownouts and bridge switchover to **diesel generators**. Recovery targets balance **RTO** (maximum tolerable downtime) and **RPO** (maximum acceptable data loss age), alongside **MTBF** and **MTTR**.

## 2. SOC & Blue-Team Operational Practice

Ransomware operators specifically target recovery planes before encrypting user files—disabling Volume Shadow Copies (`vssadmin`), stopping backup services, or deleting cloud snapshots ([MITRE ATT&CK T1486](https://attack.mitre.org/techniques/T1486/)). Verifying immutable backup integrity before starting eradication ensures clean system rebuilds:

```kql
SecurityEvent
| where EventID in (7036, 7045)
| where EventData has_any ("vss", "wbengine", "Veeam", "BackupExec")
| project TimeGenerated, Computer, EventID, EventData
```

## 3. References (Web Sources)

- [Wikipedia — Business continuity](https://en.wikipedia.org/wiki/Business_continuity)
- [Wikipedia — Business impact analysis](https://en.wikipedia.org/wiki/Business_impact_analysis)
- [CISA — #StopRansomware Recovery & Immutable Backups](https://www.cisa.gov/stopransomware)
- [NIST SP 800-53 Rev. 5 — Contingency Planning (CP)](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Data Encrypted for Impact (T1486)](https://attack.mitre.org/techniques/T1486/)
