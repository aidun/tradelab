CREATE TABLE assets (
  id UUID PRIMARY KEY,
  symbol TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  asset_type TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE markets (
  id UUID PRIMARY KEY,
  base_asset_id UUID NOT NULL REFERENCES assets(id),
  quote_asset_id UUID NOT NULL REFERENCES assets(id),
  symbol TEXT NOT NULL UNIQUE,
  exchange_code TEXT NOT NULL,
  tick_size NUMERIC(18, 8) NOT NULL,
  min_order_size NUMERIC(18, 8) NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
  id UUID PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  display_name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wallets (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  wallet_type TEXT NOT NULL,
  base_currency TEXT NOT NULL,
  starting_balance NUMERIC(18, 8) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wallet_balances (
  id UUID PRIMARY KEY,
  wallet_id UUID NOT NULL REFERENCES wallets(id),
  asset_id UUID NOT NULL REFERENCES assets(id),
  available_amount NUMERIC(18, 8) NOT NULL,
  locked_amount NUMERIC(18, 8) NOT NULL DEFAULT 0,
  average_entry_price NUMERIC(18, 8) NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (wallet_id, asset_id)
);

CREATE TABLE orders (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  wallet_id UUID NOT NULL REFERENCES wallets(id),
  market_id UUID NOT NULL REFERENCES markets(id),
  strategy_id UUID,
  order_source TEXT NOT NULL,
  side TEXT NOT NULL,
  order_type TEXT NOT NULL,
  status TEXT NOT NULL,
  requested_quantity NUMERIC(18, 8),
  requested_quote_amount NUMERIC(18, 8),
  executed_quantity NUMERIC(18, 8),
  average_execution_price NUMERIC(18, 8),
  fee_amount NUMERIC(18, 8),
  fee_asset_id UUID REFERENCES assets(id),
  slippage_bps INTEGER NOT NULL DEFAULT 0,
  submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  executed_at TIMESTAMPTZ,
  cancelled_at TIMESTAMPTZ
);
