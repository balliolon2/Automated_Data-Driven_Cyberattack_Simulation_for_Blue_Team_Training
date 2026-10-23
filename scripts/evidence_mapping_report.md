# Evidence Mapping Review

Batch: 15 scenarios · Model: z-ai/glm-5.3-flash · Prompt: evidence-mapping-v1
Drafted by: llm · Generated: 2026-10-22T09:06:13Z
Mapping SHA-256: not computed (unapproved draft)
Status: PENDING APPROVAL

## Coverage summary

| Scenario | Verdict | Findings | Evidence indices |
|---|---|---|---|
| 1. Multi-Factor Authentication Bypass (Push Fatigue) | TRUE POSITIVE | 2 | #15, #16, #17, #18, #19 |
| 2. Atypical Travel VPN Connection Alert | FALSE POSITIVE | 2 | #16, #17, #18 |
| 3. Unauthorized Privileged Access / DACL Abuse | TRUE POSITIVE | 2 | #16, #17, #18 |
| 4. SQL Injection & Database Exfiltration | TRUE POSITIVE | 2 | #15, #16, #17, #18 |
| 5. Internal Port Scan and SYN Flood Alarm | FALSE POSITIVE | 2 | #0, #25, #26 |
| 6. Ransomware Execution & Volume Shadow Copy Deletion | TRUE POSITIVE | 2 | #15, #16, #17, #18 |
| 7. VLAN Segmentation Failure / Guest Network Bypass | TRUE POSITIVE | 2 | #15, #16, #17 |
| 8. Exorbitant Bandwidth Consumption on Guest Network | FALSE POSITIVE | 1 | #0, #25, #26 |
| 9. AWS S3 Cloud Storage Misconfiguration | TRUE POSITIVE | 2 | #15, #16, #17, #18 |
| 10. DNS Tunneling / Data Exfiltration via DNS TXT Queries | TRUE POSITIVE | 2 | #15, #16, #17, #18 |
| 11. Anomalous Volume of NTP Network Requests | FALSE POSITIVE | 1 | #0, #1, #34, #35, #36 |
| 12. Spear Phishing leading to Credential Harvesting | TRUE POSITIVE | 2 | #15, #16, #17, #18 |
| 13. Insider Threat: Customer Database Exfiltration | TRUE POSITIVE | 2 | #15, #16, #17 |
| 14. Anomalous Cloud Data Archiving Alert | FALSE POSITIVE | 1 | #15, #16, #17 |
| 15. Rogue Wireless Access Point Detection | TRUE POSITIVE | 2 | #15, #17, #18 |

## Warnings

- Scenario 'Unauthorized Privileged Access / DACL Abuse': log index 16 is shared evidence for findings d1_f5, d1_f6
- Scenario 'Exorbitant Bandwidth Consumption on Guest Network': False Positive has no change/HR/scheduled/approved log in its evidence; the benign context is the observed authorized activity itself.
- Scenario 'DNS Tunneling / Data Exfiltration via DNS TXT Queries': log index 15 is shared evidence for findings d4_f1, d4_f2
- Scenario 'DNS Tunneling / Data Exfiltration via DNS TXT Queries': log index 18 is shared evidence for findings d4_f1, d4_f2
- Scenario 'Insider Threat: Customer Database Exfiltration': log index 15 is shared evidence for findings d5_f1, d5_f2

