"""
Automated API Test Suite: Tests all primary endpoints in gateway_swagger.py as an end-user.
Records status codes, response shapes, errors, and UX issues encountered during testing.
"""

import urllib.request
import urllib.error
import json
import sys

if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8')

BASE_URL = "http://localhost:8080"

results = []

def test_endpoint(name, method, path, body=None, headers=None, expected_status=200):
    url = f"{BASE_URL}{path}"
    req_headers = {"Content-Type": "application/json"}
    if headers:
        req_headers.update(headers)
    
    data = json.dumps(body).encode() if body else None
    req = urllib.request.Request(url, data=data, headers=req_headers, method=method)
    
    try:
        with urllib.request.urlopen(req) as resp:
            code = resp.getcode()
            body_res = resp.read().decode()
            try:
                parsed = json.loads(body_res)
            except:
                parsed = body_res[:100]
            
            passed = (code == expected_status)
            results.append({
                "name": name,
                "method": method,
                "path": path,
                "status": code,
                "passed": passed,
                "response": parsed,
                "error": None
            })
            return parsed
    except urllib.error.HTTPError as e:
        body_err = e.read().decode()
        try:
            err_json = json.loads(body_err)
        except:
            err_json = body_err
        passed = (e.code == expected_status)
        results.append({
            "name": name,
            "method": method,
            "path": path,
            "status": e.code,
            "passed": passed,
            "response": None,
            "error": err_json
        })
        return err_json
    except Exception as e:
        results.append({
            "name": name,
            "method": method,
            "path": path,
            "status": 0,
            "passed": False,
            "response": None,
            "error": str(e)
        })
        return None

def run_all_tests():
    print("🧪 Running API End-User Test Suite...")
    
    # 1. Health Check
    test_endpoint("Health Check", "GET", "/api/v1/health")
    
    # 2. Login Valid Admin
    login_res = test_endpoint("Login Admin", "POST", "/api/v1/auth/login", {"login": "admin@nhit.com", "password": "password123"})
    token = login_res.get("token") if isinstance(login_res, dict) else ""
    auth_header = {"Authorization": f"Bearer {token}"} if token else {}
    
    # 3. Login Invalid User
    test_endpoint("Login Non-existent User", "POST", "/api/v1/auth/login", {"login": "fake@nhit.com", "password": "wrong"}, expected_status=404)
    
    # 4. Token Validate
    test_endpoint("Validate Token", "POST", "/api/v1/auth/validate-token", {"token": token})
    
    # 5. List Organizations
    orgs_res = test_endpoint("List Organizations", "GET", "/api/v1/organizations")
    first_org_id = orgs_res.get("organizations", [{}])[0].get("org_id", "") if isinstance(orgs_res, dict) else ""
    
    # 6. Get Organization by ID
    if first_org_id:
        test_endpoint(f"Get Org ({first_org_id})", "GET", f"/api/v1/organizations/{first_org_id}")
        test_endpoint("Get Org with Projects", "GET", f"/api/v1/organizations/{first_org_id}/with-projects")
    
    # 7. Get Org by Code
    test_endpoint("Get Org by Code", "GET", "/api/v1/organizations/code/NHIT-TECH")
    
    # 8. List Departments
    depts_res = test_endpoint("List Departments", "GET", "/api/v1/departments")
    first_dept_id = depts_res.get("departments", [{}])[0].get("id", "") if isinstance(depts_res, dict) else ""
    if first_dept_id:
        test_endpoint("Get Department", "GET", f"/api/v1/departments/{first_dept_id}")
        
    # 9. List Designations
    desigs_res = test_endpoint("List Designations", "GET", "/api/v1/designations")
    first_desig_id = desigs_res.get("designations", [{}])[0].get("id", "") if isinstance(desigs_res, dict) else ""
    if first_desig_id:
        test_endpoint("Get Designation", "GET", f"/api/v1/designations/{first_desig_id}")
        
    # 10. List Users
    users_res = test_endpoint("List Users", "GET", "/api/v1/users")
    first_user_id = users_res.get("users", [{}])[0].get("user_id", "") if isinstance(users_res, dict) else ""
    if first_user_id:
        test_endpoint("Get User by ID", "GET", f"/api/v1/users/{first_user_id}")
        
    # 11. Dropdowns
    test_endpoint("Dropdown Departments", "GET", "/api/v1/users/dropdowns/departments")
    test_endpoint("Dropdown Designations", "GET", "/api/v1/users/dropdowns/designations")
    test_endpoint("Dropdown Roles", "GET", "/api/v1/users/dropdowns/roles")
    
    # 12. Roles & Permissions
    test_endpoint("List Roles", "GET", "/api/v1/roles")
    test_endpoint("List Permissions", "GET", "/api/v1/permissions")
    
    # 13. Projects
    test_endpoint("List Projects by Org", "GET", f"/api/v1/organizations/{first_org_id}/projects")
    test_endpoint("Get Project by ID", "GET", "/api/v1/projects/950e8400-e29b-41d4-a716-446655440001")
    
    # 14. Vendors
    vendors_res = test_endpoint("List Vendors", "GET", "/api/v1/vendors")
    first_vendor_id = vendors_res.get("vendors", [{}])[0].get("id", "") if isinstance(vendors_res, dict) else ""
    if first_vendor_id:
        test_endpoint("Get Vendor by ID", "GET", f"/api/v1/vendors/{first_vendor_id}")
        test_endpoint("Get Vendor Accounts", "GET", f"/api/v1/vendors/{first_vendor_id}/accounts")
    test_endpoint("Get Vendor by Code", "GET", "/api/v1/vendors/code/VEN-LNT-001")
    
    # 15. Notes
    test_endpoint("List Green Notes", "GET", "/api/v1/green-notes")
    test_endpoint("List Payment Notes", "GET", "/api/v1/payment-notes")
    test_endpoint("List Reimbursements", "GET", "/api/v1/reimbursements")
    test_endpoint("List Activity Logs", "GET", "/api/v1/activity-logs")
    
    # Print Summary Report
    print("\n" + "=" * 80)
    print("🎯 TEST EXECUTION REPORT")
    print("=" * 80)
    passed_count = sum(1 for r in results if r["passed"])
    failed_count = len(results) - passed_count
    
    for r in results:
        status_sym = "✅" if r["passed"] else "❌"
        print(f"{status_sym} [{r['status']}] {r['method']} {r['path']} - {r['name']}")
        if not r["passed"]:
            print(f"   ⚠️ ERROR: {r['error']}")
        elif isinstance(r["response"], dict) and "total_count" in r["response"]:
            print(f"   📊 Count: {r['response']['total_count']}")
        elif isinstance(r["response"], dict) and "pagination" in r["response"]:
            print(f"   📊 Total items: {r['response']['pagination'].get('total_items')}")

    print("\n" + "-" * 80)
    print(f"Total Tests: {len(results)} | Passed: {passed_count} | Failed: {failed_count}")
    print("-" * 80)

if __name__ == "__main__":
    run_all_tests()
