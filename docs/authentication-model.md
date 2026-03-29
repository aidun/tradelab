# Authentication Model

## Purpose

This document defines the TradeLab identity model for the transition from guest demo sessions to durable registered accounts.

The current product remains demo-only, but authentication is now a first-class product surface. The target identity provider is `Clerk`, with guest mode kept as the low-friction entry point.

## Identity modes

TradeLab should support two identity modes:

### 1. Guest demo mode

Guest mode is the default first-run experience.

Characteristics:

- starts without signup friction
- creates a temporary demo session and demo wallet
- uses virtual balances only
- is intended for first exploration and product evaluation
- does not imply durable ownership across devices unless explicitly upgraded

### 2. Registered demo account

Registered accounts are the durable mainline path.

Characteristics:

- backed by Clerk identity
- support repeat access across sessions and devices
- keep the product demo-only while introducing durable ownership
- become the long-term home for user portfolio state, preferences, and later strategy settings

## Locked decisions

- managed auth provider: `Clerk`
- guest mode stays
- primary social login targets:
  - Google
  - Apple
- provider model should remain extensible for future additions

## Why Clerk

Clerk is the planned auth provider because it supports:

- fast integration with Next.js
- managed session handling
- social login providers
- future support for account lifecycle features without building auth from scratch

TradeLab should avoid implementing bespoke password auth in the MVP when the product still needs to mature its core trading and simulation flows.

## Product behavior

### First launch

On first launch, TradeLab should continue to:

1. open directly into the product
2. create a guest demo session automatically
3. explain that the user is in a demo-only sandbox

### Upgrade path

At the appropriate point in the user journey, the product should offer an upgrade path from guest mode to a registered account.

The guest-upgrade path should:

- preserve the low-friction first experience
- make the durability benefit explicit
- avoid forcing registration before the user understands the product

## Account ownership model

### Guest session ownership

Guest sessions own:

- a temporary demo wallet
- temporary portfolio state
- temporary order and activity history

Guest sessions should be treated as revocable and time-bounded.

### Registered account ownership

Registered accounts should own:

- durable user identity
- one primary demo account at minimum
- durable portfolio, order, position, and activity history
- future strategy configuration and backtesting history

## Session model

TradeLab should separate:

- `identity`: handled by Clerk
- `application demo session`: handled by the TradeLab backend

This means a registered user can still have an application-level demo session, but that session is linked to a durable identity instead of existing only as an anonymous guest token.

## Future compatibility

This model should support:

- guest sessions
- guest-to-registered upgrades
- durable registered accounts
- future multi-account or team features without replacing the core identity boundary

## Related documentation

- [clerk-architecture.md](clerk-architecture.md)
- [auth-flows.md](auth-flows.md)
- [account-lifecycle.md](account-lifecycle.md)
- [onboarding-requirements.md](onboarding-requirements.md)
