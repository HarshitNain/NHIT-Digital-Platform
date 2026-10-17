-- Add financial breakdown columns to payment_notes
ALTER TABLE payment_notes ADD COLUMN IF NOT EXISTS taxable_amount DECIMAL(20, 2) DEFAULT 0;
ALTER TABLE payment_notes ADD COLUMN IF NOT EXISTS gst_amount DECIMAL(20, 2) DEFAULT 0;
ALTER TABLE payment_notes ADD COLUMN IF NOT EXISTS other_charges DECIMAL(20, 2) DEFAULT 0;

-- Add column for Green Note approval history (structured text/json)
ALTER TABLE payment_notes ADD COLUMN IF NOT EXISTS green_note_history TEXT;

-- Update existing records if any
UPDATE payment_notes SET taxable_amount = gross_amount WHERE taxable_amount = 0 AND gross_amount > 0;
