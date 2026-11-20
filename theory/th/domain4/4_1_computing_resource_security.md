# 4.1 ประยุกต์ใช้เทคนิคความปลอดภัยทั่วไปกับทรัพยากรคอมพิวเตอร์ (Apply common security techniques to computing resources)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การเสริมความแข็งแกร่งให้เซิร์ฟเวอร์ เวิร์กสเตชัน อุปกรณ์พกพา และเครือข่ายไร้สาย ช่วยจำกัดการทำงานเริ่มต้นและการฝังตัวหลังการเจาะระบบ ([NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [Endpoint Detection and Response](https://en.wikipedia.org/wiki/Endpoint_detection_and_response)):

- **ค่ามาตรฐานความปลอดภัยและการเสริมความแกร่งให้ Endpoint:** การบังคับใช้ CIS Benchmarks หรือค่ามาตรฐานผู้ผลิตผ่าน Group Policy (`GPO`) หรือระบบจัดการคอนฟิก ได้แก่ การลบบริการที่ไม่ใช้ การปิดพอร์ตไม่จำเป็น เปลี่ยนรหัสผ่านโรงงาน เปิดใช้ **UEFI Secure Boot** (ป้องกัน Bootkits) บังคับใช้ Full-Disk Encryption และเปิดระบบควบคุมสิทธิ์ระดับเคอร์เนล (`SELinux`, `AppArmor`)
- **การป้องกันระดับ Endpoint และแอปพลิเคชัน:** ติดตั้งระบบ **EDR** แบบรวมศูนย์, Anti-Malware, Host-Based Firewalls ร่วมกับ Application Sandboxing และการป้องกันความสมบูรณ์ของหน่วยความจำ
- **การจัดการอุปกรณ์พกพา (`MDM` & `MAM`):** บังคับใช้การล็อกหน้าจอ Remote Wipe การตรวจจับ Jailbreak/Root การแยกคอนเทนเนอร์ข้อมูลงานออกจากแอปส่วนตัว และโมเดลการใช้งานองค์กร:
  - **BYOD (Bring Your Own Device):** พนักงานเป็นเจ้าของเครื่อง ต้องอาศัย `MAM` containerization เพื่อรักษาความเป็นส่วนตัว
  - **COPE (Corporate-Owned, Personally Enabled):** องค์กรเป็นเจ้าของ ควบคุมผ่าน `MDM` และอนุญาตการใช้งานส่วนตัวแบบจำกัด
  - **CYOD / Corporate-Owned Business Only:** การควบคุมและข้อจำกัดด้าน Compliance สูงสุด
- **การเสริมความปลอดภัยระบบไร้สาย:** ใช้มาตรฐาน **WPA3-Enterprise** (`802.1X / EAP-TLS` หรือ `PEAP` พิสูจน์ตัวตนกับ RADIUS) แทนรหัส PSK ร่วมกัน และแยก Guest SSID ด้วย Captive Portal และ Client Isolation

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ข้อมูลสถานะความปลอดภัยของ Endpoint ช่วยยืนยันว่าโพรเซสต้องสงสัยทำงานภายใต้ระบบที่เข้มงวดหรือเกิดจากการแก้ค่าคอนฟิก ผู้โจมตีมักพยายามปิดระบบป้องกันบนเครื่องหรือแก้ค่า Registry/GPO ([MITRE ATT&CK T1562 — Impair Defenses](https://attack.mitre.org/techniques/T1562/)) หรือสร้างงานอัตโนมัติสำหรับ Persistence ([T1053](https://attack.mitre.org/techniques/T1053/)):

```kql
DeviceRegistryEvents
| where ActionType == "RegistryValueSet"
| where RegistryKey has_any ("DisableAntiSpyware", "DisableRealtimeMonitoring", "CurrentVersion\\Run")
| project Timestamp, DeviceName, InitiatingProcessAccountName, RegistryKey, RegistryValueData
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Endpoint detection and response](https://en.wikipedia.org/wiki/Endpoint_detection_and_response)
- [NIST SP 800-53 Rev. 5 — System and Information Integrity (SI)](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Impair Defenses (T1562)](https://attack.mitre.org/techniques/T1562/)
- [MITRE ATT&CK — Scheduled Task/Job (T1053)](https://attack.mitre.org/techniques/T1053/)
- [Microsoft Learn — KQL Overview](https://learn.microsoft.com/en-us/kusto/query/)
