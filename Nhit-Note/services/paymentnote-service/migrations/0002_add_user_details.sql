-- Add user details to payment_notes
ALTER TABLE payment_notes ADD COLUMN created_by_name TEXT;
ALTER TABLE payment_notes ADD COLUMN created_by_designation TEXT;
ALTER TABLE payment_notes ADD COLUMN created_by_role TEXT;

-- Add user details to approval logs
ALTER TABLE payment_note_approval_logs ADD COLUMN reviewer_designation TEXT;
ALTER TABLE payment_note_approval_logs ADD COLUMN reviewer_role TEXT;

-- Add user details to comments
ALTER TABLE payment_note_comments ADD COLUMN user_designation TEXT;
ALTER TABLE payment_note_comments ADD COLUMN user_role TEXT;
