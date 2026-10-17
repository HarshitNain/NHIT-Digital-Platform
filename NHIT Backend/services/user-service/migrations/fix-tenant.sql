-- Fix Tenant Foreign Key Constraint Issue
-- Run this SQL to create the tenant first

-- Connect to database:
-- psql -U postgres -d nhit

-- Create tenant if it doesn't exist
INSERT INTO tenants (
    tenant_id,
    tenant_name,
    tenant_code,
    is_active,
    created_at,
    updated_at
)
VALUES (
    '123e4567-e89b-12d3-a456-426614174000',
    'Test Tenant',
    'TEST001',
    true,
    NOW(),
    NOW()
)
ON CONFLICT (tenant_id) DO NOTHING;

-- Verify tenant was created
SELECT 
    tenant_id,
    tenant_name,
    tenant_code,
    is_active,
    created_at
FROM tenants
WHERE tenant_id = '123e4567-e89b-12d3-a456-426614174000';

-- Now you can register users with this tenant_id!
