-- Payment Note Status Enum
CREATE TYPE payment_note_status AS ENUM (
    'D',   -- Draft
    'P',   -- Pending Approval
    'A',   -- Approved
    'R',   -- Rejected
    'S',   -- Submitted
    'PD',  -- Pending Draft
    'H'    -- On Hold
);

-- Main payment_notes table
CREATE TABLE IF NOT EXISTS payment_notes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,                    -- Owner/creator of payment note
    
    -- Green Note Reference
    green_note_id TEXT,                         -- Reference to green note (UUID as text for cross-service)
    green_note_no TEXT,                         -- Green note number
    green_note_approver TEXT,                   -- Name of green note approver
    green_note_app_date TEXT,                   -- Green note approval date
    
    -- Reimbursement Reference
    reimbursement_note_id BIGINT,               -- Reference to reimbursement note (nullable)
    
    -- Payment Note Details
    note_no TEXT NOT NULL UNIQUE,               -- Payment note number
    subject TEXT,
    date TIMESTAMPTZ,                           -- Payment note date
    department TEXT,                            -- Department name
    
    -- Vendor Details
    vendor_code TEXT,
    vendor_name TEXT,
    
    -- Project Details
    project_name TEXT,
    
    -- Invoice Details
    invoice_no TEXT,
    invoice_date TEXT,
    invoice_amount DECIMAL(20, 2),
    invoice_approved_by TEXT,
    
    -- LOA/PO Details
    loa_po_no TEXT,
    loa_po_amount DECIMAL(20, 2),
    loa_po_date TEXT,
    
    -- Financial Calculations
    gross_amount DECIMAL(20, 2) NOT NULL DEFAULT 0,              -- Gross invoice amount
    total_additions DECIMAL(20, 2) NOT NULL DEFAULT 0,           -- Sum of add particulars
    total_deductions DECIMAL(20, 2) NOT NULL DEFAULT 0,          -- Sum of less particulars
    net_payable_amount DECIMAL(20, 2) NOT NULL DEFAULT 0,        -- Net payable after calculations
    net_payable_round_off DECIMAL(20, 2) NOT NULL DEFAULT 0,     -- Rounded off amount
    net_payable_words TEXT,                                      -- Amount in words
    
    -- TDS Details
    tds_percentage DECIMAL(5, 2),               -- TDS percentage (e.g., 2.00 for 2%)
    tds_section TEXT,                           -- TDS section (e.g., 194C)
    tds_amount DECIMAL(20, 2),                  -- Calculated TDS amount
    
    -- Bank Details
    account_holder_name TEXT,
    bank_name TEXT,
    account_number TEXT,
    ifsc_code TEXT,
    
    -- Recommendation
    recommendation_of_payment TEXT,
    
    -- Status and Flags
    status payment_note_status DEFAULT 'D',
    is_draft BOOLEAN NOT NULL DEFAULT FALSE,
    auto_created BOOLEAN NOT NULL DEFAULT FALSE, -- Flag for auto-created drafts from green note
    created_by BIGINT,                          -- User who created (may differ from user_id)
    
    -- Hold information
    hold_reason TEXT,
    hold_date TIMESTAMPTZ,
    hold_by BIGINT,                             -- User who put on hold
    
    -- UTR (Unique Transaction Reference) information
    utr_no TEXT,
    utr_date TEXT,                              -- Stored as ISO-8601 string
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT check_note_no_not_empty CHECK (note_no <> '')
);

-- Payment note particulars (add/less items)
CREATE TABLE IF NOT EXISTS payment_note_particulars (
    id BIGSERIAL PRIMARY KEY,
    payment_note_id BIGINT NOT NULL REFERENCES payment_notes(id) ON DELETE CASCADE,
    particular_type TEXT NOT NULL,              -- 'ADD' or 'LESS'
    particular TEXT NOT NULL,                   -- Description
    amount DECIMAL(20, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT check_particular_type CHECK (particular_type IN ('ADD', 'LESS'))
);

-- Payment note approval logs
CREATE TABLE IF NOT EXISTS payment_note_approval_logs (
    id BIGSERIAL PRIMARY KEY,
    payment_note_id BIGINT NOT NULL REFERENCES payment_notes(id) ON DELETE CASCADE,
    status TEXT NOT NULL,                       -- A, R, P, etc.
    comments TEXT,
    reviewer_id BIGINT NOT NULL,                -- User who approved/rejected
    reviewer_name TEXT,
    reviewer_email TEXT,
    approver_level INT,                         -- Approval priority level
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payment note comments
CREATE TABLE IF NOT EXISTS payment_note_comments (
    id BIGSERIAL PRIMARY KEY,
    payment_note_id BIGINT NOT NULL REFERENCES payment_notes(id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    status TEXT,                                -- Status when comment was made
    user_id BIGINT NOT NULL,
    user_name TEXT,
    user_email TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payment note supporting documents
CREATE TABLE IF NOT EXISTS payment_note_documents (
    id BIGSERIAL PRIMARY KEY,
    payment_note_id BIGINT NOT NULL REFERENCES payment_notes(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    mime_type TEXT,
    file_size BIGINT,
    object_key TEXT NOT NULL,                   -- S3/storage key
    uploaded_by BIGINT NOT NULL,
    uploaded_by_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_payment_notes_user_id ON payment_notes(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_notes_green_note_id ON payment_notes(green_note_id);
CREATE INDEX IF NOT EXISTS idx_payment_notes_green_note_no ON payment_notes(green_note_no);
CREATE INDEX IF NOT EXISTS idx_payment_notes_reimbursement_note_id ON payment_notes(reimbursement_note_id);
CREATE INDEX IF NOT EXISTS idx_payment_notes_status ON payment_notes(status);
CREATE INDEX IF NOT EXISTS idx_payment_notes_is_draft ON payment_notes(is_draft);
CREATE INDEX IF NOT EXISTS idx_payment_notes_created_at ON payment_notes(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_notes_note_no ON payment_notes(note_no);
CREATE INDEX IF NOT EXISTS idx_payment_notes_vendor_code ON payment_notes(vendor_code);
CREATE INDEX IF NOT EXISTS idx_payment_notes_project_name ON payment_notes(project_name);

CREATE INDEX IF NOT EXISTS idx_payment_note_particulars_note_id ON payment_note_particulars(payment_note_id);
CREATE INDEX IF NOT EXISTS idx_payment_note_approval_logs_note_id ON payment_note_approval_logs(payment_note_id);
CREATE INDEX IF NOT EXISTS idx_payment_note_comments_note_id ON payment_note_comments(payment_note_id);
CREATE INDEX IF NOT EXISTS idx_payment_note_documents_note_id ON payment_note_documents(payment_note_id);

-- Function to update updated_at column
CREATE OR REPLACE FUNCTION update_payment_note_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for updated_at
CREATE TRIGGER update_payment_notes_modtime
BEFORE UPDATE ON payment_notes
FOR EACH ROW EXECUTE FUNCTION update_payment_note_modified_column();
