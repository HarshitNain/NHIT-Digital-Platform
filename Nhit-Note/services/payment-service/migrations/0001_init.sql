-- Payment Status Enum
CREATE TYPE payment_status AS ENUM (
    'D',   -- Draft
    'S',   -- Submitted
    'A',   -- Approved
    'R',   -- Rejected
    'P',   -- Pending
    'C'    -- Completed
);

-- Template Type Enum
CREATE TYPE template_type AS ENUM (
    'RTGS',
    'NEFT',
    'CHEQUE',
    'CASH',
    'INTERNAL'
);

-- Main payments table
CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,
    sl_no TEXT NOT NULL,                       -- Serial number for grouping payments
    template_type template_type NOT NULL,
    project TEXT,
    account_full_name TEXT,
    from_account_type TEXT,
    full_account_number TEXT,
    to_account TEXT,                            -- Renamed from 'to' to avoid SQL keyword
    to_account_type TEXT,
    name_of_beneficiary TEXT,
    account_number TEXT,
    name_of_bank TEXT,
    ifsc_code TEXT,
    amount DECIMAL(20, 2) NOT NULL DEFAULT 0,
    purpose TEXT,
    status payment_status DEFAULT 'D',
    user_id BIGINT NOT NULL,
    payment_note_id BIGINT,                    -- Reference to payment note
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payment vendor accounts junction table (many-to-many)
CREATE TABLE IF NOT EXISTS payment_vendor_accounts (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    vendor_id BIGINT NOT NULL,                 -- External vendor ID
    vendor_account_id BIGINT,                  -- Specific vendor account
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(payment_id, vendor_id)
);

-- Payment shortcuts table
CREATE TABLE IF NOT EXISTS payment_shortcuts (
    id BIGSERIAL PRIMARY KEY,
    sl_no TEXT,
    shortcut_name TEXT NOT NULL,
    request_data_json TEXT NOT NULL,          -- JSON payload for recreating payments
    user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Bank letter approval logs
CREATE TABLE IF NOT EXISTS bank_letter_approval_logs (
    id BIGSERIAL PRIMARY KEY,
    sl_no TEXT NOT NULL,
    status TEXT NOT NULL,
    comments TEXT,
    reviewer_id BIGINT NOT NULL,
    reviewer_name TEXT,
    reviewer_email TEXT,
    approver_level INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_payments_sl_no ON payments(sl_no);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_payment_note_id ON payments(payment_note_id);
CREATE INDEX IF NOT EXISTS idx_payments_template_type ON payments(template_type);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payment_vendor_accounts_payment_id ON payment_vendor_accounts(payment_id);
CREATE INDEX IF NOT EXISTS idx_payment_vendor_accounts_vendor_id ON payment_vendor_accounts(vendor_id);

CREATE INDEX IF NOT EXISTS idx_payment_shortcuts_user_id ON payment_shortcuts(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_shortcuts_sl_no ON payment_shortcuts(sl_no);

CREATE INDEX IF NOT EXISTS idx_bank_letter_logs_sl_no ON bank_letter_approval_logs(sl_no);

-- Function to update updated_at column
CREATE OR REPLACE FUNCTION update_payment_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for updated_at on payments
CREATE TRIGGER update_payments_modtime
BEFORE UPDATE ON payments
FOR EACH ROW EXECUTE FUNCTION update_payment_modified_column();

-- Trigger for updated_at on payment_shortcuts
CREATE TRIGGER update_payment_shortcuts_modtime
BEFORE UPDATE ON payment_shortcuts
FOR EACH ROW EXECUTE FUNCTION update_payment_modified_column();
