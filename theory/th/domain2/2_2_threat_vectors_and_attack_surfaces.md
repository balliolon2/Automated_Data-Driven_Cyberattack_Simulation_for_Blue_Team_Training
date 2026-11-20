# 2.2 วิเคราะห์เวกเตอร์ภัยคุกคามและพื้นผิวการโจมตีที่พบบ่อย (Analyze common threat vectors and attack surfaces)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

พื้นผิวการโจมตีขององค์กร ([Attack Surface](https://en.wikipedia.org/wiki/Attack_surface)) ครอบคลุมทุกพอร์ตเครือข่ายที่เปิดเผย ซอฟต์แวร์ปลายทาง พอร์ทัลพิสูจน์ตัวตน การเชื่อมต่อห่วงโซ่อุปทาน และกระบวนการทำงานของบุคลากรที่ผู้โจมตีสามารถเข้าถึงได้ ([Phishing](https://en.wikipedia.org/wiki/Phishing), [Social Engineering](https://en.wikipedia.org/wiki/Social_engineering_%28security%29), [CISA Phishing Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)):

- **เวกเตอร์ผ่านข้อความและฟิชชิ่ง (Message-Based & Phishing Vectors):** Phishing แบบส่งวงกว้าง, **Spear Phishing** แบบเจาะจงเป้าหมาย, **Whaling** ที่มุ่งเป้าผู้บริหาร, **Vishing** ผ่านเสียงโทรศัพท์, **Smishing** ผ่าน SMS, การปลอมอีเมลทางธุรกิจ (**BEC**) และโดเมนปลอมที่สะกดคล้ายคลึง (**Typosquatting**)
- **หลักการวิศวกรรมสังคม (Social Engineering Principles):** การสร้างเรื่องราวหลอกลวง (Pretexting), การอ้างอำนาจผู้บริหาร, การสร้างความเร่งด่วน/ขาดแคลน, การสร้างความคุ้นเคย และการอ้างความเห็นพ้องของกลุ่ม เพื่อบายพาสมาตรการทางเทคนิคผ่านการคล้อยตามของมนุษย์
- **เวกเตอร์ทางเทคนิคและระบบที่ขาดการป้องกัน (Technical & Unsecured Vectors):**
  - บริการที่เปิดโดยไม่ปลอดภัย รหัสผ่านตั้งต้นจากโรงงาน พอร์ต RDP/VPN ที่ไม่ได้รับการป้องกัน และจุดเชื่อมต่อไร้สายปลอม (Evil Twin)
  - สื่อบันทึกข้อมูลถอดเสียบที่ขาดการควบคุม (**Removable Media** เช่น USB drop attacks) รวมถึงลิงก์อันตรายและการโจมตีแบบ Watering-Hole
  - **เวกเตอร์ห่วงโซ่อุปทาน (Supply Chain Vectors):** ไลบรารีซอฟต์แวร์ที่ถูกสอดไส้ ผู้ให้บริการดูแลระบบภายนอก (MSPs) ที่ถูกยึดครอง และการปลอมแปลง CI/CD build pipelines ([NIST SP 800-161 Rev. 1](https://csrc.nist.gov/pubs/sp/800/161/r1/final))
- **เวกเตอร์จากมนุษย์และการตั้งค่าเริ่มต้น:** การใช้ Shadow IT บนคลาวด์นอกเหนือการกำกับดูแล และบัญชีผู้รับเหมาที่หมดสัญญาแต่ยังไม่ถูกเพิกถอน

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

Email Gateway, Proxy Logs และ Endpoint Telemetry ใช้ตรวจจับระยะการเข้าถึงระบบเริ่มต้น ([MITRE ATT&CK T1566 — Phishing](https://attack.mitre.org/techniques/T1566/)) และการเจาะแอปพลิเคชันที่เปิดสู่สาธารณะ ([T1190](https://attack.mitre.org/techniques/T1190/)) นักวิเคราะห์จะไล่เรียงลำดับเหตุการณ์ตั้งแต่ URL ในอีเมลที่ถูกคลิกไปจนถึงการแตกโพรเซส PowerShell หรือ Office Macro:

```kql
EmailUrlInfo
| join kind=inner (UrlClickEvents | where ActionType == "ClickAllowed") on NetworkMessageId
| join kind=leftouter (DeviceProcessEvents) on $left.AccountUpn == $right.InitiatingProcessAccountUpn
| where DeviceProcessEvents.FileName in~ ("powershell.exe", "cmd.exe", "mshta.exe")
| project Url, AccountUpn, DeviceName, FileName, ProcessCommandLine
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Attack surface](https://en.wikipedia.org/wiki/Attack_surface)
- [Wikipedia — Phishing](https://en.wikipedia.org/wiki/Phishing)
- [Wikipedia — Social engineering (security)](https://en.wikipedia.org/wiki/Social_engineering_%28security%29)
- [CISA — Phishing Threat & Response Guidance](https://www.cisa.gov/topics/cyber-threats-and-responses/phishing)
- [NIST SP 800-161 Rev. 1 — Cybersecurity Supply Chain Risk Management](https://csrc.nist.gov/pubs/sp/800/161/r1/final)
- [MITRE ATT&CK — Phishing (T1566)](https://attack.mitre.org/techniques/T1566/)
