# 🏅 Feature: CRUD `member_achievements` — member self-unlocks badges from front-end

## 📋 Summary

Implement **CRUD for the `member_achievements` join table** that records which badges each member has unlocked. The front-end (member client) calls `POST /v1/member/achievements` whenever a member wins / clears something — the server validates the request, enforces ownership against the JWT claims, and persists the unlock.

Mirrors the layered pattern of **Member Social Links CRUD (#21)** (member-only, claims-driven, no cross-member access) and reuses the achievement catalog landed in **Achievement CRUD (#22)**.

> ⚠️ Read this section first — it constrains every design decision below.
>
> **A `member_achievements` row has only one writeable column besides its composite PK: `unlocked_at`.** There is no meaningful "Update" semantic — once a badge is unlocked, the row is final. So the "U" in CRUD is **intentionally omitted** for this resource. Do not invent a `PUT` endpoint; do not add one to "complete the CRUD" out of habit.

---

## 🗄️ Database Context

### Table: `member_achievements`
Source: `db/migrations_postgre/20260318203943_create_table_member_achievement.up.sql`

```sql
CREATE TABLE member_achievements (
    member_id INTEGER NOT NULL,
    achievement_id INTEGER NOT NULL,
    unlocked_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT member_achievements_pkey PRIMARY KEY (member_id, achievement_id)
);
```

### Foreign keys (existing — do NOT touch)
Source: `db/migrations_postgre/20260318204005_create_foreign_key.up.sql:65-67`

```sql
ALTER TABLE member_achievements
  ADD CONSTRAINT member_achievements_member_id_fkey
  FOREIGN KEY (member_id) REFERENCES members(member_id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE member_achievements
  ADD CONSTRAINT member_achievements_achievement_id_fkey
  FOREIGN KEY (achievement_id) REFERENCES achievements(achievement_id) ON DELETE CASCADE ON UPDATE CASCADE;
```

### Key implications

- **Composite primary key `(member_id, achievement_id)`** — a member can unlock a given achievement at most once. A second `INSERT` raises Postgres error `23505 unique_violation` → translate to `409 Conflict`.
- **`unlocked_at` is server-set** — the request body MUST NOT accept `unlocked_at`; the DB default `CURRENT_TIMESTAMP` is the single source of truth. Never trust a client timestamp here.
- **`ON DELETE CASCADE` on both sides** — if a member or achievement is deleted, all unlock rows vanish automatically. The API never needs to manually clean these up.

---

## 🧱 Existing partial implementation — DO NOT DUPLICATE

A subset of this feature already exists because the member profile dashboard (#21) needed to render badges:

- **Entity** `internal/entity/member_achievement_entity.go` — already defined as a **joined view** (carries achievement fields + `unlocked_at`). Reuse this; do NOT create a parallel "pure-row" entity.

  ```go
  type MemberAchievement struct {
      AchievementId string `db:"achievement_id"`
      Nama          string `db:"nama"`
      Deskripsi     string `db:"deskripsi"`
      BadgeIcon     string `db:"badge_icon"`
      XPRequired    int    `db:"xp_required"`
      Tier          string `db:"tier"`
      UnlockedAt    string `db:"unlocked_at"`
  }
  ```

- **Repository** `internal/repository/member_achievement_repository.go` — already has `FindAllByMemberId(ctx, memberId)` returning the joined view. **Extend the existing interface in this file**, do NOT declare a second `MemberAchievementRepository`.

- **Bootstrap wiring** `internal/config/app.go:43` — `memberAchievementRepo` is already constructed and currently passed into `MemberUseCase`. You will additionally pass it into the new `MemberAchievementUseCase`.

Anything new (request DTOs, response DTOs other than the joined view, usecase, controller, route entries) is to be added by this PR.

---

## 🎯 Goals

1. Member can record a new achievement unlock via a single `POST` call from the front-end.
2. Member can list their own unlocked achievements (joined with the catalog).
3. Member can fetch a specific unlock by `achievement_id` (joined).
4. Super-admin can delete an erroneous unlock for corrective action (member cannot — once earned, the member cannot self-revoke).
5. **Strict ownership**: a member can only act on rows where `member_id` matches the `member_id` extracted from their JWT claims. Any attempt to operate on another member's row → `403 Forbidden`.
6. **Server-side XP gating** (see *Trust Model* below).
7. **Idempotency-friendly conflict response** on duplicate unlocks.

---

## 🔐 Trust Model (read carefully)

The front-end declares "this member just won achievement X." The server is the sole authority for:

1. **Whose unlock is this?** — Always taken from JWT claims (`extractMemberIdFromClaims(claims)`), **never** from the request body. Even if the client sends `member_id`, the server ignores it.
2. **Does the achievement exist?** — `FindById` on the achievements table. `404 Not Found` if missing.
3. **Has the member already unlocked it?** — Pre-check; if yes, return `409 Conflict` with a clear message. (Belt-and-suspenders: also handle Postgres `23505` from the actual `INSERT`, in case of a race.)
4. **Does the member meet the XP threshold?** — Compare `members.total_xp` against `achievements.xp_required`. **If `member.TotalXP < achievement.XPRequired`, return `403 Forbidden` with body `{"errors":"XP belum mencukupi untuk membuka achievement ini"}`.** This is the single most important non-obvious rule — without it, a member could call `POST /v1/member/achievements {"achievement_id":"1"}` and instantly grant themselves the top-tier badge.

> 🚨 Do **not** skip the XP check on the assumption the front-end already validated it. The front-end is untrusted; this is a public REST endpoint behind only a bearer token.

---

## 🔧 Scope

### A. Models — `internal/model/member_achievement.go` (new file)

```go
package model

// ==================== Requests ====================

// POST /v1/member/achievements
// member_id is intentionally absent — taken from JWT claims.
// unlocked_at is intentionally absent — server-set via DB default.
type MemberAchievementCreateRequest struct {
    AchievementId string `json:"achievement_id" validate:"required,numeric"`
}

// ==================== Responses ====================

// Returned when joining with achievements catalog — mirrors entity.MemberAchievement.
type MemberAchievementResponse struct {
    AchievementId string `json:"achievement_id"`
    Nama          string `json:"nama"`
    Deskripsi     string `json:"deskripsi"`
    BadgeIcon     string `json:"badge_icon"`
    XPRequired    int    `json:"xp_required"`
    Tier          string `json:"tier"`
    UnlockedAt    string `json:"unlocked_at"`
}
```

### B. Converter — `internal/model/converter/member_achievement_converter.go` (new file)

- [ ] `ToMemberAchievementResponse(*entity.MemberAchievement) *model.MemberAchievementResponse`
- [ ] `ToMemberAchievementResponses([]*entity.MemberAchievement) []*model.MemberAchievementResponse` — must return a non-nil empty slice `[]` (never `null`) when input is empty, matching the convention in `member_social_link_converter.go`.

### C. Repository — extend `internal/repository/member_achievement_repository.go`

**Extend the existing interface — do not redeclare it.** The current shape is:

```go
type MemberAchievementRepository interface {
    FindAllByMemberId(ctx context.Context, memberId string) ([]*entity.MemberAchievement, error)
}
```

Add these methods:

```go
type MemberAchievementRepository interface {
    FindAllByMemberId(ctx context.Context, memberId string) ([]*entity.MemberAchievement, error)
    FindOne(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error) // joined, like FindAllByMemberId
    Exists(ctx context.Context, memberId, achievementId string) (bool, error)
    Create(ctx context.Context, memberId, achievementId string) (*entity.MemberAchievement, error)  // inserts; returns the joined row
    Delete(ctx context.Context, memberId, achievementId string) error
}
```

Implementation rules:

- [ ] Convert both `memberId` and `achievementId` from `string` → `int` with `strconv.Atoi` (consistent with existing `FindAllByMemberId`). Return an error if either fails to parse — usecase will translate to `404`.
- [ ] All `SELECT` queries reuse the same join shape already established in `FindAllByMemberId` (`a.achievement_id::text`, `a.tier::text`, `ma.unlocked_at::text`) so `pgx.RowToAddrOfStructByNameLax[entity.MemberAchievement]` continues to work.
- [ ] `Create` should return the **joined** entity (with `nama`, `tier`, etc.). Two acceptable patterns — pick one and stay consistent:
  1. **Two-trip**: `INSERT INTO member_achievements (member_id, achievement_id) VALUES ($1, $2) ON CONFLICT DO NOTHING RETURNING 1`; if zero rows returned, the row already existed → return a typed `ErrAlreadyExists` sentinel; usecase maps to `409`. Then call `FindOne(memberId, achievementId)` to fetch the joined view.
  2. **Single-trip CTE**: `WITH ins AS (INSERT ... RETURNING member_id, achievement_id) SELECT a.achievement_id::text, ... FROM ins JOIN achievements a USING (achievement_id)` — acceptable if implemented cleanly.
- [ ] `Delete` runs `DELETE FROM member_achievements WHERE member_id = $1 AND achievement_id = $2` and checks `cmd.RowsAffected()`. If `0`, return a typed `ErrNotFound` sentinel; usecase maps to `404`.
- [ ] Translate `pgconn.PgError`:
  - `23505` (unique_violation on PK) → `ErrAlreadyExists` (fallback for the race against `Exists`).
  - `23503` (foreign_key_violation on `achievement_id`) → `ErrAchievementNotFound`. Log loudly if the violation is on `member_id`, because that means the JWT references a nonexistent member — bug, not user error → return `500` upstream.

### D. Usecase — `internal/usecase/member_achievement_usecase.go` (new file)

Interface:

```go
type MemberAchievementUseCase interface {
    Create(ctx context.Context, req *model.MemberAchievementCreateRequest, claims *model.Claims) (*model.MemberAchievementResponse, error)
    FindAllMine(ctx context.Context, claims *model.Claims) ([]*model.MemberAchievementResponse, error)
    FindOne(ctx context.Context, achievementId string, claims *model.Claims) (*model.MemberAchievementResponse, error)
    Delete(ctx context.Context, memberId, achievementId string, claims *model.Claims) error // super_admin-only path
}
```

Dependencies:

```go
type memberAchievementUseCaseImpl struct {
    Repo            repository.MemberAchievementRepository
    MemberRepo      repository.MemberRepository       // for TotalXP lookup
    AchievementRepo repository.AchievementRepository  // for xp_required + existence
    Log             *logrus.Logger
    Validator       *validator.Validate
}
```

**Create rules (most important):**

- [ ] `u.Validator.Struct(req)` → on fail return `fiber.ErrBadRequest`.
- [ ] `memberId := extractMemberIdFromClaims(claims)` (shared helper at `internal/usecase/member_usecase.go:224`). If empty → `fiber.ErrForbidden`.
- [ ] Look up the achievement: `ach, err := u.AchievementRepo.FindById(ctx, req.AchievementId)`. On not found → `fiber.NewError(fiber.StatusNotFound, "achievement tidak ditemukan")`.
- [ ] Look up the member: `member, err := u.MemberRepo.FindById(ctx, memberId)`. On not found → `fiber.ErrForbidden` (claim doesn't match a real member).
- [ ] **XP gate**: if `member.TotalXP < ach.XPRequired`, return `fiber.NewError(fiber.StatusForbidden, "XP belum mencukupi untuk membuka achievement ini")`.
- [ ] Duplicate pre-check: `exists, _ := u.Repo.Exists(ctx, memberId, req.AchievementId)`; if true → `fiber.NewError(fiber.StatusConflict, "achievement sudah pernah di-unlock")`.
- [ ] Insert: `result, err := u.Repo.Create(ctx, memberId, req.AchievementId)`.
  - If `ErrAlreadyExists` (race) → `409` with same message as above.
  - If `ErrAchievementNotFound` → `404`.
  - Otherwise log + `500`.
- [ ] Return `converter.ToMemberAchievementResponse(result)`.

**FindAllMine rules:**

- [ ] Extract `memberId`; empty → `403`.
- [ ] Call `Repo.FindAllByMemberId`. Repo errors → `500`. Empty list → return `[]` (converter handles it).

**FindOne rules:**

- [ ] Extract `memberId`; empty → `403`.
- [ ] `Repo.FindOne(memberId, achievementId)` — if not found → `404`.

**Delete rules (super_admin path):**

- [ ] Verify `claims.Role == "super_admin"`. If not → `fiber.ErrForbidden`. (Defense in depth — route middleware will already gate this, but the usecase does not blindly trust its callers.)
- [ ] Call `Repo.Delete(memberId, achievementId)`. `ErrNotFound` → `404`. Other errors → `500`.

### E. Controller — `internal/delivery/http/member_achievement_controller.go` (new file)

Follow `member_social_link_controller.go` line-for-line for body parsing, claims extraction, status codes, and the `WebResponse[T]` envelope.

```go
type MemberAchievementController interface {
    Create(ctx fiber.Ctx) error      // POST   /v1/member/achievements
    FindAllMine(ctx fiber.Ctx) error // GET    /v1/member/achievements
    FindOne(ctx fiber.Ctx) error     // GET    /v1/member/achievements/:achievement_id
    Delete(ctx fiber.Ctx) error      // DELETE /v1/member/achievements/:member_id/:achievement_id  (super_admin)
}
```

Status codes:

| Endpoint | Success status | Notes |
| -------- | -------------- | ----- |
| `POST /v1/member/achievements` | `201 Created` | Body = `MemberAchievementCreateRequest` |
| `GET /v1/member/achievements` | `200 OK` | Returns array; no `paging` block — list is naturally small (≤ # achievements in catalog) |
| `GET /v1/member/achievements/:achievement_id` | `200 OK` | |
| `DELETE /v1/member/achievements/:member_id/:achievement_id` | `200 OK` | Body `{"data":"Member achievement deleted successfully"}` |

### F. Route registration — `delivery/http/route/route.go`

Add to `RouteConfig`:

```go
MemberAchievementController http.MemberAchievementController
```

Inside `SetupAuthRoutes`, add a new block **immediately after the existing MEMBER SOCIAL LINKS block** (around line 234):

```go
// ==========================================
// MEMBER ACHIEVEMENTS
//   - create / read: member only (self)
//   - delete: super_admin only (corrective)
// ==========================================
auth.Post("/member/achievements", memberOnly, c.MemberAchievementController.Create)
auth.Get("/member/achievements", memberOnly, c.MemberAchievementController.FindAllMine)
auth.Get("/member/achievements/:achievement_id", memberOnly, c.MemberAchievementController.FindOne)
auth.Delete("/member/achievements/:member_id/:achievement_id", superadminOnly, c.MemberAchievementController.Delete)
```

> ⚠️ Route ordering note: `auth.Get("/member/:id", ...)` already exists at line 211. Fiber v3 matches routes in registration order, but `/member/achievements` is a **static segment** that takes precedence over the `:id` param when registered first. Register the achievements routes **before** `/member/:id` would normally match, OR — safer — keep `/member/:id` as-is and rely on the fact that the literal string `"achievements"` will be parsed as `id` and then fail downstream (numeric validation in `FindById` will 404, not collide). **Hand-test `GET /v1/member/achievements` returns the list, not a 404 from `MemberController.FindById`.** If a collision is observed, move the achievements block above line 211.

### G. Wiring — `internal/config/app.go`

- [ ] No new repo construction needed — `memberAchievementRepo` already exists at line 43; reuse it.
- [ ] Construct the usecase (place near line 62, after `AchievementUseCase`):
  ```go
  MemberAchievementUseCase := usecase.NewMemberAchievementUseCase(memberAchievementRepo, memberRepo, achievementRepo, cfg.Log, cfg.Validate)
  ```
- [ ] Construct the controller (place near line 84):
  ```go
  MemberAchievementController := http.NewMemberAchievementController(MemberAchievementUseCase, cfg.Log)
  ```
- [ ] Add `MemberAchievementController` to the `route.RouteConfig{...}` literal (around line 106).

### H. OpenAPI — `docs/openapi.yaml`

- [ ] Add 4 paths under `/v1/member/achievements*`.
- [ ] Add schemas: `MemberAchievementCreateRequest`, `MemberAchievementResponse`.
- [ ] All paths use `security: [BearerAuth: []]`.
- [ ] Document responses: `201`, `400` (validation), `403` (XP insufficient / forbidden), `404` (achievement or unlock not found), `409` (duplicate unlock).
- [ ] Note in the `POST` description: "`unlocked_at` is server-generated; do not send it." and "`member_id` is derived from the JWT; do not send it."

### I. Tests

- [ ] **Repository** (`internal/repository/member_achievement_repository_test.go` — extend existing or create)
  - `Create` happy path → returns joined entity with all fields populated.
  - `Create` second time with same pair → `ErrAlreadyExists`.
  - `Create` with nonexistent `achievement_id` → `ErrAchievementNotFound` (translated from FK violation).
  - `FindOne` not found → `ErrNotFound`.
  - `Delete` happy path; `Delete` non-existent row → `ErrNotFound`.
- [ ] **Usecase** (`internal/usecase/member_achievement_usecase_test.go` — new)
  - Create with `req.AchievementId` missing → `400`.
  - Create when claims has no `member_id` → `403`.
  - **Create when member's `total_xp < ach.xp_required` → `403` with the Indonesian message above, and repo `Create` is never called.** (Must-have assertion — without it, the XP gate is just hopeful comments.)
  - Create when achievement does not exist → `404`.
  - Create when already unlocked (pre-check) → `409`, repo `Create` not called.
  - Create happy path → response payload populated, repo `Create` called exactly once.
  - FindAllMine when member has zero unlocks → returns `[]`, not `null`.
  - FindOne scopes by `memberId` from claims — pass two different `Claims` and assert isolation.
  - Delete invoked with non-superadmin role → `403`.
- [ ] **Controller** (`internal/delivery/http/member_achievement_controller_test.go` — new)
  - Each endpoint returns the `WebResponse[T]` envelope and correct HTTP code.
  - `POST` returns `201` on success.
  - `DELETE` returns string data payload.

---

## 🔐 Authorization Matrix

| Endpoint | member (self) | member (other) | guru | super_admin |
| -------- | :-----------: | :------------: | :--: | :---------: |
| `POST /v1/member/achievements` | ✅ | n/a* | ❌ | ❌ |
| `GET  /v1/member/achievements` | ✅ (own list) | n/a* | ❌ | ❌ |
| `GET  /v1/member/achievements/:achievement_id` | ✅ (own row) | n/a* | ❌ | ❌ |
| `DELETE /v1/member/achievements/:member_id/:achievement_id` | ❌ | ❌ | ❌ | ✅ |

`*` "Other member" cannot be addressed at all — the resource is implicitly scoped to the caller via JWT claims. There is intentionally no `GET /v1/member/:id/achievements` for cross-member viewing in this issue.

❌ = `403 Forbidden` from `RoleMiddleware`.

---

## 📦 Example payloads

### `POST /v1/member/achievements`

Request (member is logged in; their `member_id` is `7`):
```json
{
  "achievement_id": "2"
}
```

Response `201`:
```json
{
  "data": {
    "achievement_id": "2",
    "nama": "Penjelajah Awal",
    "deskripsi": "Selesaikan 10 cerita interaktif",
    "badge_icon": "https://arsiva.test/uploads/badges/penjelajah-awal.webp",
    "xp_required": 500,
    "tier": "silver",
    "unlocked_at": "2026-05-11 10:14:33"
  }
}
```

Response `403` (member has only 320 XP, achievement requires 500):
```json
{ "errors": "XP belum mencukupi untuk membuka achievement ini" }
```

Response `409` (already unlocked earlier):
```json
{ "errors": "achievement sudah pernah di-unlock" }
```

Response `404` (achievement `999` does not exist):
```json
{ "errors": "achievement tidak ditemukan" }
```

### `GET /v1/member/achievements`

Response `200`:
```json
{
  "data": [
    {
      "achievement_id": "2",
      "nama": "Penjelajah Awal",
      "deskripsi": "Selesaikan 10 cerita interaktif",
      "badge_icon": "https://arsiva.test/uploads/badges/penjelajah-awal.webp",
      "xp_required": 500,
      "tier": "silver",
      "unlocked_at": "2026-05-11 10:14:33"
    }
  ]
}
```

Empty case:
```json
{ "data": [] }
```

### `DELETE /v1/member/achievements/7/2` (super_admin only)

Response `200`:
```json
{ "data": "Member achievement deleted successfully" }
```

---

## ✅ Acceptance Criteria

- Member calling `POST /v1/member/achievements {"achievement_id":"X"}` from the front-end successfully inserts an unlock and gets the joined achievement back as `201`.
- Member with insufficient `total_xp` is rejected with `403` **at the usecase layer** — verified by a unit test that asserts the repo `Create` is never called.
- Re-unlocking the same achievement returns `409` and leaves the original `unlocked_at` untouched (no overwrite).
- Posting a nonexistent `achievement_id` returns `404`, not `500`.
- `member_id` is never read from the request body — even if the client sends `{"member_id":"99", "achievement_id":"2"}`, the unlock is recorded against the JWT's `member_id`.
- `GET /v1/member/achievements` returns `[]` (not `null`) when the member has no badges.
- Guru and other members cannot reach this resource (`403` from middleware).
- Super-admin `DELETE /v1/member/achievements/:member_id/:achievement_id` removes the row; the achievement and member rows themselves are untouched.
- `go test ./...` is green; OpenAPI spec compiles.

---

## 📚 Related

- Migration: `db/migrations_postgre/20260318203943_create_table_member_achievement.up.sql`.
- FKs: `db/migrations_postgre/20260318204005_create_foreign_key.up.sql:65-67`.
- Companion resource: Achievement catalog CRUD (#22).
- Pattern reference (member-scoped self-service, claims-driven): Member Social Links CRUD (#21).
- Existing partial implementation to **extend** (not replace): `internal/repository/member_achievement_repository.go`, `internal/entity/member_achievement_entity.go`.
- Claims helper to reuse: `extractMemberIdFromClaims` at `internal/usecase/member_usecase.go:224`.
