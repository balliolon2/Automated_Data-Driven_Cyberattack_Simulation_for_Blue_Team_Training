# 2.1 เปรียบเทียบและแยกแยะกลุ่มผู้คุกคามและแหล่งข่าวกรองภัยคุกคาม (Compare and contrast threat actors and intelligence sources)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การเข้าใจลักษณะของผู้โจมตี ทั้งด้านเงินทุน ความซับซ้อน ระดับสิทธิ์การเข้าถึง และเจตนาเชิงกลยุทธ์ ช่วยให้จัดลำดับความสำคัญของการป้องกันได้อย่างแม่นยำ ([Threat (Computer)](https://en.wikipedia.org/wiki/Threat_(computer)), [Advanced Persistent Threat](https://en.wikipedia.org/wiki/Advanced_persistent_threat)):

- **ประเภทของกลุ่มผู้คุกคาม (Threat Actor Categories):**
  - **Nation-State / APT:** ทีมปฏิบัติการที่ได้รับการสนับสนุนจากรัฐ มีเงินทุนสูง มุ่งจารกรรมข้อมูลระยะยาว ขโมยทรัพย์สินทางปัญญา หรือขัดขวางโครงสร้างพื้นฐานสำคัญโดยใช้เครื่องมือเฉพาะทางและช่องโหว่ Zero-Day
  - **Organized Cybercrime:** กลุ่มอาชญากรรมไซเบอร์เพื่อผลกำไร ดำเนินระบบ Ransomware-as-a-Service (RaaS) ตลาดซื้อขาย Credential และการหลอกลวงโอนเงินทางอีเมลองค์กร (BEC)
  - **Hacktivists:** ผู้โจมตีที่มีแรงจูงใจด้านอุดมการณ์หรือการเมือง ใช้ DDoS, การเปลี่ยนหน้าเว็บไซต์ หรือปฏิบัติการ Hack-and-Leak เพื่อสื่อสารเชิงสัญลักษณ์
  - **Unskilled Attackers (Script Kiddies):** ผู้โจมตีฉวยโอกาสที่ใช้ชุดเครื่องมือโจมตีสาธารณะโดยไม่ได้เข้าใจกลไกเชิงลึก
  - **Insider Threats (ภัยคุกคามจากคนใน):** พนักงาน ผู้รับเหมา หรือคู่ค้าที่ได้รับอนุญาตซึ่งกระทำการโดย **เจตนาร้าย (Malicious)** (ความแค้น การรับสินบน การก่อวินาศกรรม) หรือ **ความประมาท (Negligent)** (ทำข้อมูลรั่วไหลโดยไม่ตั้งใจ การใช้ Shadow IT)
- **แหล่งข่าวกรองภัยคุกคาม (Threat Intelligence Sources):**
  - **OSINT (Open-Source Intelligence):** ประกาศแจ้งเตือนสาธารณะ ฐานข้อมูล CVE การเฝ้าติดตาม Paste Sites บันทึก WHOIS/Certificate Transparency และการตรวจสอบ Dark Web
  - **ชุมชนแบ่งปันข้อมูล (Information-Sharing Communities):** ศูนย์ประสานงานเฉพาะอุตสาหกรรมอย่าง **ISACs / ISAOs** และคำแนะนำจากภาครัฐ เช่น [CISA Advisories](https://www.cisa.gov/stopransomware)
  - **ฟีดข้อมูลอัตโนมัติ (Automated Feeds):** ตัวบ่งชี้การบุกรุก (IOCs) ที่เครื่องอ่านได้ซึ่งแลกเปลี่ยนผ่านรูปแบบมาตรฐาน **STIX / TAXII** และเชื่อมโยงกับพฤติกรรมผู้โจมตีใน [MITRE ATT&CK Enterprise](https://attack.mitre.org/matrices/enterprise/)

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ข่าวกรองภัยคุกคาม (`CTI`) ช่วยยกระดับทีม SOC จากการไล่บล็อกหมายเลข IP ที่เปลี่ยนตลอดเวลาไปสู่การล่าหา TTPs (Tactics, Techniques, and Procedures) ที่คงทนกว่า เมื่อฟีดของ ISAC แจ้งเตือนการโจมตีแบบ Password Spraying ([MITRE ATT&CK T1110](https://attack.mitre.org/techniques/T1110/)) หรือการบุกรุกผ่านคู่ค้าที่ไว้วางใจ ([T1199](https://attack.mitre.org/techniques/T1199/)) นักวิเคราะห์จะกวาดค้น Log ย้อนหลังเพื่อหาตัวบ่งชี้ที่ตรงกัน:

```kql
ThreatIntelligenceIndicator
| join kind=inner (CommonSecurityLog | where TimeGenerated > ago(7d)) on $left.NetworkIP == $right.SourceIP
| project TimeGenerated, SourceIP, DestinationIP, DestinationPort, Description
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Threat (computer)](https://en.wikipedia.org/wiki/Threat_(computer))
- [Wikipedia — Advanced persistent threat](https://en.wikipedia.org/wiki/Advanced_persistent_threat)
- [MITRE ATT&CK — Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/)
- [CISA — #StopRansomware Resources](https://www.cisa.gov/stopransomware)
- [Stanford University UIT — Information Security](https://uit.stanford.edu/security)
