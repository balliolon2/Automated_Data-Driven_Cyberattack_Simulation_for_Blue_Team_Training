# คู่มือวิทยากรและแผนดำเนินกิจกรรมเวิร์กช็อป 2 ชั่วโมง (High School Cyber Detective Runbook)

คู่มือฉบับนี้ออกแบบสำหรับวิทยากรที่นำแพลตฟอร์ม **SOC Trainer** ไปใช้จัดกิจกรรมเชิงปฏิบัติการ 2 ชั่วโมง (120 นาที) ให้แก่นักเรียนระดับมัธยมศึกษาตอนปลายที่มีพื้นฐานด้าน Cybersecurity เป็นศูนย์ โดยสอดคล้องกับโครงสร้างหน้าจอและระบบจริงทั้งหมดของแพลตฟอร์ม

---

## 1. Blueprint ช่วงเปิดกิจกรรม 10 นาทีแรก (สำหรับสไลด์เปิดและสร้างแรงจูงใจ)

### 1.1 อุปมา 3 สถานการณ์ชีวิตประจำวันของวัยรุ่น → งานจริงของนักวิเคราะห์ SOC Tier 1

1. **บัญชี IG / Roblox / Discord ถูกแฮกตอนตีสอง (Identity & Brute Force)**
   - **ข้อความชวนคิด:** *"ถ้าเพื่อนทักมาบอกว่า 'เฮ้ย แกโพสต์ขายของตอนตี 2 เหรอ?' ทั้งที่เราหลับอยู่ — แสดงว่ามีคนอื่นแอบล็อกอินเข้าบัญชีเราแล้ว!"*
   - **เชื่อมโยงเข้าสู่ SOC:** นักวิเคราะห์ SOC Tier 1 คอยสอดส่อง Log การล็อกอินที่ผิดปกติ (เช่น เดาพาสเวิร์ดผิด 18 ครั้งแล้วเข้าได้จาก IP ต่างประเทศตอนตี 2) เพื่อล็อกบัญชีและเตะผู้บุกรุกออกให้ทันเวลา
2. **SMS พัสดุค้างส่ง / ลิงก์รับทุนการศึกษาปลอม (Phishing & Malicious Attachment)**
   - **ข้อความชวนคิด:** *"เคยเห็นข้อความ 'กรุณากดลิงก์เพื่อยืนยันรับเงิน/พัสดุภายใน 15 นาที' ไหม? แค่คลิกเดียวอาจพาไปหน้าเว็บปลอมที่ดูดรหัสผ่านทั้งเครื่องได้"*
   - **เชื่อมโยงเข้าสู่ SOC:** เมื่อมีคนเผลอกดลิงก์หรือดาวน์โหลดไฟล์แปลกปลอม นักวิเคราะห์ SOC ต้องค้นหา **IOC (ร่องรอยการบุกรุก)** ในบันทึกเหตุการณ์ เพื่อแยกให้ออกว่าเหตุการณ์ไหนอันตรายจริง
3. **Wi-Fi โรงเรียนหรือร้านกาแฟช้าผิดปกติ / มีเครื่องสแกนพอร์ต (Internal Threat vs. Normal IT Maintenance)**
   - **ข้อความชวนคิด:** *"เวลาไฟร์วอลล์ดังเตือน อย่าเพิ่งแตกตื่น 100%! บางครั้งอาจเป็นพี่แอดมินฝ่าย IT กำลังอัปเดตหรือสำรองข้อมูลตามรอบปกติก็ได้"*
   - **เชื่อมโยงเข้าสู่ SOC:** หัวใจของนักสืบไซเบอร์คือการตัดสิน **TP/FP Decision** ให้แม่นยำ — **True Positive (โจมตีจริง)** ต้องรีบหยุดยั้ง ส่วน **False Positive (งานปกติ)** ต้องพิสูจน์จากหลักฐานเพื่อไม่ให้สั่งปิดระบบผิดพลาด

---

### 1.2 กิจกรรมอุ่นเครื่องหน้าโปรเจกเตอร์ 2 นาที: **"Spot the Hacker" (โหวตยกมือทั้งห้องก่อนล็อกอิน)**

> แนะนำให้เปิดหน้าแรกของแพลตฟอร์มขึ้นโปรเจกเตอร์ (Guest เริ่มต้นภาษาไทยอัตโนมัติ) หรือนำ 3 บรรทัด Log ด้านล่างขึ้นสไลด์เพื่อให้นักเรียนยกมือโหวตก่อนเปิดระบบให้นักเรียนทุกคนเข้า

