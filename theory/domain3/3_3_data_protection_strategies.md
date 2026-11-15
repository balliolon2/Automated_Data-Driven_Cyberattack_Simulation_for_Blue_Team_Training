# 3.3 Compare and contrast concepts and strategies to protect data

## 1. Core Security+ Theory

Data protection aligns technical safeguards to **data classification**, **data states**, and jurisdictional mandates ([Data Loss Prevention](https://en.wikipedia.org/wiki/Data_loss_prevention_software), [GDPR](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)):

- **Data Classifications & Types:** Public, Internal/Private, Confidential, Restricted, and regulated categories including **PII** (Personally Identifiable Information), **PHI** (Protected Health Information), Payment Card Data (`PCI`), and Intellectual Property (`IP`).
- **Data Sovereignty & Geographical Considerations:** Laws such as the **GDPR** restrict cross-border storage/transfer and grant data subjects access/erasure rights.
- **Protection Across Data States:**
  - **Data at Rest:** Protected with AES-256 full-disk encryption (BitLocker, FileVault), database Transparent Data Encryption (`TDE`), and cloud KMS-managed bucket encryption.
  - **Data in Transit:** Protected with **TLS 1.2/1.3** (`HTTPS`), **IPsec**, or **SSH**; legacy protocols (HTTP, FTP, Telnet, SMTP without STARTTLS) expose credentials and payloads in cleartext.
  - **Data in Use:** Protected using **Confidential Computing / Secure Enclaves**, input validation, and strict memory access controls while actively processed in RAM.
- **Securing & De-identifying Methods:**
  - **Tokenization:** Replaces sensitive values (card numbers) with random non-mathematical vault references.
  - **Masking & Redaction:** Obscures characters (e.g., `****-****-****-1234`) in non-production DBs or UI displays.
  - **Anonymization vs. Pseudonymization:** Anonymization irreversibly strips re-identification paths; pseudonymization swaps identifiers via a separate protected lookup key.
  - **DLP & IRM/DRM:** **[Data Loss Prevention (DLP)](https://en.wikipedia.org/wiki/Data_loss_prevention_software)** inspects network egress, email, and USB writes to block policy violations, while Information Rights Management (`IRM`) embeds persistent view/print restrictions inside documents.

## 2. SOC & Blue-Team Operational Practice

SOC responders monitor DLP alerts, unusual archive creations (`7z`/`tar`), and large transfers over cloud APIs or DNS channels for **Exfiltration Over Web Service** ([MITRE ATT&CK T1567](https://attack.mitre.org/techniques/T1567/)) or **Alternative Protocol** ([T1048](https://attack.mitre.org/techniques/T1048/)):

```kql
CloudAppEvents
| where ActionType in ("FileDownloaded", "FileUploaded", "AnonymousLinkCreated")
| summarize ActionCount = count(), TotalBytes = sum(tolong(RawEventData.FileSizeBytes)) by AccountDisplayName, IPAddress
| where ActionCount >= 25 or TotalBytes >= 500000000
```

## 3. References (Web Sources)

- [Wikipedia — Data loss prevention software](https://en.wikipedia.org/wiki/Data_loss_prevention_software)
- [Wikipedia — General Data Protection Regulation](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)
- [NIST SP 800-53 Rev. 5 — System and Communications Protection (SC)](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Exfiltration Over Web Service (T1567)](https://attack.mitre.org/techniques/T1567/)
- [MITRE ATT&CK — Exfiltration Over Alternative Protocol (T1048)](https://attack.mitre.org/techniques/T1048/)
