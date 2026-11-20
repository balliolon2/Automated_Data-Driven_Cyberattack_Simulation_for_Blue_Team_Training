# 2.5 อธิบายเทคนิคการลดความเสี่ยงที่ใช้ในการป้องกันระบบองค์กร (Explain mitigation techniques used to secure the enterprise)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

มาตรการลดความเสี่ยงระดับองค์กรช่วยลดทั้งโอกาสในการถูกโจมตีสำเร็จและขอบเขตความเสียหาย (Blast Radius) ทั้งในระดับเครือข่าย โฮสต์ และตัวตนผู้ใช้ ([NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [Network Segmentation](https://en.wikipedia.org/wiki/Network_segmentation)):

- **การควบคุมการเข้าถึงและการแบ่งส่วนเครือข่าย (Access & Segmentation Mitigations):**
  - **[Network Segmentation](https://en.wikipedia.org/wiki/Network_segmentation) & Access Control Lists (ACLs):** แบ่งเครือข่ายออกเป็นโซนความเชื่อถือ (VLANs, Screened Subnets/DMZs) เพื่อไม่ให้การบุกรุกเว็บเซิร์ฟเวอร์สาธารณะข้ามเข้าถึงฐานข้อมูลภายในได้โดยเสรี ([MITRE ATT&CK T1021 — Remote Services](https://attack.mitre.org/techniques/T1021/))
  - **Least Privilege:** จำกัดสิทธิ์บัญชีผู้ใช้ Service Tokens และกฎ Firewall ไว้เท่าที่จำเป็นต่อการทำงานเท่านั้น
- **การควบคุมการรันโปรแกรมและปลายทาง (Execution & Endpoint Mitigations):**
  - **Application Allowlisting:** อนุญาตให้ทำงานได้เฉพาะไฟล์ที่เซ็นดิจิทัลหรือได้รับการอนุมัติเท่านั้น ช่วยสกัดกั้นเพย์โหลดของ Ransomware/Trojan ที่ยังไม่เคยรู้จักได้แม้ไม่มีลายเซ็น
  - **[Endpoint Detection and Response (EDR)](https://en.wikipedia.org/wiki/Endpoint_detection_and_response) & Host Firewalls:** ป้องกันตามพฤติกรรม สแกนหน่วยความจำ และสั่งแยกเครื่องที่ถูกบุกรุกออกจากเครือข่าย
- **การเสริมความแข็งแกร่ง การตั้งค่า และแพตช์ (Configuration, Hardening & Patching):**
  - บังคับใช้ค่าตั้งต้นความปลอดภัยมาตรฐาน **ปิดพอร์ตและบริการที่ไม่จำเป็น** เปลี่ยนรหัสผ่านโรงงาน และปิดโปรโตคอลเก่าที่ไม่ปลอดภัย (SMBv1, Telnet, SSLv3)
  - **Patch Management & Encryption:** กำหนดรอบเวลาการแพตช์ตามระดับความรุนแรง ([NIST SP 800-40 Rev. 4](https://csrc.nist.gov/pubs/sp/800/40/r4/final)) ควบคู่กับการเข้ารหัสวอลุ่มข้อมูลสำคัญ
- **การกักกันและการจัดการวงจรชีวิตระบบ:** การแยกเข้า Quarantine VLAN, การเปิดไฟล์แนบที่ไม่น่าเชื่อถือใน **Sandbox**, การแพตช์หรือแยกวง และ **การปลดระวาง (Decommissioning)** ฮาร์ดแวร์ที่หมดอายุสนับสนุน (EOL)

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ระหว่างการรับมือเหตุการณ์จริง ทีม Blue Team ต้องเลือกมาตรการที่ **ได้สัดส่วนกับหลักฐาน** เช่น การแยกเครื่องที่ติดเชื้อผ่าน EDR (`Corrective` containment) การเพิกถอน Session Token ที่ถูกขโมย (`Access revocation`) การบล็อกโดเมน C2 ผ่าน DNS Sinkholing หรือการเพิ่มกฎ Firewall เฉพาะจุด แทนการสั่งหยุดระบบการผลิตทั้งหมดโดยไม่จำเป็น:

```kql
DeviceEvents
| where ActionType in ("AsrExecutableBlockingAudited", "AsrLsassCredentialTheftBlocked", "AntivirusDetection")
| summarize Detections = count() by DeviceName, ActionType, FileName
| order by Detections desc
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Network segmentation](https://en.wikipedia.org/wiki/Network_segmentation)
- [Wikipedia — Endpoint detection and response](https://en.wikipedia.org/wiki/Endpoint_detection_and_response)
- [NIST SP 800-40 Rev. 4 — Enterprise Patch Management Planning](https://csrc.nist.gov/pubs/sp/800/40/r4/final)
- [NIST SP 800-53 Rev. 5 — Security Controls](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Remote Services (T1021)](https://attack.mitre.org/techniques/T1021/)
