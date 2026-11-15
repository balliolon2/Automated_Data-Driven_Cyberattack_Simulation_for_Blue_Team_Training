# 4.1 Apply common security techniques to computing resources

## 1. Core Security+ Theory

Hardening individual servers, workstations, mobile devices, and wireless networks limits initial execution and post-exploitation persistence ([NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [Endpoint Detection and Response](https://en.wikipedia.org/wiki/Endpoint_detection_and_response)):

- **Secure Baselines & Endpoint Hardening:** Applying CIS Benchmarks or vendor baselines via Group Policy (`GPO`) or configuration management: removing unused services, closing unnecessary ports, changing default credentials, enabling **UEFI Secure Boot** (preventing bootkits), enforcing full-disk encryption, and activating OS mandatory access controls (`SELinux`, `AppArmor`).
- **Endpoint & Application Defense:** Running centrally managed **EDR**, anti-malware, and host-based firewalls, combined with application sandboxing and memory integrity protections.
- **Mobile Device Management (`MDM` & `MAM`):** Enforcing screen locks, remote wipe, jailbreak/root detection, containerization (separating work profiles from personal apps), and enterprise deployment models:
  - **BYOD (Bring Your Own Device):** Employee-owned; privacy boundaries require `MAM` containerization.
  - **COPE (Corporate-Owned, Personally Enabled):** Enterprise-owned, controlled by `MDM`, limited personal use allowed.
  - **CYOD / Corporate-Owned Business Only:** Maximum control and compliance restriction.
- **Wireless Security Hardening:** Securing enterprise Wi-Fi with **WPA3-Enterprise** (`802.1X / EAP-TLS` or `PEAP` authenticating against RADIUS) rather than shared PSKs, isolating guest SSIDs via captive portals and client isolation.

## 2. SOC & Blue-Team Operational Practice

Endpoint security posture telemetry confirms whether a suspicious process launched inside a hardened context or resulted from tampered configuration. Attackers regularly disable host defenses or modify registry/GPO baselines ([MITRE ATT&CK T1562 — Impair Defenses](https://attack.mitre.org/techniques/T1562/)) or persist via scheduled jobs ([T1053](https://attack.mitre.org/techniques/T1053/)):

```kql
DeviceRegistryEvents
| where ActionType == "RegistryValueSet"
| where RegistryKey has_any ("DisableAntiSpyware", "DisableRealtimeMonitoring", "CurrentVersion\\Run")
| project Timestamp, DeviceName, InitiatingProcessAccountName, RegistryKey, RegistryValueData
```

## 3. References (Web Sources)

- [Wikipedia — Endpoint detection and response](https://en.wikipedia.org/wiki/Endpoint_detection_and_response)
- [NIST SP 800-53 Rev. 5 — System and Information Integrity (SI)](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Impair Defenses (T1562)](https://attack.mitre.org/techniques/T1562/)
- [MITRE ATT&CK — Scheduled Task/Job (T1053)](https://attack.mitre.org/techniques/T1053/)
- [Microsoft Learn — KQL Overview](https://learn.microsoft.com/en-us/kusto/query/)
