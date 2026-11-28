"""Live E2E integration verification for Registration Email Verification (Issue #69).

Requires backend running with REQUIRE_EMAIL_VERIFICATION=true pointing to Mailpit
(SMTP_HOST=mailpit SMTP_PORT=1025 SMTP_TLS_MODE=none) and Mailpit web API reachable.
"""
import os
import re
import sys
import time
import uuid
import requests

BASE_URL = os.environ.get("API_BASE_URL", "http://127.0.0.1:8080/api")
MAILPIT_URL = os.environ.get("MAILPIT_API_URL", "http://127.0.0.1:8025/api/v1")
ADMIN_EMAIL = os.environ.get("ADMIN_EMAIL", "admin@soc.local")
ADMIN_PASSWORD = os.environ.get("ADMIN_PASSWORD", "")


def load_admin_creds():
    email, pw = ADMIN_EMAIL, ADMIN_PASSWORD
    if pw:
        return email, pw
    env_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", ".env")
    if os.path.exists(env_path):
        with open(env_path, "r", encoding="utf-8") as f:
            for line in f:
                s = line.strip()
                if s.startswith("ADMIN_EMAIL="):
                    email = s.split("=", 1)[1].strip() or email
                elif s.startswith("ADMIN_PASSWORD="):
                    pw = s.split("=", 1)[1].strip()
    return email, pw


def wait_for_mailpit_token(recipient_email: str, after_ids: set, timeout_sec: float = 15.0) -> tuple[str, str]:
    deadline = time.time() + timeout_sec
    while time.time() < deadline:
        resp = requests.get(f"{MAILPIT_URL}/messages", timeout=5)
        resp.raise_for_status()
        for msg in resp.json().get("messages") or []:
            msg_id = msg.get("ID")
            if msg_id in after_ids:
                continue
            tos = [t.get("Address", "") for t in (msg.get("To") or [])]
            if recipient_email not in tos:
                continue
            detail = requests.get(f"{MAILPIT_URL}/message/{msg_id}", timeout=5)
            detail.raise_for_status()
            d = detail.json()
            text_body = (d.get("Text") or "") + "\n" + (d.get("HTML") or "")
            assert "ยืนยันอีเมล" in text_body and "Verify Email" in text_body, (
                f"Expected dual-language TH+EN body in email {msg_id}"
            )
            m = re.search(r"/verify-email\?token=([A-Za-z0-9_-]+)", text_body)
            assert m, f"Could not extract verify-email token from email body:\n{text_body[:400]}"
            return msg_id, m.group(1)
        time.sleep(0.4)
    raise AssertionError(f"Timed out waiting for verification email to {recipient_email} in Mailpit")


