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
- Login also accepts an `expected_role` field; the server rejects mismatched roles (prevents a member token being issued to someone using the guru login form, etc.)
- Server returns `access_token` and `refresh_token`
- Tokens are validated using `AuthMiddleware` on protected routes
- JWT secret configured in `config.json` under `app.jwt-secret`

### Email Verification (OTP) & Password Reset
Both flows are OTP-based (6-digit numeric codes emailed to the user) and share one Redis-backed helper in `internal/usecase/auth_usecase.go`:
- **Register → verify (Approach B):** `register` creates the user with `is_verified=false` (Postgres column, added in migration `20260715000001`) and mails a verification OTP. A mail failure does **not** fail registration — the account exists and the user can re-request a code. `Login` rejects unverified accounts with **403**, and that check runs **after** the password check so it cannot be used to enumerate accounts.
- **Forgot → reset (link-based, NOT OTP):** `POST /v1/forgot-password` mails a **reset link** (not a code) pointing at the frontend page `email.reset_password_url` (default `https://arsiva.id/reset-password`) with `?token=…&email=…` appended (`issueResetToken` → `buildResetURL`). It **always returns a generic success** regardless of whether the email exists (anti-enumeration — all internal errors are logged and swallowed). `POST /v1/reset-password` takes `{email, token, new_password}`, validates the token via `consumeOTP` (same SHA-256 compare — a reset token is stored exactly like an OTP), and calls `UserRepository.UpdatePassword` (a dedicated single-column update — never `UpdateUser`, which would clobber username/email/role). **Email verification on register is still OTP** (`issueOTP`); only password reset uses a link.
- **Secret storage:** Redis keys `otp:verify:{email}` (6-digit OTP) / `otp:reset:{email}` (32-byte URL-safe token from `utils.GenerateResetToken`) store the **SHA-256 hash** of the secret (never plaintext) plus an `attempts` counter, with a TTL. Both share the `storeSecret` helper. A resend cooldown key `otp:cooldown:{purpose}:{email}` throttles re-issues; because `storeSecret` arms it **before** the mail reaches the relay, `issueOTP`/`issueResetToken` roll it back via `clearResendCooldown` when `SendHTML` fails — otherwise one SMTP hiccup would leave the user with no mail *and* a 429 on retry for the whole window. The rollback drops only the cooldown, never the stored secret (a code that did land stays usable). Secrets are single-use (deleted on success) and invalidated after `otp_max_attempts` wrong tries. **The verify OTP and the reset link have separate TTLs** (`storeSecret` picks the TTL by purpose via `ttlFor`): `otp_ttl_minutes` (default **5**) is the verification-OTP lifetime, `reset_link_ttl_minutes` (default **15**) is the reset-link lifetime — longer because the link round-trips through email. All tunables live under the `email` block in `config.json` (`otp_ttl_minutes`, `reset_link_ttl_minutes`, `otp_max_attempts`, `otp_resend_cooldown_seconds`, `reset_password_url`).
- **FE timing hints:** `register`, `resend-otp`, and `forgot-password` responses carry the (static, non-user-specific) timing knobs so the FE can render a countdown + throttle its resend button — `AuthUseCase.OTPPolicy()` supplies them. Register/resend return `otp_expires_in_seconds` + `resend_cooldown_seconds`; forgot-password returns `reset_link_expires_in_seconds` + `resend_cooldown_seconds`. These are config-derived constants, so emitting them leaks nothing about whether an email is registered (anti-enumeration stays intact).
- **Endpoints** (all guest routes, rate-limited via `AuthLimiter`): `POST /v1/verify-email`, `POST /v1/resend-otp`, `POST /v1/forgot-password`, `POST /v1/reset-password`.
- **Mailer:** `internal/mailer/mailer.go` uses stdlib `net/smtp` (zero new deps). It is tolerant of the local relay setup (`host=localhost`, `port=25`): STARTTLS only if advertised, AUTH only if advertised and a username is set. Every message carries `Date` **and `Message-ID`** — Gmail rejects mail without a Message-ID at end-of-DATA (`550-5.7.1 ... RfcMessageNonCompliant`), and Postfix only backfills missing headers when `always_add_missing_headers=yes` (off by default), so the mailer emits them itself; `mailer_test.go` guards this. The relay's `250` means *accepted for delivery*, not delivered — a later bounce is invisible to the app, so diagnose delivery from the host's `/var/log/mail.log`, not `logs/arsiva.log`. Connect and conversation are bounded by `smtpDialTimeout`/`smtpConversationTimeout` (stdlib `smtp.Dial` has none, so a wedged relay would otherwise pin the HTTP request). Because `host=localhost`, email only sends when the app runs **on the VPS** — it cannot be exercised from a local dev machine. Email HTML/text bodies live in `internal/mailer/templates.go` (OTP, reset-link, and group-invite templates share one visual language).

