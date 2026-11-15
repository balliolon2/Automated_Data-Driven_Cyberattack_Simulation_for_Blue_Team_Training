# 3.2 Apply security principles to secure enterprise infrastructure

## 1. Core Security+ Theory

Securing enterprise infrastructure enforces [Defense in Depth](https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29) across network perimeters, internal trust boundaries, and device hardware:

- **Device Placement & Security Zones:**
  - **[Screened Subnets / DMZ](https://en.wikipedia.org/wiki/DMZ_%28computing%29):** Isolates externally reachable servers (web reverse proxies, mail relays) between dual firewall policies so external compromise cannot pivot directly to internal Active Directory or database tiers.
  - **Microsegmentation & VLANs:** Restricts **east–west lateral movement** between internal hosts inside the same subnet via host firewalls or SDN policies.
- **Security Appliances & Enforcement Points:**
  - **[Firewalls](https://en.wikipedia.org/wiki/Firewall_%28computing%29):** Stateless packet filters inspect header 5-tuples; **stateful** firewalls track active TCP/UDP connection tables; **NGFWs / UTMs** perform deep Layer-7 application inspection; **WAFs** inspect HTTP/S requests for SQLi/XSS payloads.
  - **[IDS / IPS](https://en.wikipedia.org/wiki/Intrusion_detection_system):** Signature- and anomaly-based sensors operating out-of-band (`IDS` monitor/alert) or inline (`IPS` active drop/block).
  - **Forward / Reverse Proxies, Load Balancers & Jump Servers (Bastion Hosts):** Terminate TLS centrally, distribute availability load, enforce outbound filtering, and funnel privileged admin SSH/RDP access through auditable bastion points.
  - **Network Access Control (`NAC` / 802.1X):** Verifies device identity and posture health (patch state, EDR running) prior to granting full VLAN access.
- **Secure Access & Edge Architecture:** IPsec/WireGuard **VPNs**, **SD-WAN**, **SASE (Secure Access Service Edge)**, and **ZTNA** replace implicit trust inside corporate LANs with per-session identity and device verification.

## 2. SOC & Blue-Team Operational Practice

Blue teams correlate proxy, firewall, and IPS telemetry to catch attackers routing C2 beacons through non-standard proxies ([MITRE ATT&CK T1090](https://attack.mitre.org/techniques/T1090/)) or probing internal ports via **Remote Services** ([T1021](https://attack.mitre.org/techniques/T1021/)):

```kql
CommonSecurityLog
| where DeviceAction in ("Deny", "Drop", "Reset")
| summarize BlockedConnections = count(), DestPorts = make_set(DestinationPort, 10) by SourceIP, DestinationIP
| where array_length(DestPorts) >= 8 or BlockedConnections >= 100
```

## 3. References (Web Sources)

- [Wikipedia — Defense in depth (computing)](https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29)
- [Wikipedia — DMZ (computing)](https://en.wikipedia.org/wiki/DMZ_%28computing%29)
- [Wikipedia — Firewall (computing)](https://en.wikipedia.org/wiki/Firewall_%28computing%29)
- [Wikipedia — Intrusion detection system](https://en.wikipedia.org/wiki/Intrusion_detection_system)
- [MITRE ATT&CK — Proxy (T1090)](https://attack.mitre.org/techniques/T1090/)
