# 2.5 Explain mitigation techniques used to secure the enterprise

## 1. Core Security+ Theory

Enterprise mitigation reduces both the likelihood and blast radius of successful exploitation across networks, hosts, and identities ([NIST SP 800-53 Rev. 5](https://csrc.nist.gov/pubs/sp/800/53/r5/final), [Network Segmentation](https://en.wikipedia.org/wiki/Network_segmentation)):

- **Access & Segmentation Mitigations:**
  - **[Network Segmentation](https://en.wikipedia.org/wiki/Network_segmentation) & Access Control Lists (ACLs):** Divides networks into trust zones (VLANs, screened subnets/DMZs) so compromise of a public web tier cannot freely reach internal databases ([MITRE ATT&CK T1021 — Remote Services](https://attack.mitre.org/techniques/T1021/)).
  - **Least Privilege:** Restricting user accounts, service tokens, and firewall rules to the minimum permissions needed.
- **Execution & Endpoint Mitigations:**
  - **Application Allowlisting:** Permitting only cryptographically signed or approved binaries to execute, blocking unknown ransomware/trojan payloads even without signatures.
  - **[Endpoint Detection and Response (EDR)](https://en.wikipedia.org/wiki/Endpoint_detection_and_response) & Host Firewalls:** Behavioral blocking, memory scanning, and isolation of compromised devices.
- **Configuration, Hardening & Patching:**
  - Applying secure configuration baselines, **disabling unnecessary ports/services**, changing default credentials, and turning off insecure legacy protocols (SMBv1, Telnet, SSLv3).
  - **Patch Management & Encryption:** Timelines driven by severity ([NIST SP 800-40 Rev. 4](https://csrc.nist.gov/pubs/sp/800/40/r4/final)) alongside encrypting sensitive volumes.
- **Containment & Lifecycle Isolation:** Quarantine VLANs, **sandboxing** untrusted attachments, **patching/isolating**, or **decommissioning** unpatchable EOL hardware.

## 2. SOC & Blue-Team Operational Practice

During active incident mitigation, Blue Team responders choose **proportional** controls matched to evidence: isolating a compromised workstation (`Corrective` containment), revoking stolen session tokens (`Access revocation`), blocking malicious C2 domains via DNS sinkholing, or deploying targeted firewall rules—rather than halting healthy production infrastructure unnecessarily:

```kql
DeviceEvents
| where ActionType in ("AsrExecutableBlockingAudited", "AsrLsassCredentialTheftBlocked", "AntivirusDetection")
| summarize Detections = count() by DeviceName, ActionType, FileName
| order by Detections desc
```

## 3. References (Web Sources)

- [Wikipedia — Network segmentation](https://en.wikipedia.org/wiki/Network_segmentation)
- [Wikipedia — Endpoint detection and response](https://en.wikipedia.org/wiki/Endpoint_detection_and_response)
- [NIST SP 800-40 Rev. 4 — Enterprise Patch Management Planning](https://csrc.nist.gov/pubs/sp/800/40/r4/final)
- [NIST SP 800-53 Rev. 5 — Security Controls](https://csrc.nist.gov/pubs/sp/800/53/r5/final)
- [MITRE ATT&CK — Remote Services (T1021)](https://attack.mitre.org/techniques/T1021/)
