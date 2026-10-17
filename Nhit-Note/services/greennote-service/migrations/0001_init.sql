CREATE TYPE approval_for AS ENUM (
    'APPROVAL_FOR_UNSPECIFIED',
    'APPROVAL_FOR_INVOICE',
    'APPROVAL_FOR_ADVANCE',
    'APPROVAL_FOR_ADHOC'
);

CREATE TYPE expense_category_type AS ENUM (
    'EXPENSE_CATEGORY_UNSPECIFIED',
    'EXPENSE_CATEGORY_CAPITAL',
    'EXPENSE_CATEGORY_REVENUE',
    'EXPENSE_CATEGORY_OPERATIONAL',
    'EXPENSE_CATEGORY_ADMINISTRATIVE',
    'EXPENSE_CATEGORY_MAINTENANCE'
);

CREATE TYPE nature_of_expenses AS ENUM (
    'NATURE_OF_EXPENSES_UNSPECIFIED',
    'NATURE_OHC_001_MANPOWER',
    'NATURE_OHC_002_STAFF_WELFARE',
    'NATURE_OHC_003_OFFICE_RENT_UTILITIES'
);

CREATE TYPE status_enum AS ENUM (
    'STATUS_APPROVED',
    'STATUS_PENDING',
    'STATUS_REJECTED',
    'STATUS_DRAFT'
);

CREATE TYPE yes_no_enum AS ENUM (
    'YES_NO_UNSPECIFIED',
    'YES',
    'NO'
);

-- Main green_notes table
CREATE TABLE IF NOT EXISTS green_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_name TEXT,
    supplier_name TEXT,
    expense_category TEXT,
    
    -- Yes/No fields
    protest_note_raised yes_no_enum DEFAULT 'YES_NO_UNSPECIFIED',
    whether_contract yes_no_enum DEFAULT 'YES_NO_UNSPECIFIED',
    extension_of_contract_period_executed yes_no_enum DEFAULT 'YES_NO_UNSPECIFIED',
    expense_amount_within_contract yes_no_enum DEFAULT 'YES_NO_UNSPECIFIED',
    milestone_achieved yes_no_enum DEFAULT 'YES_NO_UNSPECIFIED',
    payment_approved_with_deviation yes_no_enum DEFAULT 'YES_NO_UNSPECIFIED',
    required_documents_submitted yes_no_enum DEFAULT 'YES_NO_UNSPECIFIED',
    contract_period_completed yes_no_enum DEFAULT 'YES_NO_UNSPECIFIED',
    
    -- Date fields
    documents_verified TEXT,
    contract_start_date TEXT,
    contract_end_date TEXT,
    appointed_start_date TEXT,
    supply_period_start TEXT,
    supply_period_end TEXT,
    
    -- Financial fields
    base_value DECIMAL(20, 2) DEFAULT 0,
    other_charges DECIMAL(20, 2) DEFAULT 0,
    gst DECIMAL(20, 2) DEFAULT 0,
    total_amount DECIMAL(20, 2) DEFAULT 0,

    -- Invoice settings
    enable_multiple_invoices BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Status and type
    status TEXT DEFAULT 'pending',
    approval_for TEXT,
    
    -- Project details
    department_name TEXT,
    work_order_no TEXT,
    po_number TEXT,
    work_order_date TEXT,
    
    -- Expense and classification
    expense_category_type TEXT,
    msme_classification TEXT,
    activity_type TEXT,
    
    -- Additional information
    brief_of_goods_services TEXT,
    delayed_damages TEXT,
    nature_of_expenses TEXT,
    
    -- Budget information
    budget_expenditure DECIMAL(20, 2) DEFAULT 0,
    actual_expenditure DECIMAL(20, 2) DEFAULT 0,
    expenditure_over_budget DECIMAL(20, 2) DEFAULT 0,
    
    -- Contract and compliance
    milestone_remarks TEXT,
    specify_deviation TEXT,
    
    -- HR department
    documents_workdone_supply TEXT,
    documents_discrepancy TEXT,
    remarks TEXT,
    auditor_remarks TEXT,
    amount_retained_for_non_submission DECIMAL(20, 2) DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_green_notes_status ON green_notes(status);
CREATE INDEX IF NOT EXISTS idx_green_notes_created_at ON green_notes(created_at);
CREATE INDEX IF NOT EXISTS idx_green_notes_supplier_name ON green_notes(supplier_name);
CREATE INDEX IF NOT EXISTS idx_green_notes_project_name ON green_notes(project_name);


-- ================
-- green_note_invoices (UUID PK + FK -> green_notes.id)
-- ================
CREATE TABLE IF NOT EXISTS green_note_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    green_note_id UUID NOT NULL REFERENCES green_notes(id) ON DELETE CASCADE,
    
    -- Invoice fields
    invoice_number TEXT NOT NULL,
    invoice_date TEXT,
    taxable_value DECIMAL(20, 2) NOT NULL DEFAULT 0,
    gst DECIMAL(20, 2) NOT NULL DEFAULT 0,
    other_charges DECIMAL(20, 2) NOT NULL DEFAULT 0,
    invoice_value DECIMAL(20, 2) NOT NULL DEFAULT 0,
    description TEXT,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_green_note_invoices_note_id ON green_note_invoices(green_note_id);
CREATE INDEX IF NOT EXISTS idx_green_note_invoices_invoice_number ON green_note_invoices(invoice_number);

-- ================
-- green_note_documents (UUID PK + FK -> green_notes.id)
-- ================
CREATE TABLE IF NOT EXISTS green_note_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    green_note_id UUID NOT NULL REFERENCES green_notes(id) ON DELETE CASCADE,
    
    -- Document fields
    name TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    object_key TEXT NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_green_note_documents_note_id ON green_note_documents(green_note_id);
CREATE INDEX IF NOT EXISTS idx_green_note_documents_created_at ON green_note_documents(created_at);

-- Function to update updated_at
CREATE OR REPLACE FUNCTION update_modified_column() 
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW; 
END;
$$ LANGUAGE plpgsql;

-- Trigger for updated_at
CREATE TRIGGER update_green_note_documents_modtime
BEFORE UPDATE ON green_note_documents
FOR EACH ROW EXECUTE FUNCTION update_modified_column();

-- ================
-- order_sequences (ye int hi reh sakta hai)
-- ================
CREATE TABLE IF NOT EXISTS order_sequences (
    prefix TEXT PRIMARY KEY,
    current_value BIGINT NOT NULL DEFAULT 0
);