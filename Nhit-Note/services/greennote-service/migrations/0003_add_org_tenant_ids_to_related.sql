-- Migration to add multi-tenancy support to related green note tables
ALTER TABLE green_note_invoices ADD COLUMN IF NOT EXISTS org_id UUID;
ALTER TABLE green_note_invoices ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE green_note_documents ADD COLUMN IF NOT EXISTS org_id UUID;
ALTER TABLE green_note_documents ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_green_note_invoices_org_id ON green_note_invoices(org_id);
CREATE INDEX IF NOT EXISTS idx_green_note_documents_org_id ON green_note_documents(org_id);

-- Update existing records if any
UPDATE green_note_invoices SET 
    tenant_id = '550e8400-e29b-41d4-a716-446655440000',
    org_id = '550e8400-e29b-41d4-a716-446655440001'
WHERE tenant_id IS NULL OR org_id IS NULL;

UPDATE green_note_documents SET 
    tenant_id = '550e8400-e29b-41d4-a716-446655440000',
    org_id = '550e8400-e29b-41d4-a716-446655440001'
WHERE tenant_id IS NULL OR org_id IS NULL;
