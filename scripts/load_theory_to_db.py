import os
import re
import sys


DOMAIN_SPECS = [
    ("domain1", "Domain 1: General Security Concepts", "domain1_security_concepts.md", "https://csrc.nist.gov/pubs/sp/800/53/r5/final"),
    ("domain2", "Domain 2: Threats, Vulnerabilities, and Mitigations", "domain2_threats_vulnerabilities.md", "https://owasp.org/Top10/"),
    ("domain3", "Domain 3: Security Architecture", "domain3_security_architecture.md", "https://csrc.nist.gov/pubs/sp/800/207/final"),
    ("domain4", "Domain 4: Security Operations", "domain4_security_operations.md", "https://csrc.nist.gov/pubs/sp/800/61/r3/final"),
    ("domain5", "Domain 5: Security Program Management and Oversight", "domain5_security_management.md", "https://www.cmu.edu/iso/governance/procedures/incident-response.html"),
]

SUBTOPIC_SPECS = [
    ("1.1", "domain1", 1, "Compare and contrast various types of security controls", "domain1/1_1_security_controls.md", "https://csrc.nist.gov/pubs/sp/800/53/r5/final"),
    ("1.2", "domain1", 2, "Summarize fundamental security concepts", "domain1/1_2_fundamental_security_concepts.md", "https://csrc.nist.gov/pubs/sp/800/207/final"),
    ("1.3", "domain1", 3, "Explain change management best practices", "domain1/1_3_change_management.md", "https://en.wikipedia.org/wiki/Change_control"),
    ("1.4", "domain1", 4, "Explain the importance of using appropriate cryptographic solutions", "domain1/1_4_cryptographic_solutions.md", "https://en.wikipedia.org/wiki/Public_key_infrastructure"),
    ("2.1", "domain2", 1, "Compare and contrast threat actors and intelligence sources", "domain2/2_1_threat_actors_and_motivations.md", "https://attack.mitre.org/matrices/enterprise/"),
    ("2.2", "domain2", 2, "Analyze common threat vectors and attack surfaces", "domain2/2_2_threat_vectors_and_attack_surfaces.md", "https://www.cisa.gov/topics/cyber-threats-and-responses/phishing"),
    ("2.3", "domain2", 3, "Explain various types of vulnerabilities", "domain2/2_3_types_of_vulnerabilities.md", "https://owasp.org/Top10/"),
    ("2.4", "domain2", 4, "Analyze indicators of malicious activity", "domain2/2_4_indicators_of_malicious_activity.md", "https://en.wikipedia.org/wiki/Indicator_of_compromise"),
    ("2.5", "domain2", 5, "Explain mitigation techniques used to secure the enterprise", "domain2/2_5_mitigation_techniques.md", "https://csrc.nist.gov/pubs/sp/800/40/r4/final"),
    ("3.1", "domain3", 1, "Compare and contrast security architecture models", "domain3/3_1_security_architecture_models.md", "https://csrc.nist.gov/pubs/sp/800/207/final"),
    ("3.2", "domain3", 2, "Apply security principles to secure enterprise infrastructure", "domain3/3_2_enterprise_infrastructure_security.md", "https://en.wikipedia.org/wiki/Defense_in_depth_%28computing%29"),
    ("3.3", "domain3", 3, "Compare and contrast concepts and strategies to protect data", "domain3/3_3_data_protection_strategies.md", "https://en.wikipedia.org/wiki/Data_loss_prevention_software"),
    ("3.4", "domain3", 4, "Explain the importance of resilience and recovery in security architecture", "domain3/3_4_resilience_and_recovery.md", "https://www.cisa.gov/stopransomware"),
    ("4.1", "domain4", 1, "Apply common security techniques to computing resources", "domain4/4_1_computing_resource_security.md", "https://en.wikipedia.org/wiki/Endpoint_detection_and_response"),
    ("4.2", "domain4", 2, "Explain security implications of hardware, software, and data asset management", "domain4/4_2_asset_management.md", "https://csrc.nist.gov/pubs/sp/800/88/r1/final"),
    ("4.3", "domain4", 3, "Explain various activities associated with vulnerability management", "domain4/4_3_vulnerability_management.md", "https://en.wikipedia.org/wiki/Vulnerability_management"),
    ("4.4", "domain4", 4, "Explain security alerting and monitoring concepts and tools", "domain4/4_4_security_alerting_and_monitoring.md", "https://en.wikipedia.org/wiki/Security_information_and_event_management"),
    ("4.5", "domain4", 5, "Modify enterprise capabilities to enhance security", "domain4/4_5_modifying_enterprise_capabilities.md", "https://csrc.nist.gov/pubs/sp/800/53/r5/final"),
    ("4.6", "domain4", 6, "Implement and maintain identity and access management", "domain4/4_6_identity_and_access_management.md", "https://csrc.nist.gov/pubs/sp/800/63/3/final"),
    ("4.7", "domain4", 7, "Explain the importance of automation and orchestration", "domain4/4_7_automation_and_orchestration.md", "https://en.wikipedia.org/wiki/Security_orchestration"),
    ("4.8", "domain4", 8, "Explain appropriate incident response activities", "domain4/4_8_incident_response_activities.md", "https://csrc.nist.gov/pubs/sp/800/61/r3/final"),
    ("4.9", "domain4", 9, "Use data sources to support an investigation and digital forensics", "domain4/4_9_data_sources_and_digital_forensics.md", "https://csrc.nist.gov/pubs/sp/800/86/final"),
    ("5.1", "domain5", 1, "Summarize elements of effective security governance", "domain5/5_1_security_governance.md", "https://en.wikipedia.org/wiki/Information_security_governance"),
    ("5.2", "domain5", 2, "Explain elements of the risk management process", "domain5/5_2_risk_management_process.md", "https://csrc.nist.gov/pubs/sp/800/30/r1/final"),
    ("5.3", "domain5", 3, "Explain third-party risk assessment and management processes", "domain5/5_3_third_party_risk_management.md", "https://csrc.nist.gov/pubs/sp/800/161/r1/final"),
    ("5.4", "domain5", 4, "Summarize elements of effective security compliance", "domain5/5_4_security_compliance.md", "https://en.wikipedia.org/wiki/Regulatory_compliance"),
    ("5.5", "domain5", 5, "Explain types and purposes of audits and assessments", "domain5/5_5_audits_and_assessments.md", "https://csrc.nist.gov/pubs/sp/800/115/final"),
    ("5.6", "domain5", 6, "Implement security awareness practices", "domain5/5_6_security_awareness_practices.md", "https://csrc.nist.gov/pubs/sp/800/50/r1/final"),
]

