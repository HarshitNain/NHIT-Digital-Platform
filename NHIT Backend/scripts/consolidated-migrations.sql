-- ================================================
-- NHIT Backend - Consolidated Database Migrations
-- ================================================
-- This script creates all tables for all microservices
-- Execute this script to set up the complete NHIT database schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ================================================
-- ORGANIZATION SERVICE TABLES
-- ================================================

-- Organizations Table
CREATE TABLE IF NOT EXISTS organizations (
    org_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(10) NOT NULL UNIQUE,
    database_name VARCHAR(64) NOT NULL UNIQUE,
    description TEXT,
    logo VARCHAR(500),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- User Organizations Table (Junction Table)
CREATE TABLE IF NOT EXISTS user_organizations (
    user_id UUID NOT NULL,
    org_id UUID NOT NULL,
    role_id UUID NOT NULL,
    is_current_context BOOLEAN NOT NULL DEFAULT false,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (user_id, org_id),
    FOREIGN KEY (org_id) REFERENCES organizations(org_id) ON DELETE CASCADE
);

-- ================================================
-- USER SERVICE TABLES
-- ================================================

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

-- ================================================
-- AUTH SERVICE TABLES
-- ================================================

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    session_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    session_token TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

-- Refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    token TEXT PRIMARY KEY,
    user_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Password resets table
CREATE TABLE IF NOT EXISTS password_resets (
    token UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Email verification tokens table
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    token UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ================================================
-- DEPARTMENT SERVICE TABLES
-- ================================================

-- Departments table
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    org_id UUID NOT NULL
);

-- ================================================
-- DESIGNATION SERVICE TABLES
-- ================================================

-- Designations table
CREATE TABLE IF NOT EXISTS designations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(250) NOT NULL,
    description VARCHAR(500) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    org_id UUID NOT NULL
);

-- ================================================
-- VENDOR SERVICE TABLES
-- ================================================

-- Vendors table
CREATE TABLE IF NOT EXISTS vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    vendor_code VARCHAR(100) NOT NULL,
    vendor_name VARCHAR(255) NOT NULL,
    vendor_email VARCHAR(255) NOT NULL,
    vendor_mobile VARCHAR(20),
    vendor_type VARCHAR(50),
    vendor_nick_name VARCHAR(255),
    activity_type VARCHAR(255),
    gstin VARCHAR(50),
    pan VARCHAR(20) NOT NULL,
    pin VARCHAR(20),
    country_name VARCHAR(255),
    state_name VARCHAR(255),
    city_name VARCHAR(255),
    msme_classification VARCHAR(100),
    msme VARCHAR(100),
    msme_registration_number VARCHAR(100),
    msme_start_date DATE,
    msme_end_date DATE,
    material_nature VARCHAR(255),
    gst_defaulted VARCHAR(10),
    section_206ab_verified VARCHAR(10),
    beneficiary_name VARCHAR(255) NOT NULL,
    remarks_address TEXT,
    common_bank_details TEXT,
    income_tax_type VARCHAR(100),
    project VARCHAR(255),
    status VARCHAR(50),
    from_account_type VARCHAR(100),
    account_name VARCHAR(255),
    short_name VARCHAR(100),
    parent VARCHAR(255),
    file_paths JSONB,
    code_auto_generated BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Backward compatibility banking fields
    account_number VARCHAR(50),
    name_of_bank VARCHAR(255),
    ifsc_code VARCHAR(20),
    
    -- Constraints
    CONSTRAINT vendors_tenant_vendor_code_unique UNIQUE (tenant_id, vendor_code),
    CONSTRAINT vendors_tenant_vendor_email_unique UNIQUE (tenant_id, vendor_email),
    CONSTRAINT vendors_pan_check CHECK (pan ~ '^[A-Z]{5}[0-9]{4}[A-Z]{1}$'),
    CONSTRAINT vendors_ifsc_check CHECK (ifsc_code IS NULL OR ifsc_code ~ '^[A-Z]{4}0[A-Z0-9]{6}$')
);

-- Vendor accounts table
CREATE TABLE IF NOT EXISTS vendor_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id UUID NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    account_name VARCHAR(255) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    account_type VARCHAR(50),
    name_of_bank VARCHAR(255) NOT NULL,
    branch_name VARCHAR(255),
    ifsc_code VARCHAR(20) NOT NULL,
    swift_code VARCHAR(20),
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    remarks TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT vendor_accounts_ifsc_check CHECK (ifsc_code ~ '^[A-Z]{4}0[A-Z0-9]{6}$'),
    CONSTRAINT vendor_accounts_account_number_check CHECK (account_number ~ '^[0-9]{9,18}$')
);

-- ================================================
-- PROJECT SERVICE TABLES
-- ================================================

-- Projects table (based on migration pattern)
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    org_id UUID NOT NULL,
    project_name VARCHAR(255) NOT NULL,
    created_by TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ================================================
-- INDEXES FOR PERFORMANCE
-- ================================================

-- Organization indexes
CREATE INDEX IF NOT EXISTS idx_organizations_tenant_id ON organizations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_organizations_code ON organizations(code);
CREATE INDEX IF NOT EXISTS idx_organizations_is_active ON organizations(is_active);
CREATE INDEX IF NOT EXISTS idx_organizations_created_by ON organizations(created_by);
CREATE INDEX IF NOT EXISTS idx_organizations_created_at ON organizations(created_at DESC);

