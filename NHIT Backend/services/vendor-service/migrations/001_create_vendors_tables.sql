-- Create vendors table
CREATE TABLE IF NOT EXISTS vendors (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    vendor_code VARCHAR(100) UNIQUE NOT NULL,
    vendor_name VARCHAR(255) NOT NULL,
    vendor_email VARCHAR(255) NOT NULL,
    vendor_mobile VARCHAR(20),
    account_type INTEGER NOT NULL DEFAULT 2, -- 1=INTERNAL, 2=EXTERNAL
    vendor_nick_name VARCHAR(100),
    activity_type VARCHAR(100),
    gstin VARCHAR(20),
    pan VARCHAR(10) NOT NULL,
    pin VARCHAR(10),
    country_name VARCHAR(100),
    state_name VARCHAR(100),
    city_name VARCHAR(100),
    msme_classification INTEGER DEFAULT 0, -- 0=UNSPECIFIED, 1=MICRO, 2=SMALL, 3=MEDIUM
    msme VARCHAR(50),
    msme_registration_number VARCHAR(100),
    msme_start_date TIMESTAMP,
    msme_end_date TIMESTAMP,
    material_nature VARCHAR(255),
    gst_defaulted VARCHAR(50),
    section_206ab_verified VARCHAR(50),
    beneficiary_name VARCHAR(255) NOT NULL,
    remarks_address TEXT,
    common_bank_details TEXT,
    income_tax_type VARCHAR(50),
    project VARCHAR(255),
    status INTEGER NOT NULL DEFAULT 1, -- 1=ACTIVE, 2=INACTIVE
    from_account_type VARCHAR(100),
    account_name VARCHAR(255),
    short_name VARCHAR(100),
    parent VARCHAR(255),
    code_auto_generated BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Legacy banking fields (for backward compatibility)
    account_number VARCHAR(50),
    name_of_bank VARCHAR(255),
    ifsc_code VARCHAR(15),
    document_url TEXT
);

-- Create vendor_accounts table
CREATE TABLE IF NOT EXISTS vendor_accounts (
    id UUID PRIMARY KEY,
    vendor_id UUID NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    account_name VARCHAR(255) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    account_type VARCHAR(50),
    name_of_bank VARCHAR(255) NOT NULL,
    branch_name VARCHAR(255),
    ifsc_code VARCHAR(15) NOT NULL,
    swift_code VARCHAR(20),
    is_primary BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    remarks TEXT,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_vendors_tenant_id ON vendors(tenant_id);
CREATE INDEX idx_vendors_vendor_code ON vendors(vendor_code);
CREATE INDEX idx_vendors_vendor_email ON vendors(vendor_email);
CREATE INDEX idx_vendors_account_type ON vendors(account_type);
CREATE INDEX idx_vendors_is_active ON vendors(is_active);
CREATE INDEX idx_vendor_accounts_vendor_id ON vendor_accounts(vendor_id);
CREATE INDEX idx_vendor_accounts_is_primary ON vendor_accounts(is_primary);

-- Ensure only one primary account per vendor
CREATE UNIQUE INDEX idx_vendor_accounts_unique_primary 
ON vendor_accounts(vendor_id) 
WHERE is_primary = TRUE AND is_active = TRUE;
