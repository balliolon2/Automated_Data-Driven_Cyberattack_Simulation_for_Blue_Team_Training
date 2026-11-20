# 4.9 ใช้แหล่งข้อมูลเพื่อสนับสนุนการสืบสวนและนิติวิทยาศาสตร์ดิจิทัล (Use data sources to support an investigation and digital forensics)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

ในการสืบสวนที่ต้องใช้เป็นพยานหลักฐานทางกฎหมายหรือวิเคราะห์รากเหง้าอย่างละเอียด ผู้ตอบสนองเหตุการณ์ต้องบูรณาการการวิเคราะห์ Log เข้ากับมาตรฐาน [Digital Forensics](https://en.wikipedia.org/wiki/Digital_forensics) ตามแนวทาง [NIST SP 800-86](https://csrc.nist.gov/pubs/sp/800/86/final):

- **Order of Volatility (หลัก RFC 3227):** จัดเก็บหลักฐานโดยเริ่มจากข้อมูลที่สูญหายง่ายที่สุดก่อนที่ไฟดับหรือหมดอายุ:
  1. CPU registers และ Cache
  2. ตาราง Routing, ARP cache, รายการโพรเซส และ **หน่วยความจำระบบ (RAM dump)**
  3. ระบบไฟล์ชั่วคราว / Swap space
  4. สื่อบันทึกข้อมูลถาวร (**Bit-stream forensic disk image**)
  5. Remote logging, ข้อมูลสำรอง และสื่อทางกายภาพ
- **การรักษาความสมบูรณ์และห่วงโซ่พยานหลักฐาน ([Chain of Custody](https://en.wikipedia.org/wiki/Chain_of_custody)):**
  - การออกคำสั่ง **Legal Hold** ระงับการลบ Log/อีเมล/ข้อมูลสำรองตามรอบปกติทันทีเมื่อเข้าสู่กระบวนการกฎหมายหรือ E-Discovery
  - การใช้ฮาร์ดแวร์ **Write Blockers** ป้องกันไม่ให้การ Mount ดิสก์ไปแก้ไขค่า Timestamp บนพยานหลักฐานต้นฉบับ
  - การคำนวณค่า Cryptographic Digests (**SHA-256 / SHA-512**) ณ เวลาจัดเก็บและทวนสอบ Hash ทุกครั้งที่มีการส่งมอบ พร้อมบันทึกลงใน **Chain of Custody** (ใคร อะไร เมื่อไร ที่ไหน อย่างไร)
- **แหล่งข้อมูลสืบสวน:** SIEM Logs, Firewall/DNS/Proxy Trails, NetFlow Metadata, Packet Captures เต็มรูปแบบ (`PCAP`), ไทม์ไลน์หน่วยความจำจาก EDR และ **การปรับเทียบค่าความต่างเวลา (Time Offset Normalization)** ระหว่างเครื่องที่เวลาไม่ตรงกัน

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ห้ามรีบูตเครื่องที่ถูกบุกรุกก่อนเก็บ Volatile RAM หากสงสัยการดึงรหัสผ่านจากหน่วยความจำ (`LSASS` memory injection — [MITRE ATT&CK T1003](https://attack.mitre.org/techniques/T1003/)) หรือเพย์โหลด PowerShell แบบไร้ไฟล์ ให้สั่งแยกเครื่องผ่าน Network/EDR Isolation ก่อน จากนั้นจึงเก็บหน่วยความจำ ทำ Hash ของดิสก์ และเชื่อมโยงหลักฐานด้วยเวลา UTC ที่ปรับเทียบแล้ว:

```kql
SecurityEvent
| where EventID in (4688, 1102, 4697)
| project TimeGenerated, Computer, SubjectUserName, EventID, CommandLine
| sort by TimeGenerated asc
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Digital forensics](https://en.wikipedia.org/wiki/Digital_forensics)
- [Wikipedia — Chain of custody](https://en.wikipedia.org/wiki/Chain_of_custody)
- [NIST SP 800-86 — Integrating Forensic Techniques into Incident Response](https://csrc.nist.gov/pubs/sp/800/86/final)
- [MITRE ATT&CK — OS Credential Dumping (T1003)](https://attack.mitre.org/techniques/T1003/)
