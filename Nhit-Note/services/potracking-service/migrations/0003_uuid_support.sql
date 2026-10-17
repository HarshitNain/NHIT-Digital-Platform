-- Support UUIDs for PO Tracking and Payments
-- Convert BIGINT columns to TEXT

-- 1. PO Tracking Table
ALTER TABLE po_tracking ALTER COLUMN vendor_id TYPE TEXT;
ALTER TABLE po_tracking ALTER COLUMN supplier_id TYPE TEXT;

-- 2. Payments Table
ALTER TABLE payments ALTER COLUMN user_id TYPE TEXT;

-- 3. Payment Vendor Accounts Table
ALTER TABLE payment_vendor_accounts ALTER COLUMN vendor_id TYPE TEXT;

-- 4. Payment Shortcuts Table
ALTER TABLE payment_shortcuts ALTER COLUMN user_id TYPE TEXT;

-- 5. Bank Letter Approval Logs
ALTER TABLE bank_letter_approval_logs ALTER COLUMN reviewer_id TYPE TEXT;
