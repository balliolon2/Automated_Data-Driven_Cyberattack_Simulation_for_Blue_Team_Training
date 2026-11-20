# Domain 2: ภัยคุกคาม ช่องโหว่ และการลดความเสี่ยง (Threats, Vulnerabilities, and Mitigations)

> **Domain ID:** `domain2` | **สัดส่วนในข้อสอบ:** 22% (7 จาก 30 ข้อใน Pre-Test / Post-Test) | **แหล่งอ้างอิง:** Wikipedia, CISA, OWASP, MITRE ATT&CK

บทเรียนนี้ครอบคลุมเวกเตอร์ภัยคุกคาม วิศวกรรมสังคม ช่องโหว่ซอฟต์แวร์/เว็บแอบพลิเคชัน ตระกูลมัลแวร์ และเทคนิคการลดความเสี่ยง โดยอ้างอิงประกาศเตือนภัยของหน่วยงานภาครัฐและเอกสารมาตรฐานเปิด อาทิ [CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing), [CISA #StopRansomware](https://www.cisa.gov/stopransomware), [OWASP Top 10:2021](https://owasp.org/Top10/) และบทความด้านความมั่นคงปลอดภัยไซเบอร์บน Wikipedia

---

## 1. ทฤษฎีหลัก CompTIA Security+ และแนวคิดการสอบ (Core Security+ Theory)

### 1.1 กลุ่มผู้คุกคามและแรงจูงใจ (Threat Actors & Adversary Motivations)
ผู้โจมตีที่ถูกวิเคราะห์ในข่าวกรองภัยคุกคาม ([MITRE ATT&CK®](https://attack.mitre.org/matrices/enterprise/), [CISA](https://www.cisa.gov/stopransomware)) แตกต่างกันด้านทรัพยากร การพรางตัว และเป้าหมาย:
- **Nation-State / APT (Advanced Persistent Threat):** กลุ่มที่ได้รับการสนับสนุนจากรัฐ มีเงินทุนสูง มุ่งเป้าขโมยทรัพย์สินทางปัญญา โครงสร้างพื้นฐานสำคัญ หรือจารกรรมข้อมูลด้วยปฏิบัติการแบบซ่อนตัวหลายขั้นตอน
- **Organized Crime / Ransomware Groups:** กลุ่มอาชญากรรมที่มีแรงจูงใจทางการเงิน ใช้ Ransomware และกลยุทธ์การกรรโชกซ้ำซ้อนแบบ Double Extortion ([CISA #StopRansomware](https://www.cisa.gov/stopransomware))
- **Hacktivists:** ผู้โจมตีที่มีแรงจูงใจทางอุดมการณ์หรือการเมือง นิยมใช้ DDoS, การเปลี่ยนหน้าเว็บไซต์ (Defacement) หรือการเผยแพร่ข้อมูลรั่วไหล
- **Insider Threats (ภัยคุกคามจากคนใน):** บุคลากรที่ได้รับอนุญาต (ทั้งแบบเจตนาร้ายหรือประมาทเลินเล่อ) ที่นำสิทธิ์ถูกต้องไปใช้ในทางมิชอบ ([Stanford Information Security](https://uit.stanford.edu/security))
- **Opportunistic / Unskilled Actors:** ผู้โจมตีที่อาศัยชุดเครื่องมือสำเร็จรูป สคริปต์สแกน หรือการโจมตีแบบ Credential Stuffing อัตโนมัติ

### 1.2 วิศวกรรมสังคมและเวกเตอร์ฟิชชิ่ง (Social Engineering & Phishing Vectors)
นิยามตาม [Wikipedia: Social Engineering](https://en.wikipedia.org/wiki/Social_engineering_%28security%29), [Wikipedia: Phishing](https://en.wikipedia.org/wiki/Phishing) และ [CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing):
- **Phishing / Spear Phishing / Whaling:** ข้อความหลอกลวงที่ปลอมเป็นหน่วยงานที่น่าเชื่อถือเพื่อขโมย credentials หรือหลอกให้ดาวน์โหลดไฟล์อันตราย (แบบหว่านวงกว้าง แบบเจาะจงรายบุคคล หรือแบบมุ่งเป้าผู้บริหารระดับสูง)
- **Vishing & Smishing:** วิศวกรรมสังคมผ่านโทรศัพท์เสียง (Voice) และข้อความ SMS
- **Business Email Compromise (BEC):** การปลอมแปลงหรือยึดครองบัญชีอีเมลผู้บริหาร/คู่ค้าเพื่อเบี่ยงเบนการโอนเงินหรือข้อมูลสำคัญ
- **Pretexting, Typosquatting & Watering Hole:** การสร้างสถานการณ์เร่งด่วนหลอกลวง การจดโดเมนคล้ายคลึง (เช่น `corp-login[.]com`) หรือการวางกับดักบนเว็บไซต์ภายนอกที่กลุ่มเป้าหมายเข้าใช้งานเป็นประจำ

### 1.3 ช่องโหว่ซอฟต์แวร์และเว็บแอปพลิเคชัน (OWASP Top 10:2021)
จำแนกตามมาตรฐาน [OWASP Top 10:2021](https://owasp.org/Top10/) ร่วมกับ [Wikipedia: SQL Injection](https://en.wikipedia.org/wiki/SQL_injection) และ [Wikipedia: Cross-site Scripting (XSS)](https://en.wikipedia.org/wiki/Cross-site_scripting):
- **A01 Broken Access Control:** การขาดการตรวจสอบสิทธิ์ที่รัดกุม เปิดช่องให้เกิด IDOR (Insecure Direct Object Reference), Privilege Escalation หรือ Directory Traversal (`../../etc/passwd`)
- **A03 Injection (SQLi, Command Injection, XSS):** อินพุตที่ไม่น่าเชื่อถือถูกตีความเป็นคำสั่ง SQL (`' OR 1=1--`), คำสั่ง OS shell หรือสคริปต์ฝั่งเบราว์เซอร์ (Stored/Reflected/DOM XSS)
- **A06 Vulnerable & Outdated Components:** ไลบรารีหรืออุปกรณ์เชื่อมต่ออินเทอร์เน็ตที่ยังไม่ได้อัปเดตแพตช์และมีช่องโหว่ CVE ที่เปิดเผยแล้ว
- **A07 Identification & Authentication Failures:** การโจมตีด้วย Credential Stuffing, Brute Force, Session Token ที่คาดเดาได้ง่าย และการขาดระบบ MFA
- **A10 Server-Side Request Forgery (SSRF):** การบีบบังคับให้เซิร์ฟเวอร์ดึงข้อมูลเมตาภายใน (`169.254.169.254`) หรือเรียก Internal APIs โดยไม่ได้รับอนุญาต

### 1.4 มัลแวร์และกลไกการสร้างผลกระทบ (Malware & Impact Mechanisms)
อธิบายไว้ใน [Wikipedia: Malware](https://en.wikipedia.org/wiki/Malware) และ [Wikipedia: Ransomware](https://en.wikipedia.org/wiki/Ransomware):
- **Ransomware:** เข้ารหัสไฟล์บนเครื่องและเครือข่าย พร้อมลบ Volume Shadow Copies เพื่อทำลายความพร้อมใช้งาน (Availability) และเรียกค่าไถ่ ([CISA #StopRansomware](https://www.cisa.gov/stopransomware))
- **Trojans, RATs & Rootkits:** ซอฟต์แวร์อันตรายที่แฝงตัวเป็นโปรแกรมปกติเพื่อเปิดช่องทาง C2 หรือซ่อนหลักฐานในระดับ Kernel/User-Mode
- **Living-off-the-Land (LotL) / Fileless Execution:** สั่งทำงานโค้ดอันตรายในหน่วยความจำโดยอาศัยเครื่องมือมาตรฐานของระบบปฏิบัติการ (`powershell.exe`, `wmic.exe`, `certutil.exe`)

### 1.5 เทคนิคการบรรเทาและแก้ไขช่องโหว่ (Vulnerability Mitigation Techniques)
- **Parameterized Queries & Input Validation:** แยกข้อมูลผู้ใช้จากไวยากรณ์คำสั่ง SQL/OS เพื่อป้องกัน Injection ([OWASP Top 10](https://owasp.org/Top10/))
- **Least Privilege & Network Segmentation:** จำกัดวงความเสียหาย (Blast Radius) หาก Endpoint หรือ Web Service ถูกบุกรุก
- **Patch Management & Virtual Patching (WAF/IPS):** อุดช่องโหว่ CVE ที่ทราบแล้ว หรือใช้ WAF/IPS บล็อกรูปแบบการโจมตีที่ขอบเครือข่ายระหว่างทดสอบแพตช์
- **Email Authentication & EDR:** บังคับใช้ SPF/DKIM/DMARC ([CISA](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)) และตรวจจับพฤติกรรมโพรเซสบนเครื่องปลายทาง

---

## 2. ข้อมูล SOC Telemetry, Log Indicators และการเชื่อมโยง MITRE ATT&CK

### 2.1 แหล่ง Log สำคัญและตัวบ่งชี้การบุกรุก (IoCs)
- **Web / WAF / Reverse Proxy Logs:** เพย์โหลด URI ที่มีคีย์เวิร์ด SQL (`UNION SELECT`), Path Traversal (`../`), Shell Metacharacters หรือลายเซ็นเครื่องมือสแกน (`sqlmap`, `nikto`)
- **Email Gateway Logs:** ผลการตรวจ SPF/DKIM/DMARC ที่ไม่ผ่าน โดเมนผู้ส่งที่เพิ่งจดทะเบียนใหม่ และไฟล์แนบที่เปิดใช้ Macro (`.xlsm`, `.iso`, `.lnk`)
- **EDR / Process Creation Logs (Sysmon EventID 1):** ความสัมพันธ์ของโพรเซสแม่-ลูกที่ผิดปกติ เช่น `WINWORD.EXE -> powershell.exe -enc ...` หรือ `w3wp.exe -> cmd.exe`

### 2.2 รูปแบบการสืบค้น KQL ที่พบบ่อย

ไวยากรณ์สืบค้นอ้างอิงตาม [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/):

```kql
source == "waf" and severity == "high"
message contains "powershell" or message contains "SELECT"
source == "email" and message contains "attachment"
```

### 2.3 การเชื่อมโยง MITRE ATT&CK

รหัสเทคนิคทั้งหมดเชื่อมโยงโดยตรงไปยัง [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/):

| Technique ID | ชื่อเทคนิค (Technique Name) | สัญญาณ Telemetry ที่สังเกตได้ (Observable Indicator) |
|---|---|---|
| [T1566](https://attack.mitre.org/techniques/T1566/) | Phishing | การแจ้งเตือนจาก Email Gateway พบการปลอมชื่อผู้ส่ง ไม่ผ่าน DMARC หรือมีผู้ใช้คลิกลิงก์อันตรายก่อนเกิดการรันคำสั่งบนเครื่องปลายทาง |
| [T1190](https://attack.mitre.org/techniques/T1190/) | Exploit Public-Facing Application | Web server แตกโพรเซส Shell (`nginx`/`httpd`/`w3wp.exe` -> `sh`/`cmd.exe`) หรือตอบ HTTP 500/200 พร้อม URI แบบ SQLi/Webshell |
| [T1059](https://attack.mitre.org/techniques/T1059/) | Command and Scripting Interpreter | การรัน PowerShell แบบเข้ารหัส (`-EncodedCommand`), คำสั่ง bash ที่ถูกอำพราง หรือตัวแปลสคริปต์ที่ไม่คาดคิด |
| [T1486](https://attack.mitre.org/techniques/T1486/) | Data Encrypted for Impact | เหตุการณ์เปลี่ยนชื่อ/เข้ารหัสไฟล์ปริมาณมาก การลบ shadow copy (`vssadmin delete shadows`) และการสร้างไฟล์เรียกค่าไถ่ |

---

## 3. การคัดแยกเหตุการณ์ (TP vs. FP) และแนวทางปฏิบัติ 3 ขั้นตอนตาม Playbook

### 3.1 การจำแนก True Positive (TP) ออกจาก False Positive (FP)
- **ตัวบ่งชี้ True Positive:** การสร้าง Web shell หลังส่ง HTTP POST ไปยังจุดอัปโหลดไฟล์; โปรแกรม Microsoft Office เรียก `powershell.exe` แล้วติดต่อไปยัง C2 IP ภายนอก; หรือการส่งคำสั่ง SQL Injection ตามด้วยปริมาณข้อมูลฐานข้อมูลขนาดใหญ่ส่งออกนอกระบบ
- **ตัวบ่งชี้ False Positive (Benign Context Logs):**
  - **Authorized Vulnerability Scanner:** การสแกนพอร์ตความถี่สูงหรือการส่งเพย์โหลดทดสอบ SQLi/XSS ที่มาจาก IP ของเครื่องสแกนภายในองค์กรที่ได้รับอนุมัติตามตารางเวลาใน Log
  - **Penetration Test Window:** ความพยายามเจาะระบบที่ตรงกับบันทึกการอนุมัติทดสอบเจาะระบบที่ลงนามแล้ว ทั้งหมายเลข IP ต้นทางและช่วงเวลาที่กำหนด
  - **Legitimate Admin Script:** สคริปต์อัตโนมัติของฝ่าย IT ที่ลงนามดิจิทัลอย่างถูกต้องและรันในช่วง Maintenance Window โดยไม่มีการติดต่อ C2 ภายนอก

### 3.2 หลักการตอบสนอง 3 ขั้นตอนตาม Playbook
อ้างอิงตาม [NIST SP 800-61 Rev. 2](https://csrc.nist.gov/pubs/sp/800/61/r2/final), [NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final) และ [CISA #StopRansomware](https://www.cisa.gov/stopransomware):
- **Phase 1 — Containment (การจำกัดวงความเสียหาย):**
  - *สิ่งที่ควรทำ:* แยกเครื่องที่ติดเชื้อผ่าน EDR บล็อก IP/โดเมน C2 ของผู้โจมตีที่ Firewall/Proxy กักกันอีเมลอันตรายทุกกล่องข้อความ หรือนำ Web Endpoint ที่มีช่องโหว่ไว้หลังกฎ WAF ที่เข้มงวด
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การลบฐานข้อมูลเว็บแอปพลิเคชันทั้งหมด การปิดอินเทอร์เน็ตเกตเวย์ทั้งองค์กร หรือการเพิกเฉยต่อ Web shell ที่กำลังทำงานอยู่
- **Phase 2 — Eradication (การกำจัดภัยคุกคาม):**
  - *สิ่งที่ควรทำ:* ยุติโพรเซสอันตราย ลบ Web shells และกลไก Persistence (Registry Run keys, cron jobs, scheduled tasks) ติดตั้งแพตช์สำหรับ CVE ที่ถูกเจาะ และหมุนเวียน Service Credentials ที่รั่วไหล
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การรีสตาร์ทเครื่องเพียงอย่างเดียว (เพราะ Persistence ยังคงอยู่หลังบูตใหม่) หรือการปล่อยซอฟต์แวร์เวอร์ชันมีช่องโหว่ไว้เช่นเดิม
- **Phase 3 — Recovery (การฟื้นฟูระบบ):**
  - *สิ่งที่ควรทำ:* กู้คืนไฟล์ที่ถูกแก้ไขจากสำรองข้อมูลสะอาดที่ผ่านการตรวจสอบ ทดสอบความสมบูรณ์ของแอปพลิเคชัน เปิดการเชื่อมต่อเครือข่ายกลับคืนอย่างเป็นขั้นตอน และเฝ้าระวัง SIEM/EDR เพื่อป้องกันการติดเชื้อซ้ำ
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การนำ Backup ที่ยังไม่ได้อัปเดตแพตช์กลับมาให้บริการจริงโดยไม่ได้ปิดช่องทางเข้าเริ่มต้น

---

## 4. บทสรุปสำคัญและแหล่งอ้างอิง (Key Takeaways & References)

- **เชื่อมโยงข้ามชั้นข้อมูล (Correlate Across Layers):** ในสถานการณ์ Domain 2 ควรเชื่อมโยงเวกเตอร์เริ่มต้น (เช่น WAF HTTP payload หรืออีเมลฟิชชิ่ง) เข้ากับ Telemetry หลังการเจาะระบบ (เช่น Shell ที่ถูกสร้างขึ้นหรือการเชื่อมต่อ C2 ขาออก)
- **ตรวจหลักฐานความถูกต้องก่อนเสมอ:** การสแกนช่องโหว่อัตโนมัติและ Pentest ที่ได้รับอนุญาตจะมีหน้าตาใน Firewall/WAF Logs เหมือนการโจมตีจริง ต้องตรวจสอบบันทึกอนุญาตหรือ Change Record ก่อนตัดสินเป็น True Positive เสมอ
- **กำจัด Persistence ให้สิ้นซาก:** Containment หยุดความเสียหายที่กำลังเกิด แต่ Eradication ต้องกำจัดช่องโหว่ต้นเหตุและกลไกฝังตัวทั้งหมดให้เสร็จสมบูรณ์ก่อน Recovery

### แหล่งอ้างอิงเว็บที่น่าเชื่อถือ (Authoritative Web References)
- **Wikipedia (แหล่งความรู้เปิด):**
  - [Phishing](https://en.wikipedia.org/wiki/Phishing) & [Social Engineering (Security)](https://en.wikipedia.org/wiki/Social_engineering_%28security%29)
  - [Malware](https://en.wikipedia.org/wiki/Malware) & [Ransomware](https://en.wikipedia.org/wiki/Ransomware)
  - [SQL Injection](https://en.wikipedia.org/wiki/SQL_injection) & [Cross-site Scripting (XSS)](https://en.wikipedia.org/wiki/Cross-site_scripting)
- **หน่วยงานภาครัฐและมาตรฐาน (CISA & NIST):**
  - [CISA: Recognize and Mitigate Phishing](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)
  - [CISA #StopRansomware Resource Center](https://www.cisa.gov/stopransomware)
  - [NIST SP 800-61 Rev. 2: Computer Security Incident Handling Guide](https://csrc.nist.gov/pubs/sp/800/61/r2/final) & [NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final)
- **ฐานความรู้เปิดและมหาวิทยาลัย:**
  - [OWASP Top 10:2021 Web Application Security Risks](https://owasp.org/Top10/)
  - [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/) (`T1566`, `T1190`, `T1059`, `T1486`)
  - [Stanford University Information Security (SISA)](https://uit.stanford.edu/security)
  - [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/)
