# Domain 2: Threats, Vulnerabilities, and Mitigations

> **Domain ID:** `domain2` | **Domain Scope:** 22% Weight (7 / 30 Pre/Post-Test Questions) | **Sources:** Wikipedia, CISA, OWASP, MITRE ATT&CK

This module covers threat vectors, social engineering, software/web vulnerabilities, malware families, and mitigations grounded in public agency advisories and open security references—including [CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing), [CISA #StopRansomware](https://www.cisa.gov/stopransomware), the [OWASP Top 10:2021](https://owasp.org/Top10/), and Wikipedia cybersecurity articles.

---

## 1. Core Security+ Theory & Exam Concepts

### 1.1 Threat Actors & Adversary Motivations
Adversaries analyzed in threat intelligence ([MITRE ATT&CK®](https://attack.mitre.org/matrices/enterprise/), [CISA](https://www.cisa.gov/stopransomware)) differ in resources, stealth, and objectives:
- **Nation-State / APT (Advanced Persistent Threat):** Well-funded state-sponsored groups targeting intellectual property, critical infrastructure, or espionage using stealthy, multi-stage campaigns.
- **Organized Crime / Ransomware Groups:** Financially motivated actors deploying ransomware and double-extortion schemes ([CISA #StopRansomware](https://www.cisa.gov/stopransomware)).
- **Hacktivists:** Ideologically or politically motivated actors conducting DDoS attacks, website defacements, or data leaks.
- **Insider Threats:** Authorized personnel (malicious or negligent) abusing legitimate privileges ([Stanford Information Security](https://uit.stanford.edu/security)).
- **Opportunistic / Unskilled Actors:** Automated exploit kits and off-the-shelf credential-stuffing or scanning scripts.

### 1.2 Social Engineering & Phishing Vectors
Defined in [Wikipedia: Social Engineering](https://en.wikipedia.org/wiki/Social_engineering_%28security%29), [Wikipedia: Phishing](https://en.wikipedia.org/wiki/Phishing), and [CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing):
- **Phishing / Spear Phishing / Whaling:** Deceptive messages impersonating trusted entities to steal credentials or trigger malicious downloads (broad campaigns, targeted staff, or executives).
- **Vishing & Smishing:** Voice-based and SMS-based social engineering.
- **Business Email Compromise (BEC):** Impersonation or compromise of executive/vendor email accounts to redirect wire transfers or sensitive data.
- **Pretexting, Typosquatting & Watering Hole:** Fabricating urgent scenarios, registering look-alike domains (e.g., `corp-login[.]com`), or compromising external sites frequented by target personnel.

### 1.3 Software & Web Application Vulnerabilities (OWASP Top 10:2021)
Categorized by the [OWASP Top 10:2021](https://owasp.org/Top10/) plus [Wikipedia: SQL Injection](https://en.wikipedia.org/wiki/SQL_injection) and [Wikipedia: Cross-site Scripting (XSS)](https://en.wikipedia.org/wiki/Cross-site_scripting):
- **A01 Broken Access Control:** Missing authorization checks enabling IDOR (Insecure Direct Object Reference), privilege escalation, or directory traversal (`../../etc/passwd`).
- **A03 Injection (SQLi, Command Injection, XSS):** Untrusted input interpreted as executable SQL (`' OR 1=1--`), OS shell commands, or browser scripts (Stored/Reflected/DOM XSS).
- **A06 Vulnerable & Outdated Components:** Unpatched libraries or internet-facing appliances exposing known CVEs.
- **A07 Identification & Authentication Failures:** Credential stuffing, brute force, weak session tokens, and missing MFA.
- **A10 Server-Side Request Forgery (SSRF):** Coercing a server into fetching internal metadata (`169.254.169.254`) or internal APIs.

### 1.4 Malware & Impact Mechanisms
Described in [Wikipedia: Malware](https://en.wikipedia.org/wiki/Malware) and [Wikipedia: Ransomware](https://en.wikipedia.org/wiki/Ransomware):
- **Ransomware:** Encrypts host/network files and deletes volume shadow copies to disrupt availability and extort payment ([CISA #StopRansomware](https://www.cisa.gov/stopransomware)).
- **Trojans, RATs & Rootkits:** Disguised malicious software establishing C2 channels or hiding kernel/user-mode artifacts.
- **Living-off-the-Land (LotL) / Fileless Execution:** Executes malicious logic in memory using built-in OS tools (`powershell.exe`, `wmic.exe`, `certutil.exe`).

### 1.5 Vulnerability Mitigation Techniques
- **Parameterized Queries & Input Validation:** Separates user data from executable SQL/OS syntax to neutralize injection ([OWASP Top 10](https://owasp.org/Top10/)).
- **Least Privilege & Network Segmentation:** Limits blast radius if an endpoint or web service is compromised.
- **Patch Management & Virtual Patching (WAF/IPS):** Remediates known CVEs or blocks exploit patterns at the edge while testing patches.
- **Email Authentication & EDR:** Enforces SPF/DKIM/DMARC ([CISA](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)) and monitors endpoint process behavior.

---

## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping

### 2.1 Key Log Sources & Indicators of Compromise (IoCs)
- **Web / WAF / Reverse Proxy Logs:** URI payloads containing SQL keywords (`UNION SELECT`), path traversal (`../`), shell metacharacters, or scanner signatures (`sqlmap`, `nikto`).
- **Email Gateway Logs:** Failed SPF/DKIM/DMARC checks, newly registered sender domains, macro-enabled attachments (`.xlsm`, `.iso`, `.lnk`).
- **EDR / Process Creation Logs (Sysmon EventID 1):** Parent-child anomalies such as `WINWORD.EXE -> powershell.exe -enc ...` or `w3wp.exe -> cmd.exe`.

### 2.2 Common KQL Investigation Patterns

Query syntax aligned with [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/):

```kql
source == "waf" and severity == "high"
message contains "powershell" or message contains "SELECT"
source == "email" and message contains "attachment"
```

### 2.3 MITRE ATT&CK Mapping

All technique IDs link directly to the official [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/):

| Technique ID | Technique Name | Observable Telemetry Indicator |
|---|---|---|
| [T1566](https://attack.mitre.org/techniques/T1566/) | Phishing | Email gateway alert with spoofed display name, failed DMARC, or malicious link clicked before endpoint execution |
| [T1190](https://attack.mitre.org/techniques/T1190/) | Exploit Public-Facing Application | Web server spawning a shell (`nginx`/`httpd`/`w3wp.exe` -> `sh`/`cmd.exe`) or HTTP 500/200 with SQLi/webshell URI |
| [T1059](https://attack.mitre.org/techniques/T1059/) | Command and Scripting Interpreter | Encoded PowerShell (`-EncodedCommand`), obfuscated bash one-liners, or unexpected script interpreters |
| [T1486](https://attack.mitre.org/techniques/T1486/) | Data Encrypted for Impact | High-volume file rename/encryption events, shadow copy deletion (`vssadmin delete shadows`), and ransom note creation |

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide

### 3.1 Distinguishing True Positives (TP) from False Positives (FP)
- **True Positive Indicators:** Web shell creation following an HTTP POST to an upload endpoint; Office application spawning `powershell.exe` contacting an external C2 IP; sequential SQL injection probes followed by large database response sizes.
- **False Positive Indicators (Benign Context Logs):**
  - **Authorized Vulnerability Scanner:** High-rate port scans or SQLi/XSS test payloads originating exclusively from an authorized internal scanner IP listed in an approved scan schedule log.
  - **Penetration Test Window:** Exploitation attempts documented in a signed penetration-testing authorization record matching the exact source IP and timeframe.
  - **Legitimate Admin Script:** Signed IT automation script running during a scheduled maintenance window without external C2 beacons.

### 3.2 3-Phase Playbook Response Principles
Guided by [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final), [NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final), and [CISA #StopRansomware](https://www.cisa.gov/stopransomware):
- **Phase 1 — Containment:**
  - *Correct:* Isolate the infected host via EDR, block the attacker's C2 IP/domain at the firewall/proxy, quarantine malicious emails across mailboxes, or place the vulnerable web endpoint behind a strict WAF rule.
  - *Distractors to Avoid:* Deleting all web application databases, shutting down the entire corporate internet gateway, or ignoring an active web shell.
- **Phase 2 — Eradication:**
  - *Correct:* Terminate malicious processes, remove dropped web shells/persistence entries (Registry Run keys, cron jobs, scheduled tasks), patch the exploited CVE, and rotate exposed service credentials.
  - *Distractors to Avoid:* Merely rebooting the machine (persistence survives reboot) or leaving the vulnerable software version exposed.
- **Phase 3 — Recovery:**
  - *Correct:* Restore altered files from verified clean backups, validate application integrity, re-enable host network access gradually, and monitor SIEM/EDR for re-infection.
  - *Distractors to Avoid:* Restoring unpatched backups directly back to production without fixing the entry vector.

---

## 4. Key Takeaways & References

- **Correlate Across Layers:** In Domain 2 scenarios, link the initial threat vector (e.g., WAF HTTP payload or phishing email) to post-exploitation telemetry (e.g., spawned shell or outbound C2 connection).
- **Check Benign Authorization First:** Automated vulnerability scans and authorized pentests look identical to attacks in firewall/WAF logs—always check for an authorized scanner or change record before declaring a True Positive.
- **Eradicate Persistence:** Containment stops ongoing impact, but Eradication must remove the root vulnerability and persistence mechanism before Recovery.

### Authoritative Web References
- **Wikipedia (Open Reference):**
  - [Phishing](https://en.wikipedia.org/wiki/Phishing) & [Social Engineering (Security)](https://en.wikipedia.org/wiki/Social_engineering_%28security%29)
  - [Malware](https://en.wikipedia.org/wiki/Malware) & [Ransomware](https://en.wikipedia.org/wiki/Ransomware)
  - [SQL Injection](https://en.wikipedia.org/wiki/SQL_injection) & [Cross-site Scripting (XSS)](https://en.wikipedia.org/wiki/Cross-site_scripting)
- **Government & Standards Agencies (CISA & NIST):**
  - [CISA: Recognize and Mitigate Phishing](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)
  - [CISA #StopRansomware Resource Center](https://www.cisa.gov/stopransomware)
  - [NIST SP 800-61 Rev. 2: Computer Security Incident Handling Guide](https://csrc.nist.gov/pubs/sp/800/61/r2/final) & [NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final)
- **Open Knowledge Bases & Universities:**
  - [OWASP Top 10:2021 Web Application Security Risks](https://owasp.org/Top10/)
  - [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/) (`T1566`, `T1190`, `T1059`, `T1486`)
  - [Stanford University Information Security (SISA)](https://uit.stanford.edu/security)
  - [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/)
