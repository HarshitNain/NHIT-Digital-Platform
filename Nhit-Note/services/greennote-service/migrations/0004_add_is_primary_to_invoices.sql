-- Migration to distinguish primary invoice from multiple invoice items
ALTER TABLE green_note_invoices ADD COLUMN IF NOT EXISTS is_primary BOOLEAN DEFAULT FALSE;

-- Index for faster lookup of primary vs non-primary invoices
CREATE INDEX IF NOT EXISTS idx_green_note_invoices_is_primary ON green_note_invoices(is_primary);