## 1. Multi-Factor Authentication Bypass (Push Fatigue) — TRUE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: TRUE POSITIVE
Evidence union: 5 / 36 logs (14%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 1. Multi-Factor Authentication Bypass (Push Fatigue) (TP) | d1_f1 — Identified push fatigue indicators | #15 2026-07-13T08:03:00Z okta/warning — "MFA push sent to user 'ajones' - rejected by user"<br>#16 2026-07-13T08:03:10Z okta/warning — "MFA push sent to user 'ajones' - rejected by user"<br>#17 2026-07-13T08:03:20Z okta/warning — "MFA push sent to user 'ajones' - rejected by user" | Three consecutive MFA push notifications were rejected by the user within 30 seconds, indicating a classic push fatigue or brute-force attempt against the user 'ajones'. | malicious |
| 1. Multi-Factor Authentication Bypass (Push Fatigue) (TP) | d1_f2 — Identified suspicious login source IP | #18 2026-07-13T08:03:30Z okta/high — "MFA push sent to user 'ajones' - approved"<br>#19 2026-07-13T08:03:40Z ad_controller/high — "Admin session established with hijacked session token" | The MFA push was successfully approved from the external IP 198.51.100.42, immediately followed by an admin session establishment using a hijacked token from the same suspicious source. | malicious |

Flags: none

## 2. Atypical Travel VPN Connection Alert — FALSE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: FALSE POSITIVE
Evidence union: 3 / 34 logs (9%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 2. Atypical Travel VPN Connection Alert (FP) | d1_f3 — Discovered travel confirmation in HR logs | #18 2026-07-13T08:03:30Z hr_portal/info — "Out of Office / Travel notification logged for user mroberts: Paris Summit (Jul 12 - Jul 18)" **[BENIGN CONTEXT]** | The HR portal log confirms that user mroberts is traveling to Paris for a corporate summit, justifying the atypical location alert. | benign |
| 2. Atypical Travel VPN Connection Alert (FP) | d1_f4 — Verified device posture and MFA completion | #16 2026-07-13T08:03:10Z vpn_gateway/info — "MFA prompt completed successfully via SMS verification code"<br>#17 2026-07-13T08:03:20Z vpn_gateway/info — "Device posture check passed: Company asset Tag #4288" | The VPN gateway logs confirm that the user successfully completed SMS MFA and connected using a verified corporate asset. | benign |

Flags: none

## 3. Unauthorized Privileged Access / DACL Abuse — TRUE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: TRUE POSITIVE
Evidence union: 3 / 34 logs (9%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 3. Unauthorized Privileged Access / DACL Abuse (TP) | d1_f5 — Identified unauthorized DACL modification | #16 2026-07-13T08:03:10Z ad_controller/high — "Object permissions modified: DACL changed on CN=Payroll_Confidential,OU=Shares by user 'svc_backup'" | The AD controller log shows the backup service account 'svc_backup' modifying the DACL on the restricted Payroll_Confidential object, an unauthorized permission change by a non-administrative service account. | malicious |
| 3. Unauthorized Privileged Access / DACL Abuse (TP) | d1_f6 — Correlated DACL change with target user file access | #16 2026-07-13T08:03:10Z ad_controller/high — "Object permissions modified: DACL changed on CN=Payroll_Confidential,OU=Shares by user 'svc_backup'"<br>#17 2026-07-13T08:03:20Z file_server/high — "Access granted: read folder 'Payroll_Confidential' user 'jsmith'"<br>#18 2026-07-13T08:03:30Z file_server/warning — "File read: 'salary_matrix_2026.xlsx' by user 'jsmith'" | The DACL change by 'svc_backup' at 08:03:10 is immediately followed by 'jsmith' being granted access to Payroll_Confidential and reading salary_matrix_2026.xlsx, directly correlating the change with its exploitation. | malicious |

Flags:
- Scenario 'Unauthorized Privileged Access / DACL Abuse': log index 16 is shared evidence for findings d1_f5, d1_f6

## 4. SQL Injection & Database Exfiltration — TRUE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: TRUE POSITIVE
Evidence union: 4 / 34 logs (12%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 4. SQL Injection & Database Exfiltration (TP) | d2_f1 — Found SQL injection payload in query | #15 2026-07-13T08:03:00Z nginx/warning — "GET /users/search?q=1%%27%%20UNION%%20SELECT%%20username,password_hash%%20FROM%%20users-- HTTP/1.1 200 OK"<br>#16 2026-07-13T08:03:10Z postgresql/warning — "Query: SELECT * FROM products WHERE name LIKE '%1' UNION SELECT username,password_hash FROM users--%'" | The nginx log shows external client 198.51.100.77 submitting a UNION SELECT injection targeting username and password_hash, and the PostgreSQL log confirms the injected query executed against the users table. | malicious |
| 4. SQL Injection & Database Exfiltration (TP) | d2_f2 — Identified anomalous data volume exfiltration | #17 2026-07-13T08:03:20Z postgresql/warning — "Database output payload size: 1.2MB returned to webapp server"<br>#18 2026-07-13T08:03:30Z nginx/high — "GET /users/search - Response size 1205300 bytes HTTP/1.1 200 OK" | The database returned a 1.2MB payload and nginx logged a 1,205,300-byte response delivered to external IP 198.51.100.77, directly evidencing anomalous data volume exfiltration immediately after the injection. | malicious |

Flags: none

## 5. Internal Port Scan and SYN Flood Alarm — FALSE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: FALSE POSITIVE
Evidence union: 3 / 42 logs (7%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 5. Internal Port Scan and SYN Flood Alarm (FP) | d2_f3 — Correlated scanner IP with change management log | #26 2026-07-13T08:01:10Z change_management/info — "CR-4099 Approved: Authorized internal vulnerability scanning from test subnet (10.0.99.0/24) scheduled Jul 13 08:00 - 12:00" **[BENIGN CONTEXT]** | Change management log CR-4099 confirms an approved internal vulnerability scan from 10.0.99.0/24 during this window, providing the authorized benign explanation for the alerts. | benign |
| 5. Internal Port Scan and SYN Flood Alarm (FP) | d2_f4 — Confirmed scanning behavior is internal-only | #0 2026-07-13T08:00:00Z nids/medium — "TCP Port scan detected from 10.0.99.12 to 10.0.1.0 on Port 80"<br>#25 2026-07-13T08:01:00Z nids/medium — "SYN Flood Scan from 10.0.99.12 targeting Active Directory" | NIDS alerts show scanning originates only from internal host 10.0.99.12 toward internal targets, matching the authorized scanner subnet and supporting the false positive determination. | supporting |

Flags: none

## 6. Ransomware Execution & Volume Shadow Copy Deletion — TRUE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: TRUE POSITIVE
Evidence union: 4 / 34 logs (12%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 6. Ransomware Execution & Volume Shadow Copy Deletion (TP) | d2_f5 — Identified shadow copy deletion commands | #15 2026-07-13T08:03:00Z edr/high — "Command line execution: 'vssadmin.exe delete shadows /all /quiet'"<br>#16 2026-07-13T08:03:10Z edr/high — "Command line execution: 'wmic shadowcopy delete'" | Commands 'vssadmin.exe delete shadows /all /quiet' and 'wmic shadowcopy delete' executed by user mclark are hallmark ransomware anti-recovery actions that directly identify shadow copy deletion. | malicious |
| 6. Ransomware Execution & Volume Shadow Copy Deletion (TP) | d2_f6 — Discovered ransomware file signature modifications | #17 2026-07-13T08:03:20Z file_integrity/critical — "Rapid file modification: 145 files renamed to extension '.locked' in 3 seconds"<br>#18 2026-07-13T08:03:30Z edr/critical — "Process block: 'ransom_decrypt_instruction.txt' created on desktop" | The rapid renaming of 145 files to the '.locked' extension plus creation of 'ransom_decrypt_instruction.txt' demonstrates ransomware file encryption and signature modification on WS-FIN-08. | malicious |

Flags: none

## 7. VLAN Segmentation Failure / Guest Network Bypass — TRUE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: TRUE POSITIVE
Evidence union: 3 / 34 logs (9%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 7. VLAN Segmentation Failure / Guest Network Bypass (TP) | d3_f1 — Discovered unauthorized VLAN routing path | #15 2026-07-13T08:03:00Z firewall/high — "Inbound TCP connection allowed: 192.168.10.145:49810 -> 10.0.50.12:5432 (DB-Server-Prod)" | The firewall log confirms that an inbound TCP connection was allowed from the Guest VLAN IP (192.168.10.145) directly to the production database server (10.0.50.12). | malicious |
| 7. VLAN Segmentation Failure / Guest Network Bypass (TP) | d3_f2 — Identified database brute-force signatures | #16 2026-07-13T08:03:10Z postgresql/warning — "Database login failed for user 'sa' from host 192.168.10.145"<br>#17 2026-07-13T08:03:20Z postgresql/warning — "Database login failed for user 'admin' from host 192.168.10.145" | Multiple consecutive database login failures for common users ('sa' and 'admin') originating from the guest network IP indicate a brute-force attack attempt. | malicious |

Flags: none

## 8. Exorbitant Bandwidth Consumption on Guest Network — FALSE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: FALSE POSITIVE
Evidence union: 3 / 37 logs (8%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 8. Exorbitant Bandwidth Consumption on Guest Network (FP) | d3_f3 — Correlated traffic with video streaming domains | #0 2026-07-13T08:00:00Z proxy/info — "CONNECT youtube.com:443 HTTP/1.1"<br>#26 2026-07-13T08:02:20Z proxy/info — "CONNECT netflix.com:443 HTTP/1.1"<br>#25 2026-07-13T08:02:10Z nids/medium — "Excessive outbound bandwidth usage: 192.168.10.12 on Port 443" | Proxy logs show the high-volume 443 traffic from 192.168.10.12 is CONNECT requests to youtube.com and netflix.com, confirming the NIDS bandwidth alert is legitimate guest video streaming, not malicious exfiltration. | benign |

Flags:
- Scenario 'Exorbitant Bandwidth Consumption on Guest Network': False Positive has no change/HR/scheduled/approved log in its evidence; the benign context is the observed authorized activity itself.

## 9. AWS S3 Cloud Storage Misconfiguration — TRUE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: TRUE POSITIVE
Evidence union: 4 / 34 logs (12%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 9. AWS S3 Cloud Storage Misconfiguration (TP) | d3_f4 — Identified S3 policy alteration | #15 2026-07-13T08:03:00Z aws_cloudtrail/high — "PutBucketAcl: S3Bucket=customer-contracts-prod Policy=PublicReadOwnerAuthorizedAll" | This log shows the S3 bucket ACL was modified to allow public read access, enabling the subsequent data exfiltration. | malicious |
| 9. AWS S3 Cloud Storage Misconfiguration (TP) | d3_f5 — Found anonymous download logs | #16 2026-07-13T08:03:10Z aws_cloudtrail/critical — "ListObjects: S3Bucket=customer-contracts-prod"<br>#17 2026-07-13T08:03:20Z aws_cloudtrail/critical — "GetObject: S3Bucket=customer-contracts-prod Key=contract_99420.pdf"<br>#18 2026-07-13T08:03:30Z aws_cloudtrail/critical — "GetObject: S3Bucket=customer-contracts-prod Key=contract_99421.pdf" | These logs confirm that an anonymous external IP address (198.51.100.99) accessed and downloaded sensitive contract files from the publicly exposed bucket. | malicious |

Flags: none

## 10. DNS Tunneling / Data Exfiltration via DNS TXT Queries — TRUE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: TRUE POSITIVE
Evidence union: 4 / 34 logs (12%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 10. DNS Tunneling / Data Exfiltration via DNS TXT Queries (TP) | d4_f1 — Identified DNS TXT anomalies | #15 2026-07-13T08:03:00Z dns_server/warning — "Query: aGVs...bG8.malicious-tunnel.com Type=TXT"<br>#16 2026-07-13T08:03:10Z dns_server/warning — "Query: d29y...bGQ.malicious-tunnel.com Type=TXT"<br>#17 2026-07-13T08:03:20Z dns_server/warning — "Query: c2Vj...dXJl.malicious-tunnel.com Type=TXT"<br>#18 2026-07-13T08:03:30Z dns_server/warning — "Query: cGFz...c3dk.malicious-tunnel.com Type=TXT" | The burst of TXT queries with long base64-encoded subdomains from 10.0.1.44 is the anomalous DNS pattern characteristic of tunneling-based data exfiltration. | malicious |
| 10. DNS Tunneling / Data Exfiltration via DNS TXT Queries (TP) | d4_f2 — Discovered the rogue exfiltration domain name | #15 2026-07-13T08:03:00Z dns_server/warning — "Query: aGVs...bG8.malicious-tunnel.com Type=TXT"<br>#18 2026-07-13T08:03:30Z dns_server/warning — "Query: cGFz...c3dk.malicious-tunnel.com Type=TXT" | The encoded TXT queries explicitly target the external domain malicious-tunnel.com, identifying the rogue exfiltration domain receiving the tunneled data. | malicious |

Flags:
- Scenario 'DNS Tunneling / Data Exfiltration via DNS TXT Queries': log index 15 is shared evidence for findings d4_f1, d4_f2
- Scenario 'DNS Tunneling / Data Exfiltration via DNS TXT Queries': log index 18 is shared evidence for findings d4_f1, d4_f2

## 11. Anomalous Volume of NTP Network Requests — FALSE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: FALSE POSITIVE
Evidence union: 5 / 37 logs (14%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 11. Anomalous Volume of NTP Network Requests (FP) | d4_f3 — Correlated traffic with scheduled time-sync change | #0 2026-07-13T08:00:00Z firewall/medium — "High frequency NTP request permitted: 10.0.3.15 -> time.nist.gov:123 UDP"<br>#1 2026-07-13T08:00:01Z firewall/medium — "High frequency NTP request permitted: 10.0.3.15 -> time.nist.gov:123 UDP"<br>#34 2026-07-13T08:00:34Z firewall/medium — "High frequency NTP request permitted: 10.0.3.15 -> time.nist.gov:123 UDP"<br>#35 2026-07-13T08:00:40Z edr/info — "Process activity: w32time.exe syncing time configuration loop"<br>#36 2026-07-13T08:00:45Z change_management/info — "CR-3212: Configure NTP sync interval to 1s on time-server-backup due to oscillator failure" **[BENIGN CONTEXT]** | Consecutive firewall logs show a one-second NTP cadence matching CR-3212's authorized 1s sync interval, while the w32time sync loop and the sysadmin's approved change record confirm the traffic is a benign clock-drift fix. | benign |

Flags: none

## 12. Spear Phishing leading to Credential Harvesting — TRUE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: TRUE POSITIVE
Evidence union: 4 / 34 logs (12%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 12. Spear Phishing leading to Credential Harvesting (TP) | d4_f4 — Identified lookalike phishing domain | #15 2026-07-13T08:03:00Z email_gateway/warning — "Inbound email: Sender=admin@secure-office365-portal.com Recipient=kwilliams@corp.com Subject='Urgent: MFA Reset Required'"<br>#16 2026-07-13T08:03:10Z proxy/high — "Outbound HTTP connection: 10.0.2.77 -> secure-office365-portal.com:80"<br>#17 2026-07-13T08:03:20Z edr/high — "Credential theft attempt signature match: input fields detected on unsecured form" | The lookalike sender domain delivered the MFA reset lure, the user's proxy connection to that domain, and the EDR credential theft alert together prove the phishing and harvesting chain. | malicious |
| 12. Spear Phishing leading to Credential Harvesting (TP) | d4_f5 — Found unauthorized admin account creation | #18 2026-07-13T08:03:30Z ad_controller/critical — "New admin account registered: 'adm_backdoor' created by local machine WS-HR-11 privilege process" | The domain controller log records creation of the unauthorized 'adm_backdoor' admin account originating from the compromised host WS-HR-11, directly evidencing attacker persistence. | malicious |

Flags: none

## 13. Insider Threat: Customer Database Exfiltration — TRUE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: TRUE POSITIVE
Evidence union: 3 / 34 logs (9%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 13. Insider Threat: Customer Database Exfiltration (TP) | d5_f1 — Identified anomalous CRM data export | #15 2026-07-13T08:03:00Z crm_app/high — "Bulk report exported: Customer_Billing_Details_All.csv"<br>#17 2026-07-13T08:03:20Z edr/warning — "File deleted: Customer_Billing_Details_All.csv from desktop" | rharper bulk-exported the entire customer billing database outside normal operations and then deleted the local CSV from WS-SALES-15, showing deliberate concealment of the anomalous export. | malicious |
| 13. Insider Threat: Customer Database Exfiltration (TP) | d5_f2 — Correlated export with personal webmail exfiltration | #15 2026-07-13T08:03:00Z crm_app/high — "Bulk report exported: Customer_Billing_Details_All.csv"<br>#16 2026-07-13T08:03:10Z proxy/high — "Outbound HTTP POST to mail.personal-gmail.com bytes_sent=4508200" | The bulk CRM export is immediately followed by a 4.5MB HTTP POST to a personal webmail domain from rharper's host, directly correlating the export with exfiltration of customer data. | malicious |

Flags:
- Scenario 'Insider Threat: Customer Database Exfiltration': log index 15 is shared evidence for findings d5_f1, d5_f2

## 14. Anomalous Cloud Data Archiving Alert — FALSE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: FALSE POSITIVE
Evidence union: 3 / 33 logs (9%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 14. Anomalous Cloud Data Archiving Alert (FP) | d5_f3 — Correlated backup agent with compliance schedule | #15 2026-07-13T08:03:00Z aws_cloudtrail/medium — "High volume data synchronization: S3 bucket 'audit-logs-archive' copied to Glacier storage"<br>#16 2026-07-13T08:03:10Z change_management/info — "CR-9022 Approved: Standard monthly compliance audit archiving job scheduled Jul 13 08:30" **[BENIGN CONTEXT]**<br>#17 2026-07-13T08:03:20Z aws_cloudtrail/info — "Glacier transfer complete: 45.2 GB archived successfully" | The high-volume Glacier transfer by backup_agent_04 is directly correlated with the approved CR-9022 monthly compliance archiving change record, establishing the authorized benign explanation. | benign |

Flags: none

## 15. Rogue Wireless Access Point Detection — TRUE POSITIVE

Status: DRAFTED (llm, 1 attempt(s))
TP/FP Decision: TRUE POSITIVE
Evidence union: 3 / 34 logs (9%)

| Scenario | Finding | Evidence logs | Rationale | Support |
|---|---|---|---|---|
| 15. Rogue Wireless Access Point Detection (TP) | d5_f4 — Identified port security violation | #15 2026-07-13T08:03:00Z switch_core/high — "Port Security violation: 12 MAC addresses detected on single port interface Gi1/0/22 (Port allowed 1 MAC)"<br>#18 2026-07-13T08:03:30Z switch_core/critical — "MAC spoofing signatures matched on switch interface Gi1/0/22" | The switch logs confirm a Port Security violation due to multiple MAC addresses on Gi1/0/22, which is further corroborated by MAC spoofing signatures. | malicious |
| 15. Rogue Wireless Access Point Detection (TP) | d5_f5 — Discovered rogue wireless SSID | #17 2026-07-13T08:03:20Z ids_wifi/high — "Unauthorized SSIDs detected broadcasting: 'CORP-GUEST-FAST' broadcasting on channel 6, MAC 00:e0:4c:00:a1:05" | The wireless IDS detected an unauthorized SSID named 'CORP-GUEST-FAST' broadcasting from the compromised device. | malicious |

Flags: none
