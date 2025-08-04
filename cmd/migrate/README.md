# GoRectus Database Migrations

This document describes the database migration system for GoRectus.

## Overview

GoRectus uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management. The migration system provides:

- ✅ **Version-controlled schema changes**
- ✅ **Rollback capabilities**
- ✅ **Automated initial setup**
- ✅ **Admin user creation**
- ✅ **PostgreSQL support**

## Quick Start

### 1. Start PostgreSQL

```bash
# Using Docker Compose (recommended)
make db-up

# Or manually start PostgreSQL with connection details matching .env
```

### 2. Run Migrations

```bash
# Setup everything (database + migrations)
make setup

# Or run migrations manually
make migrate-up
```

### 3. Access Admin Account

- **Email**: `admin@gorectus.local`
- **Password**: `admin123`
- **⚠️ Change this password after first login!**

## Migration Files

### Current Migrations

1. **`000001_initial_schema.up.sql`** - Core database schema

   - Users, roles, permissions tables
   - Collections and fields (for dynamic schema)
   - Activity logging and revisions
   - Sessions management
   - Settings storage

2. **`000002_seed_data.up.sql`** - Initial data
   - Administrator role with full permissions
   - Public role for unauthenticated users
   - Default admin user (`admin@gorectus.local`)
   - Basic system settings

## Migration Commands

### Using Makefile (Recommended)

```bash
# Start database
make db-up

# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Reset database (drop all + recreate)
make migrate-reset

# Check migration status
make migrate-status

# Force to specific version (use with caution)
make migrate-force VERSION=1
```

### Direct Command Usage

```bash
# Run all pending migrations
go run cmd/migrate/main.go -up

# Rollback N migrations
go run cmd/migrate/main.go -down -steps N

# Run N migrations forward
go run cmd/migrate/main.go -up -steps N

# Force to specific version
go run cmd/migrate/main.go -force V

# Drop everything
go run cmd/migrate/main.go -drop

# Reset (drop + recreate)
go run cmd/migrate/main.go -reset

# Show current status
go run cmd/migrate/main.go

# Generate password hash
go run cmd/migrate/main.go -hash "mypassword"
```

## Database Schema

### Core Tables

#### `users`

Stores user accounts with authentication and profile information.

- UUID primary key
- Email/password authentication
- Role-based access control
- User preferences (theme, language)
- Status management (active, inactive, etc.)

#### `roles`

Defines user roles and their capabilities.

- Admin access flag
- App access permissions
- IP restrictions
- Two-factor authentication enforcement

#### `permissions`

Granular permissions for role-collection-action combinations.

- CRUD operations (create, read, update, delete)
- Field-level access control
- Conditional permissions via JSON

#### `collections`

Dynamic collection definitions (like Directus).

- Collection metadata
- Display configuration
- Sorting and grouping
- Archive/versioning settings

#### `fields`

Dynamic field definitions for collections.

- Field types and interfaces
- Validation rules
- Display options
- Conditional visibility

#### `sessions`

User session management.

- Token-based authentication
- Session expiration
- Device tracking

#### `activity`

Audit log for all system activities.

- User actions tracking
- Change history
- IP and user agent logging

#### `revisions`

Version history for content changes.

- Delta tracking
- Rollback capabilities
- Parent-child relationships

#### `settings`

System-wide configuration.

- Project branding
- Authentication settings
- Default values

## Environment Configuration

Required environment variables:

```bash
# Database connection
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gorectus
DB_SSLMODE=disable
```

## Development Workflow

### 1. Creating New Migrations

```bash
# Create new migration files
migrate create -ext sql -dir migrations -seq description_of_change
```

This creates:

- `NNNNNN_description_of_change.up.sql`
- `NNNNNN_description_of_change.down.sql`

### 2. Writing Migrations

**Up Migration (`*.up.sql`)**:

```sql
-- Add new table
CREATE TABLE example (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add index
CREATE INDEX idx_example_name ON example(name);
```

**Down Migration (`*.down.sql`)**:

```sql
-- Remove in reverse order
DROP INDEX IF EXISTS idx_example_name;
DROP TABLE IF EXISTS example;
```

### 3. Testing Migrations

```bash
# Test up migration
make migrate-up

# Test down migration
make migrate-down

# Test reset
make migrate-reset
```

## Production Deployment

### 1. Backup Database

```bash
pg_dump gorectus > backup_$(date +%Y%m%d_%H%M%S).sql
```

### 2. Run Migrations

```bash
# Check current version
make migrate-status

# Run pending migrations
make migrate-up
```

### 3. Verify Success

```bash
# Check final version
make migrate-status

# Verify data integrity
psql gorectus -c "SELECT COUNT(*) FROM users;"
```

## Troubleshooting

### Common Issues

1. **"dirty" migration state**

   ```bash
   # Force to last good version
   make migrate-force VERSION=N
   ```

2. **Connection refused**

   ```bash
   # Start database
   make db-up

   # Check logs
   make db-logs
   ```

3. **Migration conflicts**
   ```bash
   # Reset and reapply
   make migrate-reset
   ```

### Recovery Procedures

1. **Corrupted migration state**:

   - Identify last good version
   - Force to that version
   - Manually fix database
   - Resume migrations

2. **Failed migration**:
   - Check migration logs
   - Fix SQL errors
   - Force to previous version
   - Retry with corrected migration

## Security Considerations

1. **Default Admin Account**:

   - Change password immediately
   - Consider disabling after creating other admins
   - Use strong passwords

2. **Database Access**:

   - Use environment variables for credentials
   - Enable SSL in production
   - Restrict network access

3. **Migration Safety**:
   - Always backup before migrations
   - Test on staging environment
   - Review SQL before applying

## Schema Evolution

The GoRectus schema is designed to be:

- **Extensible**: Easy to add new collections and fields
- **Versionable**: Full audit trail of changes
- **Flexible**: Support for dynamic content types
- **Compatible**: Similar to Directus for familiarity

Future migrations will maintain backward compatibility while adding new features incrementally.
