# Scoring and Proficiency Contract

## Purpose

กำหนดการเก็บคะแนนดิบ การคำนวณ Proficiency Score และการอัปเดต User Skill Profile โดยไม่ทำลายข้อมูลต้นฉบับ

สถานะ: **Draft — Contract สำหรับ Antigravity**

---

## 1. Score Layers

ระบบต้องแยกคะแนนอย่างน้อย 3 ชั้น:

```text
Raw Evidence
→ Component Scores
→ Scenario Score
→ Proficiency Score
```

### Raw Evidence

- คำตอบ TP/FP
- Key Findings ที่ค้นพบ
- Response Actions ที่เลือก
- เวลาเริ่ม/จบ
- Log Queries/Investigation Events
- Scenario Snapshot ที่ผู้เรียนเห็นจริง

### Component Scores

- `triage_score`
- `finding_score`
- `response_score`
- `reasoning_score` เมื่อเปิดใช้
- `time_score` ถ้ามีการกำหนด

### Scenario Score

คะแนนรวมของ Scenario หนึ่งครั้ง ต้องเก็บพร้อม Formula Version และ Rubric Version

### Proficiency Score

คะแนนความสามารถราย Domain ช่วง 0–100 ใช้สำหรับ Adaptive Scenario Selection

---

## 2. Recommended Scenario Score

เริ่มต้นด้วยน้ำหนักที่อธิบายได้:

```text
Scenario Score =
  20% Triage Accuracy
+ 40% Key Finding Discovery
+ 40% Response Action Accuracy
```

เหตุผล:

- การค้นหา Evidence และการเลือก Response เป็นหัวใจของ Blue Team Simulation
- TP/FP ยังมีน้ำหนักเพื่อวัด Triage
- Reasoning ยังไม่รวมในคะแนนหลักจนกว่าจะมี Rubric และ Reliability ที่ผ่านการตรวจสอบ

น้ำหนักต้องเก็บใน `scoring_rubric_version` ไม่ควร Hard-code โดยไม่มี Version

---

## 3. Component Scoring Rules

### TP/FP Classification

```text
ถูกต้อง = 100
ไม่ถูกต้อง = 0
```

### Key Findings

```text
finding_score = points ของ Findings ที่ค้นพบ / points ของ Findings ทั้งหมด × 100
```

ต้องกำหนดเพดาน 0–100 และป้องกันการนับ Finding ซ้ำ

### Response Actions

เริ่มต้นใช้คะแนนจาก Action Metadata:

- Correct Action ที่เลือก: ได้คะแนนตาม Points
- Incorrect Action ที่เลือก: หักคะแนนตาม Points/Policy
- Action ที่ถูกต้องแต่ไม่เลือก: ไม่ได้คะแนน

ต้องกำหนด Minimum Score เป็น 0 เพื่อไม่ให้คะแนน Scenario ติดลบ เว้นแต่ Research Protocol ระบุเป็นอย่างอื่น

### Reasoning

ในระยะ Experimental:

- เก็บเป็น Raw Text
- ไม่รวมใน Primary Score
- ใช้ Rubric Version เมื่อเริ่มประเมิน
- เก็บ Human/LLM Reviewer และ Reliability Metadata แยกจากคะแนนหลัก

---

## 4. Proficiency Update

ระบบต้องเก็บทั้งคะแนนเดิมและคะแนนใหม่:

```text
new_proficiency =
  (1 - alpha) × previous_proficiency
  + alpha × scenario_domain_score
```

ค่าเริ่มต้นที่เสนอ:

```text
alpha = 0.30
```

แต่ต้องเก็บ:

```text
calculation_method = weighted_average
calculation_version = proficiency-v1
alpha = 0.30
```

หาก Scenario หนึ่งเกี่ยวข้องหลาย Domain ให้มี Mapping ของ Points ต่อ Domain ก่อนอัปเดต Profile

---

## 5. Pre-test Initialization

Pre-test ใช้สร้าง Baseline ของแต่ละ Domain:

```text
pre_domain_score = correct_answers_in_domain / total_questions_in_domain × 100
```

ระบบต้องเก็บ:

- Raw Answers
- Correctness ต่อข้อ
- Domain Blueprint Version
- Pre-test Score รวม
- Pre-test Score ราย Domain
- Calculation Version

ไม่ควรนำคะแนน Pre-test ไปเขียนทับด้วยคะแนน Simulation; ให้เก็บเป็นคนละ Measurement

---

## 6. Threshold Evaluation

Default:

```text
proficiency_threshold = 70
```

เงื่อนไข Eligible:

```text
ทุก Domain มี proficiency_score >= threshold
```

ต้องประเมินจาก Snapshot ของ Profile ณ จุดตรวจสอบ และบันทึก:

- Threshold ที่ใช้
- Domain Scores ทุก Domain
- Domains Passed
- Domains Remaining
- Evaluation Timestamp
- Completion Reason

---

## 7. Immutability and Audit

- Raw Score ห้ามแก้ไขย้อนหลัง
- Calculation Version ต้องเปลี่ยนเมื่อสูตรเปลี่ยน
- Recalculation ต้องสร้างผลลัพธ์ Version ใหม่
- ต้องสามารถย้อนจาก Proficiency ไป Scenario Scores และ Raw Evidence
- การเปลี่ยน Rubric ต้องไม่เปลี่ยนคะแนนประวัติเดิมโดยเงียบ ๆ

