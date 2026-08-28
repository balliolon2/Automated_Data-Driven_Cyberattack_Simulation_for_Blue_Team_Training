# Grill Decision Log

เอกสารบันทึกคำตอบและเหตุผลจากการ Grill เพื่อไม่ให้ข้อกำหนดของ Senior Project เปลี่ยนไปโดยไม่ตั้งใจ

สถานะ: **รอการยืนยัน Shared Understanding ก่อนเริ่ม Implementation**

---

## Round 1 — Project Reframing

| Question | Decision |
|---|---|
| Q1 | Workflow หลักคือ Pre-test Security+ 30 ข้อ → Scenario → Logs → Handle/Response → Evaluation |
| Q2 | ไม่เริ่มจาก Attacker/Target/Monitoring Cyber Range; เป็น Scenario-based Blue Team Decision Simulation |
| Q3 | ต้องการ Scenario Coverage หลายหมวดตาม MITRE ATT&CK ตั้งแต่ต้น |

---

## Round 2 — Clarified Direction

| Question | Decision | Consequence |
|---|---|---|
| Q4 | **3 — Hybrid** | MVP เป็น Decision-based Simulation และออกแบบให้เพิ่ม Environment-based Simulation ในอนาคต |
| Q5 | **4 — Data-Driven ครบวงจร** | ใช้ Pre-test, Skill Profile, Adaptive Scenario, Behavior, Score และ Post-test เป็นข้อมูลต่อเนื่อง |
| Q6 | **3 — Mapping สองแกน** | Scenario ผูกทั้ง CompTIA Security+ Domain และ MITRE ATT&CK Technique |
| Q7 | **1 ไป 3** | เริ่มจาก Checkbox Response Actions แล้วพัฒนาไปสู่การเขียนเหตุผลประกอบ |
| Q8 | **3 + Multi-user Statistical Evaluation** | วัด Pre-test, Simulation, Post-test และวิเคราะห์ข้อมูลจากผู้เรียนหลายคน |
| Q9 | **Real-time Generation ตั้งแต่ต้น** | ระบบปัจจุบันใช้ Real-time LLM Generation และต้องมี Static Fallback |
| Q10 | **3 — Core + DevOps Extension** | Senior Project Core มาก่อน ส่วน DevOps แยกเป็น Extension |

---

## Round 3 — Research and Evaluation

| Question | Decision | Consequence |
|---|---|---|
| Q11 | **3 — One-group เป็นหลัก และเพิ่ม Comparison Group หากหาได้** | ระบบไม่ผูกกับ Research Design แบบเดียว |
| Q12 | **ตามคำแนะนำ: นักศึกษา/ผู้เรียนที่เกี่ยวข้อง และเก็บประสบการณ์เดิม** | ใช้ประสบการณ์เดิมเป็นข้อมูลประกอบการวิเคราะห์ |
| Q13 | **4 — Adaptive ต่อเนื่องจนทุก Domain ถึง Proficiency Threshold** | จำนวน Scenario ต่อคนไม่เท่ากัน ต้องกำหนด Guardrail และบันทึกจำนวนครั้งเพื่อวิเคราะห์ |
| Q14 | **3 + 4 — Blueprint ก่อน Generate และเก็บ Snapshot** | ควบคุมความเท่าเทียมและตรวจสอบ Scenario ย้อนหลังได้ |
| Q15 | **4 — Checkbox เป็นหลัก และ Reasoning เป็น Experimental Feature** | ลดความเสี่ยงจากการให้ LLM ประเมินข้อความโดยไม่มี Reliability ที่พิสูจน์แล้ว |
| Q16 | **4 — คะแนนรวมเป็น Primary Outcome; ราย Domain/Simulation เป็น Secondary** | ลด Multiple Testing และป้องกันการเลือกผลลัพธ์ย้อนหลัง |
| Q17 | **4 — รองรับหลาย Provider และบันทึก Metadata** | ต้องเก็บ Provider, Model และ Prompt Version โดยไม่เก็บ Secret |
| Q18 | **4 — รายงาน, Demo, Portfolio และการนำเสนอเชิงวิชาการ** | ต้องมี Reproducibility, Evidence, Limitations และ Versioning |

---

## Round 4 — Adaptive Loop and Deployment Bootstrap

