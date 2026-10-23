# National Highways Infra Trust (NHIT) — 100% Robust Test Harness

> **Enterprise Automated Assurance, Financial Invariant & Contract Enforcement Engine**  
> **Repository:** `HarshitNain/NHIT-Digital-Platform`  
> **Engine:** 7 Verification Phases | 144+ Invariant Checks | Zero-Flake Guarantee

---

## 1. Overview & Architecture

The NHIT Verification Harness is an enterprise-grade automated testing and invariant-checking framework. It guarantees that the entire NHIT Digital Platform—encompassing the Python/FastAPI Gateway, PostgreSQL schema definitions, Go microservices, gRPC protobuf stubs, React/TypeScript SPA, and statutory compliance rules—strictly conforms to zero-regression architectural invariants.

```
+-----------------------------------------------------------------------------------------+
|                               NHIT VERIFICATION HARNESS                                 |
+-----------------------------------------------------------------------------------------+
  |-- Phase 1: Runtime & Dependency Matrix (Python 3.10+, Node 20+, HTTPX, Redis, Git)
  |-- Phase 2: Static Code & TypeScript Contracts (tsc -b, 10 types, 11 API methods, 11 modals)
  |-- Phase 3: PostgreSQL Schema & Triggers (14 DDL tables, IFSC/PAN regex, primary account trigger)
  |-- Phase 4: Dynamic In-Process FastAPI & OWASP Security (UUID guards, SQLi fuzzing, 422 schema)
  |-- Phase 5: Financial, Cryptographic & Statutory Invariants (HMAC, 2FA OTP, 429 rate limit, GSTIN)
  |-- Phase 6: Microservices & gRPC Protobuf Contracts (8 Go modules, 5 proto schemas, 5 stubs)
  +-- Phase 7: Multi-Tier Business Workflow Simulations (4-tier GN, PN net formula, RTGS/NEFT UTR)
```

---

## 2. Directory Structure

```
harness/
├── configs/
│   ├── harness_config.yaml      # Master configuration (ports, timeouts, statutory rules)
│   └── test.env                 # Environment variables template for isolated test runs
├── reports/
│   ├── harness_report.html      # Standalone interactive dashboard (inline CSS, zero CDN)
│   ├── harness_report.md        # GitHub-flavored Markdown execution summary
│   └── .gitkeep
├── run-harness.ps1              # PowerShell runner (ANSI color headers, parameter passthrough)
├── run-harness.sh               # POSIX Bash runner for Linux/macOS environments
└── README.md                    # Comprehensive documentation and runbook
```

Root-level convenience shortcuts are also available:
- `.\run-harness.ps1`
- `.\run-harness.bat`

---

## 3. The 7 Verification Phases

| Phase | Domain | Check Count | Key Invariants Enforced |
|---|---|---|---|
| **Phase 1** | Runtime & Dependency Matrix | 14 | Python 3.10+, FastAPI, Uvicorn, Pydantic, PyJWT, Psycopg2, HTTPX, Redis, Node.js, Git |
| **Phase 2** | Code Compilation & TypeScript Contracts | 18 | `tsc -b` strict typecheck, 10 interface models (`Vendor`, `VendorAccount`, `GreenNote`), 11 API methods, 11 modals |
| **Phase 3** | Database Schema & Migrations | 19 | 14 DDL tables, IFSC/Account number constraints, single primary bank account trigger, live DB connectivity |
| **Phase 4** | In-Process FastAPI & OWASP Security | 23 | OpenAPI 3.0 schema, `/docs`, UUID path parameter sanitization (`validate_uuid`), SQLi fuzzing, negative value guards |
| **Phase 5** | Financial & Statutory Invariants | 16 | HMAC-SHA256 tokens, 2FA OTP lifecycle & thread safety, sliding-window rate limiting (HTTP 429), PAN/GSTIN/IFSC regex, Section 206AB TDS |
| **Phase 6** | Microservices & gRPC Protobuf | 21 | 5 Proto definitions (`Auth`, `GreenNote`, `PaymentNote`, `Payment`, `Approval`), 5 `pb.go` stubs, 8 Go modules, Nhit-Note root |
| **Phase 7** | Multi-Tier Workflow Simulations | 5 | 4-tier Green Note state machine, Payment Note net formula (`gross - TDS - retention`), multi-bank fallback account synthesis, RTGS/NEFT UTR format |

