# Report E: Endpoint Inventory and API Security Matrix

> **National Highways Infra Trust (NHIT) Digital Platform**  
> **API Gateway:** Python 3.12 / FastAPI (`NHIT Backend/gateway_swagger.py`)  
> **Total Cataloged Endpoints:** 135 across 13 Service Domains  
> **Standard:** OWASP API Security Top 10:2023  
> **Date:** October 24, 2026

---

## 1. API Security Overview by Service Domain

| Service Domain | Total Endpoints | Auth Enforced | Missing Auth | Rate Limited | High/Critical Risks Identified |
|---|:---:|:---:|:---:|:---:|---|
| **🔐 Auth Service** | 17 | 3 (Internal) | 0 (Public) | 2 of 17 | 2FA bypass via change-password (`SEC-009`), no logout invalidation (`SEC-008`), backdoor fallback (`SEC-007`) |
| **👥 User Management** | 28 | 28 | 0 | 0 of 28 | Lacks role restriction (regular users can manage admins), BOLA/IDOR risk |
| **🏢 Organization Service** | 10 | 7 | 3 | 0 of 10 | Public read access to org vendors, projects, and departments |
| **🏛️ Department Service** | 5 | 5 | 0 | 0 of 5 | Lacks rate limiting |
| **🏷️ Designation Service** | 5 | 5 | 0 | 0 of 5 | Lacks rate limiting |
| **📁 Project Service** | 3 | 3 | 0 | 0 of 3 | Lacks rate limiting |
| **🤝 Vendor Service & Accounts** | 15 | 0 | **15** | 0 of 15 | **Unauthenticated bank injection & primary account hijacking** (`SEC-002`) |
| **🔔 Notification Service** | 2 | 0 | **2** | 0 of 2 | **Unauthenticated live OTP dispatch & notification flooding** (`SEC-003`) |
| **📝 GreenNote Service** | 9 | 0 | **9** | 0 of 9 | **Arbitrary approval skipping** (`SEC-014`), in-memory data loss (`SEC-011`) |
| **💰 PaymentNote Service** | 14 | 0 | **14** | 0 of 14 | **Client-controlled disbursement amount** (`SEC-015`), in-memory storage (`SEC-011`) |
| **💳 Payment Service** | 18 | 0 | **18** | 0 of 18 | Unauthenticated payment creation, cart tampering, batch execution |
| **🧾 Reimbursement Service** | 16 | 0 | **16** | 0 of 16 | Unauthenticated claim creation, approval, and PII disclosure |
| **✅ Approval Service** | 1 | 0 | **1** | 0 of 1 | Unauthenticated approval workflow initiation |
| **System & Health** | 2 | 0 | 2 (Public) | 0 of 2 | Exposed interactive Swagger UI and OpenAPI JSON specs (`SEC-005`) |
| **TOTALS** | **145** | **51** | **88 Sensitive** | **2** | **Widespread Missing Auth & Rate Limiting** |

---

## 2. Detailed Service Endpoint Inventory

### 2.1. 🔐 Authentication Service (17 Endpoints)
| Method | Endpoint Path | Function Handler | Auth Type | Input Validation | Rate Limit | Risk Rating |
|---|---|---|---|---|---|:---:|
| `POST` | `/api/v1/auth/login` | `login` | Public | Pydantic Schema | Yes (10/min) | High (`SEC-007`) |
| `POST` | `/api/v1/auth/verify-otp` | `verify_otp` | Temp Token | Pydantic Schema | Yes (5/min) | Medium |
| `POST` | `/api/v1/auth/passwordless/request-otp` | `passwordless_request_otp` | Public | Pydantic Schema | None | High (`SEC-003`) |
| `POST` | `/api/v1/auth/sso/google` | `sso_google` | Public | Google ID Token | None | Medium |
| `GET` | `/api/v1/auth/sso/google/initiate` | `sso_google_initiate` | Public | None | None | Low |
| `GET` | `/api/v1/auth/sso/google/callback` | `sso_google_callback` | Public | OAuth Code | None | Low |
| `POST` | `/api/v1/auth/sso/logout/initiate` | `sso_logout_initiate` | Public | None | None | Low |
| `POST` | `/api/v1/auth/sso/logout/complete` | `sso_logout_complete` | Public | None | None | Low |
| `POST` | `/api/v1/auth/register` | `register` | Public | Pydantic Schema | None | Medium |
| `POST` | `/api/v1/auth/verify-email` | `verify_email` | Public | Pydantic Schema | None | Low |
| `POST` | `/api/v1/auth/forgot-password` | `forgot_password` | Public | Pydantic Schema | None | Low |
| `POST` | `/api/v1/auth/reset-password` | `reset_password` | Public | Pydantic Schema | None | Low |
| `POST` | `/api/v1/auth/change-password` | `change_password` | Bearer / Temp Token | Pydantic Schema | None | **Critical (`SEC-009`)** |
| `POST` | `/api/v1/auth/forgot-password-otp` | `forgot_password_otp` | Public | Pydantic Schema | None | Medium (`SEC-008`) |
| `POST` | `/api/v1/auth/verify-otp-reset-password` | `verify_otp_reset` | Public | Pydantic Schema | None | Medium (`SEC-008`) |
| `POST` | `/api/v1/auth/logout` | `logout` | Bearer Token | Pydantic Schema | None | Medium (`SEC-008`) |
| `POST` | `/api/v1/auth/refresh` | `refresh_token` | Refresh Token | Pydantic Schema | None | Low |

