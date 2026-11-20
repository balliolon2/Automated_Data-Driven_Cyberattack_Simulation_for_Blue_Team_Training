# 1.4 อธิบายความสำคัญของการเลือกใช้โซลูชันการเข้ารหัสลับอย่างเหมาะสม (Explain the importance of using appropriate cryptographic solutions)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การเลือกใช้เทคโนโลยีการเข้ารหัสลับ (Cryptography) อย่างเหมาะสมช่วยบังคับใช้ Confidentiality, Integrity และ Authenticity ตลอดทุกสถานะข้อมูล วงจรชีวิตกุญแจ และฐานรากความน่าเชื่อถือระดับฮาร์ดแวร์ ([Public-Key Cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography), [PKI](https://en.wikipedia.org/wiki/Public_key_infrastructure), [HSM](https://en.wikipedia.org/wiki/Hardware_security_module)):

- **การเข้ารหัสแบบสมมาตร เทียบกับ แบบอสมมาตร (Symmetric vs. Asymmetric Encryption):**
  - อัลกอริทึมแบบ **Symmetric** (AES-128/256, ChaCha20) ใช้กุญแจลับชุดเดียวกันสำหรับการเข้ารหัสข้อมูลจัดเก็บปริมาณมาก (Full Disk Encryption, การเข้ารหัสฐานข้อมูล) และข้อมูลภายใน TLS Session
  - อัลกอริทึมแบบ **Asymmetric** (RSA, ECC, การแลกเปลี่ยนกุญแจ Diffie–Hellman / ECDHE) ใช้คู่กุญแจ Public/Private Key ที่เชื่อมโยงกันทางคณิตศาสตร์สำหรับสร้าง Digital Signatures, ยืนยัน Non-Repudiation และแลกเปลี่ยน Session Key อย่างปลอดภัย
- **การทำแฮชและการปกป้องรหัสผ่าน (Hashing & Password Protection):** ฟังก์ชันแฮชทางเดียว (SHA-256, SHA-3) ใช้ตรวจสอบ Integrity ของไฟล์และขับเคลื่อนการพิสูจน์ตัวตนแบบ HMAC สำหรับรหัสผ่านต้องไม่เก็บด้วยแฮชเปล่าเด็ดขาด แต่ต้องใส่ **Salt** แบบสุ่มที่ไม่ซ้ำกันร่วมกับ Key-Stretching KDFs (Argon2, bcrypt, PBKDF2) เพื่อป้องกัน Rainbow Tables และการเดาสุ่มด้วย GPU
- **โครงสร้างพื้นฐานกุญแจสาธารณะ (Public Key Infrastructure — PKI):**
  - **Certificate Authority (CA)** ทำหน้าที่ผูก Public Key เข้ากับตัวตนที่ตรวจสอบแล้วในใบรับรอง **X.509 certificates** และตรวจสอบความน่าเชื่อถือเป็นสายโซ่ขึ้นไปยัง Trusted Root CA
  - **การเพิกถอนใบรับรอง (Revocation):** ใบรับรองที่รั่วไหลหรือหมดอายุจะถูกทำให้ใช้งานไม่ได้ผ่าน Certificate Revocation Lists (**CRL**) หรือ Online Certificate Status Protocol (**OCSP** / OCSP stapling) ทั้งนี้ **Key Escrow** ใช้กู้คืนกุญแจถอดรหัสที่สูญหายเพื่อความต่อเนื่องของธุรกิจ แต่โดยทั่วไปไม่นำกุญแจสำหรับลงนาม (Signing Keys) เข้า Escrow เพื่อรักษาคุณสมบัติ Non-Repudiation
- **ฐานรากความน่าเชื่อถือระดับฮาร์ดแวร์และการพรางข้อมูล (Hardware Roots of Trust & Obfuscation):**
  - **TPM** (ชิปบนเมนบอร์ดที่จัดเก็บกุญแจดิสก์และ Boot Integrity), **HSM** (อุปกรณ์ฮาร์ดแวร์ต้านทานการแกะแงะที่ปกป้อง CA Root Keys) และ **Key Management Systems (KMS)** ช่วยรักษาความลับของกุญแจ
  - **Steganography** ซ่อนข้อมูลลับไว้ในไฟล์สื่อ (ไม่ใช่การเข้ารหัสที่แท้จริง); **Tokenization** แทนที่ข้อมูลสำคัญ (เช่น เลขบัตร PAN) ด้วย Token ทดแทนที่ไม่สามารถย้อนกลับได้; **Data Masking** ปิดบังข้อมูลบางส่วนขณะแสดงผลหน้าจอ

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

สุขอนามัยด้านการเข้ารหัสจะล้มเหลวเมื่อองค์กรปล่อยให้มีใบรับรองหมดอายุ ไบนารีที่ลงนามด้วยตัวเอง (Self-Signed) หรือชุด TLS Ciphers ที่อ่อนแอหลงเหลืออยู่ ผู้โจมตีมักอาศัย Code-Signing Certificate ที่ถูกต้องหรือถูกขโมยมาใช้หลบเลี่ยงการตรวจสอบความน่าเชื่อถือ ([MITRE ATT&CK T1553](https://attack.mitre.org/techniques/T1553/)) หรือสั่งเข้ารหัสดิสก์ทั้งองค์กรเพื่อเรียกค่าไถ่ในรูปแบบ **Data Encrypted for Impact** ([T1486](https://attack.mitre.org/techniques/T1486/)) นักวิเคราะห์จึงต้องตรวจสอบ TLS Handshakes และพฤติกรรมโปรเซสที่สั่งเข้ารหัสผิดปกติ:

```kql
DeviceProcessEvents
| where ProcessCommandLine has_any ("vssadmin delete shadows", "cipher /w", "openssl enc")
| project Timestamp, DeviceName, AccountName, FileName, ProcessCommandLine
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Public-key cryptography](https://en.wikipedia.org/wiki/Public-key_cryptography)
- [Wikipedia — Public key infrastructure](https://en.wikipedia.org/wiki/Public_key_infrastructure)
- [Wikipedia — Hardware security module](https://en.wikipedia.org/wiki/Hardware_security_module)
- [MITRE ATT&CK — Subvert Trust Controls (T1553)](https://attack.mitre.org/techniques/T1553/)
- [MITRE ATT&CK — Data Encrypted for Impact (T1486)](https://attack.mitre.org/techniques/T1486/)
