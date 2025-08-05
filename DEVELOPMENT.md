# GoRectus Development Guide

A comprehensive development guide for the GoRectus project, covering database migrations, field management, frontend-backend integration, and dashboard implementation.

## Table of Contents

1. [Migration System](#migration-system)
2. [Field Management](#field-management)
3. [Frontend-Backend Integration](#frontend-backend-integration)
4. [Dashboard Implementation](#dashboard-implementation)
5. [Quick Start](#quick-start)
6. [Development Workflow](#development-workflow)

---

## Migration System

### Overview

The GoRectus migration system provides comprehensive database schema management with version control, rollback capabilities, and automation support.

### Files Created

#### Migration Tool (`cmd/migrate/main.go`)

- Complete migration management tool using golang-migrate
- Commands: up, down, reset, force, status, hash generation
- Environment-based configuration
- Comprehensive logging with logrus
- Password hash generation utility

#### SQL Migrations

1. **`000001_initial_schema.up/down.sql`** - Core database schema
2. **`000002_seed_data.up/down.sql`** - Initial admin user and data

#### Configuration & Scripts

- **`docker-compose.yml`** - PostgreSQL development setup
- **`scripts/validate-migrations.sh`** - Migration validation
- **`cmd/migrate/README.md`** - Complete documentation
- **Updated Makefile** - Migration commands integration

### Database Schema

#### Core Tables:

- **`users`** - User accounts with authentication
- **`roles`** - Role-based access control
- **`permissions`** - Granular permission system
- **`collections`** - Dynamic content collections (Directus-style)
- **`fields`** - Dynamic field definitions
- **`sessions`** - User session management
- **`activity`** - Audit logging
- **`revisions`** - Content versioning
- **`settings`** - System configuration

#### Features:

✅ UUID primary keys  
✅ Foreign key constraints  
✅ Indexes for performance  
✅ Automatic updated_at triggers  
✅ Check constraints for data integrity  
✅ JSONB for flexible metadata

### Default Admin User

**Credentials Created:**

- **Email**: `admin@gorectus.local`
- **Password**: `admin123`
- **Role**: Administrator (full permissions)
- **Status**: Active

⚠️ **Security Note**: Change this password after first login!

### Migration Commands

#### Development Setup:

```bash
# Complete setup (database + migrations)
make setup

# Start just the database
make db-up

# Run migrations manually
make migrate-up

# Check migration status
make migrate-status
```

#### Migration Management:

```bash
# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Reset database (drop all + recreate)
make migrate-reset

# Generate password hash
make hash PASSWORD="newpassword"

# Validate migration files
make validate-migrations
```

---

## Field Management

### Overview

The field management system provides full CRUD operations for managing fields within collections, based on Directus patterns.

### Key Features

- ✅ **Dynamic field creation** with database schema management
- ✅ **Field validation** and constraint handling
- ✅ **Multiple field interfaces** (text, number, boolean, date, etc.)
- ✅ **Virtual fields** for UI presentation
- ✅ **Field metadata** management (display options, validation rules)
- ✅ **Type-safe database operations**
- ✅ **Comprehensive test coverage**

### Backend Implementation

#### Files Created

1. **`cmd/server/fields_routes.go`** - Complete field management API
2. **`cmd/server/fields_routes_test.go`** - Comprehensive test suite
3. **Updated `cmd/server/main.go`** - Added fields handler registration

#### API Endpoints

All endpoints require authentication. Admin role required for create/update/delete operations.

```http
GET    /api/v1/fields                     # List all fields (with pagination)
GET    /api/v1/fields?collection=name     # List fields for specific collection
GET    /api/v1/fields/:collection         # Get all fields for a collection
GET    /api/v1/fields/:collection/:field  # Get specific field
POST   /api/v1/fields/:collection         # Create new field
PATCH  /api/v1/fields/:collection/:field  # Update field
DELETE /api/v1/fields/:collection/:field  # Delete field
```

#### Field Types and Interfaces

The system provides 20+ field interfaces:

##### Text Inputs

- **Text Input** - Simple text field
- **WYSIWYG** - Rich text editor
- **Markdown** - Markdown editor
- **Textarea** - Multi-line text
- **Code** - Code editor with syntax highlighting

##### Numbers

- **Number** - Numeric input
- **Slider** - Numeric slider

##### Boolean

- **Boolean** - Toggle switch

##### Date & Time

- **Date & Time** - Full date/time picker
- **Date** - Date only picker
- **Time** - Time only picker

##### Selection

- **Dropdown** - Selection dropdown
- **Radio Buttons** - Radio button group
- **Checkboxes** - Multiple selection
- **Many-to-One** - Relational dropdown

##### Files & Media

- **Image** - Image upload
- **File** - File upload
- **Files** - Multiple file upload

##### Special

- **Tags** - Tag input
- **Color** - Color picker
- **UUID** - UUID input
- **Key-Value Pairs** - JSON object editor

##### Presentation (Virtual)

- **Divider** - Visual separator
- **Notice** - Information display

### Frontend Implementation

#### Files Created

1. **Updated `frontend/src/services/collectionsService.ts`** - Added field management methods
2. **`frontend/src/services/fieldsService.ts`** - Comprehensive field utilities and constants

#### Usage Example

```typescript
// Create a product collection with fields
const collection = await collectionsService.createCollection({
  collection: "products",
  icon: "inventory_2",
  note: "Product catalog",
  fields: [],
});

// Add title field
await FieldsService.createField("products", {
  field: "title",
  interface: "input",
  display: "raw",
  required: true,
  width: "full",
  schema: {
    data_type: "varchar",
    max_length: 255,
    is_nullable: false,
  },
});

// Add price field
await FieldsService.createField("products", {
  field: "price",
  interface: "input-number",
  display: "formatted-value",
  required: true,
  width: "half",
  validation: {
    min: 0,
  },
  schema: {
    data_type: "decimal",
    is_nullable: false,
    default_value: 0,
  },
});
```

---

## Frontend-Backend Integration

### Overview

The frontend has been updated to properly integrate with the GoRectus backend API, providing complete CRUD operations for users, roles, and other entities.

### Changes Made

#### New Services Created

##### `/src/services/usersService.ts`

- Comprehensive Users API service
- Handles CRUD operations for users
- Supports pagination with proper response types
- Includes proper error handling

##### `/src/services/rolesService.ts`

- Roles API service
- Handles fetching roles for user management
- Supports pagination

#### Updated User Interface

The user model has been enhanced to match the backend structure:

- Changed from simple `role` string to `role_id` and `role_name` fields
- Added comprehensive user fields:
  - `avatar`, `language`, `theme`, `status`
  - `last_access`, `provider`, `external_identifier`
  - `email_notifications`, `tags`
  - `created_at`, `updated_at`

#### Component Updates

##### `/src/pages/Users.tsx`

- Complete rewrite to use actual backend API
- Added pagination support
- Enhanced form with all user fields
- Proper error handling and loading states
- Role selection from actual roles endpoint
- Support for user creation and updates with proper validation

##### `/src/contexts/AuthContext.tsx`

- Updated to use `role_name` instead of `role` for admin checks

##### `/src/services/authService.ts`

- Updated User interface to match backend structure

### Backend API Compatibility

The frontend integrates with these backend endpoints:

#### User Endpoints

- `GET /api/v1/users` - List users with pagination
- `POST /api/v1/users` - Create new user
- `GET /api/v1/users/:id` - Get single user
- `PATCH /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

#### Role Endpoints

- `GET /api/v1/roles` - List roles with pagination

#### Authentication

- All API calls include proper Authorization headers
- Token refresh handling for expired tokens

### Key Features

1. **Pagination**: Users list supports server-side pagination
2. **Role Management**: Dynamic role selection from backend
3. **Comprehensive User Fields**: Support for all user attributes
4. **Error Handling**: Proper error display and handling
5. **Loading States**: User feedback during API operations
6. **Form Validation**: Client-side validation before API calls

---

## Dashboard Implementation

### Overview

The frontend Dashboard component has been updated to conform with the actual backend API implementation, providing real-time system insights.

### Changes Made

#### Created Dashboard Service (`src/services/dashboardService.ts`)

- **Complete API Interface**: Interfaces for all dashboard data structures
- **Comprehensive Service Methods**:
  - `getDashboardOverview()` - Complete dashboard data
  - `getSystemStats()` - Basic system statistics
  - `getRecentActivity(limit?)` - Recent activity with optional limit
  - `getUserInsights()` - User analytics and insights
  - `getCollectionInsights()` - Collection metrics
- **Error Handling**: Proper error handling with typed responses

#### Updated Main Dashboard Component (`src/pages/Dashboard.tsx`)

- **Real Data Integration**: Replaced hardcoded data with actual API calls
- **Admin Access Control**: Added proper admin role checking
- **Loading States**: Comprehensive loading, error, and empty states
- **Tabbed Interface**: Three main sections:
  - System Overview
  - User Analytics
  - Collection Insights
- **Enhanced Statistics**: Stats cards now show real data from the backend
- **Activity Feed**: Real-time activity display with proper formatting

#### Created Specialized Components

##### User Insights Component (`src/components/Dashboard/UserInsights.tsx`)

- **User Distribution**: Visual breakdown by status and role
- **Growth Metrics**: New users this week/month
- **Recent Registrations**: List of newest users with details
- **Color-coded Status**: Visual indicators for user roles and statuses

##### Collection Insights Component (`src/components/Dashboard/CollectionInsights.tsx`)

- **Collection Types**: Distribution by content type
- **Activity Metrics**: Most active collections
- **Recent Collections**: Latest created collections
- **Collection Details**: Item counts, singleton/hidden status, creation dates

### API Endpoints Consumed

The frontend consumes these backend endpoints:

- `GET /api/v1/dashboard` - Complete dashboard overview
- `GET /api/v1/dashboard/stats` - Basic system statistics
- `GET /api/v1/dashboard/activity` - Recent activity feed
- `GET /api/v1/dashboard/users` - User analytics
- `GET /api/v1/dashboard/collections` - Collection metrics

### Features

#### Admin Access Control

- Only users with "Administrator" role can access dashboard
- Proper error messages for unauthorized access
- Seamless integration with existing auth system

#### Real-time Data

- Data fetched from actual backend APIs
- Proper error handling and loading states
- Automatic retry and refresh capabilities

#### Responsive Design

- Mobile-friendly grid layouts
- Adaptive card sizing
- Optimized for different screen sizes

#### Enhanced User Experience

- Skeleton loading states during data fetch
- Comprehensive error messages
- Intuitive tabbed navigation
- Color-coded status indicators
- Formatted timestamps and dates

---

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+
- Docker & Docker Compose

### Initial Setup

1. **Clone and Setup Database:**

```bash
git clone <repository>
cd GoRectus

# Start PostgreSQL and run migrations
make setup
```

2. **Start Backend Server:**

```bash
# Build and run the server
make build
make run

# Or run directly
go run cmd/server/main.go
```

3. **Start Frontend Development Server:**

```bash
cd frontend
npm install
npm start
```

4. **Access the Application:**

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Admin Login: admin@gorectus.local / admin123

### Environment Configuration

Create `.env` file in the root directory:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=gorectus
DB_USER=gorectus
DB_PASSWORD=gorectus123
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your-jwt-secret-key

# Server
PORT=8080
```

---

## Development Workflow

### Backend Development

1. **Database Changes:**

```bash
# Create new migration
migrate create -ext sql -dir migrations -seq add_new_feature

# Edit migration files
# Run migrations
make migrate-up

# Validate
make validate-migrations
```

2. **API Development:**

```bash
# Run tests
go test -v ./cmd/server

# Build
make build

# Run with hot reload (using air)
air
```

3. **Field Management:**

```bash
# Test field operations
go test -v -run TestFieldsHandlerSuite ./cmd/server
```

### Frontend Development

1. **Service Development:**

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm start

# Run tests
npm test
```

2. **API Integration:**

- Services are in `src/services/`
- Use existing patterns for new API endpoints
- Include proper TypeScript interfaces
- Add error handling and loading states

3. **Component Development:**

- Follow existing component structure
- Use Material-UI components
- Include proper prop types and interfaces
- Add responsive design considerations

### Testing

#### Backend Tests

```bash
# Run all tests
go test -v ./...

# Run specific test suite
go test -v ./cmd/server

# Run with coverage
go test -cover ./...
```

#### Frontend Tests

```bash
cd frontend

# Run tests
npm test

# Run tests with coverage
npm test -- --coverage
```

### Production Deployment

1. **Backend:**

```bash
# Build for production
make build

# Run migrations
make migrate-up

# Start server
./bin/gorectus-server
```

2. **Frontend:**

```bash
cd frontend

# Build for production
npm run build

# Serve static files with your web server
```

### Database Management

#### Backup and Restore

```bash
# Backup
pg_dump gorectus > backup.sql

# Restore
psql gorectus < backup.sql
```

#### Migration Management

```bash
# Check status
make migrate-status

# Apply migrations
make migrate-up

# Rollback
make migrate-down

# Reset (dangerous - drops all data)
make migrate-reset
```

### Security Considerations

1. **Default Credentials**: Change admin password immediately after setup
2. **Environment Variables**: Use secure storage for production secrets
3. **Database Access**: Enable SSL in production
4. **Migration Safety**: Always backup before running migrations
5. **API Security**: Implement rate limiting and input validation
6. **Frontend Security**: Sanitize user inputs and implement CSP

### Troubleshooting

#### Common Issues

1. **Database Connection Errors:**

   - Check PostgreSQL is running
   - Verify connection parameters in `.env`
   - Ensure database exists

2. **Migration Failures:**

   - Check migration syntax
   - Verify database permissions
   - Review migration logs

3. **API Integration Issues:**

   - Check backend server is running
   - Verify API endpoints
   - Check authentication headers

4. **Field Management Issues:**
   - Ensure field names are valid
   - Check interface compatibility
   - Verify admin permissions

#### Useful Commands

```bash
# Check database status
make db-status

# View logs
docker-compose logs postgres

# Reset development environment
make clean && make setup

# Generate password hash
make hash PASSWORD="newpassword"
```

---

## Conclusion

The GoRectus development environment provides a comprehensive, production-ready foundation with:

- ✅ **Enterprise-grade database management** with migrations
- ✅ **Dynamic field management** system
- ✅ **Complete frontend-backend integration**
- ✅ **Real-time dashboard** with system insights
- ✅ **Role-based access control**
- ✅ **Comprehensive testing**
- ✅ **Developer-friendly tooling**

The system is ready for feature development and can scale to support complex content management requirements while maintaining data integrity and performance.
