-- Add additional settings columns for application configuration
ALTER TABLE settings
ADD COLUMN IF NOT EXISTS maintenance_mode BOOLEAN DEFAULT false;
ALTER TABLE settings
ADD COLUMN IF NOT EXISTS smtp_host VARCHAR(255);
ALTER TABLE settings
ADD COLUMN IF NOT EXISTS smtp_port VARCHAR(10) DEFAULT '587';
ALTER TABLE settings
ADD COLUMN IF NOT EXISTS smtp_user VARCHAR(255);
ALTER TABLE settings
ADD COLUMN IF NOT EXISTS smtp_from_email VARCHAR(255);
ALTER TABLE settings
ADD COLUMN IF NOT EXISTS email_enabled BOOLEAN DEFAULT false;
ALTER TABLE settings
ADD COLUMN IF NOT EXISTS session_timeout INTEGER DEFAULT 24;
ALTER TABLE settings
ADD COLUMN IF NOT EXISTS password_min_length INTEGER DEFAULT 8;
ALTER TABLE settings
ADD COLUMN IF NOT EXISTS require_two_factor BOOLEAN DEFAULT false;
-- Create trigger for settings table (if not exists)
DO $$ BEGIN IF NOT EXISTS (
    SELECT 1
    FROM pg_trigger
    WHERE tgname = 'update_settings_updated_at'
) THEN CREATE TRIGGER update_settings_updated_at BEFORE
UPDATE ON settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
END IF;
END $$;
-- Insert default settings if none exist
INSERT INTO settings (
        project_name,
        project_descriptor,
        public_registration,
        maintenance_mode,
        email_enabled,
        session_timeout,
        password_min_length,
        require_two_factor
    )
SELECT 'GoRectus',
    'A modern headless CMS built with Go',
    false,
    false,
    false,
    24,
    8,
    false
WHERE NOT EXISTS (
        SELECT 1
        FROM settings
        LIMIT 1
    );