-- Remove seed data
-- Delete permissions
DELETE FROM permissions
WHERE role_id IN (
        '550e8400-e29b-41d4-a716-446655440000',
        '550e8400-e29b-41d4-a716-446655440001'
    );
-- Delete settings
DELETE FROM settings
WHERE id = '550e8400-e29b-41d4-a716-446655440020';
-- Delete admin user
DELETE FROM users
WHERE email = 'admin@gorectus.local';
-- Delete roles
DELETE FROM roles
WHERE name IN ('Administrator', 'Public');