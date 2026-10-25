# Report G: Remediation Roadmap

> **National Highways Infra Trust (NHIT) Digital Platform**  
> **Engineering Action Plan:** Prioritized Security Remediation & Hardening Roadmap  
> **Standard:** OWASP Risk Rating Methodology, NIST SP 800-53 Rev 5  
> **Date:** October 24, 2026

---

## 1. Remediation Priority Framework

Remediations are organized into four strict operational tiers:
- **P0 — Critical / Blocker (Immediate Containment):** Must be resolved prior to any production exposure or external user pilot.
- **P1 — High Priority (Pre-Production Gate):** Must be resolved before commercial launch or financial disbursements.
- **P2 — Medium Priority (Sprint Hardening):** Scheduled for completion within 30 days post-launch.
- **P3 — Low / Informational (Backlog & Hygiene):** Code quality, observability, and performance optimization.

---

## 2. Prioritized Action Plan

### Tier P0: Critical / Blocker Items (Immediate Action)

| ID | Vulnerability & Remediation Action | Effort | Prerequisites | Acceptance Criteria & Verification Plan |
|---|---|:---:|---|---|
| **P0-1** | **Fix Universal Auth Bypass (`SEC-001`)**<br>Remove `if os.environ.get("ENVIRONMENT") == "test"` bypass in `gateway_swagger.py:get_current_user`. Scope mock auth exclusively to test runner fixtures. | 2 Hours | None | Calling any protected endpoint with no `Authorization` header must strictly return HTTP 401 Unauthorized. |
| **P0-2** | **Protect Vendor Bank Injections (`SEC-002`)**<br>Add `Depends(get_current_user)` and RBAC checks (`finance_controller`, `admin`) to `POST /api/v1/vendors/{vendor_id}/accounts`. | 2 Hours | P0-1 | Anonymous calls return 401. Non-finance users return 403. Primary account switches require verified admin signature. |
| **P0-3** | **Eliminate 2FA Bypass in Password Change (`SEC-009`)**<br>In `gateway_swagger.py:change_password`, disallow `2fa_pending` tokens. Enforce that 2FA must be verified before password mutation. | 2 Hours | None | Submitting a `2fa_pending` token to `/api/v1/auth/change-password` returns HTTP 401 Unauthorized. |
| **P0-4** | **Enforce Route Authentication Across All Business Routes (`SEC-016`)**<br>Add `Depends(get_current_user)` to all 88 sensitive endpoints across Vendor, Green Note, Payment Note, Payment, and Reimbursement services. | 8 Hours | P0-1 | Automated test harness confirms all 88 sensitive business routes reject unauthenticated requests with HTTP 401. |

---

### Tier P1: High Priority (Pre-Production Launch Gates)

| ID | Vulnerability & Remediation Action | Effort | Prerequisites | Acceptance Criteria & Verification Plan |
|---|---|:---:|---|---|
| **P1-1** | **Migrate Financial Notes to Persistent PostgreSQL (`SEC-011`)**<br>Replace in-memory Python lists `GREEN_NOTES_STORE` and `PAYMENT_NOTES_STORE` with PostgreSQL DDL migrations and ACID queries. | 2 Days | P0-4 | Green Notes and Payment Notes persist across gateway restarts; verified by container reboot test. |
| **P1-2** | **Secure Notification Dispatch & Remove Cleartext OTP (`SEC-003`)**<br>Add authentication and strict rate limits to `/api/v1/notifications/test-otp`. Remove `otp_generated` from response payload. | 4 Hours | P0-1 | Endpoint returns 401 for anonymous callers; OTP is never leaked in JSON; rate limited to 3 requests per 5 minutes. |
| **P1-3** | **Remove Plaintext Fallback & Password Backdoor (`SEC-007`)**<br>Enforce strict bcrypt comparison. Remove `body.password == "password123"` exception backdoor from `login`. | 2 Hours | None | Invalid passwords strictly return 401; corrupted user hashes fail closed rather than accepting `"password123"`. |
| **P1-4** | **Enforce Financial Disbursement Calculation Invariant (`SEC-015`)**<br>Compute `net_payable_amount` server-side: `gross_amount - tds_amount - retention`. Reject client overrides. | 4 Hours | P1-1 | Payment Note creation verifies formula; mismatches are rejected with HTTP 422 Unprocessable Entity. |
| **P1-5** | **Enforce Sequential State Transitions on Green Notes (`SEC-014`)**<br>Implement Finite State Machine: `DRAFT -> CHECKED -> VERIFIED -> APPROVED`. Disallow arbitrary stage skipping. | 6 Hours | P1-1 | Attempting to transition directly from `DRAFT` to `APPROVED` returns HTTP 400 with statutory progression error. |
| **P1-6** | **Rotate Exposed Credentials & Purge Committed Executable (`SEC-013`)**<br>Revoke Resend and SMTP keys. Add `.env` to `.gitignore`. Remove 18.5 MB `payment-server` binary from git history. | 3 Hours | None | Repository scan detects 0 tracked `.env` files and 0 binary executables. All live keys rotated. |
| **P1-7** | **Isolate PostgreSQL Port & Enforce Production JWT Secret (`SEC-006`)**<br>Remove `ports: - "5433:5432"` from `docker-compose.yml`. Fail container startup if `JWT_SECRET` is missing. | 2 Hours | None | PostgreSQL is inaccessible from host interface; gateway container crashes if `JWT_SECRET` is unset in production. |

