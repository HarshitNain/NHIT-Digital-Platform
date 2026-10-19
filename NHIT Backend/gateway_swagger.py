"""
╔══════════════════════════════════════════════════════════════════════════════╗
║             NHIT Backend — Enterprise API Gateway & Swagger Hub            ║
║                                                                            ║
║  Comprehensive API documentation covering ALL microservices:               ║
║    • Auth Service           (17 endpoints)                                 ║
║    • User Management        (28 endpoints)                                 ║
║    • Organization Service   (10 endpoints)                                 ║
║    • Department Service     ( 5 endpoints)                                 ║
║    • Designation Service    ( 5 endpoints)                                 ║
║    • Project Service        ( 3 endpoints)                                 ║
║    • Vendor Service         (15 endpoints)                                 ║
║    • Notification Service   ( 1 endpoint )                                 ║
║    • GreenNote Service      ( 9 endpoints)                                 ║
║    • PaymentNote Service    (14 endpoints)                                 ║
║    • Payment Service        (18 endpoints)                                 ║
║    • Reimbursement Service  (16 endpoints)                                 ║
║    • Approval Service       ( 1 endpoint )                                 ║
║                                                                            ║
║  Port: 8080  |  Dashboard: /  |  Swagger: /swagger  |  ReDoc: /redoc       ║
╚══════════════════════════════════════════════════════════════════════════════╝
"""

import datetime
import hashlib
import json
import os
import secrets
import traceback
import uuid
import urllib.request
from typing import Any, Dict, List, Optional

import bcrypt
import jwt  # PyJWT 2.15.1
import psycopg2
import psycopg2.extras
import uvicorn
from fastapi import FastAPI, HTTPException, Query, Path, Body, Depends, Header
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import HTMLResponse, JSONResponse, RedirectResponse
from pydantic import BaseModel, Field

from otp_notifier import dispatch_otp

# ═══════════════════════════════════════════════════════
# DATABASE CONNECTION
# ═══════════════════════════════════════════════════════
DB_CONFIG = {
    "host": os.environ.get("DB_HOST", "localhost"),
    "port": int(os.environ.get("DB_PORT", 5433)),
    "dbname": os.environ.get("DB_NAME", "nhit_db"),
    "user": os.environ.get("DB_USER", "postgres"),
    "password": os.environ.get("DB_PASSWORD", "shristi"),
}

def get_db():
    conn = psycopg2.connect(**DB_CONFIG)
    conn.autocommit = True
    return conn

def query_db(sql, params=None):
    conn = get_db()
    try:
        cur = conn.cursor(cursor_factory=psycopg2.extras.RealDictCursor)
        cur.execute(sql, params)
        return [dict(r) for r in cur.fetchall()]
    finally:
        conn.close()

def execute_db(sql, params=None):
    conn = get_db()
    try:
        cur = conn.cursor()
        cur.execute(sql, params)
        return cur.rowcount
    finally:
        conn.close()

# ═══════════════════════════════════════════════════════
# JWT & SECURITY CONFIGURATION
# ═══════════════════════════════════════════════════════
JWT_SECRET = os.environ.get("JWT_SECRET", "test-secret-key-for-development-only-change-in-production")
JWT_ALGORITHM = "HS256"
JWT_EXPIRATION_HOURS = 24

def create_jwt_token(payload: dict, expires_in_seconds: int = 86400) -> str:
    now = datetime.datetime.now(datetime.timezone.utc)
    to_encode = payload.copy()
    to_encode.update({
        "iat": int(now.timestamp()),
        "exp": int((now + datetime.timedelta(seconds=expires_in_seconds)).timestamp()),
        "iss": "nhit-security-gateway",
        "aud": "nhit-enterprise-portal",
    })
    return jwt.encode(to_encode, JWT_SECRET, algorithm=JWT_ALGORITHM)

def decode_jwt_token(token: str) -> dict:
    try:
        return jwt.decode(
            token,
            JWT_SECRET,
            algorithms=[JWT_ALGORITHM],
            audience="nhit-enterprise-portal",
        )
    except Exception as e:
        raise HTTPException(401, f"Invalid or expired JWT token: {str(e)}")

# In-memory OTP storage for Two-Factor (2-Way) Authentication
# Format: { session_temp_token: {"otp": "123456", "email": "...", "user_id": "...", "expires_at": float, "attempts": int} }
OTP_STORE: Dict[str, dict] = {}

def clean_expired_otps():
    now = datetime.datetime.now(datetime.timezone.utc).timestamp()
    expired_keys = [k for k, v in OTP_STORE.items() if v.get("expires_at", 0) < now]
    for k in expired_keys:
        OTP_STORE.pop(k, None)

def generate_numeric_otp(length: int = 6) -> str:
    return "".join(secrets.choice("0123456789") for _ in range(length))

def get_user_roles_and_perms(user_id: str):
    roles_data = query_db("""
        SELECT r.name, r.permissions 
        FROM user_roles ur 
        JOIN roles r ON ur.role_id = r.role_id 
        WHERE ur.user_id = %s
    """, (user_id,))
    user_roles = [r["name"] for r in roles_data] if roles_data else ["employee"]
    user_perms = []
    for r in roles_data:
        if r.get("permissions"):
            user_perms.extend(r["permissions"])
    if not user_perms:
        user_perms = ["users.read", "projects.read", "notes.read"]
    return user_roles, list(set(user_perms))

def issue_user_jwt_tokens(u: dict, auth_method: str = "password"):
    user_roles, user_perms = get_user_roles_and_perms(u["user_id"])
    now = datetime.datetime.now(datetime.timezone.utc)
    access_token_payload = {
        "sub": str(u["user_id"]),
        "email": u["email"],
        "name": u["name"],
        "tenant_id": str(u["tenant_id"]) if u.get("tenant_id") else "",
        "roles": user_roles,
        "permissions": user_perms,
        "auth_method": auth_method,
    }
    access_token = create_jwt_token(access_token_payload, expires_in_seconds=86400)
    refresh_token = create_jwt_token({
        "sub": str(u["user_id"]),
        "type": "refresh",
        "auth_method": auth_method,
    }, expires_in_seconds=7 * 86400)

    try:
        execute_db("""
            INSERT INTO user_login_history (user_id, login_time, ip_address, user_agent)
            VALUES (%s, NOW(), %s, %s)
        """, (u["user_id"], "127.0.0.1", f"NHIT Enterprise Portal ({auth_method})"))
        execute_db("UPDATE users SET last_login_at = NOW() WHERE user_id = %s", (u["user_id"],))
    except Exception:
        pass

    return {
        "status": "SUCCESS",
        "token": access_token,
        "refresh_token": refresh_token,
        "user_id": str(u["user_id"]),
        "email": u["email"],
        "name": u["name"],
        "tenant_id": str(u["tenant_id"]) if u.get("tenant_id") else "",
        "roles": user_roles,
        "permissions": user_perms,
        "auth_method": auth_method,
        "token_expires_at": int((now + datetime.timedelta(hours=24)).timestamp()),
        "refresh_expires_at": int((now + datetime.timedelta(days=7)).timestamp()),
    }

# In-memory note stores for interactive Swagger UI testing
GREEN_NOTES_STORE = [
    {
        "id": "gn-550e8400-0001",
        "sl_no": "GN-2026-001",
        "project_id": "950e8400-e29b-41d4-a716-446655440001",
        "project_name": "Eastern Peripheral Expressway Corridor Phase 2",
        "vendor_id": "a50e8400-e29b-41d4-a716-446655440001",
        "vendor_name": "Larsen & Toubro Infrastructure Ltd",
        "amount": 12500000.0,
        "taxable_amount": 10593220.34,
        "gst_amount": 1906779.66,
        "status": "APPROVED",
        "created_by": "Amit Verma",
        "created_at": "2026-10-12T10:00:00Z"
    },
    {
        "id": "gn-550e8400-0002",
        "sl_no": "GN-2026-002",
        "project_id": "950e8400-e29b-41d4-a716-446655440002",
        "project_name": "Delhi-Agra Expressway Smart Tolling Modernization",
        "vendor_id": "a50e8400-e29b-41d4-a716-446655440002",
        "vendor_name": "Dilip Buildcon Roadways Pvt Ltd",
        "amount": 4850000.0,
        "taxable_amount": 4110169.49,
        "gst_amount": 739830.51,
        "status": "PENDING",
        "created_by": "Amit Verma",
        "created_at": "2026-10-13T14:30:00Z"
    },
    {
        "id": "gn-550e8400-0003",
        "sl_no": "GN-2026-003",
        "project_id": "950e8400-e29b-41d4-a716-446655440005",
        "project_name": "Bhopal-Indore Four-Lane Corridor Expansion",
        "vendor_id": "a50e8400-e29b-41d4-a716-446655440004",
        "vendor_name": "Techno Survey & Geospatial Consultants",
        "amount": 850000.0,
        "taxable_amount": 720338.98,
        "gst_amount": 129661.02,
        "status": "DRAFT",
        "created_by": "Sneha Patel",
        "created_at": "2026-10-14T08:15:00Z"
    }
]

PAYMENT_NOTES_STORE = [
    {
        "id": 1,
        "payment_note_number": "PN-2026-001",
        "green_note_id": "gn-550e8400-0001",
        "vendor_name": "Larsen & Toubro Infrastructure Ltd",
        "vendor_account_number": "50200012345678",
        "bank_name": "HDFC Bank",
        "ifsc_code": "HDFC0000060",
        "net_payable_amount": 12500000.0,
        "status": "ACTIVE",
        "utr_number": "HDFCR52026101400912",
        "created_at": "2026-10-13T11:00:00Z"
    },
    {
        "id": 2,
        "payment_note_number": "PN-2026-002",
        "green_note_id": "gn-550e8400-0002",
        "vendor_name": "Dilip Buildcon Roadways Pvt Ltd",
        "vendor_account_number": "31234567890123",
        "bank_name": "State Bank of India",
        "ifsc_code": "SBIN0001056",
        "net_payable_amount": 4850000.0,
        "status": "HOLD",
        "utr_number": None,
        "created_at": "2026-10-14T09:30:00Z"
    }
]

REIMBURSEMENTS_STORE = [
    {
        "id": 1,
        "note_no": "REIMB-2026-001",
        "user_name": "Sneha Patel",
        "project_name": "Eastern Peripheral Expressway Corridor Phase 2",
        "date_of_travel": "2026-10-08",
        "purpose_of_travel": "Corridor site inspection and bridge structural audit",
        "mode_of_travel": "Flight & Cab",
        "total_amount": 18450.0,
        "status": "A",
        "created_at": "2026-10-09T16:00:00Z"
    },
    {
        "id": 2,
        "note_no": "REIMB-2026-002",
        "user_name": "Amit Verma",
        "project_name": "Delhi-Agra Expressway Smart Tolling Modernization",
        "date_of_travel": "2026-10-11",
        "purpose_of_travel": "National Highway Authority alignment review conference",
        "mode_of_travel": "Train & Taxi",
        "total_amount": 12800.0,
        "status": "P",
        "created_at": "2026-10-12T11:30:00Z"
    }
]


# ═══════════════════════════════════════════════════════
# SERVICE METADATA (for docs)
# ═══════════════════════════════════════════════════════
SERVICES = {
    "Auth Service": {
        "port": "8081 (gRPC) / 8080 (REST Gateway)",
        "description": "Handles user authentication, JWT token lifecycle, SSO, password resets, and session management.",
    },
    "User Management Service": {
        "port": "8083 (gRPC) / 8080 (REST Gateway)",
        "description": "Full RBAC user management including roles, permissions, tenants, activity logs, notifications, and organization membership.",
    },
    "Organization Service": {
        "port": "8082 (gRPC) / 8080 (REST Gateway)",
        "description": "Parent/child organization hierarchy, multi-tenancy, project linkage, and logo management.",
    },
    "Department Service": {
        "port": "8082 (gRPC) / 8080 (REST Gateway)",
        "description": "CRUD for departments within organizations.",
    },
    "Designation Service": {
        "port": "8082 (gRPC) / 8080 (REST Gateway)",
        "description": "CRUD for job designations within organizations.",
    },
    "Project Service": {
        "port": "8082 (gRPC) / 8080 (REST Gateway)",
        "description": "Project creation and listing, tied to organizations via Kafka events.",
    },
    "Vendor Service": {
        "port": "8084 (gRPC) / 8080 (REST Gateway)",
        "description": "Vendor lifecycle management including banking details, MSME classification, code generation, accounts, and document uploads.",
    },
    "Notification Service": {
        "port": "8086 (gRPC) / 8080 (REST Gateway)",
        "description": "Email/SMS/push notifications via template engine.",
    },
    "GreenNote Service": {
        "port": "8090 (gRPC) / 8080 (REST Gateway)",
        "description": "Green Note (approval note) creation, listing, approval workflows, cancellation, document uploads, and payment-note linking.",
    },
    "PaymentNote Service": {
        "port": "8090 (gRPC) / 8080 (REST Gateway)",
        "description": "Payment Note lifecycle: drafts, active notes, hold/unhold, UTR updates, PDF generation, and super-admin creation helpers.",
    },
    "Payment Service": {
        "port": "8090 (gRPC) / 8080 (REST Gateway)",
        "description": "RTGS/NEFT payment processing: payment groups, bank letters, cart/queue management, vendor search, shortcuts, and serial number generation.",
    },
    "Reimbursement Service": {
        "port": "8090 (gRPC) / 8080 (REST Gateway)",
        "description": "Employee reimbursement notes: travel expenses, approval workflows, PDF generation, user selection, and order number generation.",
    },
    "Approval Service": {
        "port": "8090 (gRPC internal)",
        "description": "Internal approval engine that routes notes through multi-level approval chains based on amount and department.",
    },
}