def main() -> int:
    tag = uuid.uuid4().hex[:8]
    learner_email = f"verify_learner_{tag}@example.com"
    spec_email = f"verify_spec_{tag}@example.com"
    password = "StrongPass123!"

    # 1. Verify Learner registration -> 401 on wrong pw -> 403 AUTH_EMAIL_NOT_VERIFIED -> Mailpit verify -> 200 Login
    r_reg = requests.post(
        f"{BASE_URL}/register",
        json={
            "email": learner_email,
            "password": password,
            "nickname": f"vl_{tag}",
            "preferred_language": "th",
        },
        timeout=10,
    )
    assert r_reg.status_code == 201, f"register failed: {r_reg.status_code} {r_reg.text}"
    assert r_reg.json().get("requires_verification") is True, r_reg.text
    print(f"[PASS] Registered unverified learner {learner_email} (requires_verification=true)")

    r_bad = requests.post(f"{BASE_URL}/login", json={"email": learner_email, "password": "WrongPassword!"}, timeout=10)
    assert r_bad.status_code == 401, f"expected 401 on wrong password, got {r_bad.status_code}"

    r_blocked = requests.post(f"{BASE_URL}/login", json={"email": learner_email, "password": password}, timeout=10)
    assert r_blocked.status_code == 403 and r_blocked.json().get("code") == "AUTH_EMAIL_NOT_VERIFIED", (
        f"expected 403 AUTH_EMAIL_NOT_VERIFIED, got {r_blocked.status_code} {r_blocked.text}"
    )
    print("[PASS] Login blocked with 403 AUTH_EMAIL_NOT_VERIFIED while unverified (and 401 on wrong pw)")

    # Verify resend cooldown (429 within 60s) and enumeration-safe 200 on unknown account
    r_cooldown = requests.post(f"{BASE_URL}/resend-verification", json={"email": learner_email}, timeout=10)
    assert r_cooldown.status_code == 429 and r_cooldown.json().get("code") == "AUTH_VERIFICATION_COOLDOWN", (
        f"expected 429 AUTH_VERIFICATION_COOLDOWN, got {r_cooldown.status_code} {r_cooldown.text}"
    )
    r_unknown = requests.post(f"{BASE_URL}/resend-verification", json={"email": f"unknown_{tag}@example.com"}, timeout=10)
    assert r_unknown.status_code == 200, f"expected neutral 200 on unknown email, got {r_unknown.status_code}"
    print("[PASS] Resend cooldown (429 AUTH_VERIFICATION_COOLDOWN) and neutral unknown-email (200) verified")

    seen_ids: set = set()
    msg_id, raw_token = wait_for_mailpit_token(learner_email, seen_ids)
    seen_ids.add(msg_id)
    print(f"[PASS] Retrieved dual-language TH+EN email for learner from Mailpit (id={msg_id})")

    r_ver = requests.post(f"{BASE_URL}/verify-email", json={"token": raw_token}, timeout=10)
    assert r_ver.status_code == 200, f"verify-email failed: {r_ver.status_code} {r_ver.text}"

    r_replay = requests.post(f"{BASE_URL}/verify-email", json={"token": raw_token}, timeout=10)
    assert r_replay.status_code == 400, f"expected single-use replay 400, got {r_replay.status_code}"

    r_ok = requests.post(f"{BASE_URL}/login", json={"email": learner_email, "password": password}, timeout=10)
    assert r_ok.status_code == 200 and r_ok.json().get("token"), f"expected unlocked 200 login: {r_ok.text}"
    learner_hdrs = {"Authorization": f"Bearer {r_ok.json()['token']}"}
    r_prof = requests.get(f"{BASE_URL}/user/profile", headers=learner_hdrs, timeout=10)
    assert r_prof.status_code == 200 and r_prof.json().get("email_verified") is True, r_prof.text
    print("[PASS] Learner verified email (single-use), profile reflects email_verified=true, and unlocked Login (200 JWT)")

    # 2. Verify Atomic Specialist Registration + Admin Queue Filter -> Verify -> Appears in Admin Queue
    admin_email, admin_pw = load_admin_creds()
    assert admin_pw, "ADMIN_PASSWORD is required to verify Admin application queue"
    r_adm_login = requests.post(f"{BASE_URL}/login", json={"email": admin_email, "password": admin_pw}, timeout=10)
    assert r_adm_login.status_code == 200, f"Admin login failed: {r_adm_login.status_code} {r_adm_login.text}"
    adm_hdrs = {"Authorization": f"Bearer {r_adm_login.json()['token']}"}

    r_spec = requests.post(
        f"{BASE_URL}/register-specialist",
        data={
            "email": spec_email,
            "password": password,
            "nickname": f"vs_{tag}",
            "preferred_language": "en",
            "bio": "Blue team incident responder applying via atomic registration.",
        },
        files={"resume": ("resume.pdf", b"%PDF-1.4 spec-e2e-resume", "application/pdf")},
        timeout=15,
    )
    assert r_spec.status_code == 201 and r_spec.json().get("requires_verification") is True, (
        f"register-specialist failed: {r_spec.status_code} {r_spec.text}"
    )
    app_id = r_spec.json()["application_id"]
    print(f"[PASS] Atomic register-specialist persisted application {app_id} (unverified)")

    r_spec_blocked = requests.post(f"{BASE_URL}/login", json={"email": spec_email, "password": password}, timeout=10)
    assert r_spec_blocked.status_code == 403, f"expected specialist unverified 403 login, got {r_spec_blocked.status_code}"

    pending_before = requests.get(f"{BASE_URL}/admin/applications?status=pending", headers=adm_hdrs, timeout=10)
    assert pending_before.status_code == 200
    assert all(a["application_id"] != app_id for a in pending_before.json()), (
        "Unverified specialist application leaked into Admin pending queue!"
    )

    r_premature = requests.post(f"{BASE_URL}/admin/applications/{app_id}/approve", headers=adm_hdrs, timeout=10)
    assert r_premature.status_code == 400 and r_premature.json().get("code") == "ADMIN_APPLICANT_EMAIL_UNVERIFIED", (
        f"Expected 400 ADMIN_APPLICANT_EMAIL_UNVERIFIED, got {r_premature.status_code} {r_premature.text}"
    )
    print("[PASS] Unverified specialist excluded from Admin queue and blocked from premature approval")

    _, spec_token = wait_for_mailpit_token(spec_email, seen_ids)
    r_spec_ver = requests.post(f"{BASE_URL}/verify-email", json={"token": spec_token}, timeout=10)
    assert r_spec_ver.status_code == 200, f"specialist verify-email failed: {r_spec_ver.text}"

    pending_after = requests.get(f"{BASE_URL}/admin/applications?status=pending", headers=adm_hdrs, timeout=10)
    assert any(a["application_id"] == app_id for a in pending_after.json()), (
        "Verified specialist application did not appear in Admin pending queue!"
    )
    r_spec_login = requests.post(f"{BASE_URL}/login", json={"email": spec_email, "password": password}, timeout=10)
    assert r_spec_login.status_code == 200
    print("[PASS] Specialist verified -> appeared in Admin pending queue -> unlocked Login (200 JWT)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
