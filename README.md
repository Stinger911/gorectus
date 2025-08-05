# GoRectus

A modern, headless CMS built with Go and React, inspired by Directus. GoRectus provides a powerful backend API and a sleek admin interface for managing your data.

## Features

### Backend (Go)

- **RESTful API** built with Gin framework
- **PostgreSQL** database with migration support
- **JWT Authentication** for secure API access
- **Comprehensive Logging** with Logrus
- **Full Test Coverage** with testify and sqlmock
- **Docker Compose** setup for easy development

### Frontend (React)

- **Modern React 18** with TypeScript
- **Material-UI (MUI)** for beautiful, responsive design
- **React Router** for client-side navigation
- **Authentication Context** for state management
- **Admin Dashboard** with comprehensive management tools

### Key Management Features

- **User Management** - Create, edit, and manage users with role-based access
- **Collection Management** - Define dynamic data structures with custom fields
- **Settings Panel** - Configure system settings, database, email, and security
- **Dashboard** - System overview with statistics and recent activity

## Quick Start

### Prerequisites

- Go 1.20+
- PostgreSQL 13+
- Node.js 18+
- Docker & Docker Compose (optional)

### Backend Setup

1. **Start PostgreSQL with Docker Compose:**

   ```bash
   docker-compose up -d
   ```

2. **Set up environment variables:**

   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

3. **Run database migrations:**

   ```bash
   make migrate-up
   ```

4. **Build and run the server:**
   ```bash
   make build
   make run
   ```

### Frontend Setup

1. **Install dependencies:**

   ```bash
   cd frontend
   npm install
   ```

2. **Start the development server:**

   ```bash
   npm start
   ```

3. **Access the admin interface:**
   - Open http://localhost:3000/admin
   - Login with: `admin@gorectus.local` / `admin123`

## API Endpoints

### Authentication

- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `GET /api/auth/me` - Get current user info

### Users

- `GET /api/users` - List all users
- `POST /api/users` - Create new user
- `GET /api/users/:id` - Get user by ID
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user

### Collections

- `GET /api/collections` - List all collections
- `POST /api/collections` - Create new collection
- `GET /api/collections/:id` - Get collection by ID
- `PUT /api/collections/:id` - Update collection
- `DELETE /api/collections/:id` - Delete collection

### Items (Dynamic endpoints based on collections)

- `GET /api/items/:collection` - List items in collection
- `POST /api/items/:collection` - Create new item
- `GET /api/items/:collection/:id` - Get item by ID
- `PUT /api/items/:collection/:id` - Update item
- `DELETE /api/items/:collection/:id` - Delete item

### Dashboard (Admin Only)

- `GET /api/v1/dashboard` - Get complete dashboard overview with all metrics
- `GET /api/v1/dashboard/stats` - Get system statistics (users, roles, collections, sessions)
- `GET /api/v1/dashboard/activity` - Get recent activity log (supports `?limit=N` parameter)
- `GET /api/v1/dashboard/users` - Get user insights and analytics
- `GET /api/v1/dashboard/collections` - Get collection metrics and insights

## Development

### Available Make Commands

```bash
make build          # Build the Go server
make run             # Run the Go server
make test            # Run all tests
make test-coverage   # Run tests with coverage report
make migrate-up      # Run database migrations
make migrate-down    # Rollback migrations
make migrate-reset   # Reset database
make migrate-status  # Check migration status
make validate        # Validate migrations
make docker-up       # Start PostgreSQL and Adminer
make docker-down     # Stop Docker services
make clean           # Clean build artifacts
```

### Project Structure

```
├── cmd/
│   ├── migrate/        # Database migration CLI
│   └── server/         # Main server application
├── frontend/           # React admin interface
│   ├── public/
│   ├── src/
│   │   ├── components/ # Reusable UI components
│   │   ├── contexts/   # React contexts
│   │   ├── pages/      # Main application pages
│   │   ├── services/   # API and utility services
│   │   └── theme/      # MUI theme configuration
│   └── package.json
├── internal/           # Internal Go packages
├── migrations/         # Database migration files
├── scripts/           # Utility scripts
├── docker-compose.yml # Docker services
├── Makefile          # Build and development commands
└── README.md
```

## Testing

The backend includes comprehensive tests covering:

- API endpoints and middleware
- Database operations with mocked connections
- Authentication and authorization flows
- Migration functionality

Run tests with:

```bash
make test
make test-coverage
```

Current test coverage: **59%+**

## Database Schema

### Core Tables

- **users** - System users with authentication
- **roles** - User roles and permissions
- **collections** - Dynamic collection definitions
- **collection_fields** - Field schemas for collections
- **settings** - System configuration

### Sample Data

The initial migration includes:

- Admin user (`admin@gorectus.local` / `admin123`)
- Basic roles (admin, editor, user)
- System settings

## Configuration

### Environment Variables

- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_NAME` - Database name (default: gorectus)
- `DB_USER` - Database user (default: gorectus)
- `DB_PASSWORD` - Database password
- `JWT_SECRET` - JWT signing secret
- `SERVER_PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Logging level (debug, info, warn, error)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by [Directus](https://directus.io/)
- Built with [Gin](https://github.com/gin-gonic/gin) web framework
- UI components from [Material-UI](https://mui.com/)
- Database migrations with [golang-migrate](https://github.com/golang-migrate/migrate)

**gorectus** is an open-source project inspired by [Directus](https://github.com/directus/directus), aiming to provide a modern data platform built with Go for the backend, React for the frontend, and PostgreSQL as the database.

## Features

- RESTful API built with Go
- React-based admin dashboard
- PostgreSQL database support
- Modular and extensible architecture
- User authentication and role-based access control

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (v1.20+)
- [Node.js](https://nodejs.org/) (v16+)
- [PostgreSQL](https://www.postgresql.org/) (v13+)
- [Yarn](https://yarnpkg.com/) or [npm](https://www.npmjs.com/)

### Setup

#### 1. Clone the repository

```bash
git clone https://github.com/yourusername/gorectus.git
cd gorectus
```

#### 2. Configure environment variables

Copy `.env.example` to `.env` and update the values as needed:

```bash
cp .env.example .env
```

#### 3. Start PostgreSQL

Ensure PostgreSQL is running and accessible with the credentials provided in your `.env` file.

#### 4. Run database migrations

```bash
go run cmd/migrate/main.go
```

#### 5. Start the backend server

```bash
go run cmd/server/main.go
```

#### 6. Start the frontend

```bash
cd frontend
yarn install
yarn start
```

## Project Structure

```
gorectus/
├── cmd/            # Main applications (server, migrations)
├── internal/       # Application core packages
├── frontend/       # React frontend
├── migrations/     # Database migrations
├── .env.example    # Example environment variables
└── README.md
```

## Contributing

Contributions are welcome! Please open issues or submit pull requests.

## License

This project is licensed under the MIT License.
