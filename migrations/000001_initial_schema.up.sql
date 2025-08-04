-- Initial migration for GoRectus
-- Creates core tables for the system
-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar VARCHAR(255),
    language VARCHAR(10) DEFAULT 'en-US',
    theme VARCHAR(20) DEFAULT 'auto',
    status VARCHAR(20) DEFAULT 'active' CHECK (
        status IN (
            'active',
            'inactive',
            'invited',
            'draft',
            'suspended'
        )
    ),
    role_id UUID,
    last_access TIMESTAMP,
    last_page VARCHAR(255),
    provider VARCHAR(50) DEFAULT 'default',
    external_identifier VARCHAR(255),
    auth_data JSONB,
    email_notifications BOOLEAN DEFAULT true,
    tags JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    icon VARCHAR(50) DEFAULT 'supervised_user_circle',
    description TEXT,
    ip_access TEXT [],
    enforce_tfa BOOLEAN DEFAULT false,
    admin_access BOOLEAN DEFAULT false,
    app_access BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    collection VARCHAR(64) NOT NULL,
    action VARCHAR(10) NOT NULL CHECK (
        action IN (
            'create',
            'read',
            'update',
            'delete',
            'comment',
            'explain'
        )
    ),
    permissions JSONB,
    validation JSONB,
    presets JSONB,
    fields TEXT [],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Create collections table (for dynamic collections)
CREATE TABLE IF NOT EXISTS collections (
    collection VARCHAR(64) PRIMARY KEY,
    icon VARCHAR(50),
    note TEXT,
    display_template VARCHAR(255),
    hidden BOOLEAN DEFAULT false,
    singleton BOOLEAN DEFAULT false,
    translations JSONB,
    archive_field VARCHAR(64),
    archive_app_filter BOOLEAN DEFAULT true,
    archive_value VARCHAR(255),
    unarchive_value VARCHAR(255),
    sort_field VARCHAR(64),
    accountability VARCHAR(10) DEFAULT 'all' CHECK (accountability IN ('all', 'activity')),
    color VARCHAR(10),
    item_duplication_fields JSONB,
    sort INTEGER,
    "group" VARCHAR(64),
    collapse VARCHAR(10) DEFAULT 'open' CHECK (collapse IN ('open', 'closed', 'locked')),
    preview_url VARCHAR(255),
    versioning BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Create fields table (for dynamic field definitions)
CREATE TABLE IF NOT EXISTS fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection VARCHAR(64) REFERENCES collections(collection) ON DELETE CASCADE,
    field VARCHAR(64) NOT NULL,
    special VARCHAR(64) [],
    interface VARCHAR(64),
    options JSONB,
    display VARCHAR(64),
    display_options JSONB,
    readonly BOOLEAN DEFAULT false,
    hidden BOOLEAN DEFAULT false,
    sort INTEGER,
    width VARCHAR(10) DEFAULT 'full' CHECK (
        width IN (
            'half',
            'half-left',
            'half-right',
            'full',
            'fill'
        )
    ),
    translations JSONB,
    note TEXT,
    conditions JSONB,
    required BOOLEAN DEFAULT false,
    "group" VARCHAR(64),
    validation JSONB,
    validation_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(collection, field)
);
-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    token VARCHAR(64) PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    expires TIMESTAMP NOT NULL,
    ip VARCHAR(45),
    user_agent TEXT,
    data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Create activity table (audit log)
CREATE TABLE IF NOT EXISTS activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action VARCHAR(45) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE
    SET NULL,
        timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        ip VARCHAR(45),
        user_agent TEXT,
        collection VARCHAR(64),
        item VARCHAR(255),
        comment TEXT,
        origin VARCHAR(255),
        revisions UUID [],
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Create revisions table (for versioning)
CREATE TABLE IF NOT EXISTS revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID REFERENCES activity(id) ON DELETE CASCADE,
    collection VARCHAR(64) NOT NULL,
    item VARCHAR(255) NOT NULL,
    data JSONB,
    delta JSONB,
    parent UUID REFERENCES revisions(id),
    version INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Create settings table
CREATE TABLE IF NOT EXISTS settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_name VARCHAR(100) DEFAULT 'GoRectus',
    project_descriptor VARCHAR(255),
    project_url VARCHAR(255),
    project_color VARCHAR(10) DEFAULT '#6644FF',
    project_logo UUID,
    public_foreground UUID,
    public_background UUID,
    public_note TEXT,
    auth_login_attempts INTEGER DEFAULT 25,
    auth_password_policy TEXT,
    storage_asset_transform VARCHAR(10) DEFAULT 'all',
    storage_asset_presets JSONB,
    custom_css TEXT,
    storage_default_folder UUID,
    basemaps JSONB,
    mapbox_key VARCHAR(255),
    module_bar JSONB,
    project_dashboard_note TEXT,
    default_language VARCHAR(10) DEFAULT 'en-US',
    custom_aspect_ratios JSONB,
    public_registration BOOLEAN DEFAULT false,
    default_appearance VARCHAR(10) DEFAULT 'auto',
    default_theme_light VARCHAR(50),
    default_theme_dark VARCHAR(50),
    report_error_url VARCHAR(255),
    report_bug_url VARCHAR(255),
    report_feature_url VARCHAR(255),
    public_registration_verify_email BOOLEAN DEFAULT true,
    public_registration_role UUID,
    public_registration_email_filter JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- Add foreign key constraint for users.role_id
ALTER TABLE users
ADD CONSTRAINT fk_users_role_id FOREIGN KEY (role_id) REFERENCES roles(id);
-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_permissions_role_id ON permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_permissions_collection ON permissions(collection);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires);
CREATE INDEX IF NOT EXISTS idx_activity_user_id ON activity(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_collection ON activity(collection);
CREATE INDEX IF NOT EXISTS idx_activity_timestamp ON activity(timestamp);
CREATE INDEX IF NOT EXISTS idx_revisions_activity_id ON revisions(activity_id);
CREATE INDEX IF NOT EXISTS idx_revisions_collection_item ON revisions(collection, item);
CREATE INDEX IF NOT EXISTS idx_fields_collection ON fields(collection);
-- Create a function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ language 'plpgsql';
-- Create triggers to automatically update updated_at
CREATE TRIGGER update_users_updated_at BEFORE
UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE
UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_permissions_updated_at BEFORE
UPDATE ON permissions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_collections_updated_at BEFORE
UPDATE ON collections FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_fields_updated_at BEFORE
UPDATE ON fields FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_settings_updated_at BEFORE
UPDATE ON settings FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();