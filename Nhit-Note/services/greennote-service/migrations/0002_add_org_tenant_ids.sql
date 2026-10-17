-- Migration to add org_id and tenant_id to green_notes table
ALTER TABLE green_notes ADD COLUMN IF NOT EXISTS org_id UUID;
ALTER TABLE green_notes ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_green_notes_org_id ON green_notes(org_id);
CREATE INDEX IF NOT EXISTS idx_green_notes_tenant_id ON green_notes(tenant_id);

-- Update existing records with default test IDs if applicable
UPDATE green_notes SET 
    tenant_id = '550e8400-e29b-41d4-a716-446655440000',
    org_id = '550e8400-e29b-41d4-a716-446655440001'
WHERE tenant_id IS NULL OR org_id IS NULL;
