# 5.5 อธิบายประเภทและวัตถุประสงค์ของการตรวจประเมินและการทดสอบ (Explain types and purposes of audits and assessments)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การตรวจสอบอิสระช่วยยืนยันว่ามาตรการที่ออกแบบไว้ทำงานได้จริงในทางปฏิบัติ ([IT Security Audit](https://en.wikipedia.org/wiki/Information_technology_security_audit), [Penetration Test](https://en.wikipedia.org/wiki/Penetration_test), [NIST SP 800-115](https://csrc.nist.gov/pubs/sp/800/115/final)):

- **การตรวจสอบภายในเทียบกับการตรวจสอบภายนอก (Internal vs. External Audits):**
  - **Internal Audits & Self-Assessments:** ดำเนินการโดยคณะกรรมการตรวจสอบภายในหรือทีมความปลอดภัยเพื่อทวนสอบการปฏิบัติตามนโยบายและแก้ไขช่องว่างเชิงรุกก่อนการตรวจอย่างเป็นทางการ
  - **External Audits & Regulatory Examinations:** ดำเนินการโดยผู้สอบบัญชี/ผู้ประเมินอิสระภายนอก (เช่น SOC 2, หน่วยรับรอง ISO 27001, PCI QSA) หรือหน่วยงานกำกับของรัฐเพื่อให้คำรับรองที่เป็นกลาง
- **การทดสอบเจาะระบบและระดับข้อมูลที่ผู้ทดสอบทราบ ([NIST SP 800-115](https://csrc.nist.gov/pubs/sp/800/115/final)):**
  - **Unknown Environment (Black Box):** จำลองผู้โจมตีภายนอกโดยไม่มีข้อมูลสถาปัตยกรรมภายในมาก่อนเลย
  - **Partially Known Environment (Gray Box):** มอบสิทธิ์หรือแผนผังบางส่วนเพื่อเน้นงบประมาณการทดสอบไปยังระบบภายในที่มีมูลค่าสูง
  - **Known Environment (White/Crystal Box):** เปิดซอร์สโค้ด แผนผังเครือข่าย และสิทธิ์ทดสอบระดับสูงครบถ้วนเพื่อตรวจสอบอย่างทั่วถึง
  - **Reconnaissance:** แบบ **Passive** (OSINT, คิวรี DNS/WHOIS โดยไม่แตะต้องระบบเป้าหมายโดยตรง) เทียบกับแบบ **Active** (สแกนพอร์ต ทดสอบช่องโหว่)
  - **Physical Penetration Testing & Purple Teaming:** ทดสอบการปลอมบัตรผ่าน/ลักลอบตามหลังเข้าอาคาร ควบคู่กับแบบฝึกหัดร่วม Red vs Blue เพื่อปรับเทียบความแม่นยำการตรวจจับใน SIEM

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

การทดสอบเจาะระบบที่ได้รับอนุญาตต้องมีเอกสาร **Rules of Engagement (`RoE`)** ซึ่งระบุ IP ผู้ทดสอบ ช่วงเวลาทดสอบ และผู้ติดต่อฉุกเฉินบันทึกอยู่ใน Benign Context Logs เพื่อให้ทีม Blue Team แยกแยะการทดสอบที่ได้รับอนุมัติออกจากพฤติกรรม **Network Discovery / Valid Accounts** ที่เป็นอันตราย ([MITRE ATT&CK T1078](https://attack.mitre.org/techniques/T1078/)):

```kql
SecurityAlert
| where TimeGenerated > ago(24h)
| project TimeGenerated, AlertName, CompromisedEntity, Entities, RemediationSteps
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Information technology security audit](https://en.wikipedia.org/wiki/Information_technology_security_audit)
- [Wikipedia — Penetration test](https://en.wikipedia.org/wiki/Penetration_test)
- [NIST SP 800-115 — Technical Guide to Information Security Testing and Assessment](https://csrc.nist.gov/pubs/sp/800/115/final)
- [MITRE ATT&CK — Valid Accounts (T1078)](https://attack.mitre.org/techniques/T1078/)