REQUIRED_SUBTOPIC_SECTION_PATTERNS = [
    ("Section 1 (Core Security+ Theory)", r"^##\s+1\.\s+Core Security\+ Theory"),
    ("Section 2 (SOC & Blue-Team Practice)", r"^##\s+2\.\s+SOC & Blue-Team"),
    ("Section 3 (References)", r"^##\s+3\.\s+References"),
]

REQUIRED_SECTION_PATTERNS = [
    ("Section 1 (Core Security+ Theory)", r"^##\s+1\.\s+Core Security\+ Theory"),
    ("Section 2 (SOC Telemetry & MITRE)", r"^##\s+2\.\s+SOC Telemetry"),
    ("Section 3 (Alert Triage & Playbook)", r"^##\s+3\.\s+Alert Triage"),
    ("Section 4 (Key Takeaways & References)", r"^##\s+4\.\s+Key Takeaways"),
]


def resolve_theory_dir(theory_dir=None):
    if theory_dir:
        return theory_dir
    script_dir = os.path.dirname(os.path.abspath(__file__))
    candidates = [
        os.path.join(os.getcwd(), "theory"),
        os.path.join(script_dir, "..", "theory"),
        "/app/theory",
    ]
    for candidate in candidates:
        if os.path.isdir(candidate):
            return candidate
    raise FileNotFoundError(
        "Could not locate theory directory (checked ./theory, ../theory, and /app/theory)."
    )


