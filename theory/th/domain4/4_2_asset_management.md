# 4.2 อธิบายนัยด้านความปลอดภัยของการจัดการทรัพย์สินฮาร์ดแวร์ ซอฟต์แวร์ และข้อมูล (Explain security implications of hardware, software, and data asset management)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

คุณไม่สามารถปกป้องหรือเฝ้าระวังสิ่งที่คุณไม่รู้ว่ามีอยู่ได้ การจัดการทรัพย์สินอย่างเป็นระบบติดตามฮาร์ดแวร์ ซอฟต์แวร์ และข้อมูลตั้งแต่การจัดหาไปจนถึงการทำลายอย่างยืนยันได้ ([NIST SP 800-88 Rev. 1](https://csrc.nist.gov/pubs/sp/800/88/r1/final), [Data Sanitization](https://en.wikipedia.org/wiki/Data_sanitization), [Data Remanence](https://en.wikipedia.org/wiki/Data_remanence)):

- **การจัดหาและการจัดซื้อ (Acquisition & Procurement):** ประเมินมาตรฐานความปลอดภัยของผู้จำหน่ายและแหล่งที่มาในห่วงโซ่อุปทานก่อนนำฮาร์ดแวร์หรือบริการ SaaS เข้าใช้งาน
- **การกำหนดความรับผิดชอบ การติดตาม และคลังข้อมูล (CMDB):** ติดป้ายรหัสทรัพย์สินบนทรัพยากรกายภาพและคลาวด์ ระบุผู้ดูแล ความสำคัญทางธุรกิจ และรายการซอฟต์แวร์ (`SBOM` — Software Bill of Materials) เพื่อปิดจุดบอดจาก **Shadow IT** และติดตามระยะ **End-of-Life (`EOL`)**
- **การเฝ้าระวัง การจัดเก็บ และการจัดชั้นทรัพย์สิน:** กำหนดให้ Asset ความสำคัญสูงได้รับมาตรฐานการเก็บ Log ที่เข้มงวด ระยะเวลาแพตช์ที่สั้นลง และกรอบเวลาจัดเก็บข้อมูลตามกฎหมาย
- **การปลดระวางและการทำลายข้อมูลอย่างปลอดภัย ([NIST SP 800-88 Rev. 1](https://csrc.nist.gov/pubs/sp/800/88/r1/final)):**
  - **Clearing:** การเขียนทับพื้นที่จัดเก็บข้อมูลด้วยคำสั่งอ่าน/เขียนมาตรฐาน (`diskpart clean all`) สำหรับสื่อที่นำกลับมาใช้ซ้ำภายในระดับชั้นความลับเดิม
  - **Purging:** ใช้ **Cryptographic Erase** (ทำลายกุญแจเข้ารหัสทันทีทำให้ Ciphertext กู้คืนไม่ได้) หรือ **Degaussing** (สำหรับฮาร์ดดิสก์จานแม่เหล็ก/เทป — ใช้ไม่ได้ผลกับหน่วยความจำแฟลช/โซลิดสเตต `SSD`!) เพื่อป้องกันการกู้คืนในระดับห้องปฏิบัติการ
  - **Destruction:** การทำลายทางกายภาพด้วยการ **สับย่อย บดขยี้ ป่นละเอียด หรือเผา** สื่อที่ปลดระวาง พร้อมออกใบรับรองการทำลาย (**Certificate of Destruction**) ที่ลงนามสำหรับงานตรวจสอบ

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

คลังทรัพย์สิน (CMDB พร้อมแท็กเจ้าของ) ช่วยเร่งการประเมินขอบเขตเหตุการณ์เมื่อมีสัญญาณเตือน เช่น ทราบได้ทันทีว่าเครื่องเป้าหมายเป็นโน้ตบุ๊กผู้บริหาร, Domain Controller หรือเครื่องที่ปลดระวางแล้ว ซึ่งชี้ชะตาความเร่งด่วนในการยกระดับปัญหา การตรวจพบการติดตั้งซอฟต์แวร์ไม่ได้รับอนุญาตหรืออุปกรณ์ไม่มีการจัดการบนเครือข่ายภายในจะสั่งให้ NAC กักกันทันที:

```kql
DeviceNetworkInfo
| where Timestamp > ago(24h)
| summarize LastSeen = max(Timestamp) by DeviceId, DeviceName, IPAddresses
| join kind=leftanti (CMDBInventoryTable | project DeviceId) on DeviceId
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [NIST SP 800-88 Rev. 1 — Guidelines for Media Sanitization](https://csrc.nist.gov/pubs/sp/800/88/r1/final)
- [Wikipedia — Data sanitization](https://en.wikipedia.org/wiki/Data_sanitization)
- [Wikipedia — Data remanence](https://en.wikipedia.org/wiki/Data_remanence)
- [Stanford University UIT — Information Security](https://uit.stanford.edu/security)
- [NIST CSRC Glossary](https://csrc.nist.gov/glossary)
