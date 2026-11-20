# 4.5 ปรับปรุงความสามารถขององค์กรเพื่อยกระดับความปลอดภัย (Modify enterprise capabilities to enhance security)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

วิศวกรปฏิบัติการความปลอดภัยต้องปรับจูนและเสริมความแข็งแกร่งของมาตรการในระบบการผลิตอย่างต่อเนื่องเพื่อปิดเส้นทางโจมตีใหม่ๆ ([NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [Firewall (Computing)](https://en.wikipedia.org/wiki/Firewall_%28computing%29)):

- **การปรับปรุงขีดความสามารถของ Firewall และเครือข่าย:** อัปเดต Stateful ACLs และโซนความปลอดภัยให้บังคับใช้ Explicit Allowlisting ปิดพอร์ตจัดการขาเข้าแบบเดิม (`22/3389`) ไว้หลัง Jump Hosts/ZTNA และเพิ่มระบบ **DNS Filtering / Sinkholing** เพื่อให้การค้นหาโดเมน C2/ฟิชชิ่งที่รู้จักถูกส่งไปยัง Internal Sinkholes ที่ปลอดภัย
- **การเสริมความปลอดภัยอีเมล (Email Security Hardening):** บังคับใช้มาตรฐานการยืนยันผู้ส่งและป้องกันการปลอมแปลงที่ขอบ DNS/MTA:
  - **SPF (Sender Policy Framework):** ประกาศรายชื่อ IP ของเมลเซิร์ฟเวอร์ขาออกที่ได้รับอนุญาตของโดเมน
  - **DKIM (DomainKeys Identified Mail):** ลงนามเข้ารหัสส่วน Header/Body ของอีเมลเพื่อยืนยันความถูกต้องและต้นทาง
  - **DMARC:** ผูกความสอดคล้องของ SPF/DKIM เข้ากับนโยบายบังคับใช้ (`none`, `quarantine`, `reject`) พร้อมส่งรายงานสรุปผล
- **การอัปเกรดโปรโตคอลและบริการ:** แทนที่โปรโตคอลบริหารจัดการที่ไม่เข้ารหัสหรืออ่อนแอด้วยโปรโตคอลที่เข้ารหัสและมีการยืนยันตัวตน (**Telnet → SSH**, **HTTP → HTTPS/TLS 1.3**, **SNMPv1/v2c → SNMPv3**, **LDAP → LDAPS**, **NTP → NTS**)
- **Endpoint, DLP & Content Filtering:** เปิดใช้กฎลดพื้นผิวการโจมตีของ EDR (`ASR`) เพื่อบล็อก Office Macros จากการแตกโพรเซส Child Shells, ปรับจูนหมวดหมู่ URL และกักกันไฟล์ปฏิบัติการที่เป็นอันตราย

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ข้อเสนอแนะด้าน Hardening มักมาจากช่องว่างที่พบระหว่างการสืบสวน SOC เช่น กรณีอีเมลฟิชชิ่งปลอมตัวตน ([MITRE ATT&CK T1566](https://attack.mitre.org/techniques/T1566/)) หลุดผ่านนโยบาย DMARC ที่ตั้งเป็น `p=none` หรือการรัน PowerShell ([T1059](https://attack.mitre.org/techniques/T1059/)) จากไฟล์ที่ดาวน์โหลดจากอินเทอร์เน็ตโดยไม่ถูกบล็อก:

```kql
EmailEvents
| where EmailDirection == "Inbound" and ThreatTypes has "Phish"
| summarize Count = count() by SenderFromDomain, AuthenticationDetails
| order by Count desc
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Firewall (computing)](https://en.wikipedia.org/wiki/Firewall_%28computing%29)
- [CISA — Phishing Threat & Response Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)
- [NIST SP 800-53 Rev. 5 — Configuration Management & Access Control](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Command and Scripting Interpreter (T1059)](https://attack.mitre.org/techniques/T1059/)
