# 3.4 อธิบายความสำคัญของความยืดหยุ่นและการกู้คืนในสถาปัตยกรรมความปลอดภัย (Explain the importance of resilience and recovery in security architecture)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

ความยืดหยุ่น (Resilience) ช่วยรักษาความพร้อมใช้งานในการดำเนินธุรกิจเมื่อเผชิญความล้มเหลวของฮาร์ดแวร์ การโจมตีทางไซเบอร์ (โดยเฉพาะ Ransomware) และภัยธรรมชาติ ([Business Continuity](https://en.wikipedia.org/wiki/Business_continuity)):

- **ความพร้อมใช้งานสูง (`HA`) และระบบสำรอง (Redundancy):** ขจัดจุดล้มเหลวเพียงจุดเดียว (`SPOF`) ด้วยการจัดทำ Server Clustering แบบ Active-Active/Active-Passive, **Load Balancing**, การเดินสายเครือข่ายหลายเส้นทาง, ระบบไฟฟ้า/ความเย็นสำรองแบบ **N+1** และ Availability Zones ที่กระจายตามภูมิศาสตร์
- **กลยุทธ์การจัดเก็บและสำรองข้อมูล (Storage & Backup Strategies):**
  - **Full Backup:** สำรองข้อมูลครบทั้งหมด — ใช้เวลาบันทึกนานและพื้นที่มากที่สุด แต่กู้คืนจบได้ในขั้นตอนเดียวเร็วที่สุด
  - **Incremental Backup:** สำรองเฉพาะข้อมูลที่เปลี่ยนนับจาก *Backup รอบล่าสุด* (Full หรือ Incremental) — เขียนข้อมูลเร็วที่สุด แต่เวลา Restore ต้องเรียงย้อนต่อกันหลายชุด
  - **Differential Backup:** สำรองข้อมูลที่เปลี่ยนทั้งหมดนับจาก *Full Backup ล่าสุด* — สมดุลทั้งความเร็วและการกู้คืน (`Full + Differential ล่าสุด`)
  - **Snapshots, Replication & Immutable Off-site Backups (กฎ 3-2-1):** การล็อกอ็อบเจกต์แบบเขียนครั้งเดียวอ่านได้อย่างเดียว (**WORM / immutable**) ป้องกันไม่ให้ Ransomware ลบไฟล์สำรองข้อมูลได้ ([CISA #StopRansomware](https://www.cisa.gov/stopransomware))
- **ประเภทของศูนย์กู้คืนสำรอง (Recovery Site Types):**
  - **Hot Site:** ศูนย์สำรองที่ติดตั้งอุปกรณ์ครบและจำลองข้อมูลแบบเรียลไทม์ (Downtime ต่ำที่สุด ต้นทุนสูงสุด)
  - **Warm Site:** มีโครงสร้างพื้นฐานและเครือข่ายเตรียมพร้อม แต่ต้องกู้คืนข้อมูลสำรองล่าสุดก่อนสลับระบบมาใช้งาน
  - **Cold Site:** มีเพียงพื้นที่ อาคาร ไฟฟ้า และระบบปรับอากาศ รอการจัดส่งและติดตั้งฮาร์ดแวร์ทั้งหมด (กู้คืนช้าที่สุด ต้นทุนต่ำสุด)
- **ความยืดหยุ่นด้านพลังงานและตัวชี้วัดสำคัญ:** แบตเตอรี่ **UPS** รับช่วงไฟตกสั้นๆ และเชื่อมต่อการทำงานเข้ากับ **เครื่องกำเนิดไฟฟ้าดีเซล** ส่วนเป้าหมายการกู้คืนจะพิจารณาสมดุลระหว่าง **RTO** (ระยะเวลาหยุดชะงักสูงสุดที่ยอมรับได้) และ **RPO** (ช่วงอายุข้อมูลที่สูญหายได้สูงสุด) ควบคู่กับ **MTBF** และ **MTTR**

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ผู้โจมตี Ransomware มักเจาะจงทำลายระบบกู้คืนก่อนเริ่มเข้ารหัสไฟล์ผู้ใช้ เช่น ปิด Volume Shadow Copies (`vssadmin`), หยุดบริการสำรองข้อมูล หรือลบ Cloud Snapshots ([MITRE ATT&CK T1486](https://attack.mitre.org/techniques/T1486/)) การตรวจยืนยันความสมบูรณ์ของ Immutable Backup ก่อนเริ่มขั้นตอน Eradication จึงเป็นสิ่งสำคัญเพื่อให้มั่นใจว่ากู้คืนระบบที่สะอาดกลับมาได้จริง:

```kql
SecurityEvent
| where EventID in (7036, 7045)
| where EventData has_any ("vss", "wbengine", "Veeam", "BackupExec")
| project TimeGenerated, Computer, EventID, EventData
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Business continuity](https://en.wikipedia.org/wiki/Business_continuity)
- [Wikipedia — Business impact analysis](https://en.wikipedia.org/wiki/Business_impact_analysis)
- [CISA — #StopRansomware Recovery & Immutable Backups](https://www.cisa.gov/stopransomware)
- [NIST SP 800-53 Rev. 5 — Contingency Planning (CP)](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Data Encrypted for Impact (T1486)](https://attack.mitre.org/techniques/T1486/)
