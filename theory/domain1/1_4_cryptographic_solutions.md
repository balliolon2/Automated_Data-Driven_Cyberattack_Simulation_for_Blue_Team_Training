# 1.4 Explain the importance of using appropriate cryptographic solutions

## 1. Core Security+ Theory

Appropriate cryptography enforces confidentiality, integrity, and authenticity across data states, key lifecycles, and hardware trust anchors ([Public-Key Cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography), [PKI](https://en.wikipedia.org/wiki/Public_key_infrastructure), [HSM](https://en.wikipedia.org/wiki/Hardware_security_module)):

- **Symmetric vs. Asymmetric Encryption:**
  - **Symmetric** algorithms (AES-128/256, ChaCha20) share one secret key for bulk encryption at rest (Full Disk Encryption, database encryption) and symmetric TLS session payloads.
  - **Asymmetric** algorithms (RSA, ECC, Diffie–Hellman / ECDHE key exchange) use mathematically linked public/private key pairs for digital signatures, non-repudiation, and secure session key establishment.
- **Hashing & Password Protection:** One-way cryptographic hashes (SHA-256, SHA-3) verify file integrity and drive HMAC authentication. Passwords must never use bare hashes; they require unique random **salts** and key-stretching KDFs (Argon2, bcrypt, PBKDF2) to thwart rainbow tables and GPU brute-force attacks.
- **Public Key Infrastructure (PKI):**
  - A **Certificate Authority (CA)** binds public keys to verified identities in **X.509 certificates**, validated through certificate chains up to a trusted root CA.
  - **Revocation:** Compromised or expired certificates are invalidated using Certificate Revocation Lists (**CRL**) or Online Certificate Status Protocol (**OCSP** / OCSP stapling). Key escrow recovers lost decryption keys for continuity, whereas signing keys are typically withheld from escrow to preserve non-repudiation.
- **Hardware Roots of Trust & Obfuscation:**
  - **TPM** (motherboard chip storing disk/boot keys), **HSM** (tamper-responsive hardware appliance safeguarding CA root keys), and **Key Management Systems (KMS)** protect key confidentiality.
  - **Steganography** hides payloads inside cover media (not true encryption); **Tokenization** replaces sensitive values (e.g., PANs) with non-reversible surrogate tokens; **Data Masking** obscures displayed fields.

## 2. SOC & Blue-Team Operational Practice

Cryptographic hygiene breaks down when teams leave expired certificates, self-signed binaries, or weak TLS ciphers exposed. Threat actors abuse valid or stolen code-signing certificates to bypass trust controls ([MITRE ATT&CK T1553](https://attack.mitre.org/techniques/T1553/)) or encrypt enterprise disks with **Data Encrypted for Impact** ([T1486](https://attack.mitre.org/techniques/T1486/)). Analysts audit TLS handshakes and unexpected process encryption behavior:

```kql
DeviceProcessEvents
| where ProcessCommandLine has_any ("vssadmin delete shadows", "cipher /w", "openssl enc")
| project Timestamp, DeviceName, AccountName, FileName, ProcessCommandLine
```

## 3. References (Web Sources)

- [Wikipedia — Public-key cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography)
- [Wikipedia — Public key infrastructure](https://en.wikipedia.org/wiki/Public_key_infrastructure)
- [Wikipedia — Hardware security module](https://en.wikipedia.org/wiki/Hardware_security_module)
- [MITRE ATT&CK — Subvert Trust Controls (T1553)](https://attack.mitre.org/techniques/T1553/)
- [MITRE ATT&CK — Data Encrypted for Impact (T1486)](https://attack.mitre.org/techniques/T1486/)
