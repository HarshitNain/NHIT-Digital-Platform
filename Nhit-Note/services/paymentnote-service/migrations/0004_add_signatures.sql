-- Add signature columns
ALTER TABLE payment_notes ADD COLUMN IF NOT EXISTS created_by_signature TEXT;
ALTER TABLE payment_note_approval_logs ADD COLUMN IF NOT EXISTS reviewer_signature TEXT;
ALTER TABLE payment_note_comments ADD COLUMN IF NOT EXISTS user_signature TEXT;
