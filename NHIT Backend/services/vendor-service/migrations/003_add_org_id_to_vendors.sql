-- Add org_id to vendors table
ALTER TABLE vendors ADD COLUMN org_id UUID;
CREATE INDEX idx_vendors_org_id ON vendors(org_id);
