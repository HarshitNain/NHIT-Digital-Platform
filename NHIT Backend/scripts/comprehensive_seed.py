"""
╔══════════════════════════════════════════════════════════════════════════════╗
║  NHIT Backend - Comprehensive Multi-Tenant Realistic Seed Script             ║
╚══════════════════════════════════════════════════════════════════════════════╝
Populates the NHIT Backend database with rich, realistic enterprise data:
- 2 Tenants
- 4 Organizations (Parent & Child structure)
- 6 Departments
- 8 Designations
- 15 System & Business Permissions
- 5 Roles with granular permission mappings
- 8 Enterprise Users with different roles and departments
- User-Organization associations and primary contexts
- User-Role mappings
- 5 High-Value Highway & Infrastructure Projects
- 5 Real-world Vendors (MSME & Non-MSME, contractors, consultants)
- 6 Vendor Bank Accounts with primary status
- Realistic login history records
"""

import sys
if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8')

import os
import psycopg2
import psycopg2.extras
import datetime
import uuid

DB_CONFIG = {
    "host": os.environ.get("DB_HOST", "localhost"),
    "port": int(os.environ.get("DB_PORT", 5433)),
    "dbname": os.environ.get("DB_NAME", "nhit_db"),
    "user": os.environ.get("DB_USER", "postgres"),
    "password": os.environ.get("DB_PASSWORD", "shristi"),
}

