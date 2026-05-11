# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Arsiva** is a backend service for a trivia game with visual novel experience, built for students. It's a REST API built with Go that manages interactive stories, quizzes, puzzles, articles, and user progression.

## Tech Stack

- **Language**: Go 1.25
- **Framework**: Fiber v3 (HTTP framework)
- **Database**: PostgreSQL 15 with PgxPool (connection pooling)
- **Cache**: Redis 7 (for caching)
- **Configuration**: Viper (YAML/JSON config)
- **Validation**: Go Playground Validator
- **Logging**: Logrus
- **Auth**: JWT (golang-jwt/jwt v5)
- **Password**: bcrypt
- **Image Processing**: webp support via chai2010/webp
- **Database Migrations**: golang-migrate
- **Deployment**: Docker + Docker Compose, GitHub Actions CI/CD

## Running the Project

### Prerequisites
- Go 1.25+
- PostgreSQL 15
- Redis 7
- Docker & Docker Compose (for containerized deployment)

### Configuration
1. Copy `config.example.json` to `config.json`
2. Update database credentials:
   ```json
   {
     "database": {
       "postgres": {
         "host": "localhost",
         "port": 5432,
         "user": "your_postgres_user",
         "password": "your_postgres_password",
         "dbname": "your_database_name"
       },
       "redis": {
         "host": "localhost",
         "port": 6379
       }
     }
   }
   ```

### Database Setup

> **Note**: The `Makefile` has hardcoded credentials (`artha:passwordku@localhost:5432/arsiva`) and Linux absolute paths. Update these before using `make` commands, or run `migrate` directly with your own connection string.

```bash
# Run all pending migrations
make migrate

# Rollback all migrations
make migrate-down

# Reset database and re-run all migrations
make migrate-fresh

# Seed database with initial data
make seed

# Rollback seeds
make seed-down

# Create a new migration
migrate create -ext sql -dir db/migrations_postgre create_table_xxx
```

### Starting the Application
```bash
# Run web server directly
go run cmd/web/main.go

# Or using Makefile
make start

# Run with Docker Compose (production-like setup)
docker compose up -d
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests from specific package
go test -v ./internal/usecase/

# Run specific test
go test -v ./internal/usecase -run TestNewUserUseCase
```

The API is accessible at `http://localhost:3000` by default (configured in `config.json` under `web.port`).

## Architecture

Arsiva follows **Clean Architecture** principles with clear separation of concerns:

### Layer Structure

1. **External Layer** (HTTP Requests, gRPC, Messaging)
   - Entry point for all external interactions

2. **Delivery Layer** (`internal/delivery/http/`)
   - HTTP controllers that handle incoming requests
   - Transforms request data into models
   - Calls use cases for business logic
   - Controllers: `auth_controller.go`, `user_controller.go`, `article_controller.go`, etc.
   - Middleware (auth, role-based access control)

3. **Use Case Layer** (`internal/usecase/`)
   - Business logic implementation
   - Orchestrates repositories and gateways
   - Converts entities to response models via converters
   - Examples: `user_usecase.go`, `quiz_usecase.go`, `cerita_usecase.go`

4. **Repository Layer** (`internal/repository/`)
   - Data access abstraction
   - Database operations using PgxPool
   - Handles queries and transactions
   - Each repository corresponds to a domain entity

5. **Entity Layer** (`internal/entity/`)
   - Core business objects representing domain concepts
   - No external dependencies
   - Examples: `user_entity.go`, `quiz.go`, `cerita.go`

6. **Model Layer** (`internal/model/`)
   - Request/response DTOs (Data Transfer Objects)
   - Converter functions to transform between entities and models
   - Validation tags for Go Playground Validator

### Key Architecture Patterns

**Dependency Injection via Bootstrap**
- All dependencies are wired in `internal/config/app.go` Bootstrap function
- Repositories are instantiated with database and logger
- Use cases are instantiated with repositories and dependencies
- Controllers are instantiated with use cases

