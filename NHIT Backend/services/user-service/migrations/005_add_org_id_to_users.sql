-- Add org_id to users table
ALTER TABLE users ADD COLUMN org_id UUID;
CREATE INDEX idx_users_org_id ON users(org_id);
