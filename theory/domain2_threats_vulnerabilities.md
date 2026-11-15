# Domain 2: Threats, Vulnerabilities, and Mitigations

> **Domain ID:** `domain2` | **CompTIA Security+ (SY0-701) Exam Weight:** 22% | **Pre/Post-Test Quota:** 7 / 30 Questions

This module covers threat actors, attack vectors, software and network vulnerabilities, malware families, and defensive mitigations tested in Domain 2, along with how they appear in telemetry logs and incident response playbooks.

---

## 1. Core Security+ Theory & Exam Concepts

### 1.1 Threat Actors & Motivations
CompTIA Security+ categorizes threat actors by capability, resources, and intent [R2]:
- **Nation-State / APT (Advanced Persistent Threat):** Well-funded state-sponsored groups targeting intellectual property, critical infrastructure, or espionage using stealthy, multi-stage campaigns.
- **Organized Crime / Ransomware Groups:** Financially motivated actors deploying ransomware, banking Trojans, or data-extortion schemes.
- **Hacktivists:** Ideologically or politically motivated actors conducting DDoS attacks, website defacements, or data leaks.
- **Insider Threats:** Authorized personnel (malicious or negligent) abusing legitimate privileges to exfiltrate data or sabotage systems.
- **Script Kiddies / Unskilled Actors:** Opportunistic attackers relying on off-the-shelf exploit kits and automated scanners.

### 1.2 Social Engineering & Phishing Vectors
- **Phishing / Spear Phishing / Whaling:** Deceptive messages targeting broad users, specific individuals, or high-profile executives to steal credentials or deliver payloads.
- **Vishing & Smishing:** Voice-based and SMS-based social engineering.
- **Business Email Compromise (BEC):** Impersonation or compromise of executive/vendor email accounts to redirect wire transfers or sensitive data.
- **Pretexting, Typosquatting & Watering Hole:** Fabricating an urgent scenario, registering look-alike domains (e.g., `corp-login[.]com`), or compromising websites frequented by the target group.

### 1.3 Software & Web Application Vulnerabilities (OWASP Top 10) [R5]
- **A01 Broken Access Control:** Missing authorization checks allowing IDOR (Insecure Direct Object Reference), horizontal/vertical privilege escalation, or directory traversal (`../../etc/passwd`).
- **A03 Injection (SQLi, Command Injection, XSS):** Untrusted input interpreted as executable SQL (`' OR 1=1--`), OS shell commands (`; cat /etc/shadow`), or browser scripts (Stored/Reflected/DOM XSS).
- **A06 Vulnerable & Outdated Components:** Unpatched libraries or services containing known CVEs (e.g., Log4Shell, unpatched VPN appliances).
- **A07 Identification & Authentication Failures:** Credential stuffing, brute-force attacks, weak session tokens, and missing MFA.
- **A10 Server-Side Request Forgery (SSRF):** Coercing a server to make HTTP requests to internal metadata endpoints (e.g., `169.254.169.254`) or internal APIs.
- **Memory & Race Vulnerabilities:** Buffer overflows, integer overflows, and Time-of-Check to Time-of-Use (TOCTOU) race conditions.

### 1.4 Malware & Persistence Mechanisms
- **Ransomware:** Encrypts files or locks systems while exfiltrating data for double extortion.
- **Trojans, RATs & Rootkits:** Disguised malicious programs establishing remote C2 channels or hiding kernel/user-mode artifacts.
- **Fileless / Living-off-the-Land (LotL):** Executes entirely in memory using legitimate administrative binaries (`powershell.exe`, `wmic.exe`, `certutil.exe`, `mshta.exe`).
- **Logic Bombs & Backdoors:** Dormant triggers activated by date/condition or hidden authentication bypasses.

### 1.5 Vulnerability Mitigation Techniques
- **Input Validation & Parameterized Queries:** Eliminates SQLi and command injection by separating code from data.
- **Least Privilege & Network Segmentation:** Constrains blast radius when an endpoint or web service is compromised.
- **Patch Management & Virtual Patching (WAF/IPS):** Addresses known CVEs promptly or blocks exploit payloads at the perimeter until patching is complete.
- **Application Allowlisting & EDR:** Blocks unauthorized binaries and detects anomalous script execution.