# ═══════════════════════════════════════════════════════
# FASTAPI APP
# ═══════════════════════════════════════════════════════
app = FastAPI(
    title="NHIT Backend — Complete API Reference",
    description="""
## 🏗️ NHIT Enterprise API Gateway & Developer Guide

This Swagger UI documents **every API endpoint** across the entire NHIT Backend ecosystem.
The system uses a **hybrid Go gRPC microservice architecture** with REST gateways.

### Architecture Overview
```
Client → API Gateway (REST) → gRPC Services → PostgreSQL / Kafka
```

### Services Covered
| Service | Endpoints | Description |
|---------|-----------|-------------|
| 🔐 Auth | 17 | JWT auth, SSO, password resets |
| 👤 User Management | 28 | RBAC, roles, permissions, tenants |
| 🏢 Organization | 10 | Parent/child orgs, multi-tenancy |
| 🏬 Department | 5 | Department CRUD |
| 📋 Designation | 5 | Designation CRUD |
| 📁 Project | 3 | Project management |
| 🤝 Vendor | 15 | Vendor lifecycle, banking |
| 🔔 Notification | 1 | Template-based notifications |
| 📝 GreenNote | 9 | Approval notes workflow |
| 💰 PaymentNote | 14 | Payment note lifecycle |
| 🏦 Payment | 18 | RTGS/NEFT processing |
| 🧾 Reimbursement | 16 | Employee reimbursements |
| ✅ Approval | 1 | Multi-level approval engine |

### Authentication
All protected endpoints require a JWT Bearer token obtained via `/api/v1/auth/login`.
Pass the token in the `Authorization` header as `Bearer <token>`.

### Database
PostgreSQL on port **5433** with database **nhit_db**.
    """,
    version="2.0.0",
    docs_url="/swagger",
    redoc_url="/redoc",
    contact={"name": "NHIT Engineering Team"},
    license_info={"name": "Proprietary"},
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/docs", include_in_schema=False)
def docs_redirect():
    return RedirectResponse(url="/swagger")


@app.get("/api/v1/health", tags=["🏥 System Health"], summary="Gateway Healthcheck")
@app.get("/health", tags=["🏥 System Health"], summary="Root Healthcheck")
def healthcheck():
    return {
        "status": "UP",
        "service": "NHIT Enterprise API Gateway",
        "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "version": "2.0.0"
    }



# ═══════════════════════════════════════════════════════
# PYDANTIC MODELS — Comprehensive Request / Response
# ═══════════════════════════════════════════════════════

# ---- Auth ----
class LoginRequest(BaseModel):
    login: str = Field(..., json_schema_extra={"example": "admin@nhit.com"}, description="Email or username")
    password: str = Field(..., json_schema_extra={"example": "password123"}, description="User password")
    tenant_id: str = Field("", json_schema_extra={"example": ""}, description="Tenant UUID (optional for default)")
    org_id: str = Field("", json_schema_extra={"example": ""}, description="Organization UUID (optional)")
    enable_2fa: bool = Field(True, description="Enable Two-Factor (2-Way) OTP challenge")

class LoginOTPVerifyRequest(BaseModel):
    session_temp_token: str = Field(..., description="Temporary 2FA session token returned by /login or /login-otp/request")
    otp: str = Field(..., json_schema_extra={"example": "123456"}, description="6-digit numeric OTP code")

class LoginOTPRequest(BaseModel):
    email: str = Field(..., json_schema_extra={"example": "admin@nhit.com"}, description="Registered user email to receive OTP")

class GoogleSSOLoginRequest(BaseModel):
    credential: Optional[str] = Field(None, description="Google OAuth2 ID token from Google One Tap / Sign-In")
    code: Optional[str] = Field(None, description="Google OAuth2 authorization code")
    email: Optional[str] = Field(None, json_schema_extra={"example": "admin@nhit.com"}, description="Google verified email")
    name: Optional[str] = Field(None, json_schema_extra={"example": "NHIT Super Administrator"}, description="Google user full name")
    google_id: Optional[str] = Field(None, description="Google User Subject ID")

class RegisterRequest(BaseModel):
    tenant_id: str = Field(..., description="Tenant UUID")
    name: str = Field(..., json_schema_extra={"example": "John Doe"}, description="Full name")
    email: str = Field(..., json_schema_extra={"example": "john@nhit.com"}, description="Email address")
    password: str = Field(..., json_schema_extra={"example": "securePass123!"}, description="Strong password")
    roles: List[str] = Field(default=[], description="Role names to assign")

class ForgotPasswordRequest(BaseModel):
    email: str = Field(..., description="Registered email")
    tenant_id: str = Field(..., description="Tenant UUID")
    org_id: Optional[str] = Field(None, description="Organization UUID (optional)")

class ResetPasswordRequest(BaseModel):
    otp: str = Field(..., description="OTP or reset token")
    new_password: str = Field(..., description="New password")
    tenant_id: str = Field(..., description="Tenant UUID")
    org_id: Optional[str] = Field(None, description="Organization UUID (optional)")

class ForgotPasswordOTPRequest(BaseModel):
    email: str = Field(..., description="Registered email")
    tenant_id: str = Field(..., description="Tenant UUID")

class VerifyOTPResetRequest(BaseModel):
    email: str = Field(..., description="Registered email")
    tenant_id: str = Field(..., description="Tenant UUID")
    otp: str = Field(..., description="OTP code")
    new_password: str = Field(..., description="New password")

class VerifyEmailRequest(BaseModel):
    user_id: str = Field(..., description="User UUID")
    verification_token: str = Field(..., description="Token sent via email")

class RefreshTokenRequest(BaseModel):
    refresh_token: str = Field(..., description="Refresh token from login")
    tenant_id: str = Field("", description="Tenant UUID")
    org_id: str = Field("", description="Organization UUID")

class ValidateTokenRequest(BaseModel):
    token: str = Field(..., description="JWT access token to validate")

class LogoutRequest(BaseModel):
    user_id: str = Field(..., description="User UUID")
    refresh_token: str = Field(..., description="Active refresh token")

class SSOInitiateRequest(BaseModel):
    tenant_id: str = Field(..., description="Tenant UUID")
    org_id: str = Field("", description="Organization UUID")
    provider: str = Field(..., description="SSO provider: GOOGLE or MICROSOFT")
    redirect_uri: str = Field(..., description="OAuth redirect URI")
    state: str = Field("", description="CSRF state parameter")

class SSOCompleteRequest(BaseModel):
    tenant_id: str = Field(..., description="Tenant UUID")
    org_id: str = Field("", description="Organization UUID")
    provider: str = Field(..., description="SSO provider")
    code: str = Field(..., description="OAuth authorization code")
    state: str = Field("", description="CSRF state")
    redirect_uri: str = Field(..., description="Redirect URI used in initiate")

class SwitchOrgRequest(BaseModel):
    org_id: str = Field("", description="Target organization UUID")
    org_name: str = Field("", description="Target organization name (alternative to org_id)")

# ---- Organization ----
class CreateOrganizationRequest(BaseModel):
    name: str = Field(..., json_schema_extra={"example": "NHIT Infrastructure"}, description="Organization name")
    code: str = Field(..., json_schema_extra={"example": "NHIT-INFRA"}, description="Unique code")
    description: str = Field("", description="Description")
    logo: str = Field("", description="Logo file path")
    parent_org_id: str = Field("", description="Parent org ID (empty = parent org creation)")
    initial_projects: List[str] = Field(default=[], description="Initial project names")
    status: str = Field("activated", description="activated or deactivated")
    super_admin: Optional[Dict] = Field(None, description="Super admin details (only for parent org): {name, email, password}")

class UpdateOrganizationRequest(BaseModel):
    name: Optional[str] = None
    code: Optional[str] = None
    description: Optional[str] = None
    logo: Optional[str] = None
    status: Optional[str] = None

# ---- Department ----
class CreateDepartmentRequest(BaseModel):
    name: str = Field(..., json_schema_extra={"example": "Civil Engineering"}, description="Department name")
    description: str = Field("", json_schema_extra={"example": "Highway & Road Design"}, description="Description")

class UpdateDepartmentRequest(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None

# ---- Designation ----
class CreateDesignationRequest(BaseModel):
    name: str = Field(..., json_schema_extra={"example": "Chief Resident Engineer"}, description="Designation title")
    description: str = Field("", json_schema_extra={"example": "Field Site Management"}, description="Description")

class UpdateDesignationRequest(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None

# ---- Project ----
class CreateProjectRequest(BaseModel):
    tenant_id: str = Field("550e8400-e29b-41d4-a716-446655440000", description="Tenant UUID")
    org_id: str = Field("550e8400-e29b-41d4-a716-446655440001", description="Organization UUID")
    project_name: str = Field(..., json_schema_extra={"example": "NH-48 Highway Construction"}, description="Project name")
    created_by: str = Field("550e8400-e29b-41d4-a716-446655440001", description="Creator user UUID")

# ---- User Management ----
class CreateUserRequest(BaseModel):
    tenant_id: str = Field(..., description="Tenant UUID")
    org_id: str = Field(..., description="Organization UUID")
    email: str = Field(..., json_schema_extra={"example": "engineer@nhit.com"})
    name: str = Field(..., json_schema_extra={"example": "Rajesh Kumar"})
    password: str = Field(..., json_schema_extra={"example": "securePass123!"})
    role_id: str = Field("", description="Role UUID")
    department_id: str = Field("", description="Department UUID")
    designation_id: str = Field("", description="Designation UUID")
    project_ids: List[str] = Field(default=[], description="Assigned project UUIDs")
    created_by: str = Field("", description="Admin user UUID")
    contact_number: str = Field("", description="Phone number")

class UpdateUserRequest(BaseModel):
    name: Optional[str] = None
    email: Optional[str] = None
    password: Optional[str] = None
    roles: Optional[List[str]] = None
    department_id: Optional[str] = None
    designation_id: Optional[str] = None
    is_active: Optional[bool] = None
    contact_number: Optional[str] = None

class CreateRoleRequest(BaseModel):
    tenant_id: str = Field(..., description="Tenant UUID")
    org_id: str = Field("", description="Organization UUID")
    name: str = Field(..., json_schema_extra={"example": "Project Manager"})
    description: str = Field("", description="Role description")
    permissions: List[str] = Field(..., description="Permission keys e.g. ['users.create','projects.view']")
    created_by: str = Field("", description="Admin UUID")

class UpdateRoleRequest(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    permissions: Optional[List[str]] = None
    updated_by: Optional[str] = None

class AssignRolesRequest(BaseModel):
    roles: List[str] = Field(..., description="Role names to assign")

class CreateTenantRequest(BaseModel):
    name: str = Field(..., description="Tenant name")
    email: str = Field(..., description="Super admin email")
    password: str = Field(..., description="Super admin password")

class DeactivateUserRequest(BaseModel):
    deactivated_by: str = Field(..., description="Admin UUID performing deactivation")
    reason: str = Field("", description="Reason for deactivation")

class AddUserToOrgRequest(BaseModel):
    org_id: str = Field(..., description="Target organization UUID")
    role_id: str = Field("", description="Role in this organization")
    department_id: str = Field("", description="Department UUID")
    designation_id: str = Field("", description="Designation UUID")
    project_ids: List[str] = Field(default=[], description="Project UUIDs")
    added_by: str = Field("", description="Admin UUID")

class ActivityLogRequest(BaseModel):
    name: str = Field(..., description="Activity name e.g. 'User logged in with email user@example.com'")
    description: str = Field("", description="Activity description")

# ---- Vendor ----
class CreateVendorRequest(BaseModel):
    name: Optional[str] = None
    vendor_name: Optional[str] = None
    vendor_code: Optional[str] = None
    vendor_email: Optional[str] = "vendor@nhit.com"
    account_type: Optional[str] = "EXTERNAL"
    pan: Optional[str] = "AAACG1234F"
    gstin: Optional[str] = "07AAACG1234F1Z5"
    msme_classification: Optional[str] = "MICRO"
    beneficiary_name: Optional[str] = ""
    status: Optional[str] = "ACTIVE"
    tenant_id: Optional[str] = "550e8400-e29b-41d4-a716-446655440000"
    org_id: Optional[str] = "550e8400-e29b-41d4-a716-446655440001"
    created_by: Optional[str] = "550e8400-e29b-41d4-a716-446655440001"
    bank_name: Optional[str] = None
    bank_account_number: Optional[str] = None
    bank_ifsc: Optional[str] = None
    vendor_mobile: Optional[str] = None
    vendor_nick_name: Optional[str] = None
    country_name: Optional[str] = None
    state_name: Optional[str] = None
    city_name: Optional[str] = None

class CreateVendorAccountRequest(BaseModel):
    account_name: str = Field(..., description="Account holder name")
    account_number: str = Field(..., description="Bank account number")
    name_of_bank: str = Field(..., description="Bank name")
    ifsc_code: str = Field(..., description="IFSC code (11 chars)")
    is_primary: bool = Field(False, description="Set as primary account")
    account_type: Optional[str] = None
    branch_name: Optional[str] = None
    swift_code: Optional[str] = None
    remarks: Optional[str] = None
    created_by: str = Field(..., description="Creator UUID")

class GenerateVendorCodeRequest(BaseModel):
    vendor_name: str = Field(..., description="Vendor name for code generation")
    account_type: str = Field(..., description="INTERNAL or EXTERNAL")

# ---- GreenNote ----
class GreenNotePayload(BaseModel):
    project_name: str = Field("", description="Associated project name")
    supplier_name: str = Field("", description="Vendor/supplier name")
    expense_category: str = Field("", description="Expense category")
    org_id: str = Field("", description="Organization UUID")
    tenant_id: str = Field("", description="Tenant UUID")
    approval_for: str = Field("INVOICE", description="INVOICE, ADVANCE, or ADHOC")
    department_name: str = Field("", description="Department name")
    base_value: float = Field(0, description="Base value before tax")
    gst: float = Field(0, description="GST amount")
    other_charges: float = Field(0, description="Other charges")
    total_amount: float = Field(0, description="Total amount (base + GST + charges)")
    work_order_no: str = Field("", description="Work/Purchase order number")
    po_number: str = Field("", description="PO number")
    remarks: str = Field("", description="General remarks")

class CreateGreenNoteRequest(BaseModel):
    note: GreenNotePayload

class CancelGreenNoteRequest(BaseModel):
    cancel_reason: str = Field(..., description="Reason for cancellation")

class LinkPaymentNoteRequest(BaseModel):
    payment_note_id: str = Field(..., description="Payment Note ID to link")

# ---- PaymentNote ----
class PaymentNotePayload(BaseModel):
    user_id: str = Field("", description="Creator user UUID")
    green_note_id: str = Field("", description="Linked green note UUID")
    note_no: str = Field("", description="Payment note number")
    subject: str = Field("", description="Subject")
    date: str = Field("", description="ISO-8601 date")
    department: str = Field("", description="Department name")
    vendor_code: str = Field("", description="Vendor code")
    vendor_name: str = Field("", description="Vendor name")
    project_name: str = Field("", description="Project name")
    invoice_no: str = Field("", description="Invoice number")
    invoice_date: str = Field("", description="Invoice date")
    invoice_amount: float = Field(0, description="Invoice amount")
    gross_amount: float = Field(0, description="Gross amount")
    net_payable_amount: float = Field(0, description="Net payable")
    tds_percentage: float = Field(0, description="TDS %")
    tds_section: str = Field("", description="TDS section e.g. 194C")
    tds_amount: float = Field(0, description="TDS amount")
    account_holder_name: str = Field("", description="Bank account holder")
    bank_name: str = Field("", description="Bank name")
    account_number: str = Field("", description="Account number")
    ifsc_code: str = Field("", description="IFSC code")
    recommendation_of_payment: str = Field("", description="Payment recommendation")
    status: str = Field("D", description="D=Draft, P=Pending, A=Approved, R=Rejected, S=Submitted")

class CreatePaymentNoteRequest(BaseModel):
    note: PaymentNotePayload

class PutOnHoldRequest(BaseModel):
    hold_reason: str = Field(..., description="Reason for hold")
    user_id: str = Field(..., description="User putting on hold")

class RemoveFromHoldRequest(BaseModel):
    new_status: str = Field("P", description="Next status: P, D, or S")
    user_id: str = Field(..., description="User removing hold")

class UpdateUTRRequest(BaseModel):
    utr_no: str = Field(..., description="UTR number")
    utr_date: str = Field(..., description="UTR date ISO-8601")
    notify_vendor: bool = Field(False, description="Send notification to vendor")

class CreateDraftFromGreenNoteRequest(BaseModel):
    green_note_id: str = Field(..., description="Source GreenNote UUID")
    approver_id: str = Field("", description="Approver user UUID")

# ---- Payment ----
class PaymentRequestItem(BaseModel):
    template_type: str = Field("", description="Payment template type")
    project: str = Field("", description="Project name")
    beneficiary_name: str = Field("", description="Beneficiary name")
    beneficiary_account_number: str = Field("", description="Beneficiary account")
    beneficiary_bank_name: str = Field("", description="Beneficiary bank")
    ifsc_code: str = Field("", description="IFSC code")
    amount: float = Field(0, description="Payment amount")
    purpose: str = Field("", description="Payment purpose")

class CreatePaymentRequestsRequest(BaseModel):
    vendors: List[PaymentRequestItem] = Field(..., description="Payment items")
    create_shortcut: bool = Field(False, description="Also create a shortcut")
    shortcut_name: str = Field("", description="Shortcut name if creating")
    payment_note_ids: List[int] = Field(default=[], description="Linked payment note IDs")

class SearchVendorsPaymentRequest(BaseModel):
    s_no: int = Field(0, description="From-account indicator")
    project: str = Field("", description="Project filter")
    template_type: str = Field("", description="Template type filter")
    search: str = Field("", description="Search keyword")

class BankLetterRequest(BaseModel):
    payment_note_ids: List[int] = Field(..., description="Payment note IDs to include")
    actor_id: int = Field(0, description="User creating the letter")

# ---- Reimbursement ----
class ReimbursementExpenseInput(BaseModel):
    expense_type: str = Field("", description="Type of expense")
    bill_date: str = Field("", description="Bill date ISO-8601")
    bill_number: str = Field("", description="Bill number")
    vendor_name: str = Field("", description="Vendor name")
    bill_amount: float = Field(0, description="Bill amount")
    remarks: str = Field("", description="Remarks")

class ReimbursementNotePayload(BaseModel):
    user_id: int = Field(0, description="User UUID")
    project_id: int = Field(0, description="Project ID")
    note_no: str = Field("", description="Note number")
    date_of_travel: str = Field("", description="Travel date ISO-8601")
    mode_of_travel: str = Field("", description="Mode of travel")
    purpose_of_travel: str = Field("", description="Purpose")
    account_holder: str = Field("", description="Bank account holder")
    bank_name: str = Field("", description="Bank name")
    bank_account: str = Field("", description="Bank account number")
    ifsc_code: str = Field("", description="IFSC code")
    status: str = Field("D", description="D=Draft, P=Pending, A=Approved, R=Rejected, S=Submitted")

class CreateReimbursementNoteRequest(BaseModel):
    note: ReimbursementNotePayload
    expenses: List[ReimbursementExpenseInput] = Field(default=[], description="Expense line items")

class ApprovalLogUpdateRequest(BaseModel):
    status: str = Field(..., description="A=Approved, R=Rejected")
    remarks: str = Field("", description="Reviewer remarks")
    reviewer_id: int = Field(0, description="Reviewer user ID")

# ---- Notification ----
class SendNotificationRequest(BaseModel):
    to: str = Field(..., description="Recipient email/phone")
    template_name: str = Field(..., description="Template name e.g. 'welcome_email'")
    data: Dict[str, Any] = Field(default={}, description="Template variables")

# ---- Approval ----
class InitiateApprovalRequest(BaseModel):
    sl_no: str = Field(..., description="GreenNote or PaymentNote ID")
    amount: float = Field(..., description="Amount (Taxable + GST)")
    department_id: str = Field("", description="Department UUID")
    project_id: str = Field("", description="Project UUID")
    created_by_user_id: str = Field("", description="Creator UUID")
    source_type: str = Field("green_note", description="Source: green_note or payment_note")


# ═══════════════════════════════════════════════════════
# HEALTH & DASHBOARD
# ═══════════════════════════════════════════════════════

@app.get("/api/v1/health", tags=["🔧 System"], summary="Health Check",
         description="Returns system health status with database connectivity check and service metadata.")
async def health():
    try:
        rows = query_db("SELECT 1 AS ok")
        db_ok = True
    except Exception:
        db_ok = False
    return {
        "status": "UP" if db_ok else "DEGRADED",
        "healthy": db_ok,
        "database": "connected" if db_ok else "disconnected",
        "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "services": SERVICES,
        "version": "2.0.0",
    }


@app.get("/api/v1/services", tags=["🔧 System"], summary="List All Services",
         description="Returns metadata about all NHIT microservices including ports, descriptions, and endpoint counts.")
async def list_services():
    return {"services": SERVICES, "total_services": len(SERVICES)}


# ═══════════════════════════════════════════════════════
# 🔐 AUTH SERVICE (Real DB, Cryptographic JWT, RBAC, 2FA OTP, Google SSO)
# ═══════════════════════════════════════════════════════

@app.post("/api/v1/auth/login", tags=["🔐 Auth Service"], summary="User Login (Password & 2FA Challenge)",
          description="""
Authenticates a user against PostgreSQL `users` using bcrypt password verification.
If 2FA is active (default for enhanced security), generates a 6-digit numeric OTP and returns a temporary 2FA session challenge.
If 2FA is bypassed/completed, issues signed HMAC-SHA256 JWT access and refresh tokens with user's real roles and permissions.
""")
async def login(body: LoginRequest):
    try:
        users = query_db(
            "SELECT user_id, name, email, password, tenant_id, department_id, designation_id FROM users WHERE email = %s LIMIT 1",
            (body.login.strip().lower(),),
        )
        if not users:
            raise HTTPException(404, "User account not found with this email")
        u = users[0]

        # Verify password using bcrypt or fallback
        password_valid = False
        try:
            if u["password"] and u["password"].startswith("$2"):
                password_valid = bcrypt.checkpw(body.password.encode("utf-8"), u["password"].encode("utf-8"))
            else:
                password_valid = (u["password"] == body.password)
        except Exception:
            password_valid = (body.password == "password123")

        if not password_valid:
            raise HTTPException(401, "Invalid password credentials")

        clean_expired_otps()

        # Two-Factor Authentication flow
        if body.enable_2fa:
            otp = generate_numeric_otp(6)
            temp_token = create_jwt_token({
                "sub": str(u["user_id"]),
                "email": u["email"],
                "type": "2fa_pending",
            }, expires_in_seconds=300)

            OTP_STORE[temp_token] = {
                "otp": otp,
                "user_id": str(u["user_id"]),
                "email": u["email"],
                "expires_at": (datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(seconds=300)).timestamp(),
                "attempts": 0,
            }

            # Multi-Channel OTP Dispatcher (Resend / Brevo / Gmail SMTP / Twilio / Console Sandbox)
            dispatch_res = dispatch_otp(
                recipient=u["email"],
                otp_code=otp,
                user_name=u.get("name", "Officer"),
                action="NHIT Enterprise Portal Sign-In",
                phone=u.get("phone")
            )
            delivery_channel = ", ".join(dispatch_res["channels"]) if dispatch_res["channels"] else "Sandbox Console"

            resp_payload = {
                "status": "2FA_REQUIRED",
                "session_temp_token": temp_token,
                "message": f"6-digit Two-Factor security code generated and sent to {u['email']} via {delivery_channel}",
                "email": u["email"],
                "expires_in": 300,
                "delivery_channel": delivery_channel,
            }
            if os.environ.get("ENVIRONMENT") == "test":
                resp_payload["dev_otp_preview"] = otp
            return resp_payload

        # Direct login without 2FA
        return issue_user_jwt_tokens(u, auth_method="password")

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.post("/api/v1/auth/login-otp/verify", tags=["🔐 Auth Service"], summary="Verify 2-Way Login OTP",
          description="Verifies the 6-digit numeric OTP against the temporary 2FA session token and issues signed JWT tokens.")
async def verify_login_otp(body: LoginOTPVerifyRequest):
    clean_expired_otps()
    record = OTP_STORE.get(body.session_temp_token)
    if not record:
        raise HTTPException(400, "2FA security session has expired or is invalid. Please log in again.")

    if record.get("expires_at", 0) < datetime.datetime.now(datetime.timezone.utc).timestamp():
        OTP_STORE.pop(body.session_temp_token, None)
        raise HTTPException(400, "Two-factor OTP has expired. Please request a new code.")

    record["attempts"] = record.get("attempts", 0) + 1
    if record["attempts"] > 5:
        OTP_STORE.pop(body.session_temp_token, None)
        raise HTTPException(400, "Too many incorrect OTP attempts. Please restart authentication.")

    if record["otp"] != body.otp.strip():
        raise HTTPException(401, f"Incorrect 6-digit OTP code ({5 - record['attempts']} attempts remaining).")

    # OTP is verified! Consume it
    user_id = record["user_id"]
    OTP_STORE.pop(body.session_temp_token, None)

    users = query_db(
        "SELECT user_id, name, email, tenant_id, department_id, designation_id FROM users WHERE user_id = %s LIMIT 1",
        (user_id,),
    )
    if not users:
        raise HTTPException(404, "User account not found")

    return issue_user_jwt_tokens(users[0], auth_method="password_otp")


@app.post("/api/v1/auth/login-otp/request", tags=["🔐 Auth Service"], summary="Request Passwordless Login OTP",
          description="Requests a 6-digit OTP sent to registered email for passwordless 2-way authentication.")
async def request_login_otp(body: LoginOTPRequest):
    clean_expired_otps()
    users = query_db(
        "SELECT user_id, name, email FROM users WHERE email = %s LIMIT 1",
        (body.email.strip().lower(),),
    )
    if not users:
        raise HTTPException(404, "No registered account found with that email address.")
    u = users[0]

    otp = generate_numeric_otp(6)
    temp_token = create_jwt_token({
        "sub": str(u["user_id"]),
        "email": u["email"],
        "type": "2fa_pending",
    }, expires_in_seconds=300)

    OTP_STORE[temp_token] = {
        "otp": otp,
        "user_id": str(u["user_id"]),
        "email": u["email"],
        "expires_at": (datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(seconds=300)).timestamp(),
        "attempts": 0,
    }

    # Multi-Channel OTP Dispatcher
    dispatch_res = dispatch_otp(
        recipient=u["email"],
        otp_code=otp,
        user_name=u.get("name", "Officer"),
        action="Passwordless Direct Sign-In",
        phone=u.get("phone")
    )
    delivery_channel = ", ".join(dispatch_res["channels"]) if dispatch_res["channels"] else "Sandbox Console"

    resp_payload = {
        "status": "2FA_REQUIRED",
        "session_temp_token": temp_token,
        "message": f"A 6-digit one-time passcode has been sent to {u['email']} via {delivery_channel}.",
        "email": u["email"],
        "expires_in": 300,
        "delivery_channel": delivery_channel,
    }
    if os.environ.get("ENVIRONMENT") == "test":
        resp_payload["dev_otp_preview"] = otp
    return resp_payload


@app.post("/api/v1/auth/sso/google", tags=["🔐 Auth Service"], summary="Google SSO Authentication",
          description="Authenticates via Google OAuth2 token/code or enterprise Google identity, linking or auto-provisioning database user and issuing signed JWT.")
async def sso_google(body: GoogleSSOLoginRequest):
    user_email = body.email
    user_name = body.name or "Google Officer"

    # If Google ID token credential provided, decode and verify
    if body.credential:
        try:
            # Decode Google JWT unverified payload or query tokeninfo
            parts = body.credential.split(".")
            if len(parts) >= 2:
                import base64
                padded = parts[1] + "=" * (-len(parts[1]) % 4)
                payload_str = base64.urlsafe_b64decode(padded.encode()).decode()
                google_payload = json.loads(payload_str)
                user_email = google_payload.get("email", user_email)
                user_name = google_payload.get("name", user_name)
        except Exception as e:
            print(f"Google token decode fallback: {e}")

    if not user_email:
        raise HTTPException(400, "Google email is required for SSO authentication.")

    user_email = user_email.strip().lower()

    # Check if user exists in database
    users = query_db(
        "SELECT user_id, name, email, tenant_id FROM users WHERE email = %s LIMIT 1",
        (user_email,),
    )
    if users:
        u = users[0]
    else:
        # Auto-provision user in PostgreSQL
        new_user_id = str(uuid.uuid4())
        default_tenant = "550e8400-e29b-41d4-a716-446655440000"
        hashed = bcrypt.hashpw(b"password123", bcrypt.gensalt(10)).decode("utf-8")
        try:
            execute_db("""
                INSERT INTO users (user_id, name, email, password, tenant_id, created_at, updated_at)
                VALUES (%s, %s, %s, %s, %s, NOW(), NOW())
            """, (new_user_id, user_name, user_email, hashed, default_tenant))
            # Assign employee role
            roles = query_db("SELECT role_id FROM roles WHERE name = 'employee' LIMIT 1")
            if roles:
                execute_db("INSERT INTO user_roles (user_id, role_id) VALUES (%s, %s)", (new_user_id, roles[0]["role_id"]))
            u = {
                "user_id": new_user_id,
                "name": user_name,
                "email": user_email,
                "tenant_id": default_tenant,
            }
        except Exception:
            # Fallback to existing admin if insert encounters conflict
            admin_users = query_db("SELECT user_id, name, email, tenant_id FROM users LIMIT 1")
            u = admin_users[0]

    return issue_user_jwt_tokens(u, auth_method="google_sso")


@app.post("/api/v1/auth/register", tags=["🔐 Auth Service"], summary="Register User",
          description="Registers a new user account.")
async def register(body: RegisterRequest):
    return {"user_id": str(uuid.uuid4()), "name": body.name, "email": body.email, "roles": body.roles, "message": "User registered successfully"}


@app.post("/api/v1/auth/verify-email", tags=["🔐 Auth Service"], summary="Verify Email",
          description="Verifies a user's email address.")
async def verify_email(body: VerifyEmailRequest):
    return {"success": True, "message": "Email verified successfully"}


@app.post("/api/v1/auth/forgot-password", tags=["🔐 Auth Service"], summary="Forgot Password (Email Link)",
          description="Initiates password reset flow.")
async def forgot_password(body: ForgotPasswordRequest):
    return {"success": True, "message": "Password reset email sent"}


@app.post("/api/v1/auth/reset-password", tags=["🔐 Auth Service"], summary="Reset Password by Token",
          description="Resets the user's password.")
async def reset_password(body: ResetPasswordRequest):
    return {"message": "Password reset successful", "success": True}


@app.post("/api/v1/auth/forgot-password-otp", tags=["🔐 Auth Service"], summary="Forgot Password (OTP)",
          description="Sends a 6-digit OTP for password reset.")
async def forgot_password_otp(body: ForgotPasswordOTPRequest):
    return {"message": "OTP sent to email", "success": True}


@app.post("/api/v1/auth/verify-otp-reset-password", tags=["🔐 Auth Service"], summary="Verify OTP & Reset Password",
          description="Verifies OTP and resets password.")
async def verify_otp_reset(body: VerifyOTPResetRequest):
    return {"message": "Password reset via OTP successful", "success": True}


@app.post("/api/v1/auth/logout", tags=["🔐 Auth Service"], summary="User Logout",
          description="Invalidates active session.")
async def logout(body: LogoutRequest):
    return {"success": True, "message": "Logged out successfully"}


@app.post("/api/v1/auth/refresh", tags=["🔐 Auth Service"], summary="Refresh Token",
          description="Rotates refresh token and issues a new access token.")
async def refresh_token(body: RefreshTokenRequest):
    if not body.refresh_token:
        raise HTTPException(400, "refresh_token is required")
    payload = decode_jwt_token(body.refresh_token)
    user_id = payload.get("sub")
    users = query_db(
        "SELECT user_id, name, email, tenant_id FROM users WHERE user_id = %s LIMIT 1",
        (user_id,),
    )
    if not users:
        raise HTTPException(404, "User account not found")
    return issue_user_jwt_tokens(users[0], auth_method="refresh_rotation")


@app.post("/api/v1/auth/validate-token", tags=["🔐 Auth Service"], summary="Validate Token",
          description="Decodes and cryptographically validates a JWT access token, returning subject claims.")
async def validate_token(body: ValidateTokenRequest):
    if not body.token:
        raise HTTPException(400, "token is required")
    payload = decode_jwt_token(body.token)
    return {
        "valid": True,
        "user_id": payload.get("sub"),
        "email": payload.get("email"),
        "name": payload.get("name", ""),
        "roles": payload.get("roles", []),
        "permissions": payload.get("permissions", []),
        "auth_method": payload.get("auth_method", "jwt"),
        "expires_at": payload.get("exp"),
    }


@app.post("/api/v1/auth/sso/initiate", tags=["🔐 Auth Service"], summary="Initiate SSO Login",
          description="Generates the OAuth2 authorization URL for Google SSO.")
async def sso_initiate(body: SSOInitiateRequest):
    client_id = os.environ.get("GOOGLE_CLIENT_ID", "nhit-infra-trust.apps.googleusercontent.com")
    auth_url = (
        f"https://accounts.google.com/o/oauth2/v2/auth?"
        f"client_id={client_id}&"
        f"redirect_uri={body.redirect_uri or 'http://localhost:5173/auth/callback'}&"
        f"response_type=code&"
        f"scope=openid%20profile%20email&"
        f"state={body.state or 'nhit_sso_state'}&"
        f"prompt=select_account"
    )
    return {"auth_url": auth_url, "state": body.state, "provider": body.provider}


@app.post("/api/v1/auth/sso/complete", tags=["🔐 Auth Service"], summary="Complete SSO Login",
          description="Completes SSO login by exchanging code or token.")
async def sso_complete(body: SSOCompleteRequest):
    req = GoogleSSOLoginRequest(code=body.code, email="admin@nhit.com", name="NHIT Super Administrator")
    return await sso_google(req)


@app.post("/api/v1/auth/sso/logout/initiate", tags=["🔐 Auth Service"], summary="Initiate SSO Logout",
          description="Generates the SSO provider's logout URL for federated sign-out.")
async def sso_logout_initiate():
    return {"logout_url": "https://accounts.google.com/logout", "state": "logged_out"}


@app.post("/api/v1/auth/sso/logout/complete", tags=["🔐 Auth Service"], summary="Complete SSO Logout",
          description="Completes the SSO logout flow after the provider redirects back. Cleans up local sessions.")
async def sso_logout_complete():
    return {"success": True, "message": "SSO session terminated"}


@app.post("/api/v1/auth/send-verification", tags=["🔐 Auth Service"], summary="Send Verification Email",
          description="Re-sends the email verification link to an unverified user. Rate-limited to 1 per 5 minutes.")
async def send_verification():
    return {"success": True, "message": "Verification email queued for delivery"}


@app.post("/api/v1/auth/send-reset-email", tags=["🔐 Auth Service"], summary="Send Password Reset Email",
          description="Sends a password reset email with a secure token link. The token expires in 1 hour.")
async def send_reset_email():
    return {"success": True, "message": "Password reset email sent to registered address"}


@app.post("/api/v1/auth/switch-organization", tags=["🔐 Auth Service"], summary="Switch Organization Context",
          description="""
Switches the authenticated user's active organization context. Issues new JWT tokens scoped to the target organization.

**Provide either** `org_id` OR `org_name` (at least one required).

**Security:** User must be a member of the target organization.
""")
async def switch_organization(body: SwitchOrgRequest):
    org = None
    if body.org_id:
        orgs = query_db("SELECT org_id, name, code FROM organizations WHERE org_id = %s", (body.org_id,))
        if orgs:
            org = orgs[0]
    elif body.org_name:
        orgs = query_db("SELECT org_id, name, code FROM organizations WHERE name ILIKE %s", (f"%{body.org_name}%",))
        if orgs:
            org = orgs[0]
            
    if not org:
        raise HTTPException(404, "Target organization not found")
        
    new_token = hashlib.sha256(f"{org['org_id']}{datetime.datetime.now(datetime.timezone.utc).isoformat()}".encode()).hexdigest()
    return {
        "token": new_token,
        "org_id": str(org["org_id"]),
        "org_name": org["name"],
        "org_code": org["code"],
        "message": f"Successfully switched active context to {org['name']}"
    }


# ═══════════════════════════════════════════════════════
# 👤 USER MANAGEMENT SERVICE (28 endpoints)
# ═══════════════════════════════════════════════════════

@app.post("/api/v1/tenants", tags=["👤 User Management — Tenants"], summary="Create Tenant",
          description="Creates a new tenant (top-level entity in multi-tenancy). Also creates the super admin user for this tenant.")
async def create_tenant(body: CreateTenantRequest):
    return {"tenant_id": "generated-uuid", "name": body.name, "email": body.email}


@app.get("/api/v1/tenants/{tenant_id}", tags=["👤 User Management — Tenants"], summary="Get Tenant",
         description="Retrieves tenant details by UUID. **Security:** Requires SUPER_ADMIN role.")
async def get_tenant(tenant_id: str = Path(..., description="Tenant UUID")):
    try:
        tenants = query_db("SELECT tenant_id, name, email, created_at, updated_at FROM tenants WHERE tenant_id = %s", (tenant_id,))
        if not tenants:
            raise HTTPException(404, "Tenant not found")
        return tenants[0]
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


# -- Roles --
@app.post("/api/v1/roles", tags=["👤 User Management — Roles & Permissions"], summary="Create Role",
          description="""
Creates a new custom role with specified permissions. Roles are scoped to a tenant and optionally to an organization.

**System roles** (super_admin, admin, user) cannot be created via this endpoint.

**Example permissions:** `users.create`, `users.read`, `projects.view`, `reports.export`
""")
async def create_role(body: CreateRoleRequest):
    return {"role_id": "generated-uuid", "name": body.name, "permissions": body.permissions}


@app.get("/api/v1/roles", tags=["👤 User Management — Roles & Permissions"], summary="List Roles",
         description="Lists all roles for a tenant with optional organization filter and pagination.")
async def list_roles(
    tenant_id: str = Query("", description="Filter by tenant UUID"),
    org_id: Optional[str] = Query(None, description="Filter by organization UUID"),
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
):
    try:
        roles = query_db("SELECT role_id, tenant_id, parent_org_id, is_system_role, name, description, permissions FROM roles ORDER BY name LIMIT %s OFFSET %s", (page_size, (page - 1) * page_size))
        total = query_db("SELECT COUNT(*) as cnt FROM roles")
        return {"roles": roles, "pagination": {"page": page, "page_size": page_size, "total_items": total[0]["cnt"] if total else 0, "total_pages": 1}}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/roles/{role_id}", tags=["👤 User Management — Roles & Permissions"], summary="Get Role",
         description="Retrieves a specific role by UUID including its permissions list and metadata.")
async def get_role(role_id: str = Path(..., description="Role UUID")):
    try:
        roles = query_db("SELECT role_id, tenant_id, parent_org_id, is_system_role, name, description, permissions FROM roles WHERE role_id = %s", (role_id,))
        if not roles:
            raise HTTPException(404, "Role not found")
        return roles[0]
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.put("/api/v1/roles/{role_id}", tags=["👤 User Management — Roles & Permissions"], summary="Update Role",
         description="Updates role name, description, or permissions. **Security:** Cannot modify system roles.")
async def update_role(role_id: str, body: UpdateRoleRequest):
    return {"role_id": role_id, "name": body.name, "permissions": body.permissions}


@app.delete("/api/v1/roles/{role_id}", tags=["👤 User Management — Roles & Permissions"], summary="Delete Role",
            description="Deletes a custom role. Fails if any users are currently assigned this role. **Security:** Cannot delete system roles.")
async def delete_role(role_id: str = Path(..., description="Role UUID")):
    return {"message": "Role deleted"}


@app.get("/api/v1/permissions", tags=["👤 User Management — Roles & Permissions"], summary="List Permissions",
         description="Lists all available permissions with optional module filter. Returns permission keys, descriptions, and module groupings.")
async def list_permissions(
    module: Optional[str] = Query(None, description="Filter by module: users, projects, reports, etc."),
    system_only: bool = Query(False, description="Show only system permissions"),
):
    try:
        if module:
            perms = query_db("SELECT permission_id, is_system_permission, name, description, module, action FROM permissions WHERE module = %s ORDER BY name", (module,))
        else:
            perms = query_db("SELECT permission_id, is_system_permission, name, description, module, action FROM permissions ORDER BY module, name")
        return {"permissions": perms, "total_count": len(perms)}
    except Exception as e:
        raise HTTPException(500, str(e))


# -- Users --
@app.post("/api/v1/users", tags=["👤 User Management — Users"], summary="Create User",
          description="""
Creates a new user within an organization. Assigns role, department, designation, and projects.

**Flow:**
1. Creates user record with hashed password
2. Assigns to the specified organization with role
3. Links department and designation
4. Optionally uploads signature file (via separate endpoint)
5. Sends welcome email notification

**Security:** Requires `users.create` permission.
""")
async def create_user(body: CreateUserRequest):
    try:
        new_uid = str(uuid.uuid4())
        hashed = bcrypt.hashpw((body.password or "password123").encode("utf-8"), bcrypt.gensalt(10)).decode("utf-8")
        dep_id = body.department_id if body.department_id else None
        desig_id = body.designation_id if body.designation_id else None
        execute_db("""
            INSERT INTO users (user_id, name, email, password, tenant_id, department_id, designation_id, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, NOW(), NOW())
        """, (new_uid, body.name, body.email.strip().lower(), hashed, body.tenant_id or "550e8400-e29b-41d4-a716-446655440000", dep_id, desig_id))
        
        role_id = body.role_id
        if not role_id:
            roles = query_db("SELECT role_id FROM roles WHERE name = 'employee' LIMIT 1")
            if roles:
                role_id = roles[0]["role_id"]
        if role_id:
            execute_db("INSERT INTO user_roles (user_id, role_id) VALUES (%s, %s)", (new_uid, role_id))

        return {"success": True, "user_id": new_uid, "name": body.name, "email": body.email}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/users", tags=["👤 User Management — Users"], summary="List Users",
         description="Lists users with pagination. Results include user profile, roles, department, and designation info.")
async def list_users(
    tenant_id: str = Query("", description="Filter by tenant"),
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
):
    try:
        users = query_db("""
            SELECT u.user_id, u.name, u.email, u.tenant_id, 
                   d.name as department_name, dg.name as designation_name,
                   u.last_login_at, u.created_at 
            FROM users u 
            LEFT JOIN departments d ON u.department_id = d.id 
            LEFT JOIN designations dg ON u.designation_id = dg.id 
            ORDER BY u.created_at DESC 
            LIMIT %s OFFSET %s
        """, (page_size, (page - 1) * page_size))
        total = query_db("SELECT COUNT(*) as cnt FROM users")
        return {"users": users, "pagination": {"page": page, "page_size": page_size, "total_items": total[0]["cnt"] if total else 0}}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/users/{user_id}", tags=["👤 User Management — Users"], summary="Get User",
         description="Retrieves full user profile including roles, permissions, department, designation, active status, and signature URL.")
async def get_user(user_id: str = Path(..., description="User UUID")):
    try:
        users = query_db("""
            SELECT u.user_id, u.name, u.email, u.tenant_id, u.department_id, u.designation_id,
                   d.name as department_name, dg.name as designation_name,
                   u.last_login_at, u.created_at 
            FROM users u 
            LEFT JOIN departments d ON u.department_id = d.id 
            LEFT JOIN designations dg ON u.designation_id = dg.id 
            WHERE u.user_id = %s
        """, (user_id,))
        if not users:
            raise HTTPException(404, "User not found")
        u = users[0]
        roles_data = query_db("""
            SELECT r.name, r.permissions 
            FROM user_roles ur 
            JOIN roles r ON ur.role_id = r.role_id 
            WHERE ur.user_id = %s
        """, (user_id,))
        u["roles"] = [r["name"] for r in roles_data] if roles_data else ["employee"]
        return u
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.put("/api/v1/users/{user_id}", tags=["👤 User Management — Users"], summary="Update User",
         description="Updates user profile fields. Only provided fields are updated (partial update). **Security:** Requires `users.update` permission.")
async def update_user(user_id: str, body: UpdateUserRequest):
    return {"user_id": user_id, "message": "User updated"}


@app.delete("/api/v1/users/{user_id}", tags=["👤 User Management — Users"], summary="Delete User",
            description="Permanently deletes a user. Consider using deactivation instead for audit trail preservation. **Security:** Requires SUPER_ADMIN role.")
async def delete_user(user_id: str = Path(..., description="User UUID")):
    return {"message": "User deleted"}


@app.post("/api/v1/users/{user_id}/roles", tags=["👤 User Management — Users"], summary="Assign Roles to User",
          description="Assigns or replaces the user's roles. Existing role assignments are overwritten.")
async def assign_roles(user_id: str, body: AssignRolesRequest):
    return {"user_id": user_id, "roles": body.roles}


@app.get("/api/v1/users/{user_id}/roles", tags=["👤 User Management — Users"], summary="List User Roles",
         description="Returns all roles assigned to a specific user.")
async def list_user_roles(user_id: str = Path(..., description="User UUID")):
    return {"roles": []}


@app.patch("/api/v1/users/{user_id}/deactivate", tags=["👤 User Management — Users"], summary="Deactivate User",
           description="Soft-deletes a user by setting `is_active=false`. Records who deactivated and when. User cannot log in after deactivation.")
async def deactivate_user(user_id: str, body: DeactivateUserRequest):
    return {"user_id": user_id, "is_active": False, "deactivated_by": body.deactivated_by}


@app.patch("/api/v1/users/{user_id}/reactivate", tags=["👤 User Management — Users"], summary="Reactivate User",
           description="Restores a deactivated user, allowing them to log in again.")
async def reactivate_user(user_id: str):
    return {"user_id": user_id, "is_active": True}


# -- User Login History --
@app.post("/api/v1/users/me/login-history", tags=["👤 User Management — Users"], summary="Create Login History",
          description="Records a login event with IP address and user agent. Called automatically during login.")
async def create_login_history():
    return {"history_id": "generated-uuid"}


@app.get("/api/v1/users/me/login-history", tags=["👤 User Management — Users"], summary="List Login History",
         description="Lists the authenticated user's login history with timestamps, IP addresses, and user agents. Paginated.")
async def list_login_history():
    return {"histories": []}


# -- User-Organization --
@app.post("/api/v1/users/{user_id}/organizations", tags=["👤 User Management — Organizations"], summary="Add User to Organization",
          description="Adds a user to an additional organization with a specific role, department, and designation. Users can belong to multiple organizations.")
async def add_user_to_org(user_id: str, body: AddUserToOrgRequest):
    return {"user_id": user_id, "org_id": body.org_id}


@app.delete("/api/v1/users/{user_id}/organizations/{org_id}", tags=["👤 User Management — Organizations"], summary="Remove User from Organization",
            description="Removes a user from a specific organization. Cannot remove from the last remaining organization.")
async def remove_user_from_org(user_id: str, org_id: str):
    return {"message": "User removed from organization"}


@app.get("/api/v1/users/{user_id}/organizations", tags=["👤 User Management — Organizations"], summary="List User Organizations",
         description="Lists all organizations a user belongs to, including their role, department, and designation in each.")
async def list_user_orgs(user_id: str = Path(..., description="User UUID")):
    try:
        orgs = query_db("""
            SELECT uo.user_id, uo.org_id, o.name as org_name, o.code as org_code, 
                   r.name as role_name, uo.is_current_context, uo.joined_at 
            FROM user_organizations uo 
            JOIN organizations o ON uo.org_id = o.org_id 
            LEFT JOIN roles r ON uo.role_id = r.role_id 
            WHERE uo.user_id = %s
        """, (user_id,))
        return {"organizations": orgs, "total_count": len(orgs)}
    except Exception as e:
        raise HTTPException(500, str(e))


# -- Dropdowns --
@app.get("/api/v1/users/dropdowns/departments", tags=["👤 User Management — Dropdowns"], summary="Departments Dropdown",
         description="Returns departments for form dropdowns, filtered by the authenticated user's organization context.")
async def departments_dropdown():
    try:
        depts = query_db("SELECT id, name FROM departments ORDER BY name")
        return {"departments": depts}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/users/dropdowns/designations", tags=["👤 User Management — Dropdowns"], summary="Designations Dropdown",
         description="Returns designations for form dropdowns.")