**โจทย์หน้าชั้นเรียน (30 วินาที):**
> *เวลา 02:14 น. เครื่องเซิร์ฟเวอร์ชมรมคอมฯ ส่ง Log ออกมา 3 บรรทัด — นักสืบคิดว่านี่คือ **"แฮกเกอร์ตัวจริง (True Positive)"** หรือ **"ระบบทำงานปกติ (False Positive)"**?*

```text
02:14:02Z  [WARN]   [auth.sshd]  Failed password x18 for admin_club from 185.220.101.44 (Tor exit node)
02:14:19Z  [ALERT]  [auth.sshd]  LOGIN_SUCCESS user=admin_club src_ip=185.220.101.44 method=password
02:14:47Z  [ALERT]  [file.sftp]  DOWNLOAD path=/srv/club/passwords_backup.zip size=4.2MB src_ip=185.220.101.44
```

**จังหวะเฉลยของวิทยากร (30 วินาที):**
- ให้นักเรียนยกมือโหวต: *ใครว่า "โจมตีจริง" ยกมือ! / ใครว่า "ปกติ" ยกมือ!*
- กดปุ่ม **"โจมตีจริง (True Positive)"** บนแท็บ `30s-detective-case.log` หน้าแรกให้เห็นเฉลยทันที: *"สุ่มเดารหัส 18 ครั้งแล้วเข้าได้ตอนตี 2 พร้อมโหลดไฟล์รหัสผ่านออกไปทันที — ยินดีด้วย! นี่คืองานแรกของนักสืบ SOC ที่คุณเพิ่งทำสำเร็จ"*
- แจ้งนักเรียนว่าเมื่อเข้าไปทำ **Scenario** จริงบนเครื่องตัวเอง ให้ทำแบบนี้ 3 ขั้นตอนเดียวกัน (อ่านเรื่องย่อ → เลือก Log หลักฐานจริงไม่เลือกมั่ว → ตัดสิน TP/FP และเลือกวิธีรับมือ)

---

### 1.3 การเกริ่นนำ Pre-Test เพื่อลดความกดดันของนักเรียน ม.ปลาย

- **สคริปต์แนะนำวิทยากร:**
  > *"ด่านแรกคือ Pre-Test 30 ข้อ แต่อย่าเพิ่งกลัว! นี่**ไม่ใช่ข้อสอบเก็บคะแนนในห้องเรียน** ต่อให้ได้ 0% ก็ถือว่าปกติที่สุด เหมือนเกม RPG ที่ให้เรา 'วัดพลังตัวละคร Level 1' ก่อนลงดันเจี้ยน จะได้ดูกันท้ายคาบว่าพอฝึกจบแล้ว Level เราพุ่งขึ้นกี่คะแนน!"*
- **ย้ำตัวช่วยสำคัญหน้าชั้นเรียน:**
  1. มีคำแปลภาษาไทยให้อ่านเข้าใจง่ายทุกข้อ พร้อมปุ่ม **`ดูต้นฉบับ EN`** 1 คลิกถ้าอยากเทียบศัพท์ภาษาอังกฤษจริงของ CompTIA Security+
  2. ถ้าไม่รู้คำตอบจริงๆ ให้อ่านและตอบเท่าที่เข้าใจ ไม่ต้องเปิดเน็ตค้นหา เพื่อวัดพัฒนาการที่แท้จริง (เป้าเวลา ~20 นาที)

---

## 2. ตารางเวลา 120 นาทีแบบรายนาที (Minute-by-Minute Classroom Schedule)

