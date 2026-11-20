# 3.1 เปรียบเทียบและแยกแยะแบบจำลองสถาปัตยกรรมความปลอดภัย (Compare and contrast security architecture models)

## 1. ทฤษฎีหลัก CompTIA Security+ (Core Security+ Theory)

สถาปัตยกรรมความปลอดภัยสมัยใหม่กระจายภาระงานข้ามสภาพแวดล้อม On-Premises, Cloud, Containers และ Operational Technology โดยต้องรักษาสมดุลระหว่างการควบคุมขอบเขตความปลอดภัยกับความยืดหยุ่นในการขยายระบบ ([Cloud Computing Security](https://en.wikipedia.org/wiki/Cloud_computing_security), [Infrastructure as Code](https://en.wikipedia.org/wiki/Infrastructure_as_code)):

- **แบบจำลองคลาวด์และความรับผิดชอบร่วมกัน (Cloud & Shared Responsibility Models):**
  - **IaaS (Infrastructure as a Service):** ผู้ให้บริการดูแลฮาร์ดแวร์กายภาพและระบบ Virtualization; ลูกค้าต้องแพตช์ OS, ตั้งค่า VPC Firewalls, IAM และเข้ารหัสข้อมูล
  - **PaaS (Platform as a Service):** ผู้ให้บริการดูแล OS และ Runtime; ลูกค้าดูแลความปลอดภัยของโค้ดแอปพลิเคชัน การยืนยันตัวตน API และข้อมูล
  - **SaaS (Software as a Service):** ผู้ให้บริการดูแลทั้งสแตก; ลูกค้าบังคับใช้ MFA, Conditional Access, RBAC และการตั้งค่าผู้เช่า (Tenant Configuration)
  - **Deployment Models:** Public, Private, Hybrid และ Multi-Cloud (หรือ Community Clouds) ร่วมกับการใช้ **Cloud Access Security Brokers (`CASB`)** เพื่อเพิ่มการมองเห็นและ DLP สำหรับแอป SaaS
- **Serverless, Microservices & Containers:**
  - **Microservices** แยกโครงสร้าง Monolith ออกเป็นบริการย่อยที่สื่อสารกันผ่าน Mutual TLS (`mTLS`) และ API Gateways
  - **Containers & Orchestration:** แยก Workload ที่ระดับ OS-Namespace โดยไม่ใช้ Guest Kernel เต็มรูปแบบ ความปลอดภัยจึงอาศัย Minimal Immutable Images, Image Signing, นโยบาย Runtime และการทำงานแบบ Non-Root
- **[Infrastructure as Code (IaC)](https://en.wikipedia.org/wiki/Infrastructure_as_code) & Software-Defined Networking (`SDN`):** เทมเพลตเชิงประกาศ (Terraform, CloudFormation) ช่วยบังคับใช้ Baseline ที่ทำซ้ำได้และตรวจจับ Configuration Drift ในระดับ Version Control บน Cloud VPC
- **On-Premises, Centralized/Decentralized & OT/IoT/Embedded:** ระบบควบคุมอุตสาหกรรม (**ICS/SCADA**), อุปกรณ์ RTOS และเซ็นเซอร์ IoT ให้ความสำคัญสูงสุดกับ **ความปลอดภัยต่อชีวิตและความพร้อมใช้งาน** (`CIA` กลับลำดับเป็น `AIC`) จึงต้องแบ่งแยก **IT/OT Segmentation** อย่างเคร่งครัดเพราะเฟิร์มแวร์ฝังตัวไม่สามารถติดตั้ง EDR Agent มาตรฐานได้

## 2. การปฏิบัติงานจริงใน SOC และ Blue Team (SOC & Blue-Team Operational Practice)

สถาปัตยกรรมแบบคลาวด์และไฮบริดย้ายจุดเน้นการเฝ้าระวังมายัง Control-Plane Audit Logs (`CloudTrail`, `AzureActivity`, Kubernetes Audit Logs) เมื่อผู้โจมตีขโมย API Token มักพยายามค้นหาและดึงข้อมูลจาก Cloud Storage ([MITRE ATT&CK T1530](https://attack.mitre.org/techniques/T1530/)) หรือยกระดับสิทธิ์ใน IaC โดยไม่ได้รับอนุญาต:

```kql
AzureActivity
| where OperationNameValue has_any ("MICROSOFT.AUTHORIZATION/ROLEASSIGNMENTS/WRITE", "MICROSOFT.STORAGE/STORAGEACCOUNTS/LISTKEYS/ACTION")
| where ActivityStatusValue == "Success"
| project TimeGenerated, Caller, CallerIpAddress, OperationNameValue, ResourceGroup
```

## 3. แหล่งอ้างอิง (References / Web Sources)

- [Wikipedia — Cloud computing security](https://en.wikipedia.org/wiki/Cloud_computing_security)
- [Wikipedia — Infrastructure as code](https://en.wikipedia.org/wiki/Infrastructure_as_code)
- [NIST SP 800-207 — Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final)
- [CISA — Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model)
- [MITRE ATT&CK — Data from Cloud Storage (T1530)](https://attack.mitre.org/techniques/T1530/)
