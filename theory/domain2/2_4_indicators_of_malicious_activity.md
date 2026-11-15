# 2.4 Analyze indicators of malicious activity

## 1. Core Security+ Theory

Analyzing [Indicators of Compromise (IoCs)](https://en.wikipedia.org/wiki/Indicator_of_compromise) requires spotting artifacts of [Malware](https://en.wikipedia.org/wiki/Malware), credential abuse, and network attacks before full adversary impact:

- **Malware Categories:**
  - **[Ransomware](https://en.wikipedia.org/wiki/Ransomware):** Encrypts files or locks systems for extortion (often paired with double extortion data theft).
  - **Trojans & RATs:** Masquerade as legitimate software to establish remote backdoor control.
  - **Worms:** Self-replicate across network shares/vulnerabilities without human interaction.
  - **Rootkits, Keyloggers, Spyware, Bloatware, & Logic Bombs:** Hide persistence deep in OS/kernel layers, record keystrokes, or trigger on specific conditions/dates.
- **Network & Protocol Attacks:**
  - **Denial of Service / DDoS:** Volumetric floods, SYN floods, or amplification depleting availability.
  - **On-Path (MitM) & Layer-2 Poisoning:** ARP poisoning, DNS poisoning, and rogue DHCP hijacking traffic.
  - **Replay, Downgrade, & DNS Tunneling:** Re-submitting captured authentication tokens or smuggling C2 traffic inside DNS TXT/A queries.
- **Password & Application Artifacts:**
  - **Brute Force, Dictionary, Password Spraying, Credential Stuffing, Pass-the-Hash:** Trigger account lockouts, simultaneous logins from impossible geographies, or NTLM lateral movement ([MITRE ATT&CK T1003](https://attack.mitre.org/techniques/T1003/)).
  - **Persistence & Log Anomalies:** Unauthorized scheduled tasks ([T1053](https://attack.mitre.org/techniques/T1053/)), unexpected service installs, disabled AV/EDR, or cleared event logs ([T1070](https://attack.mitre.org/techniques/T1070/)).

## 2. SOC & Blue-Team Operational Practice

SOC analysts separate benign administrative noise from malicious IoCs by correlating host telemetry (parent-child process trees, unexpected outbound sockets, registry Run keys) with authentication and DNS logs:

```kql
DeviceNetworkEvents
| where RemotePort in (4444, 8080, 9001) or (ActionType == "ConnectionSuccess" and InitiatingProcessFileName in~ ("powershell.exe", "rundll32.exe", "regsvr32.exe"))
| project Timestamp, DeviceName, InitiatingProcessFileName, InitiatingProcessCommandLine, RemoteIP, RemotePort
```

## 3. References (Web Sources)

- [Wikipedia — Indicator of compromise](https://en.wikipedia.org/wiki/Indicator_of_compromise)
- [Wikipedia — Malware](https://en.wikipedia.org/wiki/Malware)
- [Wikipedia — Ransomware](https://en.wikipedia.org/wiki/Ransomware)
- [MITRE ATT&CK — Indicator Removal (T1070)](https://attack.mitre.org/techniques/T1070/)
- [MITRE ATT&CK — Scheduled Task/Job (T1053)](https://attack.mitre.org/techniques/T1053/)
