CREATE TABLE member_notifications (
 notification_id BIGSERIAL PRIMARY KEY,
 member_id INTEGER NOT NULL REFERENCES members(member_id) ON DELETE CASCADE,
 type TEXT NOT NULL CHECK(type = 'achievement_unlocked'),
 dedupe_key TEXT NOT NULL, title TEXT NOT NULL, message TEXT NOT NULL,
 payload JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), read_at TIMESTAMPTZ,
 UNIQUE(member_id,dedupe_key)
);
CREATE INDEX member_notifications_recent_idx ON member_notifications(member_id,created_at DESC,notification_id DESC);
CREATE INDEX member_notifications_unread_idx ON member_notifications(member_id) WHERE read_at IS NULL;
CREATE TABLE progress_completions (
 completion_id TEXT PRIMARY KEY,
 member_id INTEGER NOT NULL REFERENCES members(member_id) ON DELETE CASCADE,
 content_type TEXT NOT NULL, content_id INTEGER NOT NULL,
 result JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
