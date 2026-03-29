# Authentication Flows

## Purpose

This document defines the intended signup, login, logout, session-expiry, and social-login flows for TradeLab.

## Supported modes

- guest demo session
- registered account via Clerk
- social login providers:
  - Google
  - Apple

## Flow 1. Guest first launch

1. User opens the app.
2. TradeLab creates a guest demo session automatically.
3. The dashboard loads with demo-only messaging.
4. The user can explore markets and place demo trades without registration.

Goal:

- reduce first-use friction
- let users understand the product before committing to signup

## Flow 2. Sign up

1. User sees the durable-account prompt after the dashboard has loaded and first value is visible.
2. Clerk presents available signup methods.
3. User signs up with Google, Apple, or another enabled provider.
4. TradeLab asks whether guest demo data should be preserved or discarded.
5. TradeLab creates or links the internal user record.
6. TradeLab establishes a registered `HttpOnly` app session.
7. TradeLab initializes the durable demo account context.

Goal:

- convert a guest or new user into a durable product account

## Flow 3. Login

1. User returns to TradeLab.
2. Clerk authenticates the user.
3. The backend resolves the internal user and account context.
4. The backend establishes or refreshes the registered app session.
5. The product opens into the user-owned demo environment without forcing guest re-entry.

Goal:

- restore durable access across sessions and devices

## Flow 4. Logout

1. User signs out from the account controls.
2. TradeLab revokes the registered app session.
3. Clerk session is terminated in the frontend.
4. Any active registered-account application context is cleared.
5. TradeLab falls back to a new guest demo session automatically.

Goal:

- make account exit explicit and predictable

## Flow 5. Session expiry

TradeLab should distinguish:

- guest demo-session expiry
- registered identity/session expiry

Expected behavior:

- expired guest sessions prompt the user to refresh into a new guest session or sign up
- expired registered sessions prompt re-authentication through Clerk
- the UI should never fail silently when auth state is no longer valid

## Social login requirements

Google and Apple are the primary provider targets.

The UX should:

- show clearly which providers are supported
- keep the provider surface compact and trustworthy
- allow future provider additions without redesigning the auth boundary

## UX requirements

The auth surface should communicate:

- guest mode is temporary
- registered mode is durable
- the product remains demo-only in both cases

## Related documentation

- [authentication-model.md](authentication-model.md)
- [clerk-architecture.md](clerk-architecture.md)
- [account-lifecycle.md](account-lifecycle.md)
