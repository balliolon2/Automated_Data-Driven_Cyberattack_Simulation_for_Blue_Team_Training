# examination/ — แหล่งข้อสอบต้นทาง (Source Exam Markdown)

## 1. ไดเรกทอรีนี้เก็บอะไร

ไดเรกทอรีนี้เก็บไฟล์ Markdown ต้นทางของคลังข้อสอบ Pre-test/Post-test ซึ่งเป็นเนื้อหาที่ใช้ในระบบประเมินผลของผู้เรียน จำนวน **540 ข้อ** โดยแบ่งเป็น **ไฟล์ต้นทางตาม Domain จำนวน 5 ไฟล์** (`domain*.md`) พร้อมด้วย `README.md` นี้ที่อธิบายแหล่งที่มา การแมป และนโยบายการใช้งาน

ไฟล์ในไดเรกทอรีนี้เป็น **แหล่งข้อมูลต้นทาง (source of truth)** สำหรับการสร้าง Artifact ด้วย Parser ไม่ใช่ไฟล์ที่ Backend อ่านโดยตรง

## 2. แหล่งที่มาและการอ้างอิง

เนื้อหาข้อสอบทั้งหมดในไดเรกทอรีนี้มาจากหนังสือ:

Gorgotsias, N. (2024). *CompTIA Security+ SY0-701 practice tests & PBQs: Exam SY0-701* (3rd ed.). ExamsDigest LLC; LabsDigest LLC.

- รหัสอ้างอิงโครงการ: **[R1]**
- การอ้างอิงในเนื้อหารายงาน: **(Gorgotsias, 2024)**
- ทะเบียนแหล่งอ้างอิง ข้อจำกัดลิขสิทธิ์ และความเสี่ยงที่ยังเปิดอยู่: [`../docs/REFERENCES.md`](../docs/REFERENCES.md)

## 3. การแมปบท → Domain → ไฟล์

| บท | ชื่อบทในหนังสือ | Domain ในโครงการ | ช่วงข้อสอบ | ไฟล์ | จำนวน |
|---|---|---|---|---|---|
| Ch1 | General Security Concepts | domain1 — Security Concepts | Q1–Q110 | [`domain1_security_concepts.md`](domain1_security_concepts.md) | 110 |
| Ch2 | Threats, Vulnerabilities, and Mitigations | domain2 — Threats, Vulnerabilities, and Mitigations | Q111–Q220 | [`domain2_threats_vulnerabilities.md`](domain2_threats_vulnerabilities.md) | 110 |
| Ch3 | Implementation | domain3 — Security Architecture | Q221–Q310 | [`domain3_security_architecture.md`](domain3_security_architecture.md) | 90 |
| Ch4 | Security Operations | domain4 — Security Operations | Q311–Q460 | [`domain4_security_operations.md`](domain4_security_operations.md) | 150 |
| Ch5 | Security Program Management and Oversight | domain5 — Security Management | Q461–Q540 | [`domain5_security_management.md`](domain5_security_management.md) | 80 |
| **รวม** | | | **Q1–Q540** | | **540** |

หมายเหตุ: หนังสือต้นทางใช้ชื่อบทที่ 3 ว่า **"Implementation"** ขณะที่โครงการแมปบทนี้เข้ากับ CompTIA Domain 3 **"Security Architecture"** ความต่างนี้เป็นเรื่องของชื่อเรียก โดยการแมปยึดตามลำดับบทและช่วงข้อสอบ

## 4. การ Normalization — Question 444

ไฟล์ [`domain4_security_operations.md`](domain4_security_operations.md) เดิมมีข้อบกพร่องเชิงรูปแบบที่หัวข้อ **Question 444** คือไม่มีจุดต่อท้ายหมายเลขข้อ ทำให้ Parser ไม่จับหัวข้อและข้ามข้อนี้ไป

