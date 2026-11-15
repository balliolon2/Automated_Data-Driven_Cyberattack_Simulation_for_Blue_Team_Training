import os
import re
import sys


DOMAIN_SPECS = [
    ("domain1", "Domain 1: General Security Concepts", "domain1_security_concepts.md"),
    ("domain2", "Domain 2: Threats, Vulnerabilities, and Mitigations", "domain2_threats_vulnerabilities.md"),
    ("domain3", "Domain 3: Security Architecture", "domain3_security_architecture.md"),
    ("domain4", "Domain 4: Security Operations", "domain4_security_operations.md"),
    ("domain5", "Domain 5: Security Program Management and Oversight", "domain5_security_management.md"),
]

REQUIRED_SECTION_PATTERNS = [
    ("Section 1 (Core Security+ Theory)", r"^##\s+1\.\s+Core Security\+ Theory"),
    ("Section 2 (SOC Telemetry & MITRE)", r"^##\s+2\.\s+SOC Telemetry"),
    ("Section 3 (Alert Triage & Playbook)", r"^##\s+3\.\s+Alert Triage"),
    ("Section 4 (Key Takeaways & References)", r"^##\s+4\.\s+Key Takeaways"),
]

SOURCE_REF_URL = "docs/REFERENCES.md#14-learning-theory-curriculum-provenance-r2r5"


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

    return errors


def load_and_validate_modules(theory_dir=None):
    resolved_dir = resolve_theory_dir(theory_dir)
    modules = []
    all_errors = []

    for domain_id, title, filename in DOMAIN_SPECS:
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
                "source_url": SOURCE_REF_URL,
            }
        )

    if all_errors:
        raise ValueError("Theory validation failed:\n- " + "\n- ".join(all_errors))

    return modules


def load_theory_to_db(db_connection_string, theory_dir=None):
    import psycopg2

    modules = load_and_validate_modules(theory_dir)
    conn = psycopg2.connect(db_connection_string)
    cursor = conn.cursor()

    try:
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

        conn.commit()
        print(
            f"Safe theory seed completed: {inserted_count} inserted, {updated_count} updated across {len(modules)} domains."
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
