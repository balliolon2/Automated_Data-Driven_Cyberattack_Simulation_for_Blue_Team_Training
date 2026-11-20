# Domain 3: สถาปัตยกรรมความปลอดภัย (Security Architecture)

> **Domain ID:** `domain3` | **สัดส่วนในข้อสอบ:** 18% (5 จาก 30 ข้อใน Pre-Test / Post-Test) | **แหล่งอ้างอิง:** Wikipedia, NIST CSRC, CISA, MITRE ATT&CK

บทเรียนนี้สังเคราะห์แนวคิดสถาปัตยกรรมความมั่นคงปลอดภัยระดับองค์กรและระบบคลาวด์จากมาตรฐานสาธารณะและแหล่งอ้างอิงเปิด ได้แก่ [NIST SP 800-207: Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final), [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [CISA Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model) และบทความสถาปัตยกรรมระบบบน Wikipedia

---

## 1. ทฤษฎีหลัก CompTIA Security+ และแนวคิดการสอบ (Core Security+ Theory)

### 1.1 การแบ่งส่วนเครือข่าย โซนความปลอดภัย และการป้องกันเชิงลึก (Network Segmentation, Zones & Defense-in-Depth)
ตามคำอธิบายใน [Wikipedia: Defense in Depth](https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29), [Wikipedia: DMZ (Computing)](https://en.wikipedia.org/wiki/DMZ_%28computing%29) และ [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final) (SC-7 Boundary Protection):
- **Defense-in-Depth (การป้องกันเชิงลึกหลายชั้น):** การวางมาตรการควบคุมทางเทคนิคและทางบริหารที่อิสระต่อกันเป็นลำดับชั้น เพื่อให้การเจาะผ่านด่านใดด่านหนึ่งไม่นำไปสู่การเข้าถึงข้อมูลหลักได้โดยทันที
- **Network Segmentation & Micro-segmentation:** การแบ่งเครือข่ายองค์กรออกเป็นโซนความเชื่อถือที่แยกจากกัน (User LAN, Server VLAN, OT/ICS VLAN, Management Plane) ควบคุมด้วย Firewalls และ ACLs เพื่อจำกัดการเคลื่อนที่ข้ามระบบ (Lateral Movement) ([NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final))
- **DMZ (Screened Subnet):** แยกบริการที่ต้องเปิดให้ภายนอกเข้าถึง (Web Servers, Reverse Proxies, Mail Gateways) ไว้ในเครือข่ายกันชนที่คั่นจากเครือข่ายภายในด้วย Firewall ([Wikipedia: DMZ](https://en.wikipedia.org/wiki/DMZ_%28computing%29))
- **Air Gaps & Jump Hosts (Bastion Hosts):** การแยกสภาพแวดล้อมความอ่อนไหวสูง (เช่น ระบบทำความเย็น OT/SCADA) ทางกายภาพหรือเชิงตรรกะ และบังคับให้การเข้าถึงเชิงผู้ดูแลระบบต้องผ่าน Jump Server ที่เสริมความแข็งแกร่งและใช้ MFA เสมอ

### 1.2 แบบจำลองความมั่นคงปลอดภัยบนคลาวด์และ Virtualization (Cloud & Virtualization Security Models)
- **รูปแบบการใช้งานคลาวด์ (Cloud Deployment Models):** สภาพแวดล้อมแบบ Public, Private, Hybrid และ Community Cloud
- **แบบจำลองความรับผิดชอบร่วมกัน (Shared Responsibility Model):**
  - *IaaS:* ผู้ให้บริการดูแลฮาร์ดแวร์กายภาพและ Hypervisor; ลูกค้าดูแล Guest OS, การแพตช์, Network Security Groups, IAM และข้อมูล
  - *PaaS:* ผู้ให้บริการดูแล OS และ Runtime; ลูกค้าดูแลโค้ดแอปพลิเคชัน การตั้งค่า และข้อมูล
  - *SaaS:* ผู้ให้บริการดูแลโครงสร้างพื้นฐานและซอฟต์แวร์; ลูกค้าดูแลตัวตนผู้ใช้, RBAC และการจัดระดับชั้นข้อมูล
- **มาตรการควบคุมสถาปัตยกรรมคลาวด์:** Virtual Private Cloud (VPC) Security Groups, NACLs, Cloud Access Security Brokers (CASB), การสแกน Infrastructure as Code (IaC) และ Container แบบ Immutable

### 1.3 โครงสร้างพื้นฐานเครือข่ายและการสื่อสารที่มั่นคงปลอดภัย (Secure Network Infrastructure & Communications)
อธิบายใน [Wikipedia: Firewall (Computing)](https://en.wikipedia.org/wiki/Firewall_%28computing%29) และ [CISA Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model):
- **Firewalls & Application Gateways:** Stateful/Next-Generation Firewalls (NGFW), Web Application Firewalls (WAF) และ Forward/Reverse Proxies ที่บังคับใช้นโยบายข้ามโซนและตรวจจับทราฟฟิกระดับแอปพลิเคชัน
- **โปรโตคอลเข้ารหัสและการสร้างอุโมงค์เชื่อมโยง (Encrypted Transport & Tunneling Protocols):**
  - *TLS 1.2 / 1.3:* เข้ารหัสข้อมูลระหว่างการส่งผ่าน (HTTPS, LDAPS, SMTPS)
  - *IPsec VPN (AH/ESP, IKEv2):* อุโมงค์เข้ารหัสแบบ Site-to-Site และการเข้าถึงระยะไกล
  - *SSH / SFTP / SNMPv3:* โปรโตคอลจัดการระบบและเก็บ Telemetry แบบเข้ารหัส แทนที่ Telnet, FTP และ SNMPv1/v2c ที่ส่งข้อความธรรมดา
  - *DNSSEC, DoH/DoT:* ปกป้องความถูกต้องสมบูรณ์และความเป็นส่วนตัวของ DNS จากการปลอมแปลงและการดักฟัง

### 1.4 ความยืดหยุ่น ความพร้อมใช้งานสูง และสถานะของข้อมูล (Resiliency, High Availability & Data States)
- **ความยืดหยุ่นและความทนทานต่อความขัดข้อง (Resilience & Fault Tolerance):** การทำ Load Balancing แบบ Active-Active การทำสำเนาระหว่างหลายภูมิภาค (Multi-Region Replication) ระบบจัดเก็บข้อมูล RAID และพลังงานสำรอง เพื่อลดจุดล้มเหลวเพียงจุดเดียว (Single Point of Failure)
- **Immutable Backups (กฎ 3-2-1):** การเก็บสำเนาข้อมูล 3 ชุด บนสื่อต่างประเภทกัน 2 ชนิด โดยมีอย่างน้อย 1 ชุดอยู่นอกสถานที่และแก้ไขทับไม่ได้ (Immutable) เพื่อรับประกันการกู้คืนที่สะอาดจาก Ransomware ([CISA #StopRansomware](https://www.cisa.gov/stopransomware))
- **สถานะของข้อมูล (Data States):** การปกป้องข้อมูลขณะจัดเก็บ (Data at Rest ด้วย Full-Disk/Database Encryption), ข้อมูลขณะส่งผ่าน (Data in Transit ด้วย TLS/IPsec) และข้อมูลขณะประมวลผล (Data in Use ด้วย Confidential Computing/Enclaves)

---

## 2. ข้อมูล SOC Telemetry, Log Indicators และการเชื่อมโยง MITRE ATT&CK

เหตุการณ์ใน Domain 3 มุ่งเน้นการละเมิดขอบเขตสถาปัตยกรรม เช่น ทราฟฟิกข้ามโซนที่ผิดปกติ การตั้งค่า Cloud Storage/Security Group ผิดพลาด โปรโตคอลเก่าที่ไม่ปลอดภัย หรือการเคลื่อนที่ข้ามระบบ (Lateral Movement) ระหว่าง Subnets ที่แบ่งแยกไว้

### 2.1 แหล่ง Log และฟิลด์สำคัญ
- **Firewall & NGFW Logs:** `src_ip`, `dst_ip`, `dst_port`, `src_zone`, `dst_zone`, `action` (`ALLOW`/`DENY`), `protocol`, `bytes_out`
- **Cloud Audit Logs (AWS CloudTrail / Azure Activity):** `event_name` (`AuthorizeSecurityGroupIngress`, `PutBucketAcl`), `principal`, `source_ip`, `region`, `resource_arn`
- **Proxy / DNS Logs:** การสร้างอุโมงค์ `CONNECT` ออกภายนอกบนพอร์ตที่ไม่ใช่มาตรฐาน คำสั่ง DNS Tunneling หรือการอัปโหลดไฟล์ไปยัง SaaS ที่ไม่ได้รับอนุญาต

### 2.2 รูปแบบการสืบค้น KQL ที่พบบ่อย

ไวยากรณ์สืบค้นอ้างอิงตาม [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/):

```kql
source == "firewall" and dst_port == "445"
source == "cloudtrail" and message contains "SecurityGroup"
source == "proxy" and action == "allow"
```

### 2.3 การเชื่อมโยง MITRE ATT&CK

รหัสเทคนิคทั้งหมดเชื่อมโยงโดยตรงไปยัง [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/):

| Technique ID | ชื่อเทคนิค (Technique Name) | สัญญาณ Telemetry ที่สังเกตได้ (Observable Indicator) |
|---|---|---|
| [T1021](https://attack.mitre.org/techniques/T1021/) | Remote Services | เครื่องภายในเชื่อมต่อไปยังพอร์ต SMB (445), RDP (3389), WinRM (5985) หรือ SSH (22) ข้ามขอบเขต Subnet |
| [T1562](https://attack.mitre.org/techniques/T1562/) | Impair Defenses | การขยายกฎ Firewall เป็น `0.0.0.0/0` การปิด Cloud Logging หรือการแก้ Security Group เพื่อเปิดพอร์ตจัดการสู่สาธารณะ |
| [T1048](https://attack.mitre.org/techniques/T1048/) | Exfiltration Over Alternative Protocol | การส่งออกข้อมูลปริมาณมากผ่านช่องทาง DNS, ICMP หรือพอร์ตนอกมาตรฐานเพื่อหลบเลี่ยง DLP Proxy |
| [T1090](https://attack.mitre.org/techniques/T1090/) | Proxy | โฮสต์ภายในสร้างอุโมงค์ SOCKS/HTTP ที่ไม่ได้รับอนุญาต หรือทราฟฟิกส่งต่อหลายทอด (Multi-hop Relay) |

---

## 3. การคัดแยกเหตุการณ์ (TP vs. FP) และแนวทางปฏิบัติ 3 ขั้นตอนตาม Playbook

### 3.1 การจำแนก True Positive (TP) ออกจาก False Positive (FP)
- **ตัวบ่งชี้ True Positive:** เว็บเซิร์ฟเวอร์ในโซน DMZ เชื่อมต่อ SMB/RDP เข้าสู่ Subnet ฐานข้อมูลภายใน; Cloud S3 Bucket หรือ Security Group ถูกเปิดเป็น `0.0.0.0/0` โดยผู้ใช้หรือ API Key ผิดปกติ; หรือพบทราฟฟิกขาออกแบนด์วิดท์สูงต่อเนื่องไปยัง IP ภายนอกที่ไม่รู้จักนอกเวลาทำการ
- **ตัวบ่งชี้ False Positive (Benign Context Logs):**
  - **Approved Firewall / Cloud Change Ticket:** การเปิดพอร์ตใหม่หรือแก้ไข Security Group ที่ตรงกับใบคำขอเปลี่ยนแปลงโครงสร้างพื้นฐาน (CR) ที่ได้รับอนุมัติ
  - **Authorized Backup / Replication Job:** การโอนถ่ายข้อมูลเข้ารหัสปริมาณสูงภายในหรือออกนอกสถานที่ในช่วงเวลาสำรองข้อมูลที่กำหนดไปยัง IP ของเครื่องสำรองข้อมูลที่รู้จัก
  - **Observed Authorized Traffic Pattern:** ปริมาณการใช้ Proxy `CONNECT` สูงที่สอดคล้องกับกิจกรรมทางธุรกิจตามปกติ (เช่น CDN กระจายซอฟต์แวร์ หรือการสตรีมสื่อ/อบรมที่ได้รับอนุญาต) เมื่อยืนยันร่วมกับบริบทของ Asset

### 3.2 หลักการตอบสนอง 3 ขั้นตอนตาม Playbook
อ้างอิงตาม [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [NIST SP 800-207](https://csrc.nist.gov/pubs/sp/800/207/final) และ [CMU Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html):
- **Phase 1 — Containment (การจำกัดวงความเสียหาย):**
  - *สิ่งที่ควรทำ:* บังคับใช้ Micro-segmentation ACLs เพื่อปิดกั้นเส้นทางข้ามโซนที่ไม่ได้รับอนุญาต ย้อนคืนกฎ Security Group ที่เปิดกว้างเกินจริง (`0.0.0.0/0`) หรือแยก Workload ใน DMZ/Cloud ที่ถูกบุกรุกออกจากเครือข่าย
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การล้างค่า Network Switch ทั้ง Subnet หรือการปิดระบบ DNS/DHCP หลักของการผลิตทั้งหมด
- **Phase 2 — Eradication (การกำจัดภัยคุกคาม):**
  - *สิ่งที่ควรทำ:* ปิดพอร์ตโปรโตคอลเก่าที่ไม่เข้ารหัส (Telnet/FTP/SMBv1) หมุนเวียน Cloud IAM Keys ที่รั่วไหล ลบ SSH Keys หรือกฎ Firewall แปลกปลอม และแก้ไขการตั้งค่าผิดพลาดในเทมเพลต IaC
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การปล่อยท่อ IaC ที่ตั้งค่าผิดพลาดไว้ตามเดิมจนทำให้การ Deploy รอบถัดไปเปิดช่องโหว่เดิมกลับมาอีก
- **Phase 3 — Recovery (การฟื้นฟูระบบ):**
  - *สิ่งที่ควรทำ:* ตรวจสอบความถูกต้องของกฎแบ่งส่วน Zero Trust ด้วยการ Audit สถาปัตยกรรม กู้คืนค่าคอนฟิกมาตรฐานจาก Immutable Baselines ทดสอบการทำงานของระบบสำรอง/Failover และค่อยๆ เปิดทราฟฟิกระบบผลิตอย่างจำกัดสิทธิ์
  - *ตัวเลือกหลอกที่ควรหลีกเลี่ยง:* การปล่อยกฎ "Allow All" ที่ใช้แก้ปัญหาเฉพาะหน้าทิ้งไว้หลังกู้คืนระบบเสร็จ

---

## 4. บทสรุปสำคัญและแหล่งอ้างอิง (Key Takeaways & References)

- **คิดในมุมขอบเขตความเชื่อถือ (Trust Boundaries):** ตรวจสอบเสมอว่า *โซนใด* (`DMZ`, `Internal`, `Cloud`, `OT`) เป็นผู้เริ่มการเชื่อมต่อ และทิศทางนั้นละเมิดสถาปัตยกรรม Least Privilege หรือไม่
- **ตรวจสอบ Change Record โครงสร้างพื้นฐานเสมอ:** การแจ้งเตือนด้าน Firewall และ Cloud Config มักเป็น False Positive เมื่อมีใบ Change Ticket หรือบันทึกงานสำรองข้อมูล/Replication รองรับ
- **รักษาความพร้อมใช้งานขณะตัดแยก:** Containment ต้องจำกัดเฉพาะขอบเขตหรือโฮสต์ที่มีปัญหา ห้ามสั่งหยุดโครงสร้างพื้นฐานเครือข่ายหลักที่ไม่เกี่ยวข้อง

### แหล่งอ้างอิงเว็บที่น่าเชื่อถือ (Authoritative Web References)
- **Wikipedia (แหล่งความรู้เปิด):**
  - [Defense in Depth (Computing)](https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29)
  - [DMZ (Computing)](https://en.wikipedia.org/wiki/DMZ_%28computing%29)
  - [Firewall (Computing)](https://en.wikipedia.org/wiki/Firewall_%28computing%29)
  - [Zero Trust Security Model](https://en.wikipedia.org/wiki/Zero_trust_security_model)
- **หน่วยงานภาครัฐและมาตรฐาน (NIST & CISA):**
  - [NIST SP 800-207: Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final)
  - [NIST SP 800-53 Rev. 5: Security and Privacy Controls](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
  - [CISA Zero Trust Maturity Model (ZTMM)](https://www.cisa.gov/zero-trust-maturity-model)
- **มหาวิทยาลัยและฐานความรู้เปิด:**
  - [CMU Information Security Office: Incident Response Plan](https://www.cmu.edu/iso/governance/procedures/incident-response.html)
  - [MITRE ATT&CK® Enterprise Matrix](https://attack.mitre.org/matrices/enterprise/) (`T1021`, `T1562`, `T1048`, `T1090`)
  - [Microsoft Learn: Kusto Query Language (KQL)](https://learn.microsoft.com/en-us/kusto/query/)
