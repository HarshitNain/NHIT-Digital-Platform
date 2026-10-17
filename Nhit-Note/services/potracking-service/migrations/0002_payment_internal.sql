-- Payment Status Enum
DO $$ BEGIN
    CREATE TYPE payment_status AS ENUM (
        'D',   -- Draft
        'S',   -- Submitted
        'A',   -- Approved
        'R',   -- Rejected
        'P',   -- PendingLevel1
        'C'    -- Completed
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Template Type Enum
DO $$ BEGIN
    CREATE TYPE template_type AS ENUM (
        'RTGS',
        'NEFT',
        'CHEQUE',
        'CASH',
        'INTERNAL'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Main payments table
CREATE TABLE IF NOT EXISTS payments (
    id BIGSERIAL PRIMARY KEY,
    sl_no TEXT NOT NULL,
    template_type template_type NOT NULL,
    project TEXT,
    account_full_name TEXT,
    from_account_type TEXT,
    full_account_number TEXT,
    to_account TEXT,
    to_account_type TEXT,
    name_of_beneficiary TEXT,
    account_number TEXT,
    name_of_bank TEXT,
    ifsc_code TEXT,
    amount DECIMAL(20, 2) NOT NULL DEFAULT 0,
    purpose TEXT,
    status payment_status DEFAULT 'D',
    user_id BIGINT NOT NULL,
    payment_note_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payment vendor accounts junction table
CREATE TABLE IF NOT EXISTS payment_vendor_accounts (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    vendor_id BIGINT NOT NULL,
    vendor_account_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(payment_id, vendor_id)
);

-- Payment shortcuts table
CREATE TABLE IF NOT EXISTS payment_shortcuts (
    id BIGSERIAL PRIMARY KEY,
    sl_no TEXT,
    shortcut_name TEXT NOT NULL,
    request_data_json TEXT NOT NULL,
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

-- Indexes
CREATE INDEX IF NOT EXISTS idx_payments_sl_no ON payments(sl_no);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_shortcuts_user_id ON payment_shortcuts(user_id);

-- Trigger for updated_at on payments
CREATE TRIGGER update_payments_modtime
BEFORE UPDATE ON payments
FOR EACH ROW EXECUTE FUNCTION update_po_tracking_modified_column();

-- Trigger for updated_at on payment_shortcuts
CREATE TRIGGER update_payment_shortcuts_modtime
BEFORE UPDATE ON payment_shortcuts
FOR EACH ROW EXECUTE FUNCTION update_po_tracking_modified_column();
