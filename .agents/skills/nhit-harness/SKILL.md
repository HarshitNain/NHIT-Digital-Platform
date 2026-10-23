---
name: nhit-harness
description: >-
  Use this skill to run the NHIT test harness, verify database connectivity, validate frontend and backend health, and execute platform verification checks.
---

# NHIT Platform Verification & Test Harness

This skill provides step-by-step instructions to verify the integrity of the NHIT Digital Platform across backend services, database migrations, dynamic API routing, OWASP security, and frontend compilation.

## 1. Automated Verification Harness (144+ Invariant Checks Across 7 Phases)
Run the unified platform test harness script from the workspace root:
```bash
# Using PowerShell runner (recommended on Windows)
.\run-harness.ps1

# Using Linux/macOS runner
./harness/run-harness.sh

# Or directly via Python
python "NHIT Backend/scripts/harness.py"
```

### CLI Execution Options:
- **Fast Execution** (skips slower subprocesses like strict TypeScript):
  ```bash
  .\run-harness.ps1 -Fast
  # or: python "NHIT Backend/scripts/harness.py" --fast
  ```
- **Isolated Phase Execution (1-7)**:
  ```bash
  .\run-harness.ps1 -Phase 1  # Runtime & Dependencies
  .\run-harness.ps1 -Phase 2  # Compilation & Contracts
  .\run-harness.ps1 -Phase 3  # Database & Migrations
  .\run-harness.ps1 -Phase 4  # Dynamic FastAPI & OWASP
  .\run-harness.ps1 -Phase 5  # Financial & Statutory Invariants
  .\run-harness.ps1 -Phase 6  # Microservices & gRPC Protobuf Contracts
  .\run-harness.ps1 -Phase 7  # Multi-Tier Workflow Simulations
  ```
- **Fail Fast / Bail on First Failure** (ideal for CI/CD):
  ```bash
  .\run-harness.ps1 -Fast -Bail
  ```
- **Automated Report Exports**:
  ```bash
  python "NHIT Backend/scripts/harness.py" --html harness/reports/harness_report.html --markdown harness/reports/harness_report.md
  ```
- **Machine-Readable JSON Output**:
  ```bash
  python "NHIT Backend/scripts/harness.py" --json
  ```

### What the Harness Validates:
1. **Phase 1: Runtime & Dependency Matrix**: Python 3.10+, FastAPI, Uvicorn, Pydantic, PyJWT, Psycopg2, Bcrypt, HTTPX, Redis, Node.js 20+, NPM 10+, directory structure.
2. **Phase 2: Code Compilation & Contracts**: `gateway_swagger.py` & `otp_notifier.py` py-compile, frontend model types (10 interfaces), API client methods (11 methods), modal component exports (11 modals), Vendor KYC multi-bank tab binding, and `npx tsc --noEmit`.
3. **Phase 3: Database & Migrations**: 14 DDL tables, IFSC/account constraints, single-primary triggers, foreign keys, live PostgreSQL connectivity.
4. **Phase 4: Dynamic In-Process FastAPI & OWASP**: OpenAPI 3.0 specs, interactive Swagger, `/api/v1/services`, `/api/v1/health`, UUID sanitization, SQL injection fuzzing, Green Notes & Payment Notes endpoints, Project & Reimbursement APIs, status whitelists, CORS, and Bearer token verification.
5. **Phase 5: Financial, Cryptographic & Statutory Invariants**: HMAC-SHA256 JWT lifecycle, production JWT safety policy, 2FA OTP lifecycle, 10-thread concurrent stress test, sliding-window rate limiting, Income Tax PAN regex, GSTIN regex, RBI IFSC regex, MSME categorization, Section 206AB higher TDS guard, non-negative disbursement formula, and single primary account exclusivity.
6. **Phase 6: Microservices Compilation & gRPC Protobuf Contracts**: 5 gRPC Proto schemas (`auth`, `greennote`, `paymentnote`, `payment`, `approval`), 5 generated `pb.go` stubs, 8 Go backend services, Nhit-Note root module, and E2E test subpackage isolation.
7. **Phase 7: End-to-End Multi-Tier Business Workflow Simulations**: 4-tier Green Note state progression, Payment Note net payable formula with TDS/retention, Vendor multi-bank fallback account synthesis, 2FA OTP lifecycle & anti-replay, and RTGS/NEFT statutory UTR identifier validation.

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
