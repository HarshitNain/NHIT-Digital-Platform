---
name: nhit-harness
description: >-
  Use this skill to run the NHIT test harness, verify database connectivity, validate frontend and backend health, and execute platform verification checks.
---

# NHIT Platform Verification & Test Harness

This skill provides step-by-step instructions to verify the integrity of the NHIT Digital Platform across backend services, database migrations, dynamic API routing, OWASP security, and frontend compilation.

## 1. Automated Verification Harness (103+ Checks)
Run the unified platform test harness script from the workspace root:
```bash
python "NHIT Backend/scripts/harness.py"
```

### CLI Execution Options:
- **Fast Execution** (skips slower subprocesses like strict TypeScript):
  ```bash
  python "NHIT Backend/scripts/harness.py" --fast
  ```
- **Isolated Phase Execution**:
  ```bash
  python "NHIT Backend/scripts/harness.py" --phase 1  # Runtime & Dependencies
  python "NHIT Backend/scripts/harness.py" --phase 2  # Compilation & Contracts
  python "NHIT Backend/scripts/harness.py" --phase 3  # Database & Migrations
  python "NHIT Backend/scripts/harness.py" --phase 4  # Dynamic FastAPI & OWASP
  python "NHIT Backend/scripts/harness.py" --phase 5  # Financial & Statutory Invariants
  ```
- **Fail Fast / Bail on First Failure** (ideal for CI/CD):
  ```bash
  python "NHIT Backend/scripts/harness.py" --bail
  ```
- **Machine-Readable JSON Output**:
  ```bash
  python "NHIT Backend/scripts/harness.py" --json
  ```

### What the Harness Validates:
1. **Phase 1: Runtime & Dependency Matrix**: Python 3.10+, FastAPI, Uvicorn, Pydantic, PyJWT, Psycopg2, Bcrypt, Redis, Node.js 20+, NPM 10+, directory structure.
2. **Phase 2: Code Compilation & Contracts**: `gateway_swagger.py` & `otp_notifier.py` py-compile, frontend model types (9 interfaces), API client methods (10 methods), modal component exports (11 modals), Vendor KYC multi-bank tab binding, and `npx tsc --noEmit`.
3. **Phase 3: Database & Migrations**: 14 DDL tables, IFSC/account constraints, single-primary triggers, foreign keys, live PostgreSQL connectivity.
4. **Phase 4: Dynamic In-Process FastAPI & OWASP**: OpenAPI 3.0 specs, interactive Swagger, `/api/v1/services`, `/api/v1/health`, UUID sanitization, SQL injection fuzzing, Green Notes & Payment Notes endpoints, Project & Reimbursement APIs, status whitelists, CORS, and Bearer token verification.
5. **Phase 5: Financial, Cryptographic & Statutory Invariants**: HMAC-SHA256 JWT lifecycle, production JWT safety policy, 2FA OTP lifecycle, 10-thread concurrent stress test, sliding-window rate limiting, Income Tax PAN regex, GSTIN regex, RBI IFSC regex, MSME categorization, Section 206AB higher TDS guard, non-negative disbursement formula, and single primary account exclusivity.

## 2. Independent Backend Check
```bash
python -m py_compile "NHIT Backend/gateway_swagger.py"
python -m py_compile "NHIT Backend/otp_notifier.py"
```

## 3. Independent Frontend Check
From the `nhit-frontend/` directory:
```bash
npx tsc --noEmit
```

## 4. Database Reset & Reseed (If Needed)
To re-seed test data:
```bash
psql -h localhost -p 5433 -U postgres -d nhit_db -f "NHIT Backend/scripts/seed-test-data.sql"
```
Or use the Python seeder:
```bash
python "NHIT Backend/scripts/setup_test_db.py"
```
