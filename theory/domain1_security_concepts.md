# Domain 1: General Security Concepts

> **Domain ID:** `domain1` | **Domain Scope:** 12% Weight (4 / 30 Pre/Post-Test Questions) | **Sources:** Wikipedia, NIST CSRC, CISA, MITRE ATT&CK

This module synthesizes foundational cybersecurity principles directly from open encyclopedia and government standards references—including [Wikipedia: Information Security](https://en.wikipedia.org/wiki/Information_security), [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final), and the [NIST CSRC Glossary](https://csrc.nist.gov/glossary)—and connects them to SOC telemetry, alert triage, and incident response playbooks.

---

## 1. Core Security+ Theory & Exam Concepts

### 1.1 The CIA Triad & Non-Repudiation
According to [Wikipedia: Information Security](https://en.wikipedia.org/wiki/Information_security) and the [NIST CSRC Glossary](https://csrc.nist.gov/glossary), core security objectives balance three pillars plus accountability guarantees:
- **Confidentiality:** Ensuring information is not disclosed to unauthorized individuals, entities, or processes. Enforced via encryption, [Access Control](https://en.wikipedia.org/wiki/Access_control) lists, and data classification.
- **Integrity:** Protecting data against unauthorized or undetected modification or destruction, maintaining accuracy and reliability over its lifecycle via cryptographic hashing (SHA-256) and digital signatures.
- **Availability:** Ensuring timely, reliable access to information systems and data for authorized users through redundancy, failover, backups, and DoS mitigation.
- **Non-Repudiation:** Assurance that the sender of information is provided with proof of delivery and the recipient with proof of identity so neither can deny having processed the data ([Public-key cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography), digital signatures, tamper-evident audit logs).

### 1.2 Security Control Categories and Functional Types
Under [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final) and security governance taxonomy, controls are classified along two dimensions:

| Axis | Types | Practical Examples |
|---|---|---|
| Control Category | Technical, Managerial, Operational, Physical | Firewalls and MFA (Technical); Risk policies (Managerial); Security awareness training (Operational); Badge readers and bollards (Physical) |
| Control Function | Preventive, Detective, Corrective, Deterrent, Compensating, Directive | MFA and ACLs (Preventive); SIEM alerts and IDS (Detective); Patching and backup restore (Corrective); Warning banners (Deterrent); Isolated VLANs for legacy hosts (Compensating); Acceptable Use Policy (Directive) |

### 1.3 Access Control & AAA Framework
Per [Wikipedia: Authentication](https://en.wikipedia.org/wiki/Authentication) and [Wikipedia: Access Control](https://en.wikipedia.org/wiki/Access_control):
- **Authentication:** Verifying the identity of a user, process, or device before allowing access, commonly strengthened with Multi-Factor Authentication (MFA) combining something known (password), possessed (FIDO2 token/smart card), or inherent (biometrics).
- **Authorization:** Granting or restricting permissions based on verified identity (RBAC, ABAC, MAC, DAC, and Least Privilege).
- **Accounting (Auditing):** Recording authentication outcomes, privilege assignments, and access events in centralized audit logs for traceability and non-repudiation.

### 1.4 Zero Trust Architecture (ZTA)
Defined in [NIST SP 800-207: Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final), the [CISA Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model), and [Wikipedia: Zero Trust Security Model](https://en.wikipedia.org/wiki/Zero_trust_security_model):
- **Core Tenet ("Never trust, always verify"):** Network location inside a corporate LAN or VPN never implies implicit trust. Access is evaluated per request based on identity, device posture, and least privilege.
- **Logical Components:** A **Policy Engine (PE)** evaluates risk/trust telemetry, a **Policy Administrator (PA)** establishes or revokes session paths, and a **Policy Enforcement Point (PEP)** mediates connections between subjects and enterprise resources.

### 1.5 Cryptographic Fundamentals & PKI
Described in [Wikipedia: Public-key Cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography) and [Wikipedia: Public Key Infrastructure](https://en.wikipedia.org/wiki/Public_key_infrastructure):
- **Symmetric Encryption:** Uses a single shared secret key (e.g., AES-128/256) for fast bulk encryption of data at rest and symmetric TLS traffic.
- **Asymmetric Encryption:** Relies on mathematically related public and private key pairs (RSA, ECC) for key exchange, TLS authentication, and digital signatures.
- **PKI & Trust Stores:** X.509 certificates signed by a Certificate Authority (CA) bind public keys to verified identities, with revocation tracked via CRL and OCSP.

---

## 2. SOC Telemetry, Log Indicators & MITRE ATT&CK Mapping

Domain 1 incidents frequently involve authentication abuse, MFA fatigue/bypass, PKI trust tampering, or unauthorized privilege changes.

### 2.1 Key Log Sources & Fields
- **Identity Provider (IdP) / Entra ID / Okta Logs:** `user`, `src_ip`, `mfa_result`, `auth_method`, `device_trust`, `app_name`.
- **Windows Security Event Log:** `EventID 4624` (Logon success), `4625` (Logon failure), `4672` (Special privileges assigned), `4768/4769` (Kerberos TGT/TGS requests).
- **Certificate & PKI Audit Logs:** Certificate issuance, expired/self-signed TLS warnings, OCSP/CRL validation failures.

### 2.2 Common KQL Investigation Patterns

Query syntax aligned with [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/):

```kql
source == "identity" and mfa_result == "denied"
source == "auth" and status == "failed"
event_id == "4672" or message contains "privilege"
```

### 2.3 MITRE ATT&CK Mapping

All technique IDs link directly to the official [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/):

| Technique ID | Technique Name | Observable Telemetry Indicator |
|---|---|---|
| [T1621](https://attack.mitre.org/techniques/T1621/) | Multi-Factor Authentication Request Generation | Rapid burst of denied MFA push prompts followed by a single accepted prompt from a new IP |
| [T1078](https://attack.mitre.org/techniques/T1078/) | Valid Accounts | Successful login (`4624`) from anomalous geolocation/ASN using a valid credential |
| [T1556](https://attack.mitre.org/techniques/T1556/) | Modify Authentication Process | Unexpected modification to PAM modules, IdP federation trust, or MFA enrollment settings |
| [T1553](https://attack.mitre.org/techniques/T1553/) | Subvert Trust Controls | Installation of an untrusted root CA certificate on an endpoint |

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Response Guide

### 3.1 Distinguishing True Positives (TP) from False Positives (FP)
- **True Positive Indicators:** Repeated failed MFA prompts ([T1621](https://attack.mitre.org/techniques/T1621/)) followed by an approval from an unrecognized ASN/country; token reuse from distant IPs within minutes; unauthorized root certificates installed on production hosts.
- **False Positive Indicators (Benign Context Logs):**
  - **Approved Change Ticket:** A service account password rotation or certificate renewal logged alongside a matching `CHG-XXXX` ticket in change-management telemetry.
  - **HR Travel Confirmation:** An "impossible travel" or foreign IP login corroborated by an HR travel approval log for that employee.
  - **Scheduled Maintenance:** Certificate expiry alerts or service restarts occurring inside an authorized maintenance window.

### 3.2 3-Phase Playbook Response Principles
Guided by [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final), [NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final), and the [CMU Computer Security Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html):
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

### Authoritative Web References
- **Wikipedia (Open Reference):**
  - [Information Security & CIA Triad](https://en.wikipedia.org/wiki/Information_security)
  - [Authentication](https://en.wikipedia.org/wiki/Authentication) & [Access Control](https://en.wikipedia.org/wiki/Access_control)
  - [Zero Trust Security Model](https://en.wikipedia.org/wiki/Zero_trust_security_model)
  - [Public-key Cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography) & [Public Key Infrastructure (PKI)](https://en.wikipedia.org/wiki/Public_key_infrastructure)
- **Government & Standards Agencies (NIST & CISA):**
  - [NIST SP 800-53 Rev. 5: Security and Privacy Controls](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
  - [NIST SP 800-207: Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final)
  - [NIST CSRC Security Glossary](https://csrc.nist.gov/glossary)
  - [CISA Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model)
- **Universities & Open Knowledge Bases:**
  - [CMU Information Security Office: Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)
  - [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/) (`T1621`, `T1078`, `T1556`, `T1553`)
  - [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/)
