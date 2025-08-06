-- Remove additional settings columns
ALTER TABLE settings DROP COLUMN IF EXISTS maintenance_mode;
ALTER TABLE settings DROP COLUMN IF EXISTS smtp_host;
ALTER TABLE settings DROP COLUMN IF EXISTS smtp_port;
ALTER TABLE settings DROP COLUMN IF EXISTS smtp_user;
ALTER TABLE settings DROP COLUMN IF EXISTS smtp_from_email;
ALTER TABLE settings DROP COLUMN IF EXISTS email_enabled;
ALTER TABLE settings DROP COLUMN IF EXISTS session_timeout;
ALTER TABLE settings DROP COLUMN IF EXISTS password_min_length;
ALTER TABLE settings DROP COLUMN IF EXISTS require_two_factor;
-- Drop trigger
DROP TRIGGER IF EXISTS update_settings_updated_at ON settings;