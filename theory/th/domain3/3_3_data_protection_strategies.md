# 3.3 เปรียบเทียบและแยกแยะแนวคิดและกลยุทธ์ในการปกป้องข้อมูล (Compare and contrast concepts and strategies to protect data)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การปกป้องข้อมูลต้องจัดวางมาตรการทางเทคนิคให้สอดคล้องกับ **การจำแนกชั้นข้อมูล (Data Classification)**, **สถานะของข้อมูล (Data States)** และข้อบังคับตามเขตอำนาจกฎหมาย ([Data Loss Prevention](https://en.wikipedia.org/wiki/Data_loss_prevention_software), [GDPR](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)):

- **ระดับชั้นและประเภทของข้อมูล (Data Classifications & Types):** Public, Internal/Private, Confidential, Restricted และข้อมูลภายใต้ข้อกำกับ ได้แก่ **PII** (ข้อมูลส่วนบุคคล), **PHI** (ข้อมูลสุขภาพ), ข้อมูลบัตรชำระเงิน (`PCI`) และทรัพย์สินทางปัญญา (`IP`)
- **อธิปไตยของข้อมูลและข้อกำหนดเชิงพื้นที่ (Data Sovereignty & Geographical Considerations):** กฎหมายอย่าง **GDPR** จำกัดการโอน/จัดเก็บข้อมูลข้ามพรมแดนและมอบสิทธิ์แก่เจ้าของข้อมูลในการเข้าถึงหรือสั่งลบ
- **การป้องกันตามสถานะของข้อมูล (Protection Across Data States):**
  - **Data at Rest (ขณะจัดเก็บ):** ปกป้องด้วยการเข้ารหัสทั้งดิสก์ AES-256 (BitLocker, FileVault), การเข้ารหัสระดับฐานข้อมูล Transparent Data Encryption (`TDE`) และการเข้ารหัส Cloud Storage ผ่าน KMS
  - **Data in Transit (ขณะส่งผ่าน):** ปกป้องด้วย **TLS 1.2/1.3** (`HTTPS`), **IPsec** หรือ **SSH**; ส่วนโปรโตคอลเก่า (HTTP, FTP, Telnet, SMTP ที่ไม่มี STARTTLS) จะเปิดเผยข้อมูลและรหัสผ่านแบบข้อความธรรมดา
  - **Data in Use (ขณะประมวลผล):** ปกป้องด้วย **Confidential Computing / Secure Enclaves**, การตรวจสอบอินพุต และการควบคุมสิทธิ์เข้าถึงหน่วยความจำอย่างเข้มงวดขณะข้อมูลอยู่ใน RAM
- **วิธีการปกปิดและลดการระบุตัวตน (Securing & De-identifying Methods):**
  - **Tokenization:** แทนที่ค่าความลับ (เช่น เลขบัตรเครดิต) ด้วยโทเค็นสุ่มที่ไม่สัมพันธ์ทางคณิตศาสตร์จากระบบ Vault
  - **Masking & Redaction:** ปิดบังบางส่วนของอักขระ (เช่น `****-****-****-1234`) ในฐานข้อมูลทดสอบหรือหน้าจอ UI
  - **Anonymization vs. Pseudonymization:** Anonymization ลบเส้นทางย้อนกลับสู่ตัวตนอย่างถาวร; ส่วน Pseudonymization แทนที่ตัวบ่งชี้ด้วยนามแฝงที่อ้างอิงผ่านตารางคีย์แยกต่างหากที่มีการป้องกัน
  - **DLP & IRM/DRM:** **[Data Loss Prevention (DLP)](https://en.wikipedia.org/wiki/Data_loss_prevention_software)** ตรวจสอบการส่งออกข้อมูลผ่านเครือข่าย อีเมล และ USB เพื่อบล็อกการละเมิดนโยบาย ขณะที่ Information Rights Management (`IRM`) ฝังข้อจำกัดการเปิดดู/พิมพ์ไว้ภายในตัวไฟล์เอกสาร

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ผู้ปฏิบัติการ SOC เฝ้าระวังการแจ้งเตือนจาก DLP การบีบอัดไฟล์เก็บถาวรผิดปกติ (`7z`/`tar`) และการถ่ายโอนข้อมูลปริมาณมากผ่าน Cloud APIs หรือช่องทาง DNS สำหรับพฤติกรรม **Exfiltration Over Web Service** ([MITRE ATT&CK T1567](https://attack.mitre.org/techniques/T1567/)) หรือ **Alternative Protocol** ([T1048](https://attack.mitre.org/techniques/T1048/)):

```kql
CloudAppEvents
| where ActionType in ("FileDownloaded", "FileUploaded", "AnonymousLinkCreated")
| summarize ActionCount = count(), TotalBytes = sum(tolong(RawEventData.FileSizeBytes)) by AccountDisplayName, IPAddress
| where ActionCount >= 25 or TotalBytes >= 500000000
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Data loss prevention software](https://en.wikipedia.org/wiki/Data_loss_prevention_software)
- [Wikipedia — General Data Protection Regulation](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)
- [NIST SP 800-53 Rev. 5 — System and Communications Protection (SC)](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Exfiltration Over Web Service (T1567)](https://attack.mitre.org/techniques/T1567/)
- [MITRE ATT&CK — Exfiltration Over Alternative Protocol (T1048)](https://attack.mitre.org/techniques/T1048/)
