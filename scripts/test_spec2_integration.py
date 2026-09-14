import requests
import random
import sys

BASE_URL = "http://localhost:8080/api"

def run_spec2_tests():
    print("=== Running Spec 2 Integration Tests ===")
    
    rand_id = random.randint(10000, 99999)
    learner_email = f"learner_{rand_id}@soc.local"
    learner_nick = f"analyst_{rand_id}"
    spec_email = f"specialist_{rand_id}@soc.local"
    spec_nick = f"specialist_{rand_id}"
    other_email = f"other_{rand_id}@soc.local"
    other_nick = f"other_{rand_id}"
    password = "SecretPassword123!"

    # 1. Register users
    print(f"\n[Setup] Registering learner ({learner_nick}) and specialist candidate ({spec_nick})...")
    requests.post(f"{BASE_URL}/register", json={"email": learner_email, "password": password, "nickname": learner_nick})
    requests.post(f"{BASE_URL}/register", json={"email": spec_email, "password": password, "nickname": spec_nick})
    requests.post(f"{BASE_URL}/register", json={"email": other_email, "password": password, "nickname": other_nick})

    # Log in as learner
    res = requests.post(f"{BASE_URL}/login", json={"email": learner_email, "password": password})
    learner_token = res.json().get("token")

    # Log in as specialist candidate
    res = requests.post(f"{BASE_URL}/login", json={"email": spec_email, "password": password})
    spec_token = res.json().get("token")

    # Log in as other user
    res = requests.post(f"{BASE_URL}/login", json={"email": other_email, "password": password})
    other_token = res.json().get("token")

    # Onboard specialist via apply + admin approve
    # Submit application
    app_res = requests.post(
        f"{BASE_URL}/specialist/apply",
        headers={"Authorization": f"Bearer {spec_token}"},
        data={"bio": "Senior SOC Analyst", "linkedin_url": "https://linkedin.com/in/sec"},
        files={"resume": ("resume.pdf", b"%PDF-1.4 dummy", "application/pdf")}
    )
    app_id = app_res.json().get("application_id")

    # Login as admin to approve candidate
    admin_login = requests.post(f"{BASE_URL}/login", json={"email": "admin@soc.local", "password": "password"})
    if admin_login.status_code == 200:
        admin_token = admin_login.json().get("token")
        requests.post(f"{BASE_URL}/admin/applications/{app_id}/approve", headers={"Authorization": f"Bearer {admin_token}"})
        # Refresh specialist token with new role
        spec_login = requests.post(f"{BASE_URL}/login", json={"email": spec_email, "password": password})
        spec_token = spec_login.json().get("token")
        print("Specialist candidate approved and role escalated to 'specialist'")
    else:
        print("Warning: default admin@soc.local not found, using admin_token if available")

    # -------------------------------------------------------------
    # Test 1: RBAC on Specialist Review Console (/api/specialist/submissions)
    # -------------------------------------------------------------
    print("\n[Test 1] Testing RBAC on /api/specialist/submissions (Learner forbidden)...")
    res = requests.get(f"{BASE_URL}/specialist/submissions", headers={"Authorization": f"Bearer {learner_token}"})
    if res.status_code != 403:
        print(f"FAILED: Expected 403 for learner access to reviews, got {res.status_code}")
        sys.exit(1)
    print("PASSED: Learner denied access with 403 Forbidden")

    print("\n[Test 2] Testing Specialist access to /api/specialist/submissions...")
    res = requests.get(f"{BASE_URL}/specialist/submissions", headers={"Authorization": f"Bearer {spec_token}"})
    if res.status_code != 200:
        print(f"FAILED: Expected 200 for specialist access to reviews, got {res.status_code}: {res.text}")
        sys.exit(1)
    submissions = res.json()
    print(f"PASSED: Specialist retrieved submissions list ({len(submissions)} items)")

    # -------------------------------------------------------------
    # Test 3: Anonymization check (Strictly Nickname, no Email or UserID)
    # -------------------------------------------------------------
    print("\n[Test 3] Checking privacy anonymization on returned submissions...")
    for sub in submissions:
        if "email" in sub or "user_id" in sub:
            print(f"FAILED: Privacy violation! Submission contains leaked identifier: {sub}")
            sys.exit(1)
        if "learner_nickname" not in sub:
            print(f"FAILED: Missing learner_nickname: {sub}")
            sys.exit(1)
    print("PASSED: All submission records are properly anonymized with learner nicknames only")

    # -------------------------------------------------------------
    # Test 4: RBAC on Authoring Analysis Threads (POST /api/threads)
    # -------------------------------------------------------------
    print("\n[Test 4] Testing RBAC on POST /api/threads (Learner forbidden)...")
    res = requests.post(f"{BASE_URL}/threads", headers={"Authorization": f"Bearer {learner_token}"}, json={
        "title": "Unauthorized Learner Attempt at Writing Thread",
        "content": "This post should be rejected because learners cannot author specialist threads.",
        "tags": ["test"]
    })
    if res.status_code != 403:
        print(f"FAILED: Expected 403 for learner thread creation, got {res.status_code}")
        sys.exit(1)
    print("PASSED: Learner blocked from thread authoring with 403 Forbidden")

    # -------------------------------------------------------------
    # Test 5: Specialist Authoring Analysis Thread (POST /api/threads)
    # -------------------------------------------------------------
    print("\n[Test 5] Specialist authoring a new analysis thread...")
    thread_payload = {
        "title": f"Tactical SIEM Investigation Walkthrough #{rand_id}",
        "content": "### Blue Team Investigation Walkthrough\n\nDuring triage, the primary IOC was an obfuscated PowerShell encoded command.\n\n| Step | Action | Finding |\n|---|---|---|\n| 1 | Alert Triage | Suspicious command line |\n| 2 | Containment | VLAN isolation |",
        "tags": ["triage", "powershell", "threat-intel"]
    }
    res = requests.post(f"{BASE_URL}/threads", headers={"Authorization": f"Bearer {spec_token}"}, json=thread_payload)
    if res.status_code != 201:
        print(f"FAILED: Expected 201, got {res.status_code}: {res.text}")
        sys.exit(1)
    thread = res.json()
    thread_id = thread.get("thread_id")
    if not thread_id or thread.get("author", {}).get("nickname") != spec_nick:
        print(f"FAILED: Invalid thread response payload: {thread}")
        sys.exit(1)
    print(f"PASSED: Thread created successfully (ID: {thread_id}, Author: {spec_nick})")

    # -------------------------------------------------------------
    # Test 6: Public Feed Listing & Tag Filtering (GET /api/threads)
    # -------------------------------------------------------------
    print("\n[Test 6] Listing public threads and filtering by tag...")
    list_res = requests.get(f"{BASE_URL}/threads?tag=powershell")
    if list_res.status_code != 200:
        print(f"FAILED: Expected 200, got {list_res.status_code}")
        sys.exit(1)
    items = list_res.json().get("threads", [])
    if not any(t["thread_id"] == thread_id for t in items):
        print(f"FAILED: Created thread not found in filtered list")
        sys.exit(1)
    print("PASSED: Public feed retrieved and filtered by tag successfully")

    # -------------------------------------------------------------
    # Test 7: Fetch Single Thread & View Count Increment (GET /api/threads/:id)
    # -------------------------------------------------------------
    print("\n[Test 7] Fetching single thread and checking view count...")
    initial_views = thread.get("view_count", 0)
    detail_res = requests.get(f"{BASE_URL}/threads/{thread_id}")
    if detail_res.status_code != 200:
        print(f"FAILED: Expected 200, got {detail_res.status_code}")
        sys.exit(1)
    new_views = detail_res.json().get("view_count", 0)
    if new_views <= initial_views:
        print(f"FAILED: View count did not increment. Before: {initial_views}, After: {new_views}")
        sys.exit(1)
    print(f"PASSED: Single thread fetched with view count incremented to {new_views}")

    # -------------------------------------------------------------
    # Test 8: Unauthorized Edit & Delete Protection
    # -------------------------------------------------------------
    print("\n[Test 8] Testing edit/delete protection by non-author...")
    bad_edit = requests.put(
        f"{BASE_URL}/threads/{thread_id}",
        headers={"Authorization": f"Bearer {other_token}"},
        json={"title": "Hacked Title", "content": "Unauthorized content alteration attempts here."}
    )
    if bad_edit.status_code != 403:
        print(f"FAILED: Expected 403 for non-author edit, got {bad_edit.status_code}")
        sys.exit(1)

    bad_del = requests.delete(
        f"{BASE_URL}/threads/{thread_id}",
        headers={"Authorization": f"Bearer {other_token}"}
    )
    if bad_del.status_code != 403:
        print(f"FAILED: Expected 403 for non-author delete, got {bad_del.status_code}")
        sys.exit(1)
    print("PASSED: Non-author edit and delete safely rejected with 403 Forbidden")

    # -------------------------------------------------------------
    # Test 9: Author Edit (PUT /api/threads/:id)
    # -------------------------------------------------------------
    print("\n[Test 9] Author editing their own analysis thread...")
    update_res = requests.put(
        f"{BASE_URL}/threads/{thread_id}",
        headers={"Authorization": f"Bearer {spec_token}"},
        json={
            "title": f"Updated Tactical Investigation #{rand_id}",
            "content": "### Updated Analysis Content with Mitigations\n\nDeploy custom Sigma detection rules for cmdline flags.",
            "tags": ["triage", "sigma"]
        }
    )
    if update_res.status_code != 200:
        print(f"FAILED: Expected 200 for author update, got {update_res.status_code}")
        sys.exit(1)
    if update_res.json().get("title") != f"Updated Tactical Investigation #{rand_id}":
        print(f"FAILED: Title not updated: {update_res.json()}")
        sys.exit(1)
    print("PASSED: Thread updated successfully by author")

    # -------------------------------------------------------------
    # Test 10: Author Delete (DELETE /api/threads/:id)
    # -------------------------------------------------------------
    print("\n[Test 10] Author deleting their analysis thread...")
    del_res = requests.delete(f"{BASE_URL}/threads/{thread_id}", headers={"Authorization": f"Bearer {spec_token}"})
    if del_res.status_code != 200:
        print(f"FAILED: Expected 200 for author delete, got {del_res.status_code}")
        sys.exit(1)
    
    # Confirm 404
    get_res = requests.get(f"{BASE_URL}/threads/{thread_id}")
    if get_res.status_code != 404:
        print(f"FAILED: Expected 404 after deletion, got {get_res.status_code}")
        sys.exit(1)
    print("PASSED: Thread deleted and confirmed 404 Not Found")

    print("\n=======================================================")
    print("  ALL SPEC 2 INTEGRATION TESTS PASSED SUCCESSFULLY! 🎯 ")
    print("=======================================================")

if __name__ == "__main__":
    run_spec2_tests()
