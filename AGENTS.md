# NHIT Digital Platform — Project Harness & Agent Memory

> **System Memory & Engineering Playbook for AI Pair Programmers and Developers**  
> **Repository:** `HarshitNain/NHIT-Digital-Platform`  
> **Target Enterprise:** National Highways Infra Trust (NHIT)  
> **Last Updated:** 2026-10-22

---

## 1. System Overview & Architecture

NHIT Digital Platform is an enterprise-grade financial management, green-noting, vendor procurement, and statutory compliance platform built for infrastructure asset management.

### Platform Tech Stack
- **API Gateway & Core Backend**: Python 3.11/3.12 + FastAPI (`NHIT Backend/gateway_swagger.py`). Exposes REST API on port `8000`, interactive Swagger documentation at `/docs`, OpenAPI spec at `/openapi.json`.
- **Frontend SPA**: React 18 / 19 + TypeScript + Vite (`nhit-frontend/`). Runs on port `5173`. Uses Vanilla CSS design system with custom tokens and Lucide icons.
- **Primary Database**: PostgreSQL 15 on port `5433` (DB: `nhit_db`, User: `postgres`, Password: `shristi`). Managed with consolidated schema migrations in `NHIT Backend/scripts/consolidated-migrations.sql`.
- **Distributed Cache & Session Store**: Redis on port `6379` (used for Two-Factor OTP storage and session invalidation; falls back gracefully to in-process store if offline).
- **Microservices**: Go services in `Nhit-Note/` and `notification-service/` for high-throughput note tracking and automated alerts.

---

## 2. Core Business Domains & Workflows

### 2.1. Green Notes (GN) — 4-Tier Financial Approval Workflow
Green Notes govern expenditure requests across NHIT highway stretches and corporate divisions.
- **Approval Types**: `INVOICE` (contractor billings), `ADVANCE` (mobilization advances), `ADHOC` (special project clearances).
- **4-Tier Approval Progression**:
  1. **Prepared**: Project Engineer (submits claim with bill breakdown).
  2. **Checked**: Dy. General Manager (technical and physical milestone validation).
  3. **Verified**: Financial Controller (tax deductions, budget balance, statutory compliance).
  4. **Sanctioned**: CGM / Trustee Board (final release sanction).
- **Status Codes**: `DRAFT`, `SUBMITTED`, `CHECKED`, `VERIFIED`, `APPROVED`, `REJECTED`.

### 2.2. Payment Notes (PN) — Financial Disbursement & Guardrails
- **Calculation Formula**: `Net Payable Amount` = `Gross Invoice Amount` - `TDS Deductions` - `Retention / Other Withholdings`.
- **Guardrails**: Strict database-level and API-level checks prevent negative disbursements or gross mismatch.
- **Lifecycle**: `ACTIVE` (scheduled) -> `HOLD` (dispute or compliance freeze) -> `PAID` (disbursed via RTGS/NEFT with registered UTR number).

### 2.3. Vendor Master & Multi-Bank Management
- **Vendor Master** (`vendors` table): Stores corporate details, GSTIN, PAN, MSME classification (`Micro`, `Small`, `Medium`, `Corporate EPC`), Section 206AB compliance (TDS non-defaulter flag).
- **Multi-Bank Accounts** (`vendor_accounts` table):
  - Supports multiple disbursement accounts per vendor (Primary Disbursing vs. Secondary / Project Escrow).
  - Single primary account constraint: marking an account as primary atomically demotes previous primaries.
  - Endpoints: `GET /api/v1/vendors/{vendor_id}/accounts` and `POST /api/v1/vendors/{vendor_id}/accounts`.
  - Fallback logic: If no secondary records exist in `vendor_accounts`, gateway synthesizes primary account from `vendors` KYC record.

### 2.4. Reimbursement Notes
- Travel, fuel, and ad-hoc operational claims submitted by highway inspection teams.
- Stored with expense breakdown line-items and bank routing details.

### 2.5. RBAC & Security (OWASP Top 10 Hardened)
- Centralized Bearer JWT authentication on all `/api/v1/*` routes.
- Two-Factor Authentication (2FA) with OTP verification stored in Redis / distributed memory.
- Multi-tenancy with `tenant_id` and `org_id` isolation.
- Input validation: All UUID path parameters pass through `validate_uuid()` to prevent PostgreSQL 500 crashes.

---

## 3. Developer & Agent Runbook

### Standard Services & Ports
| Component | Runtime | Port | Directory | Start Command |
|---|---|---|---|---|
| API Gateway | Python | `8000` | `NHIT Backend` | `python "NHIT Backend/gateway_swagger.py"` |
| Frontend SPA | Vite / React | `5173` | `nhit-frontend` | `npm run dev` (in `nhit-frontend`) |
| PostgreSQL | Postgres 15 | `5433` | Docker/Host | `docker-compose up -d postgres` |
| Redis | Redis 7 | `6379` | Docker/Host | `docker-compose up -d redis` |

### Database Configuration (Default Development)
```json
{
  "host": "localhost",
  "port": 5433,
  "dbname": "nhit_db",
  "user": "postgres",
  "password": "shristi"
}
```

### Essential Verification Commands
1. **Frontend Typecheck**:
   ```bash
   cd nhit-frontend
   npx tsc --noEmit
   ```
2. **Backend Syntax & Compile Check**:
   ```bash
   python -m py_compile "NHIT Backend/gateway_swagger.py"
   ```
3. **Execute Full Health & Test Harness (103+ Invariant Checks)**:
   ```bash
   python "NHIT Backend/scripts/harness.py"
   ```
   - Fast mode: `python "NHIT Backend/scripts/harness.py" --fast`
   - Specific phase: `python "NHIT Backend/scripts/harness.py" --phase 1` (1-5)
   - Fail-fast mode: `python "NHIT Backend/scripts/harness.py" --bail`
   - JSON report: `python "NHIT Backend/scripts/harness.py" --json`

---

## 4. Known Gotchas & Engineering Conventions

1. **Windows PowerShell Chaining**:
   - Never use `&&` to chain commands in PowerShell; it causes a syntax error. Use `;` or separate executions.
2. **Windows Console Unicode**:
   - The default Windows console encoding (`cp1252`) fails when printing emojis. Any Python script printing logs should include:
     ```python
     import sys
     if sys.platform == "win32":
         sys.stdout.reconfigure(encoding="utf-8")
     ```
3. **Optional Redis Dependency**:
   - `redis` is optional for distributed caching. In `gateway_swagger.py`, `import redis` must always be annotated with `# type: ignore` to prevent static type checker errors in environments where redis is not pre-installed.
4. **TypeScript Strictness**:
   - Zero `any` usage. All API responses must map to interfaces declared in `nhit-frontend/src/types/index.ts`.
5. **Contract Synchronization**:
   - Whenever modifying backend route models in `gateway_swagger.py`, immediately update:
     - `nhit-frontend/src/types/index.ts`
     - `nhit-frontend/src/services/api.ts`
6. **UUID Validation**:
   - Always call `validate_uuid(id, "param_name")` in gateway route handlers before passing UUIDs to SQL queries to avoid PostgreSQL syntax errors.
