-- ================================================
-- NHIT Backend - Database Reset Script
-- ================================================
-- This script drops all tables and data for a clean reset
-- WARNING: This will delete ALL data in the database!

-- ================================================
-- DROP ALL TABLES (in correct order to handle foreign keys)
-- ================================================

-- Drop tables with foreign key dependencies first
DROP TABLE IF EXISTS user_organizations CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS user_login_history CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS password_resets CASCADE;
DROP TABLE IF EXISTS email_verification_tokens CASCADE;
DROP TABLE IF EXISTS vendor_accounts CASCADE;

-- Drop main tables
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;
DROP TABLE IF EXISTS departments CASCADE;
DROP TABLE IF EXISTS designations CASCADE;
DROP TABLE IF EXISTS vendors CASCADE;
DROP TABLE IF EXISTS projects CASCADE;

-- ================================================
-- DROP ALL FUNCTIONS AND TRIGGERS
-- ================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS update_organizations_updated_at ON organizations;
DROP TRIGGER IF EXISTS update_user_organizations_updated_at ON user_organizations;
DROP TRIGGER IF EXISTS trigger_vendors_updated_at ON vendors;
DROP TRIGGER IF EXISTS trigger_vendor_accounts_updated_at ON vendor_accounts;
DROP TRIGGER IF EXISTS trigger_ensure_single_primary_account ON vendor_accounts;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS ensure_single_primary_account() CASCADE;

-- ================================================
-- DROP EXTENSIONS (optional - uncomment if needed)
-- ================================================

-- DROP EXTENSION IF EXISTS "uuid-ossp";

-- ================================================
-- RESET COMPLETE
-- ================================================

-- Log successful reset
DO $$
BEGIN
    RAISE NOTICE 'NHIT Database reset completed successfully!';
    RAISE NOTICE 'All tables, triggers, and functions have been dropped.';
    RAISE NOTICE 'You can now run the migration script to recreate the schema.';
END $$;