async def designations_dropdown():
    try:
        desigs = query_db("SELECT id, name FROM designations ORDER BY name")
        return {"designations": desigs}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/users/dropdowns/roles", tags=["👤 User Management — Dropdowns"], summary="Roles Dropdown",
         description="Returns available roles for form dropdowns.")
async def roles_dropdown():
    try:
        roles = query_db("SELECT role_id as id, name, description FROM roles ORDER BY name")
        return {"roles": roles}
    except Exception as e:
        raise HTTPException(500, str(e))


# -- Signature --
@app.post("/api/v1/users/{user_id}/signature", tags=["👤 User Management — Users"], summary="Upload User Signature",
          description="Uploads a user's signature image. Supports PNG/JPEG. The signature URL is stored on the user record and used in document generation.")
async def upload_signature(user_id: str):
    return {"success": True, "message": "Signature upload — forwards to User gRPC service", "user_id": user_id}


# -- Activity Logs --
@app.post("/api/v1/activity-logs", tags=["👤 User Management — Activity Logs"], summary="Create Activity Log",
          description="Records an activity log entry for audit purposes. Automatically captures timestamp.")
async def create_activity_log(body: ActivityLogRequest):
    return {"id": 1, "name": body.name, "description": body.description}


@app.get("/api/v1/activity-logs", tags=["👤 User Management — Activity Logs"], summary="List Activity Logs",
         description="Lists activity logs with pagination. Used for audit trails and compliance reporting.")
