import json
import os
import sys

import psycopg2


def load_scenarios_to_db(json_file_path, db_connection_string):
    with open(json_file_path, "r", encoding="utf-8") as f:
        scenarios = json.load(f)

    if not scenarios:
        raise ValueError(f"No scenarios found in {json_file_path}; refusing to seed an empty scenario pool.")

    conn = psycopg2.connect(db_connection_string)
    cursor = conn.cursor()

    try:
        cursor.execute(
            """
            ALTER TABLE scenarios
            ADD COLUMN IF NOT EXISTS is_true_positive BOOLEAN NOT NULL DEFAULT true,
            ADD COLUMN IF NOT EXISTS tp_fp_explanation TEXT;
            """
        )
        conn.commit()
        print("Database migration check completed (added scenarios columns if missing).")

        inserted_count = 0
        updated_count = 0

        for s in scenarios:
            cursor.execute(
                """
                SELECT scenario_id FROM scenarios WHERE title = %s AND domain_id = %s
                """,
                (s["title"], s["domain_id"]),
            )
            existing = cursor.fetchone()

            if existing:
                cursor.execute(
                    """
                    UPDATE scenarios SET
                        description = %s,
                        difficulty = %s,
                        initial_logs = %s,
                        playbook_steps = %s,
                        expected_outcomes = %s,
                        is_true_positive = %s,
                        tp_fp_explanation = %s,
                        status = %s,
                        updated_at = NOW()
                    WHERE scenario_id = %s
                    """,
                    (
                        s["description"],
                        s["difficulty"],
                        json.dumps(s["initial_logs"]),
                        json.dumps(s["playbook_steps"]),
                        json.dumps(s["expected_outcomes"]),
                        s["is_true_positive"],
                        s["tp_fp_explanation"],
                        s.get("status", "active"),
                        existing[0],
                    ),
                )
                updated_count += 1
            else:
                cursor.execute(
                    """
                    INSERT INTO scenarios
                    (title, description, domain_id, difficulty, initial_logs, playbook_steps, expected_outcomes, is_true_positive, tp_fp_explanation, status)
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    """,
                    (
                        s["title"],
                        s["description"],
                        s["domain_id"],
                        s["difficulty"],
                        json.dumps(s["initial_logs"]),
                        json.dumps(s["playbook_steps"]),
                        json.dumps(s["expected_outcomes"]),
                        s["is_true_positive"],
                        s["tp_fp_explanation"],
                        s.get("status", "active"),
                    ),
                )
                inserted_count += 1

        conn.commit()
        print(f"Safe seed completed: {inserted_count} inserted, {updated_count} updated. No session data modified.")
    except Exception:
        conn.rollback()
        raise
    finally:
        cursor.close()
        conn.close()


if __name__ == "__main__":
    db_conn_str = os.environ.get("DATABASE_URL")
    json_path = "scenarios_seed.json"

    if not db_conn_str:
        print("DATABASE_URL environment variable is required (no fallback credentials).", file=sys.stderr)
        sys.exit(1)

    if not os.path.isfile(json_path):
        print(f"File {json_path} not found. Please run scripts/generate_scenarios.py first.", file=sys.stderr)
        sys.exit(1)

    try:
        load_scenarios_to_db(json_path, db_conn_str)
    except Exception as e:
        print(f"Loading {json_path} failed: {e}", file=sys.stderr)
        sys.exit(1)
