# Automatic achievements and persistent notifications

Finalization uses one PostgreSQL transaction for submit, solve, story endings and the expired-session worker. It locks the member row before progress, streak/daily-task updates, XP and achievement writes. Content XP remains one-time per member/content; an earlier zero-XP attempt does not consume the reward. Existing scoring and daily-task eligibility rules are unchanged. The coordinator computes final level after all bonuses, inserts eligible achievements with ON CONFLICT using the existing primary key, creates notifications only for returned inserts, stores the response snapshot, then commits. A notification failure rolls back all effects. Invalid/deleted group references are omitted; existing references are locked against deletion during the insert.

Start returns a backend UUID `completion_id`, also available in GET session. Store it per play session and send it in submit, solve and scene requests. Requests without the ID remain compatible while the session exists, but cannot recover a response after cleanup. The ID is scoped to the authenticated member and content. A new playthrough gets a new ID.

Redis session creation is atomic. Finalize atomically freezes and reads the session; subsequent gameplay writes cannot alter the score. Failed transactions leave this frozen session retryable, including solve and ending-scene retries. Concurrent finalizers serialize on the member lock and return the same persisted result. Processing has a 15-second deadline: retry processing errors with the same ID. A mismatched active ID returns 409; an unknown completion with no matching session returns 404. No successful empty completion is returned.

After commit, conditional Redis cleanup cannot delete a newer session. A retry checks PostgreSQL before Redis, so lost HTTP responses, failed Redis cleanup, and ambiguous commits resolve through the same snapshot. A worker retry also checks the snapshot before reading current content metadata. Redis sessions that expire before any successful commit cannot be reconstructed; keep the worker running within the existing 30-minute session grace period. Existing pre-deployment sessions have no completion ID: drain them before deployment, or let their existing TTL expire and restart them.

The response preserves `data`, `progres_id`, `previous_level`, `new_level`, `leveled_up` and story `ending_type`. Additions are `completion_id`, `skor`, `awarded_xp` (effective content XP), `bonus_xp`, `total_awarded_xp`, `total_xp`, `newly_unlocked_achievements` and `notifications`. Both arrays are always arrays, including `[]`. A non-ending scene still returns `data: null`. Worker completions do not push messages; clients retrieve the inbox later.

For example, a member at 350 XP finishing a quiz with 30 content XP and 20 daily XP reaches 400 XP and level 2. Replay returns those same values and notification IDs, even after the notification has been read. Snapshot `read_at` describes event-time state; fetch the inbox for current state.

## Frontend integration and rollout

Render score and XP from the completion response. Show level animation only when `leveled_up`. Choose notifications as the single popup source and deduplicate `notification_id` per account; do not also loop over newly unlocked achievements. Popup display does not mark an inbox notification read. Device-local popup dedupe does not guarantee exactly-once presentation across devices.

Deploy the migration and backend first, then update the frontend to persist completion IDs and stop looping over POST `/v1/member/achievements`. That endpoint is deprecated but remains an idempotent compatibility adapter to the same evaluator; the requested ID only selects its legacy response. Remove its route/controller/request schema in a coordinated release after frontend migration is confirmed. The frontend application is not in this backend repository.

GET `/v1/member/notifications?page=1&size=20&unread_only=false` returns `data` plus existing `paging` metadata, ordered by created_at DESC and notification_id DESC. Read notifications are retained; no retention deletion policy is introduced. GET never marks read. GET `/unread-count` returns `data.count`. PATCH `/:notification_id/read` preserves the first read timestamp and returns the current notification. All endpoints require authenticated member role and completed profile. Missing or another member's notification returns 404; missing authentication returns 401 and wrong role 403. Payloads retain badge metadata even when the catalog changes.

Admin DELETE is a data correction, not permanent revocation: eligible badges can be awarded again, but the existing `achievement_unlocked:<id>` notification dedupe key prevents a second inbox event. Permanent suppression is outside this release.

## Reconciliation

Run `DATABASE_URL=... go run ./cmd/reconcile-achievements -batch-size 100` in a configured environment. It scans members by ID in bounded batches, locks each member and uses the same evaluator. Default is silent: no historical notifications or unread flood. `-notify` explicitly enables inbox events for newly awarded badges. It reports awarded/skip/error counts and exits nonzero on errors; rerun safely. Newly added badges or lowered thresholds can use the same command. Raising a threshold does not revoke existing badges. Zero-XP badges are evaluated on completion or reconciliation, not registration.

## Verification

`go test ./...` runs unit/HTTP/Redis tests. Set TEST_DATABASE_URL to a PostgreSQL role with CREATEDB to run `go test ./internal/repository -run TestCompletionTransactionIntegration -v`; the test creates and drops an isolated database and applies real migrations, then tests rollback, concurrent finalization, replay, one-time XP, threshold crossing, notification ownership/pagination/read state, reconciliation and migration down/up. A historical duplicate FK migration is normalized only in the disposable test database; historical migration files are unchanged.
