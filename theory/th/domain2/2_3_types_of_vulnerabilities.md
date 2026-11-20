# 2.3 อธิบายประเภทต่างๆ ของช่องโหว่ด้านความปลอดภัย (Explain various types of vulnerabilities)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

ช่องโหว่ ([Vulnerability](https://en.wikipedia.org/wiki/Vulnerability_(computing))) คือจุดอ่อนในการออกแบบ โค้ด การตั้งค่า หรือกระบวนการปฏิบัติงานที่ผู้โจมตีสามารถใช้ประโยชน์เพื่อละเมิดนโยบายความปลอดภัยได้ ([OWASP Top 10:2021](https://owasp.org/Top10/)):

- **ข้อบกพร่องในแอปพลิเคชันและการแทรกคำสั่ง (Application & Injection Flaws):**
  - **[SQL Injection (SQLi)](https://en.wikipedia.org/wiki/SQL_injection):** อินพุตที่ไม่ได้รับการตรวจสอบเข้าเปลี่ยนแปลงคิวรีฐานข้อมูลฝั่งหลังบ้าน ป้องกันได้ด้วย Parameterized Queries / Prepared Statements
  - **[Cross-Site Scripting (XSS)](https://en.wikipedia.org/wiki/Cross-site_scripting):** สคริปต์อันตรายถูกรันบนเบราว์เซอร์ของเหยื่อผ่านรูปแบบ Reflected, Stored หรือ DOM-based บรรเทาด้วย Context-Aware Output Encoding และ Content Security Policy (`CSP`)
  - **ข้อบกพร่องหน่วยความจำและตรรกะ:** Buffer Overflow, Integer Overflow, **Race Conditions** (`TOCTOU` — Time-of-Check to Time-of-Use), Insecure Deserialization, SSRF, CSRF และ Directory Traversal (`../`)
- **จุดอ่อนในระบบปฏิบัติการ เฟิร์มแวร์ และ Virtualization:** ระบบปฏิบัติการที่สิ้นสุดการสนับสนุนแล้ว (`EOL`) ขาดแพตช์, การยกระดับสิทธิ์ใน Kernel, การฝังโค้ดใน UEFI/Firmware และ **VM Escape** จาก Guest OS ออกไปยัง Hypervisor
- **การตั้งค่าผิดพลาดและจุดอ่อนทางการเข้ารหัส (Misconfigurations & Cryptographic Weaknesses):** รหัสผ่านตั้งต้น ถังเก็บข้อมูลคลาวด์ที่เปิดสาธารณะ กฎ CORS/ACL ที่อนุญาตเกินจำเป็น โปรโตคอลไม่เข้ารหัส และชุดรหัสลับที่อ่อนแอ
- **ช่องโหว่ Zero-Day และ Supply-Chain:** ช่องโหว่ที่ยังไม่เคยถูกเปิดเผยและยังไม่มีแพตช์จากผู้พัฒนา จำเป็นต้องใช้มาตรการชดเชยทางเครือข่ายและการเฝ้าระวังระดับ Runtime ระหว่างรอแก้ไข

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

การคัดแยกเหตุการณ์ช่องโหว่จะเชื่อมโยงรายงานผลการสแกนเข้ากับทราฟฟิกการโจมตีจริงต่อบริการสาธารณะ ([MITRE ATT&CK T1190](https://attack.mitre.org/techniques/T1190/)) หรือการเรียกใช้ตัวแปลคำสั่งต้องสงสัย ([T1059](https://attack.mitre.org/techniques/T1059/)) การตรวจยืนยันว่า WAF บล็อกด้วยรหัส `403` ได้สำเร็จ หรือเซิร์ฟเวอร์ตอบรหัส `500`/`200` พร้อมมีการเชื่อมต่อขาออก จะช่วยแยกแยะระหว่างการสแกนทดสอบกับกรณีถูกเจาะระบบสำเร็จ:

```kql
W3CIISLog
| where csUriQuery matches regex @"(?i)(union\s+select|<script>|../|\$\{jndi:)"
| summarize HitCount = count(), StatusCodes = make_set(scStatus) by cIP, csUriStem
| order by HitCount desc
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Vulnerability (computing)](https://en.wikipedia.org/wiki/Vulnerability_(computing))
- [OWASP Top 10:2021 — Critical Web Application Security Risks](https://owasp.org/Top10/)
- [Wikipedia — SQL injection](https://en.wikipedia.org/wiki/SQL_injection)
- [Wikipedia — Cross-site scripting](https://en.wikipedia.org/wiki/Cross-site_scripting)
- [MITRE ATT&CK — Exploit Public-Facing Application (T1190)](https://attack.mitre.org/techniques/T1190/)
