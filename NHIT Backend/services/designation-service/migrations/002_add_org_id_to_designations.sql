-- Add org_id column to designations table for organization-specific filtering
-- This migration adds the org_id column to enable multi-organization support

ALTER TABLE designations ADD COLUMN IF NOT EXISTS org_id UUID;

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_designations_org_id ON designations(org_id);

-- Add comment for documentation
COMMENT ON COLUMN designations.org_id IS 'Organization ID for organization-specific designations (NULL for global designations)';
