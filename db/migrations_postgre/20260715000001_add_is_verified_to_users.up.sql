ALTER TABLE users ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT false;

-- Accounts that already exist (e.g. the seeded admin/guru/member) predate the
-- email-verification gate, so mark them verified to avoid locking them out.
UPDATE users SET is_verified = true;