def seed():
    conn = psycopg2.connect(**DB_CONFIG)
    conn.autocommit = False
    cur = conn.cursor()
    now = datetime.datetime.now(datetime.timezone.utc)
    
    # Password hash for "password123"
    BCRYPT_PASSWORD = "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"

    print("🚀 Starting comprehensive seeding...")

    # =========================================================================
    # 1. TENANTS
    # =========================================================================
    TENANT_1 = "550e8400-e29b-41d4-a716-446655440000"
    TENANT_2 = "550e8400-e29b-41d4-a716-446655440010"
    
    tenants = [
        (TENANT_1, "National Highway Infra Trust", "infra.trust@nhit.com", BCRYPT_PASSWORD, now, now),
        (TENANT_2, "Apex Logistics Solutions", "contact@apexlogistics.com", BCRYPT_PASSWORD, now, now),
    ]
    for t in tenants:
        cur.execute("""
            INSERT INTO tenants (tenant_id, name, email, password, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s)
            ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name, updated_at = EXCLUDED.updated_at
        """, t)
    print("✅ Seeded 2 Tenants")

    # =========================================================================
    # 2. ORGANIZATIONS
    # =========================================================================
    ORG_TECH = "550e8400-e29b-41d4-a716-446655440001"
    ORG_ROAD = "550e8400-e29b-41d4-a716-446655440002"
    ORG_ASSET = "550e8400-e29b-41d4-a716-446655440003"
    ORG_LOGISTICS = "550e8400-e29b-41d4-a716-446655440004"

    orgs = [
        (ORG_TECH, TENANT_1, "NHIT Technologies Ltd", "NHIT-TECH", "nhit_tech_db", "Main technology and enterprise platform division", None, True, ORG_TECH, now, now),
        (ORG_ROAD, TENANT_1, "NHIT Road & Expressway Projects", "NHIT-ROAD", "nhit_road_db", "National Highway corridor development and EPC management", None, True, ORG_TECH, now, now),
        (ORG_ASSET, TENANT_1, "NHIT Highway Asset Management", "NHIT-ASSET", "nhit_asset_db", "Toll operations, FASTag reconciliation, and maintenance", None, True, ORG_TECH, now, now),
        (ORG_LOGISTICS, TENANT_2, "Apex Freight & Logistics Org", "APEX-LOG", "apex_log_db", "Third-party logistics, heavy fleet, and highway freight services", None, True, ORG_TECH, now, now),
    ]
    for o in orgs:
        cur.execute("""
            INSERT INTO organizations (org_id, tenant_id, name, code, database_name, description, logo, is_active, created_by, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (org_id) DO UPDATE SET 
                name = EXCLUDED.name, 
                code = EXCLUDED.code, 
                database_name = EXCLUDED.database_name, 
                description = EXCLUDED.description,
                updated_at = EXCLUDED.updated_at
        """, o)
    print("✅ Seeded 4 Organizations")

    # =========================================================================
    # 3. DEPARTMENTS
    # =========================================================================
    DEPT_IT = "550e8400-e29b-41d4-a716-446655440002"
    DEPT_CIVIL = "650e8400-e29b-41d4-a716-446655440001"
    DEPT_FINANCE = "650e8400-e29b-41d4-a716-446655440002"
    DEPT_HR = "650e8400-e29b-41d4-a716-446655440003"
    DEPT_PROC = "650e8400-e29b-41d4-a716-446655440004"
    DEPT_OPS = "650e8400-e29b-41d4-a716-446655440005"

    departments = [
        (DEPT_IT, ORG_TECH, "Information Technology", "Cloud infrastructure, enterprise APIs, and digital platform engineering", now, now),
        (DEPT_CIVIL, ORG_ROAD, "Civil Engineering & Construction", "Highway alignment, bridges, flyovers, and structural engineering", now, now),
        (DEPT_FINANCE, ORG_TECH, "Finance & Accounts", "Budgeting, payment notes, RTGS disbursements, and financial auditing", now, now),
        (DEPT_HR, ORG_TECH, "Human Resources", "People operations, talent acquisition, payroll, and reimbursements", now, now),
        (DEPT_PROC, ORG_ROAD, "Procurement & Contracts", "Tenders, vendor onboarding, contract drafting, and MSME compliance", now, now),
        (DEPT_OPS, ORG_ASSET, "Toll Operations & Monitoring", "FASTag lanes, smart CCTV monitoring, and maintenance patrol", now, now),
    ]
    for d in departments:
        cur.execute("""
            INSERT INTO departments (id, org_id, name, description, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s)
            ON CONFLICT (name) DO UPDATE SET 
                description = EXCLUDED.description,
                org_id = EXCLUDED.org_id,
                updated_at = EXCLUDED.updated_at
        """, d)
    print("✅ Seeded 6 Departments")

    # =========================================================================
    # 4. DESIGNATIONS
    # =========================================================================
    DESIG_CEO = "750e8400-e29b-41d4-a716-446655440001"
    DESIG_CTO = "750e8400-e29b-41d4-a716-446655440002"
    DESIG_CFO = "750e8400-e29b-41d4-a716-446655440003"
    DESIG_DIR = "750e8400-e29b-41d4-a716-446655440004"
    DESIG_ENG = "750e8400-e29b-41d4-a716-446655440005"
    DESIG_DEV = "550e8400-e29b-41d4-a716-446655440003"
    DESIG_PROC = "750e8400-e29b-41d4-a716-446655440006"
    DESIG_ACC = "750e8400-e29b-41d4-a716-446655440007"

    designations = [
        (DESIG_CEO, ORG_TECH, "Chief Executive Officer", "Executive head of the enterprise", now, now),
        (DESIG_CTO, ORG_TECH, "Chief Technology Officer", "Technology executive leading engineering architecture", now, now),
        (DESIG_CFO, ORG_TECH, "Chief Financial Officer", "Financial executive overseeing treasury and note approvals", now, now),
        (DESIG_DIR, ORG_ROAD, "Project Director", "Senior director supervising road and corridor construction", now, now),
        (DESIG_ENG, ORG_ROAD, "Senior Resident Engineer", "Field resident engineer inspecting site progress and bills", now, now),
        (DESIG_DEV, ORG_TECH, "Software Developer", "Full-stack engineer building microservices and web portal", now, now),
        (DESIG_PROC, ORG_ROAD, "Procurement Officer", "Vendor management, bids, and purchase order officer", now, now),
        (DESIG_ACC, ORG_TECH, "Senior Accounts Manager", "Vendor invoice processing and bank reconciliation", now, now),
    ]
    for dg in designations:
        cur.execute("""
            INSERT INTO designations (id, org_id, name, description, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s)
            ON CONFLICT (id) DO UPDATE SET 
                name = EXCLUDED.name, 
                description = EXCLUDED.description,
                org_id = EXCLUDED.org_id,
                updated_at = EXCLUDED.updated_at
        """, dg)
    print("✅ Seeded 8 Designations")

    # =========================================================================
    # 5. PERMISSIONS
    # =========================================================================
    PERM_LIST = [
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "users.create")), True, "users.create", "Create new system users", "users", "create"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "users.read")), True, "users.read", "View users and profile details", "users", "read"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "users.update")), True, "users.update", "Update user records", "users", "update"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "users.delete")), True, "users.delete", "Delete or archive user accounts", "users", "delete"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "orgs.create")), True, "orgs.create", "Create organizations", "organizations", "create"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "orgs.read")), True, "orgs.read", "List and view organizations", "organizations", "read"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "orgs.update")), True, "orgs.update", "Update organization settings", "organizations", "update"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "projects.create")), True, "projects.create", "Create infrastructure projects", "projects", "create"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "projects.read")), True, "projects.read", "View project details and milestones", "projects", "read"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "vendors.create")), True, "vendors.create", "Onboard new vendors and bank accounts", "vendors", "create"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "vendors.read")), True, "vendors.read", "View vendors and banking records", "vendors", "read"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "notes.create")), True, "notes.create", "Create Green Notes and Payment Notes", "notes", "create"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "notes.read")), True, "notes.read", "View notes and attachments", "notes", "read"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "notes.approve")), True, "notes.approve", "Approve Green/Payment Notes in workflow", "notes", "approve"),
        (str(uuid.uuid5(uuid.NAMESPACE_DNS, "payments.process")), True, "payments.process", "Disburse funds and generate bank RTGS letters", "payments", "process"),
    ]
    for p in PERM_LIST:
        cur.execute("""
            INSERT INTO permissions (permission_id, is_system_permission, name, description, module, action)
            VALUES (%s, %s, %s, %s, %s, %s)
            ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description, module = EXCLUDED.module, action = EXCLUDED.action
        """, p)
    print("✅ Seeded 15 Permissions")

    # =========================================================================
    # 6. ROLES
    # =========================================================================
    ROLE_SUPERADMIN = "e50e8400-e29b-41d4-a716-446655440001"
    ROLE_ORGADMIN = "e50e8400-e29b-41d4-a716-446655440002"
    ROLE_PM = "e50e8400-e29b-41d4-a716-446655440003"
    ROLE_FINANCE = "e50e8400-e29b-41d4-a716-446655440004"
    ROLE_EMPLOYEE = "e50e8400-e29b-41d4-a716-446655440005"

    roles = [
        (ROLE_SUPERADMIN, TENANT_1, None, True, "super_admin", "Full system administrator with unrestricted privileges",
         ["*"], "system", now, now),
        (ROLE_ORGADMIN, TENANT_1, ORG_TECH, False, "org_admin", "Organization administrator managing users, depts, and projects",
         ["users.create", "users.read", "users.update", "orgs.read", "orgs.update", "projects.create", "projects.read", "vendors.read"], "system", now, now),
        (ROLE_PM, TENANT_1, ORG_ROAD, False, "project_manager", "Project manager leading site execution, vendor notes, and progress",
         ["projects.read", "vendors.create", "vendors.read", "notes.create", "notes.read"], "system", now, now),
        (ROLE_FINANCE, TENANT_1, ORG_TECH, False, "finance_reviewer", "Financial reviewer handling note approvals, vouchers, and disbursements",
         ["notes.read", "notes.approve", "payments.process", "vendors.read"], "system", now, now),
        (ROLE_EMPLOYEE, TENANT_1, ORG_TECH, False, "employee", "Standard employee permitted to submit reimbursement notes and view projects",
         ["notes.create", "notes.read", "projects.read"], "system", now, now),
    ]
    for r in roles:
        cur.execute("""
            INSERT INTO roles (role_id, tenant_id, parent_org_id, is_system_role, name, description, permissions, created_by, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (role_id) DO UPDATE SET 
                name = EXCLUDED.name, 
                description = EXCLUDED.description, 
                permissions = EXCLUDED.permissions,
                updated_at = EXCLUDED.updated_at
        """, r)
    print("✅ Seeded 5 Roles")

    # =========================================================================
    # 7. USERS
    # =========================================================================
    USER_ADMIN = "550e8400-e29b-41d4-a716-446655440001"
    USER_CTO = "850e8400-e29b-41d4-a716-446655440002"
    USER_CFO = "850e8400-e29b-41d4-a716-446655440003"
    USER_DIR = "850e8400-e29b-41d4-a716-446655440004"
    USER_ENG = "850e8400-e29b-41d4-a716-446655440005"
    USER_PROC = "850e8400-e29b-41d4-a716-446655440006"
    USER_ACC = "850e8400-e29b-41d4-a716-446655440007"
    USER_DEV = "850e8400-e29b-41d4-a716-446655440008"

    users = [
        (USER_ADMIN, TENANT_1, "NHIT Super Administrator", "admin@nhit.com", BCRYPT_PASSWORD, now, DEPT_IT, DESIG_CEO, "Chrome/Windows", "127.0.0.1", now, now, now, None),
        (USER_CTO, TENANT_1, "Rajesh Sharma", "cto@nhit.com", BCRYPT_PASSWORD, now, DEPT_IT, DESIG_CTO, "Firefox/Mac", "192.168.1.45", now, now, now, None),
        (USER_CFO, TENANT_1, "Priya Mehta", "cfo@nhit.com", BCRYPT_PASSWORD, now, DEPT_FINANCE, DESIG_CFO, "Edge/Windows", "192.168.1.12", now, now, now, None),
        (USER_DIR, TENANT_1, "Amit Verma", "project.director@nhit.com", BCRYPT_PASSWORD, now, DEPT_CIVIL, DESIG_DIR, "Safari/Mac", "192.168.2.100", now, now, now, None),
        (USER_ENG, TENANT_1, "Sneha Patel", "engineer@nhit.com", BCRYPT_PASSWORD, now, DEPT_CIVIL, DESIG_ENG, "Chrome/Android", "10.0.4.15", now, now, now, None),
        (USER_PROC, TENANT_1, "Vikram Malhotra", "procurement@nhit.com", BCRYPT_PASSWORD, now, DEPT_PROC, DESIG_PROC, "Chrome/Windows", "192.168.1.88", now, now, now, None),
        (USER_ACC, TENANT_1, "Ananya Sen", "accounts@nhit.com", BCRYPT_PASSWORD, now, DEPT_FINANCE, DESIG_ACC, "Edge/Windows", "192.168.1.99", now, now, now, None),
        (USER_DEV, TENANT_1, "Kunal Nain", "developer@nhit.com", BCRYPT_PASSWORD, now, DEPT_IT, DESIG_DEV, "VSCode/Windows", "127.0.0.1", now, now, now, None),
    ]

    for u in users:
        cur.execute("""
            INSERT INTO users (user_id, tenant_id, name, email, password, email_verified_at, department_id, designation_id, user_agent, last_login_ip, created_at, updated_at, last_login_at, last_logout_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (email) DO UPDATE SET 
                name = EXCLUDED.name,
                department_id = EXCLUDED.department_id,
                designation_id = EXCLUDED.designation_id,
                updated_at = EXCLUDED.updated_at
        """, u)
    print("✅ Seeded 8 Users")

    # =========================================================================
    # 8. USER ORGANIZATIONS & USER ROLES
    # =========================================================================
    user_org_roles = [
        (USER_ADMIN, ORG_TECH, ROLE_SUPERADMIN, True),
        (USER_ADMIN, ORG_ROAD, ROLE_SUPERADMIN, False),
        (USER_CTO, ORG_TECH, ROLE_ORGADMIN, True),
        (USER_CFO, ORG_TECH, ROLE_FINANCE, True),
        (USER_DIR, ORG_ROAD, ROLE_PM, True),
        (USER_ENG, ORG_ROAD, ROLE_EMPLOYEE, True),
        (USER_PROC, ORG_ROAD, ROLE_PM, True),
        (USER_ACC, ORG_TECH, ROLE_FINANCE, True),
        (USER_DEV, ORG_TECH, ROLE_EMPLOYEE, True),
    ]
    for u_id, o_id, r_id, is_curr in user_org_roles:
        cur.execute("""
            INSERT INTO user_organizations (user_id, org_id, role_id, is_current_context, joined_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s)
            ON CONFLICT (user_id, org_id) DO UPDATE SET 
                role_id = EXCLUDED.role_id,
                is_current_context = EXCLUDED.is_current_context,
                updated_at = EXCLUDED.updated_at
        """, (u_id, o_id, r_id, is_curr, now, now))

        cur.execute("""
            INSERT INTO user_roles (user_id, role_id)
            VALUES (%s, %s)
            ON CONFLICT (user_id, role_id) DO NOTHING
        """, (u_id, r_id))
    print("✅ Seeded User Organizations & User Roles")

    # =========================================================================
    # 9. PROJECTS
    # =========================================================================
    projects = [
        ("950e8400-e29b-41d4-a716-446655440001", TENANT_1, "Eastern Peripheral Expressway Corridor Phase 2", ORG_ROAD, "Amit Verma", now, now),
        ("950e8400-e29b-41d4-a716-446655440002", TENANT_1, "Delhi-Agra Expressway Smart Tolling Modernization", ORG_ROAD, "Amit Verma", now, now),
        ("950e8400-e29b-41d4-a716-446655440003", TENANT_1, "NHIT Enterprise Cloud & FASTag Reconciliation Platform", ORG_TECH, "Rajesh Sharma", now, now),
        ("950e8400-e29b-41d4-a716-446655440004", TENANT_1, "National Highway Drone Surveillance & Asset Monitoring", ORG_ASSET, "Sneha Patel", now, now),
        ("950e8400-e29b-41d4-a716-446655440005", TENANT_1, "Bhopal-Indore Four-Lane Corridor Expansion", ORG_ROAD, "Amit Verma", now, now),
    ]
    for p in projects:
        cur.execute("""
            INSERT INTO projects (id, tenant_id, project_name, org_id, created_by, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (id) DO UPDATE SET 
                project_name = EXCLUDED.project_name,
                org_id = EXCLUDED.org_id,
                updated_at = EXCLUDED.updated_at
        """, p)
    print("✅ Seeded 5 Projects")

    # =========================================================================
    # 10. VENDORS & VENDOR BANK ACCOUNTS
    # =========================================================================
    VEND_1 = "a50e8400-e29b-41d4-a716-446655440001"
    VEND_2 = "a50e8400-e29b-41d4-a716-446655440002"
    VEND_3 = "a50e8400-e29b-41d4-a716-446655440003"
    VEND_4 = "a50e8400-e29b-41d4-a716-446655440004"
    VEND_5 = "a50e8400-e29b-41d4-a716-446655440005"

    vendors = [
        (VEND_1, TENANT_1, "VEN-LNT-001", "Larsen & Toubro Infrastructure Ltd", "contact@lntecc.com", "+91-22-67525656", "EXTERNAL", "AAACL1234E", "Larsen & Toubro Ltd", "27AAACL1234E1Z1", True, "NO", "None", "EPC Contractor", USER_PROC, now, now),
        (VEND_2, TENANT_1, "VEN-DBL-002", "Dilip Buildcon Roadways Pvt Ltd", "tenders@dilipbuildcon.co.in", "+91-755-4029999", "EXTERNAL", "AABCD5678F", "Dilip Buildcon Pvt Ltd", "23AABCD5678F1Z2", True, "YES", "Medium", "Highway Concessionaire", USER_PROC, now, now),
        (VEND_3, TENANT_1, "VEN-AFC-003", "Afcons Infrastructure Limited", "procure@afcons.com", "+91-22-67191000", "EXTERNAL", "AACCA9012G", "Afcons Infrastructure Ltd", "27AACCA9012G1Z3", True, "NO", "None", "Bridges & Tunnels Specialist", USER_PROC, now, now),
        (VEND_4, TENANT_1, "VEN-TSG-004", "Techno Survey & Geospatial Consultants", "info@technosurvey.in", "+91-11-29567890", "EXTERNAL", "BBBPK3456H", "Techno Survey Geospatial LLP", "07BBBPK3456H1Z4", True, "YES", "Micro", "Geotechnical & Drone Survey", USER_PROC, now, now),
        (VEND_5, TENANT_1, "VEN-ITS-005", "InfraTech Safety Equipment & Signages", "sales@infratechsafety.com", "+91-120-4567890", "EXTERNAL", "CCCPK7890J", "InfraTech Safety Pvt Ltd", "08CCCPK7890J1Z5", True, "YES", "Small", "Highway Crash Barriers & Signage", USER_PROC, now, now),
    ]

    for v in vendors:
        cur.execute("""
            INSERT INTO vendors (
                id, tenant_id, vendor_code, vendor_name, vendor_email, vendor_mobile, 
                vendor_type, pan, beneficiary_name, gstin, is_active, msme, msme_classification,
                material_nature, created_by, created_at, updated_at
            )
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (id) DO UPDATE SET 
                vendor_name = EXCLUDED.vendor_name,
                vendor_code = EXCLUDED.vendor_code,
                vendor_email = EXCLUDED.vendor_email,
                pan = EXCLUDED.pan,
                gstin = EXCLUDED.gstin,
                msme = EXCLUDED.msme,
                msme_classification = EXCLUDED.msme_classification,
                updated_at = EXCLUDED.updated_at
        """, v)
    print("✅ Seeded 5 Vendors")

    # Vendor Bank Accounts
    vendor_accounts = [
        ("b50e8400-e29b-41d4-a716-446655440001", VEND_1, "L&T Infrastructure Master Account", "50200012345678", "Current", "HDFC Bank", "Fort Mumbai Branch", "HDFC0000060", True, True, USER_PROC, now, now),
        ("b50e8400-e29b-41d4-a716-446655440002", VEND_2, "Dilip Buildcon Main Treasury", "31234567890123", "Current", "State Bank of India", "Bhopal Main Branch", "SBIN0001056", True, True, USER_PROC, now, now),
        ("b50e8400-e29b-41d4-a716-446655440003", VEND_3, "Afcons Infrastructure Disbursal", "000405001234", "Current", "ICICI Bank", "Nariman Point Mumbai", "ICIC0000004", True, True, USER_PROC, now, now),
        ("b50e8400-e29b-41d4-a716-446655440004", VEND_4, "Techno Survey Operations", "919020012345678", "Current", "Axis Bank", "Connaught Place New Delhi", "UTIB0000007", True, True, USER_PROC, now, now),
        ("b50e8400-e29b-41d4-a716-446655440005", VEND_5, "InfraTech Safety Equipment", "123456789012", "Current", "Kotak Mahindra Bank", "Sector 18 Noida Branch", "KKBK0000180", True, True, USER_PROC, now, now),
        ("b50e8400-e29b-41d4-a716-446655440006", VEND_1, "L&T Secondary Retention Account", "50200098765432", "Current", "State Bank of India", "Industrial Finance Branch Mumbai", "SBIN0008965", False, True, USER_PROC, now, now),
    ]
    for va in vendor_accounts:
        cur.execute("""
            INSERT INTO vendor_accounts (id, vendor_id, account_name, account_number, account_type, name_of_bank, branch_name, ifsc_code, is_primary, is_active, created_by, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (id) DO UPDATE SET 
                account_name = EXCLUDED.account_name,
                account_number = EXCLUDED.account_number,
                name_of_bank = EXCLUDED.name_of_bank,
                ifsc_code = EXCLUDED.ifsc_code,
                is_primary = EXCLUDED.is_primary,
                updated_at = EXCLUDED.updated_at
        """, va)
    print("✅ Seeded 6 Vendor Bank Accounts")

    # =========================================================================
    # 11. AUDIT LOGIN HISTORY
    # =========================================================================
    history = [
        (USER_ADMIN, now - datetime.timedelta(minutes=30), "127.0.0.1", "Chrome 122.0.0 Win64"),
        (USER_CTO, now - datetime.timedelta(hours=2), "192.168.1.45", "Firefox 124.0.1 MacOS"),
        (USER_CFO, now - datetime.timedelta(hours=5), "192.168.1.12", "Edge 122.0.0 Win64"),
        (USER_DIR, now - datetime.timedelta(days=1), "192.168.2.100", "Safari 17.2 MacOS"),
    ]
    for h in history:
        cur.execute("""
            INSERT INTO user_login_history (user_id, login_time, ip_address, user_agent)
            VALUES (%s, %s, %s, %s)
        """, h)
    print("✅ Seeded Login History")

    conn.commit()
    cur.close()
    conn.close()
    print("🎉 Comprehensive database seeding completed successfully!")

if __name__ == "__main__":
    seed()