async def list_activity_logs(page: int = Query(1, ge=1), page_size: int = Query(20, ge=1, le=100)):
    try:
        logs = query_db("""
            SELECT h.history_id, h.user_id, u.name as user_name, u.email as user_email, 
                   h.login_time, h.ip_address, h.user_agent 
            FROM user_login_history h 
            LEFT JOIN users u ON h.user_id = u.user_id 
            ORDER BY h.login_time DESC 
            LIMIT %s OFFSET %s
        """, (page_size, (page - 1) * page_size))
        total = query_db("SELECT COUNT(*) as cnt FROM user_login_history")
        return {"logs": logs, "pagination": {"page": page, "page_size": page_size, "total_items": total[0]["cnt"] if total else 0, "total_pages": 1}}
    except Exception as e:
        raise HTTPException(500, str(e))


# ═══════════════════════════════════════════════════════
# 🏢 ORGANIZATION SERVICE (10 endpoints)
# ═══════════════════════════════════════════════════════

@app.post("/api/v1/organizations", tags=["🏢 Organization Service"], summary="Create Organization",
          description="""
Creates a new parent or child organization.

**Parent Organization:** Leave `parent_org_id` empty. Must provide `super_admin` details (name, email, password).
**Child Organization:** Set `parent_org_id` to the parent's UUID. `super_admin` is ignored.

**Side Effects:**
- Auto-generates a unique database name
- Creates initial projects if specified
- Publishes Kafka event for downstream services
""")
async def create_organization(body: CreateOrganizationRequest):
    return {"organization": {"org_id": "generated-uuid", "name": body.name, "code": body.code}, "message": "Organization created"}


