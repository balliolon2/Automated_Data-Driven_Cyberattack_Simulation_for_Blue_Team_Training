# Domain 4: ปฏิบัติการด้านความปลอดภัย (Security Operations)

> **Domain ID:** `domain4` | **สัดส่วนในข้อสอบ:** 28% (8 จาก 30 ข้อใน Pre-Test / Post-Test) | **แหล่งอ้างอิง:** Wikipedia, NIST CSRC, CMU ISO, Stanford UIT, MITRE ATT&CK

Domain 4 คือแกนหลักของการปฏิบัติงานในศูนย์ SOC ระดับ Tier-1/Tier-2 ได้แก่ การเฝ้าระวังความปลอดภัย การเชื่อมโยง Telemetry ในระบบ SIEM/SOAR ปฏิบัติการจัดการช่องโหว่ นิติวิทยาศาสตร์ดิจิทัล (Digital Forensics) และการรับมือเหตุการณ์ตั้งแต่ต้นจนจบตามแนวทาง [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final), [NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final), [CMU Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html) และแหล่งข้อมูล Wikipedia

---

## 1. ทฤษฎีหลัก CompTIA Security+ และแนวคิดการสอบ (Core Security+ Theory)

### 1.1 การเฝ้าระวังความปลอดภัย SIEM และ SOAR
อธิบายไว้ใน [Wikipedia: Security Operations Center (SOC)](https://en.wikipedia.org/wiki/Security_operations_center) และ [Wikipedia: Security Information and Event Management (SIEM)](https://en.wikipedia.org/wiki/Security_information_and_event_management):
- **SIEM (Security Information and Event Management):** ผสานการเฝ้าระวังเหตุการณ์แบบเรียลไทม์ (SEM) เข้ากับการจัดเก็บและเชื่อมโยง Log ย้อนหลัง (SIM) ข้ามเครื่องปลายทาง Firewalls ระบบยืนยันตัวตน และแพลตฟอร์มคลาวด์ เพื่อตรวจจับภัยคุกคามและสนับสนุน Threat Hunting
- **SOAR (Security Orchestration, Automation, and Response):** เชื่อมโยงการแจ้งเตือนจาก SIEM เข้ากับ Playbook อัตโนมัติที่ทำซ้ำได้ เพื่อเสริมข้อมูลบริบท (Threat Intelligence, GeoIP) และจำกัดวงความเสียหายอย่างมีขอบเขต
- **การรวม Log และการซิงโครไนซ์เวลา:** ตัวรวบรวม Log ส่วนกลาง (Syslog/Agents) ที่เทียบเวลาตรงกันผ่าน Network Time Protocol (NTP) ช่วยรักษาลำดับเหตุการณ์ข้ามระบบได้อย่างแม่นยำระหว่างการสืบสวนย้อนหลัง

### 1.2 ปฏิบัติการด้านความปลอดภัยของ Endpoint เครือข่าย และอีเมล
อธิบายไว้ใน [Wikipedia: Intrusion Detection System](https://en.wikipedia.org/wiki/Intrusion_detection_system):
- **EDR / XDR:** เฝ้าระวังแผนผังโพรเซส การเข้าถึง Credential การแก้ไข Registry และ Network Sockets อย่างต่อเนื่อง พร้อมสั่งแยกเฉพาะ Endpoint ที่ถูกบุกรุกได้ทันที
- **IDS vs. IPS:** Intrusion Detection Systems ตรวจสอบทราฟฟิกแบบแฝงตัวและแจ้งเตือนตามลายเซ็นหรือความผิดปกติ; ส่วน Intrusion Prevention Systems วางตัวขวางเส้นทาง (Inline) เพื่อสกัดกั้นทราฟฟิกอันตรายได้ทันที
- **การกรอง DNS และการยืนยันตัวตนอีเมล:** ทำ Sinkhole สำหรับโดเมน C2 อันตราย และตรวจสอบระเบียน SPF, DKIM และ DMARC เพื่อลดการปลอมแปลงอีเมล ([CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing))

### 1.3 การจัดการช่องโหว่และความแม่นยำในการตรวจจับ
- **การสแกนและการจัดลำดับความสำคัญ:** การสแกนช่องโหว่แบบใช้สิทธิ์ยืนยันตัวตน (Credentialed) เทียบกับไม่ใช้สิทธิ์ (Non-credentialed), คะแนนความรุนแรง CVSS และการคัดแยกตามความสำคัญของ Asset ([NIST CSRC Glossary](https://csrc.nist.gov/glossary))
- **False Positives vs. False Negatives:** *False Positive* คือการแจ้งเตือนว่ากิจกรรมปกติหรือเครื่องที่แพตช์แล้วเป็นภัยคุกคาม/มีช่องโหว่; *False Negative* คือการตรวจไม่พบการโจมตีหรือช่องโหว่ที่มีอยู่จริง (ถือเป็นความเสี่ยงเชิงปฏิบัติการสูงสุด)
- **การแก้ไขและทวนสอบ:** การติดตั้งแพตช์อย่างเป็นระบบ การวางมาตรการชดเชยช่วง Patch Freeze และการสแกนทวนสอบหลังแก้ไข

### 1.4 วงจรการตอบสนองต่อเหตุการณ์ (Incident Response Lifecycle)
สอดคล้องกับ [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final), [NIST SP 800-61 Rev. 3 / CSF 2.0 Profile](https://csrc.nist.gov/pubs/sp/800/61/r3/final), [Wikipedia: Computer Security Incident Management](https://en.wikipedia.org/wiki/Computer_security_incident_management) และ [CMU Computer Security Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html):
1. **Preparation (การเตรียมความพร้อม):** จัดทำ Response Playbooks ติดตั้งเซ็นเซอร์ SIEM/EDR ตั้งค่าการส่ง Log ส่วนกลาง และซ้อมแผน Tabletop Drills ([Stanford University IT](https://uit.stanford.edu/security))
2. **Detection and Analysis (การตรวจจับและวิเคราะห์):** คัดแยกแจ้งเตือน สืบค้น Telemetry Logs ประเมินขอบเขต ตัดสิน True vs. False Positive และเก็บรวบรวม Evidence Logs
3. **Containment, Eradication, and Recovery (การจำกัดวง กำจัด และฟื้นฟู):** แยกระบบที่ได้รับผลกระทบ กำจัดกลไก Persistence/สาเหตุรากเหง้า และคืนระบบที่สะอาดสู่การผลิต
4. **Post-Incident Activity (ถอดบทเรียน):** ทบทวนหลังเกิดเหตุอย่างเป็นระบบ ปรับจูนกฎการตรวจจับที่ส่งเสียงรบกวนมากเกินไป และอัปเดตมาตรการความปลอดภัย

### 1.5 นิติวิทยาศาสตร์ดิจิทัลและการรักษาพยานหลักฐาน (Digital Forensics & Evidence Preservation)
อธิบายใน [Wikipedia: Digital Forensics](https://en.wikipedia.org/wiki/Digital_forensics):
- **Order of Volatility (RFC 3227):** จัดเก็บหลักฐานที่สูญหายง่ายที่สุดก่อนตามลำดับ: CPU registers/cache → RAM/memory dump → โพรเซสที่ทำงานและซ็อกเก็ตเครือข่าย → ไฟล์ชั่วคราว → Disk Images → ข้อมูลสำรองระยะยาว
- **Legal Hold & Chain of Custody:** บันทึกผู้จัดเก็บ ผู้ส่งต่อ และผลตรวจความถูกต้องด้วย Cryptographic Hash `SHA-256` เพื่อรักษาความน่าเชื่อถือทางกฎหมาย

---

## 2. ข้อมูล SOC Telemetry, Log Indicators และการเชื่อมโยง MITRE ATT&CK

### 2.1 การอ่าน Telemetry Logs ในหน้า Detect & Analyze
โดยทั่วไป Log ในระบบจำลอง SOC มีฟิลด์มาตรฐาน ได้แก่ `timestamp`, `source`, `host`, `severity`, `src_ip`, `dst_ip`, `user` และ `message` นักวิเคราะห์จะไล่เรียงห่วงโซ่การโจมตีผ่าน 4 ขั้นตอน:
1. **Initial Signal:** การแจ้งเตือนระดับสูงจาก EDR/IDS/Auth ที่ระบุ `host` หรือ `user` เป้าหมาย
2. **Pivot & Correlate:** กรองตาม `host` หรือ `src_ip` รอบเวลาที่แจ้งเตือนเพื่อดูเหตุการณ์ก่อนหน้าและหลังจากนั้นทันที
3. **Confirm Scope:** มองหาร่องรอย Lateral Movement, การสร้าง Persistence หรือการลักลอบส่งออกข้อมูล
4. **Check Context:** ตรวจสอบคำอธิบายกิจกรรมปกติ (Change Tickets, งานตามกำหนด, เครื่องมือสแกนที่ได้รับอนุญาต) ก่อนสรุปผล

### 2.2 รูปแบบการสืบค้น KQL ที่พบบ่อย

ไวยากรณ์สืบค้นอ้างอิงตาม [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/):

```kql
severity == "high" or severity == "critical"
host == "WS-FIN-04"
source == "edr" and message contains "lsass"
src_ip == "185.220.101.42" or dst_ip == "185.220.101.42"
```

### 2.3 การเชื่อมโยง MITRE ATT&CK

รหัสเทคนิคทั้งหมดเชื่อมโยงโดยตรงไปยัง [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/):

| Technique ID | ชื่อเทคนิค (Technique Name) | สัญญาณ Telemetry ที่สังเกตได้ (Observable Indicator) |
|---|---|---|
| [T1003](https://attack.mitre.org/techniques/T1003/) | OS Credential Dumping | แจ้งเตือน EDR พบ `procdump`, `mimikatz` หรือการขออ่านหน่วยความจำของโพรเซส `lsass.exe` โดยไม่ได้รับอนุญาต |
| [T1053](https://attack.mitre.org/techniques/T1053/) | Scheduled Task/Job | `EventID 4698` หรือการแก้ไข `crontab` บน Linux เพื่อสั่งรันไฟล์ตามกำหนดเวลา |
| [T1110](https://attack.mitre.org/techniques/T1110/) | Brute Force | เหตุการณ์ล็อกอินล้มเหลว `EventID 4625` ถี่ผิดปกติจาก IP ภายนอกเดียว ตามด้วยการล็อกอินสำเร็จ `4624` |
| [T1071](https://attack.mitre.org/techniques/T1071/) | Application Layer Protocol | การส่งสัญญาณ Beacon ผ่าน HTTPS/DNS เป็นรอบเวลาคงที่จาก Endpoint ไปยังโดเมนภายนอกที่ไม่รู้จัก |

---

## 3. การคัดแยกเหตุการณ์ (TP vs. FP) และแนวทางปฏิบัติ 3 ขั้นตอนตาม Playbook

### 3.1 การจำแนก True Positive (TP) ออกจาก False Positive (FP)
- **ตัวบ่งชี้ True Positive:** การเข้าถึงหน่วยความจำ LSASS ตามด้วยการพิสูจน์ตัวตน SMB ข้ามเครื่อง; Scheduled Task ใหม่ที่ชี้ไปยัง `%TEMP%` หรือ `/tmp`; หรือการล็อกอินล้มเหลวจำนวนมากที่จบลงด้วยการล็อกอินสำเร็จและยกระดับสิทธิ์
- **ตัวบ่งชี้ False Positive (Benign Context Logs):**
  - **Approved IT Change / Patching:** การรัน PowerShell สิทธิ์สูงหรือสร้างบริการใหม่ที่มาพร้อมบันทึกจาก SCCM/Intune หรือใบ Change Ticket ที่อนุมัติ
  - **Authorized Endpoint Maintenance:** การสแกนไวรัสเต็มระบบหรือโปรแกรม Backup ที่อ่านไฟล์ระบบในช่วง Maintenance Window ที่กำหนด
  - **Benign Operational Anomaly:** สคริปต์ Build ของนักพัฒนาหรือโพรบเฝ้าระวังที่มีเอกสารอ้างอิงใน CMDB/ระบบ Ticketing

### 3.2 หลักการตอบสนอง 3 ขั้นตอนตาม Playbook
อ้างอิงตาม [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final), [CMU Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html) และ [Stanford Incident Reporting](https://uit.stanford.edu/security/report-incident):
- **Phase 1 — Containment (การจำกัดวงความเสียหาย):**
  - *สิ่งที่ควรทำ:* สั่งตัดการเชื่อมต่อเครือข่ายของ Endpoint ผ่าน EDR (โดยคงข้อมูลในหน่วยความจำ RAM ไว้) บล็อก IP/โดเมน C2 ที่ยืนยันแล้วที่ Firewall และระงับบัญชีที่ถูกบุกรุก
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การปิดเครื่องเซิร์ฟเวอร์ทันทีก่อนเก็บข้อมูลหน่วยความจำ (ทำลายหลักฐาน volatile RAM) หรือการปิดกฎ SIEM เพื่อล้างคิวแจ้งเตือน
- **Phase 2 — Eradication (การกำจัดภัยคุกคาม):**
  - *สิ่งที่ควรทำ:* ยุติโพรเซสอันตราย ลบ Scheduled Tasks/Services/ไฟล์ไบนารีแปลกปลอม เพิกถอน Kerberos Tickets/Sessions และเปลี่ยนรหัสผ่านที่ได้รับผลกระทบ
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การลบ Log Files ระหว่างการทำความสะอาด (ทำลายหลักฐานทางนิติวิทยาศาสตร์) หรือปล่อยให้กลไก Persistence ยังอยู่
- **Phase 3 — Recovery (การฟื้นฟูระบบ):**
  - *สิ่งที่ควรทำ:* สร้างเครื่องใหม่หรือกู้คืนจากอิมเมจที่สะอาด ตรวจสอบสุขภาพ EDR และการส่ง Log พร้อมเฝ้าระวังบัญชีและเครื่องเป้าหมายใน SIEM อย่างเข้มข้น
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การนำเครื่องที่ยังไม่ยืนยันความสะอาดกลับเข้าเครือข่ายการผลิตทันที

---

## 4. บทสรุปสำคัญและแหล่งอ้างอิง (Key Takeaways & References)

- **ค้นเชื่อมโยง อย่าเดา (Pivot, Don't Guess):** อย่าเลือกแถว Log เพียงเพราะ `severity == "high"` จงตรวจสอบว่าแถวใดพิสูจน์ Key Finding ได้จริง เพราะทุก Noise Log ที่เลือกผิดจะถูกหัก `-2` คะแนน
- **รักษาหลักฐานที่สูญหายง่าย (Preserve Volatility):** ในขั้น Containment ให้เลือก *การแยกเครือข่าย (Network Isolation)* แทน *การปิดไฟทันที (Hard Power-Off)* เพื่อรักษาหลักฐานใน RAM
- **ตรวจสอบคำอธิบายฝั่ง False Positive:** ค้นหา Benign Context Logs (Change Tickets, แผนซ่อมบำรุง, เครื่องมือที่ได้รับอนุญาต) ทุกครั้งที่พฤติกรรมอาจเป็นกิจกรรมปฏิบัติการปกติ

### แหล่งอ้างอิงเว็บที่น่าเชื่อถือ (Authoritative Web References)
- **Wikipedia (แหล่งความรู้เปิด):**
  - [Security Operations Center (SOC)](https://en.wikipedia.org/wiki/Security_operations_center)
  - [Security Information and Event Management (SIEM)](https://en.wikipedia.org/wiki/Security_information_and_event_management)
  - [Intrusion Detection System](https://en.wikipedia.org/wiki/Intrusion_detection_system)
  - [Computer Security Incident Management](https://en.wikipedia.org/wiki/Computer_security_incident_management) & [Digital Forensics](https://en.wikipedia.org/wiki/Digital_forensics)
- **หน่วยงานภาครัฐและมาตรฐาน (NIST & CISA):**
  - [NIST SP 800-61 Rev. 2: Computer Security Incident Handling Guide](https://csrc.nist.gov/pubs/sp/800/61/r2/final)
  - [NIST SP 800-61 Rev. 3: Incident Response & CSF 2.0 Profile](https://csrc.nist.gov/pubs/sp/800/61/r3/final)
  - [CISA Phishing & Threat Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)
- **มหาวิทยาลัยและฐานความรู้เปิด:**
  - [CMU Information Security Office: Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)
  - [Stanford University IT: Report an Incident & SISA Portal](https://uit.stanford.edu/security/report-incident)
  - [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/) (`T1003`, `T1053`, `T1110`, `T1071`)
  - [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/)
