# Domain 3: สถาปัตยกรรมความปลอดภัย (สรุปกระชับ 3 นาที)

> **Domain 3** | **สัดส่วนในข้อสอบ:** 18% (`5 / 30` ข้อใน Pre/Post-Test) | **เวลาอ่านโดยประมาณ:** ~3 นาที

ลองนึกภาพการออกแบบตึกที่มีโถงรับแขก โซนล็อกเข้าห้องเรียน และเครื่องปั่นไฟสำรอง — ต่อให้คนนอกเดินเข้าโถงหน้าบ้านได้ ก็ไม่มีทางเดินทะลุเข้าห้องเก็บความลับได้ตรงๆ

---

## 1. ทฤษฎีหลัก CompTIA Security+ ที่ออกสอบบ่อย (Core Security+ Theory)

- **Defense-in-Depth & การแบ่งโซนเครือข่าย** ([Wikipedia: Defense in Depth](https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29), [Wikipedia: DMZ](https://en.wikipedia.org/wiki/DMZ_%28computing%29)):
  - *Defense-in-Depth:* ซ้อนด่านป้องกันหลายชั้น (Firewall + IAM + EDR + Backup) เพื่อไม่ให้ด่านเดียวพังแล้วระบบล่มทั้งหมด
  - *Segmentation & Zero Trust:* แยกโซน (DMZ สาธารณะ, วงผู้ใช้, วงฐานข้อมูล, วง OT) และตรวจสิทธิ์ทุกคำขอ ([NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final), [CISA Zero Trust](https://www.cisa.gov/zero-trust-maturity-model))
  - *DMZ & Jump Host:* เว็บที่เปิดให้คนนอกใช้ต้องอยู่ใน DMZ และผู้ดูแลระบบต้องผ่าน Bastion/Jump Host ที่เปิด MFA ก่อนเข้าเซิร์ฟเวอร์สำคัญ
- **Cloud Shared Responsibility Model (แบ่งความรับผิดชอบบนคลาวด์):**
  - *IaaS* (เช่น เช่า VM): ผู้ให้บริการดูแลฮาร์ดแวร์/Hypervisor ส่วน **เรา** ต้องแพตช์ OS ตั้ง Firewall/IAM และดูแลข้อมูลเอง
  - *PaaS:* ผู้ให้บริการดูแล OS/Runtime ส่วน **เรา** ดูแลโค้ดแอปและข้อมูล และ *SaaS:* **เรา** ดูแลสิทธิ์ผู้ใช้และการจัดชั้นข้อมูล
- **Firewall & โปรโตคอลเข้ารหัส** ([Wikipedia: Firewall](https://en.wikipedia.org/wiki/Firewall_%28computing%29)):
  - ใช้ NGFW/WAF คัดกรองทราฟฟิก และเลือกใช้ **TLS 1.3**, **IPsec VPN**, **SSH/SFTP**, **SNMPv3** แทน Telnet/FTP แบบไม่เข้ารหัส
- **ความทนทาน (Resilience) & สถานะข้อมูล** ([CISA #StopRansomware](https://www.cisa.gov/stopransomware)):
  - **Immutable Backup กฎ 3-2-1:** สำรองข้อมูล 3 ชุดบนสื่อ 2 ชนิด และ 1 ชุดอยู่นอกสถานที่แบบแก้ไขไม่ได้ (กัน Ransomware) พร้อมปกป้องข้อมูลทั้ง **At Rest / In Transit / In Use**

---

## 2. สัญญาณอันตรายใน SIEM Log, KQL และ MITRE ATT&CK

จับการกระโดดข้ามโซน ([T1021](https://attack.mitre.org/techniques/T1021/)) การเปิดช่องโหว่ Firewall/Cloud ([T1562](https://attack.mitre.org/techniques/T1562/)) และแอบส่งข้อมูลออกช่องทางอื่น ([T1048](https://attack.mitre.org/techniques/T1048/)):

```kql
(source == "firewall" and dst_port == "445") or (source == "cloudtrail" and message contains "SecurityGroup")
```

- **จุดสังเกต:** เครื่องใน DMZ เชื่อมต่อไปพอร์ต `445` (SMB) หรือ `3389` (RDP) ในวงฐานข้อมูลภายใน หรือ Security Group ถูกเปิดเป็น `0.0.0.0/0`

---

## 3. การคัดแยก True Positive vs. False Positive และ Playbook 3 ระยะ

| True Positive (ภัยคุกคามจริง) | False Positive (งานปกติ / มี Benign Context Log ยืนยัน) |
|---|---|
| Cloud Security Group ถูกเปิด `0.0.0.0/0` โดย API Key แปลกปลอม | การเปิดพอร์ตหรือปรับกติกา Firewall ที่ตรงกับ **Change Ticket ที่ได้รับอนุมัติ** |
| ปริมาณข้อมูลเข้ารหัสไหลออกไปยัง IP ภายนอกที่ไม่รู้จัก | งาน Backup ตามตารางเวลาไปยังเครื่องสำรองข้อมูลหรือ CDN ที่อนุญาต |

ตามแนวทาง [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final):
- **Containment:** ปิดเส้นทางข้ามโซนด้วย ACL ถอนกฎ `0.0.0.0/0` ที่เปิดพลาด และกักเครื่อง DMZ ที่ถูกเจาะ
- **Eradication:** ปิดพอร์ตไม่ปลอดภัย (Telnet/FTP/SMBv1) หมุนเวียน Cloud Keys ที่รั่ว และแก้โค้ด IaC
- **Recovery:** กู้ค่าการตั้งค่าที่ปลอดภัย ตรวจซ้ำการแยกโซน และเฝ้าระวังทราฟฟิกข้ามโซนต่อเนื่อง

---

## 4. สรุปสั้นก่อนลงสนามและแหล่งอ้างอิง (Key Takeaways & References)

- เครื่องใน **DMZ** ไม่ควรเชื่อมต่อตรงมายังฐานข้อมูลภายในด้วยพอร์ตแอดมิน (`22`, `445`, `3389`) หากไม่มีใบงานอนุมัติ
- คู่มือ Query: [Microsoft Learn: KQL](https://learn.microsoft.com/en-us/kusto/query/)
