-- Seeding Projects for requested organizations

-- 1. RNR Consulting Private Limited (379d8ece-e140-41a9-ad3b-577ea64c2b27)
-- Projects: "RNR Project 1", "RNR Project 2", "RNR Project 3", "RNR Project 4", "RNR Project 5"
INSERT INTO projects (id, tenant_id, org_id, project_name, created_by, created_at, updated_at) VALUES 
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', '379d8ece-e140-41a9-ad3b-577ea64c2b27', 'RNR Project 1', 'RNR Consulting Private Limited', NOW(), NOW()),
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', '379d8ece-e140-41a9-ad3b-577ea64c2b27', 'RNR Project 2', 'RNR Consulting Private Limited', NOW(), NOW()),
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', '379d8ece-e140-41a9-ad3b-577ea64c2b27', 'RNR Project 3', 'RNR Consulting Private Limited', NOW(), NOW()),
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', '379d8ece-e140-41a9-ad3b-577ea64c2b27', 'RNR Project 4', 'RNR Consulting Private Limited', NOW(), NOW()),
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', '379d8ece-e140-41a9-ad3b-577ea64c2b27', 'RNR Project 5', 'RNR Consulting Private Limited', NOW(), NOW());

-- 2. Aaditya Tech (381e68d4-b4e2-4a1f-9d58-2f2a5020bd16)
-- Projects: "ad1"
INSERT INTO projects (id, tenant_id, org_id, project_name, created_by, created_at, updated_at) VALUES 
(gen_random_uuid(), '8d6764bb-b862-45e7-bb7e-acb668b536a7', '381e68d4-b4e2-4a1f-9d58-2f2a5020bd16', 'ad1', 'Abhishek Kumar Sah', NOW(), NOW());

-- 3. Oasis Infobyte (3e3d65fa-54ae-42ce-acb0-5a4400a57e12)
-- Projects: "Oasis Project 1", "Oasis Project 2", "Oasis Project 3"
INSERT INTO projects (id, tenant_id, org_id, project_name, created_by, created_at, updated_at) VALUES 
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', '3e3d65fa-54ae-42ce-acb0-5a4400a57e12', 'Oasis Project 1', 'RNR Consulting Private Limited', NOW(), NOW()),
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', '3e3d65fa-54ae-42ce-acb0-5a4400a57e12', 'Oasis Project 2', 'RNR Consulting Private Limited', NOW(), NOW()),
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', '3e3d65fa-54ae-42ce-acb0-5a4400a57e12', 'Oasis Project 3', 'RNR Consulting Private Limited', NOW(), NOW());

-- 4. Ab Technologies (6b9f950e-1bd5-457f-b75a-f2ac652b7dbc)
-- Projects: "AB TECH PROJECT 1", "AB TECH PROJECT 2", "AB TECH PROJECT 3"
INSERT INTO projects (id, tenant_id, org_id, project_name, created_by, created_at, updated_at) VALUES 
(gen_random_uuid(), '8d6764bb-b862-45e7-bb7e-acb668b536a7', '6b9f950e-1bd5-457f-b75a-f2ac652b7dbc', 'AB TECH PROJECT 1', 'Abhishek Kumar Sah', NOW(), NOW()),
(gen_random_uuid(), '8d6764bb-b862-45e7-bb7e-acb668b536a7', '6b9f950e-1bd5-457f-b75a-f2ac652b7dbc', 'AB TECH PROJECT 2', 'Abhishek Kumar Sah', NOW(), NOW()),
(gen_random_uuid(), '8d6764bb-b862-45e7-bb7e-acb668b536a7', '6b9f950e-1bd5-457f-b75a-f2ac652b7dbc', 'AB TECH PROJECT 3', 'Abhishek Kumar Sah', NOW(), NOW());

-- 5. PixPivot Private Limited (fab2a88d-5db3-4300-9bd8-8965742e829d)
-- Projects: "NHIT", "ERP", "Consent Manager"
INSERT INTO projects (id, tenant_id, org_id, project_name, created_by, created_at, updated_at) VALUES 
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', 'fab2a88d-5db3-4300-9bd8-8965742e829d', 'NHIT', 'RNR Consulting Private Limited', NOW(), NOW()),
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', 'fab2a88d-5db3-4300-9bd8-8965742e829d', 'ERP', 'RNR Consulting Private Limited', NOW(), NOW()),
(gen_random_uuid(), 'cf58f3f5-676e-4de3-99e3-a47009efb631', 'fab2a88d-5db3-4300-9bd8-8965742e829d', 'Consent Manager', 'RNR Consulting Private Limited', NOW(), NOW());
