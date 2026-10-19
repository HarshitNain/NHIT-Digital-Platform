-- Migration: 006_add_must_change_password_to_users.sql
-- Enforces mandatory password update upon initial login for IAM provisioned users (OPEN-03).
ALTER TABLE users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN DEFAULT FALSE;