-- User organization indexes
CREATE INDEX IF NOT EXISTS idx_user_organizations_user_id ON user_organizations(user_id);
CREATE INDEX IF NOT EXISTS idx_user_organizations_org_id ON user_organizations(org_id);
CREATE INDEX IF NOT EXISTS idx_user_organizations_role_id ON user_organizations(role_id);
CREATE INDEX IF NOT EXISTS idx_user_organizations_current_context ON user_organizations(user_id, is_current_context) WHERE is_current_context = true;

-- User indexes
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

-- Auth service indexes
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_user_expires ON sessions(user_id, expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON password_resets(expires_at);
CREATE INDEX IF NOT EXISTS idx_email_verification_user_id ON email_verification_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_email_verification_expires_at ON email_verification_tokens(expires_at);

-- Department indexes
CREATE INDEX IF NOT EXISTS idx_departments_name ON departments(name);
CREATE INDEX IF NOT EXISTS idx_departments_created_at ON departments(created_at DESC);

-- Designation indexes
CREATE INDEX IF NOT EXISTS idx_designations_name ON designations(name);
CREATE INDEX IF NOT EXISTS idx_designations_created_at ON designations(created_at DESC);

-- Vendor indexes
CREATE INDEX IF NOT EXISTS idx_vendors_tenant_id ON vendors(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vendors_vendor_code ON vendors(vendor_code);
CREATE INDEX IF NOT EXISTS idx_vendors_vendor_email ON vendors(vendor_email);
CREATE INDEX IF NOT EXISTS idx_vendors_is_active ON vendors(is_active);
CREATE INDEX IF NOT EXISTS idx_vendors_project ON vendors(project);
CREATE INDEX IF NOT EXISTS idx_vendors_vendor_type ON vendors(vendor_type);
CREATE INDEX IF NOT EXISTS idx_vendors_created_at ON vendors(created_at);

-- Vendor account indexes
CREATE INDEX IF NOT EXISTS idx_vendor_accounts_vendor_id ON vendor_accounts(vendor_id);
CREATE INDEX IF NOT EXISTS idx_vendor_accounts_is_primary ON vendor_accounts(is_primary);
CREATE INDEX IF NOT EXISTS idx_vendor_accounts_is_active ON vendor_accounts(is_active);
CREATE INDEX IF NOT EXISTS idx_vendor_accounts_created_at ON vendor_accounts(created_at);

-- Project indexes
CREATE INDEX IF NOT EXISTS idx_projects_tenant_id ON projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_projects_created_by ON projects(created_by);
CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at DESC);

-- ================================================
-- TRIGGERS AND FUNCTIONS
-- ================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for organizations
CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_organizations_updated_at
    BEFORE UPDATE ON user_organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Triggers for vendors
CREATE TRIGGER trigger_vendors_updated_at
    BEFORE UPDATE ON vendors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_vendor_accounts_updated_at
    BEFORE UPDATE ON vendor_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to ensure only one primary account per vendor
CREATE OR REPLACE FUNCTION ensure_single_primary_account()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_primary = true THEN
        UPDATE vendor_accounts 
        SET is_primary = false, updated_at = NOW()
        WHERE vendor_id = NEW.vendor_id 
        AND id != NEW.id 
        AND is_primary = true;
    END IF;
    
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ensure_single_primary_account
    BEFORE INSERT OR UPDATE ON vendor_accounts
    FOR EACH ROW
    EXECUTE FUNCTION ensure_single_primary_account();

-- ================================================
-- COMMENTS FOR DOCUMENTATION
-- ================================================

-- Organization service comments
COMMENT ON TABLE organizations IS 'Stores organization information for multi-tenancy';
COMMENT ON TABLE user_organizations IS 'Junction table for user-organization relationships';

-- User service comments
COMMENT ON TABLE users IS 'Stores user information';
COMMENT ON TABLE user_roles IS 'Many-to-many relationship between users and roles';
COMMENT ON TABLE user_login_history IS 'Tracks user login history';
COMMENT ON TABLE roles IS 'Stores dynamic roles (name + permissions) per tenant/organization';
COMMENT ON TABLE permissions IS 'Catalog of allowed permission keys used by roles';

-- Auth service comments
COMMENT ON TABLE sessions IS 'Active user sessions';
COMMENT ON TABLE refresh_tokens IS 'JWT refresh tokens';
COMMENT ON TABLE password_resets IS 'Password reset tokens';
COMMENT ON TABLE email_verification_tokens IS 'Email verification tokens';

-- Department service comments
COMMENT ON TABLE departments IS 'Stores department information';

-- Designation service comments
COMMENT ON TABLE designations IS 'Stores designation/position information with hierarchical structure';

-- Vendor service comments
COMMENT ON TABLE vendors IS 'Stores vendor information';
COMMENT ON TABLE vendor_accounts IS 'Stores vendor banking account details';

-- Project service comments
COMMENT ON TABLE projects IS 'Stores project information';

-- ================================================
-- MIGRATION COMPLETE
-- ================================================

-- Log successful migration
DO $$
BEGIN
    RAISE NOTICE 'NHIT Database schema migration completed successfully!';
    RAISE NOTICE 'Tables created: organizations, user_organizations, users, user_roles, user_login_history, sessions, refresh_tokens, password_resets, email_verification_tokens, departments, designations, vendors, vendor_accounts, projects';
    RAISE NOTICE 'All indexes, triggers, and constraints have been applied.';
END $$;