---

### 2.2. 🤝 Vendor Service & Multi-Bank Management (15 Endpoints)
| Method | Endpoint Path | Function Handler | Auth Type | Input Validation | Rate Limit | Risk Rating |
|---|---|---|---|---|---|:---:|
| `POST` | `/api/v1/vendors` | `create_vendor` | ❌ None | Pydantic Schema | None | High (`SEC-016`) |
| `GET` | `/api/v1/vendors` | `list_vendors` | ❌ None | Query Params | None | High |
| `GET` | `/api/v1/vendors/{vendor_id}` | `get_vendor` | ❌ None | UUID Validator | None | Medium |
| `GET` | `/api/v1/vendors/code/{vendor_code}` | `get_vendor_by_code` | ❌ None | String Path | None | Medium |
| `PUT` | `/api/v1/vendors/{vendor_id}` | `update_vendor` | ❌ None | UUID Validator | None | High |
| `DELETE` | `/api/v1/vendors/{vendor_id}` | `delete_vendor` | ❌ None | UUID Validator | None | High |
| `POST` | `/api/v1/vendors/generate-code` | `generate_vendor_code` | ❌ None | Pydantic Schema | None | Low |
| `PUT` | `/api/v1/vendors/{vendor_id}/code` | `update_vendor_code` | ❌ None | UUID Validator | None | Medium |
| `POST` | `/api/v1/vendors/{vendor_id}/regenerate-code` | `regenerate_vendor_code` | ❌ None | UUID Validator | None | Medium |
| `POST` | `/api/v1/vendors/{vendor_id}/accounts` | `create_vendor_account` | ❌ None | UUID + Pydantic | None | **Critical (`SEC-002`)** |
| `GET` | `/api/v1/vendors/{vendor_id}/accounts` | `get_vendor_accounts` | ❌ None | UUID Validator | None | High |
| `GET` | `/api/v1/vendors/{vendor_id}/banking-details` | `get_vendor_banking` | ❌ None | UUID Validator | None | High |
| `PUT` | `/api/v1/vendors/accounts/{account_id}` | `update_vendor_account` | ❌ None | String ID | None | High |
| `DELETE` | `/api/v1/vendors/accounts/{account_id}` | `delete_vendor_account` | ❌ None | String ID | None | High |
| `POST` | `/api/v1/vendors/accounts/{account_id}/toggle-status` | `toggle_account_status` | ❌ None | String ID | None | High |

---