**Interface-Based Design**
- Controllers, use cases, and repositories are defined as interfaces
- Implementations follow the naming pattern `{Type}Impl`
- Enables testing with mock implementations

**Error Handling**
- Errors bubble up from repository → use case → controller
- Controllers transform errors to appropriate HTTP status codes
- Global error handler in `internal/config/fiber.go`

### Domain Entities

The application manages these core domains:
- **Users** (super_admin, guru, member roles)
- **Content**: Articles, Quizzes, Puzzles, Interactive Stories (Cerita)
- **Categories**: Article categories, Quiz categories, Story categories
- **Groups**: Study groups with members and shared content
- **Schools & Teachers (Guru)**: Educational institution management
- **Achievements & Progress**: User accomplishments and learning progress
- **Assets**: Uploaded files (images, etc.) with automatic cleanup

## Authentication & Authorization

### JWT-Based Auth
- Users login via `/v1/login` endpoint with email + password
- Server returns `access_token` and `refresh_token`
- Tokens are validated using `AuthMiddleware` on protected routes
- JWT secret configured in `config.json` under `app.jwt-secret`

### Role-Based Access Control (RBAC)
Three roles defined in database enum `role_enum`:
- **super_admin**: Full system access, user management
- **guru**: Content creation/management (articles, quizzes, stories, puzzles)
- **member**: Read-only access to content, can join groups

### Route Protection
Routes are protected via middleware chain:
1. `AuthMiddleware` - Validates JWT token, extracts user claims
2. `RoleMiddleware` - Checks user role against allowed roles

Example from `delivery/http/route/route.go`:
```go
superadminOnly := middleware.RoleMiddleware("super_admin")
guruAdmin := middleware.RoleMiddleware("guru", "super_admin")
auth.Get("/users", superadminOnly, c.UserController.GetAllUsers)
auth.Post("/articles", guruAdmin, c.ArticleController.CreateArticle)
```

### Password Security
- Passwords hashed with bcrypt (cost factor 12) via `internal/utils/password.go`
- `HashPassword()` for registration
- `CheckPasswordHash()` for login validation
- Hash stored in database `password_hash` column

## Database Schema

### Key Entities
- **users**: Authentication and roles
- **guru**: Teacher information linked to users
- **sekolah**: Schools/educational institutions
- **member**: Student information linked to users
- **group**: Study groups created by users
- **group_member**: Join table for group membership
- **article**, **artikel**: Content articles with categories
- **quiz**, **pertanyaan_kuis**, **pilihan_kuis**: Quiz structure (questions and options)
- **cerita_interaktif**, **scene**, **puzzle**: Interactive story content
- **assets**: File uploads with metadata
- **achievements**, **member_achievement**: Badge/achievement system
- **member_progres**: User progress tracking
- **member_activity_logs**: Activity audit trail
- **member_social_links**: User social media profiles

### Important Enums
- `role_enum`: super_admin, guru, member
- `status_enum`: draft, published
- `content_type_enum`: kuis, cerita, puzzle
- `activity_type_enum`: login, update_profile, complete_quiz, read_story, solve_puzzle

## Deployment

### Docker Deployment
The project uses Docker Compose for production deployment:
```bash
docker compose build api-server  # Build the image
docker compose up -d              # Start all services (API, PostgreSQL, Redis, Nginx)
```

**Services**:
- **nginx**: Reverse proxy on ports 80/443
- **api-server**: Go backend (built from Dockerfile)
- **postgres-db**: PostgreSQL database
- **redis**: Redis cache
- **certbot**: SSL certificate renewal

### CI/CD Pipeline
GitHub Actions workflow (`.github/workflows/deploy.yml`) automates:
1. **Build**: Docker image compilation
2. **Test**: Unit tests via `go test ./...`
3. **Deploy**: Zero-downtime deployment with `docker compose up -d`

All runs on self-hosted runner.

## Code Organization

