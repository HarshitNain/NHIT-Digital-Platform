# Report D: Authentication and Authorization Matrix

> **National Highways Infra Trust (NHIT) Digital Platform**  
> **Assessment Scope:** Complete Role-Based Access Control (RBAC), Multi-Tenancy & Authorization Invariants  
> **Standard:** OWASP Top 10:2021 A01 (Broken Access Control), OWASP API Security Top 10:2023 API1/API2/API5  
> **Date:** October 24, 2026

---

## 1. Platform Roles & Intended Privilege Hierarchy

The NHIT enterprise architecture defines 7 distinct operational roles:

| Role Identifier | Role Name | Intended Business Authority & Scope |
|---|---|---|
| `super_admin` | Super Administrator | Full platform control: tenant creation, RBAC role management, global configuration (`*`). |
| `admin` | Organization Administrator | User management within organization, project allocation, vendor approval. |
| `finance_controller` | Financial Controller | Green Note verification (Tier 3), Payment Note validation, tax & TDS audit, disbursement release. |
| `project_engineer` | Project Engineer / DGM | Green Note initiation (Tier 1) and checking (Tier 2), milestone progress reporting. |
| `employee` | General Staff Member | Reimbursement note submission, profile viewing, internal self-service. |
| `auditor` | Statutory Compliance Auditor | Read-only access to financial logs, payment records, bank letters, and audit trails. |
| `external_vendor` | External Contractor | Restricted portal: view allocated purchase orders, submit invoices, view payment/UTR status. |

---

## 2. Comprehensive RBAC Enforcement Matrix

The following matrix contrasts **Intended Business Permissions** against **Actual Server-Side Enforcement** identified during static AST analysis and dynamic security inspection of `NHIT Backend/gateway_swagger.py`.

### Legend:
- `✓ ENFORCED`: Token and role cryptographically validated on the server.
- `⚠️ AUTH ONLY`: Validates Bearer token exists, but DOES NOT check the caller's specific role (any authenticated user can execute).
- `❌ EXPOSED`: **No authentication dependency at all**. Any anonymous external network caller can execute this operation.
- `🚨 BYPASS`: Bypassed whenever `ENVIRONMENT=test` (defaults to test in repository configuration).

| Domain | Operation / Endpoint | HTTP Method | Intended Authorized Roles | Actual Server-Side Status | Enforcement Gap / Vulnerability |
|---|---|:---:|---|:---:|---|
| **Auth** | Login (`/api/v1/auth/login`) | POST | Public | Public (Intended) | Password fallback backdoor (`SEC-007`) |
| **Auth** | Verify 2FA OTP (`/api/v1/auth/verify-otp`) | POST | Public (Session Token) | Public (Intended) | Sliding rate-limit bypass |
| **Auth** | Passwordless OTP (`/api/v1/auth/passwordless/request-otp`) | POST | Public | Public (Intended) | Email verification required |
| **Auth** | Change Password (`/api/v1/auth/change-password`) | POST | Auth / Password Reset | ❌ CRITICAL BYPASS | Accepts `2fa_pending` tokens (`SEC-009`) |
| **Auth** | Logout (`/api/v1/auth/logout`) | POST | Authenticated User | ❌ EXPOSED / NOOP | No token revocation or denylist (`SEC-008`) |
| **Users** | List Users (`/api/v1/users`) | GET | `super_admin`, `admin` | ⚠️ AUTH ONLY (🚨 BYPASS) | Role not checked; any user can list users |
| **Users** | Create User (`/api/v1/users`) | POST | `super_admin`, `admin` | ⚠️ AUTH ONLY (🚨 BYPASS) | Role not checked; any user can create admins |
| **Users** | Update User (`/api/v1/users/{id}`) | PUT | `super_admin`, `admin` | ⚠️ AUTH ONLY (🚨 BYPASS) | BOLA/IDOR: Any user can modify any account |
| **Users** | Delete User (`/api/v1/users/{id}`) | DELETE | `super_admin` | ⚠️ AUTH ONLY (🚨 BYPASS) | Regular users can delete administrator accounts |
| **Vendors** | List Vendors (`/api/v1/vendors`) | GET | All Authenticated Staff | ❌ EXPOSED | Publicly accessible without Bearer token |
| **Vendors** | Get Vendor Detail (`/api/v1/vendors/{id}`) | GET | All Authenticated Staff | ❌ EXPOSED | Publicly accessible without Bearer token |
| **Vendors** | Create Vendor Account (`/api/v1/vendors/{id}/accounts`) | POST | `admin`, `finance_controller` | ❌ EXPOSED | **Unauthenticated account injection** (`SEC-002`) |
| **Vendors** | List Vendor Accounts (`/api/v1/vendors/{id}/accounts`) | GET | `admin`, `finance_controller` | ❌ EXPOSED | Bank account numbers exposed to public |
| **Vendors** | Update Vendor Account (`/api/v1/vendors/accounts/{id}`) | PUT | `admin`, `finance_controller` | ❌ EXPOSED | Unauthenticated bank modification |
| **Vendors** | Delete Vendor Account (`/api/v1/vendors/accounts/{id}`) | DELETE | `super_admin`, `admin` | ❌ EXPOSED | Unauthenticated bank deletion |
| **Green Notes** | Create Green Note (`/api/v1/green-notes`) | POST | `project_engineer`, `admin` | ❌ EXPOSED | Unauthenticated note creation (`SEC-016`) |
| **Green Notes** | List Green Notes (`/api/v1/green-notes`) | GET | Authenticated Staff | ❌ EXPOSED | Unauthenticated note listing |
| **Green Notes** | Update Status (`/api/v1/green-notes/{id}/status`) | PATCH | Tier-Specific Approver | ❌ EXPOSED | **Arbitrary approval skipping** (`SEC-014`) |
| **Green Notes** | Cancel Note (`/api/v1/green-notes/{id}/cancel`) | POST | `project_engineer`, `admin` | ❌ EXPOSED | Unauthenticated note cancellation |
| **Payment Notes** | Create Payment Note (`/api/v1/payment-notes`) | POST | `finance_controller` | ❌ EXPOSED | Client controls net amount (`SEC-015`) |
| **Payment Notes** | List Payment Notes (`/api/v1/payment-notes`) | GET | `finance_controller`, `auditor` | ❌ EXPOSED | Unauthenticated payment listing |
| **Payment Notes** | Update UTR (`/api/v1/payment-notes/{id}/utr`) | POST | `finance_controller` | ❌ EXPOSED | Unauthenticated UTR falsification |
| **Payment Notes** | Put on Hold (`/api/v1/payment-notes/{id}/hold`) | POST | `finance_controller`, `admin` | ❌ EXPOSED | Unauthenticated disbursement freeze |
| **Payments** | List Payments (`/api/v1/payments`) | GET | `finance_controller` | ❌ EXPOSED | Financial data exposed to public |
| **Payments** | Create Payment Request (`/api/v1/payments`) | POST | `finance_controller` | ❌ EXPOSED | Unauthenticated disbursement initiation |
| **Payments** | Execute Shortcut (`/api/v1/payments/shortcuts/{id}/execute`) | POST | `finance_controller` | ❌ EXPOSED | Unauthenticated batch payout execution |
| **Reimbursements** | List Claims (`/api/v1/reimbursement-notes`) | GET | Authenticated Staff | ❌ EXPOSED | Personal employee claims publicly visible |
| **Reimbursements** | Create Claim (`/api/v1/reimbursement-notes`) | POST | `employee` | ❌ EXPOSED | Unauthenticated reimbursement filing |
| **Reimbursements** | Approve Claim (`/api/v1/reimbursement-notes/{id}/approval-log`) | POST | `finance_controller`, `admin` | ❌ EXPOSED | Unauthenticated claim self-approval |
| **Notifications** | Live Test OTP (`/api/v1/notifications/test-otp`) | POST | `super_admin` (Dev Only) | ❌ EXPOSED | Live SMS/email bombing vector (`SEC-003`) |