| เวลา (ชั่วโมง:นาที) | ระยะเวลา | กิจกรรมและหน้าจอบนระบบ | คำพูดและคิวของวิทยากร (Facilitator Cues) |
| --- | :---: | --- | --- |
| **0:00 – 0:08** | 8 นาที | **Opening Hook & Projector Game**<br>หน้า `/` (`LandingPage`) | เปิดสไลด์อุปมา 3 เรื่อง (IG ถูกยึด / SMS ลวง / แจ้งเตือนปลอม), เล่นโหวต **"Spot the Hacker"** 3 บรรทัดหน้าห้อง, แนะนำแผนที่ 4 ด่านตลอด 2 ชั่วโมง |
| **0:08 – 0:10** | 2 นาที | **Registration & Login Queue**<br>หน้า `/register` → `/login` | แจ้ง URL ของระบบ (เช่น Azure domain), ให้นักเรียนสมัครสมาชิกด้วย **อีเมลที่จำได้ง่าย + ชื่อเล่นภาษาอังกฤษสั้นๆ (ห้ามซ้ำกัน)** แล้วกดปุ่ม **"เริ่มภารกิจ 2 ชั่วโมง"** |
| **0:10 – 0:30** | 20 นาที | **Stage 1: Diagnostic Pre-Test**<br>หน้า `/pre-test` (30 ข้อ, เกณฑ์เวลา 30 นาที) | **คิว `0:10`:** ชี้ให้ดูป้ายสีเขียวบนหน้าจอก่อนกดเริ่ม (*"ได้ 0% ก็ปกติ!"*) และย้ำปุ่ม `ดูต้นฉบับ EN`<br>**คิว `0:25`:** แจ้งเตือน *"เหลือเวลาประมาณ 5 นาทีสำหรับ Pre-Test ข้อไหนไม่มั่นใจตอบตามสัญชาตญาณแล้วกดส่งข้อสอบได้เลย"* |
| **0:30 – 0:50** | 20 นาที | **Stage 2: Domain Essentials Briefs**<br>หน้า `/learn` (5 โดเมน, ~3–4 นาที/โดเมน) | **คิว `0:30`:** แนะนำหน้า `/evaluation` ดูจุดอ่อนของตัวเอง (โดเมนที่ต่ำกว่า **Proficiency Threshold 70%**) แล้วเข้า `/learn`<br>ย้ำให้นักเรียนอ่านสาระสำคัญทั้ง **5 โดเมนให้ครบ 5/5** และกดปุ่ม **"บันทึกว่าอ่านสาระสำคัญโดเมนนี้จบแล้ว"** ทุกโดเมน (ส่วนหัวข้อย่อย Deep Dive ด้านล่างเป็นตัวเลือกเสริม ไม่ต้องคลิกครบทุกหัวข้อ) |
| **0:50 – 1:30** | 40 นาที | **Stage 3: Adaptive Scenario Loop**<br>หน้า `/simulation` (เป้าหมาย 3–5 Scenarios) | **คิว `0:50`:** ทุกคนเข้า **Scenario ที่ 1** แผง **"วิธีเล่นฉบับย่อใน 30 วินาที & ศัพท์ SOC พื้นฐาน"** จะกางให้อัตโนมัติ! พาทั้งห้องอ่าน Tab 1 (วิธีเล่น 3 ขั้น) และชี้ Tab 2 (**10 คำศัพท์ SOC**)<br>**คิว `1:05`:** เตือนสติเรื่องเลือก Log หลักฐาน (*"ห้ามติ๊กมั่วทุกบรรทัด เพราะเลือก Noise ผิดถูกหัก 2 คะแนนต่อรายการ!"*) และแนะนำ **Quick-Filter Chips**<br>**คิว `1:20`:** นักเรียนที่ถึง **Scenario 2+** แผงคู่มือจะยุบเก็บให้อัตโนมัติ (กดเปิดดูซ้ำได้เสมอ) |
| **1:30 – 1:50** | 20 นาที | **Stage 4A: Final Post-Test**<br>หน้า `/post-test` (30 ข้อ Zero-Overlap) | **คิว `1:30`:** นัดหยุดเล่น Simulation ชั่วคราว (หรือคนที่ครบ Threshold 70% แล้วกด **"เริ่มทำ Post-Test"** ได้เลย) เพื่อเข้าทำ Post-Test 30 ข้อวัดระดับพลังหลังฝึก *(หากทีมงานตั้ง `BYPASS_POST_TEST_THRESHOLD=true` สำหรับเวิร์กช็อป 2 ชั่วโมง ทุกคนสามารถกดเข้าทำ `/post-test` ได้พร้อมกันเวลา 1:30)* |
| **1:50 – 1:55** | 5 นาที | **Stage 4B: Post-Test Questionnaire**<br>หน้า `/questionnaire` | **คิว `1:50`:** ให้นักเรียนกรอกแบบสอบถามความพึงพอใจ **CSUQ v3 (1–7)** และภาระงาน **NASA-TLX** พร้อมดูกราฟพัฒนาการ **Pre→Post** และเรดาร์ภาระงานของตัวเองทันทีหลังส่ง |
| **1:55 – 2:00** | 5 นาที | **Closing & Live Cohort Reveal**<br>หน้า `/admin/evaluations` & `/admin/evaluations/statistics` | **คิว `1:55`:** ขอบคุณนักเรียนทุกคนและเฉลยภาพรวมความสำเร็จ<br>**คิว `1:58` (Live Reveal):** เปิดหน้า **Admin Operational Evaluations** (2×2 Dashboard: Funnel การมีส่วนร่วม, การกระจาย CSUQ, NASA-TLX Radar, และกราฟเปรียบเทียบ **Pre→Post รายบุคคลทั้งห้อง**) ขึ้นโปรเจกเตอร์เพื่อให้นักเรียนเห็นพลังการเรียนรู้ของห้องเรียนแบบเรียลไทม์ |

