import requests
import random
import sys

BASE_URL = "http://localhost:8080/api"

def run_spec3_tests():
    print("=== Running Spec 3 Integration Tests ===")

    rand_id = random.randint(10000, 99999)
    spec_email = f"author_spec_{rand_id}@soc.local"
    spec_nick = f"spec_author_{rand_id}"
    learner1_email = f"learner1_{rand_id}@soc.local"
    learner1_nick = f"learner_one_{rand_id}"
    learner2_email = f"learner2_{rand_id}@soc.local"
    learner2_nick = f"learner_two_{rand_id}"
    password = "SecretPassword123!"

    print(f"\n[Setup] Registering users: {spec_nick} (specialist), {learner1_nick}, {learner2_nick}...")
    requests.post(f"{BASE_URL}/register", json={"email": spec_email, "password": password, "nickname": spec_nick})
    requests.post(f"{BASE_URL}/register", json={"email": learner1_email, "password": password, "nickname": learner1_nick})
    requests.post(f"{BASE_URL}/register", json={"email": learner2_email, "password": password, "nickname": learner2_nick})

    res = requests.post(f"{BASE_URL}/login", json={"email": spec_email, "password": password})
    spec_token = res.json().get("token")

    res = requests.post(f"{BASE_URL}/login", json={"email": learner1_email, "password": password})
    learner1_token = res.json().get("token")

    res = requests.post(f"{BASE_URL}/login", json={"email": learner2_email, "password": password})
    learner2_token = res.json().get("token")

    # Onboard spec_nick to specialist role
    app_res = requests.post(
        f"{BASE_URL}/specialist/apply",
        headers={"Authorization": f"Bearer {spec_token}"},
        data={"bio": "Senior SOC Analyst & Author", "linkedin_url": "https://linkedin.com/in/sec"},
        files={"resume": ("resume.pdf", b"%PDF-1.4 dummy", "application/pdf")}
    )
    app_id = app_res.json().get("application_id")

    admin_login = requests.post(f"{BASE_URL}/login", json={"email": "admin@soc.local", "password": "password"})
    admin_token = None
    if admin_login.status_code == 200:
        admin_token = admin_login.json().get("token")
        requests.post(f"{BASE_URL}/admin/applications/{app_id}/approve", headers={"Authorization": f"Bearer {admin_token}"})
        spec_login = requests.post(f"{BASE_URL}/login", json={"email": spec_email, "password": password})
        spec_token = spec_login.json().get("token")
        print("Specialist candidate approved and role escalated.")

    # Create a test thread
    thread_res = requests.post(
        f"{BASE_URL}/threads",
        headers={"Authorization": f"Bearer {spec_token}"},
        json={
            "title": f"Investigation of CVE-2024-{rand_id} Attack Vector",
            "content": "Comprehensive incident writeup analyzing C2 persistence and PowerShell obfuscation.",
            "tags": ["threat-intel", "persistence"]
        }
    )
    if thread_res.status_code != 201:
        print(f"FAILED to create thread: {thread_res.status_code} - {thread_res.text}")
        sys.exit(1)
    thread = thread_res.json()
    thread_id = thread["thread_id"]
    print(f"Created thread '{thread['title']}' (ID: {thread_id})")

    # -------------------------------------------------------------
    # Test 1: Thread Helpful Upvoting (1-vote-per-user toggle)
    # -------------------------------------------------------------
    print("\n[Test 1] Testing Thread Upvote (Toggle On & Off)...")
    # Upvote once
    up1 = requests.post(f"{BASE_URL}/threads/{thread_id}/upvote", headers={"Authorization": f"Bearer {learner1_token}"})
    if up1.status_code != 200 or not up1.json().get("upvoted") or up1.json().get("upvote_count") != 1:
        print(f"FAILED on thread upvote toggle on: {up1.text}")
        sys.exit(1)
    print("PASSED: Learner 1 upvoted thread (count = 1)")

    # Upvote again (toggle off)
    up2 = requests.post(f"{BASE_URL}/threads/{thread_id}/upvote", headers={"Authorization": f"Bearer {learner1_token}"})
    if up2.status_code != 200 or up2.json().get("upvoted") or up2.json().get("upvote_count") != 0:
        print(f"FAILED on thread upvote toggle off: {up2.text}")
        sys.exit(1)
    print("PASSED: Learner 1 toggled upvote off (count = 0)")

    # Re-upvote for learner 1
    requests.post(f"{BASE_URL}/threads/{thread_id}/upvote", headers={"Authorization": f"Bearer {learner1_token}"})

    # -------------------------------------------------------------
    # Test 2: Two-Level Threaded Commenting & 2-Level Flattening
    # -------------------------------------------------------------
    print("\n[Test 2] Testing Comment Creation & 2-Level Auto-Flattening...")
    # 1. Learner 1 posts top-level comment (Level 1)
    c1_res = requests.post(
        f"{BASE_URL}/threads/{thread_id}/comments",
        headers={"Authorization": f"Bearer {learner1_token}"},
        json={"content": "Great analysis! Could you elaborate on the beaconing interval?"}
    )
    if c1_res.status_code != 201:
        print(f"FAILED to post top-level comment: {c1_res.text}")
        sys.exit(1)
    c1 = c1_res.json()
    c1_id = c1["comment_id"]
    print(f"PASSED: Learner 1 posted top-level comment (ID: {c1_id})")

    # 2. Specialist author replies to Level 1 comment (Level 2)
    c2_res = requests.post(
        f"{BASE_URL}/threads/{thread_id}/comments",
        headers={"Authorization": f"Bearer {spec_token}"},
        json={
            "parent_comment_id": c1_id,
            "content": "@learner_one The jitter was set to 15% with a 60-second sleep cycle."
        }
    )
    if c2_res.status_code != 201:
        print(f"FAILED to post level-2 reply: {c2_res.text}")
        sys.exit(1)
    c2 = c2_res.json()
    c2_id = c2["comment_id"]
    print(f"PASSED: Specialist posted reply to c1 (Level 2, ID: {c2_id})")

    # 3. Learner 2 replies to Level 2 comment (c2) -> backend should flatten to c1!
    c3_res = requests.post(
        f"{BASE_URL}/threads/{thread_id}/comments",
        headers={"Authorization": f"Bearer {learner2_token}"},
        json={
            "parent_comment_id": c2_id,
            "content": "@spec_author Was there any TLS SNI inspection evasion observed?"
        }
    )
    if c3_res.status_code != 201:
        print(f"FAILED to post reply to level-2 comment: {c3_res.text}")
        sys.exit(1)
    c3 = c3_res.json()
    c3_id = c3["comment_id"]
    if c3.get("parent_comment_id") != c1_id:
        print(f"FAILED: 2-level flattening failed! Expected parent_comment_id={c1_id}, got {c3.get('parent_comment_id')}")
        sys.exit(1)
    print("PASSED: Level-2 reply automatically normalized to root comment (c1_id)")

    # -------------------------------------------------------------
    # Test 3: List Comments Hierarchy & Role Badges
    # -------------------------------------------------------------
    print("\n[Test 3] Verifying List Comments hierarchy & author roles...")
    list_res = requests.get(f"{BASE_URL}/threads/{thread_id}/comments", headers={"Authorization": f"Bearer {learner1_token}"})
    if list_res.status_code != 200:
        print(f"FAILED to list comments: {list_res.text}")
        sys.exit(1)
    comments = list_res.json()
    if len(comments) != 1:
        print(f"FAILED: Expected 1 top-level comment, got {len(comments)}")
        sys.exit(1)
    root = comments[0]
    if len(root["replies"]) != 2:
        print(f"FAILED: Expected 2 replies under root comment, got {len(root['replies'])}")
        sys.exit(1)
    print(f"PASSED: Comment tree correctly structured with 1 root and {len(root['replies'])} nested replies")

    # -------------------------------------------------------------
    # Test 4: Comment Helpful Upvoting (1-vote-per-user toggle)
    # -------------------------------------------------------------
    print("\n[Test 4] Testing Comment Upvote (Toggle On & Off)...")
    cup1 = requests.post(f"{BASE_URL}/comments/{c1_id}/upvote", headers={"Authorization": f"Bearer {learner2_token}"})
    if cup1.status_code != 200 or not cup1.json().get("upvoted") or cup1.json().get("upvote_count") != 1:
        print(f"FAILED on comment upvote toggle on: {cup1.text}")
        sys.exit(1)
    print("PASSED: Learner 2 upvoted c1 (count = 1)")

    cup2 = requests.post(f"{BASE_URL}/comments/{c1_id}/upvote", headers={"Authorization": f"Bearer {learner2_token}"})
    if cup2.status_code != 200 or cup2.json().get("upvoted") or cup2.json().get("upvote_count") != 0:
        print(f"FAILED on comment upvote toggle off: {cup2.text}")
        sys.exit(1)
    print("PASSED: Learner 2 toggled upvote off (count = 0)")

    # -------------------------------------------------------------
    # Test 5: Comment Pinning & Cap at 3
    # -------------------------------------------------------------
    print("\n[Test 5] Testing Comment Pinning & Max 3 Pin Limit...")
    # Learner 1 attempts to pin c1 -> should be 403
    p_unauth = requests.post(f"{BASE_URL}/comments/{c1_id}/pin", headers={"Authorization": f"Bearer {learner1_token}"})
    if p_unauth.status_code != 403:
        print(f"FAILED: Expected 403 for non-moderator pin attempt, got {p_unauth.status_code}")
        sys.exit(1)
    print("PASSED: Non-author learner denied pin with 403")

    # Specialist thread author pins c1
    p_auth = requests.post(f"{BASE_URL}/comments/{c1_id}/pin", headers={"Authorization": f"Bearer {spec_token}"})
    if p_auth.status_code != 200 or not p_auth.json().get("is_pinned"):
        print(f"FAILED: Specialist author failed to pin c1: {p_auth.text}")
        sys.exit(1)
    print("PASSED: Specialist thread author successfully pinned c1")

    # Create 3 more top-level comments to test 3-pin limit
    extra_ids = []
    for i in range(3):
        res = requests.post(
            f"{BASE_URL}/threads/{thread_id}/comments",
            headers={"Authorization": f"Bearer {spec_token}"},
            json={"content": f"Extra comment #{i+1} for pin testing"}
        )
        extra_ids.append(res.json()["comment_id"])

    # Pin extra 1 and 2 (total 3 pinned)
    requests.post(f"{BASE_URL}/comments/{extra_ids[0]}/pin", headers={"Authorization": f"Bearer {spec_token}"})
    requests.post(f"{BASE_URL}/comments/{extra_ids[1]}/pin", headers={"Authorization": f"Bearer {spec_token}"})

    # Attempt to pin 4th -> should be rejected with 400
    p_overflow = requests.post(f"{BASE_URL}/comments/{extra_ids[2]}/pin", headers={"Authorization": f"Bearer {spec_token}"})
    if p_overflow.status_code != 400:
        print(f"FAILED: Expected 400 when exceeding 3 pinned comments, got {p_overflow.status_code}")
        sys.exit(1)
    print("PASSED: 4th pin attempt rejected with 400 Bad Request (Max 3 pinned comments enforced)")

    # -------------------------------------------------------------
    # Test 6: Comment Moderation (Hide / Unhide)
    # -------------------------------------------------------------
    print("\n[Test 6] Testing Comment Hiding (Spam filter)...")
    hide_res = requests.post(f"{BASE_URL}/comments/{c3_id}/hide", headers={"Authorization": f"Bearer {spec_token}"})
    if hide_res.status_code != 200 or not hide_res.json().get("is_hidden"):
        print(f"FAILED: Failed to hide comment: {hide_res.text}")
        sys.exit(1)
    print("PASSED: Specialist thread author hid comment c3")

    # Verify regular learner sees masked content
    list_for_learner = requests.get(f"{BASE_URL}/threads/{thread_id}/comments", headers={"Authorization": f"Bearer {learner1_token}"}).json()
    replies = list_for_learner[0]["replies"]
    hidden_reply = [r for r in replies if r["comment_id"] == c3_id][0]
    if "hidden by thread moderator" not in hidden_reply["content"]:
        print(f"FAILED: Expected masked content for hidden comment, got: {hidden_reply['content']}")
        sys.exit(1)
    print("PASSED: Regular learner receives masked spam placeholder")

    # Verify author/admin sees original content
    list_for_author = requests.get(f"{BASE_URL}/threads/{thread_id}/comments", headers={"Authorization": f"Bearer {spec_token}"}).json()
    replies_auth = list_for_author[0]["replies"]
    hidden_reply_auth = [r for r in replies_auth if r["comment_id"] == c3_id][0]
    if "TLS SNI inspection" not in hidden_reply_auth["content"]:
        print(f"FAILED: Expected unmasked content for moderator, got: {hidden_reply_auth['content']}")
        sys.exit(1)
    print("PASSED: Thread author sees original content with is_hidden=True flag")

    # -------------------------------------------------------------
    # Test 7: In-App Notifications Engine & Read Status
    # -------------------------------------------------------------
    print("\n[Test 7] Testing In-App Notification Engine & Read Toggles...")
    # Check notifications for Specialist (should have notification about c1 from learner1)
    notif_res = requests.get(f"{BASE_URL}/notifications", headers={"Authorization": f"Bearer {spec_token}"})
    if notif_res.status_code != 200:
        print(f"FAILED to fetch notifications: {notif_res.text}")
        sys.exit(1)
    notif_data = notif_res.json()
    notifs = notif_data.get("notifications", [])
    unread_count = notif_data.get("unread_count", 0)
    print(f"Specialist has {len(notifs)} notifications ({unread_count} unread)")

    if len(notifs) == 0:
        print("FAILED: Expected at least 1 notification for specialist author on new comment")
        sys.exit(1)

    target_notif = notifs[0]
    target_id = target_notif["notification_id"]

    # Mark single notification as read
    read_res = requests.put(f"{BASE_URL}/notifications/{target_id}/read", headers={"Authorization": f"Bearer {spec_token}"})
    if read_res.status_code != 200:
        print(f"FAILED to mark notification as read: {read_res.text}")
        sys.exit(1)
    print("PASSED: Marked individual notification as read")

    # Mark all notifications as read
    read_all_res = requests.put(f"{BASE_URL}/notifications/read-all", headers={"Authorization": f"Bearer {spec_token}"})
    if read_all_res.status_code != 200:
        print(f"FAILED to mark all notifications as read: {read_all_res.text}")
        sys.exit(1)
    
    # Confirm unread count is 0
    after_res = requests.get(f"{BASE_URL}/notifications", headers={"Authorization": f"Bearer {spec_token}"}).json()
    if after_res.get("unread_count") != 0:
        print(f"FAILED: Expected unread_count=0 after read-all, got {after_res.get('unread_count')}")
        sys.exit(1)
    print("PASSED: Mark all as read successfully set unread_count to 0")

    print("\n=================================================")
    print(" ALL SPEC 3 INTEGRATION TESTS PASSED SUCCESSFULLY! ")
    print("=================================================")

if __name__ == "__main__":
    run_spec3_tests()
