import requests
import random
import sys

BASE_URL = "http://localhost:8080/api"

def run_tests():
    # 1. Register a test user
    email = f"sim_test_{random.randint(1000, 9999)}@example.com"
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

    # 3. Verify simulation status BEFORE Pre-test is complete (should fail/warn)
    print("Verifying simulation status before pre-test is completed...")
    status_res = requests.get(f"{BASE_URL}/simulation/status", headers=headers)
    if status_res.status_code == 200:
        data = status_res.json()
        if not data.get("error"):
            print("Error: Expected warning that skill profile is empty, but got success status:", data)
            sys.exit(1)
        print("Received expected warning:", data["error"])
    else:
        print("Received expected error response:", status_res.status_code)

    # 4. Start Pre-Test
    print("Starting Pre-Test...")
    pre_res = requests.post(f"{BASE_URL}/exams/pre-test", headers=headers)
    if pre_res.status_code != 201:
        print("Failed to start Pre-Test (Note: DB must have questions populated for this to work)", pre_res.status_code)
        sys.exit(0)
        
    pre_data = pre_res.json()
    session_id = pre_data["session"]["session_id"]
    questions = pre_data["questions"]
    print(f"Pre-Test started. Session ID: {session_id}. Questions count: {len(questions)}")

    # 5. Submit Pre-Test Answers to seed proficiency
    print("Submitting pre-test answers...")
    for i, q in enumerate(questions):
        ans_res = requests.post(f"{BASE_URL}/exams/submit-answer", headers=headers, json={
            "session_id": session_id,
            "question_id": q["question_id"],
            "user_answer": "C", # dummy answer
            "time_spent_seconds": 1
        })
        if ans_res.status_code != 200:
            print(f"Failed to submit answer for question {i}", ans_res.json())
            sys.exit(1)
            
        data = ans_res.json()
        if data.get("completed"):
            print("Pre-test completed! Final Score:", data.get("score"))
            print("Seeded domain scores in response:", data.get("domain_scores"))
            break

    # 6. Verify simulation status AFTER Pre-test is complete (should succeed)
    print("Verifying simulation status after pre-test...")
    status_res = requests.get(f"{BASE_URL}/simulation/status", headers=headers)
    print(f"Status Res Code: {status_res.status_code}, Text: {status_res.text}")
    if status_res.status_code != 200:
        print("Failed to get simulation status", status_res.text)
        sys.exit(1)
    status_data = status_res.json()
    print("Simulation status:", status_data)
    if not status_data.get("needs_training"):
        print("Error: User should need training because dummy pre-test score is low.")
        sys.exit(1)

    # 7. Start Simulation Scenario
    print("Starting simulation scenario...")
    start_res = requests.post(f"{BASE_URL}/simulation/start", {}, headers=headers)
    if start_res.status_code not in (200, 201):
        print("Failed to start simulation scenario", start_res.json())
        sys.exit(1)
    
    start_data = start_res.json()
    sim_session = start_data["session"]
    scenario = start_data["scenario"]
    print(f"Simulation started. Session ID: {sim_session['session_id']}. Scenario Title: '{scenario['title']}'")
    print(f"Scenario domain: {scenario['domain_id']}. Initial logs count: {len(scenario['initial_logs'])}")
    
    # 8. Test Log Query
    print("Logging query action...")
    query_res = requests.post(f"{BASE_URL}/simulation/log-query", headers=headers, json={
        "session_id": sim_session['session_id'],
        "query": 'source == "firewall"'
    })
    if query_res.status_code != 200:
        print("Failed to log query", query_res.json())
        sys.exit(1)
    print("Query logged successfully:", query_res.json())

    # 9. Test Session Recovery
    print("Testing Simulation Session Recovery...")
    rec_res = requests.get(f"{BASE_URL}/simulation/session", headers=headers)
    if rec_res.status_code != 200:
        print("Simulation session recovery failed", rec_res.json())
        sys.exit(1)
    rec_data = rec_res.json()
    print(f"Recovery successful. Type: {rec_data['type']}. Recovered session ID: {rec_data['session']['session_id']}")

    # 10. Submit Scenario Solution
    print("Submitting scenario solution...")
    # Gather playbook action IDs to select
    selected_actions = []
    # Let's select one action from containment just to pass validation
    playbook = scenario.get("playbook_steps", {})
    containment = playbook.get("containment", [])
    if containment:
        selected_actions.append(containment[0]["id"])
        
    submit_res = requests.post(f"{BASE_URL}/simulation/submit", headers=headers, json={
        "session_id": sim_session['session_id'],
        "is_true_positive": True,
        "tp_fp_selected": True,
        "selected_actions": selected_actions,
        "discovered_findings": []
      })
    if submit_res.status_code != 200:
        print("Failed to submit scenario solution", submit_res.json())
        sys.exit(1)
    
    submit_data = submit_res.json()
    result = submit_data["result"]
    print(f"Scenario submitted successfully! Total score: {result['total_score']}%. Triage Correct: {result['tp_fp_correct']}")

    # 11. Retrieve scenario result review
    print("Retrieving scenario result review...")
    result_res = requests.get(f"{BASE_URL}/simulation/result/{sim_session['session_id']}", headers=headers)
    if result_res.status_code != 200:
        print("Failed to retrieve scenario result", result_res.json())
        sys.exit(1)
    
    result_data = result_res.json()
    print("Successfully retrieved scenario result dashboard.")
    print("Scenario status:", result_data["session"]["status"])
    print("Domain proficiencies updated count:", len(result_data["domain_proficiencies"]))

    print("\n--- ALL TESTS COMPLETED SUCCESSFULLY ---")

if __name__ == "__main__":
    run_tests()
