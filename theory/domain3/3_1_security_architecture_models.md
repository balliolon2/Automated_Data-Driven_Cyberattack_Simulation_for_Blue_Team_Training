# 3.1 Compare and contrast security architecture models

## 1. Core Security+ Theory

Modern security architecture distributes workloads across on-premises, cloud, containerized, and operational technology environments while balancing security boundary control against scalability ([Cloud Computing Security](https://en.wikipedia.org/wiki/Cloud_computing_security), [Infrastructure as Code](https://en.wikipedia.org/wiki/Infrastructure_as_code)):

- **Cloud & Shared Responsibility Models:**
  - **IaaS (Infrastructure as a Service):** Provider secures physical hardware and virtualization; customer patches OS, configures VPC firewalls, IAM, and data encryption.
  - **PaaS (Platform as a Service):** Provider manages OS and runtime; customer secures application code, API authentication, and data.
  - **SaaS (Software as a Service):** Provider manages stack; customer enforces MFA, conditional access, RBAC, and tenant configuration.
  - **Deployment Models:** Public, Private, Hybrid, and Multi-cloud (or community clouds), paired with **Cloud Access Security Brokers (`CASB`)** for SaaS visibility and DLP.
- **Serverless, Microservices & Containers:**
  - **Microservices** decompose monoliths into loosely coupled services communicating over mutual TLS (`mTLS`) and API gateways.
  - **Containers & Orchestration:** Isolate workloads at OS-namespace level without full guest kernels; security relies on minimal immutable images, image signing, runtime policies, and non-root execution.
- **[Infrastructure as Code (IaC)](https://en.wikipedia.org/wiki/Infrastructure_as_code) & Software-Defined Networking (`SDN`):** Declarative templates (Terraform, CloudFormation) enforce repeatable baselines and version-controlled drift detection across cloud VPC topologies.
- **On-Premises, Centralized/Decentralized & OT/IoT/Embedded:** Industrial Control Systems (**ICS/SCADA**), embedded RTOS devices, and IoT sensors prioritize **safety and availability** (`CIA` inverted to `AIC`), requiring strict **IT/OT segmentation** because embedded firmware cannot run standard EDR agents.

## 2. SOC & Blue-Team Operational Practice

Cloud and hybrid architectures shift monitoring toward control-plane audit logs (`CloudTrail`, `AzureActivity`, Kubernetes Audit Logs). Adversaries exploiting stolen API tokens attempt **Cloud Storage Object Discovery & Access** ([MITRE ATT&CK T1530](https://attack.mitre.org/techniques/T1530/)) or unauthorized IaC role escalation:

```kql
AzureActivity
| where OperationNameValue has_any ("MICROSOFT.AUTHORIZATION/ROLEASSIGNMENTS/WRITE", "MICROSOFT.STORAGE/STORAGEACCOUNTS/LISTKEYS/ACTION")
| where ActivityStatusValue == "Success"
| project TimeGenerated, Caller, CallerIpAddress, OperationNameValue, ResourceGroup
```

## 3. References (Web Sources)

- [Wikipedia — Cloud computing security](https://en.wikipedia.org/wiki/Cloud_computing_security)
- [Wikipedia — Infrastructure as code](https://en.wikipedia.org/wiki/Infrastructure_as_code)
- [NIST SP 800-207 — Zero Trust Architecture](https://csrc.nist.gov/pubs/sp/800/207/final)
- [CISA — Zero Trust Maturity Model](https://www.cisa.gov/zero-trust-maturity-model)
- [MITRE ATT&CK — Data from Cloud Storage (T1530)](https://attack.mitre.org/techniques/T1530/)
