-- Add org_id to vendors table
ALTER TABLE vendors ADD COLUMN IF NOT EXISTS org_id UUID;

-- Update vendors unique constraints to include org_id
ALTER TABLE vendors DROP CONSTRAINT IF EXISTS vendors_tenant_vendor_code_unique;
ALTER TABLE vendors ADD CONSTRAINT vendors_org_vendor_code_unique UNIQUE (org_id, vendor_code);

ALTER TABLE vendors DROP CONSTRAINT IF EXISTS vendors_tenant_vendor_email_unique;
ALTER TABLE vendors ADD CONSTRAINT vendors_org_vendor_email_unique UNIQUE (org_id, vendor_email);

-- Update indexes
CREATE INDEX IF NOT EXISTS idx_users_org_id ON users(org_id);
CREATE INDEX IF NOT EXISTS idx_vendors_org_id ON vendors(org_id);