def validate_theory_content(domain_id, title, content):
    errors = []
    if not content or not content.strip():
        errors.append(f"{domain_id}: theory markdown content is empty")
        return errors

    first_line = content.lstrip().splitlines()[0].strip()
    if not first_line.startswith("# "):
        errors.append(f"{domain_id}: expected top-level '# ' heading, got '{first_line[:40]}'")

    for label, pattern in REQUIRED_SECTION_PATTERNS:
        if not re.search(pattern, content, flags=re.MULTILINE):
            errors.append(f"{domain_id} ({title}): missing required {label}")

    if "kql" not in content.lower():
        errors.append(f"{domain_id} ({title}): missing KQL investigation patterns")

    if "false positive" not in content.lower() or "true positive" not in content.lower():
        errors.append(f"{domain_id} ({title}): missing True Positive / False Positive triage guidance")

    if "](https://" not in content:
        errors.append(f"{domain_id} ({title}): missing clickable https:// reference links")

    return errors


def load_and_validate_modules(theory_dir=None):
    resolved_dir = resolve_theory_dir(theory_dir)
    modules = []
    all_errors = []

    for domain_id, title, filename, source_url in DOMAIN_SPECS:
        file_path = os.path.join(resolved_dir, filename)
        if not os.path.isfile(file_path):
            all_errors.append(f"{domain_id}: missing file {file_path}")
            continue

        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read().replace("\r\n", "\n")

        errs = validate_theory_content(domain_id, title, content)
        all_errors.extend(errs)
        modules.append(
            {
                "domain_id": domain_id,
                "title": title,
                "content": content,
                "source_url": source_url,
            }
        )

    if all_errors:
        raise ValueError("Theory validation failed:\n- " + "\n- ".join(all_errors))

    return modules


def validate_subtopic_content(sub_topic_id, title, content):
    errors = []
    if not content or not content.strip():
        errors.append(f"{sub_topic_id}: sub-topic markdown content is empty")
        return errors

    first_line = content.lstrip().splitlines()[0].strip()
    if not first_line.startswith(f"# {sub_topic_id}"):
        errors.append(f"{sub_topic_id}: expected heading starting with '# {sub_topic_id}', got '{first_line[:50]}'")

    for label, pattern in REQUIRED_SUBTOPIC_SECTION_PATTERNS:
        if not re.search(pattern, content, flags=re.MULTILINE):
            errors.append(f"{sub_topic_id} ({title}): missing required {label}")

    if "](https://" not in content:
        errors.append(f"{sub_topic_id} ({title}): missing clickable https:// reference links")

    return errors


def load_and_validate_subtopics(theory_dir=None):
    resolved_dir = resolve_theory_dir(theory_dir)
    subtopics = []
    all_errors = []

    for sub_topic_id, domain_id, order_index, title, rel_path, source_url in SUBTOPIC_SPECS:
        file_path = os.path.join(resolved_dir, rel_path)
        if not os.path.isfile(file_path):
            all_errors.append(f"{sub_topic_id}: missing file {file_path}")
            continue

        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read().replace("\r\n", "\n")

        errs = validate_subtopic_content(sub_topic_id, title, content)
        all_errors.extend(errs)
        subtopics.append(
            {
                "sub_topic_id": sub_topic_id,
                "domain_id": domain_id,
                "order_index": order_index,
                "title": title,
                "content": content,
                "source_url": source_url,
            }
        )

    if all_errors:
        raise ValueError("Sub-topic validation failed:\n- " + "\n- ".join(all_errors))

    return subtopics


