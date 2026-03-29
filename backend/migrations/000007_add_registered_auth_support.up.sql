ALTER TABLE users
  ALTER COLUMN email DROP NOT NULL,
  ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE users
  ADD COLUMN clerk_user_id TEXT UNIQUE,
  ADD COLUMN auth_provider TEXT NOT NULL DEFAULT 'guest';

CREATE INDEX users_clerk_user_id_idx ON users (clerk_user_id);
