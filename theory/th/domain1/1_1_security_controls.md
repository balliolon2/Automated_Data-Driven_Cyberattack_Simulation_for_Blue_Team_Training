# 1.1 เปรียบเทียบและแยกแยะประเภทของมาตรการควบคุมความปลอดภัย (Compare and contrast various types of security controls)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

มาตรการควบคุมความปลอดภัย (Security Controls) คือกลไกป้องกันที่นำมาใช้บริหารจัดการความเสี่ยงทั้งทางไซเบอร์และกายภาพ ครอบคลุม Confidentiality, Integrity และ Availability ตามกรอบของ [NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final) และหลักการจัดหมวดหมู่ [Security Controls](https://en.wikipedia.org/wiki/Security_controls) โดยแบ่งตาม 2 มิติหลักดังนี้:

1. **หมวดหมู่ของมาตรการควบคุม (Control Categories — จำแนกตามกลไกการติดตั้ง):**
   - **Technical / Logical (เชิงเทคนิค):** บังคับใช้ผ่านซอฟต์แวร์ ฮาร์ดแวร์ หรือเฟิร์มแวร์ (เช่น Firewall, Multi-Factor Authentication [MFA], การเข้ารหัส TLS/AES, ระบบ EDR และ Access Control Lists [ACLs])
   - **Managerial / Administrative (เชิงการบริหารจัดการ):** การประเมินความเสี่ยง (Risk Assessments), กรอบธรรมาภิบาลความมั่นคงปลอดภัย, การออกนโยบาย (Policies), การตรวจสอบผู้ให้บริการภายนอก และแผนการจัดการช่องโหว่
   - **Operational (เชิงปฏิบัติการ):** ขั้นตอนการปฏิบัติงานประจำวันที่มีบุคลากรเป็นผู้ดำเนินการหลัก (เช่น การอบรม Security Awareness, การรับมือเหตุการณ์ Incident Response, การทบทวนค่าคอนฟิก และการดูแลผู้มาติดต่อ)
   - **Physical (เชิงกายภาพ):** มาตรการป้องกันสถานที่และทรัพย์สินทางกายภาพ (เช่น เครื่องอ่านบัตรผ่าน, ระบบล็อกไบโอเมตริกซ์, เสากั้น [Bollards], Mantraps, เจ้าหน้าที่รักษาความปลอดภัย, กล้อง CCTV และกรงฟาราเดย์ [Faraday Cages])
2. **ประเภทของมาตรการควบคุม (Control Types — จำแนกตามหน้าที่การทำงาน):**
   - **Preventive (เชิงป้องกัน):** หยุดยั้งกิจกรรมที่ไม่ได้รับอนุญาตก่อนเกิดผลกระทบ (เช่น ACL ตามหลัก Least Privilege, การตรวจ Validation ข้อมูลเข้า และการบล็อกของ IPS)
   - **Deterrent (เชิงป้องปราม):** ยับยั้งให้ผู้ไม่หวังดีลังเลหรือล้มเลิกก่อนลงมือ (เช่น ข้อความ Warning Banners, กล้องที่มองเห็นได้ชัดเจน และประกาศดำเนินคดีตามกฎหมาย)
   - **Detective (เชิงตรวจจับ):** ตรวจพบและแจ้งเตือนเหตุผิดปกติหรือการกระทำที่ไม่ได้รับอนุญาตระหว่างหรือหลังเกิดเหตุ (เช่น กฎ SIEM Correlation Alerts, เซนเซอร์ IDS, File Integrity Monitoring [FIM] และสัญญาณเตือนความเคลื่อนไหว)
   - **Corrective (เชิงแก้ไขฟื้นฟู):** กู้คืนระบบที่ได้รับผลกระทบหรือจำกัดความเสียหายหลังตรวจพบเหตุ (เช่น การติดตั้งแพตช์ช่องโหว่, การแยกเครื่องที่ติดมัลแวร์ออกจากเครือข่าย [Isolate Host] และการกู้คืนจาก Backup ที่สะอาด)
   - **Compensating (เชิงชดเชย):** จัดหามาตรการทางเลือกมาทดแทนเมื่อยังไม่สามารถใช้มาตรการหลักได้ชั่วคราว (เช่น การใช้ Virtual Patching บน WAF เพื่อคุ้มครองแอปพลิเคชันเก่าที่ยังไม่สามารถแพตช์ได้)
   - **Directive (เชิงกำหนดแนวทาง):** สั่งการและกำหนดพฤติกรรมที่ต้องปฏิบัติตามอย่างชัดเจนของบุคลากรหรือระบบ (เช่น นโยบาย Acceptable Use Policy, ป้ายเขตหวงห้าม และเอกสาร SOP Checklists)

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

ในกระบวนการทำงานของ SOC แนวคิด Defense-in-Depth จะเชื่อมโยงสัญญาณข้ามหลายชั้นควบคุม เพื่อให้ความล้มเหลวของมาตรการเชิง **Preventive** ถูกจับได้อย่างรวดเร็วด้วย telemetry เชิง **Detective** และถูกยับยั้งด้วย Playbook เชิง **Corrective** ตัวอย่างเช่น หากผู้โจมตีหลบเลี่ยง MFA ที่เป็น Preventive Control ด้วยการขโมย Token ([MITRE ATT&CK T1556](https://attack.mitre.org/techniques/T1556/)) กฎของ SIEM จะตรวจจับ Impossible Travel หรือการยิงคำขอ MFA ผิดปกติ (`Detective`) เพื่อให้นักวิเคราะห์เพิกถอน Token และตัดเซสชันได้ทันท่วงที (`Corrective`):

```kql
SigninLogs
| where ResultType != "0" or RiskLevelDuringSignIn in ("medium", "high")
| summarize FailedCount = count(), DistinctIPs = dcount(IPAddress) by UserPrincipalName, AppDisplayName
| where FailedCount >= 5 or DistinctIPs >= 3
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [NIST SP 800-53 Rev. 5 — Security and Privacy Controls](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [Wikipedia — Security controls](https://en.wikipedia.org/wiki/Security_controls)
- [Wikipedia — Access control](https://en.wikipedia.org/wiki/Access_control)
- [MITRE ATT&CK — Modify Authentication Process (T1556)](https://attack.mitre.org/techniques/T1556/)
- [Microsoft Learn — KQL Overview](https://learn.microsoft.com/en-us/kusto/query/)
