-- Drop triggers first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_permissions_updated_at ON permissions;
DROP TRIGGER IF EXISTS update_collections_updated_at ON collections;
DROP TRIGGER IF EXISTS update_fields_updated_at ON fields;
DROP TRIGGER IF EXISTS update_settings_updated_at ON settings;
-- Drop the function
DROP FUNCTION IF EXISTS update_updated_at_column();
-- Drop indexes
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_role_id;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_permissions_role_id;
DROP INDEX IF EXISTS idx_permissions_collection;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_expires;
DROP INDEX IF EXISTS idx_activity_user_id;
DROP INDEX IF EXISTS idx_activity_collection;
DROP INDEX IF EXISTS idx_activity_timestamp;
DROP INDEX IF EXISTS idx_revisions_activity_id;
DROP INDEX IF EXISTS idx_revisions_collection_item;
DROP INDEX IF EXISTS idx_fields_collection;
-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS revisions;
DROP TABLE IF EXISTS activity;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS fields;
DROP TABLE IF EXISTS collections;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS settings;