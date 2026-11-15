# 4.5 Modify enterprise capabilities to enhance security

## 1. Core Security+ Theory

Security operations engineers continuously tune and harden production controls to close emerging attack paths ([NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [Firewall (Computing)](https://en.wikipedia.org/wiki/Firewall_%28computing%29)):

- **Firewall & Network Capability Modifications:** Updating stateful ACLs and zones to enforce explicit allowlisting, closing legacy inbound management ports (`22/3389`) behind jump hosts/ZTNA, and adding **DNS Filtering / Sinkholing** so lookups for known malicious C2/phishing domains resolve to safe internal sinkholes.
- **Email Security Hardening:** Enforcing sender verification and anti-spoofing standards on DNS/MTA boundaries:
  - **SPF (Sender Policy Framework):** Declares authorized outbound mail server IPs for a domain.
  - **DKIM (DomainKeys Identified Mail):** Cryptographically signs message headers/body to verify integrity and origin.
  - **DMARC:** Ties SPF/DKIM alignment to explicit enforcement policies (`none`, `quarantine`, `reject`) plus aggregate reporting.
- **Protocol & Service Upgrades:** Replacing cleartext/weak administration protocols with encrypted, authenticated equivalents (**Telnet → SSH**, **HTTP → HTTPS/TLS 1.3**, **SNMPv1/v2c → SNMPv3**, **LDAP → LDAPS**, **NTP → NTS**).
- **Endpoint, DLP & Content Filtering:** Enabling EDR attack surface reduction (`ASR`) rules blocking Office macros from spawning child shells, tuning URL categories, and quarantining malicious executables.

## 2. SOC & Blue-Team Operational Practice

Hardening recommendations directly remediate gaps exposed during SOC investigations—such as an unauthenticated email spoofing phish ([MITRE ATT&CK T1566](https://attack.mitre.org/techniques/T1566/)) slipping past permissive `p=none` DMARC records, or PowerShell execution ([T1059](https://attack.mitre.org/techniques/T1059/)) from unblocked internet downloads:

```kql
EmailEvents
| where EmailDirection == "Inbound" and ThreatTypes has "Phish"
| summarize Count = count() by SenderFromDomain, AuthenticationDetails
| order by Count desc
```

## 3. References (Web Sources)

- [Wikipedia — Firewall (computing)](https://en.wikipedia.org/wiki/Firewall_%28computing%29)
- [CISA — Phishing Threat & Response Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)
- [NIST SP 800-53 Rev. 5 — Configuration Management & Access Control](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Command and Scripting Interpreter (T1059)](https://attack.mitre.org/techniques/T1059/)
