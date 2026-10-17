-- Create test user with proper bcrypt hash for password123
INSERT INTO users (
    user_id, 
    tenant_id, 
    name, 
    email, 
    password, 
    email_verified_at, 
    created_at, 
    updated_at
) VALUES (
    '34998b4b-7b48-42ab-90f9-67a72002995b',
    '123e4567-e89b-12d3-a456-426614174000',
    'Super Admin',
    'admin@mycompany.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    NOW(),
    NOW(),
    NOW()
);

-- Link user to organization
INSERT INTO user_organizations (
    user_id, 
    org_id, 
    role_id, 
    is_current_context, 
    joined_at, 
    updated_at
) VALUES (
    '34998b4b-7b48-42ab-90f9-67a72002995b',
    '34998b4b-7b48-42ab-90f9-67a72002995b',
    '34998b4b-7b48-42ab-90f9-67a72002995b',
    true,
    NOW(),
    NOW()
) ON CONFLICT (user_id, org_id) DO NOTHING;

COMMIT;
