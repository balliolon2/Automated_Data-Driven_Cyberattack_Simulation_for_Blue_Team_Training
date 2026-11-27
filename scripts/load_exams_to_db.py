import json
import os
import sys

import psycopg2

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
if SCRIPT_DIR not in sys.path:
    sys.path.insert(0, SCRIPT_DIR)

import translate_questions_th


def load_optional_question_translations(questions, translations_path=None):
    target = translations_path or os.path.join(SCRIPT_DIR, "question_translations_th.json")
    if not os.path.isfile(target):
        return {}
    with open(target, "r", encoding="utf-8") as f:
        translations = json.load(f)
    errs, _ = translate_questions_th.validate_translations_dataset(
        questions, translations, require_full=False
    )
    if errs:
        raise ValueError("Invalid question_translations_th.json:\n- " + "\n- ".join(errs[:15]))
    return translations


def load_json_to_db(json_file_path, db_connection_string, translations_path=None):
    with open(json_file_path, "r", encoding="utf-8") as f:
        questions = json.load(f)

    if not questions:
        raise ValueError(f"No questions found in {json_file_path}; refusing to seed an empty question pool.")

    for q in questions:
        sub_topic_id = (q.get("sub_topic_id") or "").strip()
        if not sub_topic_id or sub_topic_id[0] != q["domain_id"][-1]:
            raise ValueError(
                f"Question #{q.get('source_question_no')} ({q['domain_id']}) has missing or mismatched sub_topic_id: {sub_topic_id!r}"
            )

    translations_map = load_optional_question_translations(questions, translations_path=translations_path)

    conn = psycopg2.connect(db_connection_string)
    cursor = conn.cursor()

    try:
        cursor.execute("ALTER TABLE questions ADD COLUMN IF NOT EXISTS sub_topic_id VARCHAR(16)")
        cursor.execute("ALTER TABLE questions ADD COLUMN IF NOT EXISTS translations JSONB NOT NULL DEFAULT '{}'::jsonb")
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
        print("Successfully seeded security domains.")

        inserted_count = 0
        updated_count = 0
        translated_count = 0

        for q in questions:
            q_key = translate_questions_th.question_key(q) if q.get("source_question_no") is not None else ""
            t_entry = translations_map.get(q_key)
            has_translation = bool(t_entry)
            translations_payload = {"th": t_entry["th"]} if has_translation else {}
            if has_translation:
                translated_count += 1
            translations_json = json.dumps(translations_payload, ensure_ascii=False)

            cursor.execute(
                """
                SELECT question_id
                FROM questions
                WHERE domain_id = %s
                  AND btrim(regexp_replace(regexp_replace(question_text, '[*][*]', '', 'g'), '[[:space:]]+', ' ', 'g')) = %s
                """,
                (q["domain_id"], q["question_text"]),
            )
            existing = cursor.fetchone()

            if existing:
                cursor.execute(
                    """
                    UPDATE questions SET
                        sub_topic_id = %s,
                        type = %s,
                        question_text = %s,
                        options = %s,
                        correct_answer = %s,
                        explanation = %s,
                        translations = CASE WHEN %s THEN %s::jsonb ELSE translations END,
                        difficulty = %s,
                        is_active = true
                    WHERE question_id = %s
                    """,
                    (
                        q["sub_topic_id"],
                        q["type"],
                        q["question_text"],
                        json.dumps(q["options"]),
                        q["correct_answer"],
                        q["explanation"],
                        has_translation,
                        translations_json,
                        q["difficulty"],
                        existing[0],
                    ),
                )
                updated_count += 1
            else:
                cursor.execute(
                    """
                    INSERT INTO questions
                    (domain_id, sub_topic_id, type, question_text, options, correct_answer, explanation, translations, difficulty)
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
                    """,
                    (
                        q["domain_id"],
                        q["sub_topic_id"],
                        q["type"],
                        q["question_text"],
                        json.dumps(q["options"]),
                        q["correct_answer"],
                        q["explanation"],
                        translations_json,
                        q["difficulty"],
                    ),
                )
                inserted_count += 1

        conn.commit()
        print(f"Safe question seed completed: {inserted_count} inserted, {updated_count} updated ({translated_count} with Thai translations).")
    except Exception:
        conn.rollback()
        raise
    finally:
        cursor.close()
        conn.close()


if __name__ == "__main__":
    db_conn_str = os.environ.get("DATABASE_URL")
    json_path = "parsed_questions.json"

    if not db_conn_str:
        print("DATABASE_URL environment variable is required (no fallback credentials).", file=sys.stderr)
        sys.exit(1)

    if not os.path.isfile(json_path):
        print(f"File {json_path} not found. Please run parse_exams.py first.", file=sys.stderr)
        sys.exit(1)

    try:
        load_json_to_db(json_path, db_conn_str)
    except Exception as e:
        print(f"Loading {json_path} failed: {e}", file=sys.stderr)
        sys.exit(1)
