# Security Model

## Purpose

This document defines the current security posture for TradeLab across browser storage, sessions, secrets, logs, and sensitive data handling.

TradeLab remains a demo-only product, but it still enforces explicit security boundaries for guest sessions, registered accounts, backend sessions, and operator-managed secrets.

## Identity and session model

TradeLab now uses two identity layers:

- `Clerk` for external identity
- `TradeLab app sessions` for backend authorization

### Guest mode

- guest sessions are created by the backend
- guest session tokens are visible to the browser
- guest tokens are stored in `sessionStorage` only
- guest tokens are never intended to survive browser-session shutdown

### Registered mode

- Clerk authenticates the user
- the backend verifies the Clerk token during bootstrap or upgrade
- the backend creates an opaque TradeLab app session
- the browser receives the registered app session only through an `HttpOnly` cookie
- frontend code does not store or read the registered app-session token directly

## Session lifecycle rules

### Guest sessions

- guest sessions remain temporary and lower-trust
- guest expiry is bounded
- invalid guest session behavior should create a new guest session with explicit UX messaging

### Registered app sessions

- registered app sessions use an idle timeout and an absolute lifetime
- a new registered app session is created on login/bootstrap and on guest upgrade
- older active registered sessions for the same user are revoked when a new one is created
- logout revokes the current registered app session immediately
- expired or revoked registered sessions require re-authentication through Clerk

## Browser storage rules

### Allowed browser-visible data

- guest session token in `sessionStorage`
- non-sensitive UI hints such as account mode or display name when needed
- development-only mock-auth state

### Forbidden browser-visible data

- registered app-session token
- database credentials
- backend secret material
- Clerk secret keys

## Secret handling policy

### Local development

- secrets may exist only in ignored env files or local shell environment
- no secret-bearing file should be committed into repo-tracked state

### Development Kubernetes

- secret values come from ignored local secret inputs only
- mock auth may be enabled for development and CI, but it is not a production contract

### Production

- production secrets come from an external secret store only
- publishable Clerk keys may be exposed to the frontend
- Clerk secret material and database credentials stay server-side only

### CI and release

- secrets come from GitHub Actions secret boundaries or registry auth
- workflows must never echo secret values or raw credentials into logs

## Logging and redaction rules

The backend must never log:

- raw guest session tokens
- Clerk bearer tokens
- cookies or cookie values
- `Authorization` headers
- passwords
- raw database URLs with embedded credentials
- secret values from env or operator input

The backend may log:

- session ids
- wallet ids
- user ids
- truncated external identity ids
- sanitized error strings

## Sensitive-data inventory

| Data | Where it exists | Browser-visible | Purpose | Retention expectation |
| --- | --- | --- | --- | --- |
| guest session token | backend response + browser `sessionStorage` | yes | guest authorization | browser session only |
| guest session hash | PostgreSQL `demo_sessions` | no | guest session lookup | until expiry/cleanup |
| registered app-session cookie token | browser cookie | no to JS | registered authorization | until idle/absolute expiry or logout |
| registered app-session hash | PostgreSQL `app_sessions` | no | registered session lookup and revocation | until expiry/cleanup |
| Clerk user id | PostgreSQL `users.clerk_user_id` | not by default | durable identity mapping | long-lived account data |
| email and display name | PostgreSQL `users` and frontend UI | display name yes, email only when intentionally surfaced | account identity and UX | long-lived account data |
| database credentials | env / secret store | no | infrastructure access | operator-managed |

## Mock-auth boundary

Mock auth exists only for:

- local development
- automated frontend tests
- CI flows that need deterministic sign-in behavior

Mock auth must not:

- be enabled in production manifests
- be documented as a production feature
- be treated as a user-facing identity mode

## Related documentation

- [authentication-model.md](authentication-model.md)
- [clerk-architecture.md](clerk-architecture.md)
- [deployment.md](deployment.md)
- [system-operations.md](system-operations.md)
- [developer-guide.md](developer-guide.md)
