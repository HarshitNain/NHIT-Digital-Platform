# NHIT Digital Platform — Root Agent Guidelines & Memory Harness

<!-- Mirrors AGENTS.md for Antigravity & Gemini Agent Engines -->
Please refer to [AGENTS.md](./AGENTS.md) for the complete architecture, domain workflows, runbook, and engineering conventions.

## Quick Reference Summary
- **Gateway**: FastAPI on port `8000` (`NHIT Backend/gateway_swagger.py`).
- **Frontend**: React 18 / 19 + TypeScript on port `5173` (`nhit-frontend/`).
- **Database**: PostgreSQL 15 on port `5433` (DB: `nhit_db`, User: `postgres`, Password: `shristi`).
- **Verification Harness**: `python "NHIT Backend/scripts/harness.py"` (116+ checks, `--fast`, `--bail`, `--phase`).
- **TypeScript Check**: `cd nhit-frontend && npx tsc --noEmit`.

## Key Rules
1. Never chain PowerShell commands with `&&` (use `;`).
2. Always validate UUID parameters with `validate_uuid` before running SQL queries.
3. Keep frontend `types/index.ts` synchronized with backend API models.
4. Support primary fallback in all vendor multi-bank and banking queries.
5. Prevent negative disbursements in Payment Notes (`gross_amount >= tds_amount`).
