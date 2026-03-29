# Clerk Architecture

## Purpose

This document defines how Clerk should integrate with the TradeLab frontend and backend without breaking the current guest-first demo flow.

## Boundary model

TradeLab should use two layers:

1. `Clerk identity layer`
2. `TradeLab application session layer`

Clerk is responsible for authenticating a user. The TradeLab backend remains responsible for mapping that identity to demo accounts, wallets, and trading-specific application state.

## Frontend responsibilities

The Next.js frontend should:

- render Clerk UI for signup, login, account, and logout flows
- distinguish clearly between guest mode and registered mode
- attach identity context when calling registered-account backend flows
- continue to support the current guest bootstrap path for first-time use

## Backend responsibilities

The Go backend should:

- trust only verified Clerk identity data
- map Clerk user IDs to internal TradeLab users
- create and manage application demo sessions for registered users
- keep authorization logic server-side
- avoid trusting raw client claims for wallet ownership or account identity

## Trust model

The backend should not treat the frontend as the source of truth for identity.

Instead:

- Clerk verifies the user at the edge or frontend boundary
- the backend verifies the authenticated identity context for protected registered-account routes
- TradeLab maps that identity to internal state such as user IDs, wallets, and portfolios

## Guest mode coexistence

Guest mode should remain separate from the Clerk-backed path.

This implies:

- guest session creation remains available
- guest sessions are treated as temporary and low-trust
- registered mode uses verified external identity and durable internal ownership

## Route model

The future route model should distinguish:

- public routes
- guest demo routes
- registered-account routes

Examples:

- public:
  - health
  - market list
  - market candles
- guest:
  - demo session bootstrap
  - guest portfolio access
  - guest demo orders
- registered:
  - account bootstrap
  - durable portfolio access
  - account-bound trading history
  - future strategies and backtests

## Internal mapping

TradeLab should maintain an internal user record even for Clerk-backed users.

The internal model should allow:

- one Clerk identity to map to one TradeLab user
- one TradeLab user to own wallets, orders, positions, and activity
- future expansion without coupling product data tightly to Clerk primitives

## Session verification model

The current bearer-token demo-session model is sufficient for guest mode, but the registered path should move toward:

- verified Clerk identity
- backend-issued or backend-resolved application session context
- server-side authorization checks for wallet and account access

## Implementation notes

When this architecture is implemented:

- frontend auth libraries and middleware must be added
- backend user resolution and account mapping must be introduced
- the API surface must distinguish guest and registered flows explicitly
- logging and security rules must be reviewed so identity values are useful but not sensitive

## Related documentation

- [authentication-model.md](authentication-model.md)
- [auth-flows.md](auth-flows.md)
- [account-lifecycle.md](account-lifecycle.md)
