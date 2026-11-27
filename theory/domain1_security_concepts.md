# Domain 1: General Security Concepts (3-Minute Essential)

> **Domain 1** | **Exam Weight:** 12% (`4 / 30` Pre/Post-Test Questions) | **Reading Time:** ~3 min

Think of cybersecurity like protecting a school or bank: you need locks on doors, cameras to spot intruders, and a way to prove who entered.

---

## 1. Core Security+ Theory & Exam Concepts

- **CIA Triad + Non-Repudiation** ([Wikipedia: Information Security](https://en.wikipedia.org/wiki/Information_security)):
  - **Confidentiality:** Keep secrets private via encryption and access controls (AES, TLS).
  - **Integrity:** Detect unauthorized edits using hashes (`SHA-256`) and digital signatures.
  - **Availability:** Keep systems reachable via backups, redundancy, and DDoS defense.
  - **Non-Repudiation:** Prove sender identity so nobody can deny sending data.
- **Security Controls** ([NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final)):
  - **Categories:** *Technical* (firewall, MFA), *Managerial* (policy), *Operational* (training), *Physical* (keycards, CCTV).
  - **Functions:** *Preventive* (ACL, MFA), *Detective* (SIEM, IDS), *Corrective* (patching, restore), *Deterrent* (warning banner), *Compensating* (temporary workaround), *Directive* (policy rules).
- **AAA & Zero Trust** ([NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final)):
  - **AAA:** *Authentication* (password/MFA), *Authorization* (Least Privilege), *Accounting* (audit logs).
  - **Zero Trust ("Never trust, always verify"):** Even inside office Wi-Fi or VPN, every user and device is verified per request.
- **Cryptography & PKI** ([Wikipedia: Public-key Cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography)):
  - **Symmetric (AES):** One shared secret key — fast for bulk files.
  - **Asymmetric (RSA/ECC):** Public key encrypts/verifies, private key decrypts/signs; **PKI & CA** issue trusted X.509 certificates.

---

## 2. SOC Telemetry, KQL & MITRE ATT&CK

Domain 1 attacks target logins, MFA fatigue ([T1621](https://attack.mitre.org/techniques/T1621/)), stolen accounts ([T1078](https://attack.mitre.org/techniques/T1078/)), or fake root certificates:

```kql
source == "identity" and (mfa_result == "denied" or event_id == "4672")
```

- **Watch for:** Bursts of `mfa_result == "denied"` followed by 1 approval from a strange IP/country, or unexpected `EventID 4672` (admin privilege logon).

---

## 3. Alert Triage (TP vs. FP) & 3-Phase Playbook Guide

| True Positive (Real Attack) | False Positive (Safe Routine / Benign Context Log) |
|---|---|
| Repeated MFA rejections from an unknown country followed by a login | "Impossible travel" login backed by an **HR Travel Confirmation** log |
| Untrusted root CA certificate installed on a host without approval | Service account rotation or cert renewal matching a **`CHG-XXXX` Change Ticket** |

- **Containment:** Revoke compromised session tokens or lock the targeted account (never turn off MFA globally!).
- **Eradication:** Reset passwords, re-enroll MFA, and remove rogue apps or untrusted certificates.
- **Recovery:** Re-enable verified accounts under monitoring and enforce phishing-resistant MFA.

---

## 4. Key Takeaways & References

- Always look for a **Benign Context Log** (HR travel approval or `CHG-` change ticket) before calling an identity alert a True Positive.
- Query Reference: [Microsoft Learn: KQL](https://learn.microsoft.com/en-us/kusto/query/)
