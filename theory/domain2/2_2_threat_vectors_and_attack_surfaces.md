# 2.2 Analyze common threat vectors and attack surfaces

## 1. Core Security+ Theory

An organization's [Attack Surface](https://en.wikipedia.org/wiki/Attack_surface) comprises every exposed network port, software endpoint, identity portal, supply-chain integration, and human workflow an adversary can reach ([Phishing](https://en.wikipedia.org/wiki/Phishing), [Social Engineering](https://en.wikipedia.org/wiki/Social_engineering_%28security%29), [CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)):

- **Message-Based & Phishing Vectors:** Bulk phishing, targeted **spear phishing**, executive-focused **whaling**, voice-based **vishing**, SMS-based **smishing**, **BEC (Business Email Compromise)** impersonation, and deceptive **typosquatting** domains.
- **Social Engineering Principles:** Pretexting, authority intimidation, urgency/scarcity triggers, familiarity/liking, and consensus manipulation to bypass technical controls via human compliance.
- **Technical & Unsecured Vectors:**
  - Unsecured open services, default factory credentials, vulnerable RDP/VPN concentrators, and rogue wireless networks (evil twin).
  - Uncontrolled **removable media** (USB drop attacks) and malicious links or watering-hole compromises.
  - **Supply Chain Vectors:** Compromised software dependencies, third-party managed service providers (MSPs), and tampered CI/CD build pipelines ([NIST SP 800-161 Rev. 1](https://csrc.nist.gov/pubs/sp/800/161/r1/final)).
- **Human & Default Vectors:** Shadow IT cloud sign-ups outside governance and unrevoked contractor accounts.

## 2. SOC & Blue-Team Operational Practice

Email gateways, proxy logs, and endpoint telemetry capture initial access ([MITRE ATT&CK T1566 — Phishing](https://attack.mitre.org/techniques/T1566/)) and public-facing service exploitation ([T1190](https://attack.mitre.org/techniques/T1190/)). Analysts trace click-through chains from delivery URLs to child PowerShell/Office macro execution:

```kql
EmailUrlInfo
| join kind=inner (UrlClickEvents | where ActionType == "ClickAllowed") on NetworkMessageId
| join kind=leftouter (DeviceProcessEvents) on $left.AccountUpn == $right.InitiatingProcessAccountUpn
| where DeviceProcessEvents.FileName in~ ("powershell.exe", "cmd.exe", "mshta.exe")
| project Url, AccountUpn, DeviceName, FileName, ProcessCommandLine
```

## 3. References (Web Sources)

- [Wikipedia — Attack surface](https://en.wikipedia.org/wiki/Attack_surface)
- [Wikipedia — Phishing](https://en.wikipedia.org/wiki/Phishing)
- [Wikipedia — Social engineering (security)](https://en.wikipedia.org/wiki/Social_engineering_%28security%29)
- [CISA — Phishing Threat & Response Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)
- [NIST SP 800-161 Rev. 1 — Cybersecurity Supply Chain Risk Management](https://csrc.nist.gov/pubs/sp/800/161/r1/final)
- [MITRE ATT&CK — Phishing (T1566)](https://attack.mitre.org/techniques/T1566/)
