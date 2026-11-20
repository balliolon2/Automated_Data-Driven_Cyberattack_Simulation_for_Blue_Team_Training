# 1.2 สรุปแนวคิดพื้นฐานด้านความมั่นคงปลอดภัย (Summarize fundamental security concepts)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การออกแบบความมั่นคงปลอดภัยไซเบอร์ขั้นพื้นฐานเริ่มต้นจาก **CIA Triad** ตามที่สรุปไว้ใน [Information Security](https://en.wikipedia.org/wiki/Information_security) และ [NIST CSRC Glossary](https://csrc.nist.gov/glossary):

- **Confidentiality (การรักษาความลับ):** ทำให้มั่นใจว่าข้อมูลจะถูกเปิดเผยเฉพาะผู้มีสิทธิ์ที่ได้รับอนุญาตเท่านั้น (บังคับใช้ผ่านการเข้ารหัส, Access Control และ Data Classification)
- **Integrity (ความถูกต้องสมบูรณ์):** รับรองว่าข้อมูลและระบบยังคงแม่นยำ ครบถ้วน และไม่ถูกแก้ไขดัดแปลงโดยไม่ได้รับอนุญาต (ตรวจสอบผ่าน Cryptographic Hashes, HMACs, Digital Signatures และการป้องกันการเขียนทับ)
- **Availability (ความพร้อมใช้งาน):** ทำให้ผู้ใช้งานที่ได้รับอนุญาตสามารถเข้าถึงระบบและข้อมูลได้อย่างต่อเนื่องและทันเวลา (สนับสนุนด้วย Redundancy, Failover, Capacity Planning และการบรรเทาภัย DDoS)
- **Non-Repudiation (การห้ามปฏิเสธความรับผิดชอบ):** มอบหลักฐานเชิงการเข้ารหัสที่ยืนยันแหล่งที่มาและการส่งข้อมูล — โดยทั่วไปใช้ Asymmetric Digital Signatures — เพื่อให้ผู้ส่งไม่สามารถปฏิเสธได้ว่าตนเองไม่ได้ลงนามในธุรกรรมหรือข้อความนั้น

องค์กรยุคใหม่ผสาน **AAA (Authentication, Authorization, Accounting)** เข้ากับสถาปัตยกรรม **Zero Trust Architecture** ([NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final), [CISA Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model)):
- **Control Plane เทียบกับ Data Plane:** Policy Engine / Policy Administrator ของ Zero Trust จะประเมินตัวตน ความสมบูรณ์ของอุปกรณ์ และความเสี่ยงตามบริบทใน **Control Plane** ก่อนที่ Policy Enforcement Point (`PEP`) จะอนุญาตให้แพ็กเก็ตไหลผ่าน **Data Plane**
- **เทคโนโลยีลวงและขัดขวางผู้บุกรุก (Deception & Disruption Technologies):** Honeypots, Honeynets, Honeyfiles และ Honeytokens ทำหน้าที่เป็นทรัพย์สินตัวล่อเพื่อดักจับการลาดตระเวน (Reconnaissance) และการเคลื่อนที่แนวราบ (Lateral Movement) โดยแทบไม่มี False Positive เนื่องจากไม่มีเวิร์กโฟลว์ที่ถูกต้องใดเข้ามาแตะต้อง
- **การแยกเครือข่ายทางกายภาพ (Physical Separation):** เครือข่ายแบบ **Air-gapped** แยกระบบที่มีความอ่อนไหวสูงมาก (เช่น Offline Root Certificate Authorities หรือตัวควบคุม ICS ด้านความปลอดภัย) ออกจากเครือข่ายองค์กรที่เชื่อมต่อภายนอกได้โดยสิ้นเชิง

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

นักวิเคราะห์ SOC สืบสวนการละเมิดหลักการ CIA และ AAA ผ่านบันทึกตรวจสอบของ **Accounting** (`SigninLogs`, `SecurityEvent`, syslog) การพุ่งสูงขึ้นของเหตุบัญชีถูกล็อก (`Account` lockouts) ควบคู่กับการล็อกอินสำเร็จจาก ASN ใหม่มักชี้ถึง Credential Stuffing หรือการใช้บัญชีที่ถูกต้องในทางมิชอบ ([MITRE ATT&CK T1078](https://attack.mitre.org/techniques/T1078/)) ขณะที่การเปลี่ยนแปลงค่า Hash ของไบนารีสำคัญโดยไม่ได้รับอนุญาตชี้ถึงการละเมิด Integrity:

```kql
SecurityEvent
| where EventID in (4624, 4625, 4672)
| summarize Successes = countif(EventID == 4624), Failures = countif(EventID == 4625),
            SpecialPriv = countif(EventID == 4672) by TargetUserName, IpAddress
| where Failures >= 10 and Successes >= 1
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Information security](https://en.wikipedia.org/wiki/Information_security)
- [NIST SP 800-207 — Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final)
- [CISA — Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model)
- [Wikipedia — Zero trust security model](https://en.wikipedia.org/wiki/Zero_trust_security_model)
- [NIST CSRC — Computer Security Glossary](https://csrc.nist.gov/glossary)
- [MITRE ATT&CK — Valid Accounts (T1078)](https://attack.mitre.org/techniques/T1078/)