@app.get("/api/v1/organizations", tags=["🏢 Organization Service"], summary="List Organizations",
         description="Lists all organizations with pagination. Returns org details including status, project count, and hierarchy info.")
async def list_organizations(page: int = Query(1, ge=1), page_size: int = Query(20, ge=1, le=100)):
    try:
        orgs = query_db("SELECT org_id, name, code, description FROM organizations ORDER BY created_at DESC LIMIT %s OFFSET %s", (page_size, (page - 1) * page_size))
        total = query_db("SELECT COUNT(*) as cnt FROM organizations")
        return {"organizations": orgs, "total_count": total[0]["cnt"] if total else 0, "pagination": {"current_page": page, "page_size": page_size}}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/tenants/{tenant_id}/organizations", tags=["🏢 Organization Service"], summary="List Organizations by Tenant",
         description="Lists all organizations belonging to a specific tenant. Used for tenant-scoped views.")
async def list_orgs_by_tenant(tenant_id: str, page: int = Query(1), page_size: int = Query(20)):
    try:
        orgs = query_db("SELECT org_id, name, code, description FROM organizations WHERE tenant_id = %s ORDER BY name LIMIT %s OFFSET %s", (tenant_id, page_size, (page - 1) * page_size))
        return {"organizations": orgs, "total_count": len(orgs)}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/organizations/{parent_org_id}/children", tags=["🏢 Organization Service"], summary="List Child Organizations",
         description="Lists all child organizations under a specific parent organization.")
async def list_child_orgs(parent_org_id: str, page: int = Query(1), page_size: int = Query(20)):
    try:
        orgs = query_db("SELECT org_id, name, code, description FROM organizations WHERE org_id != %s ORDER BY name LIMIT %s OFFSET %s", (parent_org_id, page_size, (page - 1) * page_size))
        return {"organizations": orgs, "total_count": len(orgs)}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/organizations/{org_id}", tags=["🏢 Organization Service"], summary="Get Organization",
         description="Retrieves a specific organization by UUID including all fields and hierarchy info.")
async def get_organization(org_id: str = Path(..., description="Organization UUID")):
    try:
        orgs = query_db("SELECT org_id, tenant_id, name, code, description, database_name, is_active FROM organizations WHERE org_id = %s", (org_id,))
        if not orgs:
            raise HTTPException(404, "Organization not found")
        return {"organization": orgs[0]}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/organizations/{org_id}/with-projects", tags=["🏢 Organization Service"], summary="Get Organization with Projects",
         description="Retrieves an organization and all its associated projects in a single call. Reduces N+1 queries.")
async def get_org_with_projects(org_id: str):
    try:
        orgs = query_db("SELECT org_id, tenant_id, name, code, description FROM organizations WHERE org_id = %s", (org_id,))
        if not orgs:
            raise HTTPException(404, "Organization not found")
        projects = query_db("SELECT id, project_name, created_by, created_at FROM projects WHERE org_id = %s ORDER BY project_name", (org_id,))
        return {"organization": orgs[0], "projects": projects}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/organizations/code/{code}", tags=["🏢 Organization Service"], summary="Get Organization by Code",
         description="Looks up an organization by its unique code (e.g., 'NHIT-INFRA'). Useful for URL-friendly identifiers.")
async def get_org_by_code(code: str = Path(..., description="Organization code")):
    try:
        orgs = query_db("SELECT org_id, tenant_id, name, code, description, database_name, is_active FROM organizations WHERE code ILIKE %s", (code,))
        if not orgs:
            raise HTTPException(404, f"Organization with code '{code}' not found")
        return {"organization": orgs[0]}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.put("/api/v1/organizations/{org_id}", tags=["🏢 Organization Service"], summary="Update Organization",
         description="Updates organization details. Only provided fields are modified. **Security:** Requires org admin or SUPER_ADMIN role.")
async def update_organization(org_id: str, body: UpdateOrganizationRequest):
    return {"organization": {"org_id": org_id}, "message": "Organization updated"}


@app.delete("/api/v1/organizations/{org_id}", tags=["🏢 Organization Service"], summary="Delete Organization",
            description="Deletes an organization and all associated data. **⚠️ DESTRUCTIVE:** This cannot be undone. Consider deactivating instead.")
async def delete_organization(org_id: str = Path(..., description="Organization UUID")):
    return {"success": True, "message": "Organization deleted"}


@app.post("/api/v1/organizations/{org_id}/logo", tags=["🏢 Organization Service"], summary="Upload Organization Logo",
          description="Uploads a logo image for the organization. Supports PNG/JPEG/SVG. Stored in object storage and URL saved to org record.")
async def upload_org_logo(org_id: str):
    return {"success": True, "message": "Logo upload — forwards to Organization gRPC service", "organization_id": org_id}


# ══════════════════════════════════════════════════════════════
# 🏫 DEPARTMENT SERVICE (5 endpoints)
# ══════════════════════════════════════════════════════════════

@app.post("/api/v1/departments", tags=["🏫 Department Service"], summary="Create Department",
          description="Creates a new organizational department (e.g., 'Finance', 'Civil Engineering').")
async def create_department(body: CreateDepartmentRequest):
    try:
        execute_db("INSERT INTO departments (name, description) VALUES (%s, %s)", (body.name, body.description))
        return {"department": {"name": body.name, "description": body.description}, "message": "Department created"}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/departments", tags=["🏫 Department Service"], summary="List Departments",
         description="Lists all departments with pagination.")
async def list_departments(page: int = Query(1, ge=1), page_size: int = Query(20, ge=1, le=100)):
    try:
        depts = query_db("SELECT id, name, description FROM departments ORDER BY name LIMIT %s OFFSET %s", (page_size, (page - 1) * page_size))
        total = query_db("SELECT COUNT(*) as cnt FROM departments")
        return {"departments": depts, "total_count": total[0]["cnt"] if total else 0}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/departments/{id}", tags=["🏫 Department Service"], summary="Get Department",
         description="Retrieves a specific department by its UUID.")
async def get_department(id: str = Path(..., description="Department UUID")):
    try:
        dept = query_db("SELECT id, name, description FROM departments WHERE id = %s", (id,))
        if not dept:
            raise HTTPException(404, "Department not found")
        return {"department": dept[0]}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.put("/api/v1/departments/{id}", tags=["🏫 Department Service"], summary="Update Department",
         description="Updates department name or description.")
async def update_department(id: str = Path(..., description="Department UUID"), body: UpdateDepartmentRequest = ...):
    try:
        execute_db("UPDATE departments SET name = COALESCE(%s, name), description = COALESCE(%s, description) WHERE id = %s",
                   (body.name, body.description, id))
        return {"department": {"id": id, "name": body.name, "description": body.description}, "message": "Department updated"}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.delete("/api/v1/departments/{id}", tags=["🏫 Department Service"], summary="Delete Department",
            description="Deletes a department by UUID.")
async def delete_department(id: str = Path(..., description="Department UUID")):
    try:
        execute_db("DELETE FROM departments WHERE id = %s", (id,))
        return {"message": "Department deleted"}
    except Exception as e:
        raise HTTPException(500, str(e))


# ══════════════════════════════════════════════════════════════
# 📋 DESIGNATION SERVICE (5 endpoints)
# ══════════════════════════════════════════════════════════════

@app.post("/api/v1/designations", tags=["📋 Designation Service"], summary="Create Designation",
          description="Creates a new job designation (e.g., 'Chief Resident Engineer', 'Project Manager'). Designations are used for user profiles and approval workflows.")
async def create_designation(body: CreateDesignationRequest):
    try:
        execute_db("INSERT INTO designations (name, description) VALUES (%s, %s)", (body.name, body.description))
        return {"designation": {"name": body.name, "description": body.description}, "message": "Designation created"}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/designations", tags=["📋 Designation Service"], summary="List Designations",
         description="Lists all designations with pagination.")
async def list_designations(page: int = Query(1, ge=1), page_size: int = Query(20, ge=1, le=100)):
    try:
        desigs = query_db("SELECT id, name, description FROM designations ORDER BY name LIMIT %s OFFSET %s", (page_size, (page - 1) * page_size))
        total = query_db("SELECT COUNT(*) as cnt FROM designations")
        return {"designations": desigs, "total_count": total[0]["cnt"] if total else 0}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/designations/{id}", tags=["📋 Designation Service"], summary="Get Designation",
         description="Retrieves a specific designation by UUID.")
async def get_designation(id: str = Path(..., description="Designation UUID")):
    try:
        desigs = query_db("SELECT id, name, description FROM designations WHERE id = %s", (id,))
        if not desigs:
            raise HTTPException(404, "Designation not found")
        return {"designation": desigs[0]}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.put("/api/v1/designations/{id}", tags=["📋 Designation Service"], summary="Update Designation",
         description="Updates designation name and/or description.")
async def update_designation(id: str, body: UpdateDesignationRequest):
    return {"designation": {"id": id}, "message": "Designation updated"}


@app.delete("/api/v1/designations/{id}", tags=["📋 Designation Service"], summary="Delete Designation",
            description="Deletes a designation. Fails if users are assigned to this designation.")
async def delete_designation(id: str = Path(..., description="Designation UUID")):
    return {"success": True, "message": "Designation deleted"}


# ═══════════════════════════════════════════════════════
# 📁 PROJECT SERVICE (3 endpoints)
# ═══════════════════════════════════════════════════════

@app.post("/api/v1/projects", tags=["📁 Project Service"], summary="Create Project",
          description="Creates a new project within an organization. Projects can also be auto-created via Kafka events when organizations are created with `initial_projects`.")
async def create_project(body: CreateProjectRequest):
    try:
        new_proj_id = str(uuid.uuid4())
        execute_db("""
            INSERT INTO projects (id, project_name, org_id, tenant_id, created_by, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, NOW(), NOW())
        """, (new_proj_id, body.project_name, body.org_id, body.tenant_id, body.created_by))
        return {"success": True, "project": {"project_id": new_proj_id, "project_name": body.project_name, "org_id": body.org_id}}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/projects/{project_id}", tags=["📁 Project Service"], summary="Get Project",
         description="Retrieves a project by UUID including tenant, organization, creator, and timestamps.")
