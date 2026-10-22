---
name: nhit-harness
description: >-
  Use this skill to run the NHIT test harness, verify database connectivity, validate frontend and backend health, and execute platform verification checks.
---

# NHIT Platform Verification & Test Harness

This skill provides step-by-step instructions to verify the integrity of the NHIT Digital Platform across backend services, database migrations, and frontend compilation.

## 1. Automated Verification Harness
Run the unified platform test harness script from the workspace root:
```bash
python "NHIT Backend/scripts/harness.py"
```
The harness performs:
1. Python environment validation & dependency check
2. Gateway code compilation & syntax check (`gateway_swagger.py`)
3. PostgreSQL connection check (port `5433`, database `nhit_db`)
4. Core database schema tables verification (`vendors`, `vendor_accounts`, `projects`, `green_notes`, `payment_notes`, `users`)
5. Frontend TypeScript compile check (`npx tsc --noEmit`)
6. Financial calculation integrity verification

## 2. Independent Backend Check
```bash
python -m py_compile "NHIT Backend/gateway_swagger.py"
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