- แก้ไขเมื่อ **2026-10-20** โดยเพิ่ม **จุดหลังหมายเลขข้อเพียงอย่างเดียว**
- **เนื้อหาคำถาม ตัวเลือก เฉลย และคำอธิบายไม่ถูกแก้ไข**
- ผลลัพธ์: จำนวนข้อที่ Ingest เพิ่มจาก **539** เป็น **540** ซึ่งเป็นจำนวนที่คาดหวังของคลังข้อสอบ

## 5. นโยบายไฟล์ที่สร้างขึ้น (Generated Artifact Policy)

- รัน Parser จาก **Root ของ Repository**:

  ```bash
  python scripts/parse_exams.py
  ```

- Artifact ที่เป็นทางการคือ **`parsed_questions.json` ที่ Root ของ Repository** เท่านั้น
- Parser อ่านเฉพาะไฟล์ต้นทาง `domain*.md` ในไดเรกทอรีนี้ และไม่ประมวลผลไฟล์อื่น เช่น `README.md`
- `examination/questions.json` เคยเป็นไฟล์ซ้ำที่ไม่ถูกอ้างอิงโดยสคริปต์ใด และ **ถูกลบออกแล้ว** เพื่อให้เหลือ Artifact เดียว
- ห้ามแก้ไข `parsed_questions.json` ด้วยมือ ให้สร้างใหม่จาก Parser เสมอ

## 6. Metadata ใน JSON

ทุกข้อใน `parsed_questions.json` มีฟิลด์กำกับที่มาแบบ Flat ดังนี้

| Field | ชนิด | ความหมาย |
|---|---|---|
| `source_ref` | string | รหัสแหล่งอ้างอิงในทะเบียน ค่าปัจจุบันคือ `"R1"` |
| `source_chapter` | integer | เลขบทของหนังสือต้นทาง (1–5) |
| `source_question_no` | integer | หมายเลขข้อในหนังสือต้นทาง (1–540) |

ค่า `source_chapter` ได้มาจากการแมป Domain → บท ในสคริปต์ Parser และ `source_question_no` อ่านจากหัวข้อคำถามในไฟล์ Markdown ต้นทาง

## 7. ข้อจำกัดด้านลิขสิทธิ์

- ข้อความคำถาม ตัวเลือก เฉลย และคำอธิบาย เป็นลิขสิทธิ์ของ **ผู้เขียนและสำนักพิมพ์**
- ใช้เพื่อ **การศึกษาและการฝึกอบรมภายในโครงการเท่านั้น**
- **ห้ามนำข้อความคำถามไปเผยแพร่หรือกระจายต่อ** นอกขอบเขตการใช้งานภายใน
- ต้องอ้างอิง `[R1]` / (Gorgotsias, 2024) ในรายงานหรือเอกสารที่อ้างถึงคลังข้อสอบ
- รายละเอียดและความเสี่ยงที่ยังเปิดอยู่: [`../docs/REFERENCES.md`](../docs/REFERENCES.md)

## 8. กฎการแก้ไขไฟล์ต้นทาง

ห้ามแก้ไขเนื้อหาข้อสอบในไฟล์ Markdown ของไดเรกทอรีนี้โดยไม่ดำเนินการต่อไปนี้พร้อมกัน:

1. แจ้งและบันทึกเหตุผลของการแก้ไข
2. อัปเดตข้อมูล Provenance ใน [`../docs/REFERENCES.md`](../docs/REFERENCES.md) หากการแก้ไขกระทบจำนวนข้อ ช่วงข้อ หรือการแมป Domain
3. อัปเดตทะเบียนแหล่งอ้างอิง (`R-ID`, Revision History) เมื่อมีการเปลี่ยนแปลงที่เกี่ยวกับที่มาของข้อมูล
4. รัน `python scripts/parse_exams.py` ใหม่และตรวจสอบว่าจำนวนข้อที่ได้ตรงกับที่คาดหวัง

การแก้ไขเชิงรูปแบบ (เช่น การเพิ่มจุดในหัวข้อคำถาม) ต้องไม่เปลี่ยนเนื้อหาคำถามและต้องบันทึกไว้ในหัวข้อ 4 ของเอกสารนี้
