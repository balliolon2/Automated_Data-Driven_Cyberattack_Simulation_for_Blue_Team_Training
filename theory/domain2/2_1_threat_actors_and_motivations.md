# 2.1 Compare and contrast threat actors and intelligence sources

## 1. Core Security+ Theory

Understanding adversaries—their funding, sophistication, access level, and strategic intent—guides prioritized defense ([Threat (Computer)](https://en.wikipedia.org/wiki/Threat_(computer)), [Advanced Persistent Threat](https://en.wikipedia.org/wiki/Advanced_persistent_threat)):

- **Threat Actor Categories:**
  - **Nation-State / APT:** Well-funded state-sponsored teams conducting long-dwell cyber espionage, intellectual property theft, or critical infrastructure disruption using custom tooling and zero-days.
  - **Organized Cybercrime:** Profit-driven groups operating Ransomware-as-a-Service (RaaS), credential marketplaces, and business email compromise (BEC).
  - **Hacktivists:** Ideologically motivated actors using DDoS, defacement, or hack-and-leak operations for political messaging.
  - **Unskilled Attackers (Script Kiddies):** Opportunistic actors running public exploit kits without deep technical understanding.
  - **Insider Threats:** Authorized employees, contractors, or partners acting either **maliciously** (revenge, bribery, sabotage) or **negligently** (accidental leaks, Shadow IT deployment).
- **Threat Intelligence Sources:**
  - **OSINT (Open-Source Intelligence):** Public advisories, CVE feeds, paste monitoring, WHOIS/certificate transparency logs, and dark-web monitoring.
  - **Information-Sharing Communities:** Sector-specific **ISACs / ISAOs** and government alerts such as [CISA Advisories](https://www.cisa.gov/stopransomware).
  - **Automated Feeds:** Machine-readable IOCs exchanged via structured formats (**STIX / TAXII**) and mapped to adversary behavior in [MITRE ATT&CK Enterprise](https://attack.mitre.org/matrices/enterprise/).

## 2. SOC & Blue-Team Operational Practice

Cyber Threat Intelligence (`CTI`) shifts SOC teams from chasing ephemeral IP addresses to hunting durable TTPs (Tactics, Techniques, and Procedures). When an ISAC feed alerts on brute-force password spraying ([MITRE ATT&CK T1110](https://attack.mitre.org/techniques/T1110/)) or trusted vendor compromise ([T1199](https://attack.mitre.org/techniques/T1199/)), analysts sweep historical logs for matching indicators:

```kql
ThreatIntelligenceIndicator
| join kind=inner (CommonSecurityLog | where TimeGenerated > ago(7d)) on $left.NetworkIP == $right.SourceIP
| project TimeGenerated, SourceIP, DestinationIP, DestinationPort, Description
```

## 3. References (Web Sources)

- [Wikipedia — Threat (computer)](https://en.wikipedia.org/wiki/Threat_(computer))
- [Wikipedia — Advanced persistent threat](https://en.wikipedia.org/wiki/Advanced_persistent_threat)
- [MITRE ATT&CK — Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/)
- [CISA — #StopRansomware Resources](https://www.cisa.gov/stopransomware)
- [Stanford University UIT — Information Security](https://uit.stanford.edu/security)
