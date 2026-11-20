# 4.8 อธิบายกิจกรรมการตอบสนองต่อเหตุการณ์ที่เหมาะสม (Explain appropriate incident response activities)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การบริหารจัดการเหตุการณ์ความปลอดภัยอย่างเป็นระบบ ([Computer Security Incident Management](https://en.wikipedia.org/wiki/Computer_security_incident_management)) ช่วยลดต้นทุนความเสียหายจากการรั่วไหลและประสานงานอย่างเป็นระเบียบตามวงจร [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final) & [Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final) ร่วมกับคู่มือปฏิบัติของสถาบัน อาทิ [CMU ISO Incident Response](https://www.cmu.edu/iso/governance/procedures/incident-response.html):

1. **Preparation (การเตรียมความพร้อม):** จัดตั้งบทบาททีม CSIRT ผังการสื่อสาร ช่องทางประสานฝ่ายกฎหมาย/ประชาสัมพันธ์ ชุดอุปกรณ์ Jump-Kits การเก็บ Log ส่วนกลาง และฝึกซ้อม **Tabletop Exercises / การจำลองการโจมตี** ก่อนเกิดเหตุจริง
2. **Detection & Analysis (การตรวจจับและวิเคราะห์):** เฝ้าระวังสัญญาณเตือน เชื่อมโยงไทม์ไลน์ข้ามโฮสต์และเครือข่าย คัดกรอง False Positives ประเมินขอบเขต และประกาศระดับความรุนแรง
3. **Containment, Eradication & Recovery (การจำกัดวง กำจัด และฟื้นฟู):**
   - **Containment:** แยกเครือข่าย/โฮสต์ในระยะสั้นหรือเพิกถอน Credential เพื่อหยุดการแพร่กระจายโดยไม่ทำลายหลักฐานที่สูญหายง่ายทันที
   - **Eradication:** ถอดถอนร่องรอยผู้โจมตี — ยุติโพรเซสอันตราย ลบ Scheduled Persistence ปิดช่องทางเจาะระบบตั้งต้น และเปลี่ยนรหัสลับทั้งหมดที่อาจรั่วไหล
   - **Recovery:** กู้คืนจากสำรองข้อมูลที่สะอาด สร้างเครื่องที่ถูกยึดครองใหม่จาก Golden Images เพิ่มความเข้มข้นในการเฝ้าระวัง Firewall/EDR และยืนยันการทำงานปกติ
4. **Post-Incident Activity (ถอดบทเรียนหลังเกิดเหตุ):** วิเคราะห์หาสาเหตุรากเหง้า (`RCA`) โดยไม่กล่าวโทษบุคคล ปรับปรุงกฎตรวจจับ SIEM และเอกสาร SOP พร้อมแจ้งเหตุละเมิดข้อมูลตามกรอบเวลากฎหมาย

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ระหว่างการคัดแยกและรายงานเหตุการณ์ ([Stanford UIT Report Incident](https://uit.stanford.edu/security/report-incident)) นักวิเคราะห์ต้องแยกแยะการบุกรุกจริง (`True Positive` ที่ต้องทำ Containment ทันที) ออกจากการทดสอบเจาะระบบหรือบำรุงรักษาตาม Change Ticket (`False Positive` ที่ไม่ควรรบกวนระบบงาน):

```kql
union DeviceProcessEvents, SigninLogs
| where TimeGenerated between (datetime(2026-11-01T00:00:00Z) .. datetime(2026-11-01T06:00:00Z))
| summarize EventCount = count(), Earliest = min(TimeGenerated), Latest = max(TimeGenerated) by Type
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [NIST SP 800-61 Rev. 2 — Computer Security Incident Handling Guide](https://csrc.nist.gov/pubs/sp/800/61/r2/final)
- [NIST SP 800-61 Rev. 3 — Incident Response Recommendations](https://csrc.nist.gov/pubs/sp/800/61/r3/final)
- [Wikipedia — Computer security incident management](https://en.wikipedia.org/wiki/Computer_security_incident_management)
- [Carnegie Mellon University ISO — Computer Security Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)
- [Stanford University UIT — Incident Reporting & Response](https://uit.stanford.edu/security/report-incident)
