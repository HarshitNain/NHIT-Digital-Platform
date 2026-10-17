-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    tenant_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email_verified_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    last_logout_at TIMESTAMPTZ,
    last_login_ip VARCHAR(50),
    user_agent TEXT,
    department_id UUID,
    designation_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User Roles (Many-to-Many)
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    PRIMARY KEY(user_id, role_id)
);

-- User Login History
CREATE TABLE IF NOT EXISTS user_login_history (
    history_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    ip_address VARCHAR(50),
    user_agent TEXT,
    login_time TIMESTAMPTZ DEFAULT NOW()
);

-- Roles table (dynamic roles with fixed permission keys)
CREATE TABLE IF NOT EXISTS roles (
    role_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    parent_org_id UUID,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    permissions TEXT[] NOT NULL DEFAULT '{}',
    is_system_role BOOLEAN NOT NULL DEFAULT FALSE,
    created_by VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Permissions catalog (fixed list of allowed permission keys)
CREATE TABLE IF NOT EXISTS permissions (
    permission_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    module VARCHAR(100),
    action VARCHAR(50),
    is_system_permission BOOLEAN NOT NULL DEFAULT TRUE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_department_id ON users(department_id);
CREATE INDEX IF NOT EXISTS idx_users_designation_id ON users(designation_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX IF NOT EXISTS idx_user_login_history_user_id ON user_login_history(user_id);
CREATE INDEX IF NOT EXISTS idx_user_login_history_login_time ON user_login_history(login_time DESC);
CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_roles_parent_org_id ON roles(parent_org_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_tenant_org_name ON roles(tenant_id, parent_org_id, lower(name));
CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_name ON permissions(LOWER(name));

-- Comments for documentation
COMMENT ON TABLE users IS 'Stores user information';
COMMENT ON TABLE user_roles IS 'Many-to-many relationship between users and roles';
COMMENT ON TABLE user_login_history IS 'Tracks user login history';
COMMENT ON TABLE roles IS 'Stores dynamic roles (name + permissions) per tenant/organization';
COMMENT ON TABLE permissions IS 'Catalog of allowed permission keys used by roles';
COMMENT ON COLUMN users.user_id IS 'Unique identifier for the user';
COMMENT ON COLUMN users.tenant_id IS 'Reference to tenant (multi-tenancy support)';
COMMENT ON COLUMN users.email IS 'User email (unique)';
COMMENT ON COLUMN users.email_verified_at IS 'Timestamp when email was verified';
COMMENT ON COLUMN users.department_id IS 'Reference to department';
COMMENT ON COLUMN users.designation_id IS 'Reference to designation';

-- Seed fixed permission list for role management UI
INSERT INTO permissions (name, description, module, action, is_system_permission)
VALUES
    -- Dashboard & Reports
    ('view-dashboard', 'View dashboard', 'dashboard', 'view', TRUE),
    ('view-reports', 'View analytical and summary reports', 'reports', 'view', TRUE),
    ('export-data', 'Export data from the system', 'reports', 'export', TRUE),

    -- User & Role management
    ('create-user', 'Create new users', 'users', 'create', TRUE),
    ('edit-user', 'Edit existing users', 'users', 'edit', TRUE),
    ('delete-user', 'Delete users', 'users', 'delete', TRUE),
    ('view-user', 'View users', 'users', 'view', TRUE),

    ('create-role', 'Create new roles', 'roles', 'create', TRUE),
    ('edit-role', 'Edit existing roles', 'roles', 'edit', TRUE),
    ('delete-role', 'Delete roles', 'roles', 'delete', TRUE),
    ('view-role', 'View roles', 'roles', 'view', TRUE),

    -- Department & Designation
    ('create-department', 'Create departments', 'departments', 'create', TRUE),
    ('edit-department', 'Edit departments', 'departments', 'edit', TRUE),
    ('delete-department', 'Delete departments', 'departments', 'delete', TRUE),
    ('view-department', 'View departments', 'departments', 'view', TRUE),

    ('create-designation', 'Create designations', 'designations', 'create', TRUE),
    ('edit-designation', 'Edit designations', 'designations', 'edit', TRUE),
    ('delete-designation', 'Delete designations', 'designations', 'delete', TRUE),
    ('view-designation', 'View designations', 'designations', 'view', TRUE),

    -- Notes & Approvals
    ('create-note', 'Create notes', 'notes', 'create', TRUE),
    ('edit-note', 'Edit notes', 'notes', 'edit', TRUE),
    ('delete-note', 'Delete notes', 'notes', 'delete', TRUE),
    ('view-note', 'View notes', 'notes', 'view', TRUE),
    ('reject-note', 'Reject notes', 'notes', 'reject', TRUE),
    ('approve-note', 'Approve notes', 'notes', 'approve', TRUE),
    ('view-all-notes', 'View all notes', 'notes', 'view', TRUE),

    -- Payment notes
    ('create-payment-note', 'Create payment notes', 'payment-notes', 'create', TRUE),
    ('edit-payment-note', 'Edit payment notes', 'payment-notes', 'edit', TRUE),
    ('delete-payment-note', 'Delete payment notes', 'payment-notes', 'delete', TRUE),
    ('view-payment-note', 'View payment notes', 'payment-notes', 'view', TRUE),
    ('view-all-payment-notes', 'View all payment notes', 'payment-notes', 'view', TRUE),
    ('approve-payment-note', 'Approve payment notes', 'payment-notes', 'approve', TRUE),

    -- Reimbursement notes
    ('create-reimbursement-note', 'Create reimbursement notes', 'reimbursement-notes', 'create', TRUE),
    ('edit-reimbursement-note', 'Edit reimbursement notes', 'reimbursement-notes', 'edit', TRUE),
    ('delete-reimbursement-note', 'Delete reimbursement notes', 'reimbursement-notes', 'delete', TRUE),
    ('view-reimbursement-note', 'View reimbursement notes', 'reimbursement-notes', 'view', TRUE),
    ('view-all-reimbursement-notes', 'View all reimbursement notes', 'reimbursement-notes', 'view', TRUE),
    ('approve-reimbursement-note', 'Approve reimbursement notes', 'reimbursement-notes', 'approve', TRUE),

    -- Vendors
    ('create-vendors', 'Create vendors', 'vendors', 'create', TRUE),
    ('edit-vendors', 'Edit vendors', 'vendors', 'edit', TRUE),
    ('delete-vendors', 'Delete vendors', 'vendors', 'delete', TRUE),
    ('view-vendors', 'View vendors', 'vendors', 'view', TRUE),

    -- Payments
    ('create-payment', 'Create payments', 'payments', 'create', TRUE),
    ('edit-payment', 'Edit payments', 'payments', 'edit', TRUE),
    ('delete-payment', 'Delete payments', 'payments', 'delete', TRUE),
    ('view-payment', 'View payments', 'payments', 'view', TRUE),
    ('approve-payment', 'Approve payments', 'payments', 'approve', TRUE),
    ('import-payment-excel', 'Import payments from Excel', 'payments', 'import', TRUE),

    -- Rules & Bank letters
    ('create-rule', 'Create business rules', 'rules', 'create', TRUE),
    ('edit-rule', 'Edit business rules', 'rules', 'edit', TRUE),
    ('delete-rule', 'Delete business rules', 'rules', 'delete', TRUE),
    ('view-rule', 'View business rules', 'rules', 'view', TRUE),

    ('create-bank-letters', 'Create bank letters', 'bank-letters', 'create', TRUE),
    ('edit-bank-letters', 'Edit bank letters', 'bank-letters', 'edit', TRUE),
    ('delete-bank-letters', 'Delete bank letters', 'bank-letters', 'delete', TRUE),
    ('view-bank-letters', 'View bank letters', 'bank-letters', 'view', TRUE),
    ('approve-bank-letters', 'Approve bank letters', 'bank-letters', 'approve', TRUE),

    -- Organizations
    ('edit-organizations', 'Edit organizations', 'organizations', 'edit', TRUE),
    ('delete-organizations', 'Delete organizations', 'organizations', 'delete', TRUE),
    ('view-organizations', 'View organizations', 'organizations', 'view', TRUE),
    ('switch-organizations', 'Switch between organizations', 'organizations', 'switch', TRUE),
    ('create-organizations', 'Create organizations', 'organizations', 'create', TRUE),

    --Projects
    ('create-projects', 'Create projects', 'projects', 'create', TRUE),
    ('edit-projects', 'Edit projects', 'projects', 'edit', TRUE),
    ('delete-projects', 'Delete projects', 'projects', 'delete', TRUE),
    ('view-projects', 'View projects', 'projects', 'view', TRUE),
    ('approve-projects', 'Approve projects', 'projects', 'approve', TRUE),
    
    -- Tickets & approvals
    ('view-tickets', 'View tickets', 'tickets', 'view', TRUE),
    ('create-tickets', 'Create tickets', 'tickets', 'create', TRUE),
    ('edit-tickets', 'Edit tickets', 'tickets', 'edit', TRUE),
    ('delete-tickets', 'Delete tickets', 'tickets', 'delete', TRUE),
    ('assign-tickets', 'Assign tickets to users', 'tickets', 'assign', TRUE),

    ('level-1-approver', 'Level 1 approver', 'approvals', 'level-1', TRUE),
    ('level-2-approver', 'Level 2 approver', 'approvals', 'level-2', TRUE),
    ('level-3-approver', 'Level 3 approver', 'approvals', 'level-3', TRUE),
    ('final-approver', 'Final approver', 'approvals', 'final', TRUE),

    -- Audit & system
    ('view-audit-trail', 'View audit trail', 'audit', 'view', TRUE),
    ('view-activity-logs', 'View activity logs', 'audit', 'view', TRUE),
    ('view-system-logs', 'View system logs', 'system', 'view', TRUE),
    ('system-configuration', 'Manage system configuration', 'system', 'configure', TRUE)
ON CONFLICT (name) DO NOTHING;
