# 3.2 ประยุกต์ใช้หลักการความปลอดภัยเพื่อป้องกันโครงสร้างพื้นฐานองค์กร (Apply security principles to secure enterprise infrastructure)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การปกป้องโครงสร้างพื้นฐานองค์กรเป็นการนำหลัก [Defense in Depth](https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29) มาบังคับใช้ทั้งที่ขอบเขตเครือข่าย จุดแบ่งความเชื่อถือภายใน และฮาร์ดแวร์อุปกรณ์:

- **การจัดวางอุปกรณ์และโซนความปลอดภัย (Device Placement & Security Zones):**
  - **[Screened Subnets / DMZ](https://en.wikipedia.org/wiki/DMZ_%28computing%29):** แยกเซิร์ฟเวอร์ที่เข้าถึงได้จากภายนอก (Web Reverse Proxies, Mail Relays) ไว้ระหว่างนโยบาย Firewall สองชั้น เพื่อไม่ให้การถูกเจาะจากภายนอกสามารถกระโดดตรงเข้าสู่ Active Directory หรือฐานข้อมูลภายในได้
  - **Microsegmentation & VLANs:** จำกัดการเคลื่อนที่แนวขวาง (**East–West Lateral Movement**) ระหว่างโฮสต์ภายใน Subnet เดียวกันผ่าน Host Firewalls หรือนโยบาย SDN
- **อุปกรณ์ความปลอดภัยและจุดบังคับนโยบาย (Security Appliances & Enforcement Points):**
  - **[Firewalls](https://en.wikipedia.org/wiki/Firewall_%28computing%29):** แบบ Stateless ตรวจ Header 5-tuple; แบบ **Stateful** ติดตามตารางสถานะการเชื่อมต่อ TCP/UDP; **NGFWs / UTMs** ตรวจลึกถึง Layer-7; และ **WAFs** ตรวจสอบคำขอ HTTP/S เพื่อป้องกัน SQLi/XSS
  - **[IDS / IPS](https://en.wikipedia.org/wiki/Intrusion_detection_system):** เซ็นเซอร์ตรวจจับตาม Signature และ Anomaly โดยทำงานแบบนอกเส้นทาง (`IDS` เฝ้าระวัง/แจ้งเตือน) หรือวางขวางเส้นทาง (`IPS` บล็อก/ตัดการเชื่อมต่อทันที)
  - **Forward / Reverse Proxies, Load Balancers & Jump Servers (Bastion Hosts):** จบการเชื่อมต่อ TLS ที่จุดศูนย์กลาง กระจายโหลดเพื่อความพร้อมใช้งาน กรองทราฟฟิกขาออก และบังคับการเชื่อมต่อ SSH/RDP ของผู้ดูแลผ่าน Bastion Host ที่ตรวจสอบย้อนหลังได้
  - **Network Access Control (`NAC` / 802.1X):** ตรวจสอบตัวตนและสถานะสุขภาพของอุปกรณ์ (ระดับแพตช์ การทำงานของ EDR) ก่อนอนุญาตเข้าใช้งาน VLAN เต็มรูปแบบ
- **สถาปัตยกรรมการเข้าถึงที่ปลอดภัย:** **VPN** แบบ IPsec/WireGuard, **SD-WAN**, **SASE (Secure Access Service Edge)** และ **ZTNA** เข้ามาแทนที่ความเชื่อถือโดยปริยายภายใน LAN ด้วยการตรวจสอบตัวตนและอุปกรณ์ในทุกเซสชัน

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ทีม Blue Team เชื่อมโยงข้อมูลจาก Proxy, Firewall และ IPS เพื่อดักจับผู้โจมตีที่ส่งสัญญาณ C2 ผ่าน Proxy นอกมาตรฐาน ([MITRE ATT&CK T1090](https://attack.mitre.org/techniques/T1090/)) หรือสแกนพอร์ตภายในผ่าน **Remote Services** ([T1021](https://attack.mitre.org/techniques/T1021/)):

```kql
CommonSecurityLog
| where DeviceAction in ("Deny", "Drop", "Reset")
| summarize BlockedConnections = count(), DestPorts = make_set(DestinationPort, 10) by SourceIP, DestinationIP
| where array_length(DestPorts) >= 8 or BlockedConnections >= 100
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Defense in depth (computing)](https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29)
- [Wikipedia — DMZ (computing)](https://en.wikipedia.org/wiki/DMZ_%28computing%29)
- [Wikipedia — Firewall (computing)](https://en.wikipedia.org/wiki/Firewall_%28computing%29)
- [Wikipedia — Intrusion detection system](https://en.wikipedia.org/wiki/Intrusion_detection_system)
- [MITRE ATT&CK — Proxy (T1090)](https://attack.mitre.org/techniques/T1090/)