---

## 3. เช็กลิสต์ก่อนเริ่มคาบ & แนวทางแก้ปัญหาหน้างาน (Azure Pre-Flight & Classroom Troubleshooting)

### 3.1 การตั้งค่าก่อนเปิดห้องเรียน (ล่วงหน้า 30 นาที)

1. **Image & Default Locale:** ตรวจสอบว่า frontend Docker image ถูก build ด้วย `VITE_DEFAULT_LOCALE=th` เพื่อให้เครื่องของนักเรียนที่ไม่เคยเข้าเว็บมาก่อนเปิดขึ้นมาเป็น **ภาษาไทยทันที**
2. **โหมด Post-Test สำหรับเวิร์กช็อป 2 ชั่วโมง:** พิจารณาการตั้งค่า `BYPASS_POST_TEST_THRESHOLD` ในไฟล์ `.env` ตามแผนเวิร์กช็อป (หากต้องการให้นักเรียนทุกคนทำ Post-Test พร้อมกันในนาทีที่ `1:30` แม้ยังทำไม่ครบทุก Scenario ให้ตั้ง `true` ก่อนเริ่มคาบ หรือคง `false` หากต้องการวัดผลตามเกณฑ์ `Threshold 70%` จริง — อย่าลืมบันทึกเงื่อนไขไว้ในรายงานวิจัย)
3. **คลัง Scenario และ AI Generation:**
   - ตรวจสอบ `docker compose ps` ว่า `cyber_sim_db`, `cyber_sim_backend`, `cyber_sim_frontend` อยู่ในสถานะ `healthy`
   - ตรวจสอบผล `seed` ว่ามีข้อสอบ CompTIA Security+ (EN + TH 540 ข้อ), เนื้อหา 5 โดเมน (EN + TH) และ Static Scenario Pool พร้อมใช้งานครบถ้วน
4. **เปิดหน้า Admin เตรียมไว้ในแท็บส่วนตัวของวิทยากร:** ล็อกอิน Admin ล่วงหน้าที่ `/admin/evaluations` และ `/admin/evaluations/statistics` เพื่อให้กด รีเฟรช โชว์กราฟช่วง **`1:58`** ได้ทันที

### 3.1.5 Azure Cloud VM Sizing, HTTPS Deployment & Clean-Slate DB Checklist (รองรับ 60 ผู้เรียนพร้อมกัน)

- **สเปก Azure VM ที่แนะนำ:** อย่างน้อย **2 vCPU / 4 GB RAM** (เช่น `Standard_B2ms` หรือ `Standard_D2s_v5`) และกำหนด **Azure NSG Inbound Rules**:
  - เปิด **TCP `80`** (HTTP / ACME redirect) และ **TCP `443`** (Auto-HTTPS ผ่าน Caddy) สำหรับสาธารณะ/ห้องเรียน
  - จำกัด **TCP `22`** (SSH) เฉพาะ IP ของวิทยากร/ผู้ดูแลระบบเท่านั้น
- **คำสั่งติดตั้งบน Azure VM (Cloud Auto-HTTPS):**
  ```bash
  ./install.sh install --mode cloud --domain <your-azure-domain> --acme-email <your-email>
  ```
- **ตรวจสอบ Compose Configuration (Local & Cloud TLS Overlay):**
  ```bash
  docker compose config -q
  docker compose -f docker-compose.yml -f docker-compose.cloud.yml config -q
  ```
- **ค่าป้องกันพีคโหลด 60 คนพร้อมกัน (ตั้งค่าเริ่มต้นไว้ใน `.env.example` และ Compose แล้ว):**
  - `DB_MAX_OPEN_CONNS=50`, `DB_MAX_IDLE_CONNS=20` (ป้องกันการเต็ม 100 connections ของ PostgreSQL)
  - `SIMULATION_LLM_MAX_CONCURRENCY=8` (เมื่อผู้เรียนกดเริ่ม Scenario พร้อมกันเกิน 8 ช่อง ระบบสลับไปใช้ Static Pool ทันทีโดยบันทึก `fallback_reason="llm_concurrency_shed"` ไม่ค้างรอ)
  - Resource caps ที่ปรับ override ได้: `DB_CPU_LIMIT=2.0`, `DB_MEM_LIMIT=2g`, `BACKEND_CPU_LIMIT=2.0`, `BACKEND_MEM_LIMIT=1g`
