DELETE FROM wallet_balances
WHERE id IN (
  'f01f1d0e-af43-4b41-a64b-8f241ae9e96e',
  '754e59e0-f584-4cb0-99b8-69daf65f3a30'
);

DELETE FROM wallets
WHERE id = '1ddb1c1c-827f-4bf0-b85a-3d5786c3b26c';

DELETE FROM users
WHERE id = 'cfbf7c8f-eaf9-47fa-8674-2a29fed1fcc9';
