# 4.4 Explain security alerting and monitoring concepts and tools

## 1. Core Security+ Theory

Continuous security monitoring collects, normalizes, and correlates cross-stack telemetry inside a [Security Operations Center (SOC)](https://en.wikipedia.org/wiki/Security_operations_center) powered by a [SIEM](https://en.wikipedia.org/wiki/Security_information_and_event_management):

- **Monitoring Targets & Telemetry Sources:**
  - **System, Application & Authentication Logs:** Centralized **Syslog**, Windows Event Logs (`4624/4625/4688`), cloud control-plane trails, and web access logs synchronized via **NTP** for accurate timeline reconstruction.
  - **Network Flows & Packet Capture (`PCAP`):** **NetFlow / IPFIX** records metadata (source/dest IPs, ports, byte volumes) while `PCAP` captures full payloads for deep inspection.
  - **Endpoint Agents & File Integrity Monitoring (`FIM`):** Real-time cryptographic checksum verification of sensitive files (`/etc/passwd`, `System32`) and process creation events.
- **SIEM Core Capabilities:**
  - **Log Aggregation & Normalization:** Parsing heterogeneous schemas into common fields (`SrcIP`, `Account`, `Action`) across collectors/sensors.
  - **Correlation Rules & Alert Tuning:** Joining multi-source signals while suppressing known benign baselines to reduce alert fatigue (`False Positives`) without losing sensitivity (`False Negatives`).
  - **User and Entity Behavior Analytics (`UEBA`):** Statistical anomaly detection spotting off-hours data staging, impossible travel, or rare parent-child process chains.
  - **Threat Hunting & Dashboards:** Proactive hypothesis-driven searches for stealthy **Command and Control (`C2`)** ([MITRE ATT&CK T1071](https://attack.mitre.org/techniques/T1071/)) bypassing signature alerts.

## 2. SOC & Blue-Team Operational Practice

Analysts rely on KQL pipelines across SIEM tables to correlate authentication anomalies with subsequent application-layer C2 callbacks ([MITRE ATT&CK T1071](https://attack.mitre.org/techniques/T1071/)) using synchronized UTC timestamps:

```kql
CommonSecurityLog
| where ApplicationProtocol in ("HTTP", "HTTPS", "DNS")
| summarize ConnectionCount = count(), AvgBytesOut = avg(SentBytes), StdevInterval = stdev(bin(TimeGenerated, 1m))
            by SourceIP, DestinationIP, DestinationPort
| where ConnectionCount >= 50 and StdevInterval < 0.5
```

## 3. References (Web Sources)

- [Wikipedia — Security operations center](https://en.wikipedia.org/wiki/Security_operations_center)
- [Wikipedia — Security information and event management](https://en.wikipedia.org/wiki/Security_information_and_event_management)
- [Wikipedia — Intrusion detection system](https://en.wikipedia.org/wiki/Intrusion_detection_system)
- [Microsoft Learn — KQL Overview](https://learn.microsoft.com/en-us/kusto/query/)
- [MITRE ATT&CK — Application Layer Protocol (T1071)](https://attack.mitre.org/techniques/T1071/)
