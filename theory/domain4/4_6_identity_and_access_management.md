# 4.6 Implement and maintain identity and access management

## 1. Core Security+ Theory

Identity and Access Management (`IAM`) verifies identities and bounds resource permissions across their lifecycle ([Identity Management](https://en.wikipedia.org/wiki/Identity_management), [Multi-Factor Authentication](https://en.wikipedia.org/wiki/Multi-factor_authentication), [NIST SP 800-63-3](https://csrc.nist.gov/pubs/sp/800/63/3/final)):

- **Account Lifecycle (Provisioning & Deprovisioning):** Automating joiner-mover-leaver workflows from HR directories (SCIM/SSO) so terminated accounts are revoked immediately without orphan service accounts.
- **Access Control Models:**
  - **RBAC (Role-Based):** Permissions assigned to organizational job roles.
  - **ABAC (Attribute-Based):** Dynamic evaluation of subject, resource, location, time, and risk context (central to Zero Trust).
  - **DAC (Discretionary):** Resource owners choose sharing permissions.
  - **MAC (Mandatory):** Kernel/security labels dictate clearances (`Top Secret` / `Confidential`), non-overridable by users.
  - **Rule-Based & Time-of-Day Restrictions:** Firewall ACL rules or shift-based login blocks.
- **[Multi-Factor Authentication (MFA)](https://en.wikipedia.org/wiki/Multi-factor_authentication) & Passwordless:** Combining independent factors—**something you know** (password/PIN), **something you have** (FIDO2 hardware security key, TOTP authenticator app, smart card), and **something you are** (biometrics). Phishing-resistant FIDO2/WebAuthn prevents adversary-in-the-middle MFA relay ([MITRE ATT&CK T1621](https://attack.mitre.org/techniques/T1621/)).
- **Federation & Single Sign-On (`SSO`):** **SAML 2.0** (XML assertions for enterprise web SSO), **OAuth 2.0** (scoped delegated authorization tokens), and **OpenID Connect (`OIDC`)** (identity JSON/JWT layer atop OAuth 2.0), alongside **Kerberos** tickets and **LDAP/Active Directory**.
- **Privileged Access Management (`PAM`) & Just-in-Time (`JIT`):** Vaulting admin secrets, requiring check-in/out session recording, and elevating privileges ephemerally only during approved tasks.

## 2. SOC & Blue-Team Operational Practice

IAM telemetry flags **MFA Fatigue / Prompt Bombing** ([MITRE ATT&CK T1621](https://attack.mitre.org/techniques/T1621/)) and suspicious OAuth consent grants or valid-account token reuse ([T1078](https://attack.mitre.org/techniques/T1078/)):

```kql
SigninLogs
| where MfaDetail.authMethod != "" and ResultType in ("50074", "500121")
| summarize DeniedPrompts = count() by UserPrincipalName, IPAddress, bin(TimeGenerated, 10m)
| where DeniedPrompts >= 3
```

## 3. References (Web Sources)

- [Wikipedia — Identity management](https://en.wikipedia.org/wiki/Identity_management)
- [Wikipedia — Multi-factor authentication](https://en.wikipedia.org/wiki/Multi-factor_authentication)
- [Wikipedia — Authentication](https://en.wikipedia.org/wiki/Authentication)
- [NIST SP 800-63-3 — Digital Identity Guidelines](https://csrc.nist.gov/pubs/sp/800/63/3/final)
- [MITRE ATT&CK — Multi-Factor Authentication Request Generation (T1621)](https://attack.mitre.org/techniques/T1621/)