### Group Invitations by Email (link-based, NOT auto-add)
`POST /v1/groups/:id/invite` (`GroupUseCase.InviteMembersByEmail`, guru-owner only) emails a **join link** to each address with an optional guru `message`. It **does not add anyone directly** — recipients join by clicking the link (→ login/register → `POST /v1/groups/join`), so it works for existing members and unregistered students alike. Details:
- The link points at `email.group_invite_url` (default `https://arsiva.id/join-group`) with `?token=…` appended. The token is the **existing group-scoped `group_invite` JWT** (7-day expiry) minted by `buildGroupInviteToken` — the same helper `GenerateInviteLink` (the shareable/QR link) now uses, and the same token `JoinGroup` consumes.
- The email body is rendered **once** and reused for every recipient (the token is group-scoped, not per-recipient). Duplicate addresses are de-duplicated. The response is a per-email confirmation `{total, sent, failed, results[]}` so the FE can show which invites failed; a mail failure for one address does not abort the batch.
- `GroupUseCase` gained `Mailer` + `InviteBaseURL` deps (wired in `internal/config/app.go`). The old behavior — silently `AddMember`-ing registered emails and ignoring the rest — was **removed**.

### Profile Completion Gate
After auth, most action endpoints additionally pass through `ProfileCompleteMiddleware` so half-onboarded users (account exists but `guru`/`member` profile row not yet filled in) cannot reach action endpoints. Routes that intentionally skip this check:
- `POST /v1/guru`, `POST /v1/member` — profile creation itself
- `GET /v1/guru/me`, `GET /v1/member/me` — let the FE detect "no profile yet"

The middleware is wired in `internal/config/app.go` and applied per-route in `delivery/http/route/route.go`.

### Role-Based Access Control (RBAC)
Three roles defined in database enum `role_enum`:
- **super_admin**: Full system access, user management
- **guru**: Content creation/management (articles, quizzes, stories, puzzles)
- **member**: Read-only access to content, can join groups

### Route Protection
Routes are protected via middleware chain:
1. `AuthMiddleware` — validates JWT token, extracts user claims (applied at the `/v1` group level)
2. `RoleMiddleware` — checks user role against allowed roles (per-route)
3. `ProfileCompleteMiddleware` — gates action endpoints on profile completion (per-route, see above)

Example from `delivery/http/route/route.go`:
```go
superadminOnly := middleware.RoleMiddleware("super_admin")
guruAdminRole  := middleware.RoleMiddleware("guru", "super_admin")
auth.Get("/users", superadminOnly, c.UserController.GetAllUsers)
auth.Post("/articles", guruAdminRole, c.ProfileCompleteMiddleware, c.ArticleController.CreateArticle)
```

### Fiber v3 middleware gotcha (IMPORTANT)
Fiber v3's `Group(prefix, ...handlers)` registers the handlers via `app.Use(prefix, ...)` under the hood. **Empty-prefix sub-groups leak middleware to every route under the parent prefix** — e.g. `auth.Group("", roleMW)` causes `roleMW` to run on every `/v1` route, not just routes registered on that sub-group. With multiple role-based sub-groups this means every role check runs on every request and the strictest one wins, so every endpoint returns 403.

To compose per-route middleware, **pass it inline** on the route call:
```go
auth.Get(path, roleMW, profileCompleteMW, handler)
```

Also: route methods (`Get`/`Post`/etc.) take individual `fiber.Handler` args, not a `[]fiber.Handler`. Defining middleware as a slice and passing the slice as one argument panics at startup with `group: invalid handler #0 ([]func(fiber.Ctx) error)`. Build each role middleware once at the top of `SetupAuthRoutes` and reuse the variable across calls.

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

## Gameplay & Progression Flow

This is the most cross-cutting subsystem and spans Redis, Postgres, and the gamification domain. Reading a single file does not reveal it.

### Server-authoritative scoring (member play = ephemeral Redis sessions)
Members *play* quizzes/stories/puzzles through `/v1/progress/*` (`ProgressController` → `ProgressSessionUseCase`, `internal/usecase/progress_session_usecase.go`). The browser is **never** the source of truth for score:
- `POST /v1/progress/start` creates a Redis session (`progress:session:{memberId}:{contentType}:{contentId}`) and registers it in a sorted set `progress:active` scored by expiry time.
- `POST /v1/progress/answer` (quiz), `/scene` (story), `/solve` (puzzle) advance the session. The server looks up the authoritative score from the DB (e.g. `GetSceneEndingInfo`, option `score`) and accumulates it **into Redis**, not from the client.
- `POST /v1/progress/submit` (or session expiry) calls `Finalize`, which moves the Redis session into `member_progres`, credits `total_xp`, and updates level/gamification. `Finalize` is **idempotent** — a second call (e.g. submit racing the expiry sweeper) is a no-op.
- `GET /v1/progress/session/:content_type/:content_id` reads the in-flight session.