- **Clean-Slate Pre-Class Database Verification (เช็กก่อนเปิดห้องเรียน 15 นาที):**
  ```bash
  # 1) ตรวจสอบจำนวนตารางหลักสูตรด้วย script มาตรฐาน (ใช้ได้ทั้ง Local และ Cloud):
  ./install.sh check-seed

  # 2) ตรวจสอบสถานะ Clean-Slate (โหลด .env เพื่อให้ตรงกับ POSTGRES_USER/POSTGRES_DB ที่ตั้งไว้):
  set -a; [ -f .env ] && . ./.env; set +a
  docker exec cyber_sim_db psql -U "${POSTGRES_USER:-admin}" -d "${POSTGRES_DB:-cyber_sim}" -c "
    SELECT
      (SELECT count(*) FROM users WHERE role = 'learner') AS learner_accounts,
      (SELECT count(*) FROM evaluation_metrics) AS evaluation_rows,
      (SELECT count(*) FROM exam_sessions) AS exam_sessions,
      (SELECT count(*) FROM simulation_sessions) AS sim_sessions,
      (SELECT count(*) FROM security_domains) AS domains,
      (SELECT count(*) FROM questions) AS questions,
      (SELECT count(*) FROM theory_subtopics) AS subtopics,
      (SELECT count(*) FROM theory_translations WHERE locale = 'th') AS th_theory_docs,
      (SELECT count(*) FROM scenarios WHERE status = 'active') AS active_scenarios;
  "

  # 3) กรณี Local Testing ใช้คำสั่งนี้เพื่อล้างข้อมูลฝึก/ผลทดสอบโดยคงหลักสูตรไว้ครบ:
  # ./install.sh reset-dev-data
  #
  # หมายเหตุสำหรับ Azure Cloud (--mode cloud): ./install.sh reset-dev-data ถูกบล็อกเพื่อความปลอดภัย;
  # แนะนำให้เปิดใช้งานจาก VM/Volume ใหม่ในวันจริง หรือหากเคยสร้างบัญชีทดสอบ dryrun_* บนเครื่อง staging ก่อนเปิดห้อง
  # ให้ลบตามลำดับ Foreign Key ใน transaction ดังนี้ (copy-paste ได้ทันที):
  # docker exec cyber_sim_db psql -U "${POSTGRES_USER:-admin}" -d "${POSTGRES_DB:-cyber_sim}" -v ON_ERROR_STOP=1 -c "
  #   BEGIN;
  #   CREATE TEMP TABLE _tu ON COMMIT DROP AS SELECT user_id FROM users WHERE role = 'learner' AND email LIKE 'dryrun\_%' ESCAPE '\';
  #   DELETE FROM feedback_logs WHERE session_id IN (SELECT session_id FROM simulation_sessions WHERE user_id IN (SELECT user_id FROM _tu));
  #   DELETE FROM session_actions WHERE session_id IN (SELECT session_id FROM simulation_sessions WHERE user_id IN (SELECT user_id FROM _tu));
  #   DELETE FROM scenario_snapshots WHERE session_id IN (SELECT session_id FROM simulation_sessions WHERE user_id IN (SELECT user_id FROM _tu));
  #   DELETE FROM help_click_analytics WHERE user_id IN (SELECT user_id FROM _tu) OR session_id IN (SELECT session_id FROM simulation_sessions WHERE user_id IN (SELECT user_id FROM _tu));
  #   DELETE FROM simulation_sessions WHERE user_id IN (SELECT user_id FROM _tu);
  #   DELETE FROM exam_session_questions WHERE session_id IN (SELECT session_id FROM exam_sessions WHERE user_id IN (SELECT user_id FROM _tu));
  #   DELETE FROM exam_sessions WHERE user_id IN (SELECT user_id FROM _tu);
  #   DELETE FROM user_skill_profiles WHERE user_id IN (SELECT user_id FROM _tu);
  #   DELETE FROM evaluation_metrics WHERE user_id IN (SELECT user_id FROM _tu);
  #   DELETE FROM users WHERE user_id IN (SELECT user_id FROM _tu);
  #   COMMIT;
  # "
  ```
  *เกณฑ์พร้อมเปิดห้อง:* `learner_accounts = 0`, `evaluation_rows = 0`, `exam_sessions = 0`, `sim_sessions = 0` ขณะที่ `domains = 5`, `questions = 540`, `subtopics = 28`, `th_theory_docs >= 33`, `active_scenarios >= 15`

