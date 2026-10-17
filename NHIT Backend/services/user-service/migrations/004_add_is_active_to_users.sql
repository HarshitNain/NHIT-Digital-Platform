-- Add is_active and deactivation fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deactivated_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deactivated_by UUID;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deactivated_by_name VARCHAR(255);

-- Comments
COMMENT ON COLUMN users.is_active IS 'Soft delete flag (TRUE = active, FALSE = deactivated)';
COMMENT ON COLUMN users.deactivated_at IS 'When the user was deactivated';
COMMENT ON COLUMN users.deactivated_by IS 'Who deactivated the user';
COMMENT ON COLUMN users.deactivated_by_name IS 'Name of the user who deactivated the user';
