-- ================================================
-- NHIT Backend - Fresh Seed Data
-- ================================================
-- Clean seed data for fresh database setup

-- Insert minimal test data for fresh start
INSERT INTO organizations (org_id, tenant_id, name, code, database_name, description, logo, is_active, created_by, created_at, updated_at)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'NHIT Technologies', 'NHIT', 'nhit_db', 'Main NHIT organization', NULL, true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW());

-- Insert test user with bcrypt hashed password "password123"
INSERT INTO users (user_id, tenant_id, name, email, password, email_verified_at, department_id, designation_id, created_at, updated_at)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'NHIT Admin', 'admin@nhit.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NOW(), NULL, NULL, NOW(), NOW());

-- Link user to organization
INSERT INTO user_organizations (user_id, org_id, role_id, is_current_context, joined_at, updated_at)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', true, NOW(), NOW());

-- Insert sample department
INSERT INTO departments (id, org_id, name, description, created_at, updated_at)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'Information Technology', 'IT Department', NOW(), NOW());

-- Insert sample designation
INSERT INTO designations (id, org_id, name, description, created_at, updated_at)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 'Software Developer', 'Full Stack Developer', NOW(), NOW());

COMMIT;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'Fresh seed data inserted successfully!';
    RAISE NOTICE 'Test Organization: NHIT Technologies (Code: NHIT)';
    RAISE NOTICE 'Test User: admin@nhit.com (Password: password123)';
    RAISE NOTICE 'Database is ready for fresh testing!';
END $$;
