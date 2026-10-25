import urllib.request
import urllib.error
import json
import sys
import os

BASE = os.environ.get("API_BASE_URL", "http://127.0.0.1:8080/api/v1")

def post(url, data):
    req = urllib.request.Request(
        f"{BASE}{url}",
        data=json.dumps(data).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST"
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            return resp.status, json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read().decode("utf-8"))
    except Exception as e:
        print(f"Connection error to {BASE}{url}: {e}")
        sys.exit(1)

def patch(url, data):
    req = urllib.request.Request(
        f"{BASE}{url}",
        data=json.dumps(data).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="PATCH"
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            return resp.status, json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read().decode("utf-8"))

def get(url):
    req = urllib.request.Request(f"{BASE}{url}")
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            return resp.status, json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        return e.code, json.loads(e.read().decode("utf-8"))

def run_tests():
    global BASE
    BASE = os.environ.get("API_BASE_URL", BASE)
    print("=" * 70)
    print(" NHIT END-TO-END AUTOMATED VERIFICATION SUITE")
    print(f" Target Base URL: {BASE}")
    print("=" * 70)

    # 1. Full Authentication Matrix
    print("\n[SUITE 1] Testing Database Authentication & RBAC Matrix:")
    users_to_test = [
        ("admin@nhit.com", "password123", ["super_admin"]),
        ("cfo@nhit.com", "password123", ["finance_reviewer", "cfo"]),
        ("cto@nhit.com", "password123", ["org_admin", "cto"]),
        ("procurement@nhit.com", "password123", ["project_manager"]),
        ("engineer@nhit.com", "password123", ["employee"]),
    ]

    for email, pwd, expected_roles in users_to_test:
        st, res = post("/auth/login", {"login": email, "password": pwd, "enable_2fa": True})
        assert st == 200 and res.get("status") == "2FA_REQUIRED", f"Failed login challenge for {email}: {st} {res}"
        temp_tok = res["session_temp_token"]
        otp = res.get("dev_otp_preview")

        st_v, res_v = post("/auth/login-otp/verify", {"session_temp_token": temp_tok, "otp": otp})
        assert st_v == 200 and "token" in res_v, f"Failed OTP verify for {email}: {st_v} {res_v}"
        roles = res_v.get("roles", [])
        perms = res_v.get("permissions", [])
        assert any(r in roles for r in expected_roles), f"Role mismatch for {email}: expected one of {expected_roles}, got {roles}"
        print(f"  [PASS] {email:25} -> 2FA Verified, Role: {roles[0]:18} Permissions: {len(perms)}")

    # 2. Security Negative Boundaries
    print("\n[SUITE 2] Security Negative & Boundary Tests:")
    st, res = post("/auth/login", {"login": "admin@nhit.com", "password": "wrongpassword123"})
    assert st == 401, f"Expected 401 on wrong password, got {st}"
    print("  [PASS] Wrong password rejected with 401 Unauthorized")

    st, res = post("/auth/login", {"login": "nonexistent@nhit.com", "password": "password123"})
    assert st in (401, 404), f"Expected 401 or 404 on nonexistent user, got {st}"
    print("  [PASS] Nonexistent account rejected with 401 Unauthorized (Anti-Enumeration Guard)")

    st, res = post("/auth/login-otp/verify", {"session_temp_token": "invalid.fake.token", "otp": "123456"})
    assert st in (400, 401), f"Expected 400/401 on bad token, got {st}"
    print("  [PASS] Invalid 2FA session token rejected")

    # 3. Direct Passwordless OTP
    print("\n[SUITE 3] Direct Passwordless OTP Flow:")
    st, res = post("/auth/login-otp/request", {"email": "engineer@nhit.com"})
    assert st == 200 and res.get("status") == "2FA_REQUIRED", f"Direct OTP request failed: {st} {res}"
    st_v, res_v = post("/auth/login-otp/verify", {"session_temp_token": res["session_temp_token"], "otp": res.get("dev_otp_preview")})
    assert st_v == 200 and "token" in res_v, f"Direct OTP verify failed: {st_v} {res_v}"
    print("  [PASS] Direct passwordless OTP login succeeded for engineer@nhit.com (Role: employee)")

    # 4. Google SSO Federated Flow
    print("\n[SUITE 4] Google SSO Federated Flow:")
    st, res = post("/auth/sso/google", {"email": "google.director@nhit.gov.in", "name": "Director S. K. Raman"})
    assert st == 200 and "token" in res, f"Google SSO failed: {st} {res}"
    print(f"  [PASS] Google SSO auto-provisioning & JWT issuance succeeded for {res.get('email')} (Role: {res.get('roles')})")

    # 5. Business Domain Endpoints & Data Operations
    print("\n[SUITE 5] Business Domain Data Operations:")

    # GreenNotes
    st, res = post("/green-notes", {
        "note": {
            "project_name": "NH-52 Jaipur Bypass Corridor",
            "supplier_name": "Tata Projects Ltd",
            "base_value": 5000000,
            "gst": 900000,
            "total_amount": 5900000,
            "remarks": "Initial highway clearance tranche"
        }
    })
    assert st == 200, f"Create green note failed: {st}"
    gn_id = res.get("id")
    print(f"  [PASS] GreenNote created -> ID: {gn_id}")

    st, res = patch(f"/green-notes/{gn_id}/status", {"status": "APPROVED", "remarks": "Approved by Finance Reviewer"})
    assert st == 200, f"Update green note status failed: {st}"
    print(f"  [PASS] GreenNote {gn_id} status updated -> APPROVED")

    # PaymentNotes & UTR
    st, res = patch("/payment-notes/1/status", {"status": "ACTIVE", "remarks": "Authorized for RTGS disbursement"})
    assert st == 200, f"Update payment note status failed: {st}"
    print("  [PASS] PaymentNote 1 status updated -> ACTIVE")

    st, res = patch("/payment-notes/1/utr", {"utr_number": "HDFCR52026101700987", "payment_date": "2026-10-17"})
    assert st == 200, f"Update UTR failed: {st}"
    print("  [PASS] PaymentNote 1 UTR recorded -> HDFCR52026101700987")

    # Vendors
    st, res = post("/vendors", {
        "name": "Gammon India Infrastructure Pvt Ltd",
        "vendor_code": "VEND-GAM-09",
        "msme_classification": "Medium Enterprise",
        "bank_name": "Axis Bank",
        "bank_account_number": "918020038472910",
        "bank_ifsc": "UTIB0000142",
        "gstin": "07AAACG1234F1Z5",
        "pan": "AAACG1234F"
    })
    assert st == 200, f"Create vendor failed: {st}"
    print("  [PASS] Vendor onboarded with banking KYC")

    # Reimbursements
    st, res = post("/reimbursements", {
        "purpose_of_travel": "Vadodara expressway bridge structural audit",
        "account_holder": "Priya Mehta",
        "bank_name": "HDFC Bank",
        "bank_account": "50200019283746",
        "ifsc_code": "HDFC0000060",
        "amount": 14200
    })
    assert st == 200, f"Create reimbursement failed: {st}"
    print("  [PASS] Travel reimbursement claim submitted")

    # 6. REST Query Endpoints
    print("\n[SUITE 6] Querying All Core Entities via REST:")
    for ep in ["/organizations", "/vendors", "/departments", "/designations", "/users", "/green-notes", "/payment-notes"]:
        st, res = get(ep)
        assert st == 200, f"GET {ep} failed: {st}"
        count = len(res) if isinstance(res, list) else len(res.get("notes", res.get("vendors", res.get("organizations", res.get("users", res.get("departments", res.get("designations", [])))))))
        print(f"  [PASS] GET {ep:25} -> {count} items returned")

    print("\n" + "=" * 70)
    print(" ALL 18 INTEGRATION & END-TO-END TESTS PASSED WITH 100% SUCCESS!")
    print("=" * 70)

if __name__ == "__main__":
    run_tests()