---

## 3. High-Risk Authorization Findings

### Finding 1: Pervasive Missing Route-Level Authentication
- **Measurement:** 88 out of 104 sensitive business routes lack `Depends(get_current_user)`.
- **Mechanism:** In FastAPI, unless a dependency like `current_user: dict = Depends(get_current_user)` is explicitly declared in the route handler signature, FastAPI treats the endpoint as completely unauthenticated.
- **Risk:** Any client with HTTP access can trigger financial disbursements, update UTR codes, create fake green notes, and delete bank accounts without presenting any credentials.

### Finding 2: Universal `ENVIRONMENT=test` Authentication Nullification
- **Mechanism:** `get_current_user` returns a mock `Super Admin` payload whenever `os.environ.get("ENVIRONMENT") == "test"`.
- **Risk:** Because `NHIT Backend/.env` defaults to `ENVIRONMENT=test`, any deployment that does not manually override this variable will run with authentication completely disabled, granting full super-admin access to any caller providing an empty token header.

### Finding 3: Multi-Tenancy and Broken Object Level Authorization (BOLA / IDOR)
- **Mechanism:** While the database schema includes `tenant_id` and `org_id` columns, query handlers in `gateway_swagger.py` frequently execute queries such as:
  ```sql
  SELECT * FROM users WHERE user_id = %s
  ```
  without constraining `AND tenant_id = %s`.
- **Risk:** A user in Tenant A can view, modify, or delete resources belonging to Tenant B by passing the victim's UUID in the path parameter.

---

## 4. Role-Based Access Control Remediation Strategy

1. **Implement Global Gateway Authentication Middleware:**
   - Configure a global FastAPI middleware that rejects any incoming request lacking a valid Bearer token by default, maintaining an explicit whitelist of public authentication endpoints (`/health`, `/api/v1/auth/login`, `/api/v1/auth/verify-otp`, `/api/v1/auth/sso/google`).
2. **Implement Role-Enforcement Decorator (`@require_roles`):**
   ```python
   def require_roles(*allowed_roles: str):
       def dependency(user: dict = Depends(get_current_user)):
           user_roles = set(user.get("roles", []))
           if "*" in user_roles or "super_admin" in user_roles:
               return user
           if not user_roles.intersection(allowed_roles):
               raise HTTPException(403, f"Access denied. Requires one of roles: {allowed_roles}")
           return user
       return Depends(dependency)
   ```
3. **Enforce Tenant Isolation in All Queries:**
   - Always extract `tenant_id` from the verified JWT claims and bind it to SQL queries (`WHERE id = %s AND tenant_id = %s`).
