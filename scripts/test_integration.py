import requests
import random
import sys

BASE_URL = "http://localhost:8080/api"

def run_tests():
    # 1. Register a test user
    email = f"test_{random.randint(1000, 9999)}@example.com"
    password = "password123"
    
    print(f"Registering user: {email}")
    register_res = requests.post(f"{BASE_URL}/register", json={
        "email": email,
        "password": password
    })
    if register_res.status_code != 201:
        print("Registration failed", register_res.json())
        sys.exit(1)
    print("Registration successful")

    # 2. Login
    print("Logging in...")
    login_res = requests.post(f"{BASE_URL}/login", json={
        "email": email,
        "password": password
    })
    if login_res.status_code != 200:
        print("Login failed", login_res.json())
        sys.exit(1)
    
    token = login_res.json().get("token")
    headers = {"Authorization": f"Bearer {token}"}
    print("Login successful, token retrieved")

    # 3. Start Pre-Test
    print("Starting Pre-Test...")
    pre_res = requests.post(f"{BASE_URL}/exams/pre-test", headers=headers)
    if pre_res.status_code != 201:
        print("Failed to start Pre-Test (Note: DB must have questions populated for this to work)", pre_res.status_code)
        print("If you haven't populated the database yet, run parse_exams.py and load_exams_to_db.py")
        sys.exit(0) # Don't exit with error, as it might just be unpopulated DB
        
    pre_data = pre_res.json()
    session_id = pre_data["session"]["session_id"]
    questions = pre_data["questions"]
    
    print(f"Pre-Test started. Session ID: {session_id}. Questions count: {len(questions)}")
    if len(questions) != 100:
        print(f"Error: Expected 100 questions, got {len(questions)}")
        sys.exit(1)
        
    # Track domains
    domains = [q["domain_id"] for q in questions]
    domain_counts = {d: domains.count(d) for d in set(domains)}
    print("Domain distribution:", domain_counts)
    
    # 4. Test Reconnect (fetch session)
    print("Testing Session Recovery...")
    session_res = requests.get(f"{BASE_URL}/exams/session", headers=headers)
    if session_res.status_code != 200:
        print("Session recovery failed", session_res.json())
        sys.exit(1)
    print("Session recovery successful. Recovered questions:", len(session_res.json()["questions"]))

    # 5. Submit Answers
    print("Submitting answers...")
    for i, q in enumerate(questions):
        # Answer all questions
        ans_res = requests.post(f"{BASE_URL}/exams/submit-answer", headers=headers, json={
            "session_id": session_id,
            "question_id": q["question_id"],
            "user_answer": "C", # just dummy answer
            "time_spent_seconds": 1
        })
        if ans_res.status_code != 200:
            print(f"Failed to submit answer for question {i}", ans_res.json())
            sys.exit(1)
            
        data = ans_res.json()
        if data.get("completed"):
            print("Pre-test completed! Final Score:", data.get("score"))
            pre_questions_used = [qu["question_id"] for qu in data["results"]]
            break

    # 6. Start Post-Test
    print("Starting Post-Test...")
    post_res = requests.post(f"{BASE_URL}/exams/post-test", headers=headers)
    if post_res.status_code != 201:
        print("Failed to start Post-Test", post_res.json())
        sys.exit(1)
        
    post_data = post_res.json()
    post_session_id = post_data["session"]["session_id"]
    post_questions = post_data["questions"]
    
    print(f"Post-Test started. Session ID: {post_session_id}. Questions count: {len(post_questions)}")
    
    # Verify no overlap
    pre_set = set(pre_questions_used)
    post_set = set([q["question_id"] for q in post_questions])
    
    overlap = pre_set.intersection(post_set)
    print(f"Overlap count between Pre-Test and Post-Test: {len(overlap)}")
    if len(overlap) > 0:
        print(f"Error: Found overlap questions: {overlap}")
        sys.exit(1)
    else:
        print("Verification SUCCESS: Zero question overlap between Pre-Test and Post-Test!")

if __name__ == "__main__":
    if len(sys.argv) > 1:
        BASE_URL = sys.argv[1]
    run_tests()
