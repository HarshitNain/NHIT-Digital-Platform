-- Designation Service Database Schema

-- Designations table
CREATE TABLE IF NOT EXISTS designations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(250) NOT NULL,
    description VARCHAR(500) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create unique index for case-insensitive name
CREATE UNIQUE INDEX IF NOT EXISTS idx_designations_name_lower ON designations (LOWER(name));

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_designations_name ON designations(name);

-- Comments for documentation
COMMENT ON TABLE designations IS 'Stores designation details';
COMMENT ON COLUMN designations.id IS 'Unique identifier for the designation';
COMMENT ON COLUMN designations.name IS 'Designation name';
