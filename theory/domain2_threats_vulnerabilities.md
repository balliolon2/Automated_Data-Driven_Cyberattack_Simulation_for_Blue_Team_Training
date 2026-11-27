# Domain 2: Threats, Vulnerabilities, and Mitigations (3-Minute Essential)

> **Domain 2** | **Exam Weight:** 22% (`7 / 30` Pre/Post-Test Questions) | **Reading Time:** ~3 min

If Domain 1 is the lock on your door, Domain 2 is knowing *who* is rattling the doorknob and *what tricks* they use — from fake school-announcement messages to sneaky web inputs and file-locking ransomware.

---

## 1. Core Security+ Theory & Exam Concepts

- **Threat Actors & Motivations:**
  - *Nation-State (APT):* Well-funded espionage or critical-infrastructure groups playing the long game.
  - *Organized Crime / Ransomware:* Financial extortion groups locking files and leaking stolen data ([CISA #StopRansomware](https://www.cisa.gov/stopransomware)).
  - *Hacktivists / Insiders / Unskilled:* Cause-driven disruptors, trusted employees misusing access, or script runners using ready-made exploit kits.
- **Social Engineering & Phishing** ([Wikipedia: Phishing](https://en.wikipedia.org/wiki/Phishing), [CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)):
  - Tricks people rather than software: *Phishing* (mass email), *Spear Phishing* (targeted), *Whaling* (executives/BEC wire fraud), *Smishing* (SMS), *Vishing* (voice), and *Typosquatting* (look-alike domains like `sch00l-login.org`).
  - **Email Defenses:** *SPF* verifies sender IP, *DKIM* checks cryptographic signature, and *DMARC* sets the reject/quarantine policy.
- **OWASP Web Vulnerabilities & Mitigations** ([OWASP Top 10:2021](https://owasp.org/Top10/)):
  - *Broken Access Control:* Missing permission checks let users read someone else's record (IDOR) or traverse folders (`../../`).
  - *Injection ([SQL Injection](https://en.wikipedia.org/wiki/SQL_injection) / XSS / Command Injection):* Untrusted input runs as code (`' OR 1=1--`); neutralize with **parameterized queries** and input validation.
  - *Vulnerable Components & Auth Failures:* Unpatched servers or weak logins without MFA — patch rapidly and use a **WAF** as virtual patching.
- **Malware & Living-off-the-Land (LotL)** ([Wikipedia: Ransomware](https://en.wikipedia.org/wiki/Ransomware)):
  - *Ransomware* encrypts files and deletes backups (`vssadmin delete shadows`). *LotL* abuses built-in tools like `powershell.exe` or `certutil.exe` to blend in without dropping extra executables.

---

## 2. SOC Telemetry, KQL & MITRE ATT&CK

Watch for suspicious mail links ([T1566](https://attack.mitre.org/techniques/T1566/)), web exploits ([T1190](https://attack.mitre.org/techniques/T1190/)), encoded PowerShell ([T1059](https://attack.mitre.org/techniques/T1059/)), and mass file encryption ([T1486](https://attack.mitre.org/techniques/T1486/)) tracked in [MITRE ATT&CK®](https://attack.mitre.org/matrices/enterprise/):

```kql
(source == "waf" and severity == "high") or (source == "edr" and message contains "powershell")
```

- **Watch for:** `WINWORD.EXE -> powershell.exe -EncodedCommand`, web servers spawning `cmd.exe`, or bursts of file renames.

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Guide

| True Positive (Real Attack) | False Positive (Safe Routine / Benign Context Log) |
|---|---|
| `w3wp.exe` spawns a shell + drops a webshell after SQLi probes | SQLi/port-scan payloads coming from an **Approved Vulnerability Scanner** IP during its scheduled window |
| Office macro launches PowerShell reaching an unknown external C2 IP | Signed IT automation script run inside an authorized maintenance ticket window |

- **Containment:** Isolate the host in EDR, quarantine the phishing emails across inboxes, and block C2 IPs/URLs on the firewall.
- **Eradication:** Kill malicious processes, remove webshells/persistence tasks, patch the CVE, and fix vulnerable queries.
- **Recovery:** Restore encrypted files from clean immutable backups and monitor for reinfection.

---

## 4. Key Takeaways & References

- Separate real attacker behavior from **authorized vulnerability scanners or pen-tests** by checking scanner IP allowlists first.
- Query Reference: [Microsoft Learn: KQL](https://learn.microsoft.com/en-us/kusto/query/)
