# Domain 3: Security Architecture

> **Domain ID:** `domain3` | **CompTIA Security+ (SY0-701) Exam Weight:** 18% | **Pre/Post-Test Quota:** 5 / 30 Questions

This module covers enterprise and cloud security architecture principles tested in Domain 3, explaining how network segmentation, DMZs, cloud shared responsibility, secure protocols, and infrastructure resilience guide Blue Team investigation and containment.

---

## 1. Core Security+ Theory & Exam Concepts

### 1.1 Network Segmentation, Zones & Defense-in-Depth
**Defense-in-Depth** layers independent technical and administrative controls so no single failure compromises critical assets [R2, R5]:
- **Network Segmentation & Micro-segmentation:** Dividing networks into isolated zones (User LAN, Server VLAN, OT/ICS VLAN, Management Plane) controlled by firewalls, ACLs, and Software-Defined Networking (SDN) policies to block lateral movement.
- **DMZ (Screened Subnet):** Isolates public-facing services (web servers, mail gateways, reverse proxies) between an external firewall and an internal firewall so compromised DMZ hosts cannot directly reach internal databases.
- **Air Gaps & Jump Hosts (Bastion Hosts):** Physically or logically separating high-security enclaves (e.g., OT chillers/SCADA or root CA) and requiring hardened jump servers with MFA for administrative access.

### 1.2 Cloud & Virtualization Security Models
- **Cloud Deployment Models:** Public, Private, Hybrid, and Community cloud.
- **Service Models & Shared Responsibility:**
  - *IaaS:* Provider secures physical hardware and hypervisor; customer secures OS, patching, network security groups, IAM, and data.
  - *PaaS:* Provider manages OS and runtime; customer secures application code, configurations, and data.
  - *SaaS:* Provider manages infrastructure and app; customer manages user identities, RBAC, and data governance.
- **Cloud-Native Architecture Controls:** VPC security groups, NACLs, CASB (Cloud Access Security Broker), IaC (Infrastructure as Code) scanning, and container immutability.

### 1.3 Secure Network Infrastructure & Protocols
- **Secure Appliances & Gateways:** Next-Generation Firewalls (NGFW), Web Application Firewalls (WAF), Forward/Reverse Proxies, Load Balancers, and Unified Threat Management (UTM).
- **Secure Communications:**
  - *TLS 1.2 / 1.3:* Encrypts data in transit (HTTPS, LDAPS, SMTPS).
  - *IPsec VPN (AH/ESP, IKEv2):* Site-to-site or remote-access encrypted tunneling.
  - *SSH / SFTP / SNMPv3:* Encrypted management and telemetry protocols replacing plaintext Telnet, FTP, and SNMPv1/v2c.
  - *DNSSEC, DoH/DoT:* Protects DNS integrity and privacy against spoofing and eavesdropping.
- **SD-WAN & SASE:** Integrates cloud-delivered security services (SWG, ZTNA, CASB, FWaaS) with software-defined WAN routing.

### 1.4 Resiliency, High Availability & Data Protection
- **Resilience & Fault Tolerance:** Active-active load balancing, RAID storage, multi-region replication, UPS, and generators prevent single points of failure.
- **Backup Strategies:** Full, incremental, and differential backups following the **3-2-1 rule** (3 copies, 2 media types, 1 offsite/immutable copy) to ensure ransomware recovery.
- **Data States & Protection:** Data at rest (full-disk/database encryption), data in transit (TLS/IPsec), and data in use (confidential computing/enclaves), reinforced by DLP (Data Loss Prevention) and tokenization.

---

## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping

Domain 3 incidents focus on architectural boundary violations: unexpected cross-zone traffic, misconfigured cloud storage/security groups, insecure legacy protocols, or lateral movement across segmented subnets.

### 2.1 Key Log Sources & Fields
- **Firewall & NGFW Logs:** `src_ip`, `dst_ip`, `dst_port`, `src_zone`, `dst_zone`, `action` (`ALLOW`/`DENY`), `protocol`, `bytes_out`.
- **Cloud Audit Logs (AWS CloudTrail / Azure Activity):** `event_name` (`AuthorizeSecurityGroupIngress`, `PutBucketAcl`), `principal`, `source_ip`, `region`, `resource_arn`.
- **Proxy / DNS Logs:** Outbound `CONNECT` tunnels on non-standard ports, DNS tunneling queries, or unauthorized SaaS uploads.

