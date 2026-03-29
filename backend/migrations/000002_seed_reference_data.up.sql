INSERT INTO assets (id, symbol, name, asset_type)
VALUES
  ('0f7c4db4-1d41-42a2-88fb-7eb4c8ef9ef8', 'USDT', 'Tether USD', 'token'),
  ('a2860ecf-d2d2-4128-a0f5-11dd7f373939', 'XRP', 'XRP', 'coin'),
  ('65da84fd-c91a-4bde-8dfa-e4c5f3478a11', 'BTC', 'Bitcoin', 'coin'),
  ('f4379a54-1a24-4654-bb3d-cfcb954965ef', 'ETH', 'Ethereum', 'coin')
ON CONFLICT (symbol) DO NOTHING;

INSERT INTO markets (id, base_asset_id, quote_asset_id, symbol, exchange_code, tick_size, min_order_size)
VALUES
  ('2d7de9b6-8f17-4faa-a36a-e1df1714367f', 'a2860ecf-d2d2-4128-a0f5-11dd7f373939', '0f7c4db4-1d41-42a2-88fb-7eb4c8ef9ef8', 'XRP/USDT', 'demo', 0.0001, 10),
  ('4cc80fb3-88f2-48fc-9b3d-8e64ebee5294', '65da84fd-c91a-4bde-8dfa-e4c5f3478a11', '0f7c4db4-1d41-42a2-88fb-7eb4c8ef9ef8', 'BTC/USDT', 'demo', 0.01, 10),
  ('df5d9756-53f2-4fea-a5cf-408f33db8ba5', 'f4379a54-1a24-4654-bb3d-cfcb954965ef', '0f7c4db4-1d41-42a2-88fb-7eb4c8ef9ef8', 'ETH/USDT', 'demo', 0.01, 10)
ON CONFLICT (symbol) DO NOTHING;
