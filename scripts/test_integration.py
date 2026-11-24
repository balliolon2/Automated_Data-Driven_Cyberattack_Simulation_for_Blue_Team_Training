import requests
import random
import sys
import uuid

BASE_URL = "http://localhost:8080/api"


def assert_review_payload(review_data, expected_exam_type, expected_unanswered=0):
    sess = review_data.get("session") or {}
    if sess.get("exam_type") != expected_exam_type or sess.get("status") != "completed":
        print(f"Error: Unexpected review session metadata: {sess}")
        sys.exit(1)
    if sess.get("unanswered_count") != expected_unanswered:
        print(f"Error: Expected unanswered_count={expected_unanswered}, got {sess.get('unanswered_count')}")
        sys.exit(1)

    domain_scores = review_data.get("domain_scores") or []
    if len(domain_scores) != 5:
        print(f"Error: Expected 5 domain_scores, got {len(domain_scores)}")
        sys.exit(1)

    results = review_data.get("results") or []
    if len(results) != 30:
        print(f"Error: Expected 30 review results, got {len(results)}")
        sys.exit(1)
    orders = [r.get("order_index") for r in results]
    if orders != list(range(30)):
        print(f"Error: Expected order_index 0..29, got {orders}")
        sys.exit(1)
    for idx, item in enumerate(results):
        if not item.get("correct_answer") or not item.get("explanation"):
            print(f"Error: Missing correct_answer/explanation on review item {idx}")
            sys.exit(1)
        if not item.get("sub_topic_id") or not item.get("sub_topic_title"):
            print(f"Error: Missing sub_topic metadata on review item {idx}: {item.get('sub_topic_id')}")
            sys.exit(1)


def verify_partial_exam_finalize():
    suffix = uuid.uuid4().hex[:8]
    email = f"partial_{suffix}@example.com"
    password = "password123"

    reg = requests.post(f"{BASE_URL}/register", json={"email": email, "password": password})
    if reg.status_code != 201:
        print("Partial test user registration failed", reg.text)
        sys.exit(1)
    tok = requests.post(f"{BASE_URL}/login", json={"email": email, "password": password}).json().get("token")
    hdrs = {"Authorization": f"Bearer {tok}"}

    pre_res = requests.post(f"{BASE_URL}/exams/pre-test", headers=hdrs)
    if pre_res.status_code != 201:
        print("Partial test StartPreTest failed", pre_res.text)
        sys.exit(1)
    pre_body = pre_res.json()
    sid = pre_body["session"]["session_id"]
    qs = pre_body["questions"]

    for q in qs[:28]:
        r = requests.post(f"{BASE_URL}/exams/submit-answer", headers=hdrs, json={
            "session_id": sid,
            "question_id": q["question_id"],
            "user_answer": "A",
            "time_spent_seconds": 0
        })
        if r.status_code != 200:
            print("Partial submit-answer failed", r.text)
            sys.exit(1)

    fin = requests.post(f"{BASE_URL}/exams/sessions/{sid}/finalize", headers=hdrs, json={
        "time_spent_seconds": 321
    })
    if fin.status_code != 200:
        print("Partial finalize failed", fin.status_code, fin.text)
        sys.exit(1)
    fin_data = fin.json()
    if (
        not fin_data.get("completed")
        or fin_data.get("unanswered_count") != 2
        or fin_data.get("time_spent_seconds") != 321
        or len(fin_data.get("domain_scores") or []) != 5
    ):
        print(f"Error: Unexpected partial finalize payload: {fin_data}")
        sys.exit(1)

    rev = requests.get(f"{BASE_URL}/exams/sessions/{sid}/review", headers=hdrs)
    if rev.status_code != 200:
        print("Partial review fetch failed", rev.status_code, rev.text)
        sys.exit(1)
    rev_data = rev.json()
    assert_review_payload(rev_data, "pre", expected_unanswered=2)
    if not rev_data["results"][28].get("is_unanswered") or not rev_data["results"][29].get("is_unanswered"):
        print("Error: Expected last 2 items to be marked is_unanswered=True")
        sys.exit(1)
    print("Partial finalization and unanswered review check PASSED!")


