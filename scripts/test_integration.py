import os
import requests
import random
import sys
import uuid

BASE_URL = "http://localhost:8080/api"


def has_thai_chars(text):
    return any("\u0E00" <= ch <= "\u0E7F" for ch in (text or ""))


def assert_bilingual_question_shape(item, label, expect_explanation):
    th = (item.get("translations") or {}).get("th")
    if not isinstance(th, dict):
        print(f"Error: [{label}] missing translations.th on question {item.get('question_id')}: {item}")
        sys.exit(1)
    q_text = th.get("question_text") or ""
    if not has_thai_chars(q_text):
        print(f"Error: [{label}] translations.th.question_text missing Thai text on {item.get('question_id')}")
        sys.exit(1)
    th_opts = th.get("options") or {}
    if set(th_opts.keys()) != {"A", "B", "C", "D"} or any(not str(v).strip() for v in th_opts.values()):
        print(f"Error: [{label}] invalid translations.th.options on {item.get('question_id')}: {th_opts}")
        sys.exit(1)

    if expect_explanation:
        exp = th.get("explanation") or ""
        if not has_thai_chars(exp) or "คำตอบที่ถูกต้อง:" not in exp or "ตัวเลือก" not in exp:
            print(f"Error: [{label}] translations.th.explanation missing structured Thai explanation on {item.get('question_id')}")
            sys.exit(1)
    else:
        if "explanation" in th or "correct_answer" in th or "explanation" in item or "correct_answer" in item:
            print(f"Error: [{label}] pre-submission answer/explanation leakage on {item.get('question_id')}!")
            sys.exit(1)


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
        assert_bilingual_question_shape(item, f"{expected_exam_type} review #{idx}", expect_explanation=True)


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

    if fin_data.get("results"):
        for r_idx, r_item in enumerate(fin_data["results"]):
            assert_bilingual_question_shape(r_item, f"Partial finalize result #{r_idx}", expect_explanation=True)

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


def resolve_admin_headers():
    admin_email = os.environ.get("ADMIN_EMAIL", "admin@soc.local")
    admin_password = os.environ.get("ADMIN_PASSWORD", "password")
    if not os.environ.get("ADMIN_PASSWORD") and os.path.exists(".env"):
        with open(".env", "r", encoding="utf-8") as envf:
            for line in envf:
                line = line.strip()
                if line.startswith("ADMIN_EMAIL=") and not os.environ.get("ADMIN_EMAIL"):
                    admin_email = line.split("=", 1)[1].strip()
                elif line.startswith("ADMIN_PASSWORD="):
                    val = line.split("=", 1)[1].strip()
                    if val:
                        admin_password = val

    admin_login = requests.post(f"{BASE_URL}/login", json={"email": admin_email, "password": admin_password})
    if admin_login.status_code != 200:
        print(f"Error: Admin login failed for {admin_email}: {admin_login.status_code} {admin_login.text}")
        sys.exit(1)
    return {"Authorization": f"Bearer {admin_login.json()['token']}"}


