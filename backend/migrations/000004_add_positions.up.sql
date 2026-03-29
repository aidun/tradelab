CREATE TABLE positions (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  wallet_id UUID NOT NULL REFERENCES wallets(id),
  market_id UUID NOT NULL REFERENCES markets(id),
  status TEXT NOT NULL,
  entry_quantity NUMERIC(18, 8) NOT NULL,
  entry_price_avg NUMERIC(18, 8) NOT NULL,
  opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX positions_open_wallet_market_idx
  ON positions (wallet_id, market_id, status);