### 2.2 Common KQL Investigation Patterns

```kql
source == "firewall" and dst_port == "445"
source == "cloudtrail" and message contains "SecurityGroup"
source == "proxy" and action == "allow"
```

### 2.3 MITRE ATT&CK Mapping [R4]

| Technique ID | Technique Name | Observable Telemetry Indicator |
|---|---|---|
| T1021 | Remote Services | Internal host initiating SMB (445), RDP (3389), WinRM (5985), or SSH (22) across subnet boundaries |
| T1562 | Impair Defenses | Firewall rule widened to `0.0.0.0/0`, cloud logging disabled, or security group modified to expose management ports |
| T1048 | Exfiltration Over Alternative Protocol | Large outbound transfers over DNS, ICMP, or non-standard ports bypassing DLP proxies |
| T1090 | Proxy | Internal host establishing unauthorized SOCKS/HTTP tunnels or multi-hop relay traffic |

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide

### 3.1 Distinguishing True Positives (TP) from False Positives (FP)
- **True Positive Indicators:** DMZ web server initiating SMB/RDP connections into the internal database subnet; cloud S3 bucket or security group opened to `0.0.0.0/0` by an anomalous user/API key; sustained high-bandwidth outbound traffic to an unknown external IP outside business hours.
- **False Positive Indicators (Benign Context Logs):**
  - **Approved Firewall / Cloud Change Ticket:** A new port allowance or security group modification matching a documented infrastructure change request.
  - **Authorized Backup / Replication Job:** High-volume internal or offsite encrypted transfer occurring at the scheduled backup window to a known backup appliance IP.
  - **Observed Authorized Traffic Pattern:** High proxy `CONNECT` bandwidth matching known legitimate business activity (e.g., approved software distribution CDN or authorized media/training stream) when corroborated by asset context.

### 3.2 3-Phase Playbook Response Principles [R3]
- **Phase 1 — Containment:**
  - *Correct:* Enforce micro-segmentation ACLs blocking the unauthorized cross-zone path, revert overly permissive security group ingress rules (`0.0.0.0/0`), or isolate the compromised DMZ/cloud workload.
  - *Distractors to Avoid:* Reformatting network switches across the entire subnet or shutting down core production DNS/DHCP infrastructure.
- **Phase 2 — Eradication:**
  - *Correct:* Close exposed legacy plaintext ports (Telnet/FTP/SMBv1), rotate leaked cloud IAM keys, remove unauthorized SSH keys or firewall rules, and remediate IaC template misconfigurations.
  - *Distractors to Avoid:* Leaving the misconfigured IaC pipeline active so the next deployment re-introduces the exposure.
- **Phase 3 — Recovery:**
  - *Correct:* Validate zero-trust segmentation rules via architecture audit, restore immutable configuration baselines, verify redundancy/failover health, and re-enable restricted production traffic.
  - *Distractors to Avoid:* Leaving temporary "allow all" troubleshooting rules in place after restoration.

---

## 4. Key Takeaways & References

- **Think in Trust Boundaries:** Always check *which zone* (`DMZ`, `Internal`, `Cloud`, `OT`) initiated the connection and whether that direction violates least-privilege architecture.
- **Verify Infrastructure Changes:** Firewall and cloud configuration alerts are frequently False Positives when backed by a change ticket or scheduled backup/replication log.
- **Preserve Availability While Isolating:** Containment targets the violated boundary or compromised host—never take down unrelated core network infrastructure.

### References
- `[R2]` CompTIA, Inc. (2023). *CompTIA Security+ Certification Exam Objectives (Exam SY0-701)* — Domain 3.0 Security Architecture.
- `[R3]` NIST. (2020). *NIST SP 800-53 Rev. 5: Security and Privacy Controls for Information Systems and Organizations* (SC & AC control families).
- `[R4]` The MITRE Corporation. (2024). *MITRE ATT&CK Enterprise Matrix* (`T1021`, `T1562`, `T1048`, `T1090`).
- `[R5]` Wikipedia contributors. (2026). *Defense in depth (computing)* and *DMZ (computing)*.
