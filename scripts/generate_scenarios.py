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
                {"id": "d1_c1", "label": "Revoke all active Okta SSO tokens and web sessions for user ajones immediately", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Terminates the attacker's hijacked session immediately."},
                {"id": "d1_c2", "label": "Temporarily disable the ajones Active Directory account and block external IP 198.51.100.42 at the perimeter firewall", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Prevents further domain access and stops connections from the attacker's source IP."},
                {"id": "d1_c3", "label": "Demote domain controller DC-01 offline immediately and flush all volatile Kerberos ticket caches across the forest", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Demoting the primary domain controller causes an enterprise-wide outage and destroys volatile forensic evidence."},
                {"id": "d1_c4", "label": "Switch the Okta tenant MFA policy for ajones from push notifications to unverified email links while keeping existing sessions active", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Leaves the attacker's active hijacked session intact and weakens authentication controls."}
            ],
            "eradication": [
                {"id": "d1_e1", "label": "Enforce phishing-resistant FIDO2 WebAuthn hardware keys for ajones and revoke legacy push notification enrollment", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "FIDO2 hardware tokens prevent push fatigue and session hijacking."},
                {"id": "d1_e2", "label": "Factory reset the user's mobile phone while leaving the compromised ajones and admin_backup AD password hashes unchanged", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Push fatigue does not require wiping the phone, and leaving compromised AD passwords allows immediate re-entry."},
                {"id": "d1_e3", "label": "Delete all Okta authentication audit logs from 198.51.100.42 to clear the security alert queue on DC-01", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Erasing audit logs destroys forensic evidence without removing the threat actor's access."}
            ],
            "recovery": [
                {"id": "d1_r1", "label": "Reset the admin_backup password on DC-01, revert unauthorized privilege changes, and re-enable ajones after credential rotation", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Reverts the unauthorized password change on admin_backup and safely restores user access."},
                {"id": "d1_r2", "label": "Restore ajones domain access using the previous password hash so cached workstation scripts continue running without interruption", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Reusing the compromised password allows the attacker to immediately log back in."},
                {"id": "d1_r3", "label": "Exempt ajones and admin_backup from MFA requirements for 48 hours on DC-01 to prevent further push fatigue lockouts", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Removing MFA on targeted accounts exposes them to trivial credential stuffing."}
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
                {"id": "d1_c4", "label": "Revoke active VPN session for user mroberts and lock their Active Directory account immediately", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Unnecessary lockout that disrupts verified legitimate business travel operations."},
                {"id": "d1_c5", "label": "Allow the verified VPN connection from 195.154.122.90 for mroberts and document HR travel confirmation", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Correct non-disruptive action since MFA, asset posture #4288, and HR travel logs succeeded."},
                {"id": "d1_c6", "label": "Block the entire French ISP IP block 195.154.0.0/16 at the corporate VPN gateway firewall", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Over-reactive geo-blocking that cuts off legitimate traveling employees in Paris."}
            ],
            "eradication": [
                {"id": "d1_e3", "label": "Correlate VPN gateway logs with HR travel records for mroberts to confirm no anomalous lateral activity occurred", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Ensures no unauthorized activity accompanied the verified travel session."},
                {"id": "d1_e4", "label": "Trigger a remote wipe on corporate laptop Asset Tag #4288 and revoke mroberts's VPN client certificate", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Destructive action against a compliant corporate asset during an authorized business trip."},
                {"id": "d1_e5", "label": "Permanently disable geolocation anomaly detection rules on the VPN gateway for all corporate users", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Creates a permanent visibility blind spot across the enterprise instead of logging a travel exception."}
            ],
            "recovery": [
                {"id": "d1_r3", "label": "Sync approved HR travel itinerary dates (Jul 12 - Jul 18) for mroberts with the SIEM geolocation whitelist", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Prevents recurring false positive alert fatigue during the remainder of the Paris summit."},
                {"id": "d1_r4", "label": "Force mroberts to connect over an unencrypted split-tunnel guest proxy until returning from France", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Degrades transport security for a legitimate remote employee."},
                {"id": "d1_r5", "label": "Reset all marketing department file share permissions and purge marketing_strategy.pdf from the server", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Disrupts legitimate business file access when no breach occurred."}
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
                {"id": "d1_c6", "label": "Revert the modified DACL on CN=Payroll_Confidential, lock svc_backup, and isolate jsmith workstation 10.0.2.14", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Restores payroll folder confidentiality, halts svc_backup abuse, and isolates the target host."},
                {"id": "d1_c7", "label": "Delete the entire Payroll_Confidential share on the file server before capturing SMB access logs for 10.0.2.14", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Destroys critical business data and forensic file access timestamps."},
                {"id": "d1_c8", "label": "Grant all finance users temporary read permissions to Payroll_Confidential while monitoring jsmith on 10.0.2.14", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Expands unauthorized exposure of sensitive payroll records instead of containing the breach."}
            ],
            "eradication": [
                {"id": "d1_e4", "label": "Rotate svc_backup and jsmith credentials, remove unauthorized ACL entries, and inspect 10.0.2.14 for exfiltrated copies of salary_matrix_2026.xlsx", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Remediates the compromised accounts and removes unauthorized local copies of sensitive files."},
                {"id": "d1_e5", "label": "Restart the Windows Server SMB service without rotating svc_backup credentials or removing jsmith from the Payroll_Confidential DACL", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Leaves both the compromised service account and the unauthorized permission grant active."},
                {"id": "d1_e6", "label": "Purge all Active Directory security event logs on the domain controller to clear DACL modification alerts for svc_backup", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Destroys the audit trail required for incident investigation."}
            ],
            "recovery": [
                {"id": "d1_r4", "label": "Audit all historical permission changes made by svc_backup, enforce least-privilege delegation, and restore monitored access to Payroll_Confidential", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Verifies no other backdoor ACLs were created and hardens directory permission governance."},
                {"id": "d1_r5", "label": "Re-enable svc_backup with Domain Admin privileges so automated nightly backup scripts do not encounter permission errors", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Violates least privilege and re-exposes the domain to privilege escalation."},
                {"id": "d1_r6", "label": "Disable file server object access auditing on Payroll_Confidential to reduce event log volume after restoring 10.0.2.14", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Blinds the SOC to future unauthorized reads of sensitive payroll files."}
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
                {"id": "d2_c1", "label": "Block external attacker IP 198.51.100.77 at the WAF and disable or enforce strict input filtering on the /users/search endpoint", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Blocks further requests from the attacker and mitigates the vulnerable search API endpoint immediately."},
                {"id": "d2_c2", "label": "Restart the nginx web server daemon while keeping the /users/search endpoint open to 198.51.100.77 to observe further queries", "is_correct": False, "domain_id": "domain2", "points": -3, "explanation": "Leaves the SQL injection vulnerability actively exploitable by the external threat actor."},
                {"id": "d2_c3", "label": "Block loopback traffic on 127.0.0.1 between nginx and PostgreSQL without restricting external requests from 198.51.100.77", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Breaks all internal database connectivity randomly without addressing the attacker's perimeter access."}
            ],
            "eradication": [
                {"id": "d2_e1", "label": "Refactor the /users/search query to use parameterized prepared statements and force a password reset for all exposed user hashes", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Permanently remediates the SQL injection flaw and invalidates credentials leaked in the 1.2MB dump."},
                {"id": "d2_e2", "label": "Drop the entire PostgreSQL users table and delete all nginx access logs referencing 198.51.100.77", "is_correct": False, "domain_id": "domain3", "points": -4, "explanation": "Causes catastrophic production data loss and destroys forensic web access evidence."},
                {"id": "d2_e3", "label": "Add a client-side JavaScript character filter on the search form while leaving the backend SQL string concatenation unchanged", "is_correct": False, "domain_id": "domain2", "points": -3, "explanation": "Client-side validation is trivially bypassed via direct HTTP GET requests from 198.51.100.77."}
            ],
            "recovery": [
                {"id": "d2_r1", "label": "Audit PostgreSQL for unauthorized database functions or web shells, deploy WAF SQLi signatures, and bring /users/search back online", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Ensures no secondary persistence remains in the database and validates defense-in-depth before restoring service."},
                {"id": "d2_r2", "label": "Restore the PostgreSQL database from a snapshot without patching /users/search and set WAF mode to pass-through for 198.51.100.77", "is_correct": False, "domain_id": "domain2", "points": -3, "explanation": "Leaves the application vulnerable to immediate re-exploitation."},
                {"id": "d2_r3", "label": "Configure PostgreSQL to run under the superuser account with trust authentication to speed up recovery after the 198.51.100.77 incident", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Severely degrades database privilege separation and authentication security."}
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
                {"id": "d2_c3", "label": "Shut down switch ports for test subnet 10.0.99.0/24 and isolate the Active Directory controller immediately", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Interrupts an authorized security compliance audit (CR-4099) and causes unnecessary domain downtime."},
                {"id": "d2_c4", "label": "Verify Change Request CR-4099 for scanner 10.0.99.12 and document the NIDS alert as an authorized False Positive", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Correctly validates the approved maintenance window without disrupting internal operations."},
                {"id": "d2_c5", "label": "Deploy an emergency firewall blackhole rule dropping all internal traffic between 10.0.99.12 and 10.0.1.0/24", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Invalidates the scheduled vulnerability assessment approved under CR-4099."}
            ],
            "eradication": [
                {"id": "d2_e4", "label": "Coordinate with the security assessment team to tune NIDS threshold suppressions for 10.0.99.12 during approved CR-4099 windows", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Reduces alert fatigue during scheduled scans while preserving detection outside maintenance windows."},
                {"id": "d2_e5", "label": "Re-image the internal vulnerability scanner at 10.0.99.12 and revoke the sysadmin account that approved CR-4099", "is_correct": False, "domain_id": "domain2", "points": -3, "explanation": "Destructive action against an authorized internal security appliance and administrator."},
                {"id": "d2_e6", "label": "Permanently disable SYN flood and port scan signatures on the NIDS across all internal subnets", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Blinds the NIDS to real internal lateral movement and reconnaissance outside the maintenance window."}
            ],
            "recovery": [
                {"id": "d2_r2", "label": "Confirm the vulnerability scan from 10.0.99.12 concludes at 12:00 per CR-4099 and verify full NIDS alerting resumes afterward", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Ensures temporary scan exclusions expire on schedule and normal monitoring posture is restored."},
                {"id": "d2_r3", "label": "Restore all hosts in 10.0.1.0/24 from offline backups to undo the TCP port scan packets sent by 10.0.99.12", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Port scanning does not corrupt host state; restoring from backup causes severe unnecessary downtime."},
                {"id": "d2_r4", "label": "Whitelist 10.0.99.0/24 permanently on all production firewalls without requiring future change management tickets", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Bypasses change management governance and creates a permanent unmonitored pivot zone."}
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
                {"id": "d2_c5", "label": "Isolate finance workstation WS-FIN-08 immediately via EDR network containment and disable mclark's Active Directory session", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Prevents ransomware from spreading laterally across SMB shares and revokes compromised user access."},
                {"id": "d2_c6", "label": "Mount all corporate SMB file shares onto WS-FIN-08 with write permissions to test which directories the ransomware targets", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Directly exposes enterprise network shares to active file encryption."},
                {"id": "d2_c7", "label": "Reboot WS-FIN-08 into normal Windows mode while keeping it connected to the 10.0.1.0/24 finance VLAN", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Allows persistent ransomware tasks to resume encryption and lateral propagation."}
            ],
            "eradication": [
                {"id": "d2_e5", "label": "Terminate processes holding .locked file handles on WS-FIN-08, quarantine the ransomware binary, and clean-reimage the workstation OS", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Halts active encryption and ensures zero ransomware persistence remains on the endpoint."},
                {"id": "d2_e6", "label": "Rename the .locked file extensions back to their original names on WS-FIN-08 without removing the ransomware executable", "is_correct": False, "domain_id": "domain2", "points": -3, "explanation": "Renaming encrypted files does not decrypt their contents and leaves the malware active."},
                {"id": "d2_e7", "label": "Delete the EDR agent on WS-FIN-08 to stop critical file modification alerts for user mclark", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Removes endpoint protection while allowing encryption to continue unchecked."}
            ],
            "recovery": [
                {"id": "d2_r3", "label": "Restore affected finance data on WS-FIN-08 from a verified immutable offline backup and rotate mclark's domain credentials", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Safely recovers clean data after shadow copy deletion and secures the user account."},
                {"id": "d2_r4", "label": "Pay the threat actor's ransom demand noted in ransom_decrypt_instruction.txt to obtain the untrusted decryption binary", "is_correct": False, "domain_id": "domain5", "points": -5, "explanation": "Violates security policy, funds threat actors, and does not guarantee safe data recovery."},
                {"id": "d2_r5", "label": "Attempt to recover files on WS-FIN-08 solely from local Volume Shadow Copies without checking offline backups", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Local shadow copies were explicitly destroyed by vssadmin.exe and wmic shadowcopy delete."}
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
                {"id": "d3_c1", "label": "Remove the misconfigured firewall rule permitting Guest VLAN 192.168.10.0/24 to reach 10.0.50.12:5432 and quarantine client 192.168.10.145 on the WLC", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Immediately severs the unauthorized inter-VLAN path and isolates the brute-forcing wireless client."},
                {"id": "d3_c2", "label": "Shut down the PostgreSQL database service on 10.0.50.12 for all internal production applications while leaving the Guest VLAN route open", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Causes a production outage without fixing the underlying network segmentation flaw."},
                {"id": "d3_c3", "label": "Bridge Guest-WiFi VLAN 192.168.10.0/24 directly into the production database subnet 10.0.50.0/24 to bypass firewall inspection overhead", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Completely eliminates network segmentation and exposes all production servers to guest devices."}
            ],
            "eradication": [
                {"id": "d3_e1", "label": "Audit the inter-VLAN firewall ACL hierarchy, enforce an explicit deny-all rule between Guest-WiFi and internal zones, and lock default 'sa' accounts", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Permanently enforces zero-trust zone segmentation and hardens database authentication."},
                {"id": "d3_e2", "label": "Change the PostgreSQL listening port on 10.0.50.12 from 5432 to 5433 while retaining full IP routing from 192.168.10.145", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Port obfuscation does not remediate the missing inter-VLAN segmentation control."},
                {"id": "d3_e3", "label": "Clear the PostgreSQL failed authentication logs on 10.0.50.12 without modifying the Guest-WiFi firewall ruleset", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Erases brute-force evidence while leaving the database exposed to 192.168.10.145."}
            ],
            "recovery": [
                {"id": "d3_r1", "label": "Verify PostgreSQL audit logs on 10.0.50.12 to confirm zero successful logins from 192.168.10.145 and validate segmentation with a port scan test", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Confirms data confidentiality remained intact and verifies the firewall fix blocks guest traffic."},
                {"id": "d3_r2", "label": "Set the PostgreSQL 'sa' and 'admin' accounts on 10.0.50.12 to passwordless trust authentication for internal recovery testing", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Introduces a critical authentication vulnerability on the production database server."},
                {"id": "d3_r3", "label": "Disable firewall connection logging between 192.168.10.0/24 and 10.0.50.0/24 to prevent future high-severity SIEM alerts", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Removes network visibility required to detect future segmentation breaches."}
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
                {"id": "d3_c3", "label": "Shut down the entire Guest Wi-Fi SSID network and quarantine all connected guest MAC addresses immediately", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Over-reactive disruption to all guest users for benign video streaming on 192.168.10.12."},
                {"id": "d3_c4", "label": "Apply per-client QoS bandwidth rate limiting on the Guest VLAN interface for 192.168.10.12 and classify the alert as a False Positive", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Controls bandwidth consumption proportionally without treating benign streaming as a security breach."},
                {"id": "d3_c5", "label": "Block outbound TCP port 443 across both corporate and guest firewall zones to halt traffic from 192.168.10.12", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Blocking HTTPS globally cripples business web operations for a benign guest event."}
            ],
            "eradication": [
                {"id": "d3_e2", "label": "Tune the NIDS bandwidth anomaly threshold for Guest VLAN 192.168.10.0/24 to account for permitted media streaming domains", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Aligns NIDS alerting thresholds with acceptable guest network usage policies."},
                {"id": "d3_e3", "label": "Deploy mandatory corporate root CA TLS decryption certificates onto unowned guest device 192.168.10.12", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Forcing corporate root certificates onto personal visitor devices is invasive and inappropriate for a guest network."},
                {"id": "d3_e4", "label": "Wipe the DHCP lease database and disable proxy logging for 192.168.10.12 to silence bandwidth warnings", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Disrupts IP allocation and removes proxy visibility without managing bandwidth."}
            ],
            "recovery": [
                {"id": "d3_r2", "label": "Monitor Guest VLAN QoS utilization metrics to verify 192.168.10.12 remains within capped bandwidth limits and close the ticket", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Verifies that traffic shaping works as intended and completes benign incident closure."},
                {"id": "d3_r3", "label": "Route Guest VLAN 192.168.10.12 traffic through the internal corporate MPLS backbone to provide higher streaming throughput", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Violates network segmentation by routing untrusted guest traffic over internal corporate links."},
                {"id": "d3_r4", "label": "File a mandatory external regulatory data breach notification regarding the 1.2MB transfer from 192.168.10.12 to netflix.com", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Streaming media from Netflix is not data exfiltration and does not warrant a breach notification."}
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
                {"id": "d3_c5", "label": "Enable S3 Block Public Access on bucket customer-contracts-prod immediately and revoke active AWS IAM sessions for temp_admin", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Immediately halts anonymous public downloads from 198.51.100.99 and locks the account that altered the ACL."},
                {"id": "d3_c6", "label": "Delete the entire customer-contracts-prod S3 bucket and all stored PDF objects to stop downloads from 198.51.100.99", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Destroys critical production customer contracts instead of restricting bucket access permissions."},
                {"id": "d3_c7", "label": "Keep the PublicReadOwnerAuthorizedAll ACL active on customer-contracts-prod while renaming contract_99420.pdf", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Leaves the bucket publicly listable and readable by anonymous users at 198.51.100.99."}
            ],
            "eradication": [
                {"id": "d3_e3", "label": "Apply an AWS Organizations Service Control Policy denying PutBucketAcl public grants and deploy AWS Config auto-remediation on customer-contracts-prod", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Permanently prevents unauthorized public ACL modifications and remediates configuration drift."},
                {"id": "d3_e4", "label": "Disable AWS CloudTrail logging on customer-contracts-prod to stop generating critical GetObject alerts from 198.51.100.99", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Eliminates cloud audit visibility without fixing the bucket permission vulnerability."},
                {"id": "d3_e5", "label": "Attach the AdministratorAccess policy to temp_admin so the user can manually manage S3 bucket ACLs without restrictions", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Violates least privilege and expands the blast radius of the compromised or mismanaged IAM account."}
            ],
            "recovery": [
                {"id": "d3_r3", "label": "Audit CloudTrail GetObject events from 198.51.100.99 to inventory exfiltrated contracts and initiate the compliance breach notification workflow", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Quantifies the exact scope of exposed customer contracts and fulfills regulatory reporting duties."},
                {"id": "d3_r4", "label": "Move customer-contracts-prod objects to an unencrypted public FTP server while AWS S3 permissions are reviewed", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Exposes sensitive contracts over an even less secure cleartext protocol."},
                {"id": "d3_r5", "label": "Purge all CloudTrail history for temp_admin and 198.51.100.99 to avoid reporting the exposure during the next compliance audit", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Violates legal and regulatory audit integrity requirements."}
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
                {"id": "d4_c1", "label": "Sinkhole malicious-tunnel.com on corporate DNS resolvers and isolate internal host 10.0.1.44 from the network segment", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Severs the covert DNS exfiltration channel and contains the compromised endpoint."},
                {"id": "d4_c2", "label": "Shut down port 53 for all internal corporate DNS servers so no hosts in the enterprise can resolve domain names", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Disabling enterprise DNS causes a total business outage instead of targeting malicious-tunnel.com and 10.0.1.44."},
                {"id": "d4_c3", "label": "Configure host 10.0.1.44 to bypass internal DNS resolvers and query external public DNS 8.8.8.8 directly for malicious-tunnel.com", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Allows the DNS tunneling malware on 10.0.1.44 to continue exfiltrating data outside internal DNS visibility."}
            ],
            "eradication": [
                {"id": "d4_e1", "label": "Identify and terminate the DNS tunneling process on 10.0.1.44, remove its scheduled persistence, and rotate credentials cached on the host", "is_correct": True, "domain_id": "domain2", "points": 5, "explanation": "Removes the active tunneling toolkit and invalidates credentials exposed on 10.0.1.44."},
                {"id": "d4_e2", "label": "Format and reinstall the corporate DNS server software while leaving workstation 10.0.1.44 untouched on the VLAN", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "The corporate DNS resolver merely forwarded queries; the infection resides on endpoint 10.0.1.44."},
                {"id": "d4_e3", "label": "Flush the DNS resolver cache for malicious-tunnel.com without scanning 10.0.1.44 for scheduled tasks or background daemons", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Flushing the DNS cache does not stop the malware on 10.0.1.44 from issuing new TXT queries."}
            ],
            "recovery": [
                {"id": "d4_r1", "label": "Decode captured DNS TXT payloads from 10.0.1.44 to assess data exposure, deploy DNS query length anomaly rules, and restore the clean host", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Quantifies exfiltrated data and implements detection rules against high-entropy DNS queries."},
                {"id": "d4_r2", "label": "Whitelist TXT queries to malicious-tunnel.com on the SIEM so host 10.0.1.44 can return to production without generating alerts", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Permits ongoing covert data exfiltration to the attacker's domain."},
                {"id": "d4_r3", "label": "Disable DNS query logging on the corporate DNS server after reconnecting 10.0.1.44 to improve resolver latency", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Eliminates essential DNS telemetry needed to detect covert C2 and exfiltration channels."}
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
                {"id": "d4_c3", "label": "Block outbound UDP port 123 entirely at the corporate firewall and isolate srv-sync-time-02 (10.0.3.15) from the network", "is_correct": False, "domain_id": "domain3", "points": -4, "explanation": "Breaks Kerberos and log time synchronization across the enterprise for an authorized change (CR-3212)."},
                {"id": "d4_c4", "label": "Correlate the UDP port 123 traffic from 10.0.3.15 with Change Request CR-3212 and classify the alert as a benign False Positive", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Confirms the high-frequency NTP polling to time.nist.gov (129.6.15.28) is an approved temporary workaround."},
                {"id": "d4_c5", "label": "Null-route government time server 129.6.15.28 on all core routers and revoke the sysadmin account that submitted CR-3212", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Disrupts legitimate upstream time synchronization and penalizes an authorized administrator."}
            ],
            "eradication": [
                {"id": "d4_e3", "label": "Add a temporary SIEM threshold exception for 10.0.3.15 polling time.nist.gov linked to CR-3212 until hardware replacement completes", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Suppresses repetitive benign alerts while maintaining visibility on other hosts."},
                {"id": "d4_e4", "label": "Kill the w32time.exe service on srv-sync-time-02 (10.0.3.15) and delete its system32 binary to stop UDP 123 packets", "is_correct": False, "domain_id": "domain2", "points": -3, "explanation": "Causes immediate clock drift on the backup time server and breaks Windows system integrity."},
                {"id": "d4_e5", "label": "Re-image srv-sync-time-02 (10.0.3.15) from cold storage backups assuming the NTP traffic to 129.6.15.28 is a DDoS amplification trojan", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Wastes incident response resources re-imaging a server experiencing a documented hardware oscillator issue."}
            ],
            "recovery": [
                {"id": "d4_r2", "label": "Verify the scheduled motherboard replacement for srv-sync-time-02 (10.0.3.15) and revert the NTP polling interval to baseline once repaired", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Tracks permanent remediation of the hardware root cause and restores standard NTP polling intervals."},
                {"id": "d4_r3", "label": "Configure all domain workstations to use unauthenticated external HTTP headers instead of NTP on 10.0.3.15 for clock sync", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Introduces insecure time synchronization that can break Kerberos authentication."},
                {"id": "d4_r4", "label": "Permanently disable UDP traffic monitoring on the perimeter firewall for all internal servers in subnet 10.0.3.0/24", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Creates a permanent blind spot for UDP-based data exfiltration or C2 traffic."}
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
                {"id": "d4_c5", "label": "Disable the rogue 'adm_backdoor' account on DC-01, revoke kwilliams's sessions, and block secure-office365-portal.com (192.0.2.14) at the proxy", "is_correct": True, "domain_id": "domain1", "points": 5, "explanation": "Removes the attacker's backdoor account, secures the compromised user, and blocks the phishing infrastructure."},
                {"id": "d4_c6", "label": "Shut down domain controller DC-01 completely while leaving kwilliams logged into WS-HR-11 and able to reach 192.0.2.14", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Causes a domain authentication outage while leaving the compromised endpoint and phishing channel active."},
                {"id": "d4_c7", "label": "Reply to admin@secure-office365-portal.com from WS-HR-11 requesting removal of kwilliams from their mailing list", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Tips off the threat actor at 192.0.2.14 without containing the compromised credentials or backdoor account."}
            ],
            "eradication": [
                {"id": "d4_e4", "label": "Purge all emails from secure-office365-portal.com across corporate mailboxes, remove adm_backdoor from DC-01, and scan WS-HR-11 for persistence", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Eliminates remaining phishing lures, removes the unauthorized admin account, and verifies endpoint cleanliness."},
                {"id": "d4_e5", "label": "Quarantine and format domain controller DC-01 without resetting kwilliams's password or purging the phishing email", "is_correct": False, "domain_id": "domain3", "points": -5, "explanation": "Formatting the primary DC breaks enterprise authentication while leaving the harvested credentials valid."},
                {"id": "d4_e6", "label": "Delete proxy logs for 10.0.2.77 on WS-HR-11 while leaving the 'adm_backdoor' account enabled in Active Directory", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Erases evidentiary proxy logs while preserving the attacker's administrative persistence."}
            ],
            "recovery": [
                {"id": "d4_r3", "label": "Re-enable kwilliams with a new password and phishing-resistant MFA, audit DC-01 group memberships, and assign targeted phishing training", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Safely restores user access with stronger authentication controls and addresses human-factor risk."},
                {"id": "d4_r4", "label": "Whitelist secure-office365-portal.com (192.0.2.14) on the email gateway so legitimate Microsoft notifications are not delayed", "is_correct": False, "domain_id": "domain2", "points": -3, "explanation": "secure-office365-portal.com is an external lookalike phishing domain, not a legitimate Microsoft service."},
                {"id": "d4_r5", "label": "Grant WS-HR-11 local administrator privileges on DC-01 so kwilliams can self-manage MFA resets in the future", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Violates least privilege and creates a severe privilege escalation path."}
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
                {"id": "d5_c1", "label": "Suspend rharper's Active Directory and CRM accounts immediately and isolate workstation WS-SALES-15 (10.0.1.18) to preserve forensic evidence", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Halts ongoing insider data exfiltration and preserves volatile and disk evidence on WS-SALES-15."},
                {"id": "d5_c2", "label": "Wipe and reformat workstation WS-SALES-15 immediately to erase all traces of Customer_Billing_Details_All.csv", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Destroys critical forensic evidence needed by Legal and HR to prove intentional insider exfiltration."},
                {"id": "d5_c3", "label": "Shut down the production CRM application for all sales staff while leaving rharper's account active on 10.0.1.18", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Disrupts enterprise sales operations while leaving the insider threat actor's account active."}
            ],
            "eradication": [
                {"id": "d5_e1", "label": "Perform a chain-of-custody forensic image of WS-SALES-15, revoke all API tokens for rharper, and block unauthorized personal webmail uploads via DLP", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Secures legally admissible forensic evidence, revokes residual access, and blocks personal webmail exfiltration vectors."},
                {"id": "d5_e2", "label": "Delete the CRM database audit logs and proxy records for 10.0.1.18 to keep the rharper investigation confidential", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Destroys primary audit evidence required for regulatory compliance and legal proceedings."},
                {"id": "d5_e3", "label": "Allow rharper to log back into WS-SALES-15 unmonitored to see if they delete additional files from the desktop", "is_correct": False, "domain_id": "domain1", "points": -3, "explanation": "Permits the insider to destroy remaining evidence or exfiltrate additional customer records."}
            ],
            "recovery": [
                {"id": "d5_r1", "label": "Coordinate with Legal and HR on insider incident handling, file GDPR/PDPA breach notifications within 72 hours, and enforce CRM bulk-export DLP limits", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Fulfills legal and regulatory PII breach mandates and implements preventive DLP controls on bulk CRM exports."},
                {"id": "d5_r2", "label": "Conceal the exfiltration of Customer_Billing_Details_All.csv from the Data Protection Officer to avoid regulatory fines", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Violates mandatory GDPR/PDPA breach notification laws and corporate governance policies."},
                {"id": "d5_r3", "label": "Grant all sales representatives unrestricted CRM bulk-export permissions so they can back up customer lists locally on their workstations", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Increases data exfiltration risk across the entire sales department."}
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
                {"id": "d5_c3", "label": "Revoke AWS IAM permissions for service principal backup_agent_04 and abort the Glacier transfer immediately", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Disrupts the pre-approved monthly compliance archiving job (CR-9022)."},
                {"id": "d5_c4", "label": "Verify Change Request CR-9022 for backup_agent_04 and acknowledge the S3-to-Glacier transfer alert as a benign False Positive", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Correctly validates the authorized monthly compliance archive without interrupting retention workflows."},
                {"id": "d5_c5", "label": "Delete the 'audit-logs-archive' S3 bucket to prevent further data migration to Amazon Glacier storage", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Destroys mandatory compliance audit logs during an authorized archiving task."}
            ],
            "eradication": [
                {"id": "d5_e2", "label": "Verify that the S3 lifecycle policy on 'audit-logs-archive' matches the retention parameters approved in CR-9022", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Confirms automated cloud data lifecycle governance is operating within approved compliance bounds."},
                {"id": "d5_e3", "label": "Delete the Glacier vault containing the 45.2 GB archive and lock the compliance_lead user account", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Destroys archived compliance records and locks out the authorizing compliance lead for a benign event."},
                {"id": "d5_e4", "label": "Disable AWS CloudTrail logging for all S3 and Glacier API calls made by service principals", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Removes essential cloud audit visibility across all storage buckets."}
            ],
            "recovery": [
                {"id": "d5_r3", "label": "Link the recurring monthly CR-9022 schedule for backup_agent_04 to the SIEM suppression ruleset and close the incident ticket", "is_correct": True, "domain_id": "domain5", "points": 5, "explanation": "Prevents recurring false positive alarms during future scheduled monthly compliance archives."},
                {"id": "d5_r4", "label": "Download the 45.2 GB Glacier archive to an unencrypted local workstation to manually inspect every audit log file", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Incurs unnecessary retrieval costs and exposes sensitive audit logs on an unencrypted endpoint."},
                {"id": "d5_r5", "label": "Configure 'audit-logs-archive' with public read permissions so external auditors can verify the CR-9022 backup directly", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Exposes sensitive internal security audit logs to the public internet."}
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
                {"id": "d5_c5", "label": "Administratively shut down switch interface Gi1/0/22 on SW-HQ-02 immediately and dispatch physical security to confiscate the rogue AP (00:e0:4c:00:a1:05)", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Immediately severs the rogue access point's physical LAN uplink and removes the unauthorized hardware."},
                {"id": "d5_c6", "label": "Increase the Port Security maximum MAC limit on SW-HQ-02 interface Gi1/0/22 from 1 to 50 so violation alerts stop firing", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Permits all unauthorized wireless clients on 'CORP-GUEST-FAST' to access the internal LAN."},
                {"id": "d5_c7", "label": "Power off core switch SW-HQ-02 completely, disconnecting all headquarters wired workstations and legitimate W-AP-3 access points", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Causes a building-wide network outage when shutting down interface Gi1/0/22 alone contains the rogue device."}
            ],
            "eradication": [
                {"id": "d5_e3", "label": "Enforce 802.1X port-based authentication with sticky MAC filtering on SW-HQ-02 access ports and revoke DHCP leases issued via Gi1/0/22", "is_correct": True, "domain_id": "domain3", "points": 5, "explanation": "Prevents unauthorized hardware from bridging onto the LAN and evicts rogue DHCP clients."},
                {"id": "d5_e4", "label": "Leave switch port Gi1/0/22 enabled on SW-HQ-02 while broadcasting a stronger signal on channel 6 from W-AP-3", "is_correct": False, "domain_id": "domain3", "points": -3, "explanation": "Leaves the physical LAN bridge active for any client connected to the rogue hardware."},
                {"id": "d5_e5", "label": "Clear Port Security violation counters on SW-HQ-02 without disconnecting MAC 00:e0:4c:00:a1:05 or auditing connected clients", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Resets the switch alarm while leaving the rogue wireless access point operating on the network."}
            ],
            "recovery": [
                {"id": "d5_r4", "label": "Audit DHCP and proxy logs for all 12 MAC addresses seen on Gi1/0/22, rotate credentials for affected corporate users, and conduct a wireless site survey", "is_correct": True, "domain_id": "domain4", "points": 5, "explanation": "Identifies any corporate devices that associated with 'CORP-GUEST-FAST' and verifies no other rogue APs remain."},
                {"id": "d5_r5", "label": "Adopt the rogue SSID 'CORP-GUEST-FAST' into the official corporate wireless controller configuration without inspecting the hardware", "is_correct": False, "domain_id": "domain5", "points": -3, "explanation": "Legitimizes an untrusted, potentially backdoored consumer device on the corporate LAN."},
                {"id": "d5_r6", "label": "Disable wireless IDS rogue SSID detection on W-AP-3 and Port Security logging on SW-HQ-02 to reduce SOC ticket volume", "is_correct": False, "domain_id": "domain4", "points": -3, "explanation": "Removes both physical port security and wireless intrusion detection controls."}
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
