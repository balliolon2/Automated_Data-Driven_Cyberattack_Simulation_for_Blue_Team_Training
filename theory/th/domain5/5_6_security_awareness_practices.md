# 5.6 ปฏิบัติตามแนวทางสร้างความตระหนักรู้ด้านความปลอดภัย (Implement security awareness practices)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การตัดสินใจของมนุษย์ยังคงเป็นทั้งเป้าหมายหลักของการโจมตีและเป็นเซ็นเซอร์แจ้งเตือนล่วงหน้าอันทรงพลัง ([Security Awareness](https://en.wikipedia.org/wiki/Security_awareness), [NIST SP 800-50 Rev. 1](https://csrc.nist.gov/pubs/sp/800/50/r1/final), [CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)):

- **วงจรโครงการและความถี่ในการกำกับดูแล:**
  - **การอบรมเริ่มต้นและการทบทวนตามรอบ:** การอบรมภาคบังคับเมื่อเข้าทำงาน (ครอบคลุมการลงนามรับทราบ AUP, การจัดการข้อมูล, ช่องทางแจ้งเหตุ) เสริมด้วยโมดูลทบทวนประจำปีและประกาศเตือนภัยทันเหตุการณ์
  - **การอบรมตามบทบาทหน้าที่ (Role-Based Training):** ออกแบบเนื้อหาเฉพาะกลุ่มเสี่ยงสูง เช่น **นักพัฒนาซอฟต์แวร์** (การเขียนโค้ดปลอดภัยตาม OWASP, สุขอนามัย Secrets), **ผู้บริหาร** (การทวนสอบ Whaling/BEC) และ **ผู้ดูแลระบบ IT สิทธิ์สูง** (Change Control, ขั้นตอน PAM)
- **การสร้างความตระหนักรู้เฉพาะภัยคุกคาม:**
  - **Phishing Campaigns & Simulations:** การส่งอีเมลจำลองฟิชชิ่งเพื่อวัดอัตราการคลิกเทียบกับ **อัตราการรายงานแจ้งเตือน** ([MITRE ATT&CK T1566](https://attack.mitre.org/techniques/T1566/)) สำหรับกระตุ้นการเรียนรู้ย่อยเชิงสร้างสรรค์แทนการลงโทษด้วยความกลัว
  - **วิศวกรรมสังคม สุขอนามัยรหัสผ่าน และภัยคนใน:** การสังเกตข้ออ้างเร่งด่วน การหลีกเลี่ยงการใช้รหัสผ่านซ้ำด้วย Password Manager องค์กร การปิดใช้สื่อถอดเสียบ USB/สายเคเบิลที่เสี่ยง และการสังเกตพฤติกรรมปฏิบัติงานที่ผิดปกติ
- **ตัวชี้วัดและวัฒนธรรมองค์กร:** วัดความรวดเร็วในการกดปุ่มรายงานฟิชชิ่ง การลดลงของอัตราการคลิกซ้ำ และการสร้างวัฒนธรรม **รายงานก่อนโดยไม่กล่าวโทษ (Report-First Culture)** ([Stanford UIT Report Incident](https://uit.stanford.edu/security/report-incident))

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

การส่งต่ออีเมลฟิชชิ่งจากผู้ใช้งานอย่างรวดเร็วช่วยลดค่า **Mean Time to Detect (`MTTD`)** ของทีม SOC ได้โดยตรง ทำให้พนักงานที่ผ่านการอบรมกลายเป็นเซ็นเซอร์บุคคลากรที่กระจายอยู่ทั่วองค์กรเพื่อป้อนข้อมูลเข้าสู่ SOAR Playbook กวาดลบอีเมลก่อนพนักงานคนอื่นหลงคลิกลิงก์อันตราย:

```kql
EmailEvents
| where LatestDeliveryAction == "Delivered" and ReportId != ""
| summarize ReportedCount = count() by SenderFromAddress, Subject
| order by ReportedCount desc
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Security awareness](https://en.wikipedia.org/wiki/Security_awareness)
- [NIST SP 800-50 Rev. 1 — Building a Cybersecurity and Privacy Learning Program](https://csrc.nist.gov/pubs/sp/800/50/r1/final)
- [CISA — Phishing Threat & Response Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)
- [Stanford University UIT — Incident Reporting & Response](https://uit.stanford.edu/security/report-incident)
- [MITRE ATT&CK — Phishing (T1566)](https://attack.mitre.org/techniques/T1566/)
