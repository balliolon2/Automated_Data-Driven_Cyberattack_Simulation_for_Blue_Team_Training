# Domain 1: General Security Concepts

> **Domain ID:** `domain1` | **CompTIA Security+ (SY0-701) Exam Weight:** 12% | **Pre/Post-Test Quota:** 4 / 30 Questions

This module covers foundational security concepts tested in Domain 1 and explains how principles such as the CIA Triad, AAA, Zero Trust, security controls, and cryptography manifest in real-world SOC telemetry, alert triage, and incident response playbooks.

---

## 1. Core Security+ Theory & Exam Concepts

### 1.1 The CIA Triad & Non-Repudiation
The **CIA Triad** defines the three primary goals of information security [R5]:
- **Confidentiality:** Ensuring information is accessible only to authorized individuals, entities, or processes. Enforced through encryption (AES, TLS), access control lists (ACLs), and data classification.
- **Integrity:** Maintaining the accuracy, completeness, and trustworthiness of data over its entire lifecycle. Enforced through cryptographic hashing (SHA-256, SHA-3), digital signatures, HMAC, and file integrity monitoring (FIM).
- **Availability:** Ensuring reliable and timely access to systems and data for authorized users. Enforced through redundancy, clustering, failover, backups, and DDoS mitigation.
- **Non-Repudiation:** Guaranteeing that the sender or actor cannot credibly deny having performed an action or sent a message. Achieved via asymmetric digital signatures, PKI certificates, and tamper-evident audit logs [R2, R5].

### 1.2 Security Control Categories and Functional Types
Under NIST SP 800-53 and CompTIA Security+ [R2, R3], security controls are classified along two axes:

| Axis | Types | Practical Examples |
|---|---|---|
| Control Category | Technical, Managerial, Operational, Physical | Firewalls and MFA (Technical); Risk policies (Managerial); Security awareness training (Operational); Badge readers and bollards (Physical) |
| Control Function | Preventive, Detective, Corrective, Deterrent, Compensating, Directive | MFA and ACLs (Preventive); SIEM alerts and IDS (Detective); Patching and backup restore (Corrective); Warning banners (Deterrent); Network isolation for unpatchable legacy hosts (Compensating); Acceptable Use Policy (Directive) |

### 1.3 AAA Framework (Authentication, Authorization, Accounting)
- **Authentication:** Verifying who a subject is using factors: *something you know* (password/PIN), *something you have* (hardware token/FIDO2/smart card), *something you are* (biometrics), *somewhere you are* (geolocation), or *something you do* (keystroke dynamics).
- **Authorization:** Determining what actions or resources an authenticated subject may access (RBAC, ABAC, MAC, DAC, and the Principle of Least Privilege).
- **Accounting (Auditing):** Recording every authentication attempt, privilege change, and resource access in centralized logs for forensic traceability and non-repudiation.

### 1.4 Zero Trust Architecture
**Zero Trust** operates on the principle *"Never trust, always verify"* [R2, R5]. Network location (such as being inside a corporate LAN or VPN) never implies implicit trust. Key pillars include:
- **Control Plane vs. Data Plane separation:** Adaptive identity and device health verification occur before granting a session to the data plane.
- **Policy Engine (PE) & Policy Administrator (PA):** Evaluate dynamic context (user identity, MFA state, device posture, geolocation, anomalous behavior) to approve or revoke session tokens.
- **Policy Enforcement Point (PEP):** Enforces allow/deny decisions at micro-perimeters around each application or workload.

### 1.5 Cryptographic Fundamentals & PKI
- **Symmetric Encryption:** Uses one shared secret key (e.g., AES-128/256, ChaCha20). Fast for bulk data at rest, but requires secure key exchange.
- **Asymmetric Encryption:** Uses mathematically linked public/private key pairs (e.g., RSA, ECC, Diffie-Hellman/ECDH). Enables key exchange, digital signatures, and Public Key Infrastructure (PKI).
- **Hashing & Salting:** One-way deterministic digests (SHA-256) verify integrity; salted adaptive hashes (Argon2, bcrypt) protect stored passwords against rainbow tables.
- **PKI & Certificates:** X.509 certificates bind a public key to an identity, signed by a Certificate Authority (CA). Revocation is checked via CRL or OCSP.

### 1.6 Change Management & Deception Technologies
- **Change Management:** Formal approval workflows, maintenance windows, rollback plans, and impact analysis prevent unauthorized modifications and help SOC analysts verify whether an alert was caused by planned work.
- **Deception & Disruption:** Honeypots, honeynets, honeyfiles, and honeytokens attract adversaries to decoy assets so any interaction produces a high-fidelity alert.