---

## 4. Quickstart & Command-Line Usage

### PowerShell (Windows)
```powershell
# Run full verification suite (generates HTML and Markdown reports)
.\run-harness.ps1

# Fast mode (skips full frontend TypeScript compile, ~18s)
.\run-harness.ps1 -Fast

# Run only a specific phase (1 to 7)
.\run-harness.ps1 -Phase 7

# Fail-fast mode (halts on first critical failure)
.\run-harness.ps1 -Fast -Bail

# Emit machine-readable JSON for automated tooling
.\run-harness.ps1 -Json
```

### Windows Batch (CMD)
```cmd
run-harness.bat -Fast
run-harness.bat -Phase 4
```

### Bash (Linux / macOS / CI)
```bash
# Ensure executable permissions
chmod +x ./harness/run-harness.sh

# Run full harness
./harness/run-harness.sh --fast

# Target specific phase
./harness/run-harness.sh --phase 5
```

---

## 5. Statutory & Enterprise Financial Rules Enforced

1. **4-Tier Green Note Financial Approval Progression**:
   - `DRAFT` / `SUBMITTED` (Project Engineer)
   - `CHECKED` (Dy. General Manager)
   - `VERIFIED` (Financial Controller)
   - `APPROVED` (CGM / Trustee Board)
   - Strict whitelist blocks unauthorized status transitions with HTTP 422.

2. **Payment Note Net Payable Guard**:
   - `net_payable_amount = gross_invoice_amount - tds_deductions - retention_amount`
   - Zero tolerance for negative disbursements (`gross_amount >= tds_amount`).

3. **Vendor Multi-Bank Account Exclusivity**:
   - One primary disbursement account per vendor.
   - Setting `is_primary = true` atomically demotes previous primary accounts.
   - Dynamic fallback synthesis: if no records exist in `vendor_accounts`, primary account is synthesized from the vendor KYC master.

4. **Statutory Tax Identifiers**:
   - **PAN**: `^[A-Z]{5}[0-9]{4}[A-Z]$` (Income Tax Act Section 139A).
   - **GSTIN**: `^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z][1-9A-Z]Z[0-9A-Z]$` (15-character standard).
   - **RBI IFSC**: `^[A-Z]{4}0[A-Z0-9]{6}$` (RTGS/NEFT routing format with 5th character mandatory zero).
   - **UTR Identifier**: 16-character alphanumeric sequence with valid clearing prefix (`UTIB`, `SBIN`, `HDFC`, `ICIC`).

---

## 6. Continuous Integration (CI/CD)

The harness is fully integrated into `.github/workflows/ci.yml`. On every pull request and push to `main`:
1. Services are launched with PostgreSQL and live Redis.
2. The harness executes with `--fast --html harness/reports/harness_report.html --markdown harness/reports/harness_report.md`.
3. The generated markdown summary is automatically appended to `$GITHUB_STEP_SUMMARY`, rendering an interactive executive dashboard in GitHub Actions.
4. If any invariant check fails, the pipeline halts immediately, preventing broken builds from reaching staging or production.

---

## 7. Artifacts & Reporting

Every harness run automatically produces:
- **`harness/reports/harness_report.html`**: Self-contained, responsive dashboard featuring dark/light aesthetic, phase breakdown cards, pass/fail badges, and duration meters. Zero external network calls or CDNs required.
- **`harness/reports/harness_report.md`**: Clean GitHub-flavored Markdown table with check descriptions, execution timestamps, and invariant assertions.
