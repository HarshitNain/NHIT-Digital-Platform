-- Department Service Database Schema

-- Departments table
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_departments_name ON departments(name);
CREATE INDEX IF NOT EXISTS idx_departments_created_at ON departments(created_at DESC);

-- Comments for documentation
COMMENT ON TABLE departments IS 'Stores department information';
COMMENT ON COLUMN departments.id IS 'Unique identifier for the department';
COMMENT ON COLUMN departments.name IS 'Department name (unique)';
COMMENT ON COLUMN departments.description IS 'Department description';
