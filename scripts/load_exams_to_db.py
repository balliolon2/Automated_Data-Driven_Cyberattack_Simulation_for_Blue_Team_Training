import json
import psycopg2
import os

def load_json_to_db(json_file_path, db_connection_string):
    with open(json_file_path, 'r', encoding='utf-8') as f:
        questions = json.load(f)

    # Connect to your postgres DB
    conn = psycopg2.connect(db_connection_string)
    cursor = conn.cursor()

    try:
        # Pre-populate security_domains to satisfy foreign key constraints
        domains = [
            ("domain1", "Security Concepts", "Core security concepts, principles, and control types", 1.0, "Security+ Domain 1"),
            ("domain2", "Threats, Vulnerabilities, and Mitigations", "Analyzing threat actors, vectors, vulnerabilities, and applying mitigations", 1.0, "Security+ Domain 2"),
            ("domain3", "Security Architecture", "Designing secure network, cloud, and host infrastructures", 1.0, "Security+ Domain 3"),
            ("domain4", "Security Operations", "Security monitoring, incident response, and operational security controls", 1.0, "Security+ Domain 4"),
            ("domain5", "Security Management", "Governance, risk management, compliance, and security policies", 1.0, "Security+ Domain 5")
        ]
        for domain_id, name, desc, weight, ref in domains:
            cursor.execute("""
                INSERT INTO security_domains (domain_id, name, description, weight_score, compTIA_reference)
                VALUES (%s, %s, %s, %s, %s)
                ON CONFLICT (domain_id) DO NOTHING
            """, (domain_id, name, desc, weight, ref))
        print("Successfully seeded security domains.")

        inserted_count = 0
        for q in questions:
            # Assuming your schema allows omitting question_id to auto-generate UUID
            # and that 'domain_id' matches the existing security_domains
            cursor.execute("""
                INSERT INTO questions 
                (domain_id, type, question_text, options, correct_answer, explanation, difficulty) 
                VALUES (%s, %s, %s, %s, %s, %s, %s)
            """, (
                q['domain_id'],
                q['type'],
                q['question_text'],
                json.dumps(q['options']),
                q['correct_answer'],
                q['explanation'],
                q['difficulty']
            ))
            inserted_count += 1
            
        conn.commit()
        print(f"Successfully inserted {inserted_count} questions into the database.")
    except Exception as e:
        conn.rollback()
        print(f"Database insertion failed: {e}")
    finally:
        cursor.close()
        conn.close()

if __name__ == "__main__":
    # Update with your actual database connection string
    # e.g. "dbname=cyber_sim user=admin password=password host=localhost port=5432"
    db_conn_str = os.environ.get("DATABASE_URL", "dbname=cyber_sim user=admin password=password host=localhost port=5432")
    json_path = "parsed_questions.json"
    
    if os.path.exists(json_path):
        load_json_to_db(json_path, db_conn_str)
    else:
        print(f"File {json_path} not found. Please run parse_exams.py first.")
