-- Migration to make GreenNote fields dynamic
-- 1. Alter columns to TEXT to avoid enum restrictions
ALTER TABLE green_notes ALTER COLUMN status TYPE TEXT USING status::TEXT;
ALTER TABLE green_notes ALTER COLUMN status SET DEFAULT 'pending';

ALTER TABLE green_notes ALTER COLUMN approval_for TYPE TEXT USING approval_for::TEXT;
ALTER TABLE green_notes ALTER COLUMN expense_category_type TYPE TEXT USING expense_category_type::TEXT;
ALTER TABLE green_notes ALTER COLUMN nature_of_expenses TYPE TEXT USING nature_of_expenses::TEXT;

-- Update existing NULL or empty statuses to 'pending'
UPDATE green_notes SET status = 'pending' WHERE status IS NULL OR status = '';