def load_theory_to_db(db_connection_string, theory_dir=None):
    import psycopg2

    modules = load_and_validate_modules(theory_dir)
    subtopics = load_and_validate_subtopics(theory_dir)
    conn = psycopg2.connect(db_connection_string)
    cursor = conn.cursor()

    try:
        cursor.execute(
            """
            CREATE TABLE IF NOT EXISTS theory_subtopics (
                sub_topic_id VARCHAR(16) PRIMARY KEY,
                domain_id VARCHAR REFERENCES security_domains(domain_id) ON DELETE CASCADE NOT NULL,
                order_index INT NOT NULL,
                title VARCHAR(255) NOT NULL,
                content TEXT NOT NULL,
                source_url TEXT,
                updated_at TIMESTAMP DEFAULT now()
            )
            """
        )
        domains = [
            ("domain1", "Security Concepts", "Core security concepts, principles, and control types", 1.0, "Security+ Domain 1"),
            ("domain2", "Threats, Vulnerabilities, and Mitigations", "Analyzing threat actors, vectors, vulnerabilities, and applying mitigations", 1.0, "Security+ Domain 2"),
            ("domain3", "Security Architecture", "Designing secure network, cloud, and host infrastructures", 1.0, "Security+ Domain 3"),
            ("domain4", "Security Operations", "Security monitoring, incident response, and operational security controls", 1.0, "Security+ Domain 4"),
            ("domain5", "Security Management", "Governance, risk management, compliance, and security policies", 1.0, "Security+ Domain 5"),
        ]
        for domain_id, name, desc, weight, ref in domains:
            cursor.execute(
                """
                INSERT INTO security_domains (domain_id, name, description, weight_score, compTIA_reference)
                VALUES (%s, %s, %s, %s, %s)
                ON CONFLICT (domain_id) DO NOTHING
                """,
                (domain_id, name, desc, weight, ref),
            )

        inserted_count = 0
        updated_count = 0

        for mod in modules:
            cursor.execute(
                """
                SELECT doc_id FROM rag_documents
                WHERE domain_id = %s AND title = %s
                """,
                (mod["domain_id"], mod["title"]),
            )
            existing = cursor.fetchone()
            if existing:
                cursor.execute(
                    """
                    UPDATE rag_documents
                    SET content = %s,
                        source_url = %s,
                        updated_at = NOW()
                    WHERE doc_id = %s
                    """,
                    (mod["content"], mod["source_url"], existing[0]),
                )
                updated_count += 1
            else:
                cursor.execute(
                    """
                    INSERT INTO rag_documents (title, content, domain_id, source_url)
                    VALUES (%s, %s, %s, %s)
                    """,
                    (mod["title"], mod["content"], mod["domain_id"], mod["source_url"]),
                )
                inserted_count += 1

        for sub in subtopics:
            cursor.execute(
                """
                INSERT INTO theory_subtopics (sub_topic_id, domain_id, order_index, title, content, source_url, updated_at)
                VALUES (%s, %s, %s, %s, %s, %s, NOW())
                ON CONFLICT (sub_topic_id) DO UPDATE SET
                    domain_id = EXCLUDED.domain_id,
                    order_index = EXCLUDED.order_index,
                    title = EXCLUDED.title,
                    content = EXCLUDED.content,
                    source_url = EXCLUDED.source_url,
                    updated_at = NOW()
                """,
                (
                    sub["sub_topic_id"],
                    sub["domain_id"],
                    sub["order_index"],
                    sub["title"],
                    sub["content"],
                    sub["source_url"],
                ),
            )

        conn.commit()
        print(
            f"Safe theory seed completed: {inserted_count} inserted, {updated_count} updated across {len(modules)} domains and {len(subtopics)} sub-topics."
        )
    except Exception:
        conn.rollback()
        raise
    finally:
        cursor.close()
        conn.close()


if __name__ == "__main__":
    db_conn_str = os.environ.get("DATABASE_URL")
    if not db_conn_str:
        print(
            "DATABASE_URL environment variable is required (no fallback credentials).",
            file=sys.stderr,
        )
        sys.exit(1)

    try:
        load_theory_to_db(db_conn_str)
    except Exception as e:
        print(f"Loading theory modules failed: {e}", file=sys.stderr)
        sys.exit(1)