---

### Tier P2: Medium Priority (Sprint Hardening — 30-Day Window)

| ID | Vulnerability & Remediation Action | Effort | Prerequisites | Acceptance Criteria & Verification Plan |
|---|---|:---:|---|---|
| **P2-1** | **Configure Nginx HTTPS / TLS Termination & HSTS (`SEC-005`)**<br>Add `listen 443 ssl` block, redirect HTTP port 80 to 443, emit HSTS header, and restrict `/docs` and `/openapi.json`. | 6 Hours | Infrastructure TLS cert | `curl -I http://domain/` returns 301 to HTTPS. SSL Labs achieves A+ rating. Documentation requires auth. |
| **P2-2** | **Migrate Frontend JWT Storage to HttpOnly Cookies (`SEC-004`)**<br>Replace `localStorage.setItem('nhit_auth_token')` with `HttpOnly`, `Secure`, `SameSite=Strict` cookies. | 1 Day | P0-1 | Access tokens cannot be accessed via browser `document.cookie` or JavaScript console; XSS exfiltration mitigated. |
| **P2-3** | **Implement Server-Side JWT Token Invalidation (`SEC-008`)**<br>Implement Redis token denylist in `logout`. Check denylist in `get_current_user`. Implement real password reset. | 1 Day | Redis | Calling `/api/v1/auth/logout` immediately invalidates the token; subsequent requests with that token return 401. |
| **P2-4** | **Deploy Distributed Redis Rate Limiting (`SEC-012`)**<br>Replace `RATE_LIMIT_STORE` with Redis sliding-window algorithm and apply global limits to all public and authenticated endpoints. | 1 Day | Redis | Volumetric request bursts return HTTP 429 across multi-instance clusters; keys expire cleanly in Redis. |

---

### Tier P3: Low Priority & Technical Hygiene (Backlog)

| ID | Vulnerability & Remediation Action | Effort | Prerequisites | Acceptance Criteria & Verification Plan |
|---|---|:---:|---|---|
| **P3-1** | **Remove Synchronous Debug Disk I/O in Go Interceptor (`SEC-010`)**<br>Remove synchronous writes to `rbac_debug.log` in `NHIT Backend/pkg/middleware/rbac.go`. | 1 Hour | None | No `rbac_debug.log` file created; logging redirected to structured stdout logger. |
| **P3-2** | **Tighten Gateway Reverse-Proxy Isolation**<br>Change gateway in `docker-compose.yml` from `ports: - "8080:8080"` to `expose: - "8080"`, forcing all traffic through Nginx. | 1 Hour | P2-1 | Direct connections to port 8080 fail; all client traffic is logged and filtered by Nginx. |
| **P3-3** | **Multi-Tenancy Query Scoping Audit**<br>Audit all database queries in `gateway_swagger.py` to ensure `AND tenant_id = %s` is enforced across all CRUD operations. | 1 Day | P0-4 | Cross-tenant data access tests confirm Tenant A cannot view or update resources in Tenant B. |

---

## 3. Verification & Regression Testing Strategy

Every patch must be validated using the project's automated verification harness:

```bash
# 1. Verify frontend TypeScript compilation with zero errors
cd nhit-frontend
npx tsc --noEmit

# 2. Verify backend Python syntax and bytecode compilation
python -m py_compile "NHIT Backend/gateway_swagger.py"

# 3. Execute the full 147-check verification harness
python "NHIT Backend/scripts/harness.py" --fast --bail

# 4. Export consolidated HTML and Markdown audit verification reports
python "NHIT Backend/scripts/harness.py" --fast --html harness/reports/post_remediation_report.html --markdown harness/reports/post_remediation_report.md
```