---

## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping

### 2.1 Key Log Sources & Indicators of Compromise (IoCs)
- **Web / WAF / Reverse Proxy Logs:** URI payloads containing SQL keywords (`UNION SELECT`), path traversal (`../`), shell metacharacters, or suspicious user agents (`sqlmap`, `nikto`).
- **Email Gateway Logs:** Failed SPF/DKIM/DMARC checks, newly registered sender domains, macro-enabled attachments (`.xlsm`, `.iso`, `.lnk`).
- **EDR / Process Creation Logs (Sysmon EventID 1):** Parent-child anomalies such as `WINWORD.EXE -> powershell.exe -enc ...` or `w3wp.exe -> cmd.exe`.

### 2.2 Common KQL Investigation Patterns

```kql
source == "waf" and severity == "high"
message contains "powershell" or message contains "SELECT"
source == "email" and message contains "attachment"
```

### 2.3 MITRE ATT&CK Mapping [R4]

| Technique ID | Technique Name | Observable Telemetry Indicator |
|---|---|---|
| T1566 | Phishing | Email gateway alert with spoofed display name, failed DMARC, or malicious link clicked before endpoint execution |
| T1190 | Exploit Public-Facing Application | Web server spawning a shell (`nginx`/`httpd`/`w3wp.exe` -> `sh`/`cmd.exe`) or HTTP 500/200 with SQLi/webshell URI |
| T1059 | Command and Scripting Interpreter | Encoded PowerShell (`-EncodedCommand`), obfuscated bash one-liners, or unexpected script interpreters |
| T1486 | Data Encrypted for Impact | High-volume file rename/encryption events, shadow copy deletion (`vssadmin delete shadows`), and ransom note creation |

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide

### 3.1 Distinguishing True Positives (TP) from False Positives (FP)
- **True Positive Indicators:** Web shell creation following an HTTP POST to an upload endpoint; Office application spawning `powershell.exe` contacting an external C2 IP; sequential SQL injection probes followed by large database response sizes.
- **False Positive Indicators (Benign Context Logs):**
  - **Authorized Vulnerability Scanner:** High-rate port scans or SQLi/XSS test payloads originating exclusively from an authorized internal scanner IP listed in an approved scan schedule log.
  - **Penetration Test Window:** Exploitation attempts documented in a signed penetration-testing authorization record matching the exact source IP and timeframe.
  - **Legitimate Admin Script:** Signed IT automation script running during a scheduled maintenance window without external C2 beacons.

### 3.2 3-Phase Playbook Response Principles [R3]
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

- **Correlate Across Layers:** In Domain 2 scenarios, link the initial threat vector (e.g., WAF HTTP payload or phishing email) to the post-exploitation telemetry (e.g., spawned shell or outbound C2 connection).
- **Check Benign Authorization First:** Automated vulnerability scans and pentest traffic look identical to attacks in firewall/WAF logs—always look for an authorized scanner or change record before declaring a True Positive.
- **Eradicate Persistence:** Containment stops bleeding, but Eradication must remove the root vulnerability and persistence mechanism before Recovery.

### References
- `[R2]` CompTIA, Inc. (2023). *CompTIA Security+ Certification Exam Objectives (Exam SY0-701)* — Domain 2.0 Threats, Vulnerabilities, and Mitigations.
- `[R3]` Cichonski, P., et al. (2012). *NIST SP 800-61 Rev. 2: Computer Security Incident Handling Guide*.
- `[R4]` The MITRE Corporation. (2024). *MITRE ATT&CK Enterprise Matrix* (`T1566`, `T1190`, `T1059`, `T1486`).
- `[R5]` OWASP Foundation. (2021). *OWASP Top 10:2021 — The Ten Most Critical Web Application Security Risks*.
