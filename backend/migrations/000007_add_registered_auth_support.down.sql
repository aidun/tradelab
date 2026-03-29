DROP INDEX IF EXISTS users_clerk_user_id_idx;

ALTER TABLE users
  DROP COLUMN IF EXISTS auth_provider,
  DROP COLUMN IF EXISTS clerk_user_id;

ALTER TABLE users
  ALTER COLUMN password_hash SET NOT NULL,
  ALTER COLUMN email SET NOT NULL;
