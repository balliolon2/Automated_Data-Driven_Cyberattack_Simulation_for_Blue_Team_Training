# Domain 4: ปฏิบัติการด้านความปลอดภัย (สรุปกระชับ 3 นาที)

> **Domain 4** | **สัดส่วนในข้อสอบ:** 28% (`8 / 30` ข้อใน Pre/Post-Test — น้ำหนักมากที่สุด!) | **เวลาอ่านโดยประมาณ:** ~3 นาที

ลองนึกภาพศูนย์ฉุกเฉินโรงพยาบาล: **SIEM** คือจอมอนิเตอร์รวมสัญญาณชีพ (Log) **SOAR** ช่วยทำงานอัตโนมัติ และทีม SOC ต้องตัดสินใจว่าจะกักโรคและเก็บหลักฐานอย่างไรไม่ให้พยานหลักฐานเสียหาย

---

## 1. ทฤษฎีหลัก CompTIA Security+ ที่ออกสอบบ่อย (Core Security+ Theory)

- **เครื่องมือ SOC: SIEM, SOAR & EDR** ([Wikipedia: SIEM](https://en.wikipedia.org/wiki/Security_information_and_event_management)):
  - *SIEM:* รวบรวมและเชื่อมโยง Log จาก Firewall/เครื่อง/ระบบล็อกอิน โดยตั้งเวลา **NTP** ให้ตรงกันทุกเครื่อง
  - *SOAR:* ช่วยตรวจสอบข้อมูลซ้ำๆ และสั่งกักกันอัตโนมัติตาม Playbook
  - *EDR vs. IDS/IPS:* EDR เฝ้าระวังโปรเซสและหน่วยความจำบนเครื่อง ส่วน IDS คอยแจ้งเตือน แต่ **IPS** จะวางขวางและบล็อกการโจมตีทันที
- **การจัดการช่องโหว่ & ความแม่นยำการตรวจจับ:**
  - จัดลำดับความสำคัญแพตช์ด้วยคะแนน **CVSS** และความสำคัญของเครื่อง
  - *False Positive* = เตือนผิดทั้งที่ปกติ ส่วน *False Negative* = มองไม่เห็นการโจมตีจริง (อันตรายที่สุด!)
- **วงจร Incident Response ของ NIST** ([NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final), [CMU IR Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)):
  1. *Preparation* (เตรียมเครื่องมือและซ้อมแผน) → 2. *Detection & Analysis* (คัดแยกและเก็บ Log หลักฐาน) → 3. *Containment, Eradication & Recovery* → 4. *Post-Incident Activity (ถอดบทเรียน)*
- **นิติวิทยาศาสตร์คอมพิวเตอร์ (Digital Forensics)** ([Wikipedia: Digital Forensics](https://en.wikipedia.org/wiki/Digital_forensics)):
  - **Order of Volatility (RFC 3227):** เก็บข้อมูลที่หายง่ายที่สุดก่อนเสมอ: CPU Cache → RAM → โปรเซส/การเชื่อมต่อเครือข่าย → ไฟล์ชั่วคราว → ฮาร์ดดิสก์ → Backup พร้อมบันทึก **Chain of Custody + SHA-256 Hash**

---

## 2. สัญญาณอันตรายใน SIEM Log, KQL และ MITRE ATT&CK

งานหลักของ SOC Tier-1 คือจับการขโมยรหัสในหน่วยความจำ ([T1003](https://attack.mitre.org/techniques/T1003/)), ตั้งงานแอบรันอัตโนมัติ ([T1053](https://attack.mitre.org/techniques/T1053/)), ยิงเดารหัส Brute Force ([T1110](https://attack.mitre.org/techniques/T1110/)) และการติดต่อ C2 ([T1071](https://attack.mitre.org/techniques/T1071/)):

```kql
severity == "high" or severity == "critical" or (source == "edr" and message contains "lsass")
```

- **สูตรสืบสวน SIEM 4 ขั้น:** ดู Alert ตั้งต้น → กรอง `host` หรือ `src_ip` ช่วงเวลาเดียวกัน → หาการแพร่กระจาย/ฝังตัว → เช็กใบงานบำรุงรักษาปกติก่อนฟันธง

---

## 3. การคัดแยก True Positive vs. False Positive และ Playbook 3 ระยะ

| True Positive (ภัยคุกคามจริง) | False Positive (งานปกติ / มี Benign Context Log ยืนยัน) |
|---|---|
| โปรแกรมแปลกปลอมอ่านหน่วยความจำ `lsass.exe` แล้วต่อ SMB ข้ามเครื่อง | โปรแกรม Antivirus สแกนตามตาราง หรือ SCCM อัปเดตซอฟต์แวร์ปกติ |
| ล็อกอินผิดรัวๆ (`4625`) แล้วผ่าน 1 ครั้ง พร้อมสร้าง Scheduled Task ใหม่ | สคริปต์ตรวจสอบระบบของทีม IT ที่ลงทะเบียนไว้ใน CMDB แล้ว |

- **Containment:** สั่ง **Network Isolate ผ่าน EDR** (ห้ามถอดปลั๊กปิดเครื่องทันทีเพราะข้อมูลใน RAM จะหาย!) บล็อก C2 และล็อกบัญชีที่เสี่ยง
- **Eradication:** หยุดโปรเซสและลบ Scheduled Task ร้าย เพิกถอน Session/Kerberos Token แล้วเปลี่ยนรหัสผ่าน
- **Recovery:** นำเครื่องสะอาดกลับเข้าสู่ระบบ สแกนซ้ำ และปรับกฎตรวจจับเพิ่มเติม

---

## 4. สรุปสั้นก่อนลงสนามและแหล่งอ้างอิง (Key Takeaways & References)

- ห้ามปิดเครื่องที่ถูกโจมตีทันทีก่อนเก็บ **RAM**; ให้ใช้ **EDR Isolate** แทนเพื่อรักษาหลักฐานสำคัญ!
- คู่มือ Query: [Microsoft Learn: KQL](https://learn.microsoft.com/en-us/kusto/query/)