| Question | Decision | Consequence |
|---|---|---|
| Q19 | **4 — Configurable Threshold, Default 70%** | Threshold ต้องบันทึกต่อ Research Session และไม่ Hard-code |
| Q20 | **4 — Maximum Scenario Cap + Maximum Time** | ต้องมี Guardrail ป้องกัน Loop ไม่จบ; ค่า Default เสนอ 10 Scenarios / 120 นาที |
| Q21 | **1 — ไม่ให้เข้า Post-test หากยังไม่ถึง Threshold** | ต้องกำหนดสถานะ Incomplete/Failed และระวัง Selection Bias ในการวิจัย |
| Q22 | **4 — เก็บจำนวน Scenario และวิเคราะห์เป็น Exposure Variable** | จำนวน Scenario เป็นตัวแปรร่วมในการวิเคราะห์ผลลัพธ์ |
| Q23 | **3 — เก็บ Snapshot ของทุก Scenario ที่นำเสนอ** | ทำ Research Reproducibility และตรวจสอบสิ่งที่ผู้เรียนเห็นจริงได้ |
| Q24 | **4 — Retry จำกัดครั้ง แล้ว Fallback** | ป้องกัน API Cost, Infinite Loop และความล่าช้า |
| Q25 | **3 — รองรับทั้ง Build จาก Source และ Versioned Images** | Development Build จาก Source; Reference Deployment ใช้ Versioned Images เป็นค่าเริ่มต้น |
| Q26 | **3 — เริ่มด้วย PostgreSQL Container และออกแบบย้ายไป Managed PostgreSQL** | ฟรีและทำซ้ำง่ายในระยะแรก แต่ต้องมี Backup/Restore |
| Q27 | **4 — Interactive Hidden Prompt + Secret File/Environment Variable** | ไม่รับ Secret ผ่าน CLI Argument และไม่แสดง Key ใน Output |
| Q28 | **Phase 1: Application Stack บน Ubuntu VM; Phase 2: HTTPS/Firewall/Backup** | `install.sh` ไม่ Provision VM ทุก Cloud Provider ในระยะแรก |

---

## Round 5 — Scoring, Fallback, Snapshot and Failure Classification

| Question | Decision | Consequence |
|---|---|---|
| Q41 | **4 — เก็บ Raw Score และ Calculated Proficiency Score แยกกัน; ใช้ Weighted Average เป็นหลัก** | เปลี่ยนสูตรได้โดยไม่ทำลาย Raw Data และต้องมี Calculation Version |
| Q42 | **4 — LLM → Static Pool → Domain ใกล้เคียง พร้อมบันทึกเหตุผล** | Adaptive Selection อธิบายได้และไม่สุ่มข้าม Domain โดยไม่มี Policy |
| Q43 | **2 — เก็บ Reasoning เป็น Optional Field แต่ยังไม่ใช้ตัดสินคะแนนหลัก** | เริ่มเก็บข้อมูลเพื่อพัฒนา Rubric ภายหลัง โดย Checkbox ยังเป็น Ground Truth หลัก |
| Q44 | **4 — Fallback ตามลำดับ แล้วจบเป็น `scenario_unavailable`** | ไม่ลงโทษผู้เรียนและไม่เพิ่ม Scenario Count/Proficiency เมื่อระบบล้มเหลว |
| Q45 | **3 — Parsed Snapshot + Metadata และ Raw Payload แบบจำกัด/ควบคุม** | ทำ Reproducibility ได้โดยไม่เก็บ Secret และลด Privacy Risk |
| Q46 | **3 — Pull Versioned Images เป็น Default สำหรับ Cloud** | Local Build จาก Source ได้; Cloud ต้อง Rollback และ Reproduce ได้ |
| Q47 | **3 — รองรับ Container และ External DB; เริ่มจาก Container** | ฟรีและทำซ้ำง่าย พร้อมเส้นทางย้ายไป Managed PostgreSQL |
| Q48 | **3 — แยกเป็น `system_incomplete`** | ไม่ปน System Failure กับ Learner Failure ใน Research Dataset |

---

## Current Shared Understanding

> แพลตฟอร์มฝึก Blue Team ที่เริ่มจาก Pre-test CompTIA Security+ 30 ข้อ แล้วสร้าง Scenario แบบ Adaptive และ Real-time ให้ผู้เรียนตรวจสอบ Logs, ค้นหา Evidence, จำแนก TP/FP, เลือก Response Actions และพัฒนาต่อด้วย Post-test โดยใช้ข้อมูลจากผู้เรียนหลายคนเพื่อประเมินผลเชิงสถิติ

ระบบปัจจุบันเป็น Scenario-based Decision Simulation ไม่ใช่ Cyber Range ที่ยิงโจมตีระบบจริง และมีเป้าหมายให้เพิ่ม Environment-based Simulation ได้ในอนาคต

---

## New Deployment Direction

ผู้ใช้ต้องการวางแผน `install.sh` สำหรับติดตั้งระบบตั้งแต่ต้นจนพร้อมใช้งาน โดยครอบคลุม:

- สร้าง `.env`
- กรอก/จัดการ AI API Key
- ติดตั้งและตรวจสอบ Database
- Apply Schema/Migration
- โหลด Security+ Questions
- โหลด Static Scenarios
- Start Backend/Frontend
- Health Check และ Smoke Test
- ใช้เป็น Reference Deployment บน Cloud ได้ง่าย

แนวทางที่ยืนยันในเอกสาร `docs/DEPLOYMENT_BOOTSTRAP_PLAN.md` คือ **Ubuntu VM + Docker Compose เป็น Reference Deployment** ก่อน ไม่สัญญาว่า Shell Script เดียวจะ Provision ได้ทุก Cloud Provider

ข้อควรระวังที่ค้นพบจากการสำรวจโค้ด:

