-- Migration: Add approver fields to green_notes
ALTER TABLE green_notes 
ADD COLUMN IF NOT EXISTS approver_name TEXT,
ADD COLUMN IF NOT EXISTS approver_designation TEXT,
ADD COLUMN IF NOT EXISTS approver_role TEXT,
ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
