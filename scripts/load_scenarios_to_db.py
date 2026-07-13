import json
import psycopg2
import os

def load_scenarios_to_db(json_file_path, db_connection_string):
    with open(json_file_path, 'r', encoding='utf-8') as f:
        scenarios = json.load(f)

    conn = psycopg2.connect(db_connection_string)
    cursor = conn.cursor()

    try:
        # Run schema migration for scenarios columns if they don't exist
        cursor.execute("""
            ALTER TABLE scenarios 
            ADD COLUMN IF NOT EXISTS is_true_positive BOOLEAN NOT NULL DEFAULT true,
            ADD COLUMN IF NOT EXISTS tp_fp_explanation TEXT;
        """)
        conn.commit()
        print("Database migration check completed (added scenarios columns if missing).")
        
        inserted_count = 0
        
        # Clear existing scenarios and dependent sessions to prevent duplicates on re-run
        cursor.execute("DELETE FROM session_actions WHERE session_id IN (SELECT session_id FROM simulation_sessions)")
        cursor.execute("DELETE FROM simulation_sessions")
        cursor.execute("DELETE FROM scenarios")
        conn.commit()
        
        for s in scenarios:
            cursor.execute("""
                INSERT INTO scenarios 
                (title, description, domain_id, difficulty, initial_logs, playbook_steps, expected_outcomes, is_true_positive, tp_fp_explanation, status) 
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """, (
                s['title'],
                s['description'],
                s['domain_id'],
                s['difficulty'],
                json.dumps(s['initial_logs']),
                json.dumps(s['playbook_steps']),
                json.dumps(s['expected_outcomes']),
                s['is_true_positive'],
                s['tp_fp_explanation'],
                s.get('status', 'active')
            ))
            inserted_count += 1
            
        conn.commit()
        print(f"Successfully inserted {inserted_count} scenarios into the database.")
    except Exception as e:
        conn.rollback()
        print(f"Database insertion failed: {e}")
    finally:
        cursor.close()
        conn.close()

if __name__ == "__main__":
    db_conn_str = os.environ.get("DATABASE_URL", "dbname=cyber_sim user=admin password=password host=localhost port=5432")
    json_path = "scenarios_seed.json"
    
    if os.path.exists(json_path):
        load_scenarios_to_db(json_path, db_conn_str)
    else:
        print(f"File {json_path} not found. Please run scripts/generate_scenarios.py first.")
