# Domain 2: ภัยคุกคาม ช่องโหว่ และการลดความเสี่ยง (สรุปกระชับ 3 นาที)

> **Domain 2** | **สัดส่วนในข้อสอบ:** 22% (`7 / 30` ข้อใน Pre/Post-Test) | **เวลาอ่านโดยประมาณ:** ~3 นาที

ถ้า Domain 1 คือแม่กุญแจหน้าบ้าน Domain 2 ก็คือการรู้ทันว่า *"ใคร"* กำลังแอบงัดประตูและ *"ใช้กลโกงแบบไหน"* — ตั้งแต่ข้อความหลอกโอนเงิน ช่องโหว่เว็บ ไปจนถึงมัลแวร์เรียกค่าไถ่

---

## 1. ทฤษฎีหลัก CompTIA Security+ ที่ออกสอบบ่อย (Core Security+ Theory)

- **กลุ่มผู้คุกคาม (Threat Actors) & แรงจูงใจ:**
  - *Nation-State (APT):* ทีมระดับรัฐที่มีทุนหนา เน้นจารกรรมข้อมูลระยะยาวอย่างเงียบเชียบ
  - *อาชญากรไซเบอร์ / Ransomware:* ล่าเงินค่าไถ่ด้วยการล็อกไฟล์และขู่ปล่อยข้อมูล ([CISA #StopRansomware](https://www.cisa.gov/stopransomware))
  - *Hacktivists / Insider Threat / ผู้ใช้อาวุธสำเร็จรูป:* โจมตีตามอุดมการณ์ คนในที่ทำข้อมูลรั่ว (ตั้งใจหรือเผลอ) หรือใช้สคริปต์สำเร็จรูปสแกนช่องโหว่ทั่วไป
- **วิศวกรรมสังคม & Phishing** ([Wikipedia: Phishing](https://en.wikipedia.org/wiki/Phishing), [CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)):
  - หลอกคนแทนการแฮกซอฟต์แวร์: *Phishing* (ส่งหว่าน), *Spear Phishing* (เจาะจงตัว), *Whaling* (เล็งผู้บริหาร/BEC หลอกโอนเงิน), *Smishing* (SMS), *Vishing* (โทรหลอก), และ *Typosquatting* (ปลอมชื่อเว็บคล้ายของจริง เช่น `sch00l-login.org`)
  - **เกราะป้องกันอีเมล:** *SPF* ตรวจ IP ผู้ส่ง, *DKIM* ตรวจลายเซ็นดิจิทัล, *DMARC* สั่งกักหรือตีกลับเมลปลอม
- **ช่องโหว่เว็บยอดฮิต OWASP & การป้องกัน** ([OWASP Top 10:2021](https://owasp.org/Top10/)):
  - *Broken Access Control:* ไม่ตรวจสิทธิ์จนเปิดดูข้อมูลคนอื่น (IDOR) หรือไต่โฟลเดอร์ (`../../`) ได้
  - *Injection ([SQL Injection](https://en.wikipedia.org/wiki/SQL_injection) / XSS / Command Injection):* แทรกคำสั่งปลอมในช่องกรอกข้อมูล (`' OR 1=1--`) ป้องกันขาดด้วย **Parameterized Queries** และคัดกรองอินพุต
  - *ซอฟต์แวร์เก่าที่ไม่อัปเดต & Auth อ่อนแอ:* รีบแพตช์และใช้ **WAF** ช่วยกรองระหว่างรอทดสอบแพตช์
- **มัลแวร์ & Living-off-the-Land (LotL)** ([Wikipedia: Ransomware](https://en.wikipedia.org/wiki/Ransomware)):
  - *Ransomware* เข้ารหัสไฟล์แล้วลบ Backup (`vssadmin delete shadows`) ส่วน *LotL* แอบใช้เครื่องมือประจำเครื่องอย่าง `powershell.exe` เพื่อพรางตัวไม่ให้จับได้ง่าย

---

## 2. สัญญาณอันตรายใน SIEM Log, KQL และ MITRE ATT&CK

จับตาสัญญาณฟิชชิง ([T1566](https://attack.mitre.org/techniques/T1566/)), เจาะเว็บ ([T1190](https://attack.mitre.org/techniques/T1190/)), รันคำสั่ง PowerShell แปลกปลอม ([T1059](https://attack.mitre.org/techniques/T1059/)) และเข้ารหัสไฟล์จำนวนมาก ([T1486](https://attack.mitre.org/techniques/T1486/)) ตามกรอบ [MITRE ATT&CK®](https://attack.mitre.org/matrices/enterprise/):

```kql
(source == "waf" and severity == "high") or (source == "edr" and message contains "powershell")
```

- **จุดสังเกต:** `WINWORD.EXE -> powershell.exe -EncodedCommand`, เว็บเซิร์ฟเวอร์แตกโปรเซส `cmd.exe` หรือไฟล์ถูกเปลี่ยนนามสกุลรัวๆ

---

## 3. การคัดแยก True Positive vs. False Positive และ Playbook 3 ระยะ

| True Positive (ภัยคุกคามจริง) | False Positive (งานปกติ / มี Benign Context Log ยืนยัน) |
|---|---|
| `w3wp.exe` สั่งเปิด shell และวาง webshell หลังมีคำขอ SQLi | การสแกน SQLi/Port จาก **IP เครื่อง Scanner ที่ได้รับอนุญาต** ตามตารางเวลา |
| ไฟล์ Office รัน PowerShell ติดต่อไปยัง C2 IP ภายนอก | สคริปต์บำรุงรักษาระบบที่มีลายเซ็นถูกต้องและตรงกับช่วงเวลาใบงาน |

- **Containment:** กักเครื่องผ่าน EDR กักเก็บเมลฟิชชิงทุกกล่อง และบล็อก C2 IP ที่ Firewall
- **Eradication:** หยุดโปรเซสร้าย ลบ webshell/งานแอบรัน ปิดช่องโหว่ด้วยแพตช์และแก้โค้ด SQL Query
- **Recovery:** กู้ไฟล์จาก Immutable Backup ที่ปลอดภัย และเฝ้าระวังไม่ให้ติดซ้ำ

---

## 4. สรุปสั้นก่อนลงสนามและแหล่งอ้างอิง (Key Takeaways & References)

- แยกการโจมตีจริงออกจาก **Vulnerability Scanner หรือทีม Pentest ที่ได้รับอนุญาต** โดยเช็ก IP Allowlist ก่อนเสมอ
- คู่มือ Query: [Microsoft Learn: KQL](https://learn.microsoft.com/en-us/kusto/query/)
