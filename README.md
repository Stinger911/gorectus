# gorectus

Attempt to create copy of [Directus](https://github.com/directus/directus) with Go

## Overview

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
