-- Add banking information and signature columns to users table
-- This migration adds fields for banking details and signature file storage

-- Add banking information columns
ALTER TABLE users ADD COLUMN IF NOT EXISTS account_holder_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS bank_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS bank_account_number VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS ifsc_code VARCHAR(11);

-- Add signature column (stores MinIO file path/URL)
ALTER TABLE users ADD COLUMN IF NOT EXISTS signature_url VARCHAR(500);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_bank_account ON users(bank_account_number);
CREATE INDEX IF NOT EXISTS idx_users_ifsc_code ON users(ifsc_code);

-- Add comments for documentation
COMMENT ON COLUMN users.account_holder_name IS 'Bank account holder name';
COMMENT ON COLUMN users.bank_name IS 'Name of the bank';
COMMENT ON COLUMN users.bank_account_number IS 'Bank account number';
COMMENT ON COLUMN users.ifsc_code IS 'IFSC code of the bank branch (11 characters)';
COMMENT ON COLUMN users.signature_url IS 'MinIO URL/path to user signature image';
