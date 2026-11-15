# 1.2 Summarize fundamental security concepts

## 1. Core Security+ Theory

Fundamental cybersecurity design begins with the **CIA Triad** summarized in [Information Security](https://en.wikipedia.org/wiki/Information_security) and the [NIST CSRC Glossary](https://csrc.nist.gov/glossary):

- **Confidentiality:** Ensuring information is disclosed only to authorized principals (enforced via encryption, access control, and data classification).
- **Integrity:** Guaranteeing data and systems remain accurate, complete, and unaltered without authorization (verified via cryptographic hashes, HMACs, digital signatures, and write-protection).
- **Availability:** Ensuring timely, reliable access to systems and data for authorized users (supported via redundancy, failover, capacity planning, and DDoS mitigation).
- **Non-Repudiation:** Providing cryptographic proof of origin and submission—typically through asymmetric digital signatures—so a sender cannot credibly deny having signed a transaction or message.

Modern enterprise security layers **AAA (Authentication, Authorization, Accounting)** alongside **Zero Trust Architecture** ([NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final), [CISA Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model)):
- **Control Plane vs. Data Plane:** The Zero Trust policy engine/administrator evaluates identity, device health, and contextual risk in the **Control Plane** before the Policy Enforcement Point (`PEP`) allows packet flow across the **Data Plane**.
- **Deception & Disruption Technologies:** Honeypots, honeynets, honeyfiles, and honeytokens act as decoy assets that capture reconnaissance and lateral movement with zero false-positive noise because no legitimate workflow interacts with them.
- **Physical Separation:** **Air-gapped** networks physically isolate ultra-sensitive workloads (e.g., offline root Certificate Authorities or safety ICS controllers) from routable enterprise networks.

## 2. SOC & Blue-Team Operational Practice

Analysts investigate CIA and AAA violations using **Accounting** audit trails (`SigninLogs`, `SecurityEvent`, syslog). A spike in `Account` lockouts alongside successful logins from new ASNs indicates credential stuffing or valid-account misuse ([MITRE ATT&CK T1078](https://attack.mitre.org/techniques/T1078/)), while unauthorized hash changes in critical binaries point to an integrity breach:

```kql
SecurityEvent
| where EventID in (4624, 4625, 4672)
| summarize Successes = countif(EventID == 4624), Failures = countif(EventID == 4625),
            SpecialPriv = countif(EventID == 4672) by TargetUserName, IpAddress
| where Failures >= 10 and Successes >= 1
```

## 3. References (Web Sources)

- [Wikipedia — Information security](https://en.wikipedia.org/wiki/Information_security)
- [NIST SP 800-207 — Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final)
- [CISA — Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model)
- [Wikipedia — Zero trust security model](https://en.wikipedia.org/wiki/Zero_trust_security_model)
- [NIST CSRC — Computer Security Glossary](https://csrc.nist.gov/glossary)
- [MITRE ATT&CK — Valid Accounts (T1078)](https://attack.mitre.org/techniques/T1078/)
