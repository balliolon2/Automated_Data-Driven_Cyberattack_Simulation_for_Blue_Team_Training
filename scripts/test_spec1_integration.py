import requests
import random
import sys
import io

BASE_URL = "http://localhost:8080/api"

def run_spec1_tests():
    print("=== Running Spec 1 Integration Tests ===")
    
    rand_id = random.randint(10000, 99999)
    email = f"learner_{rand_id}@soc.local"
    nickname = f"analyst_{rand_id}"
    password = "SecretPassword123!"

    # 1. Register new learner with nickname
    print(f"\n[Test 1] Registering user with nickname: {nickname}")
    res = requests.post(f"{BASE_URL}/register", json={
        "email": email,
        "password": password,
        "nickname": nickname
    })
    if res.status_code != 201:
        print(f"FAILED: Expected 201, got {res.status_code}: {res.text}")
        sys.exit(1)
    print("PASSED: User registered successfully with nickname")

    # 2. Attempt duplicate nickname registration
    print("\n[Test 2] Testing duplicate nickname rejection...")
    dup_res = requests.post(f"{BASE_URL}/register", json={
        "email": f"other_{rand_id}@soc.local",
        "password": password,
        "nickname": nickname
    })
    if dup_res.status_code != 400:
        print(f"FAILED: Expected 400 for duplicate nickname, got {dup_res.status_code}")
        sys.exit(1)
    print("PASSED: Duplicate nickname rejected with 400 Bad Request")

    # 3. Login
    print("\n[Test 3] Logging in...")
    login_res = requests.post(f"{BASE_URL}/login", json={
        "email": email,
        "password": password
    })
    if login_res.status_code != 200:
        print(f"FAILED: Login failed: {login_res.text}")
        sys.exit(1)
    
    data = login_res.json()
    token = data.get("token")
    user_id = data.get("user_id")
    user_role = data.get("role")
    user_nick = data.get("nickname")

    if not token or user_nick != nickname or user_role != "learner":
        print(f"FAILED: Invalid login response attributes: {data}")
        sys.exit(1)
    print("PASSED: Login returned valid token, nickname, and role 'learner'")

    headers = {"Authorization": f"Bearer {token}"}

    # 4. Get User Profile
    print("\n[Test 4] Fetching user profile...")
    prof_res = requests.get(f"{BASE_URL}/user/profile", headers=headers)
    if prof_res.status_code != 200:
        print(f"FAILED: Profile fetch failed: {prof_res.text}")
        sys.exit(1)
    prof_data = prof_res.json()
    if prof_data.get("nickname") != nickname or prof_data.get("latest_application") is not None:
        print(f"FAILED: Unexpected profile data: {prof_data}")
        sys.exit(1)
    print("PASSED: Profile fetched with no active application")

    # 5. Update Nickname
    updated_nickname = f"nick_upd_{rand_id}"
    print(f"\n[Test 5] Updating nickname to: {updated_nickname}")
    upd_res = requests.put(f"{BASE_URL}/user/nickname", headers=headers, json={"nickname": updated_nickname})
    if upd_res.status_code != 200:
        print(f"FAILED: Nickname update failed: {upd_res.text}")
        sys.exit(1)
    print("PASSED: Nickname updated")

    # 6. Submit Specialist Application with documents
    print("\n[Test 6] Submitting Specialist application with resume...")
    dummy_pdf = io.BytesIO(b"%PDF-1.4 dummy test resume content")
    files = {
        "resume": ("resume.pdf", dummy_pdf, "application/pdf")
    }
    form_data = {
        "bio": "Certified SOC Analyst with 3 years of experience in SIEM triage.",
        "linkedin_url": "https://linkedin.com/in/test-analyst",
        "portfolio_url": "https://credly.com/badges/test"
    }
    app_res = requests.post(f"{BASE_URL}/specialist/apply", headers=headers, data=form_data, files=files)
    if app_res.status_code not in (200, 201):
        print(f"FAILED: Specialist application submission failed: {app_res.text}")
        sys.exit(1)
    app_id = app_res.json().get("application_id")
    print(f"PASSED: Specialist application submitted with ID: {app_id}")

    # 7. Check application status
    print("\n[Test 7] Checking application status...")
    status_res = requests.get(f"{BASE_URL}/specialist/application-status", headers=headers)
    if status_res.status_code != 200:
        print(f"FAILED: Status fetch failed: {status_res.text}")
        sys.exit(1)
    status_data = status_res.json()
    if status_data.get("status") != "pending":
        print(f"FAILED: Expected status 'pending', got: {status_data.get('status')}")
        sys.exit(1)
    print("PASSED: Specialist application status is 'pending'")

    # 8. Learner attempts admin endpoint -> 403 Forbidden
    print("\n[Test 8] Verifying RBAC: Learner accessing /admin/applications...")
    admin_access_res = requests.get(f"{BASE_URL}/admin/applications", headers=headers)
    if admin_access_res.status_code != 403:
        print(f"FAILED: Expected 403 Forbidden for learner, got: {admin_access_res.status_code}")
        sys.exit(1)
    print("PASSED: Learner denied access to admin endpoint (403 Forbidden)")

    print("\n=== ALL SPEC 1 INTEGRATION TESTS PASSED ===")

if __name__ == "__main__":
    run_spec1_tests()
