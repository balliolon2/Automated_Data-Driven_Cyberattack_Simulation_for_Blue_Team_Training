# 5.2 อธิบายองค์ประกอบของกระบวนการบริหารความเสี่ยง (Explain elements of the risk management process)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

[Risk Management](https://en.wikipedia.org/wiki/Risk_management) ประเมินว่าภัยคุกคามใช้ประโยชน์จากช่องโหว่เพื่อสร้างผลกระทบต่อทรัพย์สินทางธุรกิจอย่างไร เพื่อชี้นำการลงทุนด้านความปลอดภัยอย่างสมเหตุสมผล ([NIST SP 800-30 Rev. 1](https://csrc.nist.gov/pubs/sp/800/30/r1/final), [Business Impact Analysis](https://en.wikipedia.org/wiki/Business_impact_analysis)):

- **ความถี่ในการระบุและประเมินความเสี่ยง (Risk Assessment Cadence):** การเฝ้าระวังต่อเนื่อง การประเมินประจำปี การทบทวนเฉพาะกิจเมื่อเกิดเหตุ (หลังการรั่วไหลหรือควบรวมกิจการ) และการประเมินระดับโครงการครั้งเดียว
- **การวิเคราะห์ความเสี่ยงเชิงคุณภาพ vs. เชิงปริมาณ:**
  - **Qualitative (เชิงคุณภาพ):** ให้คะแนนโอกาสเกิดและผลกระทบเชิงเปรียบเทียบ (`Low / Medium / High`) ผ่าน Heat Maps และความเห็นผู้เชี่ยวชาญ
  - **Quantitative (เชิงปริมาณ):** คำนวณเป็นตัวเงินด้วย **Asset Value (`AV`)**, **Exposure Factor (`EF`)**, **Single Loss Expectancy (`SLE = AV × EF`)**, **Annualized Rate of Occurrence (`ARO`)** และ **Annualized Loss Expectancy (`ALE = SLE × ARO`)** โดยค่าใช้จ่ายของมาตรการควบคุมต่อปีจะคุ้มค่าเมื่อต่ำกว่ามูลค่า ALE ที่ลดลงได้
- **ระดับความเสี่ยงที่ยอมรับได้และการติดตาม:** บันทึกใน **Risk Register** ส่วนกลางที่ระบุ **Risk Owner**, Key Risk Indicators (`KRIs`), เกณฑ์ความเสี่ยง และระดับทัศนคติต่อความเสี่ยง (**Expansionary**, **Neutral**, **Conservative**)
- **กลยุทธ์จัดการความเสี่ยง (Risk Treatment Strategies):** **Mitigate** (วางมาตรการลดความเสี่ยง), **Transfer** (ทำประกันภัยไซเบอร์ หรือส่งมอบตาม SLA ให้ผู้ให้บริการ), **Accept** (บันทึกข้อยกเว้นความเสี่ยงอย่างเป็นทางการ) หรือ **Avoid** (ยุติบริการที่มีความเสี่ยงสูง)
- **[Business Impact Analysis (BIA)](https://en.wikipedia.org/wiki/Business_impact_analysis):** นิยามภารกิจหลักที่ห้ามหยุดชะงักควบคู่กับค่า **RTO** (Recovery Time Objective), **RPO** (Recovery Point Objective), **MTBF** และ **MTTR**

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

การจัดลำดับความสำคัญใน SOC จะผูกข่าวกรองช่องโหว่ CVSS และความสำคัญของเครื่องใน SIEM เข้ากับระดับชั้น **BIA** และ **Risk Register** เพื่อให้เซิร์ฟเวอร์ชำระเงินหรือระบบยืนยันตัวตนหลักได้รับการตอบสนองทันทีก่อนเครื่องทดสอบที่มีผลกระทบต่ำ:

```kql
SecurityAlert
| where AlertSeverity in ("High", "Critical")
| summarize AlertCount = count(), LatestTime = max(TimeGenerated) by CompromisedEntity, AlertName
| order by LatestTime desc
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Risk management](https://en.wikipedia.org/wiki/Risk_management)
- [Wikipedia — Business impact analysis](https://en.wikipedia.org/wiki/Business_impact_analysis)
- [NIST SP 800-30 Rev. 1 — Guide for Conducting Risk Assessments](https://csrc.nist.gov/pubs/sp/800/30/r1/final)
- [NIST — Cybersecurity Framework (CSF 2.0)](https://www.nist.gov/cyberframework)
