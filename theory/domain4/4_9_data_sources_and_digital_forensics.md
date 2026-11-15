# 4.9 Use data sources to support an investigation and digital forensics

## 1. Core Security+ Theory

During investigations requiring legal admissibility or deep root-cause reconstruction, responders integrate log analysis with [Digital Forensics](https://en.wikipedia.org/wiki/Digital_forensics) standards from [NIST SP 800-86](https://csrc.nist.gov/pubs/sp/800/86/final):

- **Order of Volatility (RFC 3227 Principle):** Acquire evidence starting with the most ephemeral data before it vanishes on power-off or timeout:
  1. CPU registers and cache
  2. Routing tables, ARP cache, process list, and **system memory (RAM dump)**
  3. Temporary file systems / swap space
  4. Non-volatile storage (**bit-stream forensic disk image**)
  5. Remote logging, backups, and physical media
- **Preservation, Integrity & [Chain of Custody](https://en.wikipedia.org/wiki/Chain_of_custody):**
  - Issuing a **Legal Hold** immediately suspends routine log/email/backup deletion before litigation or e-discovery.
  - Using hardware **write blockers** prevents OS mounts from altering timestamps on seized disks.
  - Computing cryptographic digests (**SHA-256 / SHA-512**) at capture time and verifying hashes upon every handoff proves the copy remained unaltered, documented via signed **Chain of Custody** custody logs (who, what, when, where, why).
- **Investigation Data Sources:** SIEM logs, firewall/DNS/proxy trails, NetFlow metadata, full packet captures (`PCAP`), EDR memory timelines, and **time offset normalization** across unsynchronized hosts.

## 2. SOC & Blue-Team Operational Practice

Never reboot a compromised workstation before capturing volatile RAM if credential dumping (`LSASS` memory injection — [MITRE ATT&CK T1003](https://attack.mitre.org/techniques/T1003/)) or fileless PowerShell payloads are suspected; isolate at the network/EDR layer first, capture memory and disk hashes, and correlate artifacts using normalized UTC logs:

```kql
SecurityEvent
| where EventID in (4688, 1102, 4697)
| project TimeGenerated, Computer, SubjectUserName, EventID, CommandLine
| sort by TimeGenerated asc
```

## 3. References (Web Sources)

- [Wikipedia — Digital forensics](https://en.wikipedia.org/wiki/Digital_forensics)
- [Wikipedia — Chain of custody](https://en.wikipedia.org/wiki/Chain_of_custody)
- [NIST SP 800-86 — Integrating Forensic Techniques into Incident Response](https://csrc.nist.gov/pubs/sp/800/86/final)
- [MITRE ATT&CK — OS Credential Dumping (T1003)](https://attack.mitre.org/techniques/T1003/)