---

### 3.2 ปัญหาหน้างานที่พบบ่อยและวิธีแก้ทันที

| สถานการณ์ที่เกิดขึ้น | สาเหตุทั่วไป | วิธีแก้หน้างานทันที (10–30 วินาที) |
| --- | --- | --- |
| **สมัครไม่ได้ / ขึ้น "Nickname taken"** | นักเรียนหลายคนตั้งชื่อเล่นซ้ำ เช่น `admin`, `hacker`, `test` | บอกนักเรียนเติมเลขประจำตัว 2–3 หลักต่อท้ายชื่อเล่น เช่น `alex_07`, `mook_14` |
| **สมัครพร้อมกัน 40–60 คนแล้วหน้า Register หมุนนาน** | พีคโหลด bcrypt + database registration พร้อมกันใน 30 วินาทีแรก | ให้รอข้อความยืนยัน 3–5 วินาที ห้ามกดซ้ำรัวๆ; หาก timeout ให้รีเฟรชแล้วลองล็อกอินก่อนสมัครใหม่ (บัญชีอาจสร้างสำเร็จแล้ว) |
| **นักเรียนลืมรหัสผ่านระหว่างเบรก / เปลี่ยนเครื่อง** | จำรหัสผ่านที่เพิ่งสมัครไม่ได้ | สำหรับเวิร์กช็อป 2 ชั่วโมง แนะนำให้นักเรียนใช้อีเมล+รหัสผ่านมาตรฐานเดียวกันชั่วคราว หรือสมัครบัญชีใหม่หากยังอยู่ในช่วง 10 นาทีแรก (แจ้งชื่อเล่นใหม่ให้ทีมผู้ช่วยบันทึกคู่รหัสผู้เรียนไว้) |
| **เบราว์เซอร์เด้ง / กดปิดแท็บระหว่างทำ Pre-Test หรือ Simulation** | แบตหมด เผลอรีเฟรช หรือเน็ต Wi-Fi หลุดชั่วคราว | **ไม่ต้องตกใจ!** ระบบบันทึกคำตอบ Pre-Test, หลักฐาน Log ที่เลือก และ Session ของ Simulation บน PostgreSQL ตลอดเวลา ให้เปิดเว็บและล็อกอินเดิม ระบบจะกู้คืนต่อจากจุดล่าสุดอัตโนมัติ |
| **สร้าง Simulation นานกว่าปกติ / ขึ้นป้าย "Static Scenario"** | LLM API ตอบช้าหรือไม่ได้ตั้งค่า `LLM_API_KEY` | ระบบสลับไปใช้ **Static Fallback Pool** โดยอัตโนมัติอย่างโปร่งใส ไม่ทำให้ห้องเรียนสะดุด — ย้ำนักเรียนว่าทำต่อไปได้เลย คะแนนความเชี่ยวชาญคำนวณเหมือนเดิม 100% |
| **นักเรียนหาปุ่ม Post-Test ไม่เจอ / กดเข้าไม่ได้** | ยังอ่าน `/learn` ไม่ครบ `5/5 Domain Essentials` หรือติดเกณฑ์ `Threshold 70%` | 1. ตรวจหน้า `/learn` ว่ามี badge **"จบสาระสำคัญโดเมนแล้ว"** ครบทั้ง 5 โดเมน<br>2. ถ้าเปิดใช้เกณฑ์ `Threshold 70%` ให้ดู `/evaluation` ว่าโดเมนไหนสีเหลือง/ต่ำกว่า 70 แล้วฝึก Simulation ของโดเมนนั้นต่อ |
| **ช่วง `1:58` กราฟ Spearman Figure 2 ยังล็อกอยู่** | มีผู้ส่งแบบสอบถาม (`/questionnaire`) สำเร็จยังไม่ถึง **3 คน (`n < 3`)** | ย้ำนักเรียนส่งหน้า `/questionnaire` ช่วง `1:50–1:55` ให้ครบ; ทันทีที่มีผู้ส่งครบ `n >= 3` กด Refresh บน `/admin/evaluations/statistics` กราฟ Scatter 3 แกนและค่าสถิติ Spearman จะปลดล็อกขึ้นโปรเจกเตอร์ทันที |
