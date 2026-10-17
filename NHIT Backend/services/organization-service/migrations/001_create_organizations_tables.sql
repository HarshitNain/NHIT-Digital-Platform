-- ================================================
-- Organizations Table (Matches Proto Definition)
-- ================================================

CREATE TABLE IF NOT EXISTS organizations (
    org_id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(tenant_id),

    -- Parent Org = NULL
    -- Child Org = parent_org_id = parent organization's UUID
    parent_org_id UUID,

    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    database_name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    logo VARCHAR(500),

    -- Super admin details (ONLY for parent orgs)
    super_admin_name VARCHAR(255),
    super_admin_email VARCHAR(255),
    super_admin_password VARCHAR(255),

    -- Array of project strings
    initial_projects TEXT[],

    -- Status field: 0 = activated, 1 = deactivated
    status SMALLINT NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key for parent org (self reference)
    CONSTRAINT fk_parent_org
        FOREIGN KEY (parent_org_id)
        REFERENCES organizations(org_id)
        ON DELETE SET NULL
);

-- Indexes
CREATE INDEX idx_organizations_tenant_id ON organizations(tenant_id);
CREATE INDEX idx_organizations_code ON organizations(code);
CREATE INDEX idx_organizations_status ON organizations(status);
CREATE INDEX idx_organizations_parent_org_id ON organizations(parent_org_id);
CREATE INDEX idx_organizations_created_at ON organizations(created_at DESC);

-- ================================================
-- User Organizations Table (Keeps Existing)
-- ================================================

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

-- Indexes
CREATE INDEX idx_user_orgs_user_id ON user_organizations(user_id);
CREATE INDEX idx_user_orgs_org_id ON user_organizations(org_id);
CREATE INDEX idx_user_orgs_role_id ON user_organizations(role_id);
CREATE INDEX idx_user_orgs_current_context ON user_organizations(user_id, is_current_context)
    WHERE is_current_context = true;

-- ================================================
-- Trigger to Auto-update updated_at
-- ================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_orgs_updated_at
    BEFORE UPDATE ON user_organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ================================================
-- Documentation Comments
-- ================================================

COMMENT ON TABLE organizations IS 'Stores parent and child organizations. Matches proto Organization message.';

COMMENT ON COLUMN organizations.parent_org_id IS 'NULL for parent orgs; contains parent org UUID for child orgs.';

COMMENT ON COLUMN organizations.super_admin_name IS 'Only filled for parent orgs. Represents initial super admin for system onboarding.';
COMMENT ON COLUMN organizations.super_admin_email IS 'Super admin email for parent orgs.';
COMMENT ON COLUMN organizations.super_admin_password IS 'Super admin password for parent orgs.';

COMMENT ON COLUMN organizations.initial_projects IS 'List of initial projects stored as TEXT[]';

COMMENT ON COLUMN organizations.status IS '0 = activated, 1 = deactivated (matches proto enum OrganizationStatus)';

COMMENT ON TABLE user_organizations IS 'Junction table mapping users to organizations.';
