-- Add org_id column to departments table for organization-specific filtering
-- This migration adds the org_id column to enable multi-organization support

ALTER TABLE departments ADD COLUMN IF NOT EXISTS org_id UUID;

-- Create index for better query performance
CREATE INDEX IF NOT EXISTS idx_departments_org_id ON departments(org_id);

-- Add comment for documentation
COMMENT ON COLUMN departments.org_id IS 'Organization ID for organization-specific departments (NULL for global departments)';