### 2.3. 📝 GreenNote Financial Approval Service (9 Endpoints)
| Method | Endpoint Path | Function Handler | Auth Type | Input Validation | Rate Limit | Risk Rating |
|---|---|---|---|---|---|:---:|
| `POST` | `/api/v1/green-notes` | `create_green_note` | ❌ None | Pydantic Schema | None | High (`SEC-016`) |
| `GET` | `/api/v1/green-notes` | `list_green_notes` | ❌ None | Query Filters | None | Medium |
| `GET` | `/api/v1/green-notes/{id}` | `get_green_note` | ❌ None | String ID | None | Medium |
| `PATCH` | `/api/v1/green-notes/{id}` | `update_green_note` | ❌ None | Pydantic Schema | None | High |
| `PATCH` | `/api/v1/green-notes/{id}/status` | `update_green_note_status` | ❌ None | Enum Whitelist | None | **High (`SEC-014`)** |
| `POST` | `/api/v1/green-notes/{id}/cancel` | `cancel_green_note` | ❌ None | Pydantic Schema | None | High |
| `POST` | `/api/v1/green-notes/{note_id}/documents` | `upload_green_note_docs` | ❌ None | Form Data | None | Medium |
| `POST` | `/api/v1/green-notes/{id}/link-payment` | `link_payment_note` | ❌ None | Pydantic Schema | None | High |
| `GET` | `/api/v1/green-notes/history` | `green_note_history` | ❌ None | Query Filters | None | Medium |

---

### 2.4. 💰 PaymentNote Financial Disbursement Service (14 Endpoints)
| Method | Endpoint Path | Function Handler | Auth Type | Input Validation | Rate Limit | Risk Rating |
|---|---|---|---|---|---|:---:|
| `GET` | `/api/v1/payment-notes` | `list_payment_notes` | ❌ None | Query Filters | None | High |
| `GET` | `/api/v1/payment-notes/drafts` | `list_draft_payment_notes` | ❌ None | Query Filters | None | Medium |
| `GET` | `/api/v1/payment-notes/{id}` | `get_payment_note` | ❌ None | Int ID | None | Medium |
| `POST` | `/api/v1/payment-notes` | `create_payment_note` | ❌ None | Pydantic Schema | None | **High (`SEC-015`)** |
| `PATCH` | `/api/v1/payment-notes/{id}` | `update_payment_note` | ❌ None | Int ID | None | High |
| `DELETE` | `/api/v1/payment-notes/{id}` | `delete_payment_note` | ❌ None | Int ID | None | High |
| `POST` | `/api/v1/payment-notes/draft` | `create_draft_from_green_note` | ❌ None | Pydantic Schema | None | High |
| `POST` | `/api/v1/payment-notes/{id}/convert` | `convert_draft_to_active` | ❌ None | Int ID | None | High |
| `DELETE` | `/api/v1/payment-notes/{id}/draft` | `delete_draft_payment_note` | ❌ None | Int ID | None | High |
| `POST` | `/api/v1/payment-notes/{id}/hold` | `put_on_hold` | ❌ None | Pydantic Schema | None | High |
| `POST` | `/api/v1/payment-notes/{id}/unhold` | `remove_from_hold` | ❌ None | Pydantic Schema | None | High |
| `POST` | `/api/v1/payment-notes/{id}/utr` | `update_utr` | ❌ None | Pydantic Schema | None | High |
| `PATCH` | `/api/v1/payment-notes/{id}/status` | `update_payment_note_status`| ❌ None | Pydantic Schema | None | High |
| `GET` | `/api/v1/payment-notes/{id}/pdf` | `download_payment_note_pdf` | ❌ None | Int ID | None | Medium |

---

### 2.5. 💳 Payment Processing & Bank Letter Service (18 Endpoints)
| Method | Endpoint Path | Function Handler | Auth Type | Input Validation | Rate Limit | Risk Rating |
|---|---|---|---|---|---|:---:|
| `GET` | `/api/v1/payments` | `list_payments` | ❌ None | Query Filters | None | High |
| `GET` | `/api/v1/payments/datatable` | `list_payments_datatable` | ❌ None | Query Filters | None | High |
| `GET` | `/api/v1/payments/{sl_no}` | `get_payment_group` | ❌ None | String Path | None | Medium |
| `POST` | `/api/v1/payments` | `create_payment_requests` | ❌ None | Pydantic Schema | None | High |
| `PATCH` | `/api/v1/payments/{sl_no}` | `update_payment_group` | ❌ None | String Path | None | High |
| `DELETE` | `/api/v1/payments/{id}` | `delete_payment` | ❌ None | String Path | None | High |
| `DELETE` | `/api/v1/payments/{sl_no}/items/{id}`| `delete_payment_item` | ❌ None | String Path | None | High |
| `POST` | `/api/v1/payments/search-vendors` | `search_vendors_payment` | ❌ None | Pydantic Schema | None | Low |
| `GET` | `/api/v1/payments/vendors/internal` | `search_internal_vendors` | ❌ None | Query Params | None | Low |
| `GET` | `/api/v1/payments/projects` | `search_projects_payment` | ❌ None | Query Params | None | Low |
| `GET` | `/api/v1/payments/from-accounts/{type}` | `get_from_account_options` | ❌ None | String Path | None | Medium |
| `GET` | `/api/v1/payments/vendors` | `get_all_vendors_payment` | ❌ None | Query Params | None | Medium |
| `POST` | `/api/v1/payments/cart` | `add_to_cart` | ❌ None | Pydantic Schema | None | Medium |
| `DELETE` | `/api/v1/payments/cart` | `remove_from_cart` | ❌ None | Query Params | None | Medium |
| `GET` | `/api/v1/payments/cart` | `get_cart` | ❌ None | None | None | Medium |
| `POST` | `/api/v1/payments/shortcuts` | `create_shortcut` | ❌ None | Pydantic Schema | None | Medium |
| `POST` | `/api/v1/payments/shortcuts/{id}/execute` | `execute_shortcut` | ❌ None | String ID | None | High |
| `POST` | `/api/v1/payments/bank-letters` | `create_bank_letter` | ❌ None | Pydantic Schema | None | High |

