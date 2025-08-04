-- Seed data migration for GoRectus
-- Creates initial admin role, admin user, and basic settings
-- Insert the default admin role
INSERT INTO roles (
        id,
        name,
        icon,
        description,
        admin_access,
        app_access,
        enforce_tfa
    )
VALUES (
        '550e8400-e29b-41d4-a716-446655440000',
        'Administrator',
        'verified_user',
        'System administrators with full access to all features and settings',
        true,
        true,
        false
    ) ON CONFLICT (name) DO NOTHING;
-- Insert a public role for unauthenticated users
INSERT INTO roles (
        id,
        name,
        icon,
        description,
        admin_access,
        app_access,
        enforce_tfa
    )
VALUES (
        '550e8400-e29b-41d4-a716-446655440001',
        'Public',
        'public',
        'Public access role for unauthenticated users',
        false,
        false,
        false
    ) ON CONFLICT (name) DO NOTHING;
-- Insert default admin user
-- Password: 'admin123' hashed with bcrypt (cost 10)
-- You should change this password after first login
INSERT INTO users (
        id,
        email,
        password,
        first_name,
        last_name,
        role_id,
        status,
        language,
        theme,
        email_notifications
    )
VALUES (
        '550e8400-e29b-41d4-a716-446655440010',
        'admin@gorectus.local',
        '$2a$10$h0uOsJ2KCG1YKAnzfPe2UuNZpVKYEXKZhYiDrB1jXNTUBteYZu0Ku',
        'System',
        'Administrator',
        '550e8400-e29b-41d4-a716-446655440000',
        'active',
        'en-US',
        'auto',
        true
    ) ON CONFLICT (email) DO NOTHING;
-- Insert default settings
INSERT INTO settings (
        id,
        project_name,
        project_descriptor,
        project_color,
        default_language,
        auth_login_attempts,
        public_registration,
        public_registration_verify_email,
        storage_asset_transform,
        default_appearance
    )
VALUES (
        '550e8400-e29b-41d4-a716-446655440020',
        'GoRectus',
        'A modern data platform built with Go',
        '#6644FF',
        'en-US',
        25,
        false,
        true,
        'all',
        'auto'
    ) ON CONFLICT (id) DO NOTHING;
-- Insert admin permissions for all core collections
INSERT INTO permissions (role_id, collection, action, permissions, fields)
VALUES (
        '550e8400-e29b-41d4-a716-446655440000',
        'users',
        'create',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'users',
        'read',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'users',
        'update',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'users',
        'delete',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'roles',
        'create',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'roles',
        'read',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'roles',
        'update',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'roles',
        'delete',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'permissions',
        'create',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'permissions',
        'read',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'permissions',
        'update',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'permissions',
        'delete',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'collections',
        'create',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'collections',
        'read',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'collections',
        'update',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'collections',
        'delete',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'fields',
        'create',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'fields',
        'read',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'fields',
        'update',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'fields',
        'delete',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'settings',
        'create',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'settings',
        'read',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'settings',
        'update',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'settings',
        'delete',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'activity',
        'read',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'revisions',
        'read',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'sessions',
        'read',
        '{}',
        '{"*"}'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'sessions',
        'delete',
        '{}',
        '{"*"}'
    ) ON CONFLICT DO NOTHING;
-- Public role permissions (very limited)
INSERT INTO permissions (role_id, collection, action, permissions, fields)
VALUES (
        '550e8400-e29b-41d4-a716-446655440001',
        'settings',
        'read',
        '{"project_name": {"_eq": true}, "project_descriptor": {"_eq": true}, "project_color": {"_eq": true}}',
        '{"project_name", "project_descriptor", "project_color"}'
    ) ON CONFLICT DO NOTHING;