# 4.4 อธิบายแนวคิดและเครื่องมือในการแจ้งเตือนและเฝ้าระวังความปลอดภัย (Explain security alerting and monitoring concepts and tools)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

การเฝ้าระวังความปลอดภัยอย่างต่อเนื่องทำหน้าที่รวบรวม ปรับรูปแบบ (Normalize) และเชื่อมโยง Telemetry จากทุกชั้นเทคโนโลยีภายในศูนย์ [Security Operations Center (SOC)](https://en.wikipedia.org/wiki/Security_operations_center) ซึ่งขับเคลื่อนด้วยระบบ [SIEM](https://en.wikipedia.org/wiki/Security_information_and_event_management):

- **เป้าหมายการเฝ้าระวังและแหล่ง Telemetry:**
  - **System, Application & Authentication Logs:** ระบบรวม **Syslog**, Windows Event Logs (`4624/4625/4688`), Cloud Control-Plane Trails และ Web Access Logs ที่ซิงโครไนซ์เวลาผ่าน **NTP** เพื่อเรียงไทม์ไลน์ได้อย่างถูกต้อง
  - **Network Flows & Packet Capture (`PCAP`):** **NetFlow / IPFIX** บันทึก Metadata (IP ต้นทาง/ปลายทาง พอร์ต ปริมาณไบต์) ขณะที่ `PCAP` เก็บเพย์โหลดเต็มรูปแบบเพื่อวิเคราะห์เชิงลึก
  - **Endpoint Agents & File Integrity Monitoring (`FIM`):** ตรวจเช็กค่า Checksum ทางคณิตศาสตร์แบบเรียลไทม์สำหรับไฟล์สำคัญ (`/etc/passwd`, `System32`) และเหตุการณ์สร้างโพรเซส
- **ความสามารถหลักของ SIEM:**
  - **Log Aggregation & Normalization:** แปลงโครงสร้างข้อมูลจากหลายเซ็นเซอร์ให้อยู่ในชื่อฟิลด์มาตรฐานกลาง (`SrcIP`, `Account`, `Action`)
  - **Correlation Rules & Alert Tuning:** เชื่อมโยงสัญญาณข้ามหลายแหล่งพร้อมตัดทอน Baseline ปกติเพื่อลด Alert Fatigue (`False Positives`) โดยไม่ลดความไวในการตรวจจับ (`False Negatives`)
  - **User and Entity Behavior Analytics (`UEBA`):** การตรวจจับความผิดปกติเชิงสถิติ เช่น การเตรียมข้อมูลนอกเวลาทำการ Impossible Travel หรือสายโพรเซสแม่-ลูกที่เกิดยาก
  - **Threat Hunting & Dashboards:** การค้นหาเชิงรุกตามสมมติฐานเพื่อดักจับช่องทาง **Command and Control (`C2`)** ([MITRE ATT&CK T1071](https://attack.mitre.org/techniques/T1071/)) ที่เล็ดลอดกฎ Signature ทั่วไป

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

นักวิเคราะห์ใช้คำสั่ง KQL ในตาราง SIEM เพื่อเชื่อมโยงความผิดปกติของการยืนยันตัวตนเข้ากับการส่งสัญญาณ C2 ระดับแอปพลิเคชัน ([MITRE ATT&CK T1071](https://attack.mitre.org/techniques/T1071/)) โดยอาศัยเวลา UTC ที่ตรงกันทุกอุปกรณ์:

```kql
CommonSecurityLog
| where ApplicationProtocol in ("HTTP", "HTTPS", "DNS")
| summarize ConnectionCount = count(), AvgBytesOut = avg(SentBytes), StdevInterval = stdev(bin(TimeGenerated, 1m))
            by SourceIP, DestinationIP, DestinationPort
| where ConnectionCount >= 50 and StdevInterval < 0.5
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Security operations center](https://en.wikipedia.org/wiki/Security_operations_center)
- [Wikipedia — Security information and event management](https://en.wikipedia.org/wiki/Security_information_and_event_management)
- [Wikipedia — Intrusion detection system](https://en.wikipedia.org/wiki/Intrusion_detection_system)
- [Microsoft Learn — KQL Overview](https://learn.microsoft.com/en-us/kusto/query/)
- [MITRE ATT&CK — Application Layer Protocol (T1071)](https://attack.mitre.org/techniques/T1071/)
