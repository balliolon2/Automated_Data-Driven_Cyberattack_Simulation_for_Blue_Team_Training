# 2.4 วิเคราะห์ตัวบ่งชี้กิจกรรมที่เป็นอันตราย (Analyze indicators of malicious activity)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การวิเคราะห์ตัวบ่งชี้การบุกรุก ([Indicators of Compromise (IoCs)](https://en.wikipedia.org/wiki/Indicator_of_compromise)) จำเป็นต้องสังเกตเห็นร่องรอยของ [Malware](https://en.wikipedia.org/wiki/Malware), การใช้ Credential โดยมิชอบ และการโจมตีเครือข่ายก่อนที่จะเกิดผลกระทบเต็มรูปแบบ:

- **ประเภทของมัลแวร์ (Malware Categories):**
  - **[Ransomware](https://en.wikipedia.org/wiki/Ransomware):** เข้ารหัสไฟล์หรือล็อกระบบเพื่อเรียกค่าไถ่ (มักพบร่วมกับการขโมยข้อมูลมากดดันซ้ำแบบ Double Extortion)
  - **Trojans & RATs:** แฝงตัวเป็นซอฟต์แวร์ปกติเพื่อเปิดช่องทางควบคุมระยะไกล (Backdoor) ให้ผู้โจมตี
  - **Worms:** แพร่กระจายตัวเองผ่าน Network Shares หรือช่องโหว่โดยไม่ต้องอาศัยการโต้ตอบจากผู้ใช้
  - **Rootkits, Keyloggers, Spyware, Bloatware & Logic Bombs:** ซ่อนการฝังตัวลึกในระดับ OS/Kernel บันทึกการกดแป้นพิมพ์ หรือทำงานเมื่อเข้าเงื่อนไข/วันที่กำหนด
- **การโจมตีเครือข่ายและโปรโตคอล (Network & Protocol Attacks):**
  - **Denial of Service / DDoS:** การท่วมทราฟฟิกปริมาณมหาศาล, SYN floods หรือ Amplification เพื่อทำลายความพร้อมใช้งาน
  - **On-Path (MitM) & Layer-2 Poisoning:** ARP poisoning, DNS poisoning และ Rogue DHCP เพื่อดักจับหรือเบี่ยงเบนทราฟฟิก
  - **Replay, Downgrade & DNS Tunneling:** การส่งซ้ำ Token พิสูจน์ตัวตนที่ดักจับได้ หรือลักลอบส่งข้อมูล C2 แฝงใน DNS TXT/A queries
- **ร่องรอยการโจมตีรหัสผ่านและแอปพลิเคชัน (Password & Application Artifacts):**
  - **Brute Force, Dictionary, Password Spraying, Credential Stuffing, Pass-the-Hash:** ก่อให้เกิดบัญชีถูกล็อก การเข้าสู่ระบบจากต่างภูมิศาสตร์ที่เป็นไปไม่ได้พร้อมกัน หรือการเคลื่อนที่ข้ามระบบผ่าน NTLM ([MITRE ATT&CK T1003](https://attack.mitre.org/techniques/T1003/))
  - **Persistence & Log Anomalies:** การสร้าง Scheduled Tasks ที่ไม่ได้รับอนุญาต ([T1053](https://attack.mitre.org/techniques/T1053/)) การติดตั้งบริการแปลกปลอม การปิด AV/EDR หรือการลบ Event Logs ([T1070](https://attack.mitre.org/techniques/T1070/))

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

นักวิเคราะห์ SOC จะแยกแยะสัญญาณรบกวนทั่วไปของผู้ดูแลระบบออกจาก IoCs ที่เป็นอันตราย โดยเชื่อมโยงข้อมูล Telemetry บนโฮสต์ (แผนผังโพรเซสแม่-ลูก, การเปิดซ็อกเก็ตขาออกที่ผิดปกติ, Registry Run keys) เข้ากับข้อมูล Authentication และ DNS Logs:

```kql
DeviceNetworkEvents
| where RemotePort in (4444, 8080, 9001) or (ActionType == "ConnectionSuccess" and InitiatingProcessFileName in~ ("powershell.exe", "rundll32.exe", "regsvr32.exe"))
| project Timestamp, DeviceName, InitiatingProcessFileName, InitiatingProcessCommandLine, RemoteIP, RemotePort
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Indicator of compromise](https://en.wikipedia.org/wiki/Indicator_of_compromise)
- [Wikipedia — Malware](https://en.wikipedia.org/wiki/Malware)
- [Wikipedia — Ransomware](https://en.wikipedia.org/wiki/Ransomware)
- [MITRE ATT&CK — Indicator Removal (T1070)](https://attack.mitre.org/techniques/T1070/)
- [MITRE ATT&CK — Scheduled Task/Job (T1053)](https://attack.mitre.org/techniques/T1053/)
