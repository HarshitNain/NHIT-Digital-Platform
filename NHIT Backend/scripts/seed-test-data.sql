-- ================================================
-- NHIT Backend - Test Data Seeding Script
-- ================================================
-- This script inserts sample data for testing purposes

-- ================================================
-- ORGANIZATIONS
-- ================================================

-- Insert test organizations
INSERT INTO organizations (org_id, tenant_id, name, code, database_name, description, is_active, created_by) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'NHIT Technologies', 'NHIT', 'nhit_db', 'Main NHIT organization for technology services', true, '550e8400-e29b-41d4-a716-446655440002'),
    ('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440000', 'NHIT Consulting', 'NHITC', 'nhit_consulting_db', 'NHIT consulting division', true, '550e8400-e29b-41d4-a716-446655440002')
ON CONFLICT (org_id) DO NOTHING;

-- ================================================
-- DEPARTMENTS
-- ================================================

-- Insert test departments
INSERT INTO departments (id, name, description) VALUES
    ('650e8400-e29b-41d4-a716-446655440001', 'Engineering', 'Software development and engineering team'),
    ('650e8400-e29b-41d4-a716-446655440002', 'Human Resources', 'HR and people operations'),
    ('650e8400-e29b-41d4-a716-446655440003', 'Finance', 'Financial planning and accounting'),
    ('650e8400-e29b-41d4-a716-446655440004', 'Marketing', 'Marketing and business development'),
    ('650e8400-e29b-41d4-a716-446655440005', 'Operations', 'Business operations and support')
ON CONFLICT (name) DO NOTHING;

-- ================================================
-- DESIGNATIONS
-- ================================================

-- Insert test designations with hierarchy
INSERT INTO designations (id, name, description) VALUES
    ('750e8400-e29b-41d4-a716-446655440001', 'Chief Executive Officer', 'CEO - Top executive position'),
    ('750e8400-e29b-41d4-a716-446655440002', 'Chief Technology Officer', 'CTO - Technology leadership'),
    ('750e8400-e29b-41d4-a716-446655440003', 'Engineering Manager', 'Manager for engineering teams')
ON CONFLICT (slug) DO NOTHING;

-- ================================================
-- USERS
-- ================================================

