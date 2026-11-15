# 2.3 Explain various types of vulnerabilities

## 1. Core Security+ Theory

A [Vulnerability](https://en.wikipedia.org/wiki/Vulnerability_(computing)) is a weakness in design, code, configuration, or operations that an attacker can exploit to violate security policy ([OWASP Top 10:2021](https://owasp.org/Top10/)):

- **Application & Injection Flaws:**
  - **[SQL Injection (SQLi)](https://en.wikipedia.org/wiki/SQL_injection):** Untrusted user input alters backend database queries; prevented by parameterized queries / prepared statements.
  - **[Cross-Site Scripting (XSS)](https://en.wikipedia.org/wiki/Cross-site_scripting):** Malicious script executes in a victim's browser via reflected, stored, or DOM-based vectors; mitigated via context-aware output encoding and Content Security Policy (`CSP`).
  - **Memory Corruption & Logic Bugs:** Buffer overflows, integer overflows, **race conditions** (Time-of-Check to Time-of-Use / `TOCTOU`), insecure deserialization, SSRF, CSRF, and directory traversal (`../`).
- **OS, Firmware & Virtualization Weaknesses:** End-of-life (`EOL`) operating systems lacking vendor patches, unpatched kernel privileges, malicious UEFI/firmware implants, and **VM escape** from guest to hypervisor.
- **Misconfigurations & Cryptographic Weaknesses:** Default passwords, open cloud storage buckets, permissive CORS/ACL rules, cleartext protocols, and weak cipher suites.
- **Zero-Day & Supply-Chain Exposure:** Previously unknown flaws (`zero-day`) with no vendor patch yet available, requiring compensating network and runtime controls until remediated.

## 2. SOC & Blue-Team Operational Practice

Vulnerability triage correlates scan outputs with live exploit traffic against public web applications ([MITRE ATT&CK T1190](https://attack.mitre.org/techniques/T1190/)) or suspicious command-interpreter spawns ([T1059](https://attack.mitre.org/techniques/T1059/)). Confirming whether a WAF block (`403`) stopped an injection probe versus an application returning `500`/`200` with outbound callbacks distinguishes attempted scanning from confirmed compromise:

```kql
W3CIISLog
| where csUriQuery matches regex @"(?i)(union\s+select|<script>|../|\$\{jndi:)"
| summarize HitCount = count(), StatusCodes = make_set(scStatus) by cIP, csUriStem
| order by HitCount desc
```

## 3. References (Web Sources)

- [Wikipedia — Vulnerability (computing)](https://en.wikipedia.org/wiki/Vulnerability_(computing))
- [OWASP Top 10:2021 — Critical Web Application Security Risks](https://owasp.org/Top10/)
- [Wikipedia — SQL injection](https://en.wikipedia.org/wiki/SQL_injection)
- [Wikipedia — Cross-site scripting](https://en.wikipedia.org/wiki/Cross-site_scripting)
- [MITRE ATT&CK — Exploit Public-Facing Application (T1190)](https://attack.mitre.org/techniques/T1190/)