---

### 2.6. 🧾 Reimbursement Note Service (16 Endpoints)
| Method | Endpoint Path | Function Handler | Auth Type | Input Validation | Rate Limit | Risk Rating |
|---|---|---|---|---|---|:---:|
| `GET` | `/api/v1/reimbursement-notes` | `list_reimbursement_notes` | ❌ None | Query Filters | None | High |
| `GET` | `/api/v1/reimbursements` | `list_reimbursements_alias`| ❌ None | Query Filters | None | High |
| `GET` | `/api/v1/reimbursement-notes/datatable` | `list_reimbursement_datatable`| ❌ None | Query Filters | None | High |
| `GET` | `/api/v1/reimbursement-notes/{id}` | `get_reimbursement_note` | ❌ None | String ID | None | Medium |
| `GET` | `/api/v1/reimbursement-notes/create` | `get_reimbursement_create_data`| ❌ None | None | None | Low |
| `POST` | `/api/v1/reimbursement-notes` | `create_reimbursement_note`| ❌ None | Pydantic Schema | None | High |
| `POST` | `/api/v1/reimbursements` | `create_reimbursements_alias`| ❌ None | Pydantic Schema | None | High |
| `GET` | `/api/v1/reimbursement-notes/{id}/edit`| `get_reimbursement_edit_data`| ❌ None | String ID | None | Medium |
| `PATCH` | `/api/v1/reimbursement-notes/{id}` | `update_reimbursement_note`| ❌ None | String ID | None | High |
| `DELETE` | `/api/v1/reimbursement-notes/{id}` | `delete_reimbursement_note`| ❌ None | String ID | None | High |
| `POST` | `/api/v1/reimbursement-notes/{id}/approval-log` | `update_reimbursement_approval_log`| ❌ None | Form Data | None | High |
| `GET` | `/api/v1/reimbursement-notes/{id}/pdf` | `download_reimbursement_pdf`| ❌ None | String ID | None | Medium |
| `GET` | `/api/v1/reimbursement-notes/{id}/pdf/preview` | `preview_reimbursement_pdf`| ❌ None | String ID | None | Medium |
| `GET` | `/api/v1/reimbursement-notes/{id}/payment-note` | `get_reimbursement_payment_data`| ❌ None | String ID | None | Medium |
| `GET` | `/api/v1/reimbursement-notes/users` | `list_selectable_users` | ❌ None | None | None | High (PII) |
| `POST` | `/api/v1/reimbursement-notes/users/select` | `select_user_for_reimbursement`| ❌ None | Pydantic Schema | None | Medium |

---

### 2.7. 👥 User Management & Master Setup (28 Endpoints)
All 28 user management routes (`/api/v1/users/*`, `/api/v1/roles/*`, `/api/v1/tenants/*`) declare `current_user: dict = Depends(get_current_user)`. However, **all 28 routes lack specific role validation checks** (e.g., verifying `current_user["roles"]` contains `"super_admin"` or `"admin"`). Any authenticated user with a basic `employee` token can list all users, create new administrator accounts, or reset other accounts.