def verify_nfr_portal(learner_headers):
    print("Verifying /api/health DB ping & Admin-only NFR endpoints...")
    health_res = requests.get(f"{BASE_URL}/health")
    if health_res.status_code != 200:
        print(f"Error: GET /health failed: {health_res.status_code} {health_res.text}")
        sys.exit(1)
    h_data = health_res.json()
    if (
        h_data.get("status") != "healthy"
        or h_data.get("db_status") != "up"
        or not isinstance(h_data.get("db_latency_ms"), (int, float))
        or h_data.get("db_latency_ms", 0) <= 0
    ):
        print(f"Error: Unexpected /health payload: {h_data}")
        sys.exit(1)

    if requests.get(f"{BASE_URL}/admin/nfr/live").status_code != 401:
        print("Error: Expected 401 for unauthenticated /admin/nfr/live")
        sys.exit(1)

    for method, ep in [
        ("GET", "/admin/nfr/live"),
        ("GET", "/admin/nfr/runs"),
        ("POST", "/admin/nfr/runs"),
        ("POST", "/admin/nfr/probe/scalability"),
        ("POST", "/admin/nfr/telemetry/reset"),
        ("GET", "/admin/nfr/export?format=csv"),
    ]:
        r_forb = requests.request(method, f"{BASE_URL}{ep}", headers=learner_headers)
        if r_forb.status_code != 403:
            print(f"Error: Expected 403 for learner on {method} {ep}, got {r_forb.status_code}")
            sys.exit(1)

    admin_hdrs = resolve_admin_headers()

    nfr_live = requests.get(f"{BASE_URL}/admin/nfr/live", headers=admin_hdrs)
    if nfr_live.status_code != 200:
        print(f"Error: GET /admin/nfr/live failed: {nfr_live.status_code} {nfr_live.text}")
        sys.exit(1)
    nfr_data = nfr_live.json()
    dims = nfr_data.get("dimensions") or []
    if len(dims) != 10:
        print(f"Error: Expected 10 NFR dimensions, got {len(dims)}")
        sys.exit(1)
    dim_map = {d["id"]: d for d in dims}
    for key in ["data_integrity", "safety", "security"]:
        if dim_map.get(key, {}).get("status") == "fail":
            print(f"Error: Live NFR dimension {key} failed unexpectedly: {dim_map[key]}")
            sys.exit(1)
    safety_dim = dim_map.get("safety") or {}
    safety_checks = {c.get("key"): c for c in (safety_dim.get("checks") or [])}
    en_check_target = (safety_checks.get("english_only_invariant") or {}).get("target") or ""
    if safety_dim.get("status") != "pass" or "canonical EN Exams" not in en_check_target:
        print(f"Error: Expected NFR safety PASS with updated bilingual-safe EN target, got: {safety_dim}")
        sys.exit(1)

    for env_key in ["JWT_SECRET", "LLM_API_KEY", "POSTGRES_PASSWORD"]:
        sec_val = os.environ.get(env_key, "").strip()
        if len(sec_val) >= 12 and sec_val in nfr_live.text:
            print(f"Error: Secret leakage detected for {env_key} in /admin/nfr/live!")
            sys.exit(1)

    probe_res = requests.post(
        f"{BASE_URL}/admin/nfr/probe/scalability",
        headers=admin_hdrs,
        json={"preset": "light", "save_run": True, "run_label": "Integration Probe"},
    )
    if probe_res.status_code != 200:
        print(f"Error: Safe scalability probe failed: {probe_res.status_code} {probe_res.text}")
        sys.exit(1)
    probe_data = probe_res.json()
    saved_run = probe_data.get("saved_run") or {}
    saved_run_id = saved_run.get("run_id")
    if not saved_run_id or (probe_data.get("benchmark") or {}).get("levels", [{}])[0].get("total_requests") != 50:
        print(f"Error: Unexpected scalability probe response: {probe_data}")
        sys.exit(1)

    runs_list = requests.get(f"{BASE_URL}/admin/nfr/runs", headers=admin_hdrs)
    if runs_list.status_code != 200 or not any(r.get("run_id") == saved_run_id for r in (runs_list.json().get("runs") or [])):
        print(f"Error: Saved probe run missing from GET /admin/nfr/runs: {runs_list.text}")
        sys.exit(1)

    run_detail = requests.get(f"{BASE_URL}/admin/nfr/runs/{saved_run_id}", headers=admin_hdrs)
    if run_detail.status_code != 200 or not (run_detail.json().get("report") or {}).get("scalability"):
        print(f"Error: GET /admin/nfr/runs/{saved_run_id} failed: {run_detail.status_code}")
        sys.exit(1)

    nfr_csv = requests.get(f"{BASE_URL}/admin/nfr/export?format=csv&run_id={saved_run_id}", headers=admin_hdrs)
    if (
        nfr_csv.status_code != 200
        or not nfr_csv.content.startswith(b"\xef\xbb\xbf")
        or "Integration Probe" not in nfr_csv.text
        or "scalability_level" not in nfr_csv.text
    ):
        print(f"Error: Unexpected NFR CSV export: {nfr_csv.status_code}")
        sys.exit(1)

    nfr_json = requests.get(f"{BASE_URL}/admin/nfr/export?format=json&run_id={saved_run_id}", headers=admin_hdrs)
    if nfr_json.status_code != 200 or nfr_json.json().get("run_id") != saved_run_id:
        print(f"Error: Unexpected NFR JSON export: {nfr_json.status_code}")
        sys.exit(1)

    nfr_del = requests.delete(f"{BASE_URL}/admin/nfr/runs/{saved_run_id}", headers=admin_hdrs)
    if nfr_del.status_code != 200 or requests.get(f"{BASE_URL}/admin/nfr/runs/{saved_run_id}", headers=admin_hdrs).status_code != 404:
        print(f"Error: Expected run delete + 404 verification: {nfr_del.status_code}")
        sys.exit(1)

    nfr_reset = requests.post(f"{BASE_URL}/admin/nfr/telemetry/reset", headers=admin_hdrs)
    if nfr_reset.status_code != 200 or nfr_reset.json().get("reset") is not True:
        print(f"Error: Expected 200 telemetry reset, got {nfr_reset.status_code}: {nfr_reset.text}")
        sys.exit(1)

    print("System-Level NFR Evaluation Portal, Probes, Run History & Export VERIFIED!")


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
    for idx, q in enumerate(questions):
        assert_bilingual_question_shape(q, f"Pre-Test #{idx}", expect_explanation=False)

    # 3b. Verify public sample question endpoint returns complete bilingual translations
    sample_res = requests.get(f"{BASE_URL}/exams/sample")
    if sample_res.status_code != 200:
        print(f"Error: GET /exams/sample failed: {sample_res.status_code}")
        sys.exit(1)
    assert_bilingual_question_shape(sample_res.json(), "Public /exams/sample", expect_explanation=True)
            
    # 4. Test Reconnect (fetch session)
    print("Testing Session Recovery...")
    session_res = requests.get(f"{BASE_URL}/exams/session", headers=headers)
    if session_res.status_code != 200:
        print("Session recovery failed", session_res.json())
        sys.exit(1)
    recovered_qs = session_res.json()["questions"]
    for idx, q in enumerate(recovered_qs):
        assert_bilingual_question_shape(q, f"Session Recovery #{idx}", expect_explanation=False)
    print("Session recovery successful. Recovered questions:", len(recovered_qs))

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
            for r_idx, r_item in enumerate(data.get("results") or []):
                assert_bilingual_question_shape(r_item, f"Pre-test final submit result #{r_idx}", expect_explanation=True)
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

    # 5a-3. Verify questionnaire is locked before Post-Test completion
    print("Verifying questionnaire eligibility is locked prior to Post-Test...")
    q_status_pre = requests.get(f"{BASE_URL}/evaluation/questionnaire/status", headers=headers)
    if q_status_pre.status_code != 200:
        print("Error: GET /evaluation/questionnaire/status failed", q_status_pre.status_code, q_status_pre.text)
        sys.exit(1)
    q_pre_body = q_status_pre.json()
    if q_pre_body.get("eligible") is not False or q_pre_body.get("ineligible_reason") != "post_test_required":
        print(f"Error: Expected ineligible post_test_required before Post-Test, got {q_pre_body}")
        sys.exit(1)

    q_payload = {
        "locale": "th",
        "csuq": {
            "q1": 6, "q2": 7, "q3": 6, "q4": 6, "q5": 7, "q6": 6,
            "q7": None, "q8": 6, "q9": 6, "q10": 6, "q11": 7, "q12": 6,
            "q13": 7, "q14": 6, "q15": 6, "q16": 7,
        },
        "nasa_tlx": {
            "mental_demand": 60,
            "physical_demand": 10,
            "temporal_demand": 45,
            "performance_rating": 85,
            "effort": 55,
            "frustration": 20,
        },
        "qualitative_feedback": {
            "most_impressive": "=SUM(1,1) SIEM logs felt very realistic",
            "areas_to_improve": "More cloud detection drills",
            "additional_comments": "แบบสอบถามสองภาษาชัดเจนมาก",
        },
    }
    q_early_post = requests.post(f"{BASE_URL}/evaluation/questionnaire", headers=headers, json=q_payload)
    if q_early_post.status_code != 403 or q_early_post.json().get("code") != "QUESTIONNAIRE_POST_TEST_REQUIRED":
        print(f"Error: Expected 403 QUESTIONNAIRE_POST_TEST_REQUIRED, got {q_early_post.status_code}: {q_early_post.text}")
        sys.exit(1)

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
            if not sq.get("submitted_answer"):
                assert_bilingual_question_shape(sq, f"Learn {sub_id} unattempted sample", expect_explanation=False)
    first_sub_samples = requests.get(f"{BASE_URL}/learn/subtopics/1.1", headers=headers).json().get("sample_questions", [])
    if first_sub_samples and not first_sub_samples[0].get("submitted_answer"):
        sq0 = first_sub_samples[0]
        att_res = requests.post(
            f"{BASE_URL}/learn/subtopics/1.1/sample-questions/{sq0['question_id']}/attempt",
            headers=headers,
            json={"answer": "A"},
        )
        if att_res.status_code != 200:
            print(f"Error: Learn sample attempt failed: {att_res.status_code} {att_res.text}")
            sys.exit(1)
        assert_bilingual_question_shape(att_res.json(), "Learn 1.1 attempted sample", expect_explanation=True)
        post_sub_samples = requests.get(f"{BASE_URL}/learn/subtopics/1.1", headers=headers).json().get("sample_questions", [])
        for s_item in post_sub_samples:
            assert_bilingual_question_shape(
                s_item,
                f"Learn 1.1 post-attempt sample {s_item.get('question_id')}",
                expect_explanation=bool(s_item.get("submitted_answer")),
            )
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
            verify_nfr_portal(headers)
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
    for idx, q in enumerate(post_questions):
        assert_bilingual_question_shape(q, f"Post-Test #{idx}", expect_explanation=False)
    
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

    # 8. Verify Post-Test Evaluation Questionnaire lifecycle, Research Summary & Admin endpoints
    print("Verifying Post-Test Evaluation Questionnaire submission & server snapshot...")
    q_status_ready = requests.get(f"{BASE_URL}/evaluation/questionnaire/status", headers=headers)
    if q_status_ready.status_code != 200 or not q_status_ready.json().get("eligible"):
        print(f"Error: Expected questionnaire eligible after Post-Test, got {q_status_ready.status_code}: {q_status_ready.text}")
        sys.exit(1)

    q_submit = requests.post(f"{BASE_URL}/evaluation/questionnaire", headers=headers, json=q_payload)
    if q_submit.status_code != 201:
        print(f"Error: Expected 201 Created submitting questionnaire, got {q_submit.status_code}: {q_submit.text}")
        sys.exit(1)
    sub_data = q_submit.json().get("submission") or {}
    expected_delta = round(sub_data.get("post_score", 0) - sub_data.get("pre_score", 0), 2)
    if (
        sub_data.get("improvement_pct") != expected_delta
        or (sub_data.get("nasa_tlx_scores") or {}).get("dimensions", {}).get("performance_workload") != 15
        or (sub_data.get("csuq_scores") or {}).get("subscales", {}).get("overall") != 6.33
    ):
        print(f"Error: Unexpected questionnaire snapshot/scoring payload: {sub_data}")
        sys.exit(1)

    # Proficiency Invariance check after questionnaire submit
    status_after_q = requests.get(f"{BASE_URL}/simulation/status", headers=headers).json()
    prof_after_q = {d["domain_id"]: d["proficiency_score"] for d in (status_after_q.get("domains") or [])}
    if prof_after_q != prof_before:
        print(f"Error: Questionnaire modified proficiency scores! before={prof_before}, after={prof_after_q}")
        sys.exit(1)

    # Research Summary additive fields
    rs_res = requests.get(f"{BASE_URL}/analytics/research-summary", headers=headers)
    if rs_res.status_code != 200:
        print(f"Error: GET /analytics/research-summary failed: {rs_res.status_code}")
        sys.exit(1)
    rs_data = rs_res.json()
    if (
        rs_data.get("questionnaire_submitted") is not True
        or rs_data.get("csuq_overall") != 6.33
        or rs_data.get("nasa_tlx_overall") != 34.17
    ):
        print(f"Error: Missing additive questionnaire telemetry in research-summary: {rs_data}")
        sys.exit(1)

    # Duplicate submission blocked with 409
    q_dup = requests.post(f"{BASE_URL}/evaluation/questionnaire", headers=headers, json=q_payload)
    if q_dup.status_code != 409 or q_dup.json().get("code") != "QUESTIONNAIRE_ALREADY_SUBMITTED":
        print(f"Error: Expected 409 QUESTIONNAIRE_ALREADY_SUBMITTED, got {q_dup.status_code}: {q_dup.text}")
        sys.exit(1)

    # Learner RBAC check on admin endpoints
    for admin_ep in ["/admin/evaluations", "/admin/evaluations/statistics", "/admin/evaluations/export?format=csv"]:
        forbidden_res = requests.get(f"{BASE_URL}{admin_ep}", headers=headers)
        if forbidden_res.status_code != 403:
            print(f"Error: Expected 403 for learner on {admin_ep}, got {forbidden_res.status_code}")
            sys.exit(1)

    # Admin evaluations, paper statistics, and export checks
    admin_hdrs = resolve_admin_headers()

    adm_eval = requests.get(f"{BASE_URL}/admin/evaluations", headers=admin_hdrs)
    if adm_eval.status_code != 200:
        print(f"Error: GET /admin/evaluations failed: {adm_eval.status_code} {adm_eval.text}")
        sys.exit(1)
    adm_eval_data = adm_eval.json()
    if not adm_eval_data.get("responses") or any("email" in r for r in adm_eval_data["responses"]):
        print(f"Error: Unexpected admin evaluations payload or email PII leak: {adm_eval_data}")
        sys.exit(1)

    adm_stats = requests.get(f"{BASE_URL}/admin/evaluations/statistics", headers=admin_hdrs)
    if adm_stats.status_code != 200 or "NaN" in adm_stats.text or "Infinity" in adm_stats.text:
        print(f"Error: GET /admin/evaluations/statistics invalid: {adm_stats.status_code} {adm_stats.text}")
        sys.exit(1)
    stats_body = adm_stats.json()
    resp_count = (stats_body.get("cohort") or {}).get("questionnaire_respondents", 0)
    if resp_count < 1:
        print(f"Error: Expected >= 1 questionnaire respondent in statistics: {stats_body}")
        sys.exit(1)
    pts = stats_body.get("participant_points") or []
    if len(pts) != resp_count:
        print(f"Error: Expected {resp_count} participant_points, got {len(pts)}: {pts}")
        sys.exit(1)
    for pt in pts:
        code = pt.get("participant_code") or ""
        if not code.startswith("P-") or "email" in pt or "user_id" in pt:
            print(f"Error: Invalid or leaky participant_point entry: {pt}")
            sys.exit(1)

    adm_csv = requests.get(f"{BASE_URL}/admin/evaluations/export?format=csv", headers=admin_hdrs)
    if (
        adm_csv.status_code != 200
        or not adm_csv.content.startswith(b"\xef\xbb\xbf")
        or "participant_code" not in adm_csv.text
        or "'=SUM(1,1)" not in adm_csv.text
    ):
        print(f"Error: Unexpected CSV export response: status={adm_csv.status_code}, prefix={adm_csv.content[:20]}")
        sys.exit(1)

    adm_json = requests.get(f"{BASE_URL}/admin/evaluations/export?format=json", headers=admin_hdrs)
    if adm_json.status_code != 200 or not (adm_json.json().get("responses") or []):
        print(f"Error: Unexpected JSON export response: {adm_json.status_code}")
        sys.exit(1)

    print("Questionnaire, Research Summary & Admin Evaluation/Statistics/Export VERIFIED!")
    verify_nfr_portal(headers)

if __name__ == "__main__":
    if len(sys.argv) > 1:
        BASE_URL = sys.argv[1]
    run_tests()