-- Insert test users (password is 'password123' hashed with bcrypt)
INSERT INTO users (user_id, tenant_id, name, email, password, email_verified_at, department_id, designation_id) VALUES
    ('850e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'John Doe', 'john.doe@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), '650e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440002'),
    ('850e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440000', 'Jane Smith', 'jane.smith@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), '650e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440003'),
    ('850e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440000', 'Mike Johnson', 'mike.johnson@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), '650e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440004'),
    ('850e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440000', 'Sarah Wilson', 'sarah.wilson@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), '650e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440005'),
    ('850e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440000', 'David Brown', 'david.brown@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), '650e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440006'),
    ('850e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440000', 'Lisa Davis', 'lisa.davis@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), '650e8400-e29b-41d4-a716-446655440002', '750e8400-e29b-41d4-a716-446655440007'),
    ('850e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440000', 'Robert Miller', 'robert.miller@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), '650e8400-e29b-41d4-a716-446655440003', '750e8400-e29b-41d4-a716-446655440008'),
    ('850e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440000', 'Emily Taylor', 'emily.taylor@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), '650e8400-e29b-41d4-a716-446655440004', '750e8400-e29b-41d4-a716-446655440009'),
    ('850e8400-e29b-41d4-a716-446655440009', '550e8400-e29b-41d4-a716-446655440000', 'Chris Anderson', 'chris.anderson@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), '650e8400-e29b-41d4-a716-446655440005', '750e8400-e29b-41d4-a716-446655440010'),
    ('850e8400-e29b-41d4-a716-446655440010', '550e8400-e29b-41d4-a716-446655440000', 'Admin User', 'admin@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), '650e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440001')
ON CONFLICT (email) DO NOTHING;

-- ================================================
-- USER ORGANIZATIONS
-- ================================================

-- Assign users to organizations
INSERT INTO user_organizations (user_id, org_id, role_id, is_current_context) VALUES
    ('850e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440002', true),
    ('850e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440003', true),
    ('850e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440004', true),
    ('850e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440005', true),
    ('850e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440006', true),
    ('850e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440007', true),
    ('850e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440008', true),
    ('850e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440009', true),
    ('850e8400-e29b-41d4-a716-446655440009', '550e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440010', true),
    ('850e8400-e29b-41d4-a716-446655440010', '550e8400-e29b-41d4-a716-446655440001', '750e8400-e29b-41d4-a716-446655440001', true)
ON CONFLICT (user_id, org_id) DO NOTHING;

-- ================================================
-- PROJECTS
-- ================================================

-- Insert test projects
-- Insert test projects
INSERT INTO projects (id, tenant_id, project_name, org_id, created_by, created_at, updated_at) VALUES
    ('950e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'NHIT Backend API', '550e8400-e29b-41d4-a716-446655440001', 'mohit', NOW(), NOW()),
    ('950e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440000', 'NHIT Frontend Portal', '550e8400-e29b-41d4-a716-446655440001', 'aman', NOW(), NOW()),
    ('950e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440000', 'Mobile Application', '550e8400-e29b-41d4-a716-446655440003', 'om', NOW(), NOW()),
    ('950e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440000', 'Data Analytics Platform', '550e8400-e29b-41d4-a716-446655440003', 'shristi', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ================================================
-- VENDORS
-- ================================================

-- Insert test vendors
INSERT INTO vendors (id, tenant_id, vendor_code, vendor_name, vendor_email, vendor_mobile, vendor_type, pan, beneficiary_name, is_active, created_by) VALUES
    ('a50e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'VEN001', 'Tech Solutions Pvt Ltd', 'contact@techsolutions.com', '+91-9876543210', 'Service Provider', 'ABCDE1234F', 'Tech Solutions Pvt Ltd', true, '850e8400-e29b-41d4-a716-446655440001'),
    ('a50e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440000', 'VEN002', 'Cloud Services Inc', 'info@cloudservices.com', '+91-9876543211', 'Technology', 'FGHIJ5678K', 'Cloud Services Inc', true, '850e8400-e29b-41d4-a716-446655440001'),
    ('a50e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440000', 'VEN003', 'Office Supplies Co', 'sales@officesupplies.com', '+91-9876543212', 'Supplier', 'LMNOP9012Q', 'Office Supplies Co', true, '850e8400-e29b-41d4-a716-446655440001')
ON CONFLICT (tenant_id, vendor_code) DO NOTHING;

-- ================================================
-- VENDOR ACCOUNTS
-- ================================================

-- Insert vendor banking accounts
INSERT INTO vendor_accounts (id, vendor_id, account_name, account_number, account_type, name_of_bank, branch_name, ifsc_code, is_primary, is_active, created_by) VALUES
    ('b50e8400-e29b-41d4-a716-446655440001', 'a50e8400-e29b-41d4-a716-446655440001', 'Tech Solutions Pvt Ltd', '1234567890123456', 'Current', 'State Bank of India', 'Mumbai Main Branch', 'SBIN0000001', true, true, '850e8400-e29b-41d4-a716-446655440001'),
    ('b50e8400-e29b-41d4-a716-446655440002', 'a50e8400-e29b-41d4-a716-446655440002', 'Cloud Services Inc', '2345678901234567', 'Current', 'HDFC Bank', 'Bangalore Branch', 'HDFC0000002', true, true, '850e8400-e29b-41d4-a716-446655440001'),
    ('b50e8400-e29b-41d4-a716-446655440003', 'a50e8400-e29b-41d4-a716-446655440003', 'Office Supplies Co', '3456789012345678', 'Savings', 'ICICI Bank', 'Delhi Branch', 'ICIC0000003', true, true, '850e8400-e29b-41d4-a716-446655440001')
ON CONFLICT (id) DO NOTHING;

-- ================================================
-- UPDATE DESIGNATION USER COUNTS
-- ================================================

-- Update user counts for designations
UPDATE designations SET user_count = (
    SELECT COUNT(*) FROM users WHERE designation_id = designations.id
);

-- ================================================
-- SEED DATA COMPLETE
-- ================================================

-- Log successful seeding
DO $$
BEGIN
    RAISE NOTICE 'NHIT Test data seeding completed successfully!';
    RAISE NOTICE 'Seeded data includes:';
    RAISE NOTICE '- 2 Organizations';
    RAISE NOTICE '- 5 Departments';
    RAISE NOTICE '- 10 Designations (with hierarchy)';
    RAISE NOTICE '- 10 Users (password: password123)';
    RAISE NOTICE '- 4 Projects';
    RAISE NOTICE '- 3 Vendors with banking accounts';
    RAISE NOTICE 'You can now test the NHIT backend services with this sample data!';
END $$;
