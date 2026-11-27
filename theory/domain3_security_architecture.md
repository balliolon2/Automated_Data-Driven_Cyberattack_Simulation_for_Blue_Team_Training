# Domain 3: Security Architecture (3-Minute Essential)

> **Domain 3** | **Exam Weight:** 18% (`5 / 30` Pre/Post-Test Questions) | **Reading Time:** ~3 min

Think of designing a building with public reception in the lobby, locked hallways to classrooms, and an emergency backup generator — even if a visitor enters the lobby, they cannot walk straight into the vault.

---

## 1. Core Security+ Theory & Exam Concepts

- **Defense-in-Depth & Network Segmentation** ([Wikipedia: Defense in Depth](https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29), [Wikipedia: DMZ](https://en.wikipedia.org/wiki/DMZ_%28computing%29)):
  - *Defense-in-Depth:* Stack independent layers (firewall + IAM + EDR + backups) so bypassing one layer fails to reach core assets.
  - *Segmentation & Zero Trust:* Split networks into trust zones (Public DMZ, User LAN, Database VLAN, OT/ICS) using ACLs ([NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final), [CISA Zero Trust](https://www.cisa.gov/zero-trust-maturity-model)).
  - *DMZ & Bastion/Jump Hosts:* Public websites sit in the DMZ buffer zone; admins must pass through a hardened MFA jump host to touch critical servers.
- **Cloud Shared Responsibility Model:**
  - *IaaS* (e.g., cloud VMs): Provider secures hardware & hypervisor; **you** secure guest OS patches, firewall groups, IAM, and data.
  - *PaaS:* Provider runs the OS/runtime; **you** secure app code and data. *SaaS:* Provider runs the app; **you** manage user access and data classification.
- **Firewalls & Encrypted Protocols** ([Wikipedia: Firewall](https://en.wikipedia.org/wiki/Firewall_%28computing%29)):
  - NGFWs/WAFs filter zones and web traffic; use **TLS 1.3** (HTTPS), **IPsec VPN**, **SSH/SFTP**, and **SNMPv3** instead of plaintext Telnet, FTP, or HTTP.
- **Resilience & Data States** ([CISA #StopRansomware](https://www.cisa.gov/stopransomware)):
  - **3-2-1 Immutable Backups:** Keep 3 copies, 2 media types, 1 offsite immutable copy that ransomware cannot wipe.
  - Protect data **at rest** (disk/DB encryption), **in transit** (TLS/VPN), and **in use** (secure enclaves).

---

## 2. SOC Telemetry, KQL & MITRE ATT&CK

Domain 3 incidents catch attackers moving sideways across subnets ([T1021](https://attack.mitre.org/techniques/T1021/)), weakening firewall/cloud rules ([T1562](https://attack.mitre.org/techniques/T1562/)), or smuggling data out over alternate ports ([T1048](https://attack.mitre.org/techniques/T1048/)):

```kql
(source == "firewall" and dst_port == "445") or (source == "cloudtrail" and message contains "SecurityGroup")
```

- **Watch for:** A DMZ web server reaching port `445` (SMB) or `3389` (RDP) inside the internal DB subnet, or cloud rules opened to `0.0.0.0/0`.

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Guide

| True Positive (Real Attack) | False Positive (Safe Routine / Benign Context Log) |
|---|---|
| Cloud Security Group opened to `0.0.0.0/0` by a strange API key at night | Firewall/port rule added to match an **approved Infrastructure Change Ticket** |
| Sudden large encrypted stream to an unknown external IP | Scheduled nightly traffic to the verified backup appliance or authorized CDN |

Guided by [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final):
- **Containment:** Block the cross-zone path via ACL, revert the permissive `0.0.0.0/0` rule, and isolate the exposed DMZ host.
- **Eradication:** Disable plaintext Telnet/FTP/SMBv1 ports, rotate leaked cloud keys, and fix Infrastructure-as-Code templates.
- **Recovery:** Restore validated configs, re-verify boundary isolation, and monitor cross-zone flows.

---

## 4. Key Takeaways & References

- Never let a **DMZ host** talk straight to internal databases on admin ports (`22`, `445`, `3389`) without an approved change ticket.
- Query Reference: [Microsoft Learn: KQL](https://learn.microsoft.com/en-us/kusto/query/)
