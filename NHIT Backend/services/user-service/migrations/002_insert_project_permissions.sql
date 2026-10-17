-- Insert project permissions if they don't exist
INSERT INTO permissions (name, description, module, action, is_system_permission)
VALUES
    ('create-projects', 'Create projects', 'projects', 'create', TRUE),
    ('edit-projects', 'Edit projects', 'projects', 'edit', TRUE),
    ('delete-projects', 'Delete projects', 'projects', 'delete', TRUE),
    ('view-projects', 'View projects', 'projects', 'view', TRUE),
    ('approve-projects', 'Approve projects', 'projects', 'approve', TRUE)
ON CONFLICT (name) DO NOTHING;
