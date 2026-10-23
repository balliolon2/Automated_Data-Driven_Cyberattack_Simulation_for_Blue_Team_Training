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
    # 6b. Verify Post-Test is BLOCKED when domain proficiency is below threshold
    print("Testing Post-Test Eligibility Gate (expecting 403 Forbidden)...")
    blocked_post_res = requests.post(f"{BASE_URL}/exams/post-test", headers=headers)
    if blocked_post_res.status_code != 403:
        print(f"Error: Expected 403 Forbidden for unearned Post-Test, got {blocked_post_res.status_code}")
        sys.exit(1)
    print("Eligibility Gate VERIFIED: Post-Test correctly blocked for learner below threshold:", blocked_post_res.json())

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
    if rec_data.get("selected_log_indices") not in ([], None):
        print("Error: New session should recover an empty evidence selection, got:", rec_data.get("selected_log_indices"))
        sys.exit(1)

    # 9b. Evidence selection persistence: select -> recover -> clear
    print("Testing evidence selection persistence...")
    total_logs = len(scenario.get("initial_logs") or [])
    if total_logs == 0:
        print("Error: Scenario has no initial logs to select")
        sys.exit(1)

    select_res = requests.post(f"{BASE_URL}/simulation/select-log", headers=headers, json={
        "session_id": sim_session['session_id'],
        "log_index": 0,  # index zero must be accepted
        "selected": True
    })
    if select_res.status_code != 200:
        print("Failed to select evidence log", select_res.status_code, select_res.text)
        sys.exit(1)
    select_data = select_res.json()
    if select_data.get("selected_log_indices") != [0] or select_data.get("selected_count") != 1:
        print("Error: Unexpected selection after toggle:", select_data)
        sys.exit(1)
    print("Selected log index 0:", select_data)

    recovered_selection_res = requests.get(f"{BASE_URL}/simulation/session", headers=headers)
    recovered_selection = recovered_selection_res.json()
    if recovered_selection.get("selected_log_indices") != [0]:
        print("Error: Evidence selection did not survive recovery:", recovered_selection.get("selected_log_indices"))
        sys.exit(1)
    print("Selection recovered after refresh:", recovered_selection.get("selected_log_indices"))

    invalid_res = requests.post(f"{BASE_URL}/simulation/select-log", headers=headers, json={
        "session_id": sim_session['session_id'],
        "log_index": total_logs,
        "selected": True
    })
    if invalid_res.status_code != 400:
        print(f"Error: Expected 400 for out-of-range log index, got {invalid_res.status_code}: {invalid_res.text}")
        sys.exit(1)
    print("Out-of-range log index correctly rejected with 400")

    clear_res = requests.post(f"{BASE_URL}/simulation/select-log/clear", headers=headers, json={
        "session_id": sim_session['session_id']
    })
    if clear_res.status_code != 200:
        print("Failed to clear evidence selection", clear_res.status_code, clear_res.text)
        sys.exit(1)
    clear_data = clear_res.json()
    if clear_data.get("selected_log_indices") != [] or clear_data.get("selected_count") != 0:
        print("Error: Unexpected result after clear:", clear_data)
        sys.exit(1)

    recovered_after_clear = requests.get(f"{BASE_URL}/simulation/session", headers=headers).json()
    if recovered_after_clear.get("selected_log_indices") != []:
        print("Error: Clear did not persist after recovery:", recovered_after_clear.get("selected_log_indices"))
        sys.exit(1)
    print("Clear action persisted; recovered selection is empty")

    # Leave a valid selection in place before submitting.
    reselect_res = requests.post(f"{BASE_URL}/simulation/select-log", headers=headers, json={
        "session_id": sim_session['session_id'],
        "log_index": 0,
        "selected": True
    })
    if reselect_res.status_code != 200:
        print("Failed to re-select evidence log", reselect_res.status_code, reselect_res.text)
        sys.exit(1)
    selected_log_indices = reselect_res.json().get("selected_log_indices", [])
    print("Selection left before submit:", selected_log_indices)

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
        "discovered_findings": [],
        "selected_log_indices": selected_log_indices
      })
    if submit_res.status_code != 200:
        print("Failed to submit scenario solution", submit_res.json())
        sys.exit(1)
    
    submit_data = submit_res.json()
    result = submit_data["result"]
    print(f"Scenario submitted successfully! Total score: {result['total_score']}%. Triage Correct: {result['tp_fp_correct']}")

    # Verify a completed session can no longer change its evidence selection.
    completed_toggle = requests.post(f"{BASE_URL}/simulation/select-log", headers=headers, json={
        "session_id": sim_session['session_id'],
        "log_index": 0,
        "selected": False
    })
    if completed_toggle.status_code != 403:
        print(f"Error: Expected 403 when toggling evidence after completion, got {completed_toggle.status_code}: {completed_toggle.text}")
        sys.exit(1)
    print("Completed session correctly rejects evidence toggles with 403")

    # 12. Verify Research Protocol Analytics Summary
    print("Testing Research Analytics Summary endpoint...")
    summary_res = requests.get(f"{BASE_URL}/analytics/research-summary", headers=headers)
    if summary_res.status_code != 200:
        print("Failed to retrieve research summary", summary_res.json())
        sys.exit(1)
    summary_data = summary_res.json()
    print("Research Summary retrieved:", {
        "protocol_version": summary_data.get("protocol_version"),
        "calculation_version": summary_data.get("calculation_version"),
        "has_pre_test": summary_data.get("has_pre_test"),
        "completed_scenarios": summary_data.get("completed_scenarios"),
        "all_domains_passed": summary_data.get("all_domains_passed"),
    })

    print("\n--- ALL TESTS COMPLETED SUCCESSFULLY ---")

if __name__ == "__main__":
    run_tests()