### Answer-key protection: member views use separate `Public*` DTOs (IMPORTANT)
Because scoring is server-side, member-facing **get-by-id** endpoints must not serialize scoring/answer metadata. The pattern (see issues #41 quiz, #42 story):
- `GET /v1/quizzes/:id` returns `PublicQuizResponse` (omits option `score` — the answer key — and question `poin`).
- `GET /v1/stories/:id` returns `PublicCeritaResponse` (omits scene `ending_point`, `ending_type`, `urutan`).
- The shared `QuizResponse`/`SceneResponse` and their converters are **left untouched** — the guru/admin `manage`/create/update endpoints legitimately need those fields to author content.
- **Never strip fields from the shared response struct.** Add a parallel `Public*Response` + `ToPublic*` converter, and switch only the member get-by-id usecase/controller to it. The repo query may keep selecting the sensitive columns (the entity carries them for server-side scoring); just ensure no member serializer emits them.
- List endpoints (`GET /v1/quizzes`, `/v1/stories`) don't embed `Soal`/`Scenes` (`omitempty` + the list query skips them), so the leak only ever affects get-by-id.

### Gamification (`GamificationUseCase`)
`Finalize` → `HandleContentFinished` drives streaks, daily tasks, XP totals, and level-ups (`internal/utils/level.go`). XP is awarded **once per content per member** (dedup on `member_progres`); crossing a level threshold auto-levels-up.

## Background Workers (run in-process, NOT cmd/worker)

`cmd/worker/main.go` is still an empty placeholder. The real background jobs run as goroutines started in `Bootstrap` (`internal/config/app.go`) and are **guarded by `if !fiber.IsChild()`** so prefork mode runs them only on the master process (otherwise every prefork child would double-run them):
- `startAssetCleanupCron` — deletes orphaned uploads every 24h (`AssetUseCase.CleanupOrphanedAssets`).
- `startProgressFlushWorker` (`internal/config/progress_worker.go`) — every 15min, finalizes expired play sessions via `ListExpiredSessionKeys` → `Finalize(key, "expired")`; one bad session is logged and skipped, never aborting the batch.

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

Notable non-obvious entries:
- `cmd/worker/main.go` — placeholder entrypoint for future background jobs; not built or run by anything today.
- `testing/script.js` — load-test script (k6-style), not Go tests. Unit tests are still `go test ./...`.
- `architecture.png` — Clean Architecture diagram referenced from README.

```
Arsiva/
├── cmd/
│   ├── web/main.go           # HTTP server entry point
│   └── worker/main.go        # placeholder, no jobs wired yet
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

The OpenAPI spec lives at `docs/openapi.yaml`. The server mounts that folder as static files at `/docs/*` (`app.Get("/docs/*", static.New("./docs"))` in `internal/config/fiber.go`) — it serves the raw YAML, not a Swagger UI. Use the petstore.swagger.io link in `README.md` to view it rendered.

Routes are organized by resource and role requirements in `delivery/http/route/route.go`:
- Guest routes: `POST /v1/login`, `POST /v1/register/member`, `POST /v1/register/guru`, `GET /uploads/*`
- Authenticated routes: protected by `AuthMiddleware` (applied at `/v1` group level)
- Admin routes: additionally guarded by `RoleMiddleware("super_admin")`
- Content management: additionally guarded by `RoleMiddleware("guru", "super_admin")` + `ProfileCompleteMiddleware`

## Important Configuration Details

### JWT Secret
Configured in `config.json` under `app.jwt-secret`. Used by auth middleware to validate tokens.

### Asset Management
- Uploaded files stored in `./uploads` directory
- Automatic cleanup of orphaned assets runs every 24 hours via an in-process goroutine started in Bootstrap (guarded by `!fiber.IsChild()` — see "Background Workers" above)
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

Follow the existing layer layout (entity → repository → usecase → model/converter → controller → route → wire in `internal/config/app.go` Bootstrap). Mirror a nearby domain (e.g. `sekolah` or `achievement`) for boilerplate — every layer pairs an interface with a `*Impl` implementation.

### Adding Database Schema
1. `migrate create -ext sql -dir db/migrations_postgre create_table_xxx`
2. Fill in up/down SQL
3. `make migrate`

## Debugging

- Enable detailed logging via `internal/config/logrus.go` (LogLevel configuration)
- Logs are written to **both stderr and a rotating file on disk**. The file path comes from `log.file` in `config.json` (default `./logs/arsiva.log`); rotation (`max_size_mb`, `max_backups`, `max_age_days`, `compress`) is configured under the same `log` block — see `config.example.json`. Rotation is handled by `gopkg.in/natefinch/lumberjack.v2`.
- Check logs in running Docker container: `docker compose logs api-server` (stderr), or read the file under `/app/logs` (bind-mounted to a persistent host dir in `docker-compose.yml` so it survives container recreation).
- The `logs/` directory is gitignored.
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
