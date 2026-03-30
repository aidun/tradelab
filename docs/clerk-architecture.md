# Clerk Architecture

## Purpose

This document defines how Clerk integrates with the TradeLab frontend and backend without breaking the current guest-first demo flow.

## Boundary model

TradeLab should use two layers:

1. `Clerk identity layer`
2. `TradeLab application session layer`

Clerk is responsible for authenticating a user. The TradeLab backend remains responsible for mapping that identity to demo accounts, wallets, and trading-specific application state.

## Frontend responsibilities

The Next.js frontend now:

- renders Clerk UI for signup, login, account, and logout flows when `NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY` is configured at runtime
- supports a local, CI, and development-cluster-friendly mock auth mode through `NEXT_PUBLIC_AUTH_MOCK_MODE=true`
- distinguish clearly between guest mode and registered mode
- attach identity context when calling registered-account backend flows
- continue to support the current guest bootstrap path for first-time use

## Backend responsibilities

The Go backend now:

- trusts only verified Clerk identity data for registered-account routes
- map Clerk user IDs to internal TradeLab users
- create or load the primary demo wallet for registered users
- create TradeLab app sessions for registered users after Clerk verification
- keep authorization logic server-side
- avoid trusting raw client claims for wallet ownership or account identity

## Trust model

The backend does not treat the frontend as the source of truth for identity.

Instead:

- Clerk verifies the user at the edge or frontend boundary
- the backend verifies the authenticated identity context for protected registered-account routes through Clerk JWT verification or explicit mock mode
- the backend then establishes an opaque registered app session for normal product use
- TradeLab maps that identity to internal state such as user IDs, wallets, and portfolios

## Guest mode coexistence

Guest mode should remain separate from the Clerk-backed path.

This implies:

- guest session creation remains available
- guest sessions are treated as temporary and low-trust
- registered mode uses verified external identity plus durable internal ownership through HttpOnly app sessions

## Route model

The current route model distinguishes:

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
  - guest-to-registered upgrade
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

Current implementation notes:

- frontend auth libraries and provider wrappers are in place
- backend user resolution and account mapping are in place
- the API surface distinguishes guest and registered flows explicitly
- Phase 3 will still harden storage, redaction, and secret handling on top of this boundary

## Related documentation

- [authentication-model.md](authentication-model.md)
- [auth-flows.md](auth-flows.md)
- [account-lifecycle.md](account-lifecycle.md)