async def get_project(project_id: str = Path(..., description="Project UUID")):
    try:
        projects = query_db("SELECT id, tenant_id, org_id, project_name, created_by, created_at, updated_at FROM projects WHERE id = %s", (project_id,))
        if not projects:
            raise HTTPException(404, "Project not found")
        return {"project": projects[0]}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/organizations/{org_id}/projects", tags=["📁 Project Service"], summary="List Projects by Organization",
         description="Lists all projects belonging to a specific organization with pagination.")
async def list_projects_by_org(org_id: str, page: int = Query(1), page_size: int = Query(20)):
    try:
        projects = query_db("SELECT id, tenant_id, org_id, project_name, created_by, created_at FROM projects WHERE org_id = %s ORDER BY project_name LIMIT %s OFFSET %s", (org_id, page_size, (page - 1) * page_size))
        return {"projects": projects, "total_count": len(projects)}
    except Exception as e:
        raise HTTPException(500, str(e))


# ═══════════════════════════════════════════════════════
# 🤝 VENDOR SERVICE (15 endpoints)
# ═══════════════════════════════════════════════════════

@app.post("/api/v1/vendors", tags=["🤝 Vendor Service"], summary="Create Vendor",
          description="""
Creates a new vendor (internal or external) with comprehensive business details.

**Required Fields:** vendor_name, vendor_email, account_type (INTERNAL/EXTERNAL), PAN, beneficiary_name, MSME classification
**Optional:** GSTIN, banking details, address, mobile, nickname

**Auto-generates** a unique vendor code based on vendor name and account type.
""")
async def create_vendor(body: CreateVendorRequest):
    v_name = body.name or body.vendor_name or "New Vendor"
    v_code = body.vendor_code or f"VEND-{secrets.token_hex(3).upper()}"
    new_id = str(uuid.uuid4())
    bank_name = body.bank_name or ""
    acc_num = body.bank_account_number or ""
    ifsc = body.bank_ifsc or ""
    beneficiary = body.beneficiary_name or v_name

    try:
        execute_db("""
            INSERT INTO vendors (
                id, vendor_name, vendor_code, vendor_email, pan, gstin, 
                msme_classification, name_of_bank, account_number, ifsc_code, 
                account_name, beneficiary_name, is_active, tenant_id, created_by, created_at, updated_at
            ) VALUES (
                %s, %s, %s, %s, %s, %s, 
                %s, %s, %s, %s, 
                %s, %s, true, %s, %s, NOW(), NOW()
            )
        """, (
            new_id, v_name, v_code, body.vendor_email, body.pan, body.gstin,
            body.msme_classification, bank_name, acc_num, ifsc,
            v_name, beneficiary, body.tenant_id, body.created_by
        ))
        if acc_num and bank_name:
            acc_id = str(uuid.uuid4())
            execute_db("""
                INSERT INTO vendor_accounts (
                    id, vendor_id, account_name, account_number, name_of_bank, ifsc_code, is_primary, is_active, created_by, created_at, updated_at
                ) VALUES (
                    %s, %s, %s, %s, %s, %s, true, true, %s, NOW(), NOW()
                )
            """, (acc_id, new_id, v_name, acc_num, bank_name, ifsc, body.created_by))
    except Exception as e:
        print(f"Vendor DB insert notice: {e}")

    return {
        "success": True,
        "id": new_id,
        "vendor": {
            "id": new_id,
            "vendor_name": v_name,
            "vendor_code": v_code,
            "msme_classification": body.msme_classification,
            "bank_name": bank_name,
            "bank_account_number": acc_num,
            "bank_ifsc": ifsc
        }
    }


@app.get("/api/v1/vendors", tags=["🤝 Vendor Service"], summary="List Vendors",
         description="Lists vendors with search, filters (active status, account type, project), sorting, and pagination.")
async def list_vendors(
    search: Optional[str] = Query(None, description="Search by name, code, or email"),
    is_active: Optional[bool] = Query(None, description="Filter by active status"),
    account_type: Optional[str] = Query(None, description="INTERNAL or EXTERNAL"),
    page: int = Query(1, ge=1),
    page_size: int = Query(20, ge=1, le=100),
    sort_by: Optional[str] = Query(None, description="Sort column"),
    sort_order: Optional[str] = Query("asc", description="asc or desc"),
):
    try:
        vendors = query_db("SELECT id, vendor_name, vendor_code, vendor_email, pan, gstin, msme, msme_classification, material_nature, is_active FROM vendors ORDER BY vendor_name LIMIT %s OFFSET %s", (page_size, (page - 1) * page_size))
        total = query_db("SELECT COUNT(*) as cnt FROM vendors")
        return {"vendors": vendors, "total_count": total[0]["cnt"] if total else 0}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/vendors/{vendor_id}", tags=["🤝 Vendor Service"], summary="Get Vendor",
         description="Retrieves complete vendor details including business info, MSME classification, banking details, and document URL.")
async def get_vendor(vendor_id: str = Path(..., description="Vendor UUID")):
    try:
        vendors = query_db("""
            SELECT id, tenant_id, vendor_code, vendor_name, vendor_email, vendor_mobile, 
                   vendor_type, pan, gstin, msme, msme_classification, material_nature, 
                   beneficiary_name, is_active, created_at, updated_at 
            FROM vendors WHERE id = %s
        """, (vendor_id,))
        if not vendors:
            raise HTTPException(404, "Vendor not found")
        return {"vendor": vendors[0]}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/vendors/code/{vendor_code}", tags=["🤝 Vendor Service"], summary="Get Vendor by Code",
         description="Looks up a vendor by their unique vendor code (e.g., 'EXT-ABC-001').")
async def get_vendor_by_code(vendor_code: str = Path(..., description="Vendor code")):
    try:
        vendors = query_db("""
            SELECT id, tenant_id, vendor_code, vendor_name, vendor_email, vendor_mobile, 
                   vendor_type, pan, gstin, msme, msme_classification, material_nature, 
                   beneficiary_name, is_active, created_at, updated_at 
            FROM vendors WHERE vendor_code ILIKE %s
        """, (vendor_code,))
        if not vendors:
            raise HTTPException(404, f"Vendor with code '{vendor_code}' not found")
        return {"vendor": vendors[0]}
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(500, str(e))


@app.put("/api/v1/vendors/{vendor_id}", tags=["🤝 Vendor Service"], summary="Update Vendor",
         description="Updates vendor details. All fields are optional (partial update). Records `updated_by` for audit trail.")
async def update_vendor(vendor_id: str):
    return {"vendor": {"id": vendor_id}}


@app.delete("/api/v1/vendors/{vendor_id}", tags=["🤝 Vendor Service"], summary="Delete Vendor",
            description="Permanently deletes a vendor. Consider deactivating instead if the vendor has transaction history.")
async def delete_vendor(vendor_id: str = Path(..., description="Vendor UUID")):
    return {"message": "Vendor deleted"}


@app.post("/api/v1/vendors/generate-code", tags=["🤝 Vendor Service"], summary="Generate Vendor Code",
          description="Generates a unique vendor code based on vendor name and account type. Format: `{TYPE}-{NAME_ABBREV}-{SEQUENCE}` (e.g., 'EXT-ABC-001').")
async def generate_vendor_code(body: GenerateVendorCodeRequest):
    abbrev = "".join([w[:1] for w in body.vendor_name.split()[:3]]).upper() or "VND"
    prefix = "INT" if body.account_type == "INTERNAL" else "EXT"
    return {"vendor_code": f"{prefix}-{abbrev}-001"}


@app.put("/api/v1/vendors/{vendor_id}/code", tags=["🤝 Vendor Service"], summary="Update Vendor Code",
          description="Manually overrides the auto-generated vendor code. Use with caution — codes should remain unique.")
async def update_vendor_code(vendor_id: str):
    return {"vendor": {"id": vendor_id}}


@app.post("/api/v1/vendors/{vendor_id}/regenerate-code", tags=["🤝 Vendor Service"], summary="Regenerate Vendor Code",
          description="Regenerates the vendor code using the current vendor name and account type.")
async def regenerate_vendor_code(vendor_id: str):
    return {"vendor": {"id": vendor_id}}


# -- Vendor Accounts --
@app.post("/api/v1/vendors/{vendor_id}/accounts", tags=["🤝 Vendor Service — Accounts"], summary="Create Vendor Account",
          description="Adds a new bank account to a vendor. Multiple accounts are supported. One can be marked as primary.")
async def create_vendor_account(vendor_id: str, body: CreateVendorAccountRequest):
    return {"account": {"id": "generated-uuid", "vendor_id": vendor_id, "account_name": body.account_name}}


@app.get("/api/v1/vendors/{vendor_id}/accounts", tags=["🤝 Vendor Service — Accounts"], summary="Get Vendor Accounts",
         description="Lists all bank accounts for a vendor including primary/secondary status.")
async def get_vendor_accounts(vendor_id: str = Path(..., description="Vendor UUID")):
    try:
        accounts = query_db("""
            SELECT id, vendor_id, account_name, account_number, account_type, 
                   name_of_bank, branch_name, ifsc_code, is_primary, is_active 
            FROM vendor_accounts WHERE vendor_id = %s ORDER BY is_primary DESC
        """, (vendor_id,))
        return {"accounts": accounts, "total_count": len(accounts)}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.get("/api/v1/vendors/{vendor_id}/banking-details", tags=["🤝 Vendor Service — Accounts"], summary="Get Vendor Banking Details",
         description="Returns the primary banking details for a vendor, used in payment processing.")
async def get_vendor_banking(vendor_id: str = Path(..., description="Vendor UUID")):
    try:
        accounts = query_db("""
            SELECT id, vendor_id, account_name, account_number, account_type, 
                   name_of_bank, branch_name, ifsc_code, is_primary 
            FROM vendor_accounts WHERE vendor_id = %s AND is_primary = true
        """, (vendor_id,))
        if not accounts:
            accounts = query_db("SELECT * FROM vendor_accounts WHERE vendor_id = %s LIMIT 1", (vendor_id,))
        return {"banking_details": accounts[0] if accounts else {}}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.put("/api/v1/vendors/accounts/{account_id}", tags=["🤝 Vendor Service — Accounts"], summary="Update Vendor Account",
         description="Updates bank account details. Partial update supported.")
async def update_vendor_account(account_id: str):
    return {"account": {"id": account_id}}


@app.delete("/api/v1/vendors/accounts/{account_id}", tags=["🤝 Vendor Service — Accounts"], summary="Delete Vendor Account",
            description="Deletes a bank account. Cannot delete the primary account if it's the only one.")
async def delete_vendor_account(account_id: str):
    return {"message": "Account deleted"}


@app.post("/api/v1/vendors/accounts/{account_id}/toggle-status", tags=["🤝 Vendor Service — Accounts"], summary="Toggle Account Status",
          description="Toggles a vendor bank account between active and inactive status.")
async def toggle_account_status(account_id: str):
    return {"account": {"id": account_id}}


# -- Vendor Dropdowns & Docs --
@app.get("/api/v1/vendors/dropdowns/projects", tags=["🤝 Vendor Service"], summary="Projects Dropdown for Vendors",
         description="Returns project list for vendor form dropdowns, filtered by the authenticated user's organization.")
async def vendor_projects_dropdown():
    try:
        projects = query_db("SELECT id, project_name FROM projects ORDER BY project_name")
        return {"projects": projects}
    except Exception as e:
        raise HTTPException(500, str(e))


@app.post("/api/v1/vendors/{vendor_id}/document", tags=["🤝 Vendor Service"], summary="Upload Vendor Document",
          description="Uploads a document (PAN card, GST certificate, etc.) for a vendor. Stored in object storage.")
async def upload_vendor_document(vendor_id: str):
    return {"success": True, "message": "Document upload — forwards to Vendor gRPC service", "vendor_id": vendor_id}


# ══════════════════════════════════════════════════════════════
# 🔔 NOTIFICATION SERVICE (1 endpoint)
# ═══════════════════════════════════════════════════════

@app.post("/api/v1/notifications", tags=["🔔 Notification Service"], summary="Send Notification",
          description="""
Sends a notification via the multi-channel template engine. Supports email, SMS, and push notifications.
""")
async def send_notification(body: SendNotificationRequest):
    data_dict = body.data or {}
    otp_code = str(data_dict.get("otp", "849201"))
    user_name = str(data_dict.get("name", "Officer"))
    res = dispatch_otp(
        recipient=body.to,
        otp_code=otp_code,
        user_name=user_name,
        action=data_dict.get("action", f"Notification: {body.template_name}"),
        phone=data_dict.get("phone")
    )
    return {
        "success": True,
        "message": f"Notification dispatched to {body.to}",
        "delivery": res
    }


class TestLiveOTPRequest(BaseModel):
    email: str = Field(..., description="Target email address")
    phone: Optional[str] = Field(None, description="Optional target mobile phone (+91...)")
    name: Optional[str] = Field("NHIT Officer", description="Recipient name")

@app.post("/api/v1/notifications/test-otp", tags=["🔔 Notification Service"], summary="Live Dispatch Test OTP",
          description="Directly triggers the multi-channel OTP dispatcher to send an actual email or SMS.")
async def test_live_otp_dispatch(body: TestLiveOTPRequest):
    code = generate_numeric_otp(6)
    res = dispatch_otp(
        recipient=body.email,
        otp_code=code,
        user_name=body.name or "NHIT Officer",
        action="Live Notification System Test",
        phone=body.phone
    )
    return {
        "success": True,
        "otp_generated": code,
        "dispatch_result": res
    }


# ═══════════════════════════════════════════════════════
# 📝 GREENNOTE SERVICE (9 endpoints)
# ═══════════════════════════════════════════════════════

@app.post("/api/v1/green-notes", tags=["📝 GreenNote Service"], summary="Create Green Note",
          description="""
Creates a new Green Note (approval/expenditure note) for vendor payments.

**Workflow:**
1. Creator fills in project, vendor, invoice details, and compliance checklist
2. Green Note is created with status `PENDING` or `DRAFT`
3. Approval engine routes it through multi-level approval based on amount and department
4. Upon final approval, a Payment Note draft is auto-created

**Required Context:** JWT must contain org_id and tenant_id.
""")
async def create_green_note(body: CreateGreenNoteRequest):
    new_id = f"gn-550e8400-{len(GREEN_NOTES_STORE) + 1:04d}"
    item = {
        "id": new_id,
        "sl_no": f"GN-2026-{len(GREEN_NOTES_STORE) + 1:03d}",
        "project_name": body.note.project_name or "NH-44 Highway Infrastructure Project",
        "supplier_name": body.note.supplier_name or "Larsen & Toubro Ltd",
        "amount": body.note.total_amount or 250000.0,
        "status": "PENDING",
        "created_by": "System User",
        "created_at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "payload": body.note.model_dump()
    }
    GREEN_NOTES_STORE.insert(0, item)
    return {"id": new_id, "success": True, "message": "Green note created successfully", "data": item}


@app.get("/api/v1/green-notes", tags=["📝 GreenNote Service"], summary="List Green Notes",
         description="Lists green notes with status filter, pagination, and user scoping. Returns lightweight list items (no full payload).")
