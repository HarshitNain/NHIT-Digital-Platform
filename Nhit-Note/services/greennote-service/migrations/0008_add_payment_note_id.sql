-- Migration to link Green Note to its generated Payment Note
ALTER TABLE green_notes ADD COLUMN payment_note_id TEXT;
