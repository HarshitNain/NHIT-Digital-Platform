-- Migration to update vendor schema with enum types
-- This migration adds enum types and updates the vendors table to use them

-- Create enum types
CREATE TYPE account_type_enum AS ENUM ('INTERNAL', 'EXTERNAL');
CREATE TYPE vendor_status_enum AS ENUM ('ACTIVE', 'INACTIVE');
CREATE TYPE msme_classification_enum AS ENUM ('MICRO', 'SMALL', 'MEDIUM');

-- Update vendors table to use enums
-- First, add new columns with enum types
ALTER TABLE vendors 
  ADD COLUMN account_type_new account_type_enum;

ALTER TABLE vendors 
  ADD COLUMN status_new vendor_status_enum DEFAULT 'ACTIVE';

ALTER TABLE vendors 
  ADD COLUMN msme_classification_new msme_classification_enum;

ALTER TABLE vendors
  ADD COLUMN IF NOT EXISTS address TEXT;

ALTER TABLE vendors
  DROP COLUMN IF EXISTS country_id,
  DROP COLUMN IF EXISTS state_id,
  DROP COLUMN IF EXISTS city_id,
  DROP COLUMN IF EXISTS ifsc_code_id;

-- Migrate data from old columns to new columns
UPDATE vendors 
SET account_type_new = CASE 
  WHEN account_type = 1 THEN 'INTERNAL'::account_type_enum
  ELSE 'EXTERNAL'::account_type_enum
END
WHERE account_type IS NOT NULL;

UPDATE vendors 
SET status_new = CASE 
  WHEN status = 1 THEN 'ACTIVE'::vendor_status_enum
  ELSE 'INACTIVE'::vendor_status_enum
END;

UPDATE vendors 
SET msme_classification_new = CASE 
  WHEN msme_classification = 1 THEN 'MICRO'::msme_classification_enum
  WHEN msme_classification = 2 THEN 'SMALL'::msme_classification_enum
  WHEN msme_classification = 3 THEN 'MEDIUM'::msme_classification_enum
  ELSE NULL
END
WHERE msme_classification IS NOT NULL;

-- Drop old columns and rename new columns
ALTER TABLE vendors 
  DROP COLUMN account_type,
  DROP COLUMN status,
  DROP COLUMN msme_classification;

ALTER TABLE vendors 
  RENAME COLUMN account_type_new TO account_type;

ALTER TABLE vendors 
  RENAME COLUMN status_new TO status;

ALTER TABLE vendors 
  RENAME COLUMN msme_classification_new TO msme_classification;

-- Rename project to project_id for clarity
ALTER TABLE vendors 
  RENAME COLUMN project TO project_id;

-- Update indexes
DROP INDEX IF EXISTS idx_vendors_account_type;
DROP INDEX IF EXISTS idx_vendors_is_active; -- is_active column still exists but we might want to keep it or drop it? The migration drops it in original script but here we didn't migrate it to status_new because status column existed.
-- Original script dropped is_active. Let's see. 001 has both status (int) AND is_active (bool).
-- 002 maps is_active -> status_new. But 001 also has status. 
-- Let's assume we want to drop is_active and use status_new (enum).
-- The original 002: mapped is_active -> status_new.
-- My updated 002 above: mapped status -> status_new.
-- Let's also drop is_active if it's redundant.
ALTER TABLE vendors DROP COLUMN is_active;


CREATE INDEX idx_vendors_account_type ON vendors(account_type);
CREATE INDEX idx_vendors_status ON vendors(status);
CREATE INDEX idx_vendors_project_id ON vendors(project_id);

-- Add comments
COMMENT ON COLUMN vendors.account_type IS 'Vendor account type: INTERNAL or EXTERNAL';
COMMENT ON COLUMN vendors.status IS 'Vendor status: ACTIVE or INACTIVE';
COMMENT ON COLUMN vendors.msme_classification IS 'MSME classification: MICRO, SMALL, or MEDIUM';
COMMENT ON COLUMN vendors.project_id IS 'Project ID (UUID) associated with vendor';
