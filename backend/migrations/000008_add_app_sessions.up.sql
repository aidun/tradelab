CREATE TABLE app_sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    principal_kind TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    idle_expires_at TIMESTAMPTZ NOT NULL,
    absolute_expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_app_sessions_user_id ON app_sessions(user_id);
CREATE INDEX idx_app_sessions_wallet_id ON app_sessions(wallet_id);
CREATE INDEX idx_app_sessions_token_hash ON app_sessions(token_hash);
