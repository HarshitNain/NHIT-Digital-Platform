-- Migration: Add activity_logs table (matching UI structure)
-- Created: 2025-12-08

CREATE TABLE IF NOT EXISTS activity_logs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at);

COMMENT ON TABLE activity_logs IS 'Stores user activity logs for audit trail';
COMMENT ON COLUMN activity_logs.id IS 'Auto-incrementing ID';
COMMENT ON COLUMN activity_logs.name IS 'Activity name/title (e.g., "User Logged in with email-ID [user@example.com]")';
COMMENT ON COLUMN activity_logs.description IS 'Activity description (e.g., "User successfully logged in")';
COMMENT ON COLUMN activity_logs.created_at IS 'Timestamp when activity occurred';
