-- Migration: Add document_type to green_note_documents
ALTER TABLE green_note_documents ADD COLUMN IF NOT EXISTS document_type TEXT;
CREATE INDEX IF NOT EXISTS idx_green_note_documents_type ON green_note_documents(document_type);
