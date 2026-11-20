# Domain 1: แนวคิดความปลอดภัยทั่วไป (General Security Concepts)

> **Domain ID:** `domain1` | **สัดส่วนในข้อสอบ:** 12% (4 จาก 30 ข้อใน Pre-Test / Post-Test) | **แหล่งอ้างอิง:** Wikipedia, NIST CSRC, CISA, MITRE ATT&CK

บทเรียนนี้สรุปหลักการพื้นฐานด้านความมั่นคงปลอดภัยไซเบอร์โดยอ้างอิงจากแหล่งมาตรฐานเปิดและหน่วยงานภาครัฐ ได้แก่ [Wikipedia: Information Security](https://en.wikipedia.org/wiki/Information_security), [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final) และ [NIST CSRC Glossary](https://csrc.nist.gov/glossary) พร้อมเชื่อมโยงเข้ากับการวิเคราะห์ข้อมูล Telemetry ใน SOC, การคัดแยกสัญญาณเตือน (Alert Triage) และแนวทางปฏิบัติตาม Incident Response Playbook

---

## 1. ทฤษฎีหลัก CompTIA Security+ และแนวคิดการสอบ (Core Security+ Theory)

### 1.1 หลักการ CIA Triad และการห้ามปฏิเสธความรับผิดชอบ (Non-Repudiation)
อ้างอิงจาก [Wikipedia: Information Security](https://en.wikipedia.org/wiki/Information_security) และ [NIST CSRC Glossary](https://csrc.nist.gov/glossary) เป้าหมายหลักของความมั่นคงปลอดภัยสารสนเทศประกอบด้วย 3 เสาหลักควบคู่กับการยืนยันความรับผิดชอบ:
- **Confidentiality (การรักษาความลับ):** ป้องกันไม่ให้ข้อมูลถูกเปิดเผยต่อบุคคล หน่วยงาน หรือโปรเซสที่ไม่ได้รับอนุญาต โดยบังคับใช้ผ่านการเข้ารหัส, รายการควบคุมการเข้าถึง ([Access Control](https://en.wikipedia.org/wiki/Access_control)) และการจัดระดับชั้นข้อมูล (Data Classification)
- **Integrity (ความถูกต้องสมบูรณ์):** ป้องกันข้อมูลจากการแก้ไข ดัดแปลง หรือทำลายโดยไม่ได้รับอนุญาตหรือไม่ถูกตรวจพบ เพื่อคงความแม่นยำและความน่าเชื่อถือตลอดวงจรชีวิตข้อมูล ด้วย Cryptographic Hashing (`SHA-256`) และลายมือชื่อดิจิทัล (Digital Signatures)
- **Availability (ความพร้อมใช้งาน):** ทำให้ผู้ใช้งานที่ได้รับอนุญาตสามารถเข้าถึงระบบและข้อมูลได้อย่างทันท่วงทีและเชื่อถือได้ ผ่านระบบสำรอง (Redundancy), Failover, การสำรองข้อมูล (Backups) และการบรรเทาภัย DoS
- **Non-Repudiation (การห้ามปฏิเสธความรับผิดชอบ):** การรับประกันหลักฐานการส่งและการระบุตัวตนของผู้ส่ง เพื่อไม่ให้ฝ่ายใดสามารถปฏิเสธการดำเนินการกับข้อมูลนั้นได้ในภายหลัง โดยอาศัย [Public-key cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography), ลายมือชื่อดิจิทัล และ Audit Logs ที่ป้องกันการแก้ไข

### 1.2 หมวดหมู่และหน้าที่ของ Security Controls
ตามมาตรฐาน [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final) และโครงสร้างการกำกับดูแลความปลอดภัย การควบคุมถูกจำแนกออกเป็น 2 มิติหลัก:

| มิติ (Axis) | ประเภท (Types) | ตัวอย่างการใช้งานจริง (Practical Examples) |
|---|---|---|
| Control Category (หมวดหมู่กลไก) | Technical, Managerial, Operational, Physical | Firewall และ MFA (Technical); นโยบายความเสี่ยง (Managerial); การอบรมสร้างความตระหนักรู้ (Operational); เครื่องอ่านบัตรและเสากั้น (Physical) |
| Control Function (หน้าที่การทำงาน) | Preventive, Detective, Corrective, Deterrent, Compensating, Directive | MFA และ ACL (Preventive); การแจ้งเตือน SIEM และ IDS (Detective); การแพตช์และกู้คืนข้อมูล (Corrective); ป้ายคำเตือน (Deterrent); การแยก VLAN สำหรับระบบเก่า (Compensating); นโยบาย Acceptable Use Policy (Directive) |

### 1.3 การควบคุมการเข้าถึงและกรอบแนวคิด AAA (Access Control & AAA Framework)
อ้างอิงตาม [Wikipedia: Authentication](https://en.wikipedia.org/wiki/Authentication) และ [Wikipedia: Access Control](https://en.wikipedia.org/wiki/Access_control):
- **Authentication (การพิสูจน์ตัวตน):** ตรวจสอบตัวตนของผู้ใช้ โปรเซส หรืออุปกรณ์ก่อนอนุญาตให้เข้าถึง นิยมเสริมความแข็งแกร่งด้วย Multi-Factor Authentication (MFA) ที่ผสานสิ่งที่รู้ (รหัสผ่าน), สิ่งที่มี (FIDO2 token / smart card) หรือสิ่งที่เป็นตัวตน (Biometrics)
- **Authorization (การกำหนดสิทธิ์):** มอบหรือจำกัดสิทธิ์ตามตัวตนที่ยืนยันแล้ว (RBAC, ABAC, MAC, DAC และหลัก Least Privilege)
- **Accounting / Auditing (การบันทึกตรวจสอบ):** บันทึกผลการยืนยันตัวตน การกำหนดสิทธิ์ และกิจกรรมการเข้าถึงลงใน Audit Logs ส่วนกลางเพื่อใช้ตรวจสอบย้อนหลังและยืนยัน Non-Repudiation

### 1.4 สถาปัตยกรรม Zero Trust Architecture (ZTA)
กำหนดไว้ใน [NIST SP 800-207: Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final), [CISA Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model) และ [Wikipedia: Zero Trust Security Model](https://en.wikipedia.org/wiki/Zero_trust_security_model):
- **หลักการแกนกลาง ("Never trust, always verify"):** การอยู่ในเครือข่าย LAN องค์กรหรือ VPN ไม่ได้หมายความว่าเชื่อถือได้โดยอัตโนมัติ ทุกคำขอเข้าถึงต้องถูกประเมินตามตัวตน สถานะความปลอดภัยของอุปกรณ์ (Device Posture) และหลัก Least Privilege
- **องค์ประกอบเชิงตรรกะ (Logical Components):** **Policy Engine (PE)** ประเมินค่าความเสี่ยงและความน่าเชื่อถือจาก telemetry, **Policy Administrator (PA)** ทำหน้าที่สร้างหรือเพิกถอนเส้นทางเซสชัน และ **Policy Enforcement Point (PEP)** คั่นกลางและควบคุมการเชื่อมต่อระหว่างผู้ใช้งานกับทรัพยากรองค์กร

### 1.5 พื้นฐานการเข้ารหัสลับและ PKI (Cryptographic Fundamentals & PKI)
อธิบายไว้ใน [Wikipedia: Public-key Cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography) และ [Wikipedia: Public Key Infrastructure](https://en.wikipedia.org/wiki/Public_key_infrastructure):
- **Symmetric Encryption (การเข้ารหัสแบบสมมาตร):** ใช้กุญแจลับชุดเดียวกัน (เช่น AES-128/256) สำหรับเข้ารหัสข้อมูลปริมาณมากได้รวดเร็ว ทั้งข้อมูลจัดเก็บ (Data at Rest) และทราฟฟิก TLS
- **Asymmetric Encryption (การเข้ารหัสแบบอสมมาตร):** ใช้คู่กุญแจ Public Key และ Private Key ที่สัมพันธ์กันทางคณิตศาสตร์ (RSA, ECC) สำหรับแลกเปลี่ยนกุญแจ การพิสูจน์ตัวตน TLS และลายมือชื่อดิจิทัล
- **PKI & Trust Stores:** ใบรับรอง X.509 ที่ลงนามโดย Certificate Authority (CA) ช่วยผูก Public Key เข้ากับตัวตนที่ตรวจสอบแล้ว และติดตามการเพิกถอนใบรับรองผ่าน CRL และ OCSP

---

## 2. ข้อมูล SOC Telemetry, Log Indicators และการเชื่อมโยง MITRE ATT&CK

เหตุการณ์ใน Domain 1 มักเกี่ยวข้องกับการใช้ช่องโหว่ด้านการพิสูจน์ตัวตน, การโจมตีแบบ MFA Fatigue/Bypass, การแทรกแซง PKI Trust หรือการเปลี่ยนแปลงสิทธิ์โดยไม่ได้รับอนุญาต

### 2.1 แหล่ง Log และฟิลด์สำคัญ (Key Log Sources & Fields)
- **Identity Provider (IdP) / Entra ID / Okta Logs:** `user`, `src_ip`, `mfa_result`, `auth_method`, `device_trust`, `app_name`
- **Windows Security Event Log:** `EventID 4624` (Logon สำเร็จ), `4625` (Logon ล้มเหลว), `4672` (การมอบสิทธิ์พิเศษ), `4768/4769` (คำขอ Kerberos TGT/TGS)
- **Certificate & PKI Audit Logs:** การออกใบรับรองใหม่, คำเตือนใบรับรองหมดอายุ/Self-Signed TLS, ความล้มเหลวในการตรวจ OCSP/CRL

### 2.2 รูปแบบการสืบค้น KQL ที่พบบ่อย

ไวยากรณ์การสืบค้นอ้างอิงตาม [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/):

```kql
source == "identity" and mfa_result == "denied"
source == "auth" and status == "failed"
event_id == "4672" or message contains "privilege"
```

### 2.3 การเชื่อมโยง MITRE ATT&CK

รหัสเทคนิคทั้งหมดเชื่อมโยงโดยตรงไปยัง [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/):

| Technique ID | ชื่อเทคนิค (Technique Name) | สัญญาณ Telemetry ที่สังเกตได้ (Observable Indicator) |
|---|---|---|
| [T1621](https://attack.mitre.org/techniques/T1621/) | Multi-Factor Authentication Request Generation | คำขอ MFA Push ถูกปฏิเสธถี่รัวติดต่อกัน ตามด้วยคำขอที่ได้รับอนุมัติ 1 ครั้งจาก IP ใหม่ |
| [T1078](https://attack.mitre.org/techniques/T1078/) | Valid Accounts | การเข้าสู่ระบบสำเร็จ (`4624`) จากพื้นที่/ASN ผิดปกติโดยใช้บัญชีที่ถูกต้อง |
| [T1556](https://attack.mitre.org/techniques/T1556/) | Modify Authentication Process | การเปลี่ยนแปลง PAM modules, IdP federation trust หรือการตั้งค่า MFA โดยไม่คาดคิด |
| [T1553](https://attack.mitre.org/techniques/T1553/) | Subvert Trust Controls | การติดตั้ง Root CA certificate ที่ไม่น่าเชื่อถือลงบนเครื่องปลายทาง (Endpoint) |

---

## 3. การคัดแยกสัญญาณเตือน (Alert Triage: TP vs. FP) และแนวทาง Playbook 3 ระยะ

### 3.1 การจำแนก True Positive (TP) ออกจาก False Positive (FP)
- **สัญญาณบ่งชี้ True Positive (เหตุภัยคุกคามจริง):** มีคำขอ MFA ล้มเหลว/ปฏิเสธซ้ำๆ ([T1621](https://attack.mitre.org/techniques/T1621/)) ตามด้วยการอนุมัติจาก ASN/ประเทศที่ไม่รู้จัก; การนำ Token กลับมาใช้จาก IP ที่อยู่ห่างกันมากภายในไม่กี่นาที; หรือมีการติดตั้ง Root Certificate โดยไม่ได้รับอนุญาตบนเครื่อง Production
- **สัญญาณบ่งชี้ False Positive (Benign Context Logs / สัญญาณเตือนลวง):**
  - **Approved Change Ticket:** การหมุนเวียนรหัสผ่านของ Service Account หรือการต่ออายุใบรับรองที่ปรากฏคู่กับเลขใบคำขอเปลี่ยนระบบ `CHG-XXXX` ใน Telemetry ของระบบ Change Management
  - **HR Travel Confirmation:** การแจ้งเตือน "Impossible Travel" หรือการล็อกอินจากต่างประเทศที่มีบันทึกอนุมัติการเดินทางจากระบบ HR มายืนยันตรงกับพนักงานรายนั้น
  - **Scheduled Maintenance:** การแจ้งเตือนใบรับรองใกล้หมดอายุหรือการรีสตาร์ทบริการที่เกิดขึ้นภายในช่วง Maintenance Window ที่ได้รับอนุมัติ

### 3.2 หลักการตอบสนองตาม Playbook 3 ระยะ
อ้างอิงตาม [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final), [NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final) และ [CMU Computer Security Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html):
- **Phase 1 — การควบคุมขอบเขต (Containment):**
  - *แนวทางที่ถูกต้อง:* เพิกถอน Session Tokens ของผู้ใช้ที่ถูกบุกรุก, ปิดการใช้งานบัญชีผู้ใช้ที่ได้รับผลกระทบชั่วคราว หรือบล็อก IP ต้นทางที่เป็นอันตรายที่ IdP / Conditional Access Policy
  - *ตัวหลอกที่ต้องหลีกเลี่ยง (Distractors):* ปิดการบังคับใช้ MFA ทั้งองค์กร, สั่งปิด Domain Controller ทุกตัวทันที หรือลบบัญชี/ไดเรกทอรีผู้ใช้ออกถาวร
- **Phase 2 — การกำจัดภัยคุกคาม (Eradication):**
  - *แนวทางที่ถูกต้อง:* บังคับเปลี่ยนรหัสผ่านและลงทะเบียนอุปกรณ์ MFA ใหม่, ลบแอป OAuth/Federation แปลกปลอม หรือถอดใบรับรองที่ไม่น่าเชื่อถือออกจาก Trust Store
  - *ตัวหลอกที่ต้องหลีกเลี่ยง:* ล้างค่า Network Switch ทั้งหมด หรือข้ามการเปลี่ยนรหัสผ่านหลังเกิดเหตุละเมิดการพิสูจน์ตัวตน
- **Phase 3 — การฟื้นฟูระบบ (Recovery):**
  - *แนวทางที่ถูกต้อง:* เปิดใช้งานบัญชีที่ตรวจสอบแล้วอีกครั้งภายใต้การเฝ้าระวังเป็นพิเศษ, บังคับใช้ MFA ที่ทนต่อฟิชชิง (FIDO2) และบันทึกปิดเหตุการณ์ตามขั้นตอน Change Control
  - *ตัวหลอกที่ต้องหลีกเลี่ยง:* ยกเลิกกฎ Conditional Access ทั้งหมดเพื่อเร่งให้ผู้ใช้ล็อกอินได้เร็วขึ้นโดยไม่ตรวจสอบ

---

## 4. สรุปประเด็นสำคัญและแหล่งอ้างอิง (Key Takeaways & References)

- **ก่อนวิเคราะห์ทุกสถานการณ์ (Simulation):** ตรวจสอบก่อนเสมอว่าเหตุผิดปกติด้าน Identity หรือ Certificate มี **Benign Context Log** (เช่น Change Ticket, บันทึกเดินทาง HR, ตาราง Maintenance) ยืนยันอยู่หรือไม่ ก่อนด่วนสรุปเป็น True Positive
- **กติกาการเลือกหลักฐาน (Evidence Selection):** เลือกเฉพาะแถว Log ที่พิสูจน์ *ทั้ง* เหตุการณ์ผิดปกติและ (ในกรณี FP) บันทึกอนุญาตที่ยืนยันว่าปลอดภัย หลีกเลี่ยงการเดาสุ่มเลือก Noise Logs ที่ไม่เกี่ยวข้อง (ถูกหัก `-2` คะแนนต่อ 1 Noise Log)
- **วินัยในการเลือก Playbook:** เลือกการควบคุมขอบเขตและกำจัดภัยให้ตรงจุดกับ Identity หรือ Trust Control ที่ถูกกระทบเท่านั้น ห้ามเลือกมาตรการที่ทำลายระบบในวงกว้างเกินเหตุ

### แหล่งอ้างอิงเว็บมาตรฐาน (Authoritative Web References)
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
