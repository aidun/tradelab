INSERT INTO users (id, email, password_hash, display_name)
VALUES ('cfbf7c8f-eaf9-47fa-8674-2a29fed1fcc9', 'demo@tradelab.dev', 'development-only', 'TradeLab Demo')
ON CONFLICT (email) DO NOTHING;

INSERT INTO wallets (id, user_id, wallet_type, base_currency, starting_balance)
VALUES ('1ddb1c1c-827f-4bf0-b85a-3d5786c3b26c', 'cfbf7c8f-eaf9-47fa-8674-2a29fed1fcc9', 'paper', 'USDT', 10000)
ON CONFLICT DO NOTHING;

INSERT INTO wallet_balances (id, wallet_id, asset_id, available_amount, locked_amount, average_entry_price)
VALUES
  ('f01f1d0e-af43-4b41-a64b-8f241ae9e96e', '1ddb1c1c-827f-4bf0-b85a-3d5786c3b26c', '0f7c4db4-1d41-42a2-88fb-7eb4c8ef9ef8', 10000, 0, 1),
  ('754e59e0-f584-4cb0-99b8-69daf65f3a30', '1ddb1c1c-827f-4bf0-b85a-3d5786c3b26c', 'a2860ecf-d2d2-4128-a0f5-11dd7f373939', 0, 0, 0)
ON CONFLICT (wallet_id, asset_id) DO NOTHING;