- `.env.example` ที่ตรวจพบมีค่า API Key ถูก Redact อยู่ในไฟล์ตัวอย่าง จึงต้องตรวจว่าไม่มี Secret จริงและไม่ควรใส่ค่าที่มีรูปแบบคล้าย Key ในตัวอย่าง
- `scripts/load_scenarios_to_db.py` ลบ Simulation Sessions และ Scenarios ก่อน Seed ใหม่ จึงห้ามเรียกเป็น Default ใน Production/Research Environment โดยไม่มี Safe Mode หรือ Explicit Destructive Flag
- การใช้ `.env` บน Cloud ต้องเพิ่ม Permission, Secret Handling, Backup และไม่ส่ง Key ไป Frontend

---

## Important Corrections

### Correction 1 — Real-time Scenario Generation

Real-time LLM Generation เป็นเส้นทางหลักในระบบปัจจุบัน Static Scenario Pool เป็น Fallback และกลไกควบคุมความเสถียร/คุณภาพ

### Correction 2 — Scope of Simulation

ระบบปัจจุบันไม่ใช่การสร้าง Attacker → Target Network Environment เต็มรูปแบบ แต่เป็น Scenario ที่นำเสนอเหตุการณ์และ Telemetry ให้ Blue Team วิเคราะห์

### Correction 3 — Adaptive Stopping Rule

ผู้ใช้เลือกให้ Scenario Loop ดำเนินต่อจนทุก Domain ถึง Proficiency Threshold ไม่ใช่จำนวน Scenario คงที่ สิ่งนี้รักษาแนวคิด Adaptive Learning แต่ทำให้การวิเคราะห์ระหว่างผู้เรียนต้องพิจารณา Exposure และจำนวน Scenario เป็นตัวแปรร่วม

---

## Decisions Still Open — Next Frontier

หลัง Q41–Q48 ข้อกำหนดระดับทิศทางหลักถือว่าชัดเจนเพียงพอสำหรับการทำ Implementation Plan แล้ว ขั้นถัดไปควรเป็นการยืนยัน Shared Understanding จากเจ้าของโครงการ และแตกงานเป็น Milestone ที่ Antigravity ดำเนินการได้ ไม่ใช่เพิ่มคำถามเชิงสถาปัตยกรรมไปเรื่อย ๆ

ประเด็นที่ยังเปิดอยู่เป็น **Implementation Detail** ซึ่งควรตัดสินใจในเอกสาร/Issue ที่เกี่ยวข้อง:

- ค่า `alpha` และ Formula ของ Proficiency ต้อง Calibration ด้วยข้อมูลจริง
- JSON Schema/Validation Library ใน Go
- วิธีเลือก Static Fallback เมื่อไม่มี Candidate ตรง Domain
- วิธีนับเวลาเมื่อ Browser ปิดหรือ Network หลุด
- รูปแบบ Secret Manager ของ Cloud Provider เป้าหมาย
- วิธี Provision VM, Domain, HTTPS และ Backup ใน Phase 2
- รายละเอียด Statistical Power และ Sample Size กับอาจารย์/ผู้เชี่ยวชาญ

---

## Confirmation Gate

ก่อนเริ่มแก้โค้ดหรือสร้าง `install.sh` จริง ต้องยืนยันว่า Shared Understanding ต่อไปนี้ถูกต้อง:

- [ ] Core คือ Pre-test 30 ข้อ → Adaptive Real-time Scenario → Logs/Evidence → TP/FP → Response Actions → Eligibility → Post-test
- [ ] ระบบปัจจุบันเป็น Scenario-based Decision Simulation ไม่ใช่การโจมตีระบบจริง
- [ ] Real-time LLM Generation เป็นเส้นทางหลัก และมี Static Fallback
- [ ] ทุก Scenario ที่ผู้เรียนเห็นต้องมี Snapshot และ Metadata
- [ ] Threshold Default 70%, Cap Default 10 Scenarios, Time Limit Default 120 นาที
- [ ] ผู้เรียนที่ยังไม่ถึง Threshold ไม่มีสิทธิ์เข้า Post-test ตาม Protocol
- [ ] ผู้เรียนที่ระบบล้มเหลวต้องถูกแยกเป็น `system_incomplete`
- [ ] Safe Seed ต้องไม่ลบ User/Session/Research Data
- [ ] Cloud Reference Deployment ใช้ Ubuntu VM + Docker Compose
- [ ] Cloud Default ใช้ Versioned Images และไม่ใช้ `latest`
- [ ] DevOps เป็น Extension แยกจาก Senior Project Core

เมื่อรายการนี้ได้รับการยืนยัน จึงเริ่ม Implementation Plan/Issue Breakdown ได้

---

## Rule for Future Changes

เมื่อมีข้อเสนอใหม่ ให้ระบุว่าเป็น:

- Core Requirement
- Research Requirement
- Safety Requirement
- Deployment Requirement
- DevOps Extension
- Future Enhancement

และต้องบันทึกผลกระทบต่อ:

- User Journey
- Data Model
- Evaluation Validity
- Security Boundary
- Deployment/Operations
- Timeline

