# 1.3 อธิบายแนวปฏิบัติที่ดีด้านการบริหารจัดการการเปลี่ยนแปลง (Explain change management best practices)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

ตามที่ระบุไว้ใน [Change Management](https://en.wikipedia.org/wiki/Change_management_(engineering)) และ [Change Control](https://en.wikipedia.org/wiki/Change_control) การบริหารจัดการการเปลี่ยนแปลงอย่างเป็นระบบช่วยป้องกันความผิดพลาดจากการตั้งค่า (Misconfigurations), ระบบล่มโดยไม่ตั้งใจ และการลดทอนความปลอดภัยโดยไม่ได้รับอนุมัติ ทุกครั้งที่มีการปรับปรุงฮาร์ดแวร์ ซอฟต์แวร์ กฎ Firewall หรือนโยบาย Cloud IAM:

- **กระบวนการทางธุรกิจและสายอนุมัติ (Business Processes & Approval Flow):** การเปลี่ยนแปลงสำคัญทุกครั้งต้องเริ่มจากการยื่น **Change Request (CR)** ที่ระบุขอบเขต เหตุผล การพึ่งพาระบบ และ **Impact Analysis** (ประเมินขอบเขตผลกระทบและโซนความปลอดภัยที่เกี่ยวข้อง) จากนั้นคณะกรรมการ **Change Advisory Board (CAB)** หรือผู้ทบทวนร่วมจะเป็นผู้อนุมัติก่อนดำเนินการ ส่วนแพตช์ฉุกเฉินจะเดินตามช่องทาง Emergency CAB แบบเร่งด่วนพร้อมบังคับทบทวนหลังดำเนินการ
- **มาตรการคุ้มครองเชิงเทคนิค (Technical Safeguards):**
  - **Testing & Validation:** จำลองค่าคอนฟิกแบบ Production ในสภาพแวดล้อม Staging Sandbox เพื่อทดสอบก่อนนำขึ้นใช้งานจริง
  - **Maintenance Windows:** กำหนดการเปลี่ยนแปลง (รวมถึงการรีสตาร์ทบริการหรือการปรับโครงสร้างฐานข้อมูลที่จำเป็น) ให้อยู่ในช่วงเวลาผลกระทบต่ำที่ได้รับอนุมัติล่วงหน้าพร้อมแจ้งผู้เกี่ยวข้อง
  - **Allowlists & Restricted Lists:** บังคับใช้ Deployment Freezes ในช่วงธุรกิจหนาแน่น และควบคุมแพ็กเกจหรือไบนารีที่ได้รับอนุญาตให้ติดตั้ง
  - **Backout / Rollback Plan:** ทุกใบ CR ต้องมีขั้นตอนที่เป็นรูปธรรมและมี Snapshot/Backup ล่าสุดที่ยืนยันแล้วว่าสมบูรณ์สำหรับย้อนกลับระบบทันทีหากการตรวจสอบหลังอัปเดตไม่ผ่าน
  - **Standard Operating Procedures (SOPs) & Version Control:** ปรับปรุงแผนภาพสถาปัตยกรรม, Runbooks และโค้ด Infrastructure-as-Code (`IaC`) ที่ควบคุมเวอร์ชันให้เป็นปัจจุบันทันทีหลังติดตั้งเพื่อป้องกัน Configuration Drift
- **ระบบเก่าและข้อยกเว้น (Legacy Systems & Exceptions):** เมื่อแอปพลิเคชันรุ่นเก่าไม่สามารถอัปเกรดหรือรีสตาร์ทตามรอบปกติได้ เจ้าของความเสี่ยงต้องจัดทำเอกสาร **Exception** อย่างเป็นทางการและติดตั้ง Compensating Controls (เช่น การแบ่งโซนเครือข่ายหรือกฎ WAF)

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

บันทึกจากระบบ Change Management (เลขใบงาน `CHG-xxxx`) เป็นหนึ่งใน **Benign Context Logs** ที่มีน้ำหนักความน่าเชื่อถือสูงที่สุดระหว่างการคัดแยกสัญญาณเตือนใน SOC เมื่อกฎ SIEM แจ้งเตือนการแก้ไขนโยบาย Firewall, การรีสตาร์ทบริการตามกำหนด หรือการแก้ Group Policy Object (GPO) นักวิเคราะห์จะตรวจสอบว่าบัญชีผู้ใช้ เครื่องเป้าหมาย และเวลา UTC ตรงกับใบ CR ที่อนุมัติในช่วง Maintenance Window หรือไม่ เพื่อช่วยป้องกันการส่งต่อ False Positive พร้อมจับความเคลื่อนไหวแบบ **Impair Defenses** ที่ไม่ได้รับอนุญาต ([MITRE ATT&CK T1562](https://attack.mitre.org/techniques/T1562/)):

```kql
AuditLogs
| where OperationName has_any ("Update policy", "Modify firewall", "Install extension")
| project TimeGenerated, InitiatedBy, TargetResources, Result
| sort by TimeGenerated desc
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Change management (engineering)](https://en.wikipedia.org/wiki/Change_management_(engineering))
- [Wikipedia — Change control](https://en.wikipedia.org/wiki/Change_control)
- [NIST SP 800-53 Rev. 5 — Configuration Management (CM)](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Impair Defenses (T1562)](https://attack.mitre.org/techniques/T1562/)
- [Microsoft Learn — KQL Overview](https://learn.microsoft.com/en-us/kusto/query/)
