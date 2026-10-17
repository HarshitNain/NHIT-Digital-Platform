-- Password Reset table modifications to support OTP-based resets

-- Add new columns to password_resets table
ALTER TABLE password_resets
ADD COLUMN IF NOT EXISTS id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
ADD COLUMN IF NOT EXISTS otp VARCHAR(10) NULL,
ADD COLUMN IF NOT EXISTS reset_type VARCHAR(10) NOT NULL DEFAULT 'token',
ADD COLUMN IF NOT EXISTS used BOOLEAN NOT NULL DEFAULT FALSE;

-- Migration for existing data
-- Set ID for existing records
UPDATE password_resets
SET id = gen_random_uuid(), reset_type = 'token'
WHERE id IS NULL;

-- Create index on OTP and user_id
CREATE INDEX IF NOT EXISTS idx_password_resets_otp_user_id ON password_resets(otp, user_id);

-- Create index on reset_type
CREATE INDEX IF NOT EXISTS idx_password_resets_reset_type ON password_resets(reset_type);

COMMENT ON COLUMN password_resets.id IS 'Unique identifier for the password reset record';
COMMENT ON COLUMN password_resets.otp IS 'One-time password for OTP-based password resets';
COMMENT ON COLUMN password_resets.reset_type IS 'Type of reset: token or otp';
COMMENT ON COLUMN password_resets.used IS 'Whether this reset token has been used';
