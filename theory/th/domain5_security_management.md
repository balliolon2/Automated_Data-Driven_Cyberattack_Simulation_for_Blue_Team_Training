# Domain 5: การบริหารจัดการและกำกับดูแลความปลอดภัย (Security Program Management and Oversight)

> **Domain ID:** `domain5` | **สัดส่วนในข้อสอบ:** 20% (6 จาก 30 ข้อใน Pre-Test / Post-Test) | **แหล่งอ้างอิง:** Wikipedia, NIST CSRC, NIST CSF 2.0, CMU ISO, Stanford UIT, MITRE ATT&CK

บทเรียนนี้ครอบคลุมการกำกับดูแลความมั่นคงปลอดภัยไซเบอร์ การบริหารความเสี่ยง การปฏิบัติตามข้อกำกับ การกำกับดูแลผู้ให้บริการภายนอก และความรับผิดชอบหลังเกิดเหตุ ตามมาตรฐาน [NIST CSF 2.0](https://www.nist.gov/cyberframework), [NIST SP 800-30 Rev. 1](https://csrc.nist.gov/pubs/sp/800/30/r1/final), [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [CMU ISO Governance](https://www.cmu.edu/iso/governance/procedures/incident-response.html) และบทความด้านธรรมาภิบาลบน Wikipedia

---

## 1. ทฤษฎีหลัก CompTIA Security+ และแนวคิดการสอบ (Core Security+ Theory)

### 1.1 โครงสร้างการกำกับดูแลและลำดับชั้นเอกสาร
อ้างอิงตามฟังก์ชัน **Govern** ใน [NIST Cybersecurity Framework (CSF) 2.0](https://www.nist.gov/cyberframework) และ [Stanford University Information Security](https://uit.stanford.edu/security):
- **Policies (นโยบาย):** คำแถลงเชิงบังคับระดับบริหารที่กำหนด *สิ่งที่ต้องปกป้อง* (เช่น Acceptable Use Policy, Incident Response Policy, Data Classification Policy)
- **Standards (มาตรฐาน):** เกณฑ์ทางเทคนิคเชิงบังคับที่ระบุข้อกำหนดชัดเจน (เช่น การเข้ารหัส AES-256, การบังคับใช้ MFA)
- **Procedures (SOPs / Playbooks):** ขั้นตอนปฏิบัติงานทีละขั้นที่กำหนด *วิธีการทำงาน* ของทีม SOC ในการรับมือเหตุการณ์และรักษาพยานหลักฐานให้สม่ำเสมอ ([CMU Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html))
- **Guidelines (แนวทางปฏิบัติ):** ข้อแนะนำที่ปรับยืดหยุ่นได้ตามบริบททางวิศวกรรม
- **บทบาทและความรับผิดชอบ:** Data Owner (ผู้บริหารผู้รับผิดชอบ), Data Custodian (ผู้ดูแลทางเทคนิค) และความแตกต่างระหว่าง Data Controller กับ Data Processor ตามกฎหมายคุ้มครองข้อมูลส่วนบุคคล

### 1.2 การบริหารความเสี่ยงและตัวชี้วัดความต่อเนื่อง (Risk Management & Continuity Metrics)
กำหนดไว้ใน [NIST SP 800-30 Rev. 1: Guide for Conducting Risk Assessments](https://csrc.nist.gov/pubs/sp/800/30/r1/final), [Wikipedia: Risk Management](https://en.wikipedia.org/wiki/Risk_management) และ [Wikipedia: Business Continuity](https://en.wikipedia.org/wiki/Business_continuity):
- **ขั้นตอนประเมินความเสี่ยง:** ระบุ Asset → ระบุภัยคุกคามและช่องโหว่ → ประเมินโอกาสเกิดและผลกระทบ → เลือกแนวทางจัดการความเสี่ยง
- **ทางเลือกการจัดการความเสี่ยง (Risk Treatment Options):**
  1. *Mitigate (ลดความเสี่ยง):* ติดตั้งมาตรการทางเทคนิคหรือการบริหารเพื่อลดโอกาสเกิดหรือผลกระทบ
  2. *Transfer (ถ่ายโอนความเสี่ยง):* ทำประกันภัยไซเบอร์หรือส่งมอบหน้าที่ปฏิบัติการบางส่วนให้ MSSP ตามสัญญา
  3. *Accept (ยอมรับความเสี่ยง):* บันทึกการลงนามรับรองอย่างเป็นทางการจากผู้บริหารเมื่อต้นทุนควบคุมสูงกว่าความเสียหายที่คาดไว้
  4. *Avoid (หลีกเลี่ยงความเสี่ยง):* ยุติขั้นตอนการทำงานที่เสี่ยงหรือปลดระวางระบบเก่าที่มีช่องโหว่ทั้งหมด
- **สูตรคำนวณความเสี่ยงเชิงปริมาณ (Quantitative Risk Formulas):**
  - `SLE` (Single Loss Expectancy) = `AV` (Asset Value) × `EF` (Exposure Factor)
  - `ALE` (Annualized Loss Expectancy) = `SLE` × `ARO` (Annualized Rate of Occurrence)
- **ตัวชี้วัดความต่อเนื่องทางธุรกิจและความยืดหยุ่น:**
  - `RTO` (Recovery Time Objective): ระยะเวลาที่ระบบหยุดชะงักได้สูงสุดก่อนกระทบธุรกิจรุนแรง
  - `RPO` (Recovery Point Objective): ช่วงอายุข้อมูลที่ยอมให้สูญหายได้สูงสุด (กำหนดความถี่ในการทำ Backup และ Replication)
  - `MTTR` / `MTBF`: เวลาเฉลี่ยในการซ่อมแซมและเวลาเฉลี่ยระหว่างความขัดข้อง

### 1.3 ความเป็นส่วนตัว ข้อกำกับ และความเสี่ยงบุคคลที่สาม
อธิบายไว้ใน [Wikipedia: General Data Protection Regulation (GDPR)](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation) และ [Wikipedia: Data Loss Prevention](https://en.wikipedia.org/wiki/Data_loss_prevention_software):
- **ภาระผูกพันตามกฎหมายและความเป็นส่วนตัว:** ภายใต้ [GDPR](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation) และกฎหมายคุ้มครองข้อมูลส่วนบุคคลที่เทียบเคียงกัน (เช่น PDPA) องค์กรต้องปกป้องข้อมูลส่วนบุคคล/PII บันทึกฐานกฎหมายการประมวลผล และแจ้งหน่วยงานกำกับดูแลภายในกำหนดเวลา (เช่น 72 ชั่วโมง) หลังยืนยันเหตุข้อมูลส่วนบุคคลรั่วไหล
- **มาตรฐานความปลอดภัย:** PCI-DSS (การปกป้องข้อมูลผู้ถือบัตร) และ ISO/IEC 27001 (การรับรองระบบ ISMS)
- **ความเสี่ยงจากบุคคลที่สามและห่วงโซ่อุปทาน:** การตรวจสอบสถานะความปลอดภัยของผู้จำหน่าย (Vendor Due Diligence) เงื่อนไขสิทธิ์เข้าตรวจสอบ (Right-to-Audit) สัญญา NDA, SLA และการตรวจติดตามเซสชัน VPN/SaaS ของคู่ค้า

### 1.4 การตรวจประเมิน การซ้อมแผน และธรรมาภิบาลหลังเกิดเหตุ
- **การตรวจประเมินอิสระและใบรับรอง:** การตรวจสอบภายในและภายนอก (เช่น SOC 2 Type I/II) ยืนยันการทำงานของมาตรการในส่วน Logging และการทบทวนสิทธิ์ ([NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final) หมวด CA/AU)
- **Tabletop Exercises และการรายงานผล:** การซ้อมแผนรับมือเหตุการณ์อย่างสม่ำเสมอ ([Stanford University IT](https://uit.stanford.edu/security)) และการปิดช่องว่างด้านนโยบายระหว่างทบทวนหลังเกิดเหตุ ([NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final))

---

## 2. ข้อมูล SOC Telemetry, Log Indicators และการเชื่อมโยง MITRE ATT&CK

สถานการณ์ใน Domain 5 มักเกี่ยวข้องกับการละเมิดนโยบาย การเข้าถึงข้อมูล PII/ข้อมูลควบคุมโดยไม่ได้รับอนุญาต Shadow IT การใช้บัญชีคู่ค้าโดยมิชอบ หรือการยกระดับเหตุการณ์ตามข้อกำหนด Compliance

### 2.1 แหล่ง Log และฟิลด์สำคัญ
- **DLP (Data Loss Prevention) Logs:** `policy_matched` (`PII`, `PCI`, `Confidential`), `user`, `destination` (`USB`, `personal_cloud`, `external_email`), `action` (`BLOCK`/`AUDIT`), `file_count`
- **Vendor / Third-Party VPN & PAM Logs:** `vendor_account`, `ticket_ref`, `session_duration`, `target_system`, `commands_executed`
- **Audit & Compliance Logs:** การเข้าถึงตารางข้อมูล PII ของฝ่ายบุคคล/ลูกค้าที่ถูกจำกัดสิทธิ์ การแก้ไขนโยบายการจัดเก็บ หรือการล้าง Audit Logs (`EventID 1102`)

### 2.2 รูปแบบการสืบค้น KQL ที่พบบ่อย

ไวยากรณ์สืบค้นอ้างอิงตาม [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/):

```kql
source == "dlp" and severity == "high"
message contains "PII" or message contains "exfiltration"
user contains "vendor" or event_id == "1102"
```

### 2.3 การเชื่อมโยง MITRE ATT&CK

รหัสเทคนิคทั้งหมดเชื่อมโยงโดยตรงไปยัง [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/):

| Technique ID | ชื่อเทคนิค (Technique Name) | สัญญาณ Telemetry ที่สังเกตได้ (Observable Indicator) |
|---|---|---|
| [T1070](https://attack.mitre.org/techniques/T1070/) | Indicator Removal | มีการล้าง Security Audit Log (`EventID 1102`) ตัดทอนประวัติ Bash หรือปิดการส่งข้อมูลไป SIEM |
| [T1567](https://attack.mitre.org/techniques/T1567/) | Exfiltration Over Web Service | การอัปโหลดไฟล์ลูกค้า/PII จำนวนมากไปยังบริการ Cloud นอกเหนือการอนุญาต (Shadow IT) ที่ DLP/Proxy ตรวจจับได้ |
| [T1199](https://attack.mitre.org/techniques/T1199/) | Trusted Relationship | บัญชี VPN ของผู้รับเหมาหรือคู่ค้าภายนอกเข้าถึงฐานข้อมูลการผลิตนอกขอบเขตโดยไม่มี SLA Ticket อนุมัติ |
| [T1530](https://attack.mitre.org/techniques/T1530/) | Data from Cloud Storage | การสำรวจหรือดาวน์โหลดไฟล์จาก Storage Buckets ด้าน Compliance/PII โดยบัญชีบริการที่ไม่ใช่เจ้าของ |

---

## 3. การคัดแยกเหตุการณ์ (TP vs. FP) และแนวทางปฏิบัติ 3 ขั้นตอนตาม Playbook

### 3.1 การจำแนก True Positive (TP) ออกจาก False Positive (FP)
- **ตัวบ่งชี้ True Positive:** การส่งออกข้อมูล PII ที่ไม่ได้เข้ารหัสจำนวนมากไปยังคลาวด์ส่วนตัวที่ไม่อนุญาต; บัญชีคู่ค้าล็อกอินโดยไม่มีใบงานซ่อมบำรุงที่เปิดใช้งาน; หรือมีการจงใจล้าง Security Audit Logs ทันทีหลังคิวรีข้อมูลอ่อนไหว
- **ตัวบ่งชี้ False Positive (Benign Context Logs):**
  - **Approved Vendor Support Ticket:** การเข้าถึงเพื่อซ่อมบำรุงจากบุคคลที่สามซึ่งมีใบอนุญาต/Change Ticket ที่ถูกต้องตรงตามขอบเขตและช่วงเวลา
  - **Authorized Compliance / Audit Export:** การถ่ายโอนหลักฐาน Audit แบบเข้ารหัสโดย Data Custodian ที่ได้รับมอบหมายหรือผู้ตรวจประเมินภายนอกภายใต้คำขอทางกฎหมาย/การตรวจสอบที่เป็นลายลักษณ์อักษร
  - **DLP Test / Awareness Exercise:** กิจกรรมตามกฎ DLP ที่เกิดจากบัญชีทดสอบ Compliance ที่ได้รับอนุญาตหรือแบบฝึกซ้อม Phishing ตามแผน

### 3.2 หลักการตอบสนอง 3 ขั้นตอนตาม Playbook
อ้างอิงตาม [NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final), [CMU Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html) และ [Stanford Incident Reporting](https://uit.stanford.edu/security/report-incident):
- **Phase 1 — Containment (การจำกัดวงความเสียหาย):**
  - *สิ่งที่ควรทำ:* ระงับ Token การเข้าถึงของคู่ค้าหรือพนักงานที่ละเมิดนโยบาย เพิกถอน OAuth Grants ของแอปคลาวด์ที่ไม่อนุญาต ปิดกั้นปลายทางการอัปโหลดภายนอก และรักษา Audit Logs ให้ครบถ้วนสมบูรณ์
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การลบบัญชีพนักงานทิ้งก่อนเก็บหลักฐาน Chain-of-Custody หรือการเปิดเผยรายละเอียดเหตุละเมิดที่ยังไม่ยืนยันต่อสาธารณะ
- **Phase 2 — Eradication (การกำจัดภัยคุกคาม):**
  - *สิ่งที่ควรทำ:* บังคับใช้กฎบล็อกของ DLP ถอดถอนการเชื่อมต่อบุคคลที่สามที่ไม่ได้รับอนุญาต ลดสิทธิ์ RBAC/คู่ค้าที่เกินความจำเป็น (Least Privilege) และหมุนเวียน Shared Secrets ที่รั่วไหล
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การปิดเสียงแจ้งเตือน Compliance อย่างเงียบๆ โดยไม่แก้ไขการละเมิดนโยบายและไม่แจ้งผู้มีส่วนได้ส่วนเสียด้านธรรมาภิบาล
- **Phase 3 — Recovery (การฟื้นฟูระบบ):**
  - *สิ่งที่ควรทำ:* ยกระดับเหตุข้อมูล PII/ข้อมูลควบคุมรั่วไหลที่ยืนยันแล้วไปยังฝ่ายกฎหมาย, Compliance และเจ้าหน้าที่คุ้มครองข้อมูลส่วนบุคคล (DPO) ภายในกรอบเวลาตามกฎหมาย (เช่น กฎ 72 ชั่วโมงของ GDPR/PDPA) จัดทำเอกสาร Chain of Custody และปรับปรุงนโยบายคู่ค้า/DLP ระหว่างถอดบทเรียน
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การเปิดให้คู่ค้าเข้าถึงระยะไกลอีกครั้งก่อนลงนามทบทวน SLA/มาตรการความปลอดภัยและยืนยันระบบเก็บ Audit Log

---

## 4. บทสรุปสำคัญและแหล่งอ้างอิง (Key Takeaways & References)

- **ธรรมาภิบาลเชื่อมต่อไปถึงขั้นตอน Recovery:** แม้การควบคุมขอบเขตทางเทคนิคจะสำเร็จแล้ว ขั้นตอน Playbook ใน Domain 5 มักต้องทำ **Chain of Custody**, **การแจ้ง Compliance/DPO** และ **อัปเดตนโยบายระหว่าง Lessons Learned** ด้วย
- **ตรวจสอบขอบเขตและใบอนุญาตเสมอ:** การแจ้งเตือนการเข้าถึงของคู่ค้าและข้อมูล PII มักตัดสินได้จากว่ามี Ticket ที่ได้รับอนุมัติ เอกสารมอบอำนาจตรวจประเมิน หรือบันทึก SLA ใน Telemetry Logs หรือไม่
- **ปกป้องความสมบูรณ์ของบันทึกตรวจสอบ:** การลบ Logs ([T1070](https://attack.mitre.org/techniques/T1070/)) หรือการดัดแปลงหลักฐานทำลายความสอดคล้องทางกฎหมายและเป็นสัญญาณ True Positive ความสำคัญสูงสุดเสมอ

### แหล่งอ้างอิงเว็บที่น่าเชื่อถือ (Authoritative Web References)
- **Wikipedia (แหล่งความรู้เปิด):**
  - [Information Security](https://en.wikipedia.org/wiki/Information_security) & [Risk Management](https://en.wikipedia.org/wiki/Risk_management)
  - [Business Continuity](https://en.wikipedia.org/wiki/Business_continuity) & [Data Loss Prevention (DLP)](https://en.wikipedia.org/wiki/Data_loss_prevention_software)
  - [General Data Protection Regulation (GDPR)](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)
- **หน่วยงานภาครัฐและมาตรฐาน (NIST):**
  - [NIST Cybersecurity Framework (CSF) 2.0](https://www.nist.gov/cyberframework)
  - [NIST SP 800-30 Rev. 1: Guide for Conducting Risk Assessments](https://csrc.nist.gov/pubs/sp/800/30/r1/final)
  - [NIST SP 800-53 Rev. 5: Security and Privacy Controls](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
  - [NIST SP 800-61 Rev. 3: Incident Response Recommendations](https://csrc.nist.gov/pubs/sp/800/61/r3/final)
- **มหาวิทยาลัยและฐานความรู้เปิด:**
  - [CMU Information Security Office: Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)
  - [Stanford University Information Security (SISA)](https://uit.stanford.edu/security) & [Report an Incident](https://uit.stanford.edu/security/report-incident)
  - [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/) (`T1070`, `T1567`, `T1199`, `T1530`)
  - [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/)
