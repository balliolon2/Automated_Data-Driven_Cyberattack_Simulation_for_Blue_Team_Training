# Domain 5: การบริหารจัดการและกำกับดูแลความปลอดภัย (สรุปกระชับ 3 นาที)

> **Domain 5** | **สัดส่วนในข้อสอบ:** 20% (`6 / 30` ข้อใน Pre/Post-Test) | **เวลาอ่านโดยประมาณ:** ~3 นาที

หากทีม SOC เปรียบเหมือนนักดับเพลิงที่แนวหน้า Domain 5 คือคนที่ออกกฎหมายความปลอดภัย คำนวณงบและประกันความเสี่ยง จัดซ้อมหนีไฟ และดูแลข้อมูลลูกค้าให้ถูกต้องตามกฎหมาย

---

## 1. ทฤษฎีหลัก CompTIA Security+ ที่ออกสอบบ่อย (Core Security+ Theory)

- **ลำดับชั้นเอกสารกำกับดูแล (Governance Hierarchy)** ([NIST CSF 2.0](https://www.nist.gov/cyberframework)):
  - *Policies* (กฎใหญ่ระดับองค์กร: ปกป้อง *อะไร*) → *Standards* (เกณฑ์เทคนิคเจาะจง เช่น AES-256) → *Procedures / Playbooks* (คู่มือลงมือทำทีละขั้น) → *Guidelines* (ข้อแนะนำที่ยืดหยุ่นได้)
  - *บทบาทข้อมูล:* **Data Owner** (ผู้บริหารเจ้าของข้อมูล) vs. **Data Custodian** (แอดมินผู้ดูแลระบบและสำรองข้อมูล)
- **การบริหารความเสี่ยง & ความต่อเนื่องทางธุรกิจ** ([NIST SP 800-30](https://csrc.nist.gov/pubs/sp/800/30/r1/final), [Wikipedia: Risk Management](https://en.wikipedia.org/wiki/Risk_management), [Wikipedia: Business Continuity](https://en.wikipedia.org/wiki/Business_continuity)):
  - **4 ทางเลือกรับมือความเสี่ยง:** *Mitigate* (ลดด้วยมาตรการป้องกัน), *Transfer* (โอนด้วยประกันไซเบอร์), *Accept* (ยอมรับอย่างเป็นทางการ), *Avoid* (เลิกทำกิจกรรมที่เสี่ยง)
  - **สูตรคำนวณความเสี่ยง:** `SLE = AV × EF` (เสียหายต่อครั้ง) | `ALE = SLE × ARO` (ความเสียหายคาดการณ์ต่อปี)
  - **ตัววัดการกู้คืน:** *RTO* = เวลาระบบล่มสูงสุดที่รับได้ และ *RPO* = ช่วงเวลาข้อมูลสูญหายสูงสุดที่รับได้ (ตัวกำหนดความถี่ในการ Backup)
- **กฎหมายคุ้มครองข้อมูลส่วนบุคคล & ความเสี่ยงคู่ค้า (Vendor Risk)** ([Wikipedia: GDPR](https://en.wikipedia.org/wiki/General_Data_Protection_Regulation)):
  - กฎหมาย GDPR / PDPA บังคับให้ปกป้องข้อมูลส่วนบุคคล (PII) และแจ้งเหตุข้อมูลรั่วภายในเวลาที่กำหนด (เช่น 72 ชั่วโมง) ส่วนมาตรฐาน **PCI-DSS** คุ้มครองข้อมูลบัตรเครดิต และ **ISO 27001 / SOC 2** ตรวจสอบผ่านผู้ตรวจประเมินอิสระ
  - ควบคุมคู่ค้าภายนอกด้วย NDA, SLA, สิทธิ์เข้าตรวจ (Right-to-Audit) และการจำกัดเวลาเข้าถึงระบบ

---

## 2. สัญญาณอันตรายใน SIEM Log, KQL และ MITRE ATT&CK

จับการลบ Log ปกปิดร่องรอย ([T1070](https://attack.mitre.org/techniques/T1070/)), การอัปโหลด PII ขึ้นคลาวด์ที่ไม่ได้รับอนุญาต ([T1567](https://attack.mitre.org/techniques/T1567/)) และบัญชีคู่ค้าที่เกินสิทธิ์ ([T1199](https://attack.mitre.org/techniques/T1199/)):

```kql
source == "dlp" or event_id == "1102" or user contains "vendor"
```

- **จุดสังเกต:** `EventID 1102` (ล้าง Security Audit Log!), การส่งไฟล์ PII จำนวนมากไปคลาวด์ส่วนตัว หรือ Vendor เข้าดูฐานข้อมูลนอกช่วงเวลาใบงาน

---

## 3. การคัดแยก True Positive vs. False Positive และ Playbook 3 ระยะ

| True Positive (ภัยคุกคามจริง) | False Positive (งานปกติ / มี Benign Context Log ยืนยัน) |
|---|---|
| บัญชี Vendor ดึงตาราง PII นอกเวลาซ่อมบำรุง หรือมีการลบ Log (`1102`) | เซสชันของ Vendor ที่มี **ใบงาน Maintenance/Support Ticket** และช่วง SLA ตรงกัน |
| ส่งไฟล์ PII ขึ้นเว็บฝากไฟล์ส่วนตัวโดยไม่ได้รับอนุญาต (Shadow IT) | การส่งออกหลักฐานตรวจสอบที่เข้ารหัสโดย Data Custodian เพื่อ Audit ตามกำหนด |

- **Containment:** ระงับ VPN/บัญชี Vendor ทันที บล็อกเว็บฝากไฟล์ที่ไม่ได้รับอนุญาตผ่าน DLP/Proxy และสำรอง Log ที่เหลือไว้
- **Eradication:** เปลี่ยนรหัสผ่าน/Secret ที่แชร์กับคู่ค้า ตรวจสอบประวัติการเข้าถึงตารางข้อมูล และใช้สิทธิ์แบบชั่วคราว (Just-in-Time)
- **Recovery:** เปิดสิทธิ์เฉพาะที่ตรวจสอบแล้ว ประเมินขอบเขตข้อมูล PII ที่รั่วเพื่อแจ้งหน่วยงานกฎหมาย และอัปเดต SOP

---

## 4. สรุปสั้นก่อนลงสนามและแหล่งอ้างอิง (Key Takeaways & References)

- **RTO** คือเวลาหยุดทำงานสูงสุดที่ยอมรับได้ **RPO** คือปริมาณข้อมูลสูญหายที่ยอมรับได้ และ `EventID 1102` (ลบ Audit Log) คือสัญญาณอันตรายร้ายแรงเสมอ!
- คู่มือ Query: [Microsoft Learn: KQL](https://learn.microsoft.com/en-us/kusto/query/)
