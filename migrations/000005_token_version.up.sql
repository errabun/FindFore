-- Invalidate JWTs after password change via monotonic token_version.

ALTER TABLE players
    ADD COLUMN IF NOT EXISTS token_version INTEGER NOT NULL DEFAULT 0;
