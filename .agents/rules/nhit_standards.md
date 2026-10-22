# NHIT Platform Engineering Standards

## 1. Security & Input Sanitization
- All UUID URL path parameters (`vendor_id`, `project_id`, `note_id`, `user_id`) must pass through `validate_uuid()` before executing raw SQL queries in PostgreSQL to avoid unhandled database crashes (500 Internal Server Errors).
- All SQL statements must use parameterized placeholders (`%s`) to prevent SQL injection.
- Endpoints modifying financial state (Green Note approvals, Payment Note status transitions, Vendor account creation) must audit `created_by` or `updated_by`.

## 2. Financial Integrity Guardrails
- In Payment Notes, the calculated `net_payable_amount` must never be negative (`gross_amount >= tds_amount`).
- Bank account IFSC codes must strictly match the regex `^[A-Z]{4}0[A-Z0-9]{6}$`.
- Vendor PAN numbers must strictly match the regex `^[A-Z]{5}[0-9]{4}[A-Z]{1}$`.

## 3. Frontend TypeScript Integrity
- No `any` types permitted in `nhit-frontend/src`.
- When adding or altering backend endpoints, update `nhit-frontend/src/types/index.ts` and `nhit-frontend/src/services/api.ts` immediately.
- Run `npx tsc --noEmit` before concluding any feature to ensure zero compile warnings.

## 4. Multi-Bank Account Invariants
- Each vendor must have at most one account with `is_primary = true`.
- When an account is marked primary, previous primary accounts for that vendor must be demoted.
- If no secondary records exist in `vendor_accounts`, APIs must fall back gracefully to the primary bank record in the `vendors` table.
