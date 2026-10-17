-- Check total GreenNotes
SELECT COUNT(*) as total_greennotes FROM green_notes;

-- Check all GreenNotes with their key fields
SELECT id, org_id, tenant_id, project_name, supplier_name, status, created_at 
FROM green_notes 
ORDER BY created_at DESC 
LIMIT 10;

-- Check distinct org_id values
SELECT DISTINCT org_id, COUNT(*) as count 
FROM green_notes 
GROUP BY org_id;

-- Check the logged-in user's details
SELECT id, org_id, tenant_id, email, name 
FROM users 
WHERE email = 'nhit@gmail.com';
