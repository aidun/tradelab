CREATE TABLE strategies (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    market_id UUID NOT NULL REFERENCES markets(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    config_json JSONB NOT NULL,
    reference_price NUMERIC(18, 8) NOT NULL DEFAULT 0,
    last_run_at TIMESTAMPTZ NULL,
    last_decision TEXT NULL,
    last_outcome TEXT NULL,
    last_reason TEXT NULL,
    evaluation_claim_token TEXT NULL,
    evaluation_claimed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (wallet_id, market_id)
);

CREATE INDEX idx_strategies_wallet_status ON strategies(wallet_id, status);
CREATE INDEX idx_strategies_claim ON strategies(status, evaluation_claimed_at);

CREATE TABLE strategy_runs (
    id UUID PRIMARY KEY,
    strategy_id UUID NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
    decision TEXT NOT NULL,
    outcome TEXT NOT NULL,
    reason TEXT NOT NULL,
    details_json JSONB NULL,
    evaluation_duration_ms BIGINT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_strategy_runs_strategy_finished ON strategy_runs(strategy_id, finished_at DESC);
