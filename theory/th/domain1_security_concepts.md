# Domain 1: แนวคิดความปลอดภัยทั่วไป (สรุปกระชับ 3 นาที)

> **Domain 1** | **สัดส่วนในข้อสอบ:** 12% (`4 / 30` ข้อใน Pre/Post-Test) | **เวลาอ่านโดยประมาณ:** ~3 นาที

ลองนึกภาพการป้องกันข้อมูลเหมือนการล็อกโรงเรียนหรือธนาคาร: ต้องมีทั้งแม่กุญแจกันคนแอบเข้า กล้องวงจรปิดคอยเตือน และหลักฐานยืนยันตัวตนว่าใครเป็นคนทำ

---

## 1. ทฤษฎีหลัก CompTIA Security+ ที่ออกสอบบ่อย (Core Security+ Theory)

- **CIA Triad + Non-Repudiation** ([Wikipedia: Information Security](https://en.wikipedia.org/wiki/Information_security)):
  - **Confidentiality:** ปิดลับข้อมูลด้วยการเข้ารหัสและจำกัดสิทธิ์ (AES, TLS)
  - **Integrity:** จับการแอบแก้ไขข้อมูลด้วยค่าแฮช (`SHA-256`) และลายมือชื่อดิจิทัล
  - **Availability:** ระบบต้องใช้งานได้ต่อเนื่องด้วย Backup, ระบบสำรอง และกัน DDoS
  - **Non-Repudiation:** ยืนยันตัวผู้ส่งจนปฏิเสธไม่ได้ว่าไม่ได้ส่ง
- **Security Controls** ([NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final)):
  - **รูปแบบ:** *Technical* (Firewall, MFA), *Managerial* (นโยบาย), *Operational* (อบรม), *Physical* (คีย์การ์ด, CCTV)
  - **หน้าที่:** *Preventive* (กันก่อน: ACL, MFA), *Detective* (เตือนภัย: SIEM, IDS), *Corrective* (แก้ไข: แพตช์, กู้ Backup), *Deterrent* (ป้ายเตือน), *Compensating* (ตัวช่วยชั่วคราว), *Directive* (กฎข้อบังคับ)
- **AAA & Zero Trust** ([NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final)):
  - **AAA:** *Authentication* (ยืนยันตัวตน), *Authorization* (สิทธิ์เท่าที่จำเป็น — Least Privilege), *Accounting* (บันทึก Log)
  - **Zero Trust ("ไม่ไว้ใจใครอัตโนมัติ"):** ต่อให้อยู่ใน Wi-Fi หรือ VPN องค์กร ก็ต้องตรวจตัวตนและอุปกรณ์ทุกคำขอ
- **การเข้ารหัส & PKI** ([Wikipedia: Public-key Cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography)):
  - **Symmetric (AES):** กุญแจลับดอกเดียว เข้ารหัสไฟล์ใหญ่ได้เร็ว
  - **Asymmetric (RSA/ECC):** คู่กุญแจ Public/Private Key และใช้ **PKI + CA** ออกใบรับรอง X.509 ยืนยันตัวตนเว็บ

---

## 2. สัญญาณอันตรายใน SIEM Log, KQL และ MITRE ATT&CK

เหตุการณ์ Domain 1 มักเป็นการยิง MFA รัวๆ ([T1621](https://attack.mitre.org/techniques/T1621/)) หรือขโมยบัญชีจริงมาล็อกอิน ([T1078](https://attack.mitre.org/techniques/T1078/)):

```kql
source == "identity" and (mfa_result == "denied" or event_id == "4672")
```

- **จุดสังเกต:** `mfa_result == "denied"` รัวติดกันแล้วจู่ๆ ผ่านจาก IP แปลกหน้า หรือพบ `EventID 4672` (มอบสิทธิ์ Admin ผิดปกติ)

---

## 3. การคัดแยก True Positive vs. False Positive และ Playbook 3 ระยะ

| True Positive (ภัยคุกคามจริง) | False Positive (งานปกติ / มี Benign Context Log ยืนยัน) |
|---|---|
| ปฏิเสธ MFA รัวๆ จากต่างประเทศแล้วผ่านจาก IP แปลกหน้า | แจ้งเตือนล็อกอินข้ามประเทศที่มี **ใบอนุมัติเดินทางจาก HR** ยืนยัน |
| ติดตั้ง Root Certificate แปลกปลอมโดยไม่ได้รับอนุญาต | ต่ออายุใบรับรองหรือเปลี่ยนรหัสที่มีใบงาน **`CHG-XXXX`** รองรับ |

- **Containment:** ยกเลิก Session Token หรือล็อกบัญชีที่ถูกบุกรุกชั่วคราว (ห้ามปิด MFA ทั้งระบบ!)
- **Eradication:** เปลี่ยนรหัสผ่าน ลงทะเบียน MFA ใหม่ และลบแอปหรือใบรับรองแปลกปลอม
- **Recovery:** เปิดบัญชีใหม่ภายใต้การเฝ้าระวัง และบังคับใช้ MFA แบบป้องกันฟิชชิง

---

## 4. สรุปสั้นก่อนลงสนามและแหล่งอ้างอิง (Key Takeaways & References)

- ก่อนฟันธงว่า **True Positive** ให้เช็กหา **Benign Context Log** (เช่น `CHG-` หรือบันทึกเดินทาง HR) ทุกครั้ง
- คู่มือ Query: [Microsoft Learn: KQL](https://learn.microsoft.com/en-us/kusto/query/)
