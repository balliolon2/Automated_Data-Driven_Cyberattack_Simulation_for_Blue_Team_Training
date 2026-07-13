import json
import random
from datetime import datetime, timedelta

def create_scenarios():
    scenarios = []

    # Helper to generate timestamps
    base_time = datetime(2026, 7, 13, 8, 0, 0)
    def t(offset_seconds):
        return (base_time + timedelta(seconds=offset_seconds)).isoformat() + "Z"

    # =========================================================================
    # DOMAIN 1: Security Concepts (domain1)
    # =========================================================================
    
    # 1. TP: MFA Session Hijacking / Push Fatigue
    logs_1 = []
    # Noise
    for i in range(15):
        logs_1.append({
            "timestamp": t(i * 10), "source": "ad_controller", "severity": "info", "severity_level": 1,
            "message": "User login successful", "src_ip": f"10.0.1.{10+i}", "user": f"user{i}", "host": f"PC-{100+i}"
        })
    # Evidence
    logs_1.extend([
        {"timestamp": t(180), "source": "okta", "severity": "warning", "severity_level": 3, "message": "MFA push sent to user 'ajones' - rejected by user", "src_ip": "198.51.100.42", "user": "ajones", "host": "External-IP"},
        {"timestamp": t(190), "source": "okta", "severity": "warning", "severity_level": 3, "message": "MFA push sent to user 'ajones' - rejected by user", "src_ip": "198.51.100.42", "user": "ajones", "host": "External-IP"},
        {"timestamp": t(200), "source": "okta", "severity": "warning", "severity_level": 3, "message": "MFA push sent to user 'ajones' - rejected by user", "src_ip": "198.51.100.42", "user": "ajones", "host": "External-IP"},
        {"timestamp": t(210), "source": "okta", "severity": "high", "severity_level": 4, "message": "MFA push sent to user 'ajones' - approved", "src_ip": "198.51.100.42", "user": "ajones", "host": "External-IP"},
        {"timestamp": t(220), "source": "ad_controller", "severity": "high", "severity_level": 4, "message": "Admin session established with hijacked session token", "src_ip": "198.51.100.42", "user": "ajones", "host": "DC-01"},
        {"timestamp": t(230), "source": "ad_controller", "severity": "high", "severity_level": 4, "message": "Password changed for user 'admin_backup'", "src_ip": "198.51.100.42", "user": "ajones", "host": "DC-01"}
    ])
    # More noise
    for i in range(15):
        logs_1.append({
            "timestamp": t(240 + i * 15), "source": "firewall", "severity": "info", "severity_level": 1,
            "message": "Outbound connection allowed", "src_ip": f"10.0.2.{50+i}", "dst_ip": "8.8.8.8", "action": "allow"
        })

    scenarios.append({
        "title": "Multi-Factor Authentication Bypass (Push Fatigue)",
        "description": "An alert is triggered for user 'ajones' experiencing multiple consecutive rejected MFA pushes within a minute, followed by a successful login from a new external IP address.",
        "domain_id": "domain1",
        "difficulty": 2,
        "is_true_positive": True,
        "tp_fp_explanation": "This is a True Positive. The multiple rejected MFA push notifications followed by a successful login indicate a classic MFA Push Fatigue attack, leading to session hijacking.",
        "initial_logs": logs_1,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d1_f1",
                    "description": "Identified push fatigue indicators",
                    "domain_id": "domain1",
                    "points": 10,
                    "explanation": "Multiple rejected MFA push requests prior to approval represent a push fatigue pattern."
                },
                {
                    "id": "d1_f2",
                    "description": "Identified suspicious login source IP",
                    "domain_id": "domain1",
                    "points": 10,
                    "explanation": "The login originated from external IP 198.51.100.42 which is not standard for ajones."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d1_c1", "label": "Revoke active Okta sessions for user ajones", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Terminates the attacker's hijacked session immediately."},
                {"id": "d1_c2", "label": "Disable user ajones AD account temporarily", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Prevents further domain access until investigated."},
                {"id": "d1_c3", "label": "Block external IP 198.51.100.42 at perimeter firewall", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Stops connections from the attacker's source IP."}
            ],
            "eradication": [
                {"id": "d1_e1", "label": "Enforce registration of FIDO2 WebAuthn keys for ajones", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "FIDO2 is resistant to push fatigue and phishing."},
                {"id": "d1_e2", "label": "Format the user's primary mobile device", "is_correct": False, "domain_id": "domain1", "points": -2, "explanation": "Unnecessary since push fatigue does not imply the physical device is compromised."}
            ],
            "recovery": [
                {"id": "d1_r1", "label": "Verify backup administrator passwords and revert unauthorized changes", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Reverts the password change done on admin_backup."},
                {"id": "d1_r2", "label": "Re-enable ajones AD account after resetting password", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Safely restores user access."}
            ]
        }
    })

    # 2. FP: VPN login from uncommon country (legitimate business trip)
    logs_2 = []
    for i in range(15):
        logs_2.append({
            "timestamp": t(i * 10), "source": "firewall", "severity": "info", "severity_level": 1,
            "message": "Outbound HTTP allow", "src_ip": "10.0.1.200", "dst_ip": "104.244.42.1"
        })
    logs_2.extend([
        {"timestamp": t(180), "source": "vpn_gateway", "severity": "warning", "severity_level": 2, "message": "VPN connection attempt from atypical location: France", "src_ip": "195.154.122.90", "user": "mroberts", "host": "VPN-Client"},
        {"timestamp": t(190), "source": "vpn_gateway", "severity": "info", "severity_level": 1, "message": "MFA prompt completed successfully via SMS verification code", "src_ip": "195.154.122.90", "user": "mroberts", "host": "VPN-Client"},
        {"timestamp": t(200), "source": "vpn_gateway", "severity": "info", "severity_level": 1, "message": "Device posture check passed: Company asset Tag #4288", "src_ip": "195.154.122.90", "user": "mroberts", "host": "VPN-Client"},
        {"timestamp": t(210), "source": "hr_portal", "severity": "info", "severity_level": 1, "message": "Out of Office / Travel notification logged for user mroberts: Paris Summit (Jul 12 - Jul 18)", "user": "mroberts"}
    ])
    for i in range(15):
        logs_2.append({
            "timestamp": t(220 + i * 10), "source": "file_share", "severity": "info", "severity_level": 1,
            "message": "Read access allowed for marketing_strategy.pdf", "user": "mroberts", "src_ip": "195.154.122.90"
        })

    scenarios.append({
        "title": "Atypical Travel VPN Connection Alert",
        "description": "The VPN gateway triggers a medium severity alert regarding an atypical login location for user 'mroberts' originating from France.",
        "domain_id": "domain1",
        "difficulty": 2,
        "is_true_positive": False,
        "tp_fp_explanation": "This is a False Positive. The user mroberts completed SMS MFA, used a verified corporate asset, and HR logs confirm they are traveling in Paris, France for a corporate summit.",
        "initial_logs": logs_2,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d1_f3",
                    "description": "Discovered travel confirmation in HR logs",
                    "domain_id": "domain5",
                    "points": 10,
                    "explanation": "HR logs verify the user is in Paris, France during this time."
                },
                {
                    "id": "d1_f4",
                    "description": "Verified device posture and MFA completion",
                    "domain_id": "domain1",
                    "points": 10,
                    "explanation": "MFA SMS was successfully entered and device posture checks passed."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d1_c4", "label": "Revoke session and lock user mroberts", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Unnecessary lockout; disrupts legitimate business operations."},
                {"id": "d1_c5", "label": "Allow connection and dismiss alert after verification", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Correct action since validation succeeded."}
            ],
            "eradication": [
                {"id": "d1_e3", "label": "Review Okta security logs for prior anomalies", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Ensures no prior malicious session activity."}
            ],
            "recovery": [
                {"id": "d1_r3", "label": "Update user travel status database to prevent future alerts", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Prevents alert fatigue by syncing travel plans."}
            ]
        }
    })

    # 3. TP: Privilege Escalation via DACL Modification
    logs_3 = []
    for i in range(15):
        logs_3.append({
            "timestamp": t(i * 10), "source": "ad_controller", "severity": "info", "severity_level": 1,
            "message": "Authentication package loaded", "user": f"worker{i}"
        })
    logs_3.extend([
        {"timestamp": t(180), "source": "file_server", "severity": "warning", "severity_level": 2, "message": "Access denied for folder 'Payroll_Confidential' user 'jsmith'", "user": "jsmith", "src_ip": "10.0.2.14"},
        {"timestamp": t(190), "source": "ad_controller", "severity": "high", "severity_level": 4, "message": "Object permissions modified: DACL changed on CN=Payroll_Confidential,OU=Shares by user 'svc_backup'", "user": "svc_backup"},
        {"timestamp": t(200), "source": "file_server", "severity": "high", "severity_level": 4, "message": "Access granted: read folder 'Payroll_Confidential' user 'jsmith'", "user": "jsmith", "src_ip": "10.0.2.14"},
        {"timestamp": t(210), "source": "file_server", "severity": "warning", "severity_level": 3, "message": "File read: 'salary_matrix_2026.xlsx' by user 'jsmith'", "user": "jsmith", "src_ip": "10.0.2.14"}
    ])
    for i in range(15):
        logs_3.append({
            "timestamp": t(220 + i * 10), "source": "firewall", "severity": "info", "severity_level": 1,
            "message": "Inbound connection allowed", "src_ip": f"10.0.1.{30+i}", "dst_ip": "10.0.2.14"
        })

    scenarios.append({
        "title": "Unauthorized Privileged Access / DACL Abuse",
        "description": "An alert is triggered when user 'jsmith' suddenly gains access to a restricted folder 'Payroll_Confidential' and reads sensitive files immediately following a DACL change by a system backup account.",
        "domain_id": "domain1",
        "difficulty": 3,
        "is_true_positive": True,
        "tp_fp_explanation": "This is a True Positive. The service account 'svc_backup' was used maliciously to alter the DACL of the restricted folder to grant 'jsmith' access, which was immediately exploited.",
        "initial_logs": logs_3,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d1_f5",
                    "description": "Identified unauthorized DACL modification",
                    "domain_id": "domain1",
                    "points": 10,
                    "explanation": "DACL modification by svc_backup bypassed normal change control processes."
                },
                {
                    "id": "d1_f6",
                    "description": "Correlated DACL change with target user file access",
                    "domain_id": "domain1",
                    "points": 10,
                    "explanation": "jsmith accessed the payroll files immediately after the permission change."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d1_c6", "label": "Revoke the modified DACL and restore standard permissions", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Restores confidentiality of payroll data."},
                {"id": "d1_c7", "label": "Lock AD account svc_backup", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Prevents further unauthorized directory modifications."},
                {"id": "d1_c8", "label": "Isolate work host 10.0.2.14 belonging to jsmith", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Prevents lateral movement or data exfiltration from host."}
            ],
            "eradication": [
                {"id": "d1_e4", "label": "Investigate svc_backup credential leakage", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Identifies how the service account was compromised."}
            ],
            "recovery": [
                {"id": "d1_r4", "label": "Audit all recent changes made by svc_backup", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Ensures no other backdoors were created."}
            ]
        }
    })

    # =========================================================================
    # DOMAIN 2: Threats, Vulnerabilities, and Mitigations (domain2)
    # =========================================================================

    # 4. TP: SQL Injection leading to exfiltration
    logs_4 = []
    for i in range(15):
        logs_4.append({
            "timestamp": t(i * 10), "source": "nginx", "severity": "info", "severity_level": 1,
            "message": "GET /index.html HTTP/1.1 200 OK", "src_ip": "198.51.100.77"
        })
    logs_4.extend([
        {"timestamp": t(180), "source": "nginx", "severity": "warning", "severity_level": 3, "message": "GET /users/search?q=1%%27%%20UNION%%20SELECT%%20username,password_hash%%20FROM%%20users-- HTTP/1.1 200 OK", "src_ip": "198.51.100.77"},
        {"timestamp": t(190), "source": "postgresql", "severity": "warning", "severity_level": 3, "message": "Query: SELECT * FROM products WHERE name LIKE '%1' UNION SELECT username,password_hash FROM users--%'", "src_ip": "127.0.0.1"},
        {"timestamp": t(200), "source": "postgresql", "severity": "warning", "severity_level": 3, "message": "Database output payload size: 1.2MB returned to webapp server", "src_ip": "127.0.0.1"},
        {"timestamp": t(210), "source": "nginx", "severity": "high", "severity_level": 4, "message": "GET /users/search - Response size 1205300 bytes HTTP/1.1 200 OK", "src_ip": "198.51.100.77"}
    ])
    for i in range(15):
        logs_4.append({
            "timestamp": t(220 + i * 10), "source": "waf", "severity": "info", "severity_level": 1,
            "message": "Request inspected: status pass", "src_ip": f"198.51.100.{100+i}"
        })

    scenarios.append({
        "title": "SQL Injection & Database Exfiltration",
        "description": "Alert from web logs showing suspicious URL queries containing SQL keywords like UNION and SELECT, followed by a unusually large HTTP response payload sent back to the external client.",
        "domain_id": "domain2",
        "difficulty": 2,
        "is_true_positive": True,
        "tp_fp_explanation": "This is a True Positive. The query payload contains a classic SQL Union Injection string targeting `password_hash`, resulting in a database dump of 1.2MB sent to external IP 198.51.100.77.",
        "initial_logs": logs_4,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d2_f1",
                    "description": "Found SQL injection payload in query",
                    "domain_id": "domain2",
                    "points": 10,
                    "explanation": "The parameter 'q' contains a UNION SELECT query used to extract credentials."
                },
                {
                    "id": "d2_f2",
                    "description": "Identified anomalous data volume exfiltration",
                    "domain_id": "domain4",
                    "points": 10,
                    "explanation": "Response payload size of 1.2MB confirms successful data extraction."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d2_c1", "label": "Block external IP 198.51.100.77 at web application firewall", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Blocks further connection from the attacker."},
                {"id": "d2_c2", "label": "Take the search endpoint offline or enable strict WAF rules", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Mitigates the vulnerable search API endpoint immediately."}
            ],
            "eradication": [
                {"id": "d2_e1", "label": "Implement parameterized queries (prepared statements) on search API", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Permanently remediates SQL injection vulnerability."},
                {"id": "d2_e2", "label": "Rotate all admin passwords stored in DB", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Secures credentials potentially exposed by data dump."},
                {"id": "d2_e3", "label": "Delete the entire PostgreSQL database", "is_correct": False, "domain_id": "domain3", "points": -4, "explanation": "Overkill; leads to total business data loss."}
            ],
            "recovery": [
                {"id": "d2_r1", "label": "Scan database for web shells or persistent anomalies", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Ensures no secondary payloads were uploaded."}
            ]
        }
    })

    # 5. FP: Port scanning from authorized internal penetration tester
    logs_5 = []
    for i in range(25):
        logs_5.append({
            "timestamp": t(i * 2), "source": "nids", "severity": "medium", "severity_level": 2,
            "message": f"TCP Port scan detected from 10.0.99.12 to 10.0.1.{i} on Port {80+i}"
        })
    logs_5.extend([
        {"timestamp": t(60), "source": "nids", "severity": "medium", "severity_level": 2, "message": "SYN Flood Scan from 10.0.99.12 targeting Active Directory", "src_ip": "10.0.99.12"},
        {"timestamp": t(70), "source": "change_management", "severity": "info", "severity_level": 1, "message": "CR-4099 Approved: Authorized internal vulnerability scanning from test subnet (10.0.99.0/24) scheduled Jul 13 08:00 - 12:00", "user": "sysadmin"}
    ])
    for i in range(15):
        logs_5.append({
            "timestamp": t(80 + i * 5), "source": "ad_controller", "severity": "info", "severity_level": 1,
            "message": "Normal audit logon event", "user": f"user{i}"
        })

    scenarios.append({
        "title": "Internal Port Scan and SYN Flood Alarm",
        "description": "An IDS alarm flags a massive network TCP sweep and port scan originating from internal host IP 10.0.99.12 across multiple network zones.",
        "domain_id": "domain2",
        "difficulty": 2,
        "is_true_positive": False,
        "tp_fp_explanation": "This is a False Positive. Change request logs indicate an authorized internal vulnerability scan was scheduled from IP range 10.0.99.0/24 during this time window.",
        "initial_logs": logs_5,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d2_f3",
                    "description": "Correlated scanner IP with change management log",
                    "domain_id": "domain5",
                    "points": 10,
                    "explanation": "Vulnerability scan was pre-authorized via Change Request CR-4099."
                },
                {
                    "id": "d2_f4",
                    "description": "Confirmed scanning behavior is internal-only",
                    "domain_id": "domain2",
                    "points": 10,
                    "explanation": "Traffic remained within internal private subnets."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d2_c3", "label": "Shut down switch ports for subnet 10.0.99.0/24", "is_correct": False, "domain_id": "domain3", "points": -2, "explanation": "Interrupts scheduled security compliance audits."},
                {"id": "d2_c4", "label": "Document alert as false positive and add scanning host to IDS exclusions", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Correct action for authorized scanning activities."}
            ],
            "eradication": [
                {"id": "d2_e4", "label": "Review scanning configurations to minimize noise", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Helps optimization of scanning scope."}
            ],
            "recovery": [
                {"id": "d2_r2", "label": "Re-verify scanning schedule at the end of the maintenance window", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Ensures scans conclude on time."}
            ]
        }
    })

    # 6. TP: Ransomware attack execution
    logs_6 = []
    for i in range(15):
        logs_6.append({
            "timestamp": t(i * 10), "source": "edr", "severity": "info", "severity_level": 1,
            "message": "Registry queried: CurrentVersion", "host": "WS-FIN-08"
        })
    logs_6.extend([
        {"timestamp": t(180), "source": "edr", "severity": "high", "severity_level": 4, "message": "Command line execution: 'vssadmin.exe delete shadows /all /quiet'", "host": "WS-FIN-08", "user": "mclark"},
        {"timestamp": t(190), "source": "edr", "severity": "high", "severity_level": 4, "message": "Command line execution: 'wmic shadowcopy delete'", "host": "WS-FIN-08", "user": "mclark"},
        {"timestamp": t(200), "source": "file_integrity", "severity": "critical", "severity_level": 5, "message": "Rapid file modification: 145 files renamed to extension '.locked' in 3 seconds", "host": "WS-FIN-08"},
        {"timestamp": t(210), "source": "edr", "severity": "critical", "severity_level": 5, "message": "Process block: 'ransom_decrypt_instruction.txt' created on desktop", "host": "WS-FIN-08"}
    ])
    for i in range(15):
        logs_6.append({
            "timestamp": t(220 + i * 10), "source": "firewall", "severity": "info", "severity_level": 1,
            "message": "Outbound HTTP allow", "src_ip": "10.0.1.55"
        })

    scenarios.append({
        "title": "Ransomware Execution & Volume Shadow Copy Deletion",
        "description": "An endpoint security alarm alerts on commands attempting to delete volume shadow copies on finance workstation WS-FIN-08, followed by rapid file encryption flags.",
        "domain_id": "domain2",
        "difficulty": 3,
        "is_true_positive": True,
        "tp_fp_explanation": "This is a True Positive. The commands executing 'vssadmin delete shadows' and 'wmic shadowcopy delete' are signature actions used by ransomware threat actors to prevent system recovery before initiating encryption.",
        "initial_logs": logs_6,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d2_f5",
                    "description": "Identified shadow copy deletion commands",
                    "domain_id": "domain2",
                    "points": 10,
                    "explanation": "vssadmin.exe delete shadows is used to inhibit system recovery."
                },
                {
                    "id": "d2_f6",
                    "description": "Discovered ransomware file signature modifications",
                    "domain_id": "domain2",
                    "points": 10,
                    "explanation": "Files renamed to .locked and ransom note creation confirm active encryption."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d2_c5", "label": "Network isolate WS-FIN-08 immediately via EDR/switch block", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Prevents ransomware from spreading laterally via SMB/network shares."},
                {"id": "d2_c6", "label": "Disable Active Directory user session for mclark", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Revokes user rights to network shares."}
            ],
            "eradication": [
                {"id": "d2_e5", "label": "Terminate processes with active .locked handles and quarantine executable", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Stops active encryption process."},
                {"id": "d2_e6", "label": "Format the hard drive of WS-FIN-08 and reinstall OS", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Ensures no malware remnants remain on endpoint."}
            ],
            "recovery": [
                {"id": "d2_r3", "label": "Restore files from offline backup repository", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Safely retrieves clean copies of encrypted data."},
                {"id": "d2_r4", "label": "Pay the threat actor ransom request to obtain key", "is_correct": False, "domain_id": "domain5", "points": -5, "explanation": "Unsafe, violates policy, and does not guarantee decryption."}
            ]
        }
    })

    # =========================================================================
    # DOMAIN 3: Security Architecture (domain3)
    # =========================================================================

    # 7. TP: Guest VLAN segmentation bypass
    logs_7 = []
    for i in range(15):
        logs_7.append({
            "timestamp": t(i * 10), "source": "dhcp", "severity": "info", "severity_level": 1,
            "message": "IP assigned: 192.168.10.145 on Guest-WiFi", "host": "Guest-Device-11"
        })
    logs_7.extend([
        {"timestamp": t(180), "source": "firewall", "severity": "high", "severity_level": 4, "message": "Inbound TCP connection allowed: 192.168.10.145:49810 -> 10.0.50.12:5432 (DB-Server-Prod)", "src_ip": "192.168.10.145", "dst_ip": "10.0.50.12"},
        {"timestamp": t(190), "source": "postgresql", "severity": "warning", "severity_level": 3, "message": "Database login failed for user 'sa' from host 192.168.10.145", "src_ip": "192.168.10.145"},
        {"timestamp": t(200), "source": "postgresql", "severity": "warning", "severity_level": 3, "message": "Database login failed for user 'admin' from host 192.168.10.145", "src_ip": "192.168.10.145"},
        {"timestamp": t(210), "source": "firewall", "severity": "high", "severity_level": 4, "message": "Inbound TCP connection allowed: 192.168.10.145:49812 -> 10.0.50.12:5432 (DB-Server-Prod)", "src_ip": "192.168.10.145", "dst_ip": "10.0.50.12"}
    ])
    for i in range(15):
        logs_7.append({
            "timestamp": t(220 + i * 10), "source": "firewall", "severity": "info", "severity_level": 1,
            "message": "Outbound HTTP allow from 192.168.10.145", "src_ip": "192.168.10.145"
        })

    scenarios.append({
        "title": "VLAN Segmentation Failure / Guest Network Bypass",
        "description": "Alert indicating firewall allowing database connections directly from the Guest Wi-Fi VLAN (192.168.10.0/24) to the production database subnet.",
        "domain_id": "domain3",
        "difficulty": 2,
        "is_true_positive": True,
        "tp_fp_explanation": "This is a True Positive. Production DB subnets should NEVER accept connections from the Guest network. The firewall configuration permits this due to a misconfigured ruleset, which is currently being brute-forced.",
        "initial_logs": logs_7,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d3_f1",
                    "description": "Discovered unauthorized VLAN routing path",
                    "domain_id": "domain3",
                    "points": 10,
                    "explanation": "Firewall permitted route 192.168.10.145 -> 10.0.50.12:5432."
                },
                {
                    "id": "d3_f2",
                    "description": "Identified database brute-force signatures",
                    "domain_id": "domain2",
                    "points": 10,
                    "explanation": "Multiple database login failures indicate active brute forcing."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d3_c1", "label": "Revoke firewall rule allowing Guest VLAN to Prod DB routing", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Corrects segmentation failure instantly."},
                {"id": "d3_c2", "label": "Block Guest client 192.168.10.145 on local WLC (Wireless Controller)", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Bans attacker device from network."}
            ],
            "eradication": [
                {"id": "d3_e1", "label": "Verify firewall ACL hierarchy and implement implicit deny between zones", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Reinforces secure network architecture."}
            ],
            "recovery": [
                {"id": "d3_r1", "label": "Verify DB access logs to confirm no successful logins were achieved", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Verifies data integrity and confidentiality."}
            ]
        }
    })

    # 8. FP: Guest user accessing video streaming services
    logs_8 = []
    for i in range(25):
        logs_8.append({
            "timestamp": t(i * 5), "source": "proxy", "severity": "info", "severity_level": 1,
            "message": "CONNECT youtube.com:443 HTTP/1.1", "src_ip": "192.168.10.12", "bytes_sent": 842000
        })
    logs_8.extend([
        {"timestamp": t(130), "source": "nids", "severity": "medium", "severity_level": 2, "message": "Excessive outbound bandwidth usage: 192.168.10.12 on Port 443", "src_ip": "192.168.10.12"},
        {"timestamp": t(140), "source": "proxy", "severity": "info", "severity_level": 1, "message": "CONNECT netflix.com:443 HTTP/1.1", "src_ip": "192.168.10.12", "bytes_sent": 1205300}
    ])
    for i in range(10):
        logs_8.append({
            "timestamp": t(150 + i * 5), "source": "dhcp", "severity": "info", "severity_level": 1,
            "message": "DHCP renew 192.168.10.12"
        })

    scenarios.append({
        "title": "Exorbitant Bandwidth Consumption on Guest Network",
        "description": "An alert warns of high outbound network traffic on port 443 originating from a guest endpoint 192.168.10.12.",
        "domain_id": "domain3",
        "difficulty": 1,
        "is_true_positive": False,
        "tp_fp_explanation": "This is a False Positive. Investigation reveals the proxy logs show the traffic consists entirely of encrypted HTTPS connections to YouTube and Netflix, which represent legitimate consumer bandwidth usage allowed on the Guest network policy.",
        "initial_logs": logs_8,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d3_f3",
                    "description": "Correlated traffic with video streaming domains",
                    "domain_id": "domain3",
                    "points": 10,
                    "explanation": "Traffic destinations were youtube.com and netflix.com."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d3_c3", "label": "Shut down the entire Guest Wi-Fi SSID network", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Inconveniences all users due to one video stream."},
                {"id": "d3_c4", "label": "Apply QoS bandwidth limiting on the Guest VLAN interface", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Mitigates high consumption without locking users out."}
            ],
            "eradication": [
                {"id": "d3_e2", "label": "Review company wireless guest usage policies", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Ensures policy aligns with bandwidth usage expectations."}
            ],
            "recovery": [
                {"id": "d3_r2", "label": "Monitor guest traffic threshold alarms", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Verifies that QoS bounds are respected."}
            ]
        }
    })

    # 9. TP: Public AWS S3 Bucket Misconfiguration
    logs_9 = []
    for i in range(15):
        logs_9.append({
            "timestamp": t(i * 10), "source": "aws_cloudtrail", "severity": "info", "severity_level": 1,
            "message": "AssumeRole: Role=ReadOnlyUser", "src_ip": "10.0.1.15"
        })
    logs_9.extend([
        {"timestamp": t(180), "source": "aws_cloudtrail", "severity": "high", "severity_level": 4, "message": "PutBucketAcl: S3Bucket=customer-contracts-prod Policy=PublicReadOwnerAuthorizedAll", "src_ip": "10.0.1.15", "user": "temp_admin"},
        {"timestamp": t(190), "source": "aws_cloudtrail", "severity": "critical", "severity_level": 5, "message": "ListObjects: S3Bucket=customer-contracts-prod", "src_ip": "198.51.100.99", "user": "Anonymous"},
        {"timestamp": t(200), "source": "aws_cloudtrail", "severity": "critical", "severity_level": 5, "message": "GetObject: S3Bucket=customer-contracts-prod Key=contract_99420.pdf", "src_ip": "198.51.100.99", "user": "Anonymous"},
        {"timestamp": t(210), "source": "aws_cloudtrail", "severity": "critical", "severity_level": 5, "message": "GetObject: S3Bucket=customer-contracts-prod Key=contract_99421.pdf", "src_ip": "198.51.100.99", "user": "Anonymous"}
    ])
    for i in range(15):
        logs_9.append({
            "timestamp": t(220 + i * 10), "source": "aws_cloudtrail", "severity": "info", "severity_level": 1,
            "message": "ListObjects allowed", "src_ip": "10.0.1.15"
        })

    scenarios.append({
        "title": "AWS S3 Cloud Storage Misconfiguration",
        "description": "Cloud security audit logs detect the S3 bucket 'customer-contracts-prod' access control list (ACL) modified to public read, followed by immediate anonymous access download activities.",
        "domain_id": "domain3",
        "difficulty": 3,
        "is_true_positive": True,
        "tp_fp_explanation": "This is a True Positive. The bucket was misconfigured to allow public anonymous reads (Policy modified), and an external attacker (198.51.100.99) began exfiltrating contract data immediately.",
        "initial_logs": logs_9,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d3_f4",
                    "description": "Identified S3 policy alteration",
                    "domain_id": "domain3",
                    "points": 10,
                    "explanation": "PutBucketAcl made the storage bucket publicly accessible."
                },
                {
                    "id": "d3_f5",
                    "description": "Found anonymous download logs",
                    "domain_id": "domain3",
                    "points": 10,
                    "explanation": "Anonymous user from 198.51.100.99 fetched contract files."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d3_c5", "label": "Revoke Public Access permissions and enable Block Public Access (BPA) on S3", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Stops public anonymous access immediately."},
                {"id": "d3_c6", "label": "Disable AWS IAM credentials for user temp_admin", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Prevents further privilege actions from compromised IAM user."}
            ],
            "eradication": [
                {"id": "d3_e3", "label": "Enable bucket policy auditing and automated remediation scripts via AWS Config", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Ensures drift from baseline config is auto-fixed."}
            ],
            "recovery": [
                {"id": "d3_r3", "label": "Analyze CloudTrail logs to quantify exact files exfiltrated", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Required for compliance reporting."},
                {"id": "d3_r4", "label": "Prepare breach notification response per compliance policy", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Necessary due to exfiltration of corporate customer contract details."}
            ]
        }
    })

    # =========================================================================
    # DOMAIN 4: Security Operations (domain4)
    # =========================================================================

    # 10. TP: DNS Tunneling exfiltration
    logs_10 = []
    for i in range(15):
        logs_10.append({
            "timestamp": t(i * 10), "source": "dns_server", "severity": "info", "severity_level": 1,
            "message": "Query: mail.google.com Type=A", "src_ip": "10.0.1.44"
        })
    logs_10.extend([
        {"timestamp": t(180), "source": "dns_server", "severity": "warning", "severity_level": 3, "message": "Query: aGVs...bG8.malicious-tunnel.com Type=TXT", "src_ip": "10.0.1.44"},
        {"timestamp": t(190), "source": "dns_server", "severity": "warning", "severity_level": 3, "message": "Query: d29y...bGQ.malicious-tunnel.com Type=TXT", "src_ip": "10.0.1.44"},
        {"timestamp": t(200), "source": "dns_server", "severity": "warning", "severity_level": 3, "message": "Query: c2Vj...dXJl.malicious-tunnel.com Type=TXT", "src_ip": "10.0.1.44"},
        {"timestamp": t(210), "source": "dns_server", "severity": "warning", "severity_level": 3, "message": "Query: cGFz...c3dk.malicious-tunnel.com Type=TXT", "src_ip": "10.0.1.44"}
    ])
    for i in range(15):
        logs_10.append({
            "timestamp": t(220 + i * 10), "source": "dns_server", "severity": "info", "severity_level": 1,
            "message": "Query: internal-wiki.corp Type=A", "src_ip": "10.0.1.44"
        })

    scenarios.append({
        "title": "DNS Tunneling / Data Exfiltration via DNS TXT Queries",
        "description": "SOC dashboard detects an anomaly: high volume of DNS TXT records containing long, encoded subdomains querying the external domain 'malicious-tunnel.com' from host 10.0.1.44.",
        "domain_id": "domain4",
        "difficulty": 3,
        "is_true_positive": True,
        "tp_fp_explanation": "This is a True Positive. The TXT DNS queries contain base64 encoded strings querying a newly registered domain. This is a classic DNS Tunneling technique used to bypass standard proxy restrictions and exfiltrate data.",
        "initial_logs": logs_10,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d4_f1",
                    "description": "Identified DNS TXT anomalies",
                    "domain_id": "domain4",
                    "points": 10,
                    "explanation": "High quantity of TXT record queries with randomized subdomains indicate tunneling."
                },
                {
                    "id": "d4_f2",
                    "description": "Discovered the rogue exfiltration domain name",
                    "domain_id": "domain4",
                    "points": 10,
                    "explanation": "All anomalous traffic pointed to domain 'malicious-tunnel.com'."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d4_c1", "label": "Blackhole domain malicious-tunnel.com on corporate DNS servers", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Breaks the communication channel instantly."},
                {"id": "d4_c2", "label": "Isolate the client host 10.0.1.44 from the network", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Stops the local script from exfiltrating files."}
            ],
            "eradication": [
                {"id": "d4_e1", "label": "Scan host 10.0.1.44 for running DNS tunneling toolkits", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Locates the malware mechanism."},
                {"id": "d4_e2", "label": "Reinstall DNS service on target DC server", "is_correct": False, "domain_id": "domain3", "points": -2, "explanation": "Unnecessary; the DNS server was just forwarding requests legitimately."}
            ],
            "recovery": [
                {"id": "d4_r1", "label": "Audit total TXT traffic volume to estimate loss", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Estimates extent of data leakage."}
            ]
        }
    })

    # 11. FP: High volume of internal NTP requests
    logs_11 = []
    for i in range(35):
        logs_11.append({
            "timestamp": t(i), "source": "firewall", "severity": "medium", "severity_level": 2,
            "message": "High frequency NTP request permitted: 10.0.3.15 -> time.nist.gov:123 UDP", "src_ip": "10.0.3.15", "dst_ip": "129.6.15.28"
        })
    logs_11.extend([
        {"timestamp": t(40), "source": "edr", "severity": "info", "severity_level": 1, "message": "Process activity: w32time.exe syncing time configuration loop", "host": "srv-sync-time-02"},
        {"timestamp": t(45), "source": "change_management", "severity": "info", "severity_level": 1, "message": "CR-3212: Configure NTP sync interval to 1s on time-server-backup due to oscillator failure", "user": "sysadmin"}
    ])

    scenarios.append({
        "title": "Anomalous Volume of NTP Network Requests",
        "description": "Firewall triggers an alert for high volume UDP port 123 traffic originating from server 10.0.3.15 directed towards a government time service.",
        "domain_id": "domain4",
        "difficulty": 2,
        "is_true_positive": False,
        "tp_fp_explanation": "This is a False Positive. The high-volume NTP requests were caused by an authorized change (CR-3212) configured by a sysadmin attempting to keep a backup server synchronized during a motherboard clock drift issue.",
        "initial_logs": logs_11,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d4_f3",
                    "description": "Correlated traffic with scheduled time-sync change",
                    "domain_id": "domain5",
                    "points": 10,
                    "explanation": "Change Request CR-3212 documents the temporary clock drift configuration."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d4_c3", "label": "Block UDP Port 123 outbound entirely at corporate firewall", "is_correct": False, "domain_id": "domain3", "points": -4, "explanation": "Breaks time synchronization for the entire enterprise."},
                {"id": "d4_c4", "label": "Document findings and close ticket as benign", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Resolves case safely without breaking systems."}
            ],
            "eradication": [
                {"id": "d4_e3", "label": "Verify internal time sync hierarchy is working", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Ensures domain controllers sync properly."}
            ],
            "recovery": [
                {"id": "d4_r2", "label": "Confirm hardware replacement schedule for oscillator drift", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Permanently fixes underlying root cause."}
            ]
        }
    })

    # 12. TP: Malicious Phishing email with link clicked
    logs_12 = []
    for i in range(15):
        logs_12.append({
            "timestamp": t(i * 10), "source": "email_gateway", "severity": "info", "severity_level": 1,
            "message": "Inbound email cleared: Sender=support@legit.com Recipient=kwilliams@corp.com"
        })
    logs_12.extend([
        {"timestamp": t(180), "source": "email_gateway", "severity": "warning", "severity_level": 3, "message": "Inbound email: Sender=admin@secure-office365-portal.com Recipient=kwilliams@corp.com Subject='Urgent: MFA Reset Required'", "src_ip": "192.0.2.14"},
        {"timestamp": t(190), "source": "proxy", "severity": "high", "severity_level": 4, "message": "Outbound HTTP connection: 10.0.2.77 -> secure-office365-portal.com:80", "src_ip": "10.0.2.77", "dst_ip": "192.0.2.14"},
        {"timestamp": t(200), "source": "edr", "severity": "high", "severity_level": 4, "message": "Credential theft attempt signature match: input fields detected on unsecured form", "host": "WS-HR-11", "user": "kwilliams"},
        {"timestamp": t(210), "source": "ad_controller", "severity": "critical", "severity_level": 5, "message": "New admin account registered: 'adm_backdoor' created by local machine WS-HR-11 privilege process", "host": "DC-01"}
    ])
    for i in range(15):
        logs_12.append({
            "timestamp": t(220 + i * 10), "source": "firewall", "severity": "info", "severity_level": 1,
            "message": "Outbound TCP connection allowed", "src_ip": "10.0.2.77"
        })

    scenarios.append({
        "title": "Spear Phishing leading to Credential Harvesting",
        "description": "An email gateway alert reports a suspicious external domain sent an email with subject 'Urgent: MFA Reset Required', followed by proxy activity of a user visiting the link and local credential anomalies.",
        "domain_id": "domain4",
        "difficulty": 2,
        "is_true_positive": True,
        "tp_fp_explanation": "This is a True Positive. The email sender domain 'secure-office365-portal.com' is a lookalike domain. The user clicked the link, entered their credentials, and the attacker immediately used them to create a backdoor administrator account.",
        "initial_logs": logs_12,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d4_f4",
                    "description": "Identified lookalike phishing domain",
                    "domain_id": "domain2",
                    "points": 10,
                    "explanation": "secure-office365-portal.com is a malicious lookalike domain."
                },
                {
                    "id": "d4_f5",
                    "description": "Found unauthorized admin account creation",
                    "domain_id": "domain1",
                    "points": 10,
                    "explanation": "Account 'adm_backdoor' was created maliciously without authorization."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d4_c5", "label": "Delete the administrator account 'adm_backdoor'", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Removes the attacker's foothold."},
                {"id": "d4_c6", "label": "Revoke active web session and change password for user kwilliams", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Secures the harvested credentials."},
                {"id": "d4_c7", "label": "Add secure-office365-portal.com to proxy blocking rules", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Blocks further link clicks across the organization."}
            ],
            "eradication": [
                {"id": "d4_e4", "label": "Search mail gateway and delete copies of this email from other mailboxes", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Prevents other users from falling victim."},
                {"id": "d4_e5", "label": "Quarantine the domain controller 'DC-01'", "is_correct": False, "domain_id": "domain3", "points": -5, "explanation": "Isolating the primary DC breaks the entire domain authentication services."}
            ],
            "recovery": [
                {"id": "d4_r3", "label": "Deploy anti-phishing training to user kwilliams", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Educates the user to prevent recurrence."}
            ]
        }
    })

    # =========================================================================
    # DOMAIN 5: Security Management (domain5)
    # =========================================================================

    # 13. TP: Disgruntled employee exfiltrating customer list
    logs_13 = []
    for i in range(15):
        logs_13.append({
            "timestamp": t(i * 10), "source": "crm_app", "severity": "info", "severity_level": 1,
            "message": "View customer account", "user": "rharper"
        })
    logs_13.extend([
        {"timestamp": t(180), "source": "crm_app", "severity": "high", "severity_level": 4, "message": "Bulk report exported: Customer_Billing_Details_All.csv", "user": "rharper"},
        {"timestamp": t(190), "source": "proxy", "severity": "high", "severity_level": 4, "message": "Outbound HTTP POST to mail.personal-gmail.com bytes_sent=4508200", "src_ip": "10.0.1.18", "user": "rharper"},
        {"timestamp": t(200), "source": "edr", "severity": "warning", "severity_level": 3, "message": "File deleted: Customer_Billing_Details_All.csv from desktop", "host": "WS-SALES-15", "user": "rharper"},
        {"timestamp": t(210), "source": "ad_controller", "severity": "info", "severity_level": 1, "message": "User session closed: rharper logged off", "user": "rharper"}
    ])
    for i in range(15):
        logs_13.append({
            "timestamp": t(220 + i * 10), "source": "firewall", "severity": "info", "severity_level": 1,
            "message": "Normal outbound connection allowed", "src_ip": f"10.0.1.{40+i}"
        })

    scenarios.append({
        "title": "Insider Threat: Customer Database Exfiltration",
        "description": "Alert raised on unusual CRM database export activity by sales representative 'rharper', followed by a massive outbound data transfer to a personal email domain.",
        "domain_id": "domain5",
        "difficulty": 2,
        "is_true_positive": True,
        "tp_fp_explanation": "This is a True Positive. The employee exported customer billing databases outside normal business operations, exfiltrated 4.5MB to personal webmail, and deleted the files locally to hide traces.",
        "initial_logs": logs_13,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d5_f1",
                    "description": "Identified anomalous CRM data export",
                    "domain_id": "domain5",
                    "points": 10,
                    "explanation": "Bulk data export of customer billing records was atypical for this role."
                },
                {
                    "id": "d5_f2",
                    "description": "Correlated export with personal webmail exfiltration",
                    "domain_id": "domain4",
                    "points": 10,
                    "explanation": "Post to personal-gmail.com matches size of exported data."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d5_c1", "label": "Suspend rharper's Active Directory and CRM accounts immediately", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Stops ongoing data exfiltration."},
                {"id": "d5_c2", "label": "Disable sales host PC WS-SALES-15 network access", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Secures local forensic evidence."}
            ],
            "eradication": [
                {"id": "d5_e1", "label": "Initiate forensic acquisition of workstation hard drive", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Ensures evidence is preserved for potential legal actions."}
            ],
            "recovery": [
                {"id": "d5_r1", "label": "Report security breach to Legal & HR departments", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Required by policy for employee policy breaches."},
                {"id": "d5_r2", "label": "File GDPR/PDPA compliance breach notifications within 72 hours", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Necessary due to exfiltration of customer PII data."}
            ]
        }
    })

    # 14. FP: Pre-approved data backup archiving job
    logs_14 = []
    for i in range(15):
        logs_14.append({
            "timestamp": t(i * 10), "source": "aws_cloudtrail", "severity": "info", "severity_level": 1,
            "message": "API Call: DescribeInstances", "user": "backup_agent_04"
        })
    logs_14.extend([
        {"timestamp": t(180), "source": "aws_cloudtrail", "severity": "medium", "severity_level": 2, "message": "High volume data synchronization: S3 bucket 'audit-logs-archive' copied to Glacier storage", "user": "backup_agent_04"},
        {"timestamp": t(190), "source": "change_management", "severity": "info", "severity_level": 1, "message": "CR-9022 Approved: Standard monthly compliance audit archiving job scheduled Jul 13 08:30", "user": "compliance_lead"},
        {"timestamp": t(200), "source": "aws_cloudtrail", "severity": "info", "severity_level": 1, "message": "Glacier transfer complete: 45.2 GB archived successfully", "user": "backup_agent_04"}
    ])
    for i in range(15):
        logs_14.append({
            "timestamp": t(210 + i * 10), "source": "aws_cloudtrail", "severity": "info", "severity_level": 1,
            "message": "API Call: GetObject", "user": "backup_agent_04"
        })

    scenarios.append({
        "title": "Anomalous Cloud Data Archiving Alert",
        "description": "AWS compliance engine alerts on massive data migration of S3 audit log bucket contents into cold storage Glacier vaults.",
        "domain_id": "domain5",
        "difficulty": 1,
        "is_true_positive": False,
        "tp_fp_explanation": "This is a False Positive. The activity represents the standard monthly compliance log archiving task pre-approved under CR-9022, executing using the authorized service principal backup_agent_04.",
        "initial_logs": logs_14,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d5_f3",
                    "description": "Correlated backup agent with compliance schedule",
                    "domain_id": "domain5",
                    "points": 10,
                    "explanation": "CR-9022 outlines the scheduled archiving window."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d5_c3", "label": "Revoke AWS IAM permissions of backup_agent_04", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Breaks automated data compliance retention pipeline."},
                {"id": "d5_c4", "label": "Acknowledge alert as benign and confirm backup job completion", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Standard action for authorized scheduled tasks."}
            ],
            "eradication": [
                {"id": "d5_e2", "label": "Verify cloud retention policy limits match configuration", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Confirms data lifecycle operates correctly."}
            ],
            "recovery": [
                {"id": "d5_r3", "label": "Update compliance schedule record to reduce monitoring alarms", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Prevents future security alert triggers."}
            ]
        }
    })

    # 15. TP: Rogue Access Point installation
    logs_15 = []
    for i in range(15):
        logs_15.append({
            "timestamp": t(i * 10), "source": "dhcp_server", "severity": "info", "severity_level": 1,
            "message": "IP assigned: 10.0.4.155 to MAC 00:e0:4c:00:a1:05", "host": "corporate-device"
        })
    logs_15.extend([
        {"timestamp": t(180), "source": "switch_core", "severity": "high", "severity_level": 4, "message": "Port Security violation: 12 MAC addresses detected on single port interface Gi1/0/22 (Port allowed 1 MAC)", "host": "SW-HQ-02"},
        {"timestamp": t(190), "source": "dhcp_server", "severity": "warning", "severity_level": 3, "message": "IP requests: 8 unauthenticated devices requested IPs from port Gi1/0/22", "host": "SW-HQ-02"},
        {"timestamp": t(200), "source": "ids_wifi", "severity": "high", "severity_level": 4, "message": "Unauthorized SSIDs detected broadcasting: 'CORP-GUEST-FAST' broadcasting on channel 6, MAC 00:e0:4c:00:a1:05", "host": "W-AP-3"},
        {"timestamp": t(210), "source": "switch_core", "severity": "critical", "severity_level": 5, "message": "MAC spoofing signatures matched on switch interface Gi1/0/22", "host": "SW-HQ-02"}
    ])
    for i in range(15):
        logs_15.append({
            "timestamp": t(220 + i * 10), "source": "dhcp_server", "severity": "info", "severity_level": 1,
            "message": "IP assigned normal host", "host": "PC-502"
        })

    scenarios.append({
        "title": "Rogue Wireless Access Point Detection",
        "description": "Network switch flags a Port Security violation on port Gi1/0/22 indicating multiple unauthorized MAC addresses connecting to a single LAN port, while wireless IDS detects a rogue SSID 'CORP-GUEST-FAST'.",
        "domain_id": "domain5",
        "difficulty": 3,
        "is_true_positive": True,
        "tp_fp_explanation": "This is a True Positive. An unauthorized hardware router/access point was plugged into switch port Gi1/0/22 to bypass physical security controls. It is broadcasting a rogue SSID to harvest corporate employee credentials.",
        "initial_logs": logs_15,
        "expected_outcomes": {
            "key_findings": [
                {
                    "id": "d5_f4",
                    "description": "Identified port security violation",
                    "domain_id": "domain3",
                    "points": 10,
                    "explanation": "Gi1/0/22 exceeded its maximum MAC limit due to downstream router."
                },
                {
                    "id": "d5_f5",
                    "description": "Discovered rogue wireless SSID",
                    "domain_id": "domain5",
                    "points": 10,
                    "explanation": "Rogue SSID 'CORP-GUEST-FAST' broadcasted from the same device MAC."
                }
            ]
        },
        "playbook_steps": {
            "containment": [
                {"id": "d5_c5", "label": "Shut down switch interface Gi1/0/22 immediately via CLI", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Isolates rogue device from physical network."},
                {"id": "d5_c6", "label": "Notify physical security teams to locate and retrieve the rogue access point", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Required since hardware must be confiscated physically."}
            ],
            "eradication": [
                {"id": "d5_e3", "label": "Enforce MAC filtering and 802.1X authentication on all access switch ports", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Prevents unauthorized hardware from connecting in the future."}
            ],
            "recovery": [
                {"id": "d5_r4", "label": "Verify DHCP logs to list all devices that associated with rogue IP pool", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Identifies potentially compromised corporate endpoints."}
            ]
        }
    })

    return scenarios


def assign_evidence_indices(scenarios):
    """
    Post-process scenarios to add evidence_log_indices to each key finding.
    
    All scenarios follow the structure:
      - 15 noise logs (indices 0-14)
      - N evidence logs (indices 15 through len(logs)-16)
      - 15 noise logs (indices len(logs)-15 through len(logs)-1)
    
    This function distributes the evidence indices across findings so
    the frontend's KQL discovery mechanic works correctly.
    """
    for scenario in scenarios:
        logs = scenario.get("initial_logs", [])
        total_logs = len(logs)
        
        # Evidence entries are between the two noise blocks of 15 each
        evidence_start = 15
        evidence_end = total_logs - 16  # inclusive
        
        if evidence_end < evidence_start:
            continue
        
        evidence_indices = list(range(evidence_start, evidence_end + 1))
        
        key_findings = scenario.get("expected_outcomes", {}).get("key_findings", [])
        if not key_findings:
            continue
        
        num_findings = len(key_findings)
        num_evidence = len(evidence_indices)
        
        # Distribute evidence indices across findings
        # Each finding gets a roughly equal slice, with overlap on boundaries
        chunk_size = max(1, num_evidence // num_findings)
        
        for i, finding in enumerate(key_findings):
            start = i * chunk_size
            # Last finding gets all remaining indices
            if i == num_findings - 1:
                end = num_evidence
            else:
                # Include one overlapping index with next finding for realism
                end = min(start + chunk_size + 1, num_evidence)
            
            finding["evidence_log_indices"] = evidence_indices[start:end]
    
    return scenarios


if __name__ == "__main__":
    scs = create_scenarios()
    scs = assign_evidence_indices(scs)
    with open("scenarios_seed.json", "w", encoding="utf-8") as f:
        json.dump(scs, f, indent=2, ensure_ascii=False)
    print("Successfully generated scenarios_seed.json containing 15 scenarios!")