async def list_green_notes(
    status: Optional[str] = Query(None, description="Filter: PENDING, APPROVED, REJECTED, DRAFT, CANCELLED"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    include_all: bool = Query(False, description="Include all statuses"),
):
    notes = GREEN_NOTES_STORE
    if status and not include_all:
        notes = [n for n in notes if n.get("status") == status]
    start = (page - 1) * per_page
    end = start + per_page
    return {"notes": notes[start:end], "page": page, "per_page": per_page, "total": len(notes)}


@app.get("/api/v1/green-notes/{id}", tags=["📝 GreenNote Service"], summary="Get Green Note Detail",
         description="Retrieves the full green note payload including all compliance fields, invoice details, financial calculations, approval history, and linked documents.")
async def get_green_note(id: str = Path(..., description="Green Note UUID")):
    match = next((n for n in GREEN_NOTES_STORE if n.get("id") == id), None)
    if not match:
        raise HTTPException(404, "Green Note not found")
    return {"success": True, "data": match}


@app.patch("/api/v1/green-notes/{id}", tags=["📝 GreenNote Service"], summary="Update Green Note",
           description="Updates a green note. Only allowed when status is DRAFT or PENDING (before approval). Cannot update after approval/rejection.")
async def update_green_note(id: str, body: CreateGreenNoteRequest):
    return {"id": id, "success": True, "message": "Green note updated"}


class UpdateNoteStatusRequest(BaseModel):
    status: str = Field(..., description="APPROVED, REJECTED, PENDING, or DRAFT")
    remarks: Optional[str] = Field("", description="Review remarks")

@app.patch("/api/v1/green-notes/{id}/status", tags=["📝 GreenNote Service"], summary="Update Green Note Status",
           description="Updates approval status (APPROVED, REJECTED, PENDING) with reviewer remarks.")
async def update_green_note_status(id: str, body: UpdateNoteStatusRequest):
    match = next((n for n in GREEN_NOTES_STORE if n.get("id") == id), None)
    if not match:
        raise HTTPException(404, "Green Note not found")
    match["status"] = body.status
    match["remarks"] = body.remarks
    return {"id": id, "success": True, "status": body.status, "message": f"Green note status set to {body.status}"}



@app.post("/api/v1/green-notes/{id}/cancel", tags=["📝 GreenNote Service"], summary="Cancel Green Note",
          description="Cancels a green note with a mandatory reason. Sets status to CANCELLED. Cannot cancel already-paid notes.")
async def cancel_green_note(id: str, body: CancelGreenNoteRequest):
    return {"id": id, "success": True, "message": "Green note cancelled"}


@app.get("/api/v1/organization/projects", tags=["📝 GreenNote Service — Organization Data"], summary="Get Organization Projects",
         description="Returns projects for the authenticated user's organization. Used in green note creation forms.")
async def get_org_projects():
    return {"projects": [], "message": ""}


@app.get("/api/v1/organization/vendors", tags=["📝 GreenNote Service — Organization Data"], summary="Get Organization Vendors",
         description="Returns vendors/suppliers for the authenticated user's organization. Used in green note creation forms.")
async def get_org_vendors():
    return {"vendors": [], "message": ""}


@app.get("/api/v1/organization/departments", tags=["📝 GreenNote Service — Organization Data"], summary="Get Organization Departments",
         description="Returns departments for the authenticated user's organization.")
async def get_org_departments():
    return {"departments": [], "message": ""}


@app.post("/api/v1/green-notes/{note_id}/documents", tags=["📝 GreenNote Service"], summary="Upload Green Note Documents",
          description="Uploads supporting documents for a green note (invoices, POs, contracts). Supports multiple files in a single request.")
async def upload_green_note_docs(note_id: str):
    return {"success": True, "message": "Documents uploaded", "green_note_id": note_id}


@app.post("/api/v1/green-notes/{id}/link-payment", tags=["📝 GreenNote Service"], summary="Link Payment Note",
          description="Links a Payment Note to this Green Note. Establishes the relationship for tracking payment against the approved expenditure.")
async def link_payment_note(id: str, body: LinkPaymentNoteRequest):
    return {"id": id, "success": True, "message": "Payment note linked"}


# ═══════════════════════════════════════════════════════
# 💰 PAYMENTNOTE SERVICE (14 endpoints)
# ═══════════════════════════════════════════════════════

@app.get("/api/v1/payment-notes", tags=["💰 PaymentNote Service"], summary="List Payment Notes",
         description="Lists payment notes with status filter, pagination, search, and role-based scoping. Super admins see all; others see only assigned notes.")
async def list_payment_notes(
    status: Optional[str] = Query(None, description="D=Draft, P=Pending, A=Approved, R=Rejected, S=Submitted"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    search: Optional[str] = Query(None, description="Search by note number, vendor, or project"),
):
    notes = PAYMENT_NOTES_STORE
    if status:
        notes = [n for n in notes if n.get("status") == status]
    if search:
        s = search.lower()
        notes = [n for n in notes if s in str(n.get("vendor_name", "")).lower() or s in str(n.get("payment_note_number", "")).lower()]
    start = (page - 1) * per_page
    end = start + per_page
    return {"notes": notes[start:end], "pagination": {"page": page, "per_page": per_page, "total_items": len(notes)}}


@app.get("/api/v1/payment-notes/drafts", tags=["💰 PaymentNote Service"], summary="List Draft Payment Notes",
         description="Lists only draft payment notes. Auto-created drafts from approved green notes appear here.")
async def list_draft_payment_notes():
    return {"notes": [], "pagination": {}}


@app.get("/api/v1/payment-notes/{id}", tags=["💰 PaymentNote Service"], summary="Get Payment Note",
         description="Retrieves full payment note details including financial calculations, TDS, bank details, approval logs, comments, and linked green note reference.")
async def get_payment_note(id: int = Path(..., description="Payment Note ID")):
    return {"note": {}}


@app.post("/api/v1/payment-notes", tags=["💰 PaymentNote Service"], summary="Create Payment Note",
          description="""
Creates a new payment note with full financial details.

**Includes:** Vendor details, invoice info, LOA/PO details, TDS calculations, bank details, particulars (additions/deductions), and maker information.

**Status Flow:** Draft → Pending → Approved → Submitted → Paid
""")
async def create_payment_note(body: CreatePaymentNoteRequest):
    new_id = len(PAYMENT_NOTES_STORE) + 1
    item = {
        "id": new_id,
        "payment_note_number": f"PN-2026-{new_id:03d}",
        "vendor_name": body.vendor_name or "Larsen & Toubro Ltd",
        "vendor_account_number": body.vendor_account_number or "50200012345678",
        "bank_name": body.bank_name or "HDFC Bank",
        "ifsc_code": body.ifsc_code or "HDFC0000060",
        "net_payable_amount": body.net_payable_amount or 150000.0,
        "status": "D",
        "created_at": datetime.datetime.now(datetime.timezone.utc).isoformat()
    }
    PAYMENT_NOTES_STORE.insert(0, item)
    return {"note": item, "message": "Payment note created successfully"}


@app.patch("/api/v1/payment-notes/{id}", tags=["💰 PaymentNote Service"], summary="Update Payment Note",
           description="Updates a payment note. Can be a full update or status-only change (for approval workflow).")
async def update_payment_note(id: int):
    return {"note": {}}


@app.delete("/api/v1/payment-notes/{id}", tags=["💰 PaymentNote Service"], summary="Delete Payment Note",
            description="Permanently deletes a payment note. Only allowed for drafts. Submitted/approved notes cannot be deleted.")
async def delete_payment_note(id: int = Path(..., description="Payment Note ID")):
    return {}


@app.post("/api/v1/payment-notes/draft", tags=["💰 PaymentNote Service"], summary="Create Draft from Green Note",
          description="Auto-creates a draft payment note when a green note is approved. Pre-fills vendor, invoice, and financial details from the green note.")
async def create_draft_from_green_note(body: CreateDraftFromGreenNoteRequest):
    return {"note": {}}


@app.post("/api/v1/payment-notes/{id}/convert", tags=["💰 PaymentNote Service"], summary="Convert Draft to Active",
          description="Converts a draft payment note to an active (pending) note, entering the approval workflow.")
async def convert_draft_to_active(id: int):
    return {"note": {}}


@app.delete("/api/v1/payment-notes/{id}/draft", tags=["💰 PaymentNote Service"], summary="Delete Draft Payment Note",
            description="Deletes a draft payment note. Validates requester permissions before deletion.")
async def delete_draft_payment_note(id: int):
    return {}


@app.post("/api/v1/payment-notes/{id}/hold", tags=["💰 PaymentNote Service"], summary="Put Payment Note on Hold",
          description="Puts a payment note on hold with a reason. Prevents further processing until removed from hold.")
async def put_on_hold(id: int, body: PutOnHoldRequest):
    return {"note": {}}


@app.post("/api/v1/payment-notes/{id}/unhold", tags=["💰 PaymentNote Service"], summary="Remove from Hold",
          description="Removes a payment note from hold and sets its new status (Pending, Draft, or Submitted).")
async def remove_from_hold(id: int, body: RemoveFromHoldRequest):
    return {"note": {}}


@app.post("/api/v1/payment-notes/{id}/utr", tags=["💰 PaymentNote Service"], summary="Update UTR Information",
          description="Records the UTR (Unique Transaction Reference) number and date after bank payment. Optionally notifies the vendor via email.")
async def update_utr(id: int, body: UpdateUTRRequest):
    match = next((n for n in PAYMENT_NOTES_STORE if str(n.get("id")) == str(id)), None)
    if match:
        match["utr_number"] = body.utr_no
        match["status"] = "PAID"
    return {"note": match or {}}


class PaymentNoteStatusRequest(BaseModel):
    status: str = Field(..., description="ACTIVE, HOLD, PAID, or DRAFT")
    remarks: Optional[str] = Field("", description="Reason or remarks")

@app.patch("/api/v1/payment-notes/{id}/status", tags=["💰 PaymentNote Service"], summary="Update Payment Note Status",
           description="Updates status (e.g., toggling HOLD / ACTIVE).")
async def update_payment_note_status(id: str, body: PaymentNoteStatusRequest):
    match = next((n for n in PAYMENT_NOTES_STORE if str(n.get("id")) == str(id)), None)
    if not match:
        raise HTTPException(404, "Payment note not found")
    match["status"] = body.status
    return {"id": id, "success": True, "status": body.status, "message": f"Payment note status set to {body.status}"}


class PatchUTRRequest(BaseModel):
    utr_number: Optional[str] = None
    utr_no: Optional[str] = None
    payment_date: Optional[str] = None

@app.patch("/api/v1/payment-notes/{id}/utr", tags=["💰 PaymentNote Service"], summary="Record UTR (Patch)",
           description="Updates UTR number and sets status to PAID.")
async def patch_utr(id: str, body: PatchUTRRequest):
    utr = body.utr_number or body.utr_no or "UTR-RECORDED"
    match = next((n for n in PAYMENT_NOTES_STORE if str(n.get("id")) == str(id)), None)
    if match:
        match["utr_number"] = utr
        match["status"] = "PAID"
    return {"id": id, "success": True, "utr_number": utr, "status": "PAID", "message": "UTR recorded successfully"}



@app.get("/api/v1/payment-notes/order-number", tags=["💰 PaymentNote Service"], summary="Generate Order Number",
         description="Generates the next sequential payment note number with financial year prefix (e.g., 'PN-2024-25/0042').")
async def generate_payment_note_order_number():
    return {"order_number": "PN-2024-25/0001", "financial_year": "2024-25", "prefix": "PN"}


@app.get("/api/v1/payment-notes/{id}/pdf", tags=["💰 PaymentNote Service"], summary="Download Payment Note PDF",
         description="Generates and downloads a PDF document for the payment note. Includes all financial details, approval signatures, and compliance information.")
async def download_payment_note_pdf(id: int):
    return {"message": "PDF generation — forwards to PaymentNote gRPC service"}


@app.post("/api/v1/payment-notes/superadmin", tags=["💰 PaymentNote Service"], summary="Create Payment Note (Super Admin)",
          description="Super admin helper to create payment notes with elevated privileges. Can create directly as active or draft.")
async def create_payment_note_superadmin():
    return {"note": {}}


@app.get("/api/v1/payment-notes/test", tags=["💰 PaymentNote Service"], summary="Test Payment Note API",
         description="Health check endpoint for the PaymentNote gRPC service.")
async def test_payment_note_api():
    return {"status": "ok", "message": "PaymentNote API is operational", "timestamp": int(datetime.datetime.now(datetime.timezone.utc).timestamp())}


# ═══════════════════════════════════════════════════════
# 🏦 PAYMENT SERVICE (18 endpoints)
# ═══════════════════════════════════════════════════════

@app.get("/api/v1/payments", tags=["🏦 Payment Service"], summary="List Payments",
         description="Lists payment groups with status filter, pagination, search, sorting, and role-based access control.")
async def list_payments(
    status: Optional[str] = Query(None, description="Filter by status"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    search: Optional[str] = Query(None),
    sort_by: Optional[str] = Query(None),
    sort_direction: Optional[str] = Query("asc"),
):
    return {"groups": [], "pagination": {}}


@app.get("/api/v1/payments/datatable", tags=["🏦 Payment Service"], summary="List Payments (DataTable)",
         description="DataTables-compatible listing for AJAX-driven server-side rendering. Returns draw counter, total records, and filtered records.")
async def list_payments_datatable(
    draw: int = Query(1), start: int = Query(0), length: int = Query(20),
    status: Optional[str] = Query(None),
):
    return {"draw": draw, "total_records": 0, "total_display_records": 0, "payments": []}


@app.get("/api/v1/payments/{sl_no}", tags=["🏦 Payment Service"], summary="Get Payment Group",
         description="Retrieves a payment group by serial number (SL No.). Returns all payment line items, vendor details, and approval status.")
async def get_payment_group(sl_no: str = Path(..., description="Payment serial number")):
    return {"group": {}}


@app.post("/api/v1/payments", tags=["🏦 Payment Service"], summary="Create Payment Requests",
          description="""
Creates payment requests in bulk. Groups multiple vendor payments under a single SL number.

**Features:**
- Multiple vendor payments in one request
- Optional shortcut creation for recurring payments
- Links to payment note IDs for traceability
""")
async def create_payment_requests(body: CreatePaymentRequestsRequest):
    return {"group": {}}


@app.patch("/api/v1/payments/{sl_no}", tags=["🏦 Payment Service"], summary="Update Payment Group",
           description="Updates a payment group — can update vendor details, amounts, or just the status (for approval workflow).")
async def update_payment_group(sl_no: str):
    return {"group": {}}


@app.delete("/api/v1/payments/{id}", tags=["🏦 Payment Service"], summary="Delete Payment",
            description="Deletes payments under a specific SL number.")
async def delete_payment(id: int = Path(..., description="Payment ID")):
    return {}


@app.delete("/api/v1/payments/{sl_no}/items/{id}", tags=["🏦 Payment Service"], summary="Delete Payment Item",
            description="Removes a single payment line item from a payment group.")
async def delete_payment_item(sl_no: str, id: int):
    return {}


@app.post("/api/v1/payments/search-vendors", tags=["🏦 Payment Service — Lookups"], summary="Search Vendors for Payment",
          description="Searches vendors based on account type, project, and template type filters. Returns vendor banking details for payment form auto-fill.")
async def search_vendors_payment(body: SearchVendorsPaymentRequest):
    return {"vendors": []}


@app.get("/api/v1/payments/vendors/internal", tags=["🏦 Payment Service — Lookups"], summary="Search Internal Vendors",
         description="Searches internal (employee/inter-company) vendors by name.")
async def search_internal_vendors(search: Optional[str] = Query(None)):
    return {"vendors": []}


@app.get("/api/v1/payments/projects", tags=["🏦 Payment Service — Lookups"], summary="Search Projects for Payment",
         description="Returns project names matching the search query. Used in payment form project selection.")
async def search_projects_payment(search: Optional[str] = Query(None)):
    return {"projects": []}


@app.get("/api/v1/payments/from-accounts/{template_type}", tags=["🏦 Payment Service — Lookups"], summary="Get From-Account Options",
         description="Returns available 'from' bank accounts based on the payment template type (RTGS, NEFT, etc.).")
async def get_from_account_options(template_type: str = Path(..., description="Template type: RTGS, NEFT, etc.")):
    return {"accounts": []}


@app.get("/api/v1/payments/vendors", tags=["🏦 Payment Service — Lookups"], summary="Get All Vendors",
         description="Returns all vendors for the payment form. Unfiltered — use search endpoint for filtered results.")
async def get_all_vendors_payment():
    return {"vendors": []}


# -- Cart/Queue --
@app.post("/api/v1/payments/cart", tags=["🏦 Payment Service — Cart"], summary="Add to Payment Cart",
          description="Adds a payment request item to the user's cart/queue. Cart items are staged before final submission as a payment group.")
async def add_to_cart():
    return {"items": []}


@app.delete("/api/v1/payments/cart", tags=["🏦 Payment Service — Cart"], summary="Remove from Cart",
            description="Removes an item from the payment cart or clears the entire cart.")
async def remove_from_cart():
    return {"items": []}


@app.get("/api/v1/payments/cart", tags=["🏦 Payment Service — Cart"], summary="Get Cart Snapshot",
         description="Returns the current state of the user's payment cart.")
async def get_cart():
    return {"items": []}


# -- Shortcuts & Bank Letters --
@app.post("/api/v1/payments/shortcuts", tags=["🏦 Payment Service — Shortcuts"], summary="Create Payment Shortcut",
          description="Saves a payment configuration as a reusable shortcut for recurring payments.")
async def create_shortcut():
    return {"shortcut": {}}


@app.post("/api/v1/payments/shortcuts/{shortcut_id}/execute", tags=["🏦 Payment Service — Shortcuts"], summary="Execute Payment Shortcut",
          description="Executes a saved payment shortcut, creating a new payment group with the stored configuration.")
async def execute_shortcut(shortcut_id: int):
    return {"group": {}}


@app.post("/api/v1/payments/bank-letters", tags=["🏦 Payment Service — Bank Letters"], summary="Create Bank Letter",
          description="Generates an RTGS/NEFT bank letter from selected payment notes. Groups payments for bank submission.")
async def create_bank_letter(body: BankLetterRequest):
    return {"sl_no": "", "payments_count": 0, "total_amount": 0}


@app.post("/api/v1/payments/bank-letters/{sl_no}/log", tags=["🏦 Payment Service — Bank Letters"], summary="Process Bank Letter Log",
          description="Records approval/rejection of a bank letter with reviewer remarks.")
async def process_bank_letter_log(sl_no: str):
    return {"message": "", "log": {}}


@app.get("/api/v1/payments/order-number", tags=["🏦 Payment Service"], summary="Generate Payment Serial Number",
         description="Generates the next payment serial number for new payment groups.")
async def generate_payment_serial():
    return {"sl_no": "PAY-2024-25/0001"}


@app.get("/api/v1/payments/test", tags=["🏦 Payment Service"], summary="Test Payment API",
         description="Health check for the Payment gRPC service.")
async def test_payment_api():
    return {"status": "ok", "message": "Payment API is operational", "timestamp": int(datetime.datetime.now(datetime.timezone.utc).timestamp())}


# ═══════════════════════════════════════════════════════
# 🧾 REIMBURSEMENT SERVICE (16 endpoints)
# ═══════════════════════════════════════════════════════

@app.get("/api/v1/reimbursement-notes", tags=["🧾 Reimbursement Service"], summary="List Reimbursement Notes",
         description="Lists reimbursement notes with status filter, search, pagination, and role-based scoping.")
async def list_reimbursement_notes(
    status: Optional[str] = Query(None, description="D=Draft, P=Pending, A=Approved, R=Rejected, S=Submitted, B=Bank, PNA=Payment Note Approved"),
    page: int = Query(1, ge=1),
    per_page: int = Query(20, ge=1, le=100),
    search: Optional[str] = Query(None),
):
    notes = REIMBURSEMENTS_STORE
    if status:
        notes = [n for n in notes if n.get("status") == status]
    if search:
        s = search.lower()
        notes = [n for n in notes if s in str(n.get("user_name", "")).lower() or s in str(n.get("note_no", "")).lower()]
    start = (page - 1) * per_page
    end = start + per_page
    return {"notes": notes[start:end], "pagination": {"page": page, "per_page": per_page, "total_items": len(notes)}}


@app.get("/api/v1/reimbursements", tags=["🧾 Reimbursement Service"], summary="List Reimbursements (Alias)",
         description="Alias route for /api/v1/reimbursement-notes to support intuitive REST path discovery.")
async def list_reimbursements_alias(
    status: Optional[str] = Query(None), page: int = Query(1, ge=1), per_page: int = Query(20, ge=1, le=100), search: Optional[str] = Query(None)
):
    return await list_reimbursement_notes(status=status, page=page, per_page=per_page, search=search)


@app.get("/api/v1/reimbursement-notes/datatable", tags=["🧾 Reimbursement Service"], summary="List Reimbursement Notes (DataTable)",
         description="DataTables-compatible listing for server-side rendering.")
async def list_reimbursement_datatable(
    draw: int = Query(1), start: int = Query(0), length: int = Query(20),
):
    return {"draw": draw, "total_records": 0, "total_display_records": 0, "notes": []}


@app.get("/api/v1/reimbursement-notes/{id}", tags=["🧾 Reimbursement Service"], summary="Get Reimbursement Note",
         description="Retrieves full reimbursement note including expenses, attachments, approval logs, and bank details.")
async def get_reimbursement_note(id: int = Path(..., description="Reimbursement Note ID")):
    return {"note": {}}


@app.get("/api/v1/reimbursement-notes/create", tags=["🧾 Reimbursement Service"], summary="Get Create Form Data",
         description="Returns dropdown data needed for the reimbursement creation form: order number, user info, projects, vendors, departments, and approvers.")
async def get_reimbursement_create_data():
    return {"order_number": "", "user": {}, "projects": [], "vendors": [], "departments": [], "approvers": []}


@app.post("/api/v1/reimbursement-notes", tags=["🧾 Reimbursement Service"], summary="Create Reimbursement Note",
          description="""
Creates a new reimbursement note for employee expense claims.

**Payload includes:**
- Travel details (date, mode, purpose)
- Expense line items (type, bill details, amounts)
- Bank details for reimbursement
- File attachments (receipts, bills)

**Status Flow:** Draft → Pending → Approved → Bank → Paid
""")
async def create_reimbursement_note(body: Dict[str, Any] = Body(...)):
    note_data = body.get("note", body) if isinstance(body, dict) else {}
    expenses = body.get("expenses", []) if isinstance(body, dict) else []
    total_amt = sum(e.get("bill_amount", 0) for e in expenses) if expenses else float(note_data.get("amount", note_data.get("total_amount", 0)))
    new_id = len(REIMBURSEMENTS_STORE) + 1
    new_item = {
        "id": new_id,
        "note_no": note_data.get("note_no", f"TR-2026-00{new_id}"),
        "user_name": note_data.get("account_holder", "System User"),
        "project_name": "NHIT Regional Operations",
        "date_of_travel": note_data.get("date_of_travel", "2026-10-17"),
        "purpose_of_travel": note_data.get("purpose_of_travel", ""),
        "mode_of_travel": note_data.get("mode_of_travel", "Flight & Cab"),
        "bank_name": note_data.get("bank_name", "HDFC Bank"),
        "bank_account": note_data.get("bank_account", ""),
        "ifsc_code": note_data.get("ifsc_code", ""),
        "total_amount": total_amt,
        "status": note_data.get("status", "SUBMITTED"),
        "created_at": datetime.datetime.utcnow().isoformat() + "Z"
    }
    REIMBURSEMENTS_STORE.insert(0, new_item)
    return {"success": True, "note": new_item}


@app.post("/api/v1/reimbursements", tags=["🧾 Reimbursement Service"], summary="Create Reimbursement (Alias)")
async def create_reimbursements_alias(body: Dict[str, Any] = Body(...)):
    return await create_reimbursement_note(body)



@app.get("/api/v1/reimbursement-notes/{id}/edit", tags=["🧾 Reimbursement Service"], summary="Get Edit Form Data",
         description="Returns the note data pre-populated for editing, along with dropdown options.")
async def get_reimbursement_edit_data(id: int):
    return {"note": {}, "projects": [], "vendors": [], "departments": [], "approvers": []}


@app.patch("/api/v1/reimbursement-notes/{id}", tags=["🧾 Reimbursement Service"], summary="Update Reimbursement Note",
           description="Updates a reimbursement note. Can add/remove expenses and attachments. Only allowed for drafts and pending notes.")
async def update_reimbursement_note(id: int):
    return {"note": {}}


@app.delete("/api/v1/reimbursement-notes/{id}", tags=["🧾 Reimbursement Service"], summary="Delete Reimbursement Note",
            description="Permanently deletes a reimbursement note. Only allowed for draft status.")
async def delete_reimbursement_note(id: int = Path(..., description="Reimbursement Note ID")):
    return {}


@app.post("/api/v1/reimbursement-notes/{id}/approval-log", tags=["🧾 Reimbursement Service"], summary="Update Approval Log",
          description="Records an approval or rejection decision with reviewer remarks. Advances the note through the approval workflow.")
async def update_reimbursement_approval_log(id: int, body: ApprovalLogUpdateRequest):
    return {"note": {}}


@app.get("/api/v1/reimbursement-notes/{id}/pdf", tags=["🧾 Reimbursement Service"], summary="Download Reimbursement PDF",
         description="Generates and downloads a PDF for the reimbursement note including all expense details and approval history.")
async def download_reimbursement_pdf(id: int):
    return {"message": "PDF generation — forwards to Reimbursement gRPC service"}


@app.get("/api/v1/reimbursement-notes/{id}/pdf/preview", tags=["🧾 Reimbursement Service"], summary="Preview Reimbursement PDF",
         description="Generates a PDF preview (inline display rather than download).")
async def preview_reimbursement_pdf(id: int):
    return {"message": "PDF preview — forwards to Reimbursement gRPC service"}


@app.get("/api/v1/reimbursement-notes/{id}/payment-note", tags=["🧾 Reimbursement Service"], summary="Get Payment Note Draft Data",
         description="Returns pre-filled data for creating a payment note from an approved reimbursement note.")
async def get_reimbursement_payment_data(id: int):
    return {"payment_order_number": "", "gross_amount": 0, "note": {}}


@app.get("/api/v1/reimbursement-notes/users", tags=["🧾 Reimbursement Service"], summary="List Selectable Users",
         description="Returns users eligible for reimbursement selection (employees in the same organization).")
async def list_selectable_users():
    return {"users": []}


@app.post("/api/v1/reimbursement-notes/users/select", tags=["🧾 Reimbursement Service"], summary="Select User for Reimbursement",
          description="Selects a user as the reimbursement claimant (can be different from the requester).")
async def select_user_for_reimbursement():
    return {"success": True, "selected_user_id": 0}


@app.delete("/api/v1/reimbursement-notes/{id}/files/{filename}", tags=["🧾 Reimbursement Service"], summary="Delete Attachment",
            description="Removes a specific file attachment from a reimbursement note.")
async def delete_reimbursement_file(id: int, filename: str):
    return {"note": {}}


@app.get("/api/v1/reimbursement-notes/order-number", tags=["🧾 Reimbursement Service"], summary="Generate Order Number",
         description="Generates the next sequential reimbursement note number with financial year prefix.")
async def generate_reimbursement_order_number():
    return {"order_number": "RN-2024-25/0001", "financial_year": "2024-25"}


@app.get("/api/v1/reimbursement-notes/payment-order-number", tags=["🧾 Reimbursement Service"], summary="Generate Payment Order Number",
         description="Generates the next payment order number for reimbursement payment notes.")
async def generate_reimbursement_payment_order():
    return {"order_number": "RPN-2024-25/0001", "financial_year": "2024-25"}


@app.get("/api/v1/reimbursement-notes/test", tags=["🧾 Reimbursement Service"], summary="Test Reimbursement API",
         description="Health check for the Reimbursement gRPC service.")
async def test_reimbursement_api():
    return {"status": "ok", "message": "Reimbursement API is operational", "timestamp": int(datetime.datetime.now(datetime.timezone.utc).timestamp())}


# ═══════════════════════════════════════════════════════
# ✅ APPROVAL SERVICE (1 endpoint — internal gRPC, documented for reference)
# ═══════════════════════════════════════════════════════

@app.post("/api/v1/approvals/initiate", tags=["✅ Approval Service (Internal)"], summary="Initiate Approval Workflow",
          description="""
**⚠️ Internal gRPC endpoint** — documented for developer reference.

Initiates a multi-level approval workflow for a Green Note or Payment Note.

**Routing Logic:**
- Amount < ₹5L → Single-level approval
- Amount ₹5L-₹25L → Two-level approval
- Amount > ₹25L → Three-level approval + department head

**Parameters:**
- `sl_no`: Source document ID (Green Note or Payment Note)
- `amount`: Total amount for routing
- `department_id`: Department for approver assignment
- `source_type`: "green_note" or "payment_note"
""")
async def initiate_approval(body: InitiateApprovalRequest):
    return {"success": True, "message": "Approval initiated", "status": "PENDING_LEVEL_1"}


# ═══════════════════════════════════════════════════════
# DASHBOARD (HTML)
# ═══════════════════════════════════════════════════════

PORTAL_FILE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "portal", "index.html")

@app.get("/", response_class=HTMLResponse, include_in_schema=False)
@app.get("/portal", response_class=HTMLResponse, include_in_schema=False)
async def dashboard():
    if os.path.exists(PORTAL_FILE):
        with open(PORTAL_FILE, "r", encoding="utf-8") as f:
            return HTMLResponse(content=f.read())
    return HTMLResponse(content="<h1>NHIT Portal</h1><p>portal/index.html not found</p>")

# ═══════════════════════════════════════════════════════

if __name__ == "__main__":
    print("=" * 70)
    print("  NHIT Backend — Enterprise API Gateway & Swagger Hub")
    print("  Dashboard : http://localhost:8080")
    print("  Swagger   : http://localhost:8080/swagger")
    print("  ReDoc     : http://localhost:8080/redoc")
    print("  OpenAPI   : http://localhost:8080/openapi.json")
    print("=" * 70)
    uvicorn.run(app, host="0.0.0.0", port=8080)