---

## 8. Score Test Cases

- TP ถูกต้อง + Findings ครบ + Actions ถูกต้อง → คะแนนสูงสุด
- TP ผิด แต่ Findings/Actions ถูก → หักเฉพาะ Triage ตามน้ำหนัก
- ไม่ค้นพบ Finding → Finding Score เป็น 0 ในส่วนนั้น
- เลือก Incorrect Action → หักตาม Policy แต่ Scenario Score ไม่ต่ำกว่า 0
- Submit ซ้ำ → ไม่สร้างคะแนนซ้ำ
- Recalculate ด้วย Version ใหม่ → เก็บผลเดิมและผลใหม่แยกกัน
- Domain ต่ำกว่า 70 → Loop ดำเนินต่อ
- ทุก Domain ตั้งแต่ 70 ขึ้นไป → Eligible for Post-test

---

## Open Decisions

- ค่า `alpha` สุดท้ายต้องยืนยันด้วยการ Calibration
- จะ Normalize คะแนนจาก Scenario Difficulty อย่างไร
- จะจัดการ Scenario ที่ครอบคลุมหลาย Domain อย่างไร
- จะรวม Time Score หรือไม่

---

**เอกสารนี้กำหนด Contract ของคะแนน ไม่ใช่ผลการทดลอง**

---

## Review Note

สูตรน้ำหนักและค่า `alpha` ในเอกสารนี้เป็นค่าเริ่มต้นสำหรับการออกแบบ ไม่ใช่ค่าที่พิสูจน์แล้วทางวิจัย ต้องบันทึกการเปลี่ยนแปลงและทดสอบ Sensitivity Analysis ก่อนใช้สรุปผลจริง

---

## Research Metrics

Primary Outcome ยังคงเป็น:

```text
Post-test Total Score - Pre-test Total Score
```

Scenario Score และ Proficiency Score เป็น Secondary/Process Outcomes เว้นแต่ Protocol ฉบับอนุมัติจะกำหนดต่างออกไป

---

## Implementation Boundary

การคำนวณคะแนนควรอยู่ที่ Backend และใช้ข้อมูล Ground Truth จาก Scenario Snapshot ไม่ควรเชื่อคะแนนที่ Frontend ส่งมาโดยตรง

ไม่ควรใช้ LLM เป็นผู้ตัดสินคำตอบ Checkbox ซึ่งมี Ground Truth อยู่แล้ว

---

## Acceptance Criteria

- [ ] เก็บ Raw Evidence ครบ
- [ ] Component Scores คำนวณซ้ำได้
- [ ] Scenario Score มี Rubric Version
- [ ] Proficiency Score มี Calculation Version
- [ ] คะแนนไม่ติดลบโดยไม่ตั้งใจ
- [ ] Submit ซ้ำไม่สร้างผลซ้ำ
- [ ] ทุก Domain ถูกตรวจ Threshold
- [ ] คะแนนย้อนหลังตรวจสอบกลับไปยัง Snapshot ได้


---

## Revision History

| Version | Change |
|---|---|
| v1 | Initial scoring and proficiency contract |


---

## Terminology

ใช้คำตาม `CONTEXT.md`:

- `Proficiency Score`
- `Proficiency Threshold`
- `Key Finding`
- `Response Action`
- `Simulation Session`
- `Scenario Snapshot`


---

## Operational Warning

หากเปลี่ยน Formula หรือ Threshold ระหว่างการเก็บข้อมูล ต้องสร้าง Protocol/Calculation Version ใหม่ และไม่รวมผลต่าง Version โดยไม่มีแผนวิเคราะห์รองรับ


---

## End State

เมื่อทุก Domain ผ่าน Threshold ระบบจะเปลี่ยน Session เป็น `eligible_for_post_test` และรอให้ผู้เรียนกดเริ่ม Post-test ผ่าน Completion Page


---

## Not Eligible State

เมื่อหมด Scenario Cap หรือ Time Limit ก่อนผ่าน Threshold ระบบจะบันทึก `not_eligible_for_post_test` พร้อม Reason Code และ Domain Breakdown


---

## Data Quality

ค่าที่ขาดหายต้องไม่ถูกแทนด้วยศูนย์โดยอัตโนมัติโดยไม่มีการระบุว่า Missing เพราะศูนย์และ Missing เป็นคนละความหมายในการวิเคราะห์


---

## Future Extension

สามารถเพิ่ม Difficulty-Adjusted Score หรือ Item Response Theory ได้ภายหลัง แต่ต้องไม่เพิ่มความซับซ้อนก่อนมี Dataset และ Protocol เพียงพอ


---

## Final Check

สูตรทั้งหมดในเอกสารนี้ต้องถูก Implement เป็น Unit Tests ก่อนนำไปใช้กับข้อมูลผู้เรียนจริง


---

## Ownership

- Product/Research Decision: เจ้าของ Senior Project
- Code Implementation: Antigravity
- Review: ผู้จัดทำร่วมกับอาจารย์/ผู้เชี่ยวชาญ


---

## Status

Draft — ยังไม่ใช่คำสั่งแก้โค้ดอัตโนมัติ


---

## End