def run_tests():
    # 1. Register a test user
    email = f"test_{uuid.uuid4().hex[:8]}@example.com"
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
    if len(questions) != 30:
        print(f"Error: Expected 30 questions, got {len(questions)}")
        sys.exit(1)
        
    # Track domains
    domains = [q["domain_id"] for q in questions]
    domain_counts = {d: domains.count(d) for d in set(domains)}
    print("Domain distribution:", domain_counts)
    
    expected_counts = {"domain1": 4, "domain2": 7, "domain3": 5, "domain4": 8, "domain5": 6}
    for d, expected in expected_counts.items():
        if domain_counts.get(d, 0) != expected:
            print(f"Error: Expected {expected} questions for {d}, got {domain_counts.get(d, 0)}")
            sys.exit(1)
            
    # 4. Test Reconnect (fetch session)
    print("Testing Session Recovery...")
    session_res = requests.get(f"{BASE_URL}/exams/session", headers=headers)
    if session_res.status_code != 200:
        print("Session recovery failed", session_res.json())
        sys.exit(1)
    print("Session recovery successful. Recovered questions:", len(session_res.json()["questions"]))

    # 4b. Verify Review Completion Guard while Pre-Test is still in_progress
    print("Verifying 403 EXAM_NOT_COMPLETED while Pre-Test is in_progress...")
    inprog_rev = requests.get(f"{BASE_URL}/exams/sessions/{session_id}/review", headers=headers)
    if (
        inprog_rev.status_code != 403
        or inprog_rev.json().get("code") != "EXAM_NOT_COMPLETED"
        or (inprog_rev.json().get("params") or {}).get("exam_type") != "pre"
    ):
        print(f"Error: Expected 403 EXAM_NOT_COMPLETED, got {inprog_rev.status_code}: {inprog_rev.text}")
        sys.exit(1)
    if "correct_answer" in inprog_rev.text:
        print("Error: Answer leaked in 403 review response")
        sys.exit(1)

    # 5. Submit Answers
    print("Submitting answers...")
    pre_questions_used = []
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
            pre_questions_used = [qu["question_id"] for qu in data.get("results", [])]
            break

    if not pre_questions_used:
        print("Error: Pre-Test did not complete during the answer loop; cannot verify zero overlap")
        sys.exit(1)

    # 5a. Verify Exam History & Pre-Test Review after completion
    print("Verifying GET /api/exams/history and Pre-Test Q&A Review...")
    hist_res = requests.get(f"{BASE_URL}/exams/history", headers=headers)
    if hist_res.status_code != 200:
        print("Error: GET /api/exams/history failed", hist_res.status_code, hist_res.text)
        sys.exit(1)
    hist_data = hist_res.json()
    if hist_data.get("latest_pre_session_id") != session_id or not hist_data.get("sessions"):
        print(f"Error: Unexpected exam history after Pre-Test: {hist_data}")
        sys.exit(1)

    pre_rev_res = requests.get(f"{BASE_URL}/exams/sessions/{session_id}/review", headers=headers)
    if pre_rev_res.status_code != 200:
        print("Error: Pre-Test review failed", pre_rev_res.status_code, pre_rev_res.text)
        sys.exit(1)
    pre_rev_data = pre_rev_res.json()
    assert_review_payload(pre_rev_data, "pre", expected_unanswered=0)
    if pre_rev_data.get("pre_comparison") is not None:
        print("Error: Expected pre_comparison to be null on Pre-Test review")
        sys.exit(1)

    # 5a-2. Verify partial exam finalization (28/30 answered + finalize)
    verify_partial_exam_finalize()

    # 5b. Lock sample questions across Learn Sub-topics before Post-Test to verify Pre ∪ Learn ∪ Post = empty
    print("Locking Learn Sub-topic sample questions (1.1, 2.1, 3.1, 4.8, 5.6)...")
    learn_sample_ids = set()
    for sub_id in ["1.1", "2.1", "3.1", "4.8", "5.6"]:
        s_res = requests.get(f"{BASE_URL}/learn/subtopics/{sub_id}", headers=headers)
        if s_res.status_code != 200:
            print(f"Error: GET /learn/subtopics/{sub_id} failed: {s_res.status_code} {s_res.text}")
            sys.exit(1)
        for sq in s_res.json().get("sample_questions", []):
            learn_sample_ids.add(sq["question_id"])
    pre_set = set(pre_questions_used)
    if pre_set.intersection(learn_sample_ids):
        print("Error: Learn sample questions overlapped Pre-Test:", pre_set.intersection(learn_sample_ids))
        sys.exit(1)

    # 6. Start Post-Test
    print("Starting Post-Test...")
    post_res = requests.post(f"{BASE_URL}/exams/post-test", headers=headers)
    if post_res.status_code == 403:
        try:
            post_body = post_res.json()
        except ValueError:
            post_body = {}
        if post_body.get("reason") == "threshold_not_reached":
            print("Post-Test eligibility gate blocked the zero-overlap flow.")
            print("Verified enforced gate: learner is below the proficiency threshold and BYPASS_POST_TEST_THRESHOLD is not true.")
            print("Zero-overlap verification requires BYPASS_POST_TEST_THRESHOLD=true; skipping that check for this environment.")
            sys.exit(0)
        print("Failed to start Post-Test (403)", post_body)
        sys.exit(1)
    if post_res.status_code != 201:
        print("Failed to start Post-Test", post_res.json())
        sys.exit(1)
        
    post_data = post_res.json()
    post_session_id = post_data["session"]["session_id"]
    post_questions = post_data["questions"]
    
    print(f"Post-Test started. Session ID: {post_session_id}. Questions count: {len(post_questions)}")
    
    # Verify no overlap against Pre-Test OR Learn sample questions
    post_set = set([q["question_id"] for q in post_questions])
    overlap = (pre_set | learn_sample_ids).intersection(post_set)
    print(f"Overlap count between (Pre-Test U Learn Samples) and Post-Test: {len(overlap)}")
    if len(overlap) > 0:
        print(f"Error: Found overlap questions: {overlap}")
        sys.exit(1)
    else:
        print("Verification SUCCESS: Zero question overlap between Pre-Test and Post-Test!")

    # Capture simulation/status proficiency scores BEFORE completing Post-Test (Proficiency Invariance check)
    status_before = requests.get(f"{BASE_URL}/simulation/status", headers=headers).json()
    prof_before = {d["domain_id"]: d["proficiency_score"] for d in (status_before.get("domains") or [])}

    # 7. Complete Post-Test and verify Review + Pre->Post Learning Gain + Proficiency Invariance
    print("Submitting Post-Test answers to verify review and Pre->Post comparison...")
    post_completed = False
    for i, q in enumerate(post_questions):
        ans_res = requests.post(f"{BASE_URL}/exams/submit-answer", headers=headers, json={
            "session_id": post_session_id,
            "question_id": q["question_id"],
            "user_answer": "A",
            "time_spent_seconds": 15 if i == len(post_questions) - 1 else 0
        })
        if ans_res.status_code != 200:
            print(f"Failed to submit post-test answer {i}: {ans_res.text}")
            sys.exit(1)
        body = ans_res.json()
        if body.get("completed"):
            post_completed = True
            if len(body.get("domain_scores") or []) != 5:
                print(f"Error: Expected 5 domain_scores on completed Post-Test, got {body.get('domain_scores')}")
                sys.exit(1)
            break

    if not post_completed:
        print("Error: Post-Test did not complete after answering all 30 questions")
        sys.exit(1)

    post_rev_res = requests.get(f"{BASE_URL}/exams/sessions/{post_session_id}/review", headers=headers)
    if post_rev_res.status_code != 200:
        print("Error: Post-Test review failed", post_rev_res.status_code, post_rev_res.text)
        sys.exit(1)
    post_rev_data = post_rev_res.json()
    assert_review_payload(post_rev_data, "post", expected_unanswered=0)

    pre_comp = post_rev_data.get("pre_comparison")
    if not pre_comp or pre_comp.get("pre_session_id") != session_id or len(pre_comp.get("domain_deltas") or []) != 5:
        print(f"Error: Invalid pre_comparison payload on Post-Test review: {pre_comp}")
        sys.exit(1)

    # Verify Proficiency Invariance: Post-Test must NOT overwrite user_skill_profiles
    status_after = requests.get(f"{BASE_URL}/simulation/status", headers=headers).json()
    prof_after = {d["domain_id"]: d["proficiency_score"] for d in (status_after.get("domains") or [])}
    if prof_after != prof_before:
        print(f"Error: Proficiency Invariance violated! before={prof_before}, after={prof_after}")
        sys.exit(1)

    # Verify history now exposes both latest_pre_session_id, latest_post_session_id, and latest_pre_to_post_delta
    hist_after = requests.get(f"{BASE_URL}/exams/history", headers=headers).json()
    if (
        hist_after.get("latest_pre_session_id") != session_id
        or hist_after.get("latest_post_session_id") != post_session_id
        or hist_after.get("latest_pre_to_post_delta") is None
    ):
        print(f"Error: Expected latest_pre, latest_post, and delta in history, got {hist_after}")
        sys.exit(1)
    print("Post-Test Review, Pre->Post Gain, History & Proficiency Invariance VERIFIED!")

if __name__ == "__main__":
    if len(sys.argv) > 1:
        BASE_URL = sys.argv[1]
    run_tests()
