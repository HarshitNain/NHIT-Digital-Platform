-- Add approval_levels column to green_notes table with JSONB type
ALTER TABLE green_notes ADD COLUMN approval_levels JSONB DEFAULT '{"history": []}'::jsonb;
