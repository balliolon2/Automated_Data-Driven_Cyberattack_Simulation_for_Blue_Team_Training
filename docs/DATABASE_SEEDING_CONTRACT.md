# Database Seeding Contract

## Purpose

กำหนด Contract สำหรับ Schema, Migration, Question Seeding, Static Scenario Seeding และ Real-time Scenario Snapshot โดยเน้น Idempotency และการปกป้องข้อมูลผู้เรียน/งานวิจัย

สถานะ: **Draft — สำหรับ Antigravity ใช้ออกแบบ Implementation**

---

## 1. Seed Layers

```text
Schema / Extensions
→ Reference Data
→ Security+ Questions
→ Static Scenario Pool
→ Optional Demo Data (Local Only)
```

### Production/Research Default

- Schema และ Reference Data: อนุญาต
- Security+ Questions: อนุญาตแบบ Idempotent
- Static Scenario Pool: อนุญาตแบบ Versioned Upsert
- Demo User/Research Data: ห้ามโหลดอัตโนมัติ
- Destructive Reset: ห้ามใช้ใน Cloud/Research Mode

---

## 2. Data Classes

| Data Class | ตัวอย่าง | Seed Policy |
|---|---|---|
| Schema | tables, enums, extensions | Migration/Apply ได้ |
| Reference | Security+ Domains, MITRE References | Idempotent Upsert |
| Learning Content | Questions, Static Scenarios | Versioned Upsert |
| Runtime Data | Users, Sessions, Actions | ห้ามลบระหว่าง Seed |
| Research Data | Responses, Snapshots, Metrics | Append/Preserve |
| Demo Data | Demo User, Sample Sessions | Local-only Explicit Flag |

---

## 3. Required Identity and Versioning

Content ที่ Seed ต้องมี Identity ที่เสถียรและ Version:

- `content_key`
- `content_version`
- `source_file`
- `source_hash`
- `active`
- `created_at`
- `updated_at`

ห้ามใช้ Array Index เป็น Identity หลักของ Question, Finding หรือ Action

---

## 4. Safe Seed Rules

`seed` ปกติต้อง:

1. ตรวจ Database Connection
2. ตรวจ Schema Version
3. ตรวจ Source File
4. Validate JSON/Markdown ก่อนเขียน
5. Upsert Reference และ Learning Content
6. ไม่ลบ User, Exam Session, Simulation Session, Session Action หรือ Research Data
7. รายงานจำนวน Insert/Update/Skip/Error
8. ทำ Transaction เท่าที่เหมาะสม
9. Rollback เมื่อ Validation/Database Error

### Prohibited Default Behavior

ห้ามใช้พฤติกรรมปัจจุบันของ `scripts/load_scenarios_to_db.py` ที่ลบ `simulation_sessions` และ `scenarios` ก่อน Seed เป็น Default ใน Production/Research

หากจำเป็นสำหรับ Development ให้แยกคำสั่ง:

```bash
./install.sh reset-dev-data
```

และต้องมีเงื่อนไข:

- ตรวจ `DEPLOYMENT_MODE=local`
- ปฏิเสธเมื่อเป็น Cloud/Research
- แสดงผลกระทบ
- ขอ Confirmation
- ไม่แตะข้อมูลภายนอก Scope ที่ระบุ

---

## 5. Real-time Scenario Snapshot

ทุก Scenario ที่ถูกนำเสนอแก่ผู้เรียนต้องมี Snapshot แยกจาก Scenario Template/Pool:

```text
scenario_snapshot_id
simulation_session_id
scenario_id หรือ generated_scenario_key
scenario_json
source_type: ai_generated | static_fallback
provider
model
prompt_version
blueprint_version
rubric_version
validation_status
fallback_reason
selection_reason
rendered_at
```

User Response ต้องอ้างอิง Snapshot ไม่ใช่อ่าน Scenario ล่าสุดจาก Pool เพราะ Pool อาจเปลี่ยนภายหลัง

---

## 6. Validation Contract

ก่อน Seed หรือใช้ Scenario ต้องตรวจ:

- Required Fields ครบ
- Domain ID มีอยู่จริง
- MITRE Technique Format ถูกต้อง
- Difficulty อยู่ในช่วงที่กำหนด
- Logs เป็น Array และมี Schema
- Key Findings อ้างอิง Log Index ที่มีอยู่
- Response Actions มี Correctness และ Points
- TP/FP Ground Truth มีค่า
- ไม่มี Secret หรือ PII ที่ไม่จำเป็น

LLM Scenario ที่ไม่ผ่าน Validation ต้องไม่ถูกนำเสนอแก่ผู้เรียน

---

## 7. Seed Commands Contract

```bash
./install.sh seed
./install.sh seed --content-version vX
./install.sh check-seed
./install.sh reset-dev-data
```

### Expected Semantics

- `seed`: Safe, repeatable, non-destructive
- `seed --content-version`: ใช้ Content Version ที่ระบุ
- `check-seed`: Validate โดยไม่เขียนข้อมูล
- `reset-dev-data`: Local-only destructive operation

---

## 8. Verification Queries/Checks

หลัง Seed ต้องตรวจอย่างน้อย:

- Security+ Domain ครบ 5 Domains
- Pre-test Candidate Pool รองรับ 30 ข้อตาม Blueprint
- Post-test Pool ไม่ใช้คำถามชุดเดียวกับ Pre-test Session เดิม
- Scenario มี Domain และ Difficulty ถูกต้อง
- Key Finding ทุกตัวอ้างอิง Log ได้
- Response Action ทุกตัวมี Scoring Metadata
- Seed ซ้ำแล้วจำนวน Content ไม่เพิ่มซ้ำ
- Runtime/Research Data ยังอยู่ครบ

---

## 9. Backup and Recovery

ก่อน Migration ที่เปลี่ยน Schema หรือ Content สำคัญ:

1. ตรวจ Backup ล่าสุด
2. บันทึก Schema/Content Version
3. Apply Migration
4. Run Validation
5. ถ้าล้มเหลวให้หยุดและ Restore ตาม Runbook

ห้ามใช้การลบ Volume เป็นวิธีแก้ปัญหา Default

---

## 10. Acceptance Criteria

- [ ] Seed ซ้ำได้โดยไม่เกิด Duplicate
- [ ] Safe Seed ไม่ลบข้อมูลผู้เรียน
- [ ] Destructive Reset ทำได้เฉพาะ Local
- [ ] Scenario Snapshot ตรวจสอบย้อนหลังได้
- [ ] LLM Output ผ่าน Validation ก่อนใช้งาน
- [ ] Static Fallback มี Version และ Source
- [ ] Questions/Scenarios มี Stable Identity
- [ ] มี Seed Report
- [ ] มี Backup/Restore Procedure

