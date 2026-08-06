-- Unique player identity (case-insensitive).

-- Keep lowest id when duplicate emails exist.
DELETE FROM players a
USING players b
WHERE a.id > b.id
  AND LOWER(a.email) = LOWER(b.email)
  AND a.email IS NOT NULL
  AND b.email IS NOT NULL
  AND a.email <> ''
  AND b.email <> '';

-- Keep lowest id when duplicate usernames exist.
DELETE FROM players a
USING players b
WHERE a.id > b.id
  AND LOWER(a.username) = LOWER(b.username)
  AND a.username IS NOT NULL
  AND b.username IS NOT NULL
  AND a.username <> ''
  AND b.username <> '';

CREATE UNIQUE INDEX IF NOT EXISTS uq_players_email_lower
    ON players (LOWER(email));

CREATE UNIQUE INDEX IF NOT EXISTS uq_players_username_lower
    ON players (LOWER(username));