```
Arsiva/
├── cmd/
│   ├── web/main.go           # HTTP server entry point
│   └── worker/main.go        # (placeholder for background jobs)
├── delivery/
│   └── http/
│       ├── middleware/       # Auth & role middleware
│       └── route/           # Route configuration
├── internal/
│   ├── config/              # Viper, Fiber, PgxPool, Redis, Validator, Logrus setup
│   ├── delivery/http/       # Controllers (one per domain)
│   ├── entity/              # Core business objects
│   ├── model/               # DTOs and converters
│   ├── repository/          # Data access layer
│   ├── usecase/             # Business logic
│   └── utils/               # Helper functions (password, slug)
├── db/
│   ├── migrations_postgre/  # Schema migrations (numbered by date)
│   └── migrations_seed/     # Data seeding scripts
├── docs/                    # OpenAPI/Swagger documentation
├── Makefile                 # Development commands
├── config.example.json      # Configuration template
├── docker-compose.yml       # Container orchestration
├── Dockerfile              # Multi-stage build for production
└── go.mod/go.sum           # Go module dependencies
```

## Testing Strategy

Tests are unit tests located alongside implementation:
- `*_test.go` files in each package
- Use `testing` package and interfaces for mocking
- Controllers, use cases, and repositories have test files
- Tests verify business logic without external dependencies

Example test pattern:
```go
func TestNewUserUseCase(t *testing.T) {
    uc := NewUserUseCase(nil, nil, nil, validator.New())
    if uc == nil {
        t.Fatal("expected usecase instance")
    }
}
```

## API Documentation

OpenAPI/Swagger documentation is served at `/docs` endpoint and at:
- **Docs folder**: `docs/openapi.yaml`
- **Live URL**: [Swagger UI](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/ArthaFreestyle/Arsiva/main/docs/openapi.yaml)

Routes are organized by resource and role requirements in `delivery/http/route/route.go`:
- Guest routes: `/v1/login`, `/uploads/*`
- Authenticated routes: Protected by `AuthMiddleware`
- Admin routes: Protected by `RoleMiddleware` (super_admin only)
- Content management: Protected by `RoleMiddleware` (guru, super_admin)

## Important Configuration Details

### JWT Secret
Configured in `config.json` under `app.jwt-secret`. Used by auth middleware to validate tokens.

### Asset Management
- Uploaded files stored in `./uploads` directory
- Automatic cleanup of orphaned assets runs every 24 hours via cron in Bootstrap
- `AssetUseCase` handles file operations and cleanup

### CORS Configuration
- Allowed origins configured in `config.json` under `app.allowance`
- Default headers: Origin, Content-Type, Accept, Authorization
- Default methods: GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS

### Database Connection Pool
- Min connections: 3
- Max connections: 10
- Max lifetime: 1 hour
- Max idle time: 10 minutes
- Configured in `internal/config/pgx.go`

### Web Server Settings
- Port: Configured in `config.json` under `web.port` (default 3000)
- Prefork mode: Configurable via `web.prefork` (default false)

## Adding a New API Endpoint

1. **Create entity** in `internal/entity/` if needed
2. **Create repository interface & implementation** in `internal/repository/`
3. **Create use case interface & implementation** in `internal/usecase/`
4. **Create request/response models** in `internal/model/`
5. **Create converter functions** in `internal/model/converter/`
6. **Create controller** in `internal/delivery/http/`
7. **Register route** in `delivery/http/route/route.go`
8. **Wire dependencies** in `internal/config/app.go` Bootstrap function

### Adding Database Schema
1. Create migration: `migrate create -ext sql -dir db/migrations_postgre create_table_xxx`
2. Implement up and down SQL files
3. Run: `make migrate`

## Debugging

- Enable detailed logging via `internal/config/logrus.go` (LogLevel configuration)
- Check logs in running Docker container: `docker compose logs api-server`
- All repositories log their SQL queries

## Git Workflow

The project uses GitHub Actions for automated testing and deployment:
- **Trigger**: Push to `main` branch or manual workflow dispatch
- **Build stage**: Compiles Docker image
- **Test stage**: Runs `go test ./...` in Docker
- **Deploy stage**: Applies migrations and starts services with `docker compose up -d`

Always ensure tests pass locally before pushing:
```bash
go test ./...
```
