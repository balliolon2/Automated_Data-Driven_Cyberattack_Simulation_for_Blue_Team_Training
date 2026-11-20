# 4.6 ติดตั้งและดูแลรักษาระบบจัดการตัวตนและการเข้าถึง (Implement and maintain identity and access management)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การจัดการตัวตนและสิทธิ์การเข้าถึง (`IAM`) ทำหน้าที่ตรวจสอบตัวตนและจำกัดสิทธิ์ทรัพยากรตลอดวงจรชีวิตบัญชี ([Identity Management](https://en.wikipedia.org/wiki/Identity_management), [Multi-Factor Authentication](https://en.wikipedia.org/wiki/Multi-factor_authentication), [NIST SP 800-63-3](https://csrc.nist.gov/pubs/sp/800/63/3/final)):

- **วงจรชีวิตบัญชี (Provisioning & Deprovisioning):** จัดทำระบบอัตโนมัติเชื่อมกับทะเบียน HR (Joiner-Mover-Leaver ผ่าน SCIM/SSO) เพื่อเพิกถอนบัญชีพนักงานที่ลาออกทันทีและไม่ให้เกิดบัญชีค้างตกหล่น (Orphan Service Accounts)
- **แบบจำลองการควบคุมการเข้าถึง (Access Control Models):**
  - **RBAC (Role-Based):** กำหนดสิทธิ์ตามบทบาทหน้าที่ในองค์กร
  - **ABAC (Attribute-Based):** ประเมินคุณลักษณะแบบพลวัตทั้งผู้ใช้ ทรัพยากร ตำแหน่งที่ตั้ง เวลา และบริบทความเสี่ยง (หัวใจหลักของ Zero Trust)
  - **DAC (Discretionary):** เจ้าของทรัพยากรเป็นผู้กำหนดสิทธิ์การแชร์เอง
  - **MAC (Mandatory):** ป้ายชั้นความลับระดับเคอร์เนลเป็นผู้กำหนด (`Top Secret` / `Confidential`) โดยผู้ใช้ไม่สามารถแก้ไขเองได้
  - **Rule-Based & Time-of-Day Restrictions:** ข้อจำกัดตามกฎ ACL ของ Firewall หรือช่วงเวลาการทำงาน
- **[Multi-Factor Authentication (MFA)](https://en.wikipedia.org/wiki/Multi-factor_authentication) & Passwordless:** ผสานปัจจัยอิสระ ได้แก่ **สิ่งที่รู้** (รหัสผ่าน/PIN), **สิ่งที่มี** (FIDO2 hardware security key, แอป TOTP, smart card) และ **สิ่งที่เป็น** (ไบโอเมตริกซ์) โดย FIDO2/WebAuthn ที่ต้านทานฟิชชิ่งช่วยป้องกันการโจมตีดักกลางแบบ Adversary-in-the-Middle ([MITRE ATT&CK T1621](https://attack.mitre.org/techniques/T1621/))
- **Federation & Single Sign-On (`SSO`):** **SAML 2.0** (XML assertions สำหรับ Enterprise Web SSO), **OAuth 2.0** (Token มอบอำนาจตามขอบเขต) และ **OpenID Connect (`OIDC`)** (ชั้นตัวตนแบบ JSON/JWT บน OAuth 2.0) ควบคู่กับ **Kerberos** Tickets และ **LDAP/Active Directory**
- **Privileged Access Management (`PAM`) & Just-in-Time (`JIT`):** เก็บรักษารหัสผ่านผู้ดูแลใน Vault บันทึกวิดีโอเซสชันเข้า-ออก และยกระดับสิทธิ์เพียงชั่วคราวเฉพาะช่วงทำงานที่ได้รับอนุมัติ

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

Telemetry จากระบบ IAM ช่วยแจ้งเตือนเหตุการณ์ **MFA Fatigue / Prompt Bombing** ([MITRE ATT&CK T1621](https://attack.mitre.org/techniques/T1621/)) รวมถึงการอนุญาต OAuth Consent ที่น่าสงสัยหรือการนำ Token บัญชีจริงกลับมาใช้ซ้ำ ([T1078](https://attack.mitre.org/techniques/T1078/)):

```kql
SigninLogs
| where MfaDetail.authMethod != "" and ResultType in ("50074", "500121")
| summarize DeniedPrompts = count() by UserPrincipalName, IPAddress, bin(TimeGenerated, 10m)
| where DeniedPrompts >= 3
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Identity management](https://en.wikipedia.org/wiki/Identity_management)
- [Wikipedia — Multi-factor authentication](https://en.wikipedia.org/wiki/Multi-factor_authentication)
- [Wikipedia — Authentication](https://en.wikipedia.org/wiki/Authentication)
- [NIST SP 800-63-3 — Digital Identity Guidelines](https://csrc.nist.gov/pubs/sp/800/63/3/final)
- [MITRE ATT&CK — Multi-Factor Authentication Request Generation (T1621)](https://attack.mitre.org/techniques/T1621/)
