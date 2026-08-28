# Failure and Session Status Contract

## Purpose

แยกความล้มเหลวของผู้เรียนออกจากความล้มเหลวของระบบ เพื่อให้ User Experience และ Research Dataset ถูกต้อง

สถานะ: **Draft — Contract สำหรับ Antigravity**

---

## 1. Status Categories

### Learner Progress

```text
in_progress
completed
abandoned
threshold_not_reached
max_scenarios_reached
time_limit_reached
```

### Eligibility

```text
eligible_for_post_test
not_eligible_for_post_test
```

### System Failure

```text
system_incomplete
scenario_unavailable
service_unavailable
```

สถานะเหล่านี้อาจถูกเก็บเป็น Field แยก แทนการยัดทุกความหมายลงใน Enum เดียว

---

## 2. Classification Rules

| เหตุการณ์ | Classification | นับเป็น Completed Scenario? | อัปเดต Proficiency? |
|---|---|---:|---:|
| ผู้เรียนส่งคำตอบตามปกติ | learner_completed | ใช่ | ใช่ |
| ผู้เรียนออกเอง | abandoned | ไม่ใช่ | ไม่ |
| หมด Scenario Cap | max_scenarios_reached | ตามจำนวนที่ทำจริง | อัปเดตเฉพาะที่เสร็จ |
| หมดเวลา | time_limit_reached | ตามจำนวนที่ทำจริง | อัปเดตเฉพาะที่เสร็จ |
| LLM/Static สร้าง Scenario ไม่ได้ | scenario_unavailable | ไม่ใช่ | ไม่ |
| Backend/DB ล้มเหลว | system_incomplete | ไม่ใช่ | ไม่ |
| ผู้เรียนทำไม่ถึง Threshold แต่ Guardrail หมด | threshold_not_reached | ตามจำนวนที่ทำจริง | อัปเดตเฉพาะที่เสร็จ |

---

## 3. Completion Rules

### Eligible

```text
ทุก Domain >= Threshold
→ eligible_for_post_test
→ แสดง Completion Page
→ รอผู้เรียนกดเริ่ม Post-test
```

### Not Eligible

```text
Threshold ไม่ครบ + Cap/Time หมด
→ not_eligible_for_post_test
→ แสดง Domain Breakdown และ Reason
→ ไม่ให้เข้า Post-test ตาม Protocol
```

### System Incomplete

```text
Failure เป็นของระบบ
→ system_incomplete
→ ไม่ตัดสินว่า Learner Failed
→ preserve session data
→ อนุญาต Resume/Retry ตาม Policy
```

---

## 4. Research Data Requirements

ทุก Session ต้องมี:

- `status`
- `completion_reason`
- `failure_class`
- `started_at`
- `completed_at` หรือ `ended_at`
- `scenario_count_completed`
- `scenario_count_attempted`
- `elapsed_seconds`
- `threshold_value`
- `max_scenarios`
- `time_limit_seconds`
- Domain Breakdown

ข้อมูล `system_incomplete` ต้องแยกออกจาก `threshold_not_reached` ใน Analysis Dataset

---

## 5. User-Facing Messages

### Learner Not Eligible

ควรสื่อสารว่า:

> ยังไม่ถึงเกณฑ์ Proficiency Threshold ภายในข้อจำกัดของรอบนี้

แสดง:

- Domain ที่ผ่าน
- Domain ที่ยังไม่ผ่าน
- จำนวน Scenario ที่ทำ
- เหตุผลที่หยุด
- คำแนะนำถัดไปตาม Research Policy

### System Incomplete

ควรสื่อสารว่า:

> ระบบไม่สามารถเตรียม Scenario ได้ครบถ้วน ข้อมูล Session ของคุณถูกเก็บไว้ และเหตุการณ์นี้จะไม่ถูกนับเป็นความผิดพลาดของผู้เรียน

ห้ามใช้คำว่า Failed กับ System Failure

---

## 6. Resume Rules

- Refresh ระหว่าง Investigation → กลับไป Snapshot เดิม
- Network หลุด → ไม่สร้าง Scenario ใหม่โดยอัตโนมัติจนกว่าจะตรวจ Session
- System Failure → Resume ได้ถ้า Snapshot ยัง Valid
- Abandoned → ต้องกำหนดว่าจะ Resume หรือเริ่มใหม่ตาม Research Protocol
- Completed → ห้าม Submit ซ้ำเพื่อสร้างคะแนนใหม่

---

## 7. State Transition Invariants

- Session หนึ่งต้องมีสถานะปัจจุบันที่ชัดเจน
- State Transition ต้องมี Timestamp
- ห้ามเปลี่ยนจาก `completed` กลับเป็น `in_progress`
- `system_incomplete` ห้ามกลายเป็น `threshold_not_reached` โดยไม่มีหลักฐานว่า Learner เป็นสาเหตุ
- Scenario ที่ไม่พร้อมใช้งานห้ามเพิ่ม `scenario_count_completed`
- Post-test ต้องตรวจ Eligibility จาก Backend ไม่เชื่อ Frontend เพียงอย่างเดียว

---

## 8. Test Cases

- Learner submits valid response → completed scenario
- Learner closes page → abandonment/recovery policy
- Cap reached → not eligible with max_scenarios_reached
- Time expired → not eligible with time_limit_reached
- LLM and Static Pool unavailable → scenario_unavailable
- Database failure during submit → system_incomplete, no duplicate score
- Refresh active session → same Snapshot
- Completed session submit again → rejected/idempotent
- Post-test request before eligibility → denied
- Post-test request after eligibility → allowed to start

---

## 9. Acceptance Criteria

- [ ] Learner Failure และ System Failure แยกกันได้
- [ ] Research Dataset แยก Status ได้
- [ ] Session Recovery ไม่สร้าง Scenario ซ้ำ
- [ ] Post-test ถูกป้องกันที่ Backend
- [ ] Scenario Unavailable ไม่ลงโทษผู้เรียน
- [ ] User Message ไม่กล่าวโทษเมื่อระบบล้มเหลว
- [ ] State Transition ตรวจสอบย้อนหลังได้

