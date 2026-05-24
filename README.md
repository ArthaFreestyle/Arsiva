# Arsiva

> Backend service for **Arsiva** — a trivia game with a visual novel experience, built for students.

Arsiva is a REST API written in Go that manages interactive stories, quizzes, puzzles, articles, study groups, and member progression — plus gamification (XP, levels, leaderboards, daily streaks, and daily tasks).

## Architecture

![Clean Architecture](architecture.png)

This project follows **Clean Architecture** principles, with each domain wired through the same layered flow:

1. External system performs a request (HTTP, gRPC, Messaging, etc).
2. The **Delivery** layer (`internal/delivery/http`) binds request data into a **Model**.
3. The Delivery calls a **Use Case** with that Model.
4. The **Use Case** (`internal/usecase`) builds **Entity** data and runs the business logic.
5. The Use Case calls a **Repository** with Entity data.
6. The **Repository** (`internal/repository`) performs the database operation.
7. The Use Case converts Entities back into response Models for the Delivery layer.

Every repository, use case, and controller is defined as an `interface` paired with a `{Type}Impl` implementation, and all dependencies are wired in `internal/config/app.go` (`Bootstrap`).

## Features

- **Auth & RBAC** — JWT access/refresh tokens, bcrypt password hashing, and three roles: `super_admin`, `guru` (teacher), `member` (student). A profile-completion gate blocks half-onboarded users from action endpoints.
- **Content** — articles, quizzes, interactive stories (cerita), and puzzles, each with categories and draft/published status.
- **Groups** — study groups created by teachers, with member management, invite links, and shared content.
- **Schools & Teachers** — educational institution management.
- **Game progress** — server-authoritative scoring sessions (Redis-backed) that finalize into `member_progress` and award XP.
- **Gamification** — XP/levels, public & group leaderboards, **daily streaks** (with freeze shields), and **daily tasks** (member-only).
- **Assets** — uploaded images (with webp support) and a 24h cron that cleans up orphaned files.

## Tech Stack

- **Go 1.25** — https://github.com/golang/go
- **PostgreSQL 15** (database) — https://github.com/postgres/postgres
- **Redis 7** (cache & progress sessions) — https://github.com/redis/redis

## Framework & Library

- **GoFiber v3** (HTTP framework) — https://github.com/gofiber/fiber
- **PgxPool** (database connection pooling) — https://github.com/jackc/pgx
- **go-redis** (Redis client) — https://github.com/redis/go-redis
- **golang-jwt v5** (JWT auth) — https://github.com/golang-jwt/jwt
- **Viper** (configuration) — https://github.com/spf13/viper
- **golang-migrate** (database migrations) — https://github.com/golang-migrate/migrate
- **Go Playground Validator** (validation) — https://github.com/go-playground/validator
- **Logrus** (logging) — https://github.com/sirupsen/logrus

## Configuration

Copy the template and fill in your own credentials:

```shell
cp config.example.json config.json
```

Key settings live under `database.postgres`, `database.redis`, `app.jwt-secret`, `app.allowance` (CORS origins), and `web.port` (default `3000`).

## Database Migrations

Migrations live in `db/migrations_postgre`; data seeds live in `db/migrations_seed`.

> **Note:** the `Makefile` has hardcoded credentials (`artha:passwordku@localhost:5432/arsiva`) and Linux absolute paths. Update these before using `make`, or run `migrate` directly with your own connection string.

### Create a migration

```shell
migrate create -ext sql -dir db/migrations_postgre create_table_xxx
```

### Run via Makefile

| Command | Description |
|---|---|
| `make migrate` | Apply all pending migrations |
| `make migrate-down` | Rollback all migrations |
| `make migrate-fresh` | Reset the schema & re-run all migrations |
| `make seed` | Apply data seeds |
| `make seed-down` | Rollback data seeds |

## Run Application

### Run the web server

```bash
go run cmd/web/main.go
# or
make start
```

The API is served at `http://localhost:3000` by default.

### Run tests

```bash
go test ./...
# verbose
go test -v ./...
```

### Run with Docker Compose

```bash
docker compose up -d
```

This brings up the API server, PostgreSQL, Redis, Nginx (reverse proxy), and Certbot.

## API Documentation

The OpenAPI spec lives at `docs/openapi.yaml` and is served as a static file at `/docs/*`.

[![Swagger](https://img.shields.io/badge/Swagger-Dokumentasi%20API%20Arsiva-85EA2D?logo=swagger&logoColor=black)](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/ArthaFreestyle/Arsiva/main/docs/openapi.yaml)

Silakan lihat dokumentasi API kami di sini:
[👉 Buka Dokumentasi API Arsiva (Swagger UI)](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/ArthaFreestyle/Arsiva/main/docs/openapi.yaml)
