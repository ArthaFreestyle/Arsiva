-- IF NOT EXISTS keeps this safe to (re-)apply even on a DB where the column was
-- added out-of-band (e.g. by hand during an incident) before migrate tracked it.
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_verified BOOLEAN NOT NULL DEFAULT false;

-- Accounts that already exist (e.g. the seeded admin/guru/member) predate the
-- email-verification gate, so mark them verified to avoid locking them out.
-- Runs once (migrate applies a version a single time), so it never re-verifies
-- accounts that register later.
UPDATE users SET is_verified = true;
