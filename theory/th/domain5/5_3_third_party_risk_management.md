# 5.3 อธิบายกระบวนการประเมินและจัดการความเสี่ยงจากบุคคลที่สาม (Explain third-party risk assessment and management processes)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

องค์กรยุคใหม่สืบทอดความเสี่ยงทางไซเบอร์มาจากผู้ให้บริการคลาวด์, SaaS, MSPs และซัพพลายเออร์โอเพนซอร์ส ([Supply Chain Security](https://en.wikipedia.org/wiki/Supply_chain_security), [NIST SP 800-161 Rev. 1](https://csrc.nist.gov/pubs/sp/800/161/r1/final)):

- **การประเมินและการตรวจสอบสถานะผู้จำหน่าย (Vendor Assessment & Due Diligence):** ตรวจรับรองอิสระ **SOC 2 Type II / ISO 27001** ของคู่ค้า ใช้แบบสอบถามความปลอดภัย ตรวจรายงานสรุป Penetration Testing ประเมินความเสี่ยง **Supply Chain** ต่อช่วงที่สี่ (Fourth-party) และคัดกรอง **Conflict of Interest** ก่อนจัดซื้อจัดจ้าง
- **ข้อตกลงปฏิบัติและมาตรการคุ้มครองทางสัญญา:**
  - **NDA (Non-Disclosure Agreement):** ปกป้องทรัพย์สินทางปัญญาที่เป็นความลับและผลการตรวจช่องโหว่ที่แลกเปลี่ยนระหว่างการเจรจา
  - **MOU / MOA (Memorandum of Understanding / Agreement):** แสดงเจตจำนงความร่วมมือเบื้องต้น (`MOU`) หรือข้อตกลงปฏิบัติการอย่างเป็นทางการ (`MOA`)
  - **MSA (Master Service Agreement) & SOW (Statement of Work):** กำหนดกรอบกฎหมายหลัก (`MSA`) และรายละเอียดส่งมอบ/กรอบเวลาเฉพาะงาน (`SOW`)
  - **SLA (Service Level Agreement):** ระบุรับประกันเวลาพร้อมใช้งาน (`99.95%`), กรอบเวลาการแจ้งเหตุละเมิด และบทปรับเมื่อแก้ปัญหาล่าช้า
  - **BPA (Business Partners Agreement) & Right-to-Audit:** มอบอำนาจในการตรวจสอบความสอดคล้องและตรวจสอบความปลอดภัยตามรอบเวลา
- **การเฝ้าระวังคู่ค้าอย่างต่อเนื่อง:** ติดตามช่องโหว่ CVE ของผู้จำหน่าย, B2B Federation Tokens ที่ผิดปกติ และสุขอนามัย Credentials ระหว่างการเชื่อมต่อพันธมิตร

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ช่องทางการเชื่อมต่อของซัพพลายเออร์ที่ได้รับความเชื่อถือมักตกเป็นเป้าของการโจมตี **Trusted Relationship** ([MITRE ATT&CK T1199](https://attack.mitre.org/techniques/T1199/)) ทีม Blue Team จึงจำกัดบัญชีระยะไกลของคู่ค้าด้วยสิทธิ์แบบกำหนดเวลา นโยบายเข้าถึงตามเงื่อนไข และการเฝ้าระวังพฤติกรรมบน B2B Integration Logs:

```kql
SigninLogs
| where UserType == "Guest" or AppDisplayName has "B2B"
| summarize Logins = count(), DistinctIPs = dcount(IPAddress) by UserPrincipalName, AppDisplayName
| order by Logins desc
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [NIST SP 800-161 Rev. 1 — Cybersecurity Supply Chain Risk Management](https://csrc.nist.gov/pubs/sp/800/161/r1/final)
- [Wikipedia — Supply chain security](https://en.wikipedia.org/wiki/Supply_chain_security)
- [MITRE ATT&CK — Trusted Relationship (T1199)](https://attack.mitre.org/techniques/T1199/)
- [Stanford University UIT — Information Security](https://uit.stanford.edu/security)
