# 4.2 Explain security implications of hardware, software, and data asset management

## 1. Core Security+ Theory

You cannot protect or monitor what you do not know exists. Structured asset management tracks hardware, software, and data from acquisition through verified disposal ([NIST SP 800-88 Rev. 1](https://csrc.nist.gov/pubs/sp/800/88/r1/final), [Data Sanitization](https://en.wikipedia.org/wiki/Data_sanitization), [Data Remanence](https://en.wikipedia.org/wiki/Data_remanence)):

- **Acquisition & Procurement:** Assessing vendor security compliance and supply-chain provenance prior to onboarding hardware or SaaS tools.
- **Assignment, Tracking & Inventory (CMDB):** Tagging physical and cloud resources with asset IDs, assigned owners, business criticality, and OS/software inventories (`SBOM` — Software Bill of Materials) to eliminate blind spots and **Shadow IT** and track **End-of-Life (`EOL`)** milestones.
- **Monitoring, Retention & Asset Classification:** Ensuring high-criticality assets receive stricter logging, faster patching SLAs, and legally mandated retention windows.
- **Decommissioning & Media Sanitization ([NIST SP 800-88 Rev. 1](https://csrc.nist.gov/pubs/sp/800/88/r1/final)):**
  - **Clearing:** Overwriting addressable storage blocks using standard read/write commands (`diskpart clean all`) for media reused internally at the same classification.
  - **Purging:** Applying **Cryptographic Erase** (destroying encryption keys immediately rendering ciphertext unrecoverable) or **degaussing** (for magnetic HDDs/tapes; ineffective on solid-state flash/`SSD`!) so laboratory recovery is infeasible.
  - **Destruction:** Physical **shredding, crushing, pulverizing, or incinerating** retired media, accompanied by a signed **Certificate of Destruction** for audit compliance.

## 2. SOC & Blue-Team Operational Practice

Asset inventories (CMDB + owner tags) accelerate incident scoping when an indicator hits: knowing immediately whether an alerting host is an executive workstation, a domain controller, or a decommissioned asset determines escalation urgency and triage priority. Detection of unauthorized software installations or unmanaged devices connecting to internal subnets triggers NAC containment:

```kql
DeviceNetworkInfo
| where Timestamp > ago(24h)
| summarize LastSeen = max(Timestamp) by DeviceId, DeviceName, IPAddresses
| join kind=leftanti (CMDBInventoryTable | project DeviceId) on DeviceId
```

## 3. References (Web Sources)

- [NIST SP 800-88 Rev. 1 — Guidelines for Media Sanitization](https://csrc.nist.gov/pubs/sp/800/88/r1/final)
- [Wikipedia — Data sanitization](https://en.wikipedia.org/wiki/Data_sanitization)
- [Wikipedia — Data remanence](https://en.wikipedia.org/wiki/Data_remanence)
- [Stanford University UIT — Information Security](https://uit.stanford.edu/security)
- [NIST CSRC Glossary](https://csrc.nist.gov/glossary)
