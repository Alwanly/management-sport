# Management Sport — Sports Management API

This repository implements a Sports Management API system written in Go. It follows a clean architecture pattern and Domain-Driven Design (DDD) principles to provide a maintainable, testable REST API for managing teams, players, matches, goals, reports and audit logs.

## Features

- ✅ Clean Architecture & Domain-Driven Design (DDD)
- ✅ RESTful API with Gin framework
- ✅ PostgreSQL database with GORM ORM
- ✅ Redis caching support
- ✅ JWT & Basic Authentication
- ✅ Structured logging with Zap
- ✅ Configuration management with Viper
- ✅ Auto-generated Swagger documentation
- ✅ Unit testing with Testify and Mockery
- ✅ Health check endpoints
- ✅ Graceful shutdown
- ✅ Docker support

## Libraries Used

This project uses several libraries to facilitate development and ensure code quality:

- **Gin**: HTTP web framework used for routing and handling requests
- **GORM**: Feature-rich ORM library for database interactions with PostgreSQL
- **Viper**: Configuration management for environment variables and config files
- **Zap**: High-performance structured logger
- **go-redis/redis**: Redis client for caching
- **golang-jwt/jwt**: JWT handling for authentication
- **Swag**: Swagger documentation generator
- **Testify & Mockery**: Testing and mock generation tools

## Project Structure

The project follows a clean architecture structure:

```
go-codebase/
│
├── cmd/main/              # Application entry point
│   ├── bootstrap.go       # Application initialization and dependency injection
│   └── main.go            # Main function and server setup
│
├── pkg/                   # Reusable library code (can be imported by external projects)
│   ├── authentication/    # Authentication utilities (JWT, Basic Auth, Password)
│   ├── binding/           # Request binding helpers
│   ├── database/          # Database connection and utilities
│   ├── deps/              # Dependency injection container
│   ├── health/            # Health check handlers
│   ├── logger/            # Logging utilities
│   ├── middleware/        # HTTP middlewares
│   ├── redis/             # Redis client setup
│   ├── utils/             # Common utilities
│   ├── validator/         # Request validation
│   └── wrapper/           # Response wrapper utilities
│
├── internal/              # Private application code (domain logic)
│   ├── auth/              # Authentication domain (handlers, usecase, repository)
│   ├── team/              # Teams domain (handlers, usecase, repository)
│   ├── player/            # Players domain
│   ├── match/             # Matches domain
│   ├── goal/              # Goals domain
│   ├── report/            # Reports and statistics
│   └── audit/             # Audit logging
│
├── model/                 # Database models
│   ├── team.go            # Team model
│   ├── player.go          # Player model
│   ├── match.go           # Match model
│   ├── goal.go            # Goal model
│   └── user.go            # User model
│
├── config/                # Configuration management
│   ├── config.go          # Config loader
│   ├── default.go         # Default values
│   └── type.go            # Config type definitions
│
├── api/                   # API documentation (auto-generated)
│   ├── docs.go            # Swagger documentation code
│   ├── swagger.json       # Swagger JSON spec
│   └── swagger.yaml       # Swagger YAML spec
│
├── db/                    # Database migrations
│   └── migration/         # SQL migration files
│
├── Dockerfile             # Production Docker image
├── Makefile              # Build and development commands
├── go.mod                # Go module dependencies
└── .env.example          # Environment variables template
```

## Getting Started

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 14 or higher
- Redis 7 or higher (optional)
- Make (optional, for using Makefile commands)

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/Alwanly/management-sport.git
   cd go-codebase
   ```

2. **Install dependencies**

   ```bash
   make install
   # or manually:
   go mod download
   ```

3. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Configure database**

   Update the following in your `.env` file (example):
   ```env
   POSTGRES_URI=postgres://user:password@localhost:5432/yourdb
   REDIS_URI=redis://localhost:6379
   JWT_SECRET=your_jwt_secret
   UPLOAD_DIRECTORY=./images
   MAX_UPLOAD_SIZE=5242880 # 5MB
   ALLOWED_IMAGE_TYPES=png,jpg,jpeg
   ```

5. **Generate RSA keys for JWT (optional)**

   ```bash
   # Generate private key
   openssl genrsa -out private.pem 2048
   
   # Generate public key
   openssl rsa -in private.pem -pubout -out public.pem
   
   # Copy keys to .env as single-line base64 or load them in code
   ```

### Running the Application

**Development mode with hot reload:**

```bash
make dev
```

**Standard mode:**

```bash
make run
```

**Build binary:**

```bash
make build
./app
```

**Using Docker:**

```bash
docker build -t go-codebase .
docker run -p 9000:9000 --env-file .env go-codebase
```

### API Documentation

Once the application is running in development mode, access the Swagger documentation at:

```
http://localhost:9000/swagger/index.html
```

### Health Checks

The application provides three health check endpoints:

- `GET /health` - General health check
- `GET /ready` - Readiness probe (for Kubernetes)
- `GET /live` - Liveness probe (for Kubernetes)

## Development

### Generating API Documentation

```bash
make docs
```

This uses `swag` to scan your code and generate Swagger documentation in the `./api` directory.

### Running Tests

```bash
make test
```

### Generating Test Coverage

```bash
make coverage
```

Coverage report will be generated in `coverage/coverage.html`.

### Generating Mocks

```bash
make mock
```

### Code Formatting

```bash
make format
```

### Linting

```bash
make lint
```

## API Endpoints

### Authentication & API Overview

Authentication flow:

- `POST /auth/v1/register` - Register a new user (initial setup may require BasicAuth depending on deployment)
- `POST /auth/v1/login` - Login to receive a JWT token

Use the received token in requests to protected endpoints with header:

```
Authorization: Bearer <token>
```

Primary API domains:

- **Authentication:** `/auth/v1/*` - Register, login (user/admin)
- **Teams:** `/teams/v1/*` - CRUD operations with logo upload (admin for create/update/delete)
- **Players:** `/players/v1/*` - CRUD operations and team assignment
- **Matches:** `/matches/v1/*` - Schedule matches, update status (scheduled → ongoing → finished)
- **Goals:** `/goals/v1/*` - Record goals linked to a player and match
- **Reports:** `/reports/v1/*` - Statistics (goals per player/team, match reports)
- **Audits:** `/audits/v1/*` - Administrative audit logs (admin-only)

## Environment Variables

Critical environment variables (see `.env.example` for a fuller list):

| Variable | Description | Default |
|----------|-------------|---------|
| `ENV` | Environment (development/production) | development |
| `PORT` | HTTP server port | 9000 |
| `POSTGRES_URI` | PostgreSQL connection string | Required |
| `REDIS_URI` | Redis connection string | Optional |
| `JWT_SECRET` | Secret key for signing JWT tokens (HMAC mode) | Required |
| `UPLOAD_DIRECTORY` | Directory path for uploaded files (team logos) | ./images |
| `MAX_UPLOAD_SIZE` | Maximum upload size in bytes | 5242880 (5MB) |
| `ALLOWED_IMAGE_TYPES` | Comma-separated allowed image extensions | png,jpg,jpeg |

## License

This project is licensed under the MIT License.
