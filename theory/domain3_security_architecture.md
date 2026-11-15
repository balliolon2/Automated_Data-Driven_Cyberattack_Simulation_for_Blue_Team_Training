# Domain 3: Security Architecture

> **Domain ID:** `domain3` | **Domain Scope:** 18% Weight (5 / 30 Pre/Post-Test Questions) | **Sources:** Wikipedia, NIST CSRC, CISA, MITRE ATT&CK

This module synthesizes enterprise and cloud security architecture concepts from public standards and open references—including [NIST SP 800-207: Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final), [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), the [CISA Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model), and Wikipedia systems architecture articles.

---

## 1. Core Security+ Theory & Exam Concepts

### 1.1 Network Segmentation, Zones & Defense-in-Depth
As described in [Wikipedia: Defense in Depth](https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29), [Wikipedia: DMZ (Computing)](https://en.wikipedia.org/wiki/DMZ_%28computing%29), and [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final) (SC-7 Boundary Protection):
- **Defense-in-Depth:** Layering independent technical and administrative controls so that bypassing any single defense does not immediately compromise core data.
- **Network Segmentation & Micro-segmentation:** Dividing enterprise networks into isolated trust zones (User LAN, Server VLAN, OT/ICS VLAN, Management Plane) controlled by firewalls and ACLs to constrain lateral movement ([NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final)).
- **DMZ (Screened Subnet):** Exposes public-facing services (web servers, reverse proxies, mail gateways) in a buffer subnet separated from the internal LAN by firewalls ([Wikipedia: DMZ](https://en.wikipedia.org/wiki/DMZ_%28computing%29)).
- **Air Gaps & Jump Hosts (Bastion Hosts):** Physically or logically isolating high-sensitivity enclaves (e.g., OT chillers/SCADA) and requiring hardened MFA jump servers for administrative sessions.

### 1.2 Cloud & Virtualization Security Models
- **Cloud Deployment Models:** Public, Private, Hybrid, and Community cloud environments.
- **Shared Responsibility Model:**
  - *IaaS:* Provider secures physical hardware and hypervisor; customer secures guest OS, patching, network security groups, IAM, and data.
  - *PaaS:* Provider manages OS and runtime; customer secures application code, configurations, and data.
  - *SaaS:* Provider manages infrastructure and software; customer manages user identities, RBAC, and data classification.
- **Cloud Architecture Controls:** Virtual Private Cloud (VPC) security groups, NACLs, Cloud Access Security Brokers (CASB), Infrastructure as Code (IaC) scanning, and immutable containers.

### 1.3 Secure Network Infrastructure & Communications
Described in [Wikipedia: Firewall (Computing)](https://en.wikipedia.org/wiki/Firewall_%28computing%29) and [CISA Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model):
- **Firewalls & Application Gateways:** Stateful/Next-Generation Firewalls (NGFW), Web Application Firewalls (WAF), and forward/reverse proxies enforce zone policy and inspect application traffic.
- **Encrypted Transport & Tunneling Protocols:**
  - *TLS 1.2 / 1.3:* Encrypts data in transit (HTTPS, LDAPS, SMTPS).
  - *IPsec VPN (AH/ESP, IKEv2):* Site-to-site and remote-access encrypted tunneling.
  - *SSH / SFTP / SNMPv3:* Encrypted management and telemetry protocols replacing plaintext Telnet, FTP, and SNMPv1/v2c.
  - *DNSSEC, DoH/DoT:* Protects DNS integrity and privacy against spoofing and eavesdropping.

### 1.4 Resiliency, High Availability & Data States
- **Resilience & Fault Tolerance:** Active-active load balancing, multi-region replication, RAID storage, and backup power eliminate single points of failure.
- **Immutable Backups (3-2-1 Rule):** Keeping 3 copies on 2 different media types with 1 offsite/immutable copy guarantees clean recovery from ransomware ([CISA #StopRansomware](https://www.cisa.gov/stopransomware)).
- **Data States:** Protecting data at rest (full-disk/database encryption), data in transit (TLS/IPsec), and data in use (confidential computing/enclaves).

---

## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping

Domain 3 incidents focus on architectural boundary violations: unexpected cross-zone traffic, misconfigured cloud storage/security groups, insecure legacy protocols, or lateral movement across segmented subnets.

### 2.1 Key Log Sources & Fields
- **Firewall & NGFW Logs:** `src_ip`, `dst_ip`, `dst_port`, `src_zone`, `dst_zone`, `action` (`ALLOW`/`DENY`), `protocol`, `bytes_out`.
- **Cloud Audit Logs (AWS CloudTrail / Azure Activity):** `event_name` (`AuthorizeSecurityGroupIngress`, `PutBucketAcl`), `principal`, `source_ip`, `region`, `resource_arn`.
- **Proxy / DNS Logs:** Outbound `CONNECT` tunnels on non-standard ports, DNS tunneling queries, or unauthorized SaaS uploads.

### 2.2 Common KQL Investigation Patterns

Query syntax aligned with [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/):

```kql
source == "firewall" and dst_port == "445"
source == "cloudtrail" and message contains "SecurityGroup"
source == "proxy" and action == "allow"
```

### 2.3 MITRE ATT&CK Mapping

All technique IDs link directly to the official [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/):

| Technique ID | Technique Name | Observable Telemetry Indicator |
|---|---|---|
| [T1021](https://attack.mitre.org/techniques/T1021/) | Remote Services | Internal host initiating SMB (445), RDP (3389), WinRM (5985), or SSH (22) across subnet boundaries |
| [T1562](https://attack.mitre.org/techniques/T1562/) | Impair Defenses | Firewall rule widened to `0.0.0.0/0`, cloud logging disabled, or security group modified to expose management ports |
| [T1048](https://attack.mitre.org/techniques/T1048/) | Exfiltration Over Alternative Protocol | Large outbound transfers over DNS, ICMP, or non-standard ports bypassing DLP proxies |
| [T1090](https://attack.mitre.org/techniques/T1090/) | Proxy | Internal host establishing unauthorized SOCKS/HTTP tunnels or multi-hop relay traffic |

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide

### 3.1 Distinguishing True Positives (TP) from False Positives (FP)
- **True Positive Indicators:** DMZ web server initiating SMB/RDP connections into the internal database subnet; cloud S3 bucket or security group opened to `0.0.0.0/0` by an anomalous user/API key; sustained high-bandwidth outbound traffic to an unknown external IP outside business hours.
- **False Positive Indicators (Benign Context Logs):**
  - **Approved Firewall / Cloud Change Ticket:** A new port allowance or security group modification matching a documented infrastructure change request.
  - **Authorized Backup / Replication Job:** High-volume internal or offsite encrypted transfer occurring at the scheduled backup window to a known backup appliance IP.
  - **Observed Authorized Traffic Pattern:** High proxy `CONNECT` bandwidth matching known legitimate business activity (e.g., approved software distribution CDN or authorized media/training stream) when corroborated by asset context.

### 3.2 3-Phase Playbook Response Principles
Guided by [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final), and [CMU Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html):
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

### Authoritative Web References
- **Wikipedia (Open Reference):**
  - [Defense in Depth (Computing)](https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29)
  - [DMZ (Computing)](https://en.wikipedia.org/wiki/DMZ_%28computing%29)
  - [Firewall (Computing)](https://en.wikipedia.org/wiki/Firewall_%28computing%29)
  - [Zero Trust Security Model](https://en.wikipedia.org/wiki/Zero_trust_security_model)
- **Government & Standards Agencies (NIST & CISA):**
  - [NIST SP 800-207: Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final)
  - [NIST SP 800-53 Rev. 5: Security and Privacy Controls](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
  - [CISA Zero Trust Maturity Model (ZTMM)](https://www.cisa.gov/zero-trust-maturity-model)
- **Universities & Open Knowledge Bases:**
  - [CMU Information Security Office: Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)
  - [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/) (`T1021`, `T1562`, `T1048`, `T1090`)
  - [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/)