---

## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping

Domain 1 incidents frequently involve authentication abuse, MFA bypass, PKI tampering, or unauthorized privilege changes.

### 2.1 Key Log Sources & Fields
- **Identity Provider (IdP) / Entra ID / Okta Logs:** `user`, `src_ip`, `mfa_result`, `auth_method`, `device_trust`, `app_name`.
- **Windows Security Event Log:** `EventID 4624` (Logon success), `4625` (Logon failure), `4672` (Special privileges assigned), `4768/4769` (Kerberos TGT/TGS requests).
- **Certificate & PKI Audit Logs:** Certificate issuance, expired/self-signed TLS errors, OCSP/CRL validation failures.

### 2.2 Common KQL Investigation Patterns

```kql
source == "identity" and mfa_result == "denied"
source == "auth" and status == "failed"
event_id == "4672" or message contains "privilege"
```

### 2.3 MITRE ATT&CK Mapping [R4]

| Technique ID | Technique Name | Observable Telemetry Indicator |
|---|---|---|
| T1621 | Multi-Factor Authentication Request Generation | Rapid burst of denied MFA push prompts followed by a single accepted prompt from a new IP |
| T1078 | Valid Accounts | Successful login (`4624`) from impossible travel or outside business hours using a valid credential |
| T1556 | Modify Authentication Process | Unexpected modification to PAM modules, IdP federation trust, or MFA enrollment settings |
| T1553 | Subvert Trust Controls | Installation of an untrusted root CA certificate on an endpoint |

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide

### 3.1 Distinguishing True Positives (TP) from False Positives (FP)
- **True Positive Indicators:** Repeated failed MFA prompts (`MFA push fatigue`) followed by an approval from an unrecognized ASN/country; token reuse from two distant IPs within minutes; unauthorized self-signed certificates on production services.
- **False Positive Indicators (Benign Context Logs):**
  - **Approved Change Ticket:** A service account password rotation or certificate renewal logged alongside a matching `CHG-XXXX` ticket in change-management telemetry.
  - **HR Travel Confirmation:** An "impossible travel" or foreign IP login corroborated by an HR travel approval log for that employee.
  - **Scheduled Maintenance:** Certificate expiry alerts or service restarts occurring inside an authorized maintenance window.

### 3.2 3-Phase Playbook Response Principles [R3]
- **Phase 1 — Containment (Limit Blast Radius):**
  - *Correct:* Revoke compromised user session tokens, temporarily disable the affected user account, or block the malicious source IP at the IdP/conditional access policy.
  - *Distractors to Avoid:* Disabling MFA enforcement globally, powering off all domain controllers, or permanently deleting user directories.
- **Phase 2 — Eradication (Remove Root Cause):**
  - *Correct:* Force password reset and re-enroll MFA devices, remove rogue OAuth/federation apps, or purge untrusted certificates from the trust store.
  - *Distractors to Avoid:* Reformatting network switches or skipping credential rotation after an authentication compromise.
- **Phase 3 — Recovery (Restore & Verify):**
  - *Correct:* Re-enable the verified user account under enhanced monitoring, enforce phishing-resistant MFA (FIDO2), and document the incident under change control.
  - *Distractors to Avoid:* Lifting all conditional access rules to speed up user login without verification.

---

## 4. Key Takeaways & References

- **Before every Simulation:** Check whether an identity or certificate anomaly has a **Benign Context Log** (change ticket, HR travel record, maintenance schedule) before classifying it as a True Positive.
- **Evidence Selection Rule:** Select the exact log lines proving *both* the anomalous authentication/control event and (for FP) the authorizing benign context record. Avoid guessing unrelated background logs (`-2` points per noise log).
- **Playbook Discipline:** Match containment and eradication strictly to the compromised identity or trust control—never choose sweeping destructive actions.

### References
- `[R2]` CompTIA, Inc. (2023). *CompTIA Security+ Certification Exam Objectives (Exam SY0-701)* — Domain 1.0 General Security Concepts.
- `[R3]` NIST. (2020). *NIST SP 800-53 Rev. 5: Security and Privacy Controls for Information Systems and Organizations*.
- `[R4]` The MITRE Corporation. (2024). *MITRE ATT&CK Enterprise Matrix* (`T1621`, `T1078`, `T1556`, `T1553`).
- `[R5]` Wikipedia contributors. (2026). *Information security*, *CIA triad*, and *Zero trust security model*.
